package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// UserRepo implements domain.UserRepository backed by PostgreSQL.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo with the given connection pool.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Save(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, role, contact_email, locale, holiday_mode, holiday_from, holiday_to, favorite_products, stripe_customer_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			contact_email = EXCLUDED.contact_email,
			locale = EXCLUDED.locale,
			holiday_mode = EXCLUDED.holiday_mode,
			holiday_from = EXCLUDED.holiday_from,
			holiday_to = EXCLUDED.holiday_to,
			favorite_products = EXCLUDED.favorite_products,
			stripe_customer_id = EXCLUDED.stripe_customer_id`

	favorites := user.FavoriteProducts
	if favorites == nil {
		favorites = []string{}
	}

	// Store NULL for empty stripe_customer_id
	var stripeCustomerID *string
	if user.StripeCustomerID != "" {
		stripeCustomerID = &user.StripeCustomerID
	}

	locale := user.Locale
	if locale == "" {
		locale = "en"
	}

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.PasswordHash,
		int(user.Role),
		user.ContactEmail,
		locale,
		user.HolidayMode,
		user.HolidayFrom,
		user.HolidayTo,
		favorites,
		stripeCustomerID,
		user.CreatedAt,
	)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, role, contact_email, locale, holiday_mode, holiday_from, holiday_to, favorite_products, stripe_customer_id, created_at
		FROM users WHERE id = $1`

	return r.scanUser(ctx, query, id)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, role, contact_email, locale, holiday_mode, holiday_from, holiday_to, favorite_products, stripe_customer_id, created_at
		FROM users WHERE username = $1`

	return r.scanUser(ctx, query, username)
}

func (r *UserRepo) scanUser(ctx context.Context, query string, arg any) (*domain.User, error) {
	var u domain.User
	var role int
	var holidayFrom, holidayTo *time.Time
	var favorites []string
	var stripeCustomerID *string
	var locale *string

	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&role,
		&u.ContactEmail,
		&locale,
		&u.HolidayMode,
		&holidayFrom,
		&holidayTo,
		&favorites,
		&stripeCustomerID,
		&u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	u.Role = domain.UserRole(role)
	u.HolidayFrom = holidayFrom
	u.HolidayTo = holidayTo
	if favorites != nil {
		u.FavoriteProducts = favorites
	} else {
		u.FavoriteProducts = []string{}
	}
	if stripeCustomerID != nil {
		u.StripeCustomerID = *stripeCustomerID
	}
	if locale != nil {
		u.Locale = *locale
	}

	return &u, nil
}
