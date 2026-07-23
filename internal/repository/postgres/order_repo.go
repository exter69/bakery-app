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

// OrderRepo implements domain.OrderRepository backed by PostgreSQL.
type OrderRepo struct {
	pool *pgxpool.Pool
}

// NewOrderRepo creates a new OrderRepo with the given connection pool.
func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	dayInt, ok := dayWeekToInt[order.ScheduledDay]
	if !ok {
		return fmt.Errorf("invalid scheduled day: %s", order.ScheduledDay)
	}

	startTime := timeOfDayToTime(order.ScheduledTime.StartTime)
	endTime := timeOfDayToTime(order.ScheduledTime.EndTime)

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, bakery_id, user_id, scheduled_day, scheduled_start_time, scheduled_end_time, status, total_amount, payment_method, payment_intent_id, refund_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			total_amount = EXCLUDED.total_amount,
			payment_intent_id = EXCLUDED.payment_intent_id,
			refund_status = EXCLUDED.refund_status,
			updated_at = EXCLUDED.updated_at`,
		order.ID, order.BakeryID, order.UserID,
		dayInt, startTime, endTime,
		string(order.Status), centsToDecimal(order.TotalAmount),
		string(order.PaymentMethod), nilIfEmpty(order.PaymentIntentID),
		order.RefundStatus,
		order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return err
	}

	// Replace items: delete existing, then insert fresh
	_, err = tx.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			order.ID, item.ProductID, item.ProductName, item.Quantity,
			centsToDecimal(item.UnitPrice), centsToDecimal(item.Subtotal))
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *OrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, user_id, scheduled_day, scheduled_start_time, scheduled_end_time,
		       status, total_amount, payment_method, payment_intent_id, refund_status, created_at, updated_at
		FROM orders WHERE id = $1`, id)

	o, err := scanOrderRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.loadOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return o, nil
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID string, filters domain.OrderFilters, params domain.PaginationParams) ([]domain.Order, int, error) {
	return r.listOrders(ctx, "user_id", userID, filters, params)
}

func (r *OrderRepo) ListByBakery(ctx context.Context, bakeryID string, filters domain.OrderFilters, params domain.PaginationParams) ([]domain.Order, int, error) {
	return r.listOrders(ctx, "bakery_id", bakeryID, filters, params)
}

func (r *OrderRepo) listOrders(ctx context.Context, filterCol, filterVal string, filters domain.OrderFilters, params domain.PaginationParams) ([]domain.Order, int, error) {
	limit, offset := paginationToOffset(params)

	// Build count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM orders WHERE %s = $1`, filterCol)
	args := []any{filterVal}
	argIdx := 2
	countQuery, args, argIdx = buildWhereStatus(countQuery, args, argIdx, filters.Status)

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Build data query
	sortCol := orderSortColumn(filters.SortBy)
	sortDir := sortDirection(filters.SortDir)
	dataQuery := fmt.Sprintf(`
		SELECT id, bakery_id, user_id, scheduled_day, scheduled_start_time, scheduled_end_time,
		       status, total_amount, payment_method, payment_intent_id, refund_status, created_at, updated_at
		FROM orders WHERE %s = $1`, filterCol)

	dataArgs := []any{filterVal}
	dataArgIdx := 2
	dataQuery, dataArgs, dataArgIdx = buildWhereStatus(dataQuery, dataArgs, dataArgIdx, filters.Status)
	dataQuery += fmt.Sprintf(` ORDER BY %s %s LIMIT $%d OFFSET $%d`, sortCol, sortDir, dataArgIdx, dataArgIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		o, err := scanOrderRows(rows)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, *o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Load items for each order
	for i := range orders {
		items, err := r.loadOrderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, 0, err
		}
		orders[i].Items = items
	}

	return orders, total, nil
}

func (r *OrderRepo) loadOrderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT product_id, product_name, quantity, unit_price, subtotal
		FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		var unitPriceDec, subtotalDec float64
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Quantity, &unitPriceDec, &subtotalDec); err != nil {
			return nil, err
		}
		item.UnitPrice = decimalToCents(unitPriceDec)
		item.Subtotal = decimalToCents(subtotalDec)
		items = append(items, item)
	}
	return items, rows.Err()
}

// scanOrderRow scans a single order from a pgx.Row.
func scanOrderRow(row pgx.Row) (*domain.Order, error) {
	var o domain.Order
	var dayInt int
	var startTime, endTime time.Time
	var totalDec float64
	var status, paymentMethod string
	var paymentIntentID *string
	var refundStatus string

	err := row.Scan(
		&o.ID, &o.BakeryID, &o.UserID, &dayInt, &startTime, &endTime,
		&status, &totalDec, &paymentMethod, &paymentIntentID, &refundStatus, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}

	o.ScheduledDay = dayIntToWeek[dayInt]
	o.ScheduledTime = domain.TimeSlot{
		StartTime: timeToTimeOfDay(startTime),
		EndTime:   timeToTimeOfDay(endTime),
	}
	o.Status = domain.OrderStatus(status)
	o.TotalAmount = decimalToCents(totalDec)
	o.PaymentMethod = domain.PaymentMethod(paymentMethod)
	if paymentIntentID != nil {
		o.PaymentIntentID = *paymentIntentID
	}
	o.RefundStatus = refundStatus

	return &o, nil
}

// scanOrderRows scans an order from a pgx.Rows iterator.
func scanOrderRows(rows pgx.Rows) (*domain.Order, error) {
	var o domain.Order
	var dayInt int
	var startTime, endTime time.Time
	var totalDec float64
	var status, paymentMethod string
	var paymentIntentID *string
	var refundStatus string

	err := rows.Scan(
		&o.ID, &o.BakeryID, &o.UserID, &dayInt, &startTime, &endTime,
		&status, &totalDec, &paymentMethod, &paymentIntentID, &refundStatus, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}

	o.ScheduledDay = dayIntToWeek[dayInt]
	o.ScheduledTime = domain.TimeSlot{
		StartTime: timeToTimeOfDay(startTime),
		EndTime:   timeToTimeOfDay(endTime),
	}
	o.Status = domain.OrderStatus(status)
	o.TotalAmount = decimalToCents(totalDec)
	o.PaymentMethod = domain.PaymentMethod(paymentMethod)
	if paymentIntentID != nil {
		o.PaymentIntentID = *paymentIntentID
	}
	o.RefundStatus = refundStatus

	return &o, nil
}
