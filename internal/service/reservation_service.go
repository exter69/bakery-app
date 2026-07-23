package service

import (
	"context"
	"fmt"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/notification"
	"github.com/lucatorrekens/bakery-app/internal/validation"
)

// reservationService is the concrete implementation of domain.ReservationService.
type reservationService struct {
	bakeryRepo      domain.BakeryRepository
	reservationRepo domain.ReservationRepository
	notifications   notification.Dispatcher
	idGen           func() string
	now             func() time.Time
}

// ReservationServiceConfig holds dependencies for the reservation service.
type ReservationServiceConfig struct {
	BakeryRepo      domain.BakeryRepository
	ReservationRepo domain.ReservationRepository
	Notifications   notification.Dispatcher
	IDGen           func() string
	Now             func() time.Time
}

// NewReservationService creates a new ReservationService with the given dependencies.
func NewReservationService(cfg ReservationServiceConfig) domain.ReservationService {
	idGen := cfg.IDGen
	if idGen == nil {
		idGen = defaultReservationIDGen
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &reservationService{
		bakeryRepo:      cfg.BakeryRepo,
		reservationRepo: cfg.ReservationRepo,
		notifications:   cfg.Notifications,
		idGen:           idGen,
		now:             now,
	}
}

var reservationIDCounter int

func defaultReservationIDGen() string {
	reservationIDCounter++
	return fmt.Sprintf("reservation-%d", reservationIDCounter)
}

// CreateReservation validates and creates a new reservation with on-spot payment.
func (s *reservationService) CreateReservation(ctx context.Context, userID string, reservation domain.Reservation) (*domain.Reservation, error) {
	// Step 1: Validate items have non-empty list and valid quantities (before price enrichment)
	qtyResult := validation.ValidateOrderItemsQuantity(reservation.Items)
	if qtyResult.HasErrors() {
		return nil, &ValidationErrors{Errors: qtyResult.Errors}
	}

	// Step 2: Fetch bakery and validate schedule
	bakery, err := s.bakeryRepo.GetBakery(ctx, reservation.BakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}

	scheduleResult := validation.ValidateSchedule(reservation.ScheduledDay, reservation.ScheduledTime, bakery.Schedule)
	if scheduleResult.HasErrors() {
		return nil, &ValidationErrors{Errors: scheduleResult.Errors}
	}

	// Step 3: Fetch products and validate availability
	products, err := s.bakeryRepo.GetProductsByBakery(ctx, reservation.BakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching products: %w", err)
	}

	availabilityResult := validation.ValidateProductAvailability(reservation.Items, products)
	if availabilityResult.HasErrors() {
		return nil, &ValidationErrors{Errors: availabilityResult.Errors}
	}

	// Step 4: Enrich items with product info (name, price) from catalog
	productMap := make(map[string]domain.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}
	for i := range reservation.Items {
		if p, ok := productMap[reservation.Items[i].ProductID]; ok {
			reservation.Items[i].ProductName = p.Name
			reservation.Items[i].UnitPrice = p.Price
		}
	}

	// Step 5: Validate enriched items (unit price > 0) and calculate total
	fullValidation := validation.ValidateOrderItems(reservation.Items)
	if fullValidation.HasErrors() {
		return nil, &ValidationErrors{Errors: fullValidation.Errors}
	}
	reservation.TotalAmount = domain.CalculateOrderTotal(reservation.Items)

	// Step 5: Always set payment method to OnSpot and status to Confirmed
	reservation.PaymentMethod = domain.PaymentMethodOnSpot
	reservation.Status = domain.ReservationStatusConfirmed

	// Step 6: Set metadata
	reservation.ID = s.idGen()
	reservation.UserID = userID
	reservation.CreatedAt = s.now()

	// Step 7: Persist
	if err := s.reservationRepo.Save(ctx, reservation); err != nil {
		return nil, fmt.Errorf("saving reservation: %w", err)
	}

	// Fire-and-forget: send confirmation email to the customer
	if s.notifications != nil {
		go s.notifications.OnReservationConfirmed(context.Background(), reservation.ID)
	}

	return &reservation, nil
}

// GetReservations returns a paginated list of reservations for a user.
func (s *reservationService) GetReservations(ctx context.Context, userID string, filters domain.ReservationFilters, params domain.PaginationParams) (*domain.ListResult[domain.Reservation], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	reservations, total, err := s.reservationRepo.ListByUser(ctx, userID, filters, params)
	if err != nil {
		return nil, fmt.Errorf("listing reservations: %w", err)
	}

	return &domain.ListResult[domain.Reservation]{
		Items:    reservations,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

// DeleteReservation cancels a reservation after verifying ownership and state.
func (s *reservationService) DeleteReservation(ctx context.Context, reservationID string, userID string) error {
	reservation, err := s.reservationRepo.Get(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("fetching reservation: %w", err)
	}
	if reservation == nil {
		return ErrReservationNotFound
	}
	if reservation.UserID != userID {
		return ErrForbidden
	}
	if reservation.Status == domain.ReservationStatusPickedUp || reservation.Status == domain.ReservationStatusCancelled {
		return ErrReservationNotCancellable
	}

	reservation.Status = domain.ReservationStatusCancelled
	return s.reservationRepo.Save(ctx, *reservation)
}
