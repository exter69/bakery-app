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

// ReservationRepo implements domain.ReservationRepository backed by PostgreSQL.
type ReservationRepo struct {
	pool *pgxpool.Pool
}

// NewReservationRepo creates a new ReservationRepo with the given connection pool.
func NewReservationRepo(pool *pgxpool.Pool) *ReservationRepo {
	return &ReservationRepo{pool: pool}
}

func (r *ReservationRepo) Save(ctx context.Context, reservation domain.Reservation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	dayInt, ok := dayWeekToInt[reservation.ScheduledDay]
	if !ok {
		return fmt.Errorf("invalid scheduled day: %s", reservation.ScheduledDay)
	}

	startTime := timeOfDayToTime(reservation.ScheduledTime.StartTime)
	endTime := timeOfDayToTime(reservation.ScheduledTime.EndTime)

	_, err = tx.Exec(ctx, `
		INSERT INTO reservations (id, bakery_id, user_id, scheduled_day, scheduled_start_time, scheduled_end_time, status, total_amount, payment_method, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			total_amount = EXCLUDED.total_amount`,
		reservation.ID, reservation.BakeryID, reservation.UserID,
		dayInt, startTime, endTime,
		string(reservation.Status), centsToDecimal(reservation.TotalAmount),
		string(reservation.PaymentMethod), reservation.CreatedAt)
	if err != nil {
		return err
	}

	// Replace items: delete existing, then insert fresh
	_, err = tx.Exec(ctx, `DELETE FROM order_items WHERE reservation_id = $1`, reservation.ID)
	if err != nil {
		return err
	}

	for _, item := range reservation.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (reservation_id, product_id, product_name, quantity, unit_price, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			reservation.ID, item.ProductID, item.ProductName, item.Quantity,
			centsToDecimal(item.UnitPrice), centsToDecimal(item.Subtotal))
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ReservationRepo) Get(ctx context.Context, id string) (*domain.Reservation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, user_id, scheduled_day, scheduled_start_time, scheduled_end_time,
		       status, total_amount, payment_method, created_at
		FROM reservations WHERE id = $1`, id)

	res, err := scanReservationRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.loadReservationItems(ctx, res.ID)
	if err != nil {
		return nil, err
	}
	res.Items = items

	return res, nil
}

func (r *ReservationRepo) ListByUser(ctx context.Context, userID string, filters domain.ReservationFilters, params domain.PaginationParams) ([]domain.Reservation, int, error) {
	return r.listReservations(ctx, "user_id", userID, filters, params)
}

func (r *ReservationRepo) ListByBakery(ctx context.Context, bakeryID string, filters domain.ReservationFilters, params domain.PaginationParams) ([]domain.Reservation, int, error) {
	return r.listReservations(ctx, "bakery_id", bakeryID, filters, params)
}

func (r *ReservationRepo) listReservations(ctx context.Context, filterCol, filterVal string, filters domain.ReservationFilters, params domain.PaginationParams) ([]domain.Reservation, int, error) {
	limit, offset := paginationToOffset(params)

	// Count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM reservations WHERE %s = $1`, filterCol)
	args := []any{filterVal}
	argIdx := 2
	countQuery, args, argIdx = buildWhereStatus(countQuery, args, argIdx, filters.Status)

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Data
	sortCol := orderSortColumn(filters.SortBy)
	sortDir := sortDirection(filters.SortDir)
	dataQuery := fmt.Sprintf(`
		SELECT id, bakery_id, user_id, scheduled_day, scheduled_start_time, scheduled_end_time,
		       status, total_amount, payment_method, created_at
		FROM reservations WHERE %s = $1`, filterCol)

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

	var reservations []domain.Reservation
	for rows.Next() {
		res, err := scanReservationRows(rows)
		if err != nil {
			return nil, 0, err
		}
		reservations = append(reservations, *res)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Load items
	for i := range reservations {
		items, err := r.loadReservationItems(ctx, reservations[i].ID)
		if err != nil {
			return nil, 0, err
		}
		reservations[i].Items = items
	}

	return reservations, total, nil
}

func (r *ReservationRepo) loadReservationItems(ctx context.Context, reservationID string) ([]domain.OrderItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT product_id, product_name, quantity, unit_price, subtotal
		FROM order_items WHERE reservation_id = $1`, reservationID)
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

// scanReservationRow scans a single reservation from a pgx.Row.
func scanReservationRow(row pgx.Row) (*domain.Reservation, error) {
	var res domain.Reservation
	var dayInt int
	var startTime, endTime time.Time
	var totalDec float64
	var status, paymentMethod string

	err := row.Scan(
		&res.ID, &res.BakeryID, &res.UserID, &dayInt, &startTime, &endTime,
		&status, &totalDec, &paymentMethod, &res.CreatedAt)
	if err != nil {
		return nil, err
	}

	res.ScheduledDay = dayIntToWeek[dayInt]
	res.ScheduledTime = domain.TimeSlot{
		StartTime: timeToTimeOfDay(startTime),
		EndTime:   timeToTimeOfDay(endTime),
	}
	res.Status = domain.ReservationStatus(status)
	res.TotalAmount = decimalToCents(totalDec)
	res.PaymentMethod = domain.PaymentMethod(paymentMethod)

	return &res, nil
}

// scanReservationRows scans a reservation from a pgx.Rows iterator.
func scanReservationRows(rows pgx.Rows) (*domain.Reservation, error) {
	var res domain.Reservation
	var dayInt int
	var startTime, endTime time.Time
	var totalDec float64
	var status, paymentMethod string

	err := rows.Scan(
		&res.ID, &res.BakeryID, &res.UserID, &dayInt, &startTime, &endTime,
		&status, &totalDec, &paymentMethod, &res.CreatedAt)
	if err != nil {
		return nil, err
	}

	res.ScheduledDay = dayIntToWeek[dayInt]
	res.ScheduledTime = domain.TimeSlot{
		StartTime: timeToTimeOfDay(startTime),
		EndTime:   timeToTimeOfDay(endTime),
	}
	res.Status = domain.ReservationStatus(status)
	res.TotalAmount = decimalToCents(totalDec)
	res.PaymentMethod = domain.PaymentMethod(paymentMethod)

	return &res, nil
}
