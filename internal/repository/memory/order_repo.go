package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// OrderRepo is an in-memory implementation of domain.OrderRepository.
type OrderRepo struct {
	mu     sync.RWMutex
	orders map[string]domain.Order
}

// NewOrderRepo creates a new in-memory OrderRepo.
func NewOrderRepo() *OrderRepo {
	return &OrderRepo{
		orders: make(map[string]domain.Order),
	}
}

// Save persists an order (create or update).
func (r *OrderRepo) Save(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = *order
	return nil
}

// GetByID returns an order by ID, or nil if not found.
func (r *OrderRepo) GetByID(_ context.Context, id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok {
		return nil, nil
	}
	return &o, nil
}

// ListByUser returns orders for a user with optional filters and pagination.
func (r *OrderRepo) ListByUser(_ context.Context, userID string, filters domain.OrderFilters, params domain.PaginationParams) ([]domain.Order, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Order
	for _, o := range r.orders {
		if o.UserID != userID {
			continue
		}
		if filters.Status != nil && o.Status != *filters.Status {
			continue
		}
		filtered = append(filtered, o)
	}

	return sortAndPaginateOrders(filtered, filters, params)
}

// ListByBakery returns orders for a bakery with optional filters and pagination.
func (r *OrderRepo) ListByBakery(_ context.Context, bakeryID string, filters domain.OrderFilters, params domain.PaginationParams) ([]domain.Order, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Order
	for _, o := range r.orders {
		if o.BakeryID != bakeryID {
			continue
		}
		if filters.Status != nil && o.Status != *filters.Status {
			continue
		}
		filtered = append(filtered, o)
	}

	return sortAndPaginateOrders(filtered, filters, params)
}

// sortAndPaginateOrders applies sorting and pagination to a slice of orders.
func sortAndPaginateOrders(filtered []domain.Order, filters domain.OrderFilters, params domain.PaginationParams) ([]domain.Order, int, error) {
	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "createdAt"
	}
	sortDir := filters.SortDir
	if sortDir == "" {
		sortDir = "desc"
	}

	sort.Slice(filtered, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "scheduledTime":
			ti := filtered[i].ScheduledTime.StartTime
			tj := filtered[j].ScheduledTime.StartTime
			less = ti.Before(tj)
		default: // "createdAt"
			less = filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
		if sortDir == "desc" {
			return !less
		}
		return less
	})

	total := len(filtered)
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		return []domain.Order{}, total, nil
	}
	end := start + params.PageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}
