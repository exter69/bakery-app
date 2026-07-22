package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/invoice"
)

// InvoiceHandler handles HTTP requests for invoice retrieval.
type InvoiceHandler struct {
	store     *invoice.Store
	orderRepo domain.OrderRepository
}

// NewInvoiceHandler creates a new InvoiceHandler.
func NewInvoiceHandler(store *invoice.Store, orderRepo domain.OrderRepository) *InvoiceHandler {
	return &InvoiceHandler{store: store, orderRepo: orderRepo}
}

// RegisterRoutes registers invoice-related routes on the given chi router.
func (h *InvoiceHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/orders/{id}/invoice", h.GetInvoice)
}

// GetInvoice handles GET /api/orders/{id}/invoice.
// Returns the HTML invoice for the specified order. Auth required (owner only).
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "order ID is required",
		})
		return
	}

	userID := extractUserID(r)

	// Verify order exists and belongs to the authenticated user
	order, err := h.orderRepo.GetByID(r.Context(), orderID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to retrieve order",
		})
		return
	}
	if order == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "ORDER_NOT_FOUND",
			Message: "order not found",
		})
		return
	}

	if order.UserID != userID {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "you do not own this order",
		})
		return
	}

	// Retrieve stored invoice
	html, found := h.store.Get(orderID)
	if !found {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "INVOICE_NOT_FOUND",
			Message: "no invoice available for this order",
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
