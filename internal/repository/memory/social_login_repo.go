package memory

import (
	"context"
	"sync"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// SocialLoginRepo is an in-memory implementation of domain.SocialLoginRepository.
type SocialLoginRepo struct {
	mu     sync.RWMutex
	logins map[string]domain.SocialLogin // keyed by ID
}

// NewSocialLoginRepo creates a new in-memory SocialLoginRepo.
func NewSocialLoginRepo() *SocialLoginRepo {
	return &SocialLoginRepo{
		logins: make(map[string]domain.SocialLogin),
	}
}

// GetByProviderUser finds a social login by provider and provider user ID.
func (r *SocialLoginRepo) GetByProviderUser(_ context.Context, provider string, providerUserID string) (*domain.SocialLogin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, sl := range r.logins {
		if sl.Provider == provider && sl.ProviderUserID == providerUserID {
			return &sl, nil
		}
	}
	return nil, nil
}

// GetByProviderEmail finds a social login by provider and email.
func (r *SocialLoginRepo) GetByProviderEmail(_ context.Context, provider string, email string) (*domain.SocialLogin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, sl := range r.logins {
		if sl.Provider == provider && sl.Email == email {
			return &sl, nil
		}
	}
	return nil, nil
}

// Create persists a new social login record.
func (r *SocialLoginRepo) Create(_ context.Context, login *domain.SocialLogin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logins[login.ID] = *login
	return nil
}

// ListByUser returns all social logins for a user.
func (r *SocialLoginRepo) ListByUser(_ context.Context, userID string) ([]domain.SocialLogin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.SocialLogin
	for _, sl := range r.logins {
		if sl.UserID == userID {
			result = append(result, sl)
		}
	}
	return result, nil
}

// DeleteByUser removes all social login records for the given user.
func (r *SocialLoginRepo) DeleteByUser(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, sl := range r.logins {
		if sl.UserID == userID {
			delete(r.logins, id)
		}
	}
	return nil
}
