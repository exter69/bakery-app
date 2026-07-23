package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// PayoutHandler handles HTTP requests for payout/Connect operations.
type PayoutHandler struct {
	payoutSvc *service.PayoutService
	bakeryRepo domain.BakeryRepository
}

// NewPayoutHandler creates a new PayoutHandler.
func NewPayoutHandler(payoutSvc *service.PayoutService, bakeryRepo domain.BakeryRepository) *PayoutHandler {
	return &PayoutHandler{
		payoutSvc:  payoutSvc,
		bakeryRepo: bakeryRepo,
	}
}

// RegisterRoutes registers payout routes on the given chi router.
// All routes require JWT auth (handled by the parent group).
func (h *PayoutHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/seller/payouts", h.ListPayouts)
	r.Post("/api/seller/connect/onboard", h.Onboard)
	r.Get("/api/seller/connect/status", h.GetConnectStatus)
}

// ListPayouts handles GET /api/seller/payouts.
func (h *PayoutHandler) ListPayouts(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	bakery, err := h.bakeryRepo.GetBakeryByOwner(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch bakery",
		})
		return
	}
	if bakery == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "you don't have a bakery yet",
		})
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil || parsed < 1 {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_PAGE",
				Message: "page must be a positive integer",
			})
			return
		}
		page = parsed
	}

	params := domain.PaginationParams{
		Page:     page,
		PageSize: 20,
	}

	result, err := h.payoutSvc.ListPayouts(r.Context(), bakery.ID, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to list payouts",
		})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Onboard handles POST /api/seller/connect/onboard.
func (h *PayoutHandler) Onboard(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	bakery, err := h.bakeryRepo.GetBakeryByOwner(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch bakery",
		})
		return
	}
	if bakery == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "you don't have a bakery yet",
		})
		return
	}

	var req struct {
		RefreshURL string `json:"refreshUrl"`
		ReturnURL  string `json:"returnUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	if req.RefreshURL == "" || req.ReturnURL == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "refreshUrl and returnUrl are required",
		})
		return
	}

	link, err := h.payoutSvc.Onboard(r.Context(), bakery.ID, req.RefreshURL, req.ReturnURL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "CONNECT_ERROR",
			Message: "failed to create onboarding link",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"url": link,
	})
}

// GetConnectStatus handles GET /api/seller/connect/status.
func (h *PayoutHandler) GetConnectStatus(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	bakery, err := h.bakeryRepo.GetBakeryByOwner(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch bakery",
		})
		return
	}
	if bakery == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "you don't have a bakery yet",
		})
		return
	}

	connected, chargesEnabled, payoutsEnabled, err := h.payoutSvc.GetConnectStatus(r.Context(), bakery.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "CONNECT_ERROR",
			Message: "failed to check connect status",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"connected":      connected,
		"chargesEnabled": chargesEnabled,
		"payoutsEnabled": payoutsEnabled,
	})
}
