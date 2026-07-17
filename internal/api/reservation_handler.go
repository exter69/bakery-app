package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// ReservationHandler handles HTTP requests related to reservations.
type ReservationHandler struct {
	svc domain.ReservationService
}

// NewReservationHandler creates a new ReservationHandler.
func NewReservationHandler(svc domain.ReservationService) *ReservationHandler {
	return &ReservationHandler{svc: svc}
}

// RegisterRoutes registers reservation-related routes on the given chi router.
func (h *ReservationHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/reservations", h.CreateReservation)
	r.Delete("/api/reservations/{id}", h.DeleteReservation)
}

// CreateReservation handles POST /api/reservations.
// Creates a new reservation with on-spot payment.
func (h *ReservationHandler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	// Parse time slots
	scheduledTime, err := parseTimeSlotDTO(req.ScheduledTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// Convert request DTO to domain types
	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	reservation := domain.Reservation{
		BakeryID:      req.BakeryID,
		Items:         items,
		ScheduledDay:  req.ScheduledDay,
		ScheduledTime: scheduledTime,
	}

	// Extract userID from request (mock for now, auth comes later)
	userID := extractUserID(r)

	result, err := h.svc.CreateReservation(r.Context(), userID, reservation)
	if err != nil {
		var validationErrs *service.ValidationErrors
		if errors.As(err, &validationErrs) {
			writeJSON(w, http.StatusUnprocessableEntity, dto.ValidationErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: "reservation validation failed",
				Errors:  toValidationErrorDTOs(validationErrs.Errors),
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
			Message: "failed to create reservation",
		})
		return
	}

	resp := toReservationResponse(result)
	writeJSON(w, http.StatusCreated, resp)
}

// toReservationResponse converts a domain.Reservation to a dto.ReservationResponse.
func toReservationResponse(r *domain.Reservation) dto.ReservationResponse {
	items := make([]dto.OrderItemResponse, len(r.Items))
	for i, item := range r.Items {
		items[i] = dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
		}
	}
	return dto.ReservationResponse{
		ID:       r.ID,
		BakeryID: r.BakeryID,
		Items:    items,
		ScheduledDay: string(r.ScheduledDay),
		ScheduledTime: dto.TimeSlotResponse{
			StartTime: r.ScheduledTime.StartTime.String(),
			EndTime:   r.ScheduledTime.EndTime.String(),
		},
		Status:        string(r.Status),
		TotalAmount:   r.TotalAmount,
		PaymentMethod: string(r.PaymentMethod),
		CreatedAt:     r.CreatedAt,
	}
}

// DeleteReservation handles DELETE /api/reservations/{id}.
// Cancels a reservation after verifying ownership and state.
func (h *ReservationHandler) DeleteReservation(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "id")
	if reservationID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "reservation ID is required",
		})
		return
	}

	userID := extractUserID(r)

	err := h.svc.DeleteReservation(r.Context(), reservationID, userID)
	if err != nil {
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
				Message: "you do not own this reservation",
			})
			return
		}
		if errors.Is(err, service.ErrReservationNotCancellable) {
			writeJSON(w, http.StatusConflict, dto.ErrorResponse{
				Code:    "RESERVATION_NOT_CANCELLABLE",
				Message: "reservation cannot be cancelled in its current state",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to delete reservation",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "reservation cancelled successfully",
	})
}
