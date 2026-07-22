package memory

import (
	"context"
	"sync"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// TokenRepo is an in-memory implementation of domain.RegistrationTokenRepository.
type TokenRepo struct {
	mu     sync.RWMutex
	tokens map[string]domain.RegistrationToken // keyed by token string
}

// NewTokenRepo creates a new in-memory TokenRepo.
func NewTokenRepo() *TokenRepo {
	return &TokenRepo{
		tokens: make(map[string]domain.RegistrationToken),
	}
}

// Save persists a registration token.
func (r *TokenRepo) Save(_ context.Context, token *domain.RegistrationToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token.Token] = *token
	return nil
}

// GetByToken returns a registration token by its token string, or nil if not found.
func (r *TokenRepo) GetByToken(_ context.Context, tokenStr string) (*domain.RegistrationToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tokens[tokenStr]
	if !ok {
		return nil, nil
	}
	return &t, nil
}

// MarkUsed marks a registration token as used.
func (r *TokenRepo) MarkUsed(_ context.Context, tokenStr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tokens[tokenStr]
	if !ok {
		return nil
	}
	t.Used = true
	r.tokens[tokenStr] = t
	return nil
}
