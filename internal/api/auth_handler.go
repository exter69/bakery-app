package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/middleware"
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
	Username string  `json:"username"`
	Password string  `json:"password"`
	Role     *int    `json:"role,omitempty"`
	Code     *string `json:"code,omitempty"`
}

// loginRequest is the JSON body for POST /api/auth/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// requestAccessRequest is the JSON body for POST /api/auth/request-access.
type requestAccessRequest struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	BakeryName    string `json:"bakeryName"`
	BakeryAddress string `json:"bakeryAddress"`
}

// createTokenRequest is the JSON body for POST /api/admin/tokens.
type createTokenRequest struct {
	Email      string `json:"email"`
	BakeryName string `json:"bakeryName"`
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

	user, err := h.authSvc.Register(r.Context(), req.Username, req.Password, role, req.Code)
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
		case errors.Is(err, service.ErrTokenRequired):
			writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    "TOKEN_REQUIRED",
				Message: err.Error(),
			})
		case errors.Is(err, service.ErrInvalidToken):
			writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    "INVALID_TOKEN",
				Message: err.Error(),
			})
		case errors.Is(err, service.ErrTokenExpired):
			writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    "TOKEN_EXPIRED",
				Message: err.Error(),
			})
		case errors.Is(err, service.ErrTokenAlreadyUsed):
			writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    "TOKEN_ALREADY_USED",
				Message: err.Error(),
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

// RequestAccess handles POST /api/auth/request-access (public).
func (h *AuthHandler) RequestAccess(w http.ResponseWriter, r *http.Request) {
	var req requestAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_BODY",
			Message: "invalid request body",
		})
		return
	}

	if req.Email == "" {
		writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "email is required",
		})
		return
	}

	_, err := h.authSvc.CreateRegistrationToken(r.Context(), req.Email, req.BakeryName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Access request received. Check your email.",
	})
}

// CreateToken handles POST /api/admin/tokens (requires auth + role=0).
func (h *AuthHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	userRole := middleware.GetUserRoleFromContext(r.Context())
	if userRole != int(domain.RoleAdmin) {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "admin access required",
		})
		return
	}

	var req createTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_BODY",
			Message: "invalid request body",
		})
		return
	}

	if req.Email == "" {
		writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "email is required",
		})
		return
	}

	token, err := h.authSvc.CreateRegistrationToken(r.Context(), req.Email, req.BakeryName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"token":     token.Token,
		"expiresAt": token.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
