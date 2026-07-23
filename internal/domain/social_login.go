package domain

import (
	"context"
	"time"
)

// SocialLogin represents a linked OAuth provider for a user.
type SocialLogin struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	Provider       string    `json:"provider"` // "google" or "apple"
	ProviderUserID string    `json:"providerUserId"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"createdAt"`
}

// SocialLoginRepository provides data access for social login records.
type SocialLoginRepository interface {
	// GetByProviderUser finds a social login by provider and provider user ID.
	GetByProviderUser(ctx context.Context, provider string, providerUserID string) (*SocialLogin, error)

	// GetByProviderEmail finds a social login by provider and email.
	GetByProviderEmail(ctx context.Context, provider string, email string) (*SocialLogin, error)

	// Create persists a new social login record.
	Create(ctx context.Context, login *SocialLogin) error

	// ListByUser returns all social logins for a user.
	ListByUser(ctx context.Context, userID string) ([]SocialLogin, error)
}
