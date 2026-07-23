package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/lucatorrekens/bakery-app/internal/validation"
)

// PaymentGateway is a local alias to avoid circular imports.
type PaymentGateway = payment.PaymentGateway

// orderService is the concrete implementation of domain.OrderService.
type orderService struct {
	orderRepo        domain.OrderRepository
	bakeryRepo       domain.BakeryRepository
	userRepo         domain.UserRepository
	paymentSvc       domain.PaymentService
	paymentGateway   PaymentGateway
	onOrderCancelled func(ctx context.Context, orderID string, refunded bool) error
	onNewOrder       func(ctx context.Context, orderID string) error
	idGen            func() string
	now              func() time.Time
	rng              *rand.Rand
}

// OrderServiceConfig holds dependencies for the order service.
type OrderServiceConfig struct {
	OrderRepo        domain.OrderRepository
	BakeryRepo       domain.BakeryRepository
	UserRepo         domain.UserRepository
	PaymentSvc       domain.PaymentService
	PaymentGateway   PaymentGateway
	OnOrderCancelled func(ctx context.Context, orderID string, refunded bool) error
	OnNewOrder       func(ctx context.Context, orderID string) error
	IDGen            func() string
	Now              func() time.Time
}

// NewOrderService creates a new OrderService with the given dependencies.
func NewOrderService(cfg OrderServiceConfig) domain.OrderService {
	idGen := cfg.IDGen
	if idGen == nil {
		idGen = defaultIDGen
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &orderService{
		orderRepo:        cfg.OrderRepo,
		bakeryRepo:       cfg.BakeryRepo,
		userRepo:         cfg.UserRepo,
		paymentSvc:       cfg.PaymentSvc,
		paymentGateway:   cfg.PaymentGateway,
		onOrderCancelled: cfg.OnOrderCancelled,
		onNewOrder:       cfg.OnNewOrder,
		idGen:            idGen,
		now:              now,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// defaultIDGen generates a simple ID. In production this would use UUID.
var idCounter int

func defaultIDGen() string {
	idCounter++
	return fmt.Sprintf("order-%d", idCounter)
}

// CreateOrder validates and creates a new delivery order.
func (s *orderService) CreateOrder(ctx context.Context, userID string, order domain.Order) (*domain.Order, *domain.PaymentLink, error) {
	// Determine selection mode (default to fixed)
	selectionMode := order.SelectionMode
	if selectionMode == "" {
		selectionMode = domain.SelectionFixed
	}

	// Step 2: Fetch bakery and validate schedule
	bakery, err := s.bakeryRepo.GetBakery(ctx, order.BakeryID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, nil, ErrBakeryNotFound
	}

	scheduleResult := validation.ValidateSchedule(order.ScheduledDay, order.ScheduledTime, bakery.Schedule)
	if scheduleResult.HasErrors() {
		return nil, nil, &ValidationErrors{Errors: scheduleResult.Errors}
	}

	// Step 3: Fetch products
	products, err := s.bakeryRepo.GetProductsByBakery(ctx, order.BakeryID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching products: %w", err)
	}

	// Handle selection modes
	switch selectionMode {
	case domain.SelectionRandomFavorites:
		// Fetch user's favorites and randomly pick 2-4 items
		items, err := s.pickRandomFavorites(ctx, userID, products)
		if err != nil {
			return nil, nil, err
		}
		order.Items = items

	case domain.SelectionBakeryChoice:
		// Leave items empty — bakery fills them later; just mark the order
		order.Items = []domain.OrderItem{}

	default:
		// "fixed" mode: validate provided items
		vr := validation.ValidateOrderItemsQuantity(order.Items)
		if vr.HasErrors() {
			return nil, nil, &ValidationErrors{Errors: vr.Errors}
		}

		availResult := validation.ValidateProductAvailability(order.Items, products)
		if availResult.HasErrors() {
			return nil, nil, &ValidationErrors{Errors: availResult.Errors}
		}
	}

	// Step 4: Enrich items with product names and prices from the product catalog
	productMap := make(map[string]domain.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}
	for i := range order.Items {
		if p, ok := productMap[order.Items[i].ProductID]; ok {
			order.Items[i].ProductName = p.Name
			order.Items[i].UnitPrice = p.Price
		}
	}

	// Step 5: Validate enriched items (unit price > 0) — skip for bakery_choice with no items
	if selectionMode != domain.SelectionBakeryChoice && len(order.Items) > 0 {
		fullValidation := validation.ValidateOrderItems(order.Items)
		if fullValidation.HasErrors() {
			return nil, nil, &ValidationErrors{Errors: fullValidation.Errors}
		}
	}

	// Step 6: Calculate total
	total := domain.CalculateOrderTotal(order.Items)

	// Step 7: Build the order
	now := s.now()
	order.ID = s.idGen()
	order.UserID = userID
	order.Status = domain.OrderStatusPendingPayment
	order.PaymentMethod = domain.PaymentMethodOnline
	order.SelectionMode = selectionMode
	order.TotalAmount = total
	order.CreatedAt = now
	order.UpdatedAt = now

	// Step 8: Persist the order
	if err := s.orderRepo.Save(ctx, &order); err != nil {
		return nil, nil, fmt.Errorf("saving order: %w", err)
	}

	// Step 9: Initiate payment (skip for bakery_choice with zero total)
	if total == 0 {
		return &order, nil, nil
	}

	paymentLink, err := s.paymentSvc.InitiatePayment(ctx, order.ID, total)
	if err != nil {
		return nil, nil, fmt.Errorf("initiating payment: %w", err)
	}

	// Fire-and-forget: alert the baker about the new order
	if s.onNewOrder != nil {
		go func() {
			if err := s.onNewOrder(context.Background(), order.ID); err != nil {
				log.Printf("[NOTIFICATION] new order alert failed for order %s: %v", order.ID, err)
			}
		}()
	}

	return &order, paymentLink, nil
}

// pickRandomFavorites selects 2-4 random items from the user's favorites that belong to the bakery's product catalog.
func (s *orderService) pickRandomFavorites(ctx context.Context, userID string, products []domain.Product) ([]domain.OrderItem, error) {
	if s.userRepo == nil {
		return nil, fmt.Errorf("user repository not configured")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user favorites: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if len(user.FavoriteProducts) == 0 {
		return nil, &ValidationErrors{Errors: []validation.ValidationError{
			{Field: "favorites", Message: "no favorite products configured; add favorites first"},
		}}
	}

	// Filter favorites to only include products available in this bakery
	productSet := make(map[string]bool, len(products))
	for _, p := range products {
		if p.IsAvailable {
			productSet[p.ID] = true
		}
	}

	var eligible []string
	for _, fav := range user.FavoriteProducts {
		if productSet[fav] {
			eligible = append(eligible, fav)
		}
	}

	if len(eligible) == 0 {
		return nil, &ValidationErrors{Errors: []validation.ValidationError{
			{Field: "favorites", Message: "none of your favorites are available at this bakery"},
		}}
	}

	// Pick 2-4 random items
	count := 2 + s.rng.Intn(3) // 2, 3, or 4
	if count > len(eligible) {
		count = len(eligible)
	}

	// Shuffle and pick
	shuffled := make([]string, len(eligible))
	copy(shuffled, eligible)
	s.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	items := make([]domain.OrderItem, count)
	for i := 0; i < count; i++ {
		items[i] = domain.OrderItem{
			ProductID: shuffled[i],
			Quantity:  1,
		}
	}

	return items, nil
}

// GetOrders returns a paginated list of orders for a user.
func (s *orderService) GetOrders(ctx context.Context, userID string, filters domain.OrderFilters, params domain.PaginationParams) (*domain.ListResult[domain.Order], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	orders, total, err := s.orderRepo.ListByUser(ctx, userID, filters, params)
	if err != nil {
		return nil, fmt.Errorf("listing orders: %w", err)
	}

	return &domain.ListResult[domain.Order]{
		Items:    orders,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

// DeleteOrder cancels an order after verifying ownership and state.
func (s *orderService) DeleteOrder(ctx context.Context, orderID string, userID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("fetching order: %w", err)
	}
	if order == nil {
		return ErrOrderNotFound
	}
	if order.UserID != userID {
		return ErrForbidden
	}
	if order.Status == domain.OrderStatusDelivered || order.Status == domain.OrderStatusCancelled {
		return ErrOrderNotCancellable
	}

	// Determine payment action needed based on state *before* cancellation
	previousStatus := order.Status
	hasPaymentIntent := order.PaymentIntentID != ""
	wasPaid := previousStatus == domain.OrderStatusConfirmed ||
		previousStatus == domain.OrderStatusPreparing ||
		previousStatus == domain.OrderStatusReady

	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = s.now()
	if err := s.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("saving order: %w", err)
	}

	// Void the authorization (delayed capture: funds were held, not captured)
	if hasPaymentIntent && wasPaid && s.paymentGateway != nil {
		refunded := false
		if err := s.paymentGateway.VoidAuthorization(ctx, order.PaymentIntentID); err != nil {
			// Void failed — payment was likely already captured. Try a full refund.
			log.Printf("[PAYMENT] void failed for order %s, attempting refund: %v", order.ID, err)
			if refErr := s.paymentGateway.RefundPayment(ctx, order.PaymentIntentID, 0); refErr != nil {
				log.Printf("[PAYMENT] refund also failed for order %s: %v", order.ID, refErr)
			} else {
				refunded = true
				order.RefundStatus = "refunded"
				_ = s.orderRepo.Save(ctx, order) // best-effort status update
			}
		}
		// Send cancellation notification (non-blocking)
		if s.onOrderCancelled != nil {
			if err := s.onOrderCancelled(ctx, order.ID, refunded); err != nil {
				log.Printf("[NOTIFICATION] cancellation notification failed for order %s: %v", order.ID, err)
			}
		}
		return nil
	}

	// Fallback: initiate refund for orders without delayed capture (legacy flow)
	if !hasPaymentIntent && wasPaid && order.TotalAmount > 0 {
		if err := s.paymentSvc.InitiateRefund(ctx, order.ID, order.TotalAmount); err != nil {
			log.Printf("[PAYMENT] failed to initiate refund for order %s: %v", order.ID, err)
		}
	}

	// Send cancellation notification for non-payment-intent orders
	if s.onOrderCancelled != nil {
		if err := s.onOrderCancelled(ctx, order.ID, false); err != nil {
			log.Printf("[NOTIFICATION] cancellation notification failed for order %s: %v", order.ID, err)
		}
	}

	return nil
}
