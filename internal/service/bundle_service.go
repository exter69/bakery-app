package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/ws"
)

// Sentinel errors for the bundle service.
var (
	ErrBundleNotFound                = errors.New("bundle not found")
	ErrBundleSoldOut                 = errors.New("bundle is sold out")
	ErrBundleNotDraft                = errors.New("bundle is not in draft status")
	ErrReservationExists             = errors.New("active reservation already exists for this bundle")
	ErrBundleReservationNotFound     = errors.New("bundle reservation not found")
	ErrBundleReservationNotCancellable = errors.New("bundle reservation cannot be cancelled in its current state")
)

// bundleService is the concrete implementation of domain.BundleService.
type bundleService struct {
	repo       domain.BundleRepository
	bakeryRepo domain.BakeryRepository
	hub        *ws.Hub
	idGen      func() string
	now        func() time.Time
}

// BundleServiceConfig holds dependencies for the bundle service.
type BundleServiceConfig struct {
	Repo       domain.BundleRepository
	BakeryRepo domain.BakeryRepository
	Hub        *ws.Hub
	IDGen      func() string
	Now        func() time.Time
}

// NewBundleService creates a new BundleService with the given dependencies.
func NewBundleService(cfg BundleServiceConfig) domain.BundleService {
	idGen := cfg.IDGen
	if idGen == nil {
		idGen = func() string { return uuid.New().String() }
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &bundleService{
		repo:       cfg.Repo,
		bakeryRepo: cfg.BakeryRepo,
		hub:        cfg.Hub,
		idGen:      idGen,
		now:        now,
	}
}

// CreateBundle validates and stores a new bundle in draft status.
func (s *bundleService) CreateBundle(ctx context.Context, sellerID string, bundle domain.SurplusBundle) (*domain.SurplusBundle, error) {
	if err := domain.ValidateBundle(bundle); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	bundle.ID = s.idGen()
	bundle.Status = domain.BundleStatusDraft
	bundle.QuantityRemaining = bundle.QuantityTotal

	// Generate IDs for each item
	for i := range bundle.Items {
		bundle.Items[i].ID = s.idGen()
		bundle.Items[i].BundleID = bundle.ID
	}

	if err := s.repo.CreateBundle(ctx, &bundle); err != nil {
		return nil, fmt.Errorf("creating bundle: %w", err)
	}

	return &bundle, nil
}

// PublishBundle transitions a draft bundle to published status.
func (s *bundleService) PublishBundle(ctx context.Context, sellerID string, bundleID string) (*domain.SurplusBundle, error) {
	// Verify the seller owns the bakery associated with this bundle
	bundle, err := s.repo.GetByID(ctx, bundleID)
	if err != nil {
		return nil, fmt.Errorf("fetching bundle: %w", err)
	}
	if bundle == nil {
		return nil, ErrBundleNotFound
	}

	bakery, err := s.bakeryRepo.GetBakery(ctx, bundle.BakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}
	if bakery.OwnerID != sellerID {
		return nil, ErrForbidden
	}

	if bundle.Status != domain.BundleStatusDraft {
		return nil, ErrBundleNotDraft
	}

	now := s.now()
	today := now.Format("2006-01-02")

	bundle.Status = domain.BundleStatusPublished
	bundle.PublishedDate = today

	// Compute expires_at from bakery's closing time on the current day of week
	todayDOW := timeToDayOfWeek(now)
	expiresAt := computeExpiresAt(now, todayDOW, bakery.Schedule)
	bundle.ExpiresAt = expiresAt

	if err := s.repo.UpdateBundle(ctx, bundle); err != nil {
		return nil, fmt.Errorf("updating bundle: %w", err)
	}

	return bundle, nil
}

// ListBundles returns published bundles, optionally filtered.
func (s *bundleService) ListBundles(ctx context.Context, filters domain.BundleFilters, params domain.PaginationParams) (*domain.ListResult[domain.SurplusBundle], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	bundles, total, err := s.repo.ListPublished(ctx, filters, params)
	if err != nil {
		return nil, fmt.Errorf("listing bundles: %w", err)
	}

	return &domain.ListResult[domain.SurplusBundle]{
		Items:    bundles,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

// GetBundle returns a single bundle by ID.
func (s *bundleService) GetBundle(ctx context.Context, bundleID string) (*domain.SurplusBundle, error) {
	bundle, err := s.repo.GetByID(ctx, bundleID)
	if err != nil {
		return nil, fmt.Errorf("fetching bundle: %w", err)
	}
	if bundle == nil {
		return nil, ErrBundleNotFound
	}
	return bundle, nil
}

// ReserveBundle creates a reservation for a customer, decrementing stock.
func (s *bundleService) ReserveBundle(ctx context.Context, customerID string, bundleID string) (*domain.BundleReservation, error) {
	bundle, err := s.repo.GetByID(ctx, bundleID)
	if err != nil {
		return nil, fmt.Errorf("fetching bundle: %w", err)
	}
	if bundle == nil {
		return nil, ErrBundleNotFound
	}

	if bundle.Status != domain.BundleStatusPublished {
		return nil, ErrBundleSoldOut
	}

	if bundle.QuantityRemaining <= 0 {
		return nil, ErrBundleSoldOut
	}

	// Check no existing active reservation for this user + bundle
	existing, err := s.repo.GetActiveReservation(ctx, customerID, bundleID)
	if err != nil {
		return nil, fmt.Errorf("checking active reservation: %w", err)
	}
	if existing != nil {
		return nil, ErrReservationExists
	}

	// Atomically decrement stock
	if err := s.repo.DecrementStock(ctx, bundleID); err != nil {
		return nil, ErrBundleSoldOut
	}

	// Create the reservation
	reservation := &domain.BundleReservation{
		ID:       s.idGen(),
		BundleID: bundleID,
		UserID:   customerID,
		Status:   domain.BundleReservationPending,
	}

	if err := s.repo.CreateReservation(ctx, reservation); err != nil {
		return nil, fmt.Errorf("creating reservation: %w", err)
	}

	// Re-fetch the bundle to check if sold out
	updated, err := s.repo.GetByID(ctx, bundleID)
	if err != nil {
		return nil, fmt.Errorf("re-fetching bundle: %w", err)
	}

	if updated != nil && updated.QuantityRemaining == 0 {
		updated.Status = domain.BundleStatusSoldOut
		if err := s.repo.UpdateBundle(ctx, updated); err != nil {
			return nil, fmt.Errorf("updating bundle to sold_out: %w", err)
		}
	}

	// Broadcast stock update event
	s.broadcastStockUpdate(bundleID, updated)

	return reservation, nil
}

// CancelReservation releases the reservation and increments stock.
func (s *bundleService) CancelReservation(ctx context.Context, customerID string, reservationID string) error {
	reservation, err := s.repo.GetReservation(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("fetching reservation: %w", err)
	}
	if reservation == nil {
		return ErrBundleReservationNotFound
	}

	if reservation.UserID != customerID {
		return ErrForbidden
	}

	// Only pending or confirmed reservations can be cancelled
	if reservation.Status != domain.BundleReservationPending && reservation.Status != domain.BundleReservationConfirmed {
		return ErrBundleReservationNotCancellable
	}

	reservation.Status = domain.BundleReservationCancelled
	if err := s.repo.UpdateReservation(ctx, reservation); err != nil {
		return fmt.Errorf("updating reservation: %w", err)
	}

	// Increment stock back
	if err := s.repo.IncrementStock(ctx, reservation.BundleID); err != nil {
		return fmt.Errorf("incrementing stock: %w", err)
	}

	// Re-fetch bundle to check if it was sold_out and should revert to published
	bundle, err := s.repo.GetByID(ctx, reservation.BundleID)
	if err != nil {
		return fmt.Errorf("re-fetching bundle: %w", err)
	}
	if bundle != nil && bundle.Status == domain.BundleStatusSoldOut {
		bundle.Status = domain.BundleStatusPublished
		if err := s.repo.UpdateBundle(ctx, bundle); err != nil {
			return fmt.Errorf("reverting bundle to published: %w", err)
		}
	}

	// Broadcast stock update event
	s.broadcastStockUpdate(reservation.BundleID, bundle)

	return nil
}

// ConfirmReservation transitions a pending reservation to confirmed.
func (s *bundleService) ConfirmReservation(ctx context.Context, customerID string, reservationID string) (*domain.BundleReservation, error) {
	reservation, err := s.repo.GetReservation(ctx, reservationID)
	if err != nil {
		return nil, fmt.Errorf("fetching reservation: %w", err)
	}
	if reservation == nil {
		return nil, ErrBundleReservationNotFound
	}

	if reservation.UserID != customerID {
		return nil, ErrForbidden
	}

	if reservation.Status != domain.BundleReservationPending {
		return nil, ErrBundleReservationNotCancellable
	}

	reservation.Status = domain.BundleReservationConfirmed
	if err := s.repo.UpdateReservation(ctx, reservation); err != nil {
		return nil, fmt.Errorf("updating reservation: %w", err)
	}

	return reservation, nil
}

// ExpireOverdueBundles finds and expires bundles past their expires_at time.
func (s *bundleService) ExpireOverdueBundles(ctx context.Context) (int, error) {
	expired, err := s.repo.GetExpiredBundles(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetching expired bundles: %w", err)
	}

	count := 0
	for i := range expired {
		expired[i].Status = domain.BundleStatusExpired
		if err := s.repo.UpdateBundle(ctx, &expired[i]); err != nil {
			return count, fmt.Errorf("expiring bundle %s: %w", expired[i].ID, err)
		}
		count++

		// Broadcast bundle_expired event
		if s.hub != nil {
			s.hub.Broadcast(ws.Event{
				Type: "bundle_expired",
				Payload: map[string]interface{}{
					"bundleId": expired[i].ID,
				},
			})
		}
	}

	return count, nil
}

// ReleaseOverdueReservations releases unconfirmed reservations past pickup_end_time.
func (s *bundleService) ReleaseOverdueReservations(ctx context.Context) (int, error) {
	overdue, err := s.repo.GetOverdueReservations(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetching overdue reservations: %w", err)
	}

	count := 0
	for i := range overdue {
		overdue[i].Status = domain.BundleReservationReleased
		if err := s.repo.UpdateReservation(ctx, &overdue[i]); err != nil {
			return count, fmt.Errorf("releasing reservation %s: %w", overdue[i].ID, err)
		}

		if err := s.repo.IncrementStock(ctx, overdue[i].BundleID); err != nil {
			return count, fmt.Errorf("incrementing stock for bundle %s: %w", overdue[i].BundleID, err)
		}

		count++
	}

	return count, nil
}

// GetImpact returns community impact metrics for the current month.
func (s *bundleService) GetImpact(ctx context.Context) (*domain.BundleImpact, error) {
	count, err := s.repo.CountPickedUpThisMonth(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting picked up: %w", err)
	}

	return &domain.BundleImpact{
		TotalSaved:    count,
		WeightAvoided: float64(count) * 0.5,
	}, nil
}

// broadcastStockUpdate sends a WebSocket event to all connected clients with updated stock info.
func (s *bundleService) broadcastStockUpdate(bundleID string, bundle *domain.SurplusBundle) {
	if s.hub == nil || bundle == nil {
		return
	}

	s.hub.Broadcast(ws.Event{
		Type: "bundle_stock_update",
		Payload: map[string]interface{}{
			"bundleId":          bundleID,
			"quantityRemaining": bundle.QuantityRemaining,
			"status":            string(bundle.Status),
		},
	})
}

// timeToDayOfWeek converts a time.Time weekday to a domain.DayOfWeek.
func timeToDayOfWeek(t time.Time) domain.DayOfWeek {
	switch t.Weekday() {
	case time.Monday:
		return domain.Monday
	case time.Tuesday:
		return domain.Tuesday
	case time.Wednesday:
		return domain.Wednesday
	case time.Thursday:
		return domain.Thursday
	case time.Friday:
		return domain.Friday
	case time.Saturday:
		return domain.Saturday
	case time.Sunday:
		return domain.Sunday
	default:
		return domain.Monday
	}
}

// computeExpiresAt determines the expiration timestamp from the bakery's closing time
// on the given day of week. Falls back to end-of-day (23:59) if no schedule is found.
func computeExpiresAt(now time.Time, dow domain.DayOfWeek, schedule []domain.DaySchedule) time.Time {
	for _, ds := range schedule {
		if ds.Day == dow && ds.IsOpen {
			return time.Date(
				now.Year(), now.Month(), now.Day(),
				ds.CloseTime.Hour, ds.CloseTime.Minute, 0, 0,
				now.Location(),
			)
		}
	}

	// Fallback: end of current day
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
}
