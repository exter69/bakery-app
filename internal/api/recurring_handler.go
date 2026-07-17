package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// RecurringHandler handles HTTP requests related to recurring orders.
type RecurringHandler struct {
	svc *service.RecurringOrderService
}

// NewRecurringHandler creates a new RecurringHandler.
func NewRecurringHandler(svc *service.RecurringOrderService) *RecurringHandler {
	return &RecurringHandler{svc: svc}
}

// RegisterRoutes registers recurring order routes on the given chi router.
func (h *RecurringHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/recurring-orders", h.Create)
	r.Get("/api/recurring-orders", h.List)
	r.Put("/api/recurring-orders/{id}/pause", h.Pause)
	r.Put("/api/recurring-orders/{id}/resume", h.Resume)
	r.Delete("/api/recurring-orders/{id}", h.Delete)
}

// Create handles POST /api/recurring-orders.
func (h *RecurringHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRecurringOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	userID := extractUserID(r)

	scheduledTime, err := parseTimeSlotDTO(req.ScheduledTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	order := domain.RecurringOrder{
		BakeryID:      req.BakeryID,
		Items:         items,
		ScheduledDay:  req.ScheduledDay,
		ScheduledTime: scheduledTime,
		Frequency:     req.Frequency,
		SelectionMode: req.SelectionMode,
	}

	created, err := h.svc.CreateRecurringOrder(r.Context(), userID, order)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, recurringOrderToResponse(created))
}

// List handles GET /api/recurring-orders.
func (h *RecurringHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

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

	result, err := h.svc.ListMyRecurringOrders(r.Context(), userID, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to list recurring orders",
		})
		return
	}

	items := make([]dto.RecurringOrderResponse, len(result.Items))
	for i := range result.Items {
		items[i] = recurringOrderToResponse(&result.Items[i])
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.RecurringOrderResponse]{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

// Pause handles PUT /api/recurring-orders/{id}/pause.
func (h *RecurringHandler) Pause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := extractUserID(r)

	if err := h.svc.PauseRecurringOrder(r.Context(), id, userID); err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "recurring order paused"})
}

// Resume handles PUT /api/recurring-orders/{id}/resume.
func (h *RecurringHandler) Resume(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := extractUserID(r)

	if err := h.svc.ResumeRecurringOrder(r.Context(), id, userID); err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "recurring order resumed"})
}

// Delete handles DELETE /api/recurring-orders/{id}.
func (h *RecurringHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := extractUserID(r)

	if err := h.svc.DeleteRecurringOrder(r.Context(), id, userID); err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "recurring order deleted"})
}

// handleServiceError maps service errors to HTTP responses for recurring orders.
func (h *RecurringHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrRecurringOrderNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "RECURRING_ORDER_NOT_FOUND",
			Message: "recurring order not found",
		})
	case errors.Is(err, service.ErrBakeryNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "bakery not found",
		})
	case errors.Is(err, service.ErrForbidden):
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "you do not own this recurring order",
		})
	case errors.Is(err, service.ErrRecurringOrderNoItems):
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
	case errors.Is(err, service.ErrInvalidFrequency):
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
	case errors.Is(err, service.ErrInvalidSelectionMode):
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
	default:
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		})
	}
}

// recurringOrderToResponse converts a domain RecurringOrder to its DTO response.
func recurringOrderToResponse(o *domain.RecurringOrder) dto.RecurringOrderResponse {
	items := make([]dto.OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
		}
	}

	return dto.RecurringOrderResponse{
		ID:       o.ID,
		BakeryID: o.BakeryID,
		Items:    items,
		ScheduledDay:  string(o.ScheduledDay),
		ScheduledTime: dto.TimeSlotResponse{
			StartTime: o.ScheduledTime.StartTime.String(),
			EndTime:   o.ScheduledTime.EndTime.String(),
		},
		Frequency:     string(o.Frequency),
		SelectionMode: string(o.SelectionMode),
		Active:        o.Active,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}
}
