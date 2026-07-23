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

// BakeryRepo implements domain.BakeryRepository backed by PostgreSQL.
type BakeryRepo struct {
	pool *pgxpool.Pool
}

// NewBakeryRepo creates a new BakeryRepo with the given connection pool.
func NewBakeryRepo(pool *pgxpool.Pool) *BakeryRepo {
	return &BakeryRepo{pool: pool}
}

func (r *BakeryRepo) ListBakeries(ctx context.Context, params domain.PaginationParams) ([]domain.Bakery, int, error) {
	limit, offset := paginationToOffset(params)

	// Count total
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM bakeries`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch bakeries page
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_id, name, photo_url, description, address, latitude, longitude, rating_avg, rating_count, created_at
		FROM bakeries
		ORDER BY name ASC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bakeries []domain.Bakery
	for rows.Next() {
		b, err := scanBakeryRow(rows)
		if err != nil {
			return nil, 0, err
		}
		bakeries = append(bakeries, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Load schedules for all fetched bakeries
	for i := range bakeries {
		schedules, err := r.loadSchedules(ctx, bakeries[i].ID)
		if err != nil {
			return nil, 0, err
		}
		bakeries[i].Schedule = schedules
	}

	return bakeries, total, nil
}

func (r *BakeryRepo) GetBakery(ctx context.Context, id string) (*domain.Bakery, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, photo_url, description, address, latitude, longitude, rating_avg, rating_count, created_at
		FROM bakeries WHERE id = $1`, id)

	b, err := scanBakeryQueryRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	schedules, err := r.loadSchedules(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	b.Schedule = schedules

	return b, nil
}

func (r *BakeryRepo) GetBakeryByOwner(ctx context.Context, ownerID string) (*domain.Bakery, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, photo_url, description, address, latitude, longitude, rating_avg, rating_count, created_at
		FROM bakeries WHERE owner_id = $1`, ownerID)

	b, err := scanBakeryQueryRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	schedules, err := r.loadSchedules(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	b.Schedule = schedules

	return b, nil
}

func (r *BakeryRepo) GetByStripeConnectID(ctx context.Context, stripeConnectID string) (*domain.Bakery, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, photo_url, description, address, latitude, longitude, rating_avg, rating_count, created_at
		FROM bakeries WHERE stripe_connect_id = $1`, stripeConnectID)

	b, err := scanBakeryQueryRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	schedules, err := r.loadSchedules(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	b.Schedule = schedules

	return b, nil
}

func (r *BakeryRepo) UpdateBakery(ctx context.Context, bakery *domain.Bakery) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE bakeries SET
			owner_id = $2, name = $3, photo_url = $4, description = $5,
			address = $6, latitude = $7, longitude = $8
		WHERE id = $1`,
		bakery.ID, bakery.OwnerID, bakery.Name, bakery.PhotoURL, bakery.Description,
		bakery.Address, bakery.Latitude, bakery.Longitude)
	if err != nil {
		return err
	}

	// Replace schedules: delete all, then re-insert
	_, err = tx.Exec(ctx, `DELETE FROM day_schedules WHERE bakery_id = $1`, bakery.ID)
	if err != nil {
		return err
	}

	for _, s := range bakery.Schedule {
		dayInt, ok := dayWeekToInt[s.Day]
		if !ok {
			continue
		}
		var openTime, closeTime *time.Time
		if s.IsOpen {
			ot := timeOfDayToTime(s.OpenTime)
			ct := timeOfDayToTime(s.CloseTime)
			openTime = &ot
			closeTime = &ct
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO day_schedules (bakery_id, day_of_week, open_time, close_time, is_open)
			VALUES ($1, $2, $3, $4, $5)`,
			bakery.ID, dayInt, openTime, closeTime, s.IsOpen)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *BakeryRepo) GetProductsByBakery(ctx context.Context, bakeryID string) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, bakery_id, name, description, price, photo_url, category, is_available, allergens, health_score
		FROM products WHERE bakery_id = $1
		ORDER BY category, name`, bakeryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		p, err := scanProductRow(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *BakeryRepo) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, name, description, price, photo_url, category, is_available, allergens, health_score
		FROM products WHERE id = $1`, id)

	var p domain.Product
	var priceDecimal float64
	var allergens []string
	var healthScore *int

	err := row.Scan(
		&p.ID, &p.BakeryID, &p.Name, &p.Description, &priceDecimal,
		&p.PhotoURL, &p.Category, &p.IsAvailable, &allergens, &healthScore)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	p.Price = decimalToCents(priceDecimal)
	p.HealthScore = healthScore
	if allergens != nil {
		p.Allergens = allergens
	} else {
		p.Allergens = []string{}
	}

	return &p, nil
}

func (r *BakeryRepo) CreateProduct(ctx context.Context, product *domain.Product) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO products (id, bakery_id, name, description, price, photo_url, category, is_available, allergens, health_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		product.ID, product.BakeryID, product.Name, product.Description,
		centsToDecimal(product.Price), product.PhotoURL, product.Category,
		product.IsAvailable, product.Allergens, product.HealthScore)
	return err
}

func (r *BakeryRepo) UpdateProduct(ctx context.Context, product *domain.Product) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE products SET
			name = $2, description = $3, price = $4, photo_url = $5,
			category = $6, is_available = $7, allergens = $8, health_score = $9
		WHERE id = $1`,
		product.ID, product.Name, product.Description,
		centsToDecimal(product.Price), product.PhotoURL, product.Category,
		product.IsAvailable, product.Allergens, product.HealthScore)
	return err
}

func (r *BakeryRepo) DeleteProduct(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

func (r *BakeryRepo) SearchProducts(ctx context.Context, params domain.ProductSearchParams) ([]domain.ProductSearchResult, int, error) {
	limit, offset := paginationToOffset(params.PaginationParams)

	// Build dynamic WHERE clause
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if params.Query != "" {
		where += fmt.Sprintf(" AND p.name ILIKE '%%' || $%d || '%%'", argIdx)
		args = append(args, params.Query)
		argIdx++
	}

	if params.Category != "" {
		where += fmt.Sprintf(" AND p.category = $%d", argIdx)
		args = append(args, params.Category)
		argIdx++
	}

	if len(params.ExcludeAllergens) > 0 {
		where += fmt.Sprintf(" AND NOT (p.allergens && $%d::text[])", argIdx)
		args = append(args, params.ExcludeAllergens)
		argIdx++
	}

	if params.MinHealthScore != nil {
		where += fmt.Sprintf(" AND p.health_score >= $%d", argIdx)
		args = append(args, *params.MinHealthScore)
		argIdx++
	}

	if params.OnlyAvailable {
		where += " AND p.is_available = true"
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products p JOIN bakeries b ON b.id = p.bakery_id %s`, where)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch results
	selectQuery := fmt.Sprintf(`
		SELECT p.id, p.bakery_id, p.name, p.description, p.price, p.photo_url,
		       p.category, p.is_available, p.allergens, p.health_score,
		       b.id, b.name
		FROM products p
		JOIN bakeries b ON b.id = p.bakery_id
		%s
		ORDER BY p.name ASC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []domain.ProductSearchResult
	for rows.Next() {
		var p domain.Product
		var priceDecimal float64
		var allergens []string
		var healthScore *int
		var bakeryID, bakeryName string

		err := rows.Scan(
			&p.ID, &p.BakeryID, &p.Name, &p.Description, &priceDecimal,
			&p.PhotoURL, &p.Category, &p.IsAvailable, &allergens, &healthScore,
			&bakeryID, &bakeryName)
		if err != nil {
			return nil, 0, err
		}

		p.Price = decimalToCents(priceDecimal)
		p.HealthScore = healthScore
		if allergens != nil {
			p.Allergens = allergens
		} else {
			p.Allergens = []string{}
		}

		results = append(results, domain.ProductSearchResult{
			Product:    p,
			BakeryID:   bakeryID,
			BakeryName: bakeryName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// loadSchedules fetches day_schedules for a bakery and maps to domain types.
func (r *BakeryRepo) loadSchedules(ctx context.Context, bakeryID string) ([]domain.DaySchedule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT day_of_week, open_time, close_time, is_open
		FROM day_schedules WHERE bakery_id = $1
		ORDER BY day_of_week`, bakeryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []domain.DaySchedule
	for rows.Next() {
		var dayInt int
		var openTime, closeTime *time.Time
		var isOpen bool

		if err := rows.Scan(&dayInt, &openTime, &closeTime, &isOpen); err != nil {
			return nil, err
		}

		day, ok := dayIntToWeek[dayInt]
		if !ok {
			continue
		}

		s := domain.DaySchedule{
			Day:   day,
			IsOpen: isOpen,
		}
		if openTime != nil {
			s.OpenTime = timeToTimeOfDay(*openTime)
		}
		if closeTime != nil {
			s.CloseTime = timeToTimeOfDay(*closeTime)
		}
		schedules = append(schedules, s)
	}

	return schedules, rows.Err()
}

// scanBakeryRow scans a bakery from a pgx.Rows iterator.
func scanBakeryRow(rows pgx.Rows) (domain.Bakery, error) {
	var b domain.Bakery
	var ownerID *string

	err := rows.Scan(
		&b.ID, &ownerID, &b.Name, &b.PhotoURL, &b.Description,
		&b.Address, &b.Latitude, &b.Longitude, &b.RatingAvg, &b.RatingCount, &b.CreatedAt)
	if err != nil {
		return b, err
	}

	if ownerID != nil {
		b.OwnerID = *ownerID
	}

	return b, nil
}

// scanBakeryQueryRow scans a bakery from a single pgx.Row.
func scanBakeryQueryRow(row pgx.Row) (*domain.Bakery, error) {
	var b domain.Bakery
	var ownerID *string

	err := row.Scan(
		&b.ID, &ownerID, &b.Name, &b.PhotoURL, &b.Description,
		&b.Address, &b.Latitude, &b.Longitude, &b.RatingAvg, &b.RatingCount, &b.CreatedAt)
	if err != nil {
		return nil, err
	}

	if ownerID != nil {
		b.OwnerID = *ownerID
	}

	return &b, nil
}

// scanProductRow scans a product from a pgx.Rows iterator.
func scanProductRow(rows pgx.Rows) (domain.Product, error) {
	var p domain.Product
	var priceDecimal float64
	var allergens []string
	var healthScore *int

	err := rows.Scan(
		&p.ID, &p.BakeryID, &p.Name, &p.Description, &priceDecimal,
		&p.PhotoURL, &p.Category, &p.IsAvailable, &allergens, &healthScore)
	if err != nil {
		return p, err
	}

	p.Price = decimalToCents(priceDecimal)
	p.HealthScore = healthScore
	if allergens != nil {
		p.Allergens = allergens
	} else {
		p.Allergens = []string{}
	}

	return p, nil
}
