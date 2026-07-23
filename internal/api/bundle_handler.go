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
	"github.com/lucatorrekens/bakery-app/internal/ws"
)

// BundleHandler handles HTTP requests related to surplus bundles.
type BundleHandler struct {
	svc        domain.BundleService
	bakeryRepo domain.BakeryRepository
	bundleRepo domain.BundleRepository
	wsHub      *ws.Hub
}

// NewBundleHandler creates a new BundleHandler.
func NewBundleHandler(svc domain.BundleService, bakeryRepo domain.BakeryRepository, bundleRepo domain.BundleRepository, wsHub *ws.Hub) *BundleHandler {
	return &BundleHandler{svc: svc, bakeryRepo: bakeryRepo, bundleRepo: bundleRepo, wsHub: wsHub}
}

// RegisterRoutes registers bundle-related routes on the given chi router.
func (h *BundleHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/bundles/impact", h.GetImpact)
	r.Get("/api/bundles", h.ListBundles)
	r.Get("/api/bundles/{id}", h.GetBundle)
	r.Post("/api/bundles", h.CreateBundle)
	r.Post("/api/bundles/{id}/publish", h.PublishBundle)
	r.Post("/api/bundles/{id}/reserve", h.ReserveBundle)
	r.Post("/api/bundles/{id}/reserve/confirm", h.ConfirmReservation)
	r.Delete("/api/bundle-reservations/{id}", h.CancelReservation)
}

// ListBundles handles GET /api/bundles.
func (h *BundleHandler) ListBundles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page := 1
	if p := query.Get("page"); p != "" {
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

	filters := domain.BundleFilters{}
	if t := query.Get("type"); t != "" {
		bt := domain.BundleType(t)
		filters.Type = &bt
	}
	if pb := query.Get("pickupBefore"); pb != "" {
		tod, err := parseTimeOfDayStr(pb)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_PICKUP_BEFORE",
				Message: "pickupBefore must be in HH:MM format",
			})
			return
		}
		filters.PickupBefore = &tod
	}

	result, err := h.svc.ListBundles(r.Context(), filters, domain.PaginationParams{
		Page:     page,
		PageSize: 20,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to list bundles",
		})
		return
	}

	items := make([]dto.BundleResponse, 0, len(result.Items))
	for _, b := range result.Items {
		resp := h.toBundleResponse(r, &b)
		items = append(items, resp)
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.BundleResponse]{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

// GetBundle handles GET /api/bundles/{id}.
func (h *BundleHandler) GetBundle(w http.ResponseWriter, r *http.Request) {
	bundleID := chi.URLParam(r, "id")
	if bundleID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bundle ID is required",
		})
		return
	}

	bundle, err := h.svc.GetBundle(r.Context(), bundleID)
	if err != nil {
		if errors.Is(err, service.ErrBundleNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "BUNDLE_NOT_FOUND",
				Message: "bundle not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch bundle",
		})
		return
	}

	writeJSON(w, http.StatusOK, h.toBundleResponse(r, bundle))
}

// CreateBundle handles POST /api/bundles.
func (h *BundleHandler) CreateBundle(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBundleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	sellerID := extractUserID(r)

	// Look up the bakery owned by this seller
	bakery, err := h.bakeryRepo.GetBakeryByOwner(r.Context(), sellerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to look up bakery",
		})
		return
	}
	if bakery == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "no bakery found for this seller",
		})
		return
	}

	// Parse pickup times
	pickupStart, err := parseTimeOfDayStr(req.PickupStartTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_PICKUP_WINDOW",
			Message: "invalid pickupStartTime: expected HH:MM format",
		})
		return
	}
	pickupEnd, err := parseTimeOfDayStr(req.PickupEndTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_PICKUP_WINDOW",
			Message: "invalid pickupEndTime: expected HH:MM format",
		})
		return
	}

	// Map DTO items to domain items
	items := make([]domain.BundleItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.BundleItem{
			ProductID:   item.ProductID,
			Description: item.Description,
			Quantity:    item.Quantity,
		}
	}

	bundle := domain.SurplusBundle{
		BakeryID:        bakery.ID,
		Name:            req.Name,
		Type:            domain.BundleType(req.Type),
		PhotoURL:        req.PhotoURL,
		Description:     req.Description,
		EstimatedValue:  req.EstimatedValue,
		OriginalPrice:   req.OriginalPrice,
		DiscountedPrice: req.DiscountedPrice,
		QuantityTotal:   req.QuantityTotal,
		PickupStartTime: pickupStart,
		PickupEndTime:   pickupEnd,
		Items:           items,
	}

	created, err := h.svc.CreateBundle(r.Context(), sellerID, bundle)
	if err != nil {
		// Check for validation errors (service wraps them with "validation:" prefix)
		if isValidationError(err) {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to create bundle",
		})
		return
	}

	writeJSON(w, http.StatusCreated, h.toBundleResponse(r, created))
}

// PublishBundle handles POST /api/bundles/{id}/publish.
func (h *BundleHandler) PublishBundle(w http.ResponseWriter, r *http.Request) {
	bundleID := chi.URLParam(r, "id")
	if bundleID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bundle ID is required",
		})
		return
	}

	sellerID := extractUserID(r)

	bundle, err := h.svc.PublishBundle(r.Context(), sellerID, bundleID)
	if err != nil {
		h.handleBundleError(w, err, "failed to publish bundle")
		return
	}

	writeJSON(w, http.StatusOK, h.toBundleResponse(r, bundle))
}

// ReserveBundle handles POST /api/bundles/{id}/reserve.
func (h *BundleHandler) ReserveBundle(w http.ResponseWriter, r *http.Request) {
	bundleID := chi.URLParam(r, "id")
	if bundleID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bundle ID is required",
		})
		return
	}

	customerID := extractUserID(r)

	reservation, err := h.svc.ReserveBundle(r.Context(), customerID, bundleID)
	if err != nil {
		h.handleBundleError(w, err, "failed to reserve bundle")
		return
	}

	// Fetch bundle name for the response
	bundleName := ""
	bundle, _ := h.svc.GetBundle(r.Context(), reservation.BundleID)
	if bundle != nil {
		bundleName = bundle.Name
	}

	writeJSON(w, http.StatusCreated, dto.BundleReservationResponse{
		ID:         reservation.ID,
		BundleID:   reservation.BundleID,
		BundleName: bundleName,
		Status:     string(reservation.Status),
		CreatedAt:  reservation.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ConfirmReservation handles POST /api/bundles/{id}/reserve/confirm.
// The {id} is the bundle ID; the handler looks up the active reservation for this user+bundle.
func (h *BundleHandler) ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	bundleID := chi.URLParam(r, "id")
	if bundleID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bundle ID is required",
		})
		return
	}

	customerID := extractUserID(r)

	// Look up the active reservation for this user on this bundle
	activeReservation, err := h.bundleRepo.GetActiveReservation(r.Context(), customerID, bundleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to confirm reservation",
		})
		return
	}
	if activeReservation == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "RESERVATION_NOT_FOUND",
			Message: "no active reservation found for this bundle",
		})
		return
	}

	reservation, err := h.svc.ConfirmReservation(r.Context(), customerID, activeReservation.ID)
	if err != nil {
		h.handleBundleError(w, err, "failed to confirm reservation")
		return
	}

	bundleName := ""
	bundle, _ := h.svc.GetBundle(r.Context(), reservation.BundleID)
	if bundle != nil {
		bundleName = bundle.Name
	}

	writeJSON(w, http.StatusOK, dto.BundleReservationResponse{
		ID:         reservation.ID,
		BundleID:   reservation.BundleID,
		BundleName: bundleName,
		Status:     string(reservation.Status),
		CreatedAt:  reservation.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// CancelReservation handles DELETE /api/bundle-reservations/{id}.
func (h *BundleHandler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "id")
	if reservationID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "reservation ID is required",
		})
		return
	}

	customerID := extractUserID(r)

	err := h.svc.CancelReservation(r.Context(), customerID, reservationID)
	if err != nil {
		h.handleBundleError(w, err, "failed to cancel reservation")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "reservation cancelled successfully",
	})
}

// GetImpact handles GET /api/bundles/impact.
func (h *BundleHandler) GetImpact(w http.ResponseWriter, r *http.Request) {
	impact, err := h.svc.GetImpact(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch impact metrics",
		})
		return
	}

	writeJSON(w, http.StatusOK, dto.BundleImpactResponse{
		TotalSaved:    impact.TotalSaved,
		WeightAvoided: impact.WeightAvoided,
	})
}

// toBundleResponse converts a domain SurplusBundle to a BundleResponse DTO,
// enriching it with bakery name and coordinates.
func (h *BundleHandler) toBundleResponse(r *http.Request, bundle *domain.SurplusBundle) dto.BundleResponse {
	var bakeryName string
	var bakeryLat, bakeryLng float64

	bakery, _ := h.bakeryRepo.GetBakery(r.Context(), bundle.BakeryID)
	if bakery != nil {
		bakeryName = bakery.Name
		bakeryLat = bakery.Latitude
		bakeryLng = bakery.Longitude
	}

	items := make([]dto.BundleItemResponse, len(bundle.Items))
	for i, item := range bundle.Items {
		items[i] = dto.BundleItemResponse{
			ProductID:   item.ProductID,
			Description: item.Description,
			Quantity:    item.Quantity,
		}
	}

	return dto.BundleResponse{
		ID:                bundle.ID,
		BakeryID:          bundle.BakeryID,
		BakeryName:        bakeryName,
		BakeryLatitude:    bakeryLat,
		BakeryLongitude:   bakeryLng,
		Name:              bundle.Name,
		Type:              string(bundle.Type),
		PhotoURL:          bundle.PhotoURL,
		Description:       bundle.Description,
		EstimatedValue:    bundle.EstimatedValue,
		OriginalPrice:     bundle.OriginalPrice,
		DiscountedPrice:   bundle.DiscountedPrice,
		QuantityTotal:     bundle.QuantityTotal,
		QuantityRemaining: bundle.QuantityRemaining,
		PickupStartTime:   bundle.PickupStartTime.String(),
		PickupEndTime:     bundle.PickupEndTime.String(),
		PublishedDate:     bundle.PublishedDate,
		ExpiresAt:         bundle.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		Status:            string(bundle.Status),
		Items:             items,
		CreatedAt:         bundle.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// handleBundleError maps service-layer errors to appropriate HTTP responses.
func (h *BundleHandler) handleBundleError(w http.ResponseWriter, err error, fallbackMsg string) {
	switch {
	case errors.Is(err, service.ErrBundleNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BUNDLE_NOT_FOUND",
			Message: "bundle not found",
		})
	case errors.Is(err, service.ErrBundleSoldOut):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "BUNDLE_SOLD_OUT",
			Message: "bundle is sold out",
		})
	case errors.Is(err, service.ErrBundleNotDraft):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "BUNDLE_NOT_DRAFT",
			Message: "bundle is not in draft status",
		})
	case errors.Is(err, service.ErrReservationExists):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "RESERVATION_EXISTS",
			Message: "active reservation already exists for this bundle",
		})
	case errors.Is(err, service.ErrBundleReservationNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "RESERVATION_NOT_FOUND",
			Message: "bundle reservation not found",
		})
	case errors.Is(err, service.ErrBundleReservationNotCancellable):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "RESERVATION_NOT_CANCELLABLE",
			Message: "reservation cannot be cancelled in its current state",
		})
	case errors.Is(err, service.ErrForbidden):
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "you do not have permission for this action",
		})
	case errors.Is(err, service.ErrBakeryNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "bakery not found",
		})
	case isValidationError(err):
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
	default:
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: fallbackMsg,
		})
	}
}

// isValidationError checks if an error is a validation error from the service layer.
func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	// The bundle service wraps validation errors with "validation:" prefix
	return len(err.Error()) > 11 && err.Error()[:11] == "validation:"
}
