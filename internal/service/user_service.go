package service

import (
	"context"
	"fmt"

	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// UserService handles user profile and holiday mode operations.
type UserService struct {
	userRepo domain.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetProfile returns the user profile for the given user ID.
func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateHoliday updates the holiday mode fields for a user.
func (s *UserService) UpdateHoliday(ctx context.Context, userID string, req dto.UpdateHolidayRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	user.HolidayMode = req.HolidayMode
	user.HolidayFrom = req.HolidayFrom
	user.HolidayTo = req.HolidayTo

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return user, nil
}

// GetFavorites returns the list of favorite product IDs for a user.
func (s *UserService) GetFavorites(ctx context.Context, userID string) ([]string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.FavoriteProducts == nil {
		return []string{}, nil
	}
	return user.FavoriteProducts, nil
}

// UpdateFavorites replaces the user's favorite product IDs.
func (s *UserService) UpdateFavorites(ctx context.Context, userID string, productIDs []string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.FavoriteProducts = productIDs

	if err := s.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("saving user: %w", err)
	}

	return nil
}
