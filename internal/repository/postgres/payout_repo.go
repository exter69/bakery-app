package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// PayoutRepo implements domain.PayoutRepository backed by PostgreSQL.
type PayoutRepo struct {
	pool *pgxpool.Pool
}

// NewPayoutRepo creates a new PayoutRepo with the given connection pool.
func NewPayoutRepo(pool *pgxpool.Pool) *PayoutRepo {
	return &PayoutRepo{pool: pool}
}

// Create persists a new payout record.
func (r *PayoutRepo) Create(ctx context.Context, payout *domain.Payout) error {
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO payouts (order_id, bakery_id, amount, commission, stripe_transfer_id, status, created_at, transferred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		payout.OrderID, payout.BakeryID,
		payout.Amount, payout.Commission,
		nilIfEmpty(payout.StripeTransferID),
		string(payout.Status),
		payout.CreatedAt, payout.TransferredAt,
	).Scan(&id)
	if err != nil {
		return err
	}
	payout.ID = id
	return nil
}

// GetByOrderID returns the payout for a given order, or nil if not found.
func (r *PayoutRepo) GetByOrderID(ctx context.Context, orderID string) (*domain.Payout, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, order_id, bakery_id, amount, commission, stripe_transfer_id, status, created_at, transferred_at
		FROM payouts
		WHERE order_id = $1`, orderID)

	return r.scanPayout(row)
}

// Update persists changes to a payout.
func (r *PayoutRepo) Update(ctx context.Context, payout *domain.Payout) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payouts
		SET stripe_transfer_id = $2, status = $3, transferred_at = $4
		WHERE id = $1`,
		payout.ID, nilIfEmpty(payout.StripeTransferID),
		string(payout.Status), payout.TransferredAt,
	)
	return err
}

// ListByBakery returns paginated payouts for a bakery, newest first.
func (r *PayoutRepo) ListByBakery(ctx context.Context, bakeryID string, params domain.PaginationParams) ([]domain.Payout, int, error) {
	// Count total
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM payouts WHERE bakery_id = $1`, bakeryID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []domain.Payout{}, 0, nil
	}

	offset := (params.Page - 1) * params.PageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, bakery_id, amount, commission, stripe_transfer_id, status, created_at, transferred_at
		FROM payouts
		WHERE bakery_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		bakeryID, params.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payouts []domain.Payout
	for rows.Next() {
		p, err := r.scanPayoutFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		payouts = append(payouts, *p)
	}

	return payouts, total, nil
}

func (r *PayoutRepo) scanPayout(row pgx.Row) (*domain.Payout, error) {
	var p domain.Payout
	var status string
	var stripeTransferID *string
	var transferredAt *time.Time

	err := row.Scan(
		&p.ID, &p.OrderID, &p.BakeryID,
		&p.Amount, &p.Commission,
		&stripeTransferID, &status,
		&p.CreatedAt, &transferredAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	p.Status = domain.PayoutStatus(status)
	if stripeTransferID != nil {
		p.StripeTransferID = *stripeTransferID
	}
	p.TransferredAt = transferredAt

	return &p, nil
}

func (r *PayoutRepo) scanPayoutFromRows(rows pgx.Rows) (*domain.Payout, error) {
	var p domain.Payout
	var status string
	var stripeTransferID *string
	var transferredAt *time.Time

	err := rows.Scan(
		&p.ID, &p.OrderID, &p.BakeryID,
		&p.Amount, &p.Commission,
		&stripeTransferID, &status,
		&p.CreatedAt, &transferredAt,
	)
	if err != nil {
		return nil, err
	}

	p.Status = domain.PayoutStatus(status)
	if stripeTransferID != nil {
		p.StripeTransferID = *stripeTransferID
	}
	p.TransferredAt = transferredAt

	return &p, nil
}
