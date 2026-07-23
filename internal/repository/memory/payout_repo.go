package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// PayoutRepo is an in-memory implementation of domain.PayoutRepository.
type PayoutRepo struct {
	mu      sync.RWMutex
	payouts map[string]domain.Payout
}

// NewPayoutRepo creates a new in-memory PayoutRepo.
func NewPayoutRepo() *PayoutRepo {
	return &PayoutRepo{
		payouts: make(map[string]domain.Payout),
	}
}

// Create persists a new payout record.
func (r *PayoutRepo) Create(_ context.Context, payout *domain.Payout) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if payout.ID == "" {
		payout.ID = uuid.New().String()
	}
	r.payouts[payout.ID] = *payout
	return nil
}

// GetByOrderID returns the payout for a given order, or nil if not found.
func (r *PayoutRepo) GetByOrderID(_ context.Context, orderID string) (*domain.Payout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.payouts {
		if p.OrderID == orderID {
			return &p, nil
		}
	}
	return nil, nil
}

// Update persists changes to a payout.
func (r *PayoutRepo) Update(_ context.Context, payout *domain.Payout) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payouts[payout.ID] = *payout
	return nil
}

// ListByBakery returns paginated payouts for a bakery.
func (r *PayoutRepo) ListByBakery(_ context.Context, bakeryID string, params domain.PaginationParams) ([]domain.Payout, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []domain.Payout
	for _, p := range r.payouts {
		if p.BakeryID == bakeryID {
			all = append(all, p)
		}
	}

	// Sort by created_at descending (newest first)
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].CreatedAt.After(all[i].CreatedAt) {
				all[i], all[j] = all[j], all[i]
			}
		}
	}

	total := len(all)
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		return []domain.Payout{}, total, nil
	}
	end := start + params.PageSize
	if end > total {
		end = total
	}

	return all[start:end], total, nil
}
