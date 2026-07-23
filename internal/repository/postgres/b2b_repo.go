package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// B2BRepo implements domain.B2BRepository backed by PostgreSQL.
type B2BRepo struct {
	pool *pgxpool.Pool
}

// NewB2BRepo creates a new B2BRepo with the given connection pool.
func NewB2BRepo(pool *pgxpool.Pool) *B2BRepo {
	return &B2BRepo{pool: pool}
}

// --- Business Profiles ---

func (r *B2BRepo) CreateProfile(ctx context.Context, profile *domain.BusinessProfile) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO business_profiles (id, user_id, company_name, vat_siret, iban, billing_email, billing_contact_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		profile.ID, profile.UserID, profile.CompanyName, profile.VATSiret,
		profile.IBAN, profile.BillingEmail, profile.BillingContactName,
		profile.CreatedAt, profile.UpdatedAt)
	return err
}

func (r *B2BRepo) GetProfileByUserID(ctx context.Context, userID string) (*domain.BusinessProfile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, company_name, vat_siret, iban, billing_email, billing_contact_name, pro_discount, current_month_spend, spend_month, created_at, updated_at
		FROM business_profiles WHERE user_id = $1`, userID)
	return scanBusinessProfile(row)
}

func (r *B2BRepo) GetProfileByVAT(ctx context.Context, vatSiret string) (*domain.BusinessProfile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, company_name, vat_siret, iban, billing_email, billing_contact_name, pro_discount, current_month_spend, spend_month, created_at, updated_at
		FROM business_profiles WHERE vat_siret = $1`, vatSiret)
	return scanBusinessProfile(row)
}

func (r *B2BRepo) UpdateProfile(ctx context.Context, profile *domain.BusinessProfile) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE business_profiles SET
			company_name = $2, iban = $3, billing_email = $4, billing_contact_name = $5, updated_at = $6
		WHERE id = $1`,
		profile.ID, profile.CompanyName, profile.IBAN, profile.BillingEmail,
		profile.BillingContactName, profile.UpdatedAt)
	return err
}

func scanBusinessProfile(row pgx.Row) (*domain.BusinessProfile, error) {
	var p domain.BusinessProfile
	err := row.Scan(&p.ID, &p.UserID, &p.CompanyName, &p.VATSiret, &p.IBAN,
		&p.BillingEmail, &p.BillingContactName, &p.ProDiscount, &p.CurrentMonthSpend, &p.SpendMonth,
		&p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// --- Delivery Sites ---

func (r *B2BRepo) CreateSite(ctx context.Context, site *domain.DeliverySite) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO delivery_sites (id, user_id, name, street_address, city, postal_code, country, delivery_instructions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		site.ID, site.UserID, site.Name, site.StreetAddress, site.City,
		site.PostalCode, site.Country, nilIfEmpty(site.DeliveryInstructions),
		site.CreatedAt, site.UpdatedAt)
	return err
}

func (r *B2BRepo) GetSiteByID(ctx context.Context, id string) (*domain.DeliverySite, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, street_address, city, postal_code, country, delivery_instructions, created_at, updated_at
		FROM delivery_sites WHERE id = $1`, id)
	return scanDeliverySite(row)
}

func (r *B2BRepo) ListSitesByUser(ctx context.Context, userID string) ([]domain.DeliverySite, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, street_address, city, postal_code, country, delivery_instructions, created_at, updated_at
		FROM delivery_sites WHERE user_id = $1 ORDER BY name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []domain.DeliverySite
	for rows.Next() {
		s, err := scanDeliverySiteRow(rows)
		if err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func (r *B2BRepo) UpdateSite(ctx context.Context, site *domain.DeliverySite) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE delivery_sites SET
			name = $2, street_address = $3, city = $4, postal_code = $5,
			country = $6, delivery_instructions = $7, updated_at = $8
		WHERE id = $1`,
		site.ID, site.Name, site.StreetAddress, site.City, site.PostalCode,
		site.Country, nilIfEmpty(site.DeliveryInstructions), site.UpdatedAt)
	return err
}

func (r *B2BRepo) DeleteSite(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM delivery_sites WHERE id = $1`, id)
	return err
}

func (r *B2BRepo) CountSitesByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM delivery_sites WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func scanDeliverySite(row pgx.Row) (*domain.DeliverySite, error) {
	var s domain.DeliverySite
	var instructions *string
	err := row.Scan(&s.ID, &s.UserID, &s.Name, &s.StreetAddress, &s.City,
		&s.PostalCode, &s.Country, &instructions, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if instructions != nil {
		s.DeliveryInstructions = *instructions
	}
	return &s, nil
}

func scanDeliverySiteRow(rows pgx.Rows) (domain.DeliverySite, error) {
	var s domain.DeliverySite
	var instructions *string
	err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.StreetAddress, &s.City,
		&s.PostalCode, &s.Country, &instructions, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return s, err
	}
	if instructions != nil {
		s.DeliveryInstructions = *instructions
	}
	return s, nil
}

// --- Access Whitelisting ---

func (r *B2BRepo) CreateAccess(ctx context.Context, access *domain.B2BAccess) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bakery_b2b_access (id, bakery_id, business_user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		access.ID, access.BakeryID, access.BusinessUserID, string(access.Status),
		access.CreatedAt, access.UpdatedAt)
	return err
}

func (r *B2BRepo) GetAccessByID(ctx context.Context, id string) (*domain.B2BAccess, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, business_user_id, status, created_at, updated_at
		FROM bakery_b2b_access WHERE id = $1`, id)
	return scanB2BAccess(row)
}

func (r *B2BRepo) GetAccess(ctx context.Context, bakeryID string, userID string) (*domain.B2BAccess, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, business_user_id, status, created_at, updated_at
		FROM bakery_b2b_access WHERE bakery_id = $1 AND business_user_id = $2`, bakeryID, userID)
	return scanB2BAccess(row)
}

func (r *B2BRepo) UpdateAccessStatus(ctx context.Context, id string, status domain.B2BAccessStatus) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bakery_b2b_access SET status = $2, updated_at = $3 WHERE id = $1`,
		id, string(status), time.Now())
	return err
}

func (r *B2BRepo) ListAccessByBakery(ctx context.Context, bakeryID string, status *domain.B2BAccessStatus) ([]domain.B2BAccess, error) {
	var rows pgx.Rows
	var err error

	if status != nil {
		rows, err = r.pool.Query(ctx, `
			SELECT id, bakery_id, business_user_id, status, created_at, updated_at
			FROM bakery_b2b_access WHERE bakery_id = $1 AND status = $2
			ORDER BY created_at DESC`, bakeryID, string(*status))
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, bakery_id, business_user_id, status, created_at, updated_at
			FROM bakery_b2b_access WHERE bakery_id = $1
			ORDER BY created_at DESC`, bakeryID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accesses []domain.B2BAccess
	for rows.Next() {
		a, err := scanB2BAccessRow(rows)
		if err != nil {
			return nil, err
		}
		accesses = append(accesses, a)
	}
	return accesses, rows.Err()
}

func (r *B2BRepo) ListApprovedBakeryIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT bakery_id FROM bakery_b2b_access
		WHERE business_user_id = $1 AND status = 'approved'`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func scanB2BAccess(row pgx.Row) (*domain.B2BAccess, error) {
	var a domain.B2BAccess
	var status string
	err := row.Scan(&a.ID, &a.BakeryID, &a.BusinessUserID, &status, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.Status = domain.B2BAccessStatus(status)
	return &a, nil
}

func scanB2BAccessRow(rows pgx.Rows) (domain.B2BAccess, error) {
	var a domain.B2BAccess
	var status string
	err := rows.Scan(&a.ID, &a.BakeryID, &a.BusinessUserID, &status, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return a, err
	}
	a.Status = domain.B2BAccessStatus(status)
	return a, nil
}

// --- B2B Config ---

func (r *B2BRepo) GetConfig(ctx context.Context, bakeryID string) (*domain.B2BConfig, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, bakery_id, cutoff_time, delivery_window_start, delivery_window_end, order_minimum, pro_discount, vat_rate, created_at, updated_at
		FROM b2b_config WHERE bakery_id = $1`, bakeryID)

	var c domain.B2BConfig
	var cutoff, windowStart, windowEnd time.Time
	err := row.Scan(&c.ID, &c.BakeryID, &cutoff, &windowStart, &windowEnd,
		&c.OrderMinimum, &c.ProDiscount, &c.VATRate, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.CutoffTime = timeToTimeOfDay(cutoff)
	c.DeliveryWindowStart = timeToTimeOfDay(windowStart)
	c.DeliveryWindowEnd = timeToTimeOfDay(windowEnd)
	return &c, nil
}

func (r *B2BRepo) SaveConfig(ctx context.Context, config *domain.B2BConfig) error {
	cutoff := timeOfDayToTime(config.CutoffTime)
	windowStart := timeOfDayToTime(config.DeliveryWindowStart)
	windowEnd := timeOfDayToTime(config.DeliveryWindowEnd)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO b2b_config (id, bakery_id, cutoff_time, delivery_window_start, delivery_window_end, order_minimum, pro_discount, vat_rate, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (bakery_id) DO UPDATE SET
			cutoff_time = EXCLUDED.cutoff_time,
			delivery_window_start = EXCLUDED.delivery_window_start,
			delivery_window_end = EXCLUDED.delivery_window_end,
			order_minimum = EXCLUDED.order_minimum,
			pro_discount = EXCLUDED.pro_discount,
			vat_rate = EXCLUDED.vat_rate,
			updated_at = EXCLUDED.updated_at`,
		config.ID, config.BakeryID, cutoff, windowStart, windowEnd,
		config.OrderMinimum, config.ProDiscount, config.VATRate, config.CreatedAt, config.UpdatedAt)
	return err
}

// --- Volume Tiers ---

func (r *B2BRepo) ListVolumeTiers(ctx context.Context) ([]domain.VolumeTier, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, min_monthly_spend, discount_percent
		FROM volume_tiers ORDER BY min_monthly_spend ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []domain.VolumeTier
	for rows.Next() {
		var t domain.VolumeTier
		if err := rows.Scan(&t.ID, &t.MinMonthlySpend, &t.DiscountPercent); err != nil {
			return nil, err
		}
		tiers = append(tiers, t)
	}
	return tiers, rows.Err()
}

// --- Monthly Spend Tracking ---

func (r *B2BRepo) UpdateMonthlySpend(ctx context.Context, profileID string, amount int64, month string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE business_profiles
		SET current_month_spend = $2, spend_month = $3, updated_at = NOW()
		WHERE id = $1`,
		profileID, amount, month)
	return err
}

// --- Saved Lists ---

func (r *B2BRepo) CreateSavedList(ctx context.Context, list *domain.SavedList) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO saved_lists (id, user_id, bakery_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		list.ID, list.UserID, list.BakeryID, list.Name, list.CreatedAt, list.UpdatedAt)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO saved_list_items (id, saved_list_id, product_id, quantity)
			VALUES ($1, $2, $3, $4)`,
			item.ID, list.ID, item.ProductID, item.Quantity)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *B2BRepo) GetSavedListByID(ctx context.Context, id string) (*domain.SavedList, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, bakery_id, name, created_at, updated_at
		FROM saved_lists WHERE id = $1`, id)

	var l domain.SavedList
	err := row.Scan(&l.ID, &l.UserID, &l.BakeryID, &l.Name, &l.CreatedAt, &l.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.loadSavedListItems(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	l.Items = items
	return &l, nil
}

func (r *B2BRepo) ListSavedLists(ctx context.Context, userID string, bakeryID string) ([]domain.SavedList, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, bakery_id, name, created_at, updated_at
		FROM saved_lists WHERE user_id = $1 AND bakery_id = $2
		ORDER BY name`, userID, bakeryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []domain.SavedList
	for rows.Next() {
		var l domain.SavedList
		if err := rows.Scan(&l.ID, &l.UserID, &l.BakeryID, &l.Name, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		items, err := r.loadSavedListItems(ctx, l.ID)
		if err != nil {
			return nil, err
		}
		l.Items = items
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func (r *B2BRepo) DeleteSavedList(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM saved_lists WHERE id = $1`, id)
	return err
}

func (r *B2BRepo) loadSavedListItems(ctx context.Context, listID string) ([]domain.SavedListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, product_id, quantity FROM saved_list_items WHERE saved_list_id = $1`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.SavedListItem
	for rows.Next() {
		var item domain.SavedListItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// --- Invoices ---

func (r *B2BRepo) CreateInvoice(ctx context.Context, invoice *domain.B2BInvoice) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO b2b_invoices (id, order_id, bakery_id, business_profile_id, invoice_number, subtotal_ht, discount_amount, tva_amount, total_ttc, payment_status, issued_at, paid_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		invoice.ID, invoice.OrderID, invoice.BakeryID, invoice.BusinessProfileID,
		invoice.InvoiceNumber, invoice.SubtotalHT, invoice.DiscountAmount,
		invoice.TVAAmount, invoice.TotalTTC, invoice.PaymentStatus,
		invoice.IssuedAt, invoice.PaidAt)
	return err
}

func (r *B2BRepo) GetInvoiceByID(ctx context.Context, id string) (*domain.B2BInvoice, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, order_id, bakery_id, business_profile_id, invoice_number, subtotal_ht, discount_amount, tva_amount, total_ttc, payment_status, issued_at, paid_at
		FROM b2b_invoices WHERE id = $1`, id)
	return scanB2BInvoice(row)
}

func (r *B2BRepo) GetInvoiceByOrder(ctx context.Context, orderID string) (*domain.B2BInvoice, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, order_id, bakery_id, business_profile_id, invoice_number, subtotal_ht, discount_amount, tva_amount, total_ttc, payment_status, issued_at, paid_at
		FROM b2b_invoices WHERE order_id = $1`, orderID)
	return scanB2BInvoice(row)
}

func (r *B2BRepo) ListInvoicesByUser(ctx context.Context, profileID string, params domain.PaginationParams) ([]domain.B2BInvoice, int, error) {
	limit, offset := paginationToOffset(params)

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM b2b_invoices WHERE business_profile_id = $1`, profileID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, bakery_id, business_profile_id, invoice_number, subtotal_ht, discount_amount, tva_amount, total_ttc, payment_status, issued_at, paid_at
		FROM b2b_invoices WHERE business_profile_id = $1
		ORDER BY issued_at DESC
		LIMIT $2 OFFSET $3`, profileID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []domain.B2BInvoice
	for rows.Next() {
		inv, err := scanB2BInvoiceRow(rows)
		if err != nil {
			return nil, 0, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, total, rows.Err()
}

func (r *B2BRepo) NextInvoiceNumber(ctx context.Context, bakeryID string) (int, error) {
	var maxNum *int
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(invoice_number) FROM b2b_invoices WHERE bakery_id = $1`, bakeryID).Scan(&maxNum)
	if err != nil {
		return 0, err
	}
	if maxNum == nil {
		return 1, nil
	}
	return *maxNum + 1, nil
}

func scanB2BInvoice(row pgx.Row) (*domain.B2BInvoice, error) {
	var inv domain.B2BInvoice
	err := row.Scan(&inv.ID, &inv.OrderID, &inv.BakeryID, &inv.BusinessProfileID,
		&inv.InvoiceNumber, &inv.SubtotalHT, &inv.DiscountAmount,
		&inv.TVAAmount, &inv.TotalTTC, &inv.PaymentStatus, &inv.IssuedAt, &inv.PaidAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func scanB2BInvoiceRow(rows pgx.Rows) (domain.B2BInvoice, error) {
	var inv domain.B2BInvoice
	err := rows.Scan(&inv.ID, &inv.OrderID, &inv.BakeryID, &inv.BusinessProfileID,
		&inv.InvoiceNumber, &inv.SubtotalHT, &inv.DiscountAmount,
		&inv.TVAAmount, &inv.TotalTTC, &inv.PaymentStatus, &inv.IssuedAt, &inv.PaidAt)
	return inv, err
}
