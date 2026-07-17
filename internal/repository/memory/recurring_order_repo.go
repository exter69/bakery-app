package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// RecurringOrderRepo is an in-memory implementation of domain.RecurringOrderRepository.
type RecurringOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]domain.RecurringOrder
}

// NewRecurringOrderRepo creates a new in-memory RecurringOrderRepo.
func NewRecurringOrderRepo() *RecurringOrderRepo {
	return &RecurringOrderRepo{
		orders: make(map[string]domain.RecurringOrder),
	}
}

// Save persists a recurring order (create or update).
func (r *RecurringOrderRepo) Save(_ context.Context, order *domain.RecurringOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = *order
	return nil
}

// GetByID returns a recurring order by ID, or nil if not found.
func (r *RecurringOrderRepo) GetByID(_ context.Context, id string) (*domain.RecurringOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok {
		return nil, nil
	}
	return &o, nil
}

// ListByUser returns recurring orders for a user with pagination.
func (r *RecurringOrderRepo) ListByUser(_ context.Context, userID string, params domain.PaginationParams) ([]domain.RecurringOrder, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.RecurringOrder
	for _, o := range r.orders {
		if o.UserID == userID {
			filtered = append(filtered, o)
		}
	}

	// Sort by createdAt descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		return []domain.RecurringOrder{}, total, nil
	}
	end := start + params.PageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// ListActive returns all active recurring orders.
func (r *RecurringOrderRepo) ListActive(_ context.Context) ([]domain.RecurringOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var active []domain.RecurringOrder
	for _, o := range r.orders {
		if o.Active {
			active = append(active, o)
		}
	}
	return active, nil
}

// Delete removes a recurring order by ID.
func (r *RecurringOrderRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.orders, id)
	return nil
}
