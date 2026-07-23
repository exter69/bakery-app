package domain

import (
	"context"
	"time"
)

// PayoutStatus represents the current state of a payout.
type PayoutStatus string

const (
	PayoutStatusPending     PayoutStatus = "pending"
	PayoutStatusTransferred PayoutStatus = "transferred"
	PayoutStatusFailed      PayoutStatus = "failed"
	PayoutStatusRefunded    PayoutStatus = "refunded"
)

// Payout represents a marketplace payout to a bakery's connected Stripe account.
type Payout struct {
	ID               string       `json:"id"`
	OrderID          string       `json:"orderId"`
	BakeryID         string       `json:"bakeryId"`
	Amount           int64        `json:"amount"`     // bakery share in cents
	Commission       int64        `json:"commission"` // platform fee in cents
	StripeTransferID string       `json:"stripeTransferId,omitempty"`
	Status           PayoutStatus `json:"status"`
	CreatedAt        time.Time    `json:"createdAt"`
	TransferredAt    *time.Time   `json:"transferredAt,omitempty"`
}

// PayoutRepository provides data access for payouts.
type PayoutRepository interface {
	// Create persists a new payout record.
	Create(ctx context.Context, payout *Payout) error

	// GetByOrderID returns the payout for a given order, or nil if not found.
	GetByOrderID(ctx context.Context, orderID string) (*Payout, error)

	// Update persists changes to a payout.
	Update(ctx context.Context, payout *Payout) error

	// ListByBakery returns paginated payouts for a bakery.
	ListByBakery(ctx context.Context, bakeryID string, params PaginationParams) ([]Payout, int, error)
}
