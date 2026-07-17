package service

import (
	"context"
	"fmt"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// SellerServiceConfig holds dependencies for the seller service.
type SellerServiceConfig struct {
	BakeryRepo      domain.BakeryRepository
	OrderRepo       domain.OrderRepository
	ReservationRepo domain.ReservationRepository
}

// SellerService handles seller-specific operations.
type SellerService struct {
	bakeryRepo      domain.BakeryRepository
	orderRepo       domain.OrderRepository
	reservationRepo domain.ReservationRepository
}

// NewSellerService creates a new SellerService.
func NewSellerService(cfg SellerServiceConfig) *SellerService {
	return &SellerService{
		bakeryRepo:      cfg.BakeryRepo,
		orderRepo:       cfg.OrderRepo,
		reservationRepo: cfg.ReservationRepo,
	}
}

// UpdateBakery updates a bakery's info after verifying ownership.
func (s *SellerService) UpdateBakery(ctx context.Context, bakeryID, ownerID string, name, description, address, photoURL *string) (*domain.Bakery, error) {
	bakery, err := s.bakeryRepo.GetBakery(ctx, bakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	if name != nil {
		bakery.Name = *name
	}
	if description != nil {
		bakery.Description = *description
	}
	if address != nil {
		bakery.Address = *address
	}
	if photoURL != nil {
		bakery.PhotoURL = *photoURL
	}

	if err := s.bakeryRepo.UpdateBakery(ctx, bakery); err != nil {
		return nil, fmt.Errorf("updating bakery: %w", err)
	}
	return bakery, nil
}

// UpdateBakerySchedule updates a bakery's schedule after verifying ownership.
func (s *SellerService) UpdateBakerySchedule(ctx context.Context, bakeryID, ownerID string, schedule []domain.DaySchedule) (*domain.Bakery, error) {
	bakery, err := s.bakeryRepo.GetBakery(ctx, bakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	bakery.Schedule = schedule

	if err := s.bakeryRepo.UpdateBakery(ctx, bakery); err != nil {
		return nil, fmt.Errorf("updating bakery schedule: %w", err)
	}
	return bakery, nil
}

// CreateProduct adds a product to a bakery after verifying ownership.
func (s *SellerService) CreateProduct(ctx context.Context, bakeryID, ownerID string, product domain.Product) (*domain.Product, error) {
	bakery, err := s.bakeryRepo.GetBakery(ctx, bakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	product.BakeryID = bakeryID
	product.ID = fmt.Sprintf("prod_%d", time.Now().UnixNano())
	product.IsAvailable = true

	if err := s.bakeryRepo.CreateProduct(ctx, &product); err != nil {
		return nil, fmt.Errorf("creating product: %w", err)
	}
	return &product, nil
}

// UpdateProduct updates a product after verifying bakery ownership.
func (s *SellerService) UpdateProduct(ctx context.Context, productID, ownerID string, updates map[string]interface{}) (*domain.Product, error) {
	product, err := s.bakeryRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("fetching product: %w", err)
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	// Verify bakery ownership
	bakery, err := s.bakeryRepo.GetBakery(ctx, product.BakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	// Apply updates
	if v, ok := updates["name"].(string); ok {
		product.Name = v
	}
	if v, ok := updates["description"].(string); ok {
		product.Description = v
	}
	if v, ok := updates["price"].(float64); ok {
		product.Price = int64(v)
	}
	if v, ok := updates["photoUrl"].(string); ok {
		product.PhotoURL = v
	}
	if v, ok := updates["category"].(string); ok {
		product.Category = v
	}
	if v, ok := updates["isAvailable"].(bool); ok {
		product.IsAvailable = v
	}

	if err := s.bakeryRepo.UpdateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("updating product: %w", err)
	}
	return product, nil
}

// DeleteProduct removes a product after verifying bakery ownership.
func (s *SellerService) DeleteProduct(ctx context.Context, productID, ownerID string) error {
	product, err := s.bakeryRepo.GetProductByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("fetching product: %w", err)
	}
	if product == nil {
		return ErrProductNotFound
	}

	// Verify bakery ownership
	bakery, err := s.bakeryRepo.GetBakery(ctx, product.BakeryID)
	if err != nil {
		return fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return ErrForbidden
	}

	if err := s.bakeryRepo.DeleteProduct(ctx, productID); err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}
	return nil
}

// ListBakeryOrders returns paginated orders for a bakery after verifying ownership.
func (s *SellerService) ListBakeryOrders(ctx context.Context, bakeryID, ownerID string, filters domain.OrderFilters, params domain.PaginationParams) (*domain.ListResult[domain.Order], error) {
	bakery, err := s.bakeryRepo.GetBakery(ctx, bakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	orders, total, err := s.orderRepo.ListByBakery(ctx, bakeryID, filters, params)
	if err != nil {
		return nil, fmt.Errorf("listing bakery orders: %w", err)
	}

	return &domain.ListResult[domain.Order]{
		Items:    orders,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

// ListBakeryReservations returns paginated reservations for a bakery after verifying ownership.
func (s *SellerService) ListBakeryReservations(ctx context.Context, bakeryID, ownerID string, filters domain.ReservationFilters, params domain.PaginationParams) (*domain.ListResult[domain.Reservation], error) {
	bakery, err := s.bakeryRepo.GetBakery(ctx, bakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	reservations, total, err := s.reservationRepo.ListByBakery(ctx, bakeryID, filters, params)
	if err != nil {
		return nil, fmt.Errorf("listing bakery reservations: %w", err)
	}

	return &domain.ListResult[domain.Reservation]{
		Items:    reservations,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

// validOrderTransitions defines allowed order status transitions for sellers.
var validOrderTransitions = map[domain.OrderStatus]domain.OrderStatus{
	domain.OrderStatusConfirmed: domain.OrderStatusPreparing,
	domain.OrderStatusPreparing: domain.OrderStatusReady,
	domain.OrderStatusReady:     domain.OrderStatusDelivered,
}

// UpdateOrderStatus updates an order's status after verifying ownership and valid transition.
func (s *SellerService) UpdateOrderStatus(ctx context.Context, orderID, ownerID string, newStatus domain.OrderStatus) (*domain.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("fetching order: %w", err)
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}

	// Verify bakery ownership
	bakery, err := s.bakeryRepo.GetBakery(ctx, order.BakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	// Validate transition
	allowed, ok := validOrderTransitions[order.Status]
	if !ok || allowed != newStatus {
		return nil, ErrInvalidStatusTransition
	}

	order.Status = newStatus
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("saving order: %w", err)
	}
	return order, nil
}

// validReservationTransitions defines allowed reservation status transitions for sellers.
var validReservationTransitions = map[domain.ReservationStatus]domain.ReservationStatus{
	domain.ReservationStatusConfirmed: domain.ReservationStatusReady,
	domain.ReservationStatusReady:     domain.ReservationStatusPickedUp,
}

// UpdateReservationStatus updates a reservation's status after verifying ownership and valid transition.
func (s *SellerService) UpdateReservationStatus(ctx context.Context, reservationID, ownerID string, newStatus domain.ReservationStatus) (*domain.Reservation, error) {
	reservation, err := s.reservationRepo.Get(ctx, reservationID)
	if err != nil {
		return nil, fmt.Errorf("fetching reservation: %w", err)
	}
	if reservation == nil {
		return nil, ErrReservationNotFound
	}

	// Verify bakery ownership
	bakery, err := s.bakeryRepo.GetBakery(ctx, reservation.BakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	// Validate transition
	allowed, ok := validReservationTransitions[reservation.Status]
	if !ok || allowed != newStatus {
		return nil, ErrInvalidStatusTransition
	}

	reservation.Status = newStatus

	if err := s.reservationRepo.Save(ctx, *reservation); err != nil {
		return nil, fmt.Errorf("saving reservation: %w", err)
	}
	return reservation, nil
}
