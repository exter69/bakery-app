package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/payment"
)

// PaymentMethodService defines the methods the payment method handler needs.
type PaymentMethodService interface {
	ListPaymentMethods(ctx context.Context, userID string) ([]payment.SavedPaymentMethod, error)
	CreateSetupIntent(ctx context.Context, userID string) (string, error)
	DetachPaymentMethod(ctx context.Context, userID string, paymentMethodID string) error
	SetDefaultPaymentMethod(ctx context.Context, userID string, paymentMethodID string) error
}

// PaymentMethodHandler handles HTTP requests for saved payment methods.
type PaymentMethodHandler struct {
	svc PaymentMethodService
}

// NewPaymentMethodHandler creates a new PaymentMethodHandler.
func NewPaymentMethodHandler(svc PaymentMethodService) *PaymentMethodHandler {
	return &PaymentMethodHandler{svc: svc}
}

// RegisterRoutes registers payment method routes on the given chi router.
// All routes require JWT authentication (handled by parent router group).
func (h *PaymentMethodHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/user/payment-methods", h.ListPaymentMethods)
	r.Post("/api/user/payment-methods/setup", h.CreateSetupIntent)
	r.Delete("/api/user/payment-methods/{id}", h.DetachPaymentMethod)
	r.Put("/api/user/payment-methods/{id}/default", h.SetDefaultPaymentMethod)
}

// ListPaymentMethods handles GET /api/user/payment-methods.
func (h *PaymentMethodHandler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return
	}

	methods, err := h.svc.ListPaymentMethods(r.Context(), userID)
	if err != nil {
		if errors.Is(err, payment.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "user not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to list payment methods",
		})
		return
	}

	writeJSON(w, http.StatusOK, methods)
}

// CreateSetupIntent handles POST /api/user/payment-methods/setup.
func (h *PaymentMethodHandler) CreateSetupIntent(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return
	}

	clientSecret, err := h.svc.CreateSetupIntent(r.Context(), userID)
	if err != nil {
		if errors.Is(err, payment.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "user not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to create setup intent",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"clientSecret": clientSecret,
	})
}

// DetachPaymentMethod handles DELETE /api/user/payment-methods/{id}.
func (h *PaymentMethodHandler) DetachPaymentMethod(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return
	}

	pmID := chi.URLParam(r, "id")
	if pmID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "payment method id is required",
		})
		return
	}

	err := h.svc.DetachPaymentMethod(r.Context(), userID, pmID)
	if err != nil {
		if errors.Is(err, payment.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "user not found",
			})
			return
		}
		if errors.Is(err, payment.ErrPaymentMethodNotOwned) {
			writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
				Code:    "FORBIDDEN",
				Message: "payment method does not belong to this user",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to detach payment method",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetDefaultPaymentMethod handles PUT /api/user/payment-methods/{id}/default.
func (h *PaymentMethodHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return
	}

	pmID := chi.URLParam(r, "id")
	if pmID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "payment method id is required",
		})
		return
	}

	err := h.svc.SetDefaultPaymentMethod(r.Context(), userID, pmID)
	if err != nil {
		if errors.Is(err, payment.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "user not found",
			})
			return
		}
		if errors.Is(err, payment.ErrPaymentMethodNotOwned) {
			writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
				Code:    "FORBIDDEN",
				Message: "payment method does not belong to this user",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to set default payment method",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "default_set",
	})
}
