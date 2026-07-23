package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucatorrekens/bakery-app/internal/auth"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

const testStateKey = "test-state-signing-key-32bytes!!"

// fakeGoogleProvider is a test double for auth.OAuthProvider.
type fakeGoogleProvider struct {
	userInfo *auth.OAuthUserInfo
	err      error
}

func (f *fakeGoogleProvider) GetAuthURL(state, redirectURI string) string {
	return "https://google.com/oauth?state=" + state
}

func (f *fakeGoogleProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (*auth.OAuthUserInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.userInfo, nil
}

func setupOAuthHandler(t *testing.T) (*OAuthHandler, *chi.Mux) {
	t.Helper()
	userRepo := memory.NewUserRepo()
	socialLoginRepo := memory.NewSocialLoginRepo()

	providers := map[string]auth.OAuthProvider{
		"google": &fakeGoogleProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "g-user-1",
				Email:          "test@example.com",
				Name:           "Test User",
			},
		},
	}

	idCounter := 0
	svc := service.NewOAuthService(service.OAuthServiceConfig{
		Providers:       providers,
		SocialLoginRepo: socialLoginRepo,
		UserRepo:        userRepo,
		JWTSecret:       "test-secret",
		RedirectBase:    "http://localhost:5173",
		IDGen: func() string {
			idCounter++
			return "test-id-" + string(rune('a'+idCounter))
		},
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)
		},
	})

	handler := NewOAuthHandler(svc, []byte(testStateKey))

	r := chi.NewRouter()
	r.Get("/api/auth/oauth/{provider}", handler.GetAuthURL)
	r.Post("/api/auth/oauth/{provider}/callback", handler.HandleCallback)

	return handler, r
}

// generateValidState creates a valid state token from the handler for use in tests.
func generateValidState(t *testing.T, handler *OAuthHandler) string {
	t.Helper()
	state, err := handler.generateOAuthState()
	require.NoError(t, err)
	return state
}

func TestGetAuthURL_returnsRedirectURLAndState(t *testing.T) {
	_, router := setupOAuthHandler(t)

	req := httptest.NewRequest("GET", "/api/auth/oauth/google", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://google.com/oauth?state=")
	// Response should include the state field
	assert.Contains(t, w.Body.String(), `"state"`)
}

func TestGetAuthURL_returns400ForUnsupportedProvider(t *testing.T) {
	_, router := setupOAuthHandler(t)

	req := httptest.NewRequest("GET", "/api/auth/oauth/github", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "UNSUPPORTED_PROVIDER")
}

func TestHandleCallback_returnsTokenForValidCodeAndState(t *testing.T) {
	handler, router := setupOAuthHandler(t)
	state := generateValidState(t, handler)

	body := `{"code":"valid-code","state":"` + state + `"}`
	req := httptest.NewRequest("POST", "/api/auth/oauth/google/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
	assert.Contains(t, w.Body.String(), "user")
}

func TestHandleCallback_rejectsInvalidState(t *testing.T) {
	_, router := setupOAuthHandler(t)

	body := `{"code":"valid-code","state":"invalid-state-value"}`
	req := httptest.NewRequest("POST", "/api/auth/oauth/google/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_STATE")
}

func TestHandleCallback_rejectsMissingState(t *testing.T) {
	_, router := setupOAuthHandler(t)

	body := `{"code":"valid-code"}`
	req := httptest.NewRequest("POST", "/api/auth/oauth/google/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_STATE")
}

func TestHandleCallback_rejectsExpiredState(t *testing.T) {
	userRepo := memory.NewUserRepo()
	socialLoginRepo := memory.NewSocialLoginRepo()

	providers := map[string]auth.OAuthProvider{
		"google": &fakeGoogleProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "g-user-1",
				Email:          "test@example.com",
				Name:           "Test User",
			},
		},
	}

	svc := service.NewOAuthService(service.OAuthServiceConfig{
		Providers:       providers,
		SocialLoginRepo: socialLoginRepo,
		UserRepo:        userRepo,
		JWTSecret:       "test-secret",
		RedirectBase:    "http://localhost:5173",
		IDGen:           func() string { return "id-1" },
		Now:             func() time.Time { return time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC) },
	})

	handler := NewOAuthHandler(svc, []byte(testStateKey))
	// Generate state "in the past"
	handler.nowFunc = func() time.Time {
		return time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC) // 30 days ago
	}
	state := generateValidState(t, handler)

	// Reset nowFunc to "now" — the state is now expired (>10 min old)
	handler.nowFunc = time.Now

	r := chi.NewRouter()
	r.Post("/api/auth/oauth/{provider}/callback", handler.HandleCallback)

	body := `{"code":"valid-code","state":"` + state + `"}`
	req := httptest.NewRequest("POST", "/api/auth/oauth/google/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_STATE")
}

func TestHandleCallback_returns400WhenCodeMissing(t *testing.T) {
	_, router := setupOAuthHandler(t)

	body := `{"code":"","state":"something"}`
	req := httptest.NewRequest("POST", "/api/auth/oauth/google/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_CODE")
}

func TestHandleCallback_returns400ForUnsupportedProvider(t *testing.T) {
	handler, router := setupOAuthHandler(t)
	state := generateValidState(t, handler)

	body := `{"code":"some-code","state":"` + state + `"}`
	req := httptest.NewRequest("POST", "/api/auth/oauth/github/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "UNSUPPORTED_PROVIDER")
}

func TestHandleCallback_rejectsLinkToPasswordProtectedAccount(t *testing.T) {
	userRepo := memory.NewUserRepo()
	socialLoginRepo := memory.NewSocialLoginRepo()

	// Pre-create a user with a password
	existingUser := &domain.User{
		ID:           "existing-1",
		Username:     "test@example.com",
		PasswordHash: "$2a$10$somehash", // has password — should NOT auto-link
		Role:         domain.RoleCustomer,
	}
	require.NoError(t, userRepo.Save(context.Background(), existingUser))

	providers := map[string]auth.OAuthProvider{
		"google": &fakeGoogleProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "g-user-2",
				Email:          "test@example.com",
				Name:           "Test User",
			},
		},
	}

	idCounter := 0
	svc := service.NewOAuthService(service.OAuthServiceConfig{
		Providers:       providers,
		SocialLoginRepo: socialLoginRepo,
		UserRepo:        userRepo,
		JWTSecret:       "test-secret",
		RedirectBase:    "http://localhost:5173",
		IDGen: func() string {
			idCounter++
			return "link-id-" + string(rune('a'+idCounter))
		},
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)
		},
	})

	handler := NewOAuthHandler(svc, []byte(testStateKey))
	state := generateValidState(t, handler)

	r := chi.NewRouter()
	r.Post("/api/auth/oauth/{provider}/callback", handler.HandleCallback)

	body := `{"code":"valid-code","state":"` + state + `"}`
	req := httptest.NewRequest("POST", "/api/auth/oauth/google/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "ACCOUNT_LINK_REQUIRES_VERIFICATION")
}

func TestHandleCallback_linksToSocialOnlyAccount(t *testing.T) {
	userRepo := memory.NewUserRepo()
	socialLoginRepo := memory.NewSocialLoginRepo()

	// Pre-create a social-only user (no password)
	existingUser := &domain.User{
		ID:           "social-only-1",
		Username:     "social@example.com",
		PasswordHash: "", // no password — safe to auto-link
		Role:         domain.RoleCustomer,
	}
	require.NoError(t, userRepo.Save(context.Background(), existingUser))

	providers := map[string]auth.OAuthProvider{
		"google": &fakeGoogleProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderUserID: "g-user-3",
				Email:          "social@example.com",
				Name:           "Social User",
			},
		},
	}

	idCounter := 0
	svc := service.NewOAuthService(service.OAuthServiceConfig{
		Providers:       providers,
		SocialLoginRepo: socialLoginRepo,
		UserRepo:        userRepo,
		JWTSecret:       "test-secret",
		RedirectBase:    "http://localhost:5173",
		IDGen: func() string {
			idCounter++
			return "link-id-" + string(rune('a'+idCounter))
		},
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)
		},
	})

	handler := NewOAuthHandler(svc, []byte(testStateKey))
	state := generateValidState(t, handler)

	r := chi.NewRouter()
	r.Post("/api/auth/oauth/{provider}/callback", handler.HandleCallback)

	body := `{"code":"valid-code","state":"` + state + `"}`
	req := httptest.NewRequest("POST", "/api/auth/oauth/google/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp oauthCallbackResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "social-only-1", resp.User.ID)
}
