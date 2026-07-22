package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// AuthService handles user registration and login.
type AuthService struct {
	userRepo     domain.UserRepository
	tokenRepo    domain.RegistrationTokenRepository
	jwtSecret    string
	contactEmail string
	idGen        func() string
	now          func() time.Time
}

// AuthServiceConfig holds dependencies for the auth service.
type AuthServiceConfig struct {
	UserRepo     domain.UserRepository
	TokenRepo    domain.RegistrationTokenRepository
	JWTSecret    string
	ContactEmail string // admin email to receive access request notifications
	IDGen        func() string
	Now          func() time.Time
}

// NewAuthService creates a new AuthService with the given dependencies.
func NewAuthService(cfg AuthServiceConfig) *AuthService {
	idGen := cfg.IDGen
	if idGen == nil {
		idGen = defaultAuthIDGen
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &AuthService{
		userRepo:     cfg.UserRepo,
		tokenRepo:    cfg.TokenRepo,
		jwtSecret:    cfg.JWTSecret,
		contactEmail: cfg.ContactEmail,
		idGen:        idGen,
		now:          now,
	}
}

var authIDCounter int

func defaultAuthIDGen() string {
	authIDCounter++
	return fmt.Sprintf("user-%d", authIDCounter)
}

// generateTokenCode produces an 8-character alphanumeric code avoiding ambiguous characters.
func generateTokenCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// Register creates a new user account.
// If registering as a seller (role=1), a valid registration code is required.
func (s *AuthService) Register(ctx context.Context, username, password string, role domain.UserRole, code *string) (*domain.User, error) {
	// Validate username
	if username == "" {
		return nil, ErrUsernameRequired
	}

	// Validate password
	if len(password) < 6 {
		return nil, ErrPasswordTooShort
	}

	var contactEmail string

	// Seller registration requires a valid token
	if role == domain.RoleSeller {
		if code == nil || *code == "" {
			return nil, ErrTokenRequired
		}

		token, err := s.tokenRepo.GetByToken(ctx, *code)
		if err != nil {
			return nil, fmt.Errorf("looking up token: %w", err)
		}
		if token == nil {
			return nil, ErrInvalidToken
		}
		if token.Used {
			return nil, ErrTokenAlreadyUsed
		}
		if s.now().After(token.ExpiresAt) {
			return nil, ErrTokenExpired
		}

		// Mark the token as used
		if err := s.tokenRepo.MarkUsed(ctx, *code); err != nil {
			return nil, fmt.Errorf("marking token used: %w", err)
		}

		contactEmail = token.Email
	}

	// Check if username already exists
	existing, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("checking username: %w", err)
	}
	if existing != nil {
		return nil, ErrUsernameExists
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &domain.User{
		ID:           s.idGen(),
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		ContactEmail: contactEmail,
		CreatedAt:    s.now(),
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return user, nil
}

// CreateRegistrationToken generates a new registration token for the given email/bakery.
func (s *AuthService) CreateRegistrationToken(ctx context.Context, email, bakeryName string) (*domain.RegistrationToken, error) {
	token := &domain.RegistrationToken{
		ID:         s.idGen(),
		Token:      generateTokenCode(),
		Email:      email,
		BakeryName: bakeryName,
		ExpiresAt:  s.now().Add(7 * 24 * time.Hour),
		Used:       false,
		CreatedAt:  s.now(),
	}

	if err := s.tokenRepo.Save(ctx, token); err != nil {
		return nil, fmt.Errorf("saving token: %w", err)
	}

	log.Printf("[EMAIL] Token %s created for baker %s (%s). Send notification to admin: %s", token.Token, email, bakeryName, s.contactEmail)

	return token, nil
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, *domain.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return "", nil, ErrInvalidCredentials
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT
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
		return "", nil, fmt.Errorf("signing token: %w", err)
	}

	return tokenString, user, nil
}
