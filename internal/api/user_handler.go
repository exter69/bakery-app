package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// UserService defines the methods the user handler needs.
type UserService interface {
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	UpdateHoliday(ctx context.Context, userID string, req dto.UpdateHolidayRequest) (*domain.User, error)
	GetFavorites(ctx context.Context, userID string) ([]string, error)
	UpdateFavorites(ctx context.Context, userID string, productIDs []string) error
}

// UserHandler handles HTTP requests related to user profile and holiday mode.
type UserHandler struct {
	svc UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetProfile handles GET /api/user/profile.
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	user, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "user not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to get user profile",
		})
		return
	}

	writeJSON(w, http.StatusOK, userToProfileResponse(user))
}

// UpdateHoliday handles PUT /api/user/holiday.
func (h *UserHandler) UpdateHoliday(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateHolidayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	userID := extractUserID(r)

	user, err := h.svc.UpdateHoliday(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "user not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to update holiday mode",
		})
		return
	}

	writeJSON(w, http.StatusOK, userToProfileResponse(user))
}

// userToProfileResponse converts a domain User to its profile DTO response.
func userToProfileResponse(u *domain.User) dto.UserProfileResponse {
	favorites := u.FavoriteProducts
	if favorites == nil {
		favorites = []string{}
	}
	return dto.UserProfileResponse{
		ID:               u.ID,
		Username:         u.Username,
		Role:             int(u.Role),
		HolidayMode:      u.HolidayMode,
		HolidayFrom:      u.HolidayFrom,
		HolidayTo:        u.HolidayTo,
		FavoriteProducts: favorites,
		CreatedAt:        u.CreatedAt,
	}
}

// GetFavorites handles GET /api/user/favorites.
func (h *UserHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	favorites, err := h.svc.GetFavorites(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "user not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to get favorites",
		})
		return
	}

	writeJSON(w, http.StatusOK, dto.FavoritesResponse{ProductIDs: favorites})
}

// UpdateFavorites handles PUT /api/user/favorites.
func (h *UserHandler) UpdateFavorites(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateFavoritesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	userID := extractUserID(r)

	err := h.svc.UpdateFavorites(r.Context(), userID, req.ProductIDs)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "user not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to update favorites",
		})
		return
	}

	writeJSON(w, http.StatusOK, dto.FavoritesResponse{ProductIDs: req.ProductIDs})
}
