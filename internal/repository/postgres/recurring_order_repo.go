package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// RecurringOrderRepo implements domain.RecurringOrderRepository backed by PostgreSQL.
type RecurringOrderRepo struct {
	pool *pgxpool.Pool
}

// NewRecurringOrderRepo creates a new RecurringOrderRepo with the given connection pool.
func NewRecurringOrderRepo(pool *pgxpool.Pool) *RecurringOrderRepo {
	return &RecurringOrderRepo{pool: pool}
}

func (r *RecurringOrderRepo) Save(ctx context.Context, order *domain.RecurringOrder) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}

	timeJSON, err := json.Marshal(order.ScheduledTime)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO recurring_orders (id, user_id, bakery_id, items, scheduled_day, scheduled_time, frequency, selection_mode, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			items = EXCLUDED.items,
			scheduled_day = EXCLUDED.scheduled_day,
			scheduled_time = EXCLUDED.scheduled_time,
			frequency = EXCLUDED.frequency,
			selection_mode = EXCLUDED.selection_mode,
			active = EXCLUDED.active,
			updated_at = EXCLUDED.updated_at`,
		order.ID, order.UserID, order.BakeryID,
		itemsJSON, string(order.ScheduledDay), timeJSON,
		string(order.Frequency), string(order.SelectionMode),
		order.Active, order.CreatedAt, order.UpdatedAt)
	return err
}

func (r *RecurringOrderRepo) GetByID(ctx context.Context, id string) (*domain.RecurringOrder, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, bakery_id, items, scheduled_day, scheduled_time, frequency, selection_mode, active, created_at, updated_at
		FROM recurring_orders WHERE id = $1`, id)

	order, err := scanRecurringOrder(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *RecurringOrderRepo) ListByUser(ctx context.Context, userID string, params domain.PaginationParams) ([]domain.RecurringOrder, int, error) {
	limit, offset := paginationToOffset(params)

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM recurring_orders WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, bakery_id, items, scheduled_day, scheduled_time, frequency, selection_mode, active, created_at, updated_at
		FROM recurring_orders WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []domain.RecurringOrder
	for rows.Next() {
		order, err := scanRecurringOrderRows(rows)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, *order)
	}
	return orders, total, rows.Err()
}

func (r *RecurringOrderRepo) ListActive(ctx context.Context) ([]domain.RecurringOrder, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, bakery_id, items, scheduled_day, scheduled_time, frequency, selection_mode, active, created_at, updated_at
		FROM recurring_orders WHERE active = true
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.RecurringOrder
	for rows.Next() {
		order, err := scanRecurringOrderRows(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}
	return orders, rows.Err()
}

func (r *RecurringOrderRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM recurring_orders WHERE id = $1`, id)
	return err
}

// scanRecurringOrder scans a recurring order from a pgx.Row.
func scanRecurringOrder(row pgx.Row) (*domain.RecurringOrder, error) {
	var o domain.RecurringOrder
	var itemsJSON, timeJSON []byte
	var scheduledDay, frequency, selectionMode string

	err := row.Scan(
		&o.ID, &o.UserID, &o.BakeryID, &itemsJSON, &scheduledDay,
		&timeJSON, &frequency, &selectionMode, &o.Active, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(timeJSON, &o.ScheduledTime); err != nil {
		return nil, err
	}

	o.ScheduledDay = domain.DayOfWeek(scheduledDay)
	o.Frequency = domain.RecurringFrequency(frequency)
	o.SelectionMode = domain.SelectionMode(selectionMode)

	return &o, nil
}

// scanRecurringOrderRows scans a recurring order from a pgx.Rows iterator.
func scanRecurringOrderRows(rows pgx.Rows) (*domain.RecurringOrder, error) {
	var o domain.RecurringOrder
	var itemsJSON, timeJSON []byte
	var scheduledDay, frequency, selectionMode string

	err := rows.Scan(
		&o.ID, &o.UserID, &o.BakeryID, &itemsJSON, &scheduledDay,
		&timeJSON, &frequency, &selectionMode, &o.Active, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(timeJSON, &o.ScheduledTime); err != nil {
		return nil, err
	}

	o.ScheduledDay = domain.DayOfWeek(scheduledDay)
	o.Frequency = domain.RecurringFrequency(frequency)
	o.SelectionMode = domain.SelectionMode(selectionMode)

	return &o, nil
}
