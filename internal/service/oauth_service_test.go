package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucatorrekens/bakery-app/internal/auth"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
)

// fakeProvider is a test double for auth.OAuthProvider.
type fakeProvider struct {
	authURL  string
	userInfo *auth.OAuthUserInfo
	err      error
}

func (f *fakeProvider) GetAuthURL(state, redirectURI string) string {
	return f.authURL
}

func (f *fakeProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (*auth.OAuthUserInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.userInfo, nil
}

func newTestOAuthService(providers map[string]auth.OAuthProvider) (*OAuthService, *memory.UserRepo, *memory.SocialLoginRepo) {
	userRepo := memory.NewUserRepo()
	socialLoginRepo := memory.NewSocialLoginRepo()

	idCounter := 0
	svc := NewOAuthService(OAuthServiceConfig{
		Providers:       providers,
		SocialLoginRepo: socialLoginRepo,
		UserRepo:        userRepo,
		JWTSecret:       "test-secret",
		RedirectBase:    "http://localhost:5173",
		IDGen: func() string {
			idCounter++
			return "id-" + string(rune('0'+idCounter))
		},
		Now: func() time.Time {
			return time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
		},
	})

	return svc, userRepo, socialLoginRepo
}

func TestGetAuthURL_returnsURLForValidProvider(t *testing.T) {
	providers := map[string]auth.OAuthProvider{
		"google": &fakeProvider{authURL: "https://google.com/auth?test=1"},
	}
	svc, _, _ := newTestOAuthService(providers)

	url, err := svc.GetAuthURL("google", "some-state")
	require.NoError(t, err)
	assert.Equal(t, "https://google.com/auth?test=1", url)
}

func TestGetAuthURL_returnsErrorForUnsupportedProvider(t *testing.T) {
	svc, _, _ := newTestOAuthService(map[string]auth.OAuthProvider{})

	_, err := svc.GetAuthURL("github", "state")
	assert.ErrorIs(t, err, ErrUnsupportedProvider)
}

func TestHandleCallback_createsNewUserWhenNoneExists(t *testing.T) {
	providers := map[string]auth.OAuthProvider{
		"google": &fakeProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "google-123",
				Email:          "alice@example.com",
				Name:           "Alice",
			},
		},
	}
	svc, userRepo, socialLoginRepo := newTestOAuthService(providers)

	result, err := svc.HandleCallback(context.Background(), "google", "auth-code")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "alice@example.com", result.User.Username)
	assert.Equal(t, domain.RoleCustomer, result.User.Role)

	// Verify user was persisted
	user, _ := userRepo.GetByUsername(context.Background(), "alice@example.com")
	require.NotNil(t, user)

	// Verify social login was created
	logins, _ := socialLoginRepo.ListByUser(context.Background(), user.ID)
	require.Len(t, logins, 1)
	assert.Equal(t, "google", logins[0].Provider)
	assert.Equal(t, "google-123", logins[0].ProviderUserID)
}

func TestHandleCallback_linksToExistingUserByUsername(t *testing.T) {
	providers := map[string]auth.OAuthProvider{
		"google": &fakeProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "google-456",
				Email:          "bob@example.com",
				Name:           "Bob",
			},
		},
	}
	svc, userRepo, socialLoginRepo := newTestOAuthService(providers)

	// Pre-create a social-only user (no password) with the same email as username
	existingUser := &domain.User{
		ID:           "existing-user-1",
		Username:     "bob@example.com",
		PasswordHash: "", // no password — safe to auto-link
		Role:         domain.RoleCustomer,
		CreatedAt:    time.Now(),
	}
	_ = userRepo.Save(context.Background(), existingUser)

	result, err := svc.HandleCallback(context.Background(), "google", "auth-code")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should use the existing user, not create a new one
	assert.Equal(t, "existing-user-1", result.User.ID)
	assert.Equal(t, "bob@example.com", result.User.Username)

	// Social login should be linked to the existing user
	logins, _ := socialLoginRepo.ListByUser(context.Background(), "existing-user-1")
	require.Len(t, logins, 1)
	assert.Equal(t, "google-456", logins[0].ProviderUserID)
}

func TestHandleCallback_rejectsLinkToPasswordProtectedUser(t *testing.T) {
	providers := map[string]auth.OAuthProvider{
		"google": &fakeProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "google-789",
				Email:          "protected@example.com",
				Name:           "Protected",
			},
		},
	}
	svc, userRepo, _ := newTestOAuthService(providers)

	// Pre-create a user with a password (not safe to auto-link)
	existingUser := &domain.User{
		ID:           "protected-user-1",
		Username:     "protected@example.com",
		PasswordHash: "$2a$10$somebcrypthash",
		Role:         domain.RoleCustomer,
		CreatedAt:    time.Now(),
	}
	_ = userRepo.Save(context.Background(), existingUser)

	_, err := svc.HandleCallback(context.Background(), "google", "auth-code")
	assert.ErrorIs(t, err, ErrOAuthAccountLinkRequiresVerification)
}

func TestHandleCallback_returnsExistingSessionWhenAlreadyLinked(t *testing.T) {
	providers := map[string]auth.OAuthProvider{
		"google": &fakeProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "google-789",
				Email:          "carol@example.com",
				Name:           "Carol",
			},
		},
	}
	svc, userRepo, socialLoginRepo := newTestOAuthService(providers)

	// Pre-create user and social login
	existingUser := &domain.User{
		ID:       "carol-id",
		Username: "carol@example.com",
		Role:     domain.RoleCustomer,
	}
	_ = userRepo.Save(context.Background(), existingUser)
	_ = socialLoginRepo.Create(context.Background(), &domain.SocialLogin{
		ID:             "sl-1",
		UserID:         "carol-id",
		Provider:       "google",
		ProviderUserID: "google-789",
		Email:          "carol@example.com",
	})

	result, err := svc.HandleCallback(context.Background(), "google", "auth-code")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "carol-id", result.User.ID)
	assert.NotEmpty(t, result.Token)

	// No new social login should be created
	logins, _ := socialLoginRepo.ListByUser(context.Background(), "carol-id")
	assert.Len(t, logins, 1)
}

func TestHandleCallback_returnsErrorForUnsupportedProvider(t *testing.T) {
	svc, _, _ := newTestOAuthService(map[string]auth.OAuthProvider{})

	_, err := svc.HandleCallback(context.Background(), "github", "code")
	assert.ErrorIs(t, err, ErrUnsupportedProvider)
}

func TestHandleCallback_returnsErrorWhenNoEmail(t *testing.T) {
	providers := map[string]auth.OAuthProvider{
		"google": &fakeProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "google-no-email",
				Email:          "",
				Name:           "NoEmail",
			},
		},
	}
	svc, _, _ := newTestOAuthService(providers)

	_, err := svc.HandleCallback(context.Background(), "google", "auth-code")
	assert.ErrorIs(t, err, ErrOAuthNoEmail)
}
