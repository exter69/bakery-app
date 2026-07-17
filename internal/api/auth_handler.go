package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// registerRequest is the JSON body for POST /api/auth/register.
type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     *int   `json:"role,omitempty"`
}

// loginRequest is the JSON body for POST /api/auth/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// registerResponse is the JSON response for a successful registration.
type registerResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
}

// loginResponse is the JSON response for a successful login.
type loginResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

// userResponse is the user portion of the login response.
type userResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_BODY",
			Message: "invalid request body",
		})
		return
	}

	// Default role to customer (2); reject admin (0) registration via API
	role := domain.RoleCustomer
	if req.Role != nil {
		r := domain.UserRole(*req.Role)
		if r == domain.RoleAdmin {
			// Cannot register as admin via API
			role = domain.RoleCustomer
		} else if r == domain.RoleSeller || r == domain.RoleCustomer {
			role = r
		}
		// Any other value defaults to customer
	}

	user, err := h.authSvc.Register(r.Context(), req.Username, req.Password, role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameRequired), errors.Is(err, service.ErrPasswordTooShort):
			writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			})
		case errors.Is(err, service.ErrUsernameExists):
			writeJSON(w, http.StatusConflict, dto.ErrorResponse{
				Code:    "USERNAME_EXISTS",
				Message: "username already exists",
			})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "an unexpected error occurred",
			})
		}
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     int(user.Role),
	})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_BODY",
			Message: "invalid request body",
		})
		return
	}

	token, user, err := h.authSvc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "INVALID_CREDENTIALS",
				Message: "invalid username or password",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
		User: userResponse{
			ID:       user.ID,
			Username: user.Username,
			Role:     int(user.Role),
		},
	})
}
