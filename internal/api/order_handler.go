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

// OrderHandler handles HTTP requests related to orders.
type OrderHandler struct {
	svc domain.OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(svc domain.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// RegisterRoutes registers order-related routes on the given chi router.
func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/orders", h.CreateOrder)
	r.Get("/api/orders", h.ListOrders)
	r.Delete("/api/orders/{id}", h.DeleteOrder)
}

// CreateOrder handles POST /api/orders.
// It parses the request, converts DTOs to domain types, calls the service, and returns the result.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	// Extract userID from request (mock for now, real auth comes later)
	userID := extractUserID(r)

	// Convert request DTO to domain types
	order, err := requestToOrder(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// Call service
	created, paymentLink, err := h.svc.CreateOrder(r.Context(), userID, order)
	if err != nil {
		h.handleCreateOrderError(w, err)
		return
	}

	// Build response
	resp := orderToResponse(created, paymentLink)
	writeJSON(w, http.StatusCreated, resp)
}

// handleCreateOrderError maps service errors to HTTP responses.
func (h *OrderHandler) handleCreateOrderError(w http.ResponseWriter, err error) {
	var valErrs *service.ValidationErrors
	if errors.As(err, &valErrs) {
		writeJSON(w, http.StatusUnprocessableEntity, dto.ValidationErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "order validation failed",
			Errors:  toValidationErrorDTOs(valErrs.Errors),
		})
		return
	}

	if errors.Is(err, service.ErrBakeryNotFound) {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "bakery not found",
		})
		return
	}

	writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "failed to create order",
	})
}

// requestToOrder converts a CreateOrderRequest DTO to a domain Order.
func requestToOrder(req dto.CreateOrderRequest) (domain.Order, error) {
	scheduledTime, err := parseTimeSlotDTO(req.ScheduledTime)
	if err != nil {
		return domain.Order{}, err
	}

	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	selectionMode := domain.SelectionMode(req.SelectionMode)
	if selectionMode == "" {
		selectionMode = domain.SelectionFixed
	}

	return domain.Order{
		BakeryID:      req.BakeryID,
		Items:         items,
		ScheduledDay:  req.ScheduledDay,
		ScheduledTime: scheduledTime,
		SelectionMode: selectionMode,
	}, nil
}

// orderToResponse converts a domain Order and payment link to an OrderResponse DTO.
func orderToResponse(order *domain.Order, paymentLink *domain.PaymentLink) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
		}
	}

	resp := dto.OrderResponse{
		ID:       order.ID,
		BakeryID: order.BakeryID,
		Items:    items,
		ScheduledDay: string(order.ScheduledDay),
		ScheduledTime: dto.TimeSlotResponse{
			StartTime: order.ScheduledTime.StartTime.String(),
			EndTime:   order.ScheduledTime.EndTime.String(),
		},
		Status:        string(order.Status),
		TotalAmount:   order.TotalAmount,
		PaymentMethod: string(order.PaymentMethod),
		SelectionMode: string(order.SelectionMode),
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	if paymentLink != nil {
		resp.PaymentLink = paymentLink.URL
	}

	return resp
}

const defaultOrderPageSize = 20

// ListOrders handles GET /api/orders.
// Returns a paginated list of orders for the authenticated user with optional filtering and sorting.
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	// Parse pagination
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

	// Parse filters
	var filters domain.OrderFilters
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.OrderStatus(statusStr)
		filters.Status = &status
	}
	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		filters.SortBy = sortBy
	}
	if sortDir := r.URL.Query().Get("sortDir"); sortDir != "" {
		filters.SortDir = sortDir
	}

	params := domain.PaginationParams{
		Page:     page,
		PageSize: defaultOrderPageSize,
	}

	result, err := h.svc.GetOrders(r.Context(), userID, filters, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to list orders",
		})
		return
	}

	// Convert domain orders to response DTOs
	items := make([]dto.OrderResponse, len(result.Items))
	for i := range result.Items {
		items[i] = orderToResponse(&result.Items[i], nil)
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.OrderResponse]{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

// DeleteOrder handles DELETE /api/orders/{id}.
// Cancels an order after verifying ownership and state.
func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "order ID is required",
		})
		return
	}

	userID := extractUserID(r)

	err := h.svc.DeleteOrder(r.Context(), orderID, userID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "ORDER_NOT_FOUND",
				Message: "order not found",
			})
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
				Code:    "FORBIDDEN",
				Message: "you do not own this order",
			})
			return
		}
		if errors.Is(err, service.ErrOrderNotCancellable) {
			writeJSON(w, http.StatusConflict, dto.ErrorResponse{
				Code:    "ORDER_NOT_CANCELLABLE",
				Message: "order cannot be cancelled in its current state",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to delete order",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "order cancelled successfully",
	})
}
