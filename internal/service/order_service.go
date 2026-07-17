package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/validation"
)

// orderService is the concrete implementation of domain.OrderService.
type orderService struct {
	orderRepo  domain.OrderRepository
	bakeryRepo domain.BakeryRepository
	userRepo   domain.UserRepository
	paymentSvc domain.PaymentService
	idGen      func() string
	now        func() time.Time
	rng        *rand.Rand
}

// OrderServiceConfig holds dependencies for the order service.
type OrderServiceConfig struct {
	OrderRepo  domain.OrderRepository
	BakeryRepo domain.BakeryRepository
	UserRepo   domain.UserRepository
	PaymentSvc domain.PaymentService
	IDGen      func() string
	Now        func() time.Time
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
		orderRepo:  cfg.OrderRepo,
		bakeryRepo: cfg.BakeryRepo,
		userRepo:   cfg.UserRepo,
		paymentSvc: cfg.PaymentSvc,
		idGen:      idGen,
		now:        now,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
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

	// Check if refund is needed before status change
	needsRefund := order.Status == domain.OrderStatusConfirmed || order.Status == domain.OrderStatusPreparing

	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = s.now()
	if err := s.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("saving order: %w", err)
	}

	// Initiate refund for orders that had payment confirmed
	if needsRefund {
		if err := s.paymentSvc.InitiateRefund(ctx, order.ID, order.TotalAmount); err != nil {
			// Log the error but don't fail the cancellation
			// In production, this would be handled by a retry mechanism
			return fmt.Errorf("initiating refund: %w", err)
		}
	}

	return nil
}
