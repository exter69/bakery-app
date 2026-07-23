package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/lucatorrekens/bakery-app/internal/auth"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// OAuthService handles social login flows.
type OAuthService struct {
	providers       map[string]auth.OAuthProvider
	socialLoginRepo domain.SocialLoginRepository
	userRepo        domain.UserRepository
	jwtSecret       string
	redirectBase    string
	idGen           func() string
	now             func() time.Time
}

// OAuthServiceConfig holds dependencies for the OAuth service.
type OAuthServiceConfig struct {
	Providers       map[string]auth.OAuthProvider
	SocialLoginRepo domain.SocialLoginRepository
	UserRepo        domain.UserRepository
	JWTSecret       string
	RedirectBase    string // e.g., "http://localhost:5173"
	IDGen           func() string
	Now             func() time.Time
}

// NewOAuthService creates a new OAuthService.
func NewOAuthService(cfg OAuthServiceConfig) *OAuthService {
	idGen := cfg.IDGen
	if idGen == nil {
		idGen = defaultOAuthIDGen
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &OAuthService{
		providers:       cfg.Providers,
		socialLoginRepo: cfg.SocialLoginRepo,
		userRepo:        cfg.UserRepo,
		jwtSecret:       cfg.JWTSecret,
		redirectBase:    cfg.RedirectBase,
		idGen:           idGen,
		now:             now,
	}
}

func defaultOAuthIDGen() string {
	return uuid.New().String()
}

// GetAuthURL returns the OAuth authorization URL for the given provider.
func (s *OAuthService) GetAuthURL(provider, state string) (string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", ErrUnsupportedProvider
	}
	redirectURI := s.redirectBase + "/auth/callback"
	return p.GetAuthURL(state, redirectURI), nil
}

// OAuthLoginResult holds the result of a successful OAuth callback.
type OAuthLoginResult struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

// HandleCallback exchanges the authorization code and logs in or creates the user.
func (s *OAuthService) HandleCallback(ctx context.Context, provider, code string) (*OAuthLoginResult, error) {
	p, ok := s.providers[provider]
	if !ok {
		return nil, ErrUnsupportedProvider
	}

	redirectURI := s.redirectBase + "/auth/callback"
	userInfo, err := p.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("exchanging OAuth code: %w", err)
	}

	if userInfo.Email == "" {
		return nil, ErrOAuthNoEmail
	}

	// Step 1: Check if social login already exists for this provider+providerUserID
	existing, err := s.socialLoginRepo.GetByProviderUser(ctx, provider, userInfo.ProviderUserID)
	if err != nil {
		return nil, fmt.Errorf("looking up social login: %w", err)
	}
	if existing != nil {
		// User already linked — issue JWT
		user, err := s.userRepo.GetByID(ctx, existing.UserID)
		if err != nil {
			return nil, fmt.Errorf("fetching user: %w", err)
		}
		if user == nil {
			return nil, fmt.Errorf("linked user not found")
		}
		token, err := s.issueJWT(user)
		if err != nil {
			return nil, err
		}
		return &OAuthLoginResult{Token: token, User: user}, nil
	}

	// Step 2: Check if a user exists with the same email (match by username=email or contactEmail)
	// We search by username since that's what the existing user repo supports.
	// Also look for an email-based social login to avoid duplicates.
	existingByEmail, err := s.socialLoginRepo.GetByProviderEmail(ctx, provider, userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("looking up social login by email: %w", err)
	}
	if existingByEmail != nil {
		user, err := s.userRepo.GetByID(ctx, existingByEmail.UserID)
		if err != nil {
			return nil, fmt.Errorf("fetching user by email link: %w", err)
		}
		if user != nil {
			token, err := s.issueJWT(user)
			if err != nil {
				return nil, err
			}
			return &OAuthLoginResult{Token: token, User: user}, nil
		}
	}

	// Step 3: Try to find existing user by username (email as username)
	existingUser, err := s.userRepo.GetByUsername(ctx, userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}

	var user *domain.User
	if existingUser != nil {
		// Security: only auto-link if the existing account has no password (social-only).
		// If the account has a password, linking by unverified provider email is an
		// account takeover vector. Require the user to log in with their password first.
		if existingUser.PasswordHash != "" {
			return nil, ErrOAuthAccountLinkRequiresVerification
		}
		// Link the social login to the existing social-only user
		user = existingUser
	} else {
		// Step 4: Create new user
		user = &domain.User{
			ID:           s.idGen(),
			Username:     userInfo.Email,
			PasswordHash: "", // no password for social-only accounts
			Role:         domain.RoleCustomer,
			ContactEmail: userInfo.Email,
			CreatedAt:    s.now(),
		}
		if err := s.userRepo.Save(ctx, user); err != nil {
			return nil, fmt.Errorf("creating user: %w", err)
		}
	}

	// Create the social login link
	socialLogin := &domain.SocialLogin{
		ID:             s.idGen(),
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: userInfo.ProviderUserID,
		Email:          userInfo.Email,
		CreatedAt:      s.now(),
	}
	if err := s.socialLoginRepo.Create(ctx, socialLogin); err != nil {
		return nil, fmt.Errorf("creating social login: %w", err)
	}

	token, err := s.issueJWT(user)
	if err != nil {
		return nil, err
	}
	return &OAuthLoginResult{Token: token, User: user}, nil
}

// issueJWT generates a signed JWT for the given user.
func (s *OAuthService) issueJWT(user *domain.User) (string, error) {
	now := s.now()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": int(user.Role),
		"exp":  jwt.NewNumericDate(now.Add(24 * time.Hour)),
		"iat":  jwt.NewNumericDate(now),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}
	return tokenString, nil
}
