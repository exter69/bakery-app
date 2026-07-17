package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// AuthService handles user registration and login.
type AuthService struct {
	userRepo  domain.UserRepository
	jwtSecret string
	idGen     func() string
	now       func() time.Time
}

// AuthServiceConfig holds dependencies for the auth service.
type AuthServiceConfig struct {
	UserRepo  domain.UserRepository
	JWTSecret string
	IDGen     func() string
	Now       func() time.Time
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
		userRepo:  cfg.UserRepo,
		jwtSecret: cfg.JWTSecret,
		idGen:     idGen,
		now:       now,
	}
}

var authIDCounter int

func defaultAuthIDGen() string {
	authIDCounter++
	return fmt.Sprintf("user-%d", authIDCounter)
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, username, password string, role domain.UserRole) (*domain.User, error) {
	// Validate username
	if username == "" {
		return nil, ErrUsernameRequired
	}

	// Validate password
	if len(password) < 6 {
		return nil, ErrPasswordTooShort
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
		CreatedAt:    s.now(),
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return user, nil
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("signing token: %w", err)
	}

	return tokenString, user, nil
}
