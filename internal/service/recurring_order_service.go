package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// RecurringOrderService handles recurring order CRUD operations.
type RecurringOrderService struct {
	recurringRepo domain.RecurringOrderRepository
	bakeryRepo    domain.BakeryRepository
	idGen         func() string
	now           func() time.Time
}

// RecurringOrderServiceConfig holds dependencies for the recurring order service.
type RecurringOrderServiceConfig struct {
	RecurringRepo domain.RecurringOrderRepository
	BakeryRepo    domain.BakeryRepository
	IDGen         func() string
	Now           func() time.Time
}

// NewRecurringOrderService creates a new RecurringOrderService.
func NewRecurringOrderService(cfg RecurringOrderServiceConfig) *RecurringOrderService {
	idGen := cfg.IDGen
	if idGen == nil {
		idGen = defaultRecurringIDGen
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &RecurringOrderService{
		recurringRepo: cfg.RecurringRepo,
		bakeryRepo:    cfg.BakeryRepo,
		idGen:         idGen,
		now:           now,
	}
}

func defaultRecurringIDGen() string {
	return uuid.New().String()
}

// CreateRecurringOrder validates and creates a new recurring order.
func (s *RecurringOrderService) CreateRecurringOrder(ctx context.Context, userID string, order domain.RecurringOrder) (*domain.RecurringOrder, error) {
	// Validate bakery exists
	bakery, err := s.bakeryRepo.GetBakery(ctx, order.BakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}

	// Validate items are not empty
	if len(order.Items) == 0 {
		return nil, ErrRecurringOrderNoItems
	}

	// Validate frequency
	if order.Frequency != domain.FrequencyWeekly && order.Frequency != domain.FrequencyBiWeekly {
		return nil, ErrInvalidFrequency
	}

	// Validate selection mode
	if order.SelectionMode != domain.SelectionFixed &&
		order.SelectionMode != domain.SelectionBakeryChoice &&
		order.SelectionMode != domain.SelectionRandomFavorites {
		return nil, ErrInvalidSelectionMode
	}

	now := s.now()
	order.ID = s.idGen()
	order.UserID = userID
	order.Active = true
	order.CreatedAt = now
	order.UpdatedAt = now

	if err := s.recurringRepo.Save(ctx, &order); err != nil {
		return nil, fmt.Errorf("saving recurring order: %w", err)
	}

	return &order, nil
}

// PauseRecurringOrder sets a recurring order to inactive.
func (s *RecurringOrderService) PauseRecurringOrder(ctx context.Context, id, userID string) error {
	order, err := s.recurringRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("fetching recurring order: %w", err)
	}
	if order == nil {
		return ErrRecurringOrderNotFound
	}
	if order.UserID != userID {
		return ErrForbidden
	}

	order.Active = false
	order.UpdatedAt = s.now()
	return s.recurringRepo.Save(ctx, order)
}

// ResumeRecurringOrder sets a recurring order to active.
func (s *RecurringOrderService) ResumeRecurringOrder(ctx context.Context, id, userID string) error {
	order, err := s.recurringRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("fetching recurring order: %w", err)
	}
	if order == nil {
		return ErrRecurringOrderNotFound
	}
	if order.UserID != userID {
		return ErrForbidden
	}

	order.Active = true
	order.UpdatedAt = s.now()
	return s.recurringRepo.Save(ctx, order)
}

// DeleteRecurringOrder removes a recurring order after ownership check.
func (s *RecurringOrderService) DeleteRecurringOrder(ctx context.Context, id, userID string) error {
	order, err := s.recurringRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("fetching recurring order: %w", err)
	}
	if order == nil {
		return ErrRecurringOrderNotFound
	}
	if order.UserID != userID {
		return ErrForbidden
	}

	return s.recurringRepo.Delete(ctx, id)
}

// ListMyRecurringOrders returns paginated recurring orders for a user.
func (s *RecurringOrderService) ListMyRecurringOrders(ctx context.Context, userID string, params domain.PaginationParams) (*domain.ListResult[domain.RecurringOrder], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	orders, total, err := s.recurringRepo.ListByUser(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("listing recurring orders: %w", err)
	}

	return &domain.ListResult[domain.RecurringOrder]{
		Items:    orders,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}
