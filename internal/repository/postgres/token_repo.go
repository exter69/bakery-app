package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// TokenRepo implements domain.RegistrationTokenRepository backed by PostgreSQL.
type TokenRepo struct {
	pool *pgxpool.Pool
}

// NewTokenRepo creates a new TokenRepo with the given connection pool.
func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{pool: pool}
}

func (r *TokenRepo) Save(ctx context.Context, token *domain.RegistrationToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO registration_tokens (id, token, email, bakery_name, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			bakery_name = EXCLUDED.bakery_name,
			expires_at = EXCLUDED.expires_at,
			used = EXCLUDED.used`,
		token.ID, token.Token, token.Email, token.BakeryName,
		token.ExpiresAt, token.Used, token.CreatedAt)
	return err
}

func (r *TokenRepo) GetByToken(ctx context.Context, tokenStr string) (*domain.RegistrationToken, error) {
	var t domain.RegistrationToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, token, email, bakery_name, expires_at, used, created_at
		FROM registration_tokens WHERE token = $1`, tokenStr).Scan(
		&t.ID, &t.Token, &t.Email, &t.BakeryName,
		&t.ExpiresAt, &t.Used, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepo) MarkUsed(ctx context.Context, tokenStr string) error {
	_, err := r.pool.Exec(ctx, `UPDATE registration_tokens SET used = true WHERE token = $1`, tokenStr)
	return err
}
