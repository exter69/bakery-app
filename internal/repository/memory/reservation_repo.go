package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// ReservationRepo is an in-memory implementation of domain.ReservationRepository.
type ReservationRepo struct {
	mu           sync.RWMutex
	reservations map[string]domain.Reservation
}

// NewReservationRepo creates a new in-memory ReservationRepo.
func NewReservationRepo() *ReservationRepo {
	return &ReservationRepo{
		reservations: make(map[string]domain.Reservation),
	}
}

// Save persists a reservation in memory.
func (r *ReservationRepo) Save(_ context.Context, reservation domain.Reservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reservations[reservation.ID] = reservation
	return nil
}

// Get returns a reservation by ID, or nil if not found.
func (r *ReservationRepo) Get(_ context.Context, id string) (*domain.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res, ok := r.reservations[id]
	if !ok {
		return nil, nil
	}
	return &res, nil
}

// ListByUser returns reservations for a user with optional filters and pagination.
func (r *ReservationRepo) ListByUser(_ context.Context, userID string, filters domain.ReservationFilters, params domain.PaginationParams) ([]domain.Reservation, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Reservation
	for _, res := range r.reservations {
		if res.UserID != userID {
			continue
		}
		if filters.Status != nil && res.Status != *filters.Status {
			continue
		}
		filtered = append(filtered, res)
	}

	return sortAndPaginateReservations(filtered, filters, params)
}

// ListByBakery returns reservations for a bakery with optional filters and pagination.
func (r *ReservationRepo) ListByBakery(_ context.Context, bakeryID string, filters domain.ReservationFilters, params domain.PaginationParams) ([]domain.Reservation, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Reservation
	for _, res := range r.reservations {
		if res.BakeryID != bakeryID {
			continue
		}
		if filters.Status != nil && res.Status != *filters.Status {
			continue
		}
		filtered = append(filtered, res)
	}

	return sortAndPaginateReservations(filtered, filters, params)
}

// sortAndPaginateReservations applies sorting and pagination to a slice of reservations.
func sortAndPaginateReservations(filtered []domain.Reservation, filters domain.ReservationFilters, params domain.PaginationParams) ([]domain.Reservation, int, error) {
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
		return []domain.Reservation{}, total, nil
	}
	end := start + params.PageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}
