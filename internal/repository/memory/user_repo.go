package memory

import (
	"context"
	"sync"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// UserRepo is an in-memory implementation of domain.UserRepository.
type UserRepo struct {
	mu    sync.RWMutex
	users map[string]domain.User // keyed by ID
}

// NewUserRepo creates a new in-memory UserRepo.
func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[string]domain.User),
	}
}

// Save persists a user (create or update).
func (r *UserRepo) Save(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = *user
	return nil
}

// GetByID returns a user by ID, or nil if not found.
func (r *UserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return &u, nil
}

// GetByUsername returns a user by username, or nil if not found.
func (r *UserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Username == username {
			return &u, nil
		}
	}
	return nil, nil
}
