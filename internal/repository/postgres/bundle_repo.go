package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// BundleRepo implements domain.BundleRepository backed by PostgreSQL.
type BundleRepo struct {
	pool *pgxpool.Pool
}

// NewBundleRepo creates a new BundleRepo with the given connection pool.
func NewBundleRepo(pool *pgxpool.Pool) *BundleRepo {
	return &BundleRepo{pool: pool}
}

func (r *BundleRepo) CreateBundle(ctx context.Context, bundle *domain.SurplusBundle) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	bundle.CreatedAt = now
	bundle.UpdatedAt = now

	pickupStart := timeOfDayToTime(bundle.PickupStartTime)
	pickupEnd := timeOfDayToTime(bundle.PickupEndTime)

	// published_date: store as *string (nullable DATE)
	var publishedDate *string
	if bundle.PublishedDate != "" {
		publishedDate = &bundle.PublishedDate
	}

	// expires_at: store as *time.Time (nullable TIMESTAMP)
	var expiresAt *time.Time
	if !bundle.ExpiresAt.IsZero() {
		expiresAt = &bundle.ExpiresAt
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO surplus_bundles (
			id, bakery_id, name, type, photo_url, description, estimated_value,
			original_price, discounted_price, quantity_total, quantity_remaining,
			pickup_start_time, pickup_end_time, published_date, expires_at,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		bundle.ID, bundle.BakeryID, bundle.Name, string(bundle.Type),
		nilIfEmpty(bundle.PhotoURL), nilIfEmpty(bundle.Description), nilInt64(bundle.EstimatedValue),
		bundle.OriginalPrice, bundle.DiscountedPrice,
		bundle.QuantityTotal, bundle.QuantityRemaining,
		pickupStart, pickupEnd, publishedDate, expiresAt,
		string(bundle.Status), bundle.CreatedAt, bundle.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert surplus_bundles: %w", err)
	}

	// Bulk insert items
	for _, item := range bundle.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO surplus_bundle_items (id, bundle_id, product_id, description, quantity)
			VALUES ($1, $2, $3, $4, $5)`,
			item.ID, bundle.ID, nilIfEmpty(item.ProductID), item.Description, item.Quantity,
		)
		if err != nil {
			return fmt.Errorf("insert surplus_bundle_items: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *BundleRepo) UpdateBundle(ctx context.Context, bundle *domain.SurplusBundle) error {
	bundle.UpdatedAt = time.Now().UTC()

	var publishedDate *string
	if bundle.PublishedDate != "" {
		publishedDate = &bundle.PublishedDate
	}

	var expiresAt *time.Time
	if !bundle.ExpiresAt.IsZero() {
		expiresAt = &bundle.ExpiresAt
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE surplus_bundles
		SET status = $2, published_date = $3, expires_at = $4,
		    quantity_remaining = $5, updated_at = $6
		WHERE id = $1`,
		bundle.ID, string(bundle.Status), publishedDate, expiresAt,
		bundle.QuantityRemaining, bundle.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update surplus_bundles: %w", err)
	}
	return nil
}

func (r *BundleRepo) GetByID(ctx context.Context, id string) (*domain.SurplusBundle, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, name, type, photo_url, description, estimated_value,
		       original_price, discounted_price, quantity_total, quantity_remaining,
		       pickup_start_time, pickup_end_time, published_date, expires_at,
		       status, created_at, updated_at
		FROM surplus_bundles WHERE id = $1`, id)

	bundle, err := scanBundle(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get bundle by id: %w", err)
	}

	items, err := r.loadBundleItems(ctx, bundle.ID)
	if err != nil {
		return nil, err
	}
	bundle.Items = items

	return bundle, nil
}

func (r *BundleRepo) ListPublished(ctx context.Context, filters domain.BundleFilters, params domain.PaginationParams) ([]domain.SurplusBundle, int, error) {
	limit, offset := paginationToOffset(params)

	// Build count query
	countQuery := `SELECT COUNT(*) FROM surplus_bundles WHERE status = 'published'`
	args := []any{}
	argIdx := 1

	if filters.Type != nil {
		countQuery += fmt.Sprintf(` AND type = $%d`, argIdx)
		args = append(args, string(*filters.Type))
		argIdx++
	}
	if filters.PickupBefore != nil {
		pickupBefore := timeOfDayToTime(*filters.PickupBefore)
		countQuery += fmt.Sprintf(` AND pickup_end_time <= $%d`, argIdx)
		args = append(args, pickupBefore)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count published bundles: %w", err)
	}

	// Build data query
	dataQuery := `
		SELECT id, bakery_id, name, type, photo_url, description, estimated_value,
		       original_price, discounted_price, quantity_total, quantity_remaining,
		       pickup_start_time, pickup_end_time, published_date, expires_at,
		       status, created_at, updated_at
		FROM surplus_bundles WHERE status = 'published'`

	dataArgs := []any{}
	dataArgIdx := 1

	if filters.Type != nil {
		dataQuery += fmt.Sprintf(` AND type = $%d`, dataArgIdx)
		dataArgs = append(dataArgs, string(*filters.Type))
		dataArgIdx++
	}
	if filters.PickupBefore != nil {
		pickupBefore := timeOfDayToTime(*filters.PickupBefore)
		dataQuery += fmt.Sprintf(` AND pickup_end_time <= $%d`, dataArgIdx)
		dataArgs = append(dataArgs, pickupBefore)
		dataArgIdx++
	}

	dataQuery += fmt.Sprintf(` ORDER BY published_date DESC LIMIT $%d OFFSET $%d`, dataArgIdx, dataArgIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list published bundles: %w", err)
	}
	defer rows.Close()

	var bundles []domain.SurplusBundle
	for rows.Next() {
		bundle, err := scanBundleRows(rows)
		if err != nil {
			return nil, 0, err
		}
		bundles = append(bundles, *bundle)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Load items for each bundle
	for i := range bundles {
		items, err := r.loadBundleItems(ctx, bundles[i].ID)
		if err != nil {
			return nil, 0, err
		}
		bundles[i].Items = items
	}

	return bundles, total, nil
}

func (r *BundleRepo) GetExpiredBundles(ctx context.Context) ([]domain.SurplusBundle, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, bakery_id, name, type, photo_url, description, estimated_value,
		       original_price, discounted_price, quantity_total, quantity_remaining,
		       pickup_start_time, pickup_end_time, published_date, expires_at,
		       status, created_at, updated_at
		FROM surplus_bundles
		WHERE status = 'published' AND expires_at < NOW()`)
	if err != nil {
		return nil, fmt.Errorf("get expired bundles: %w", err)
	}
	defer rows.Close()

	var bundles []domain.SurplusBundle
	for rows.Next() {
		bundle, err := scanBundleRows(rows)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, *bundle)
	}
	return bundles, rows.Err()
}

func (r *BundleRepo) CreateReservation(ctx context.Context, reservation *domain.BundleReservation) error {
	now := time.Now().UTC()
	reservation.CreatedAt = now
	reservation.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO bundle_reservations (id, bundle_id, user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		reservation.ID, reservation.BundleID, reservation.UserID,
		string(reservation.Status), reservation.CreatedAt, reservation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert bundle_reservations: %w", err)
	}
	return nil
}

func (r *BundleRepo) GetReservation(ctx context.Context, id string) (*domain.BundleReservation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bundle_id, user_id, status, created_at, updated_at
		FROM bundle_reservations WHERE id = $1`, id)

	res, err := scanBundleReservation(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get bundle reservation: %w", err)
	}
	return res, nil
}

func (r *BundleRepo) GetActiveReservation(ctx context.Context, userID string, bundleID string) (*domain.BundleReservation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bundle_id, user_id, status, created_at, updated_at
		FROM bundle_reservations
		WHERE user_id = $1 AND bundle_id = $2 AND status IN ('pending', 'confirmed')`, userID, bundleID)

	res, err := scanBundleReservation(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active bundle reservation: %w", err)
	}
	return res, nil
}

func (r *BundleRepo) UpdateReservation(ctx context.Context, reservation *domain.BundleReservation) error {
	reservation.UpdatedAt = time.Now().UTC()

	_, err := r.pool.Exec(ctx, `
		UPDATE bundle_reservations
		SET status = $2, updated_at = $3
		WHERE id = $1`,
		reservation.ID, string(reservation.Status), reservation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update bundle_reservations: %w", err)
	}
	return nil
}

func (r *BundleRepo) GetOverdueReservations(ctx context.Context) ([]domain.BundleReservation, error) {
	// Select pending/confirmed reservations where the bundle's pickup_end_time
	// has passed for today. We compare with current time-of-day.
	rows, err := r.pool.Query(ctx, `
		SELECT br.id, br.bundle_id, br.user_id, br.status, br.created_at, br.updated_at
		FROM bundle_reservations br
		JOIN surplus_bundles sb ON sb.id = br.bundle_id
		WHERE br.status IN ('pending', 'confirmed')
		  AND sb.status = 'published'
		  AND sb.pickup_end_time < CURRENT_TIME`)
	if err != nil {
		return nil, fmt.Errorf("get overdue reservations: %w", err)
	}
	defer rows.Close()

	var reservations []domain.BundleReservation
	for rows.Next() {
		var res domain.BundleReservation
		var status string
		if err := rows.Scan(&res.ID, &res.BundleID, &res.UserID, &status, &res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, err
		}
		res.Status = domain.BundleReservationStatus(status)
		reservations = append(reservations, res)
	}
	return reservations, rows.Err()
}

func (r *BundleRepo) CountPickedUpThisMonth(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM bundle_reservations
		WHERE status = 'picked_up'
		  AND created_at >= date_trunc('month', CURRENT_DATE)`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count picked up this month: %w", err)
	}
	return count, nil
}

func (r *BundleRepo) DecrementStock(ctx context.Context, bundleID string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE surplus_bundles
		SET quantity_remaining = quantity_remaining - 1, updated_at = NOW()
		WHERE id = $1 AND quantity_remaining > 0`, bundleID)
	if err != nil {
		return fmt.Errorf("decrement stock: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("stock exhausted for bundle %s", bundleID)
	}
	return nil
}

func (r *BundleRepo) IncrementStock(ctx context.Context, bundleID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE surplus_bundles
		SET quantity_remaining = quantity_remaining + 1, updated_at = NOW()
		WHERE id = $1`, bundleID)
	if err != nil {
		return fmt.Errorf("increment stock: %w", err)
	}
	return nil
}

// --- Helpers ---

// nilInt64 returns nil if v is 0, otherwise a pointer to v.
// Used for nullable BIGINT columns (estimated_value).
func nilInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

// scanBundle scans a single bundle from a pgx.Row.
func scanBundle(row pgx.Row) (*domain.SurplusBundle, error) {
	var b domain.SurplusBundle
	var bundleType, status string
	var photoURL, description *string
	var estimatedValue *int64
	var pickupStart, pickupEnd time.Time
	var publishedDate *string
	var expiresAt *time.Time

	err := row.Scan(
		&b.ID, &b.BakeryID, &b.Name, &bundleType, &photoURL, &description, &estimatedValue,
		&b.OriginalPrice, &b.DiscountedPrice, &b.QuantityTotal, &b.QuantityRemaining,
		&pickupStart, &pickupEnd, &publishedDate, &expiresAt,
		&status, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	b.Type = domain.BundleType(bundleType)
	b.Status = domain.BundleStatus(status)
	b.PickupStartTime = timeToTimeOfDay(pickupStart)
	b.PickupEndTime = timeToTimeOfDay(pickupEnd)

	if photoURL != nil {
		b.PhotoURL = *photoURL
	}
	if description != nil {
		b.Description = *description
	}
	if estimatedValue != nil {
		b.EstimatedValue = *estimatedValue
	}
	if publishedDate != nil {
		b.PublishedDate = *publishedDate
	}
	if expiresAt != nil {
		b.ExpiresAt = *expiresAt
	}

	return &b, nil
}

// scanBundleRows scans a single bundle from a pgx.Rows iterator.
func scanBundleRows(rows pgx.Rows) (*domain.SurplusBundle, error) {
	var b domain.SurplusBundle
	var bundleType, status string
	var photoURL, description *string
	var estimatedValue *int64
	var pickupStart, pickupEnd time.Time
	var publishedDate *string
	var expiresAt *time.Time

	err := rows.Scan(
		&b.ID, &b.BakeryID, &b.Name, &bundleType, &photoURL, &description, &estimatedValue,
		&b.OriginalPrice, &b.DiscountedPrice, &b.QuantityTotal, &b.QuantityRemaining,
		&pickupStart, &pickupEnd, &publishedDate, &expiresAt,
		&status, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	b.Type = domain.BundleType(bundleType)
	b.Status = domain.BundleStatus(status)
	b.PickupStartTime = timeToTimeOfDay(pickupStart)
	b.PickupEndTime = timeToTimeOfDay(pickupEnd)

	if photoURL != nil {
		b.PhotoURL = *photoURL
	}
	if description != nil {
		b.Description = *description
	}
	if estimatedValue != nil {
		b.EstimatedValue = *estimatedValue
	}
	if publishedDate != nil {
		b.PublishedDate = *publishedDate
	}
	if expiresAt != nil {
		b.ExpiresAt = *expiresAt
	}

	return &b, nil
}

// scanBundleReservation scans a single bundle reservation from a pgx.Row.
func scanBundleReservation(row pgx.Row) (*domain.BundleReservation, error) {
	var res domain.BundleReservation
	var status string

	err := row.Scan(&res.ID, &res.BundleID, &res.UserID, &status, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return nil, err
	}
	res.Status = domain.BundleReservationStatus(status)
	return &res, nil
}

// loadBundleItems loads all items for a given bundle.
func (r *BundleRepo) loadBundleItems(ctx context.Context, bundleID string) ([]domain.BundleItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, bundle_id, product_id, description, quantity
		FROM surplus_bundle_items WHERE bundle_id = $1`, bundleID)
	if err != nil {
		return nil, fmt.Errorf("load bundle items: %w", err)
	}
	defer rows.Close()

	var items []domain.BundleItem
	for rows.Next() {
		var item domain.BundleItem
		var productID *string
		if err := rows.Scan(&item.ID, &item.BundleID, &productID, &item.Description, &item.Quantity); err != nil {
			return nil, err
		}
		if productID != nil {
			item.ProductID = *productID
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
