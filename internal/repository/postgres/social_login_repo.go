package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// SocialLoginRepo implements domain.SocialLoginRepository backed by PostgreSQL.
type SocialLoginRepo struct {
	pool *pgxpool.Pool
}

// NewSocialLoginRepo creates a new SocialLoginRepo with the given connection pool.
func NewSocialLoginRepo(pool *pgxpool.Pool) *SocialLoginRepo {
	return &SocialLoginRepo{pool: pool}
}

func (r *SocialLoginRepo) GetByProviderUser(ctx context.Context, provider string, providerUserID string) (*domain.SocialLogin, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, created_at
		FROM social_logins
		WHERE provider = $1 AND provider_user_id = $2`

	return r.scanSocialLogin(ctx, query, provider, providerUserID)
}

func (r *SocialLoginRepo) GetByProviderEmail(ctx context.Context, provider string, email string) (*domain.SocialLogin, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, created_at
		FROM social_logins
		WHERE provider = $1 AND email = $2`

	return r.scanSocialLogin(ctx, query, provider, email)
}

func (r *SocialLoginRepo) Create(ctx context.Context, login *domain.SocialLogin) error {
	query := `
		INSERT INTO social_logins (id, user_id, provider, provider_user_id, email, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, query,
		login.ID,
		login.UserID,
		login.Provider,
		login.ProviderUserID,
		login.Email,
		login.CreatedAt,
	)
	return err
}

func (r *SocialLoginRepo) ListByUser(ctx context.Context, userID string) ([]domain.SocialLogin, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, created_at
		FROM social_logins
		WHERE user_id = $1
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logins []domain.SocialLogin
	for rows.Next() {
		var sl domain.SocialLogin
		if err := rows.Scan(&sl.ID, &sl.UserID, &sl.Provider, &sl.ProviderUserID, &sl.Email, &sl.CreatedAt); err != nil {
			return nil, err
		}
		logins = append(logins, sl)
	}
	return logins, rows.Err()
}

func (r *SocialLoginRepo) scanSocialLogin(ctx context.Context, query string, args ...any) (*domain.SocialLogin, error) {
	var sl domain.SocialLogin
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&sl.ID,
		&sl.UserID,
		&sl.Provider,
		&sl.ProviderUserID,
		&sl.Email,
		&sl.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sl, nil
}
