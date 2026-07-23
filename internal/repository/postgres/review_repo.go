package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// ReviewRepo implements domain.ReviewRepository backed by PostgreSQL.
type ReviewRepo struct {
	pool *pgxpool.Pool
}

// NewReviewRepo creates a new ReviewRepo with the given connection pool.
func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) Create(ctx context.Context, review *domain.Review) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO reviews (id, bakery_id, user_id, order_id, rating, text, hidden, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		review.ID, review.BakeryID, review.UserID, review.OrderID,
		review.Rating, nilIfEmpty(review.Text), review.Hidden, review.CreatedAt)
	if err != nil {
		return err
	}

	// Recalculate bakery rating aggregates
	_, err = tx.Exec(ctx, `
		UPDATE bakeries SET
			rating_avg = (SELECT AVG(rating)::NUMERIC(2,1) FROM reviews WHERE bakery_id = $1 AND hidden = FALSE),
			rating_count = (SELECT COUNT(*) FROM reviews WHERE bakery_id = $1 AND hidden = FALSE)
		WHERE id = $1`, review.BakeryID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ReviewRepo) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, user_id, order_id, rating, text, hidden, created_at
		FROM reviews WHERE id = $1`, id)

	review, err := scanReviewRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return review, nil
}

func (r *ReviewRepo) GetByUserAndBakery(ctx context.Context, userID, bakeryID string) (*domain.Review, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, user_id, order_id, rating, text, hidden, created_at
		FROM reviews WHERE user_id = $1 AND bakery_id = $2`, userID, bakeryID)

	review, err := scanReviewRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return review, nil
}

func (r *ReviewRepo) ListByBakery(ctx context.Context, bakeryID string, params domain.PaginationParams) ([]domain.Review, int, error) {
	limit, offset := paginationToOffset(params)

	// Count visible reviews
	var total int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM reviews WHERE bakery_id = $1 AND hidden = FALSE`, bakeryID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch page of visible reviews, newest first
	rows, err := r.pool.Query(ctx, `
		SELECT id, bakery_id, user_id, order_id, rating, text, hidden, created_at
		FROM reviews
		WHERE bakery_id = $1 AND hidden = FALSE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, bakeryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var rev domain.Review
		var text *string
		err := rows.Scan(&rev.ID, &rev.BakeryID, &rev.UserID, &rev.OrderID,
			&rev.Rating, &text, &rev.Hidden, &rev.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		if text != nil {
			rev.Text = *text
		}
		reviews = append(reviews, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

func (r *ReviewRepo) SetHidden(ctx context.Context, reviewID string, hidden bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get the bakery_id for aggregate recalculation
	var bakeryID string
	err = tx.QueryRow(ctx, `
		UPDATE reviews SET hidden = $2 WHERE id = $1 RETURNING bakery_id`,
		reviewID, hidden).Scan(&bakeryID)
	if err != nil {
		return err
	}

	// Recalculate bakery rating aggregates
	_, err = tx.Exec(ctx, `
		UPDATE bakeries SET
			rating_avg = (SELECT AVG(rating)::NUMERIC(2,1) FROM reviews WHERE bakery_id = $1 AND hidden = FALSE),
			rating_count = (SELECT COUNT(*) FROM reviews WHERE bakery_id = $1 AND hidden = FALSE)
		WHERE id = $1`, bakeryID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ReviewRepo) CreateReport(ctx context.Context, report *domain.ReviewReport) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO review_reports (id, review_id, reporter_id, reason, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		report.ID, report.ReviewID, report.ReporterID, nilIfEmpty(report.Reason), report.CreatedAt)
	return err
}

// scanReviewRow scans a single review from a pgx.Row.
func scanReviewRow(row pgx.Row) (*domain.Review, error) {
	var rev domain.Review
	var text *string
	err := row.Scan(&rev.ID, &rev.BakeryID, &rev.UserID, &rev.OrderID,
		&rev.Rating, &text, &rev.Hidden, &rev.CreatedAt)
	if err != nil {
		return nil, err
	}
	if text != nil {
		rev.Text = *text
	}
	return &rev, nil
}
