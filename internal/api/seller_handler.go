package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// SellerHandler handles HTTP requests for seller portal operations.
type SellerHandler struct {
	svc *service.SellerService
}

// NewSellerHandler creates a new SellerHandler.
func NewSellerHandler(svc *service.SellerService) *SellerHandler {
	return &SellerHandler{svc: svc}
}

// RegisterRoutes registers all seller routes on the given chi router.
// All routes require JWT auth (handled by the parent group).
func (h *SellerHandler) RegisterRoutes(r chi.Router) {
	// Bakery management
	r.Put("/api/bakeries/{id}", h.UpdateBakery)
	r.Put("/api/bakeries/{id}/schedule", h.UpdateBakerySchedule)

	// Product management
	r.Post("/api/bakeries/{id}/products", h.CreateProduct)
	r.Put("/api/products/{id}", h.UpdateProduct)
	r.Delete("/api/products/{id}", h.DeleteProduct)

	// Order/Reservation views
	r.Get("/api/bakeries/{id}/orders", h.ListBakeryOrders)
	r.Get("/api/bakeries/{id}/reservations", h.ListBakeryReservations)
	r.Put("/api/orders/{id}/status", h.UpdateOrderStatus)
	r.Put("/api/reservations/{id}/status", h.UpdateReservationStatus)
}

// requireSeller checks that the user has role=1 (seller).
// Returns the userID if valid, or writes an error response and returns empty string.
func requireSeller(w http.ResponseWriter, r *http.Request) string {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return ""
	}

	role := middleware.GetUserRoleFromContext(r.Context())
	if role != int(domain.RoleSeller) {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "seller role required",
		})
		return ""
	}

	return userID
}

// handleSellerError maps common service errors to HTTP responses.
func handleSellerError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrBakeryNotFound) {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "bakery not found",
		})
		return
	}
	if errors.Is(err, service.ErrProductNotFound) {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "PRODUCT_NOT_FOUND",
			Message: "product not found",
		})
		return
	}
	if errors.Is(err, service.ErrOrderNotFound) {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "ORDER_NOT_FOUND",
			Message: "order not found",
		})
		return
	}
	if errors.Is(err, service.ErrReservationNotFound) {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "RESERVATION_NOT_FOUND",
			Message: "reservation not found",
		})
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "you do not own this resource",
		})
		return
	}
	if errors.Is(err, service.ErrInvalidStatusTransition) {
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "INVALID_STATUS_TRANSITION",
			Message: "invalid status transition",
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "an unexpected error occurred",
	})
}

// UpdateBakery handles PUT /api/bakeries/{id}.
func (h *SellerHandler) UpdateBakery(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	bakeryID := chi.URLParam(r, "id")
	if bakeryID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bakery ID is required",
		})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Address     *string `json:"address"`
		PhotoURL    *string `json:"photoUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	bakery, err := h.svc.UpdateBakery(r.Context(), bakeryID, userID, req.Name, req.Description, req.Address, req.PhotoURL)
	if err != nil {
		handleSellerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, bakery)
}

// UpdateBakerySchedule handles PUT /api/bakeries/{id}/schedule.
func (h *SellerHandler) UpdateBakerySchedule(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	bakeryID := chi.URLParam(r, "id")
	if bakeryID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bakery ID is required",
		})
		return
	}

	var req struct {
		Schedule []domain.DaySchedule `json:"schedule"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	bakery, err := h.svc.UpdateBakerySchedule(r.Context(), bakeryID, userID, req.Schedule)
	if err != nil {
		handleSellerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, bakery)
}

// CreateProduct handles POST /api/bakeries/{id}/products.
func (h *SellerHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	bakeryID := chi.URLParam(r, "id")
	if bakeryID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bakery ID is required",
		})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       int64  `json:"price"`
		PhotoURL    string `json:"photoUrl"`
		Category    string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "product name is required",
		})
		return
	}

	product := domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		PhotoURL:    req.PhotoURL,
		Category:    req.Category,
	}

	created, err := h.svc.CreateProduct(r.Context(), bakeryID, userID, product)
	if err != nil {
		handleSellerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// UpdateProduct handles PUT /api/products/{id}.
func (h *SellerHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	productID := chi.URLParam(r, "id")
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "product ID is required",
		})
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	product, err := h.svc.UpdateProduct(r.Context(), productID, userID, updates)
	if err != nil {
		handleSellerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, product)
}

// DeleteProduct handles DELETE /api/products/{id}.
func (h *SellerHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	productID := chi.URLParam(r, "id")
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "product ID is required",
		})
		return
	}

	err := h.svc.DeleteProduct(r.Context(), productID, userID)
	if err != nil {
		handleSellerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "product deleted successfully",
	})
}

// ListBakeryOrders handles GET /api/bakeries/{id}/orders.
func (h *SellerHandler) ListBakeryOrders(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	bakeryID := chi.URLParam(r, "id")
	if bakeryID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bakery ID is required",
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
		PageSize: 20,
	}

	result, err := h.svc.ListBakeryOrders(r.Context(), bakeryID, userID, filters, params)
	if err != nil {
		handleSellerError(w, err)
		return
	}

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

// ListBakeryReservations handles GET /api/bakeries/{id}/reservations.
func (h *SellerHandler) ListBakeryReservations(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	bakeryID := chi.URLParam(r, "id")
	if bakeryID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bakery ID is required",
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

	var filters domain.ReservationFilters
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.ReservationStatus(statusStr)
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
		PageSize: 20,
	}

	result, err := h.svc.ListBakeryReservations(r.Context(), bakeryID, userID, filters, params)
	if err != nil {
		handleSellerError(w, err)
		return
	}

	items := make([]dto.ReservationResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toReservationResponse(&result.Items[i])
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ReservationResponse]{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

// UpdateOrderStatus handles PUT /api/orders/{id}/status.
func (h *SellerHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "order ID is required",
		})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	if req.Status == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "status is required",
		})
		return
	}

	order, err := h.svc.UpdateOrderStatus(r.Context(), orderID, userID, domain.OrderStatus(req.Status))
	if err != nil {
		handleSellerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, orderToResponse(order, nil))
}

// UpdateReservationStatus handles PUT /api/reservations/{id}/status.
func (h *SellerHandler) UpdateReservationStatus(w http.ResponseWriter, r *http.Request) {
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	reservationID := chi.URLParam(r, "id")
	if reservationID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "reservation ID is required",
		})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	if req.Status == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "status is required",
		})
		return
	}

	reservation, err := h.svc.UpdateReservationStatus(r.Context(), reservationID, userID, domain.ReservationStatus(req.Status))
	if err != nil {
		handleSellerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toReservationResponse(reservation))
}
