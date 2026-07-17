package api

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

const defaultBakeryPageSize = 50

// BakeryHandler handles HTTP requests related to bakeries.
type BakeryHandler struct {
	svc domain.BakeryService
	now func() time.Time // injectable clock for testing
}

// NewBakeryHandler creates a new BakeryHandler.
func NewBakeryHandler(svc domain.BakeryService) *BakeryHandler {
	return &BakeryHandler{svc: svc, now: time.Now}
}

// NewBakeryHandlerWithClock creates a new BakeryHandler with an injectable clock.
func NewBakeryHandlerWithClock(svc domain.BakeryService, now func() time.Time) *BakeryHandler {
	return &BakeryHandler{svc: svc, now: now}
}

// RegisterRoutes registers bakery-related routes on the given chi router.
func (h *BakeryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/bakeries", h.ListBakeries)
	r.Get("/api/bakeries/{id}/menu", h.GetMenu)
}

// ListBakeries handles GET /api/bakeries with pagination (50 per page).
// Optionally accepts lat, lng, radius query params for location-based sorting.
func (h *BakeryHandler) ListBakeries(w http.ResponseWriter, r *http.Request) {
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

	params := domain.BakeryListParams{
		PaginationParams: domain.PaginationParams{
			Page:     page,
			PageSize: defaultBakeryPageSize,
		},
	}

	// Parse optional lat/lng/radius
	if latStr := r.URL.Query().Get("lat"); latStr != "" {
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil || lat < -90 || lat > 90 {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_LAT",
				Message: "lat must be a valid latitude between -90 and 90",
			})
			return
		}
		params.Lat = &lat
	}

	if lngStr := r.URL.Query().Get("lng"); lngStr != "" {
		lng, err := strconv.ParseFloat(lngStr, 64)
		if err != nil || lng < -180 || lng > 180 {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_LNG",
				Message: "lng must be a valid longitude between -180 and 180",
			})
			return
		}
		params.Lng = &lng
	}

	if radiusStr := r.URL.Query().Get("radius"); radiusStr != "" {
		radius, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil || radius <= 0 {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_RADIUS",
				Message: "radius must be a positive number (in km)",
			})
			return
		}
		params.Radius = radius
	}

	result, err := h.svc.ListBakeries(r.Context(), params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to list bakeries",
		})
		return
	}

	today := timeToDayOfWeek(h.now())
	cards := make([]dto.BakeryCardResponse, 0, len(result.Items))
	for _, item := range result.Items {
		card := toBakeryCardResponse(item.Bakery, today)
		if item.Distance != nil {
			rounded := math.Round(*item.Distance*10) / 10
			card.Distance = &rounded
		}
		cards = append(cards, card)
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.BakeryCardResponse]{
		Items:    cards,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

// toBakeryCardResponse converts a domain Bakery to a BakeryCardResponse with today's schedule.
func toBakeryCardResponse(bakery domain.Bakery, today domain.DayOfWeek) dto.BakeryCardResponse {
	schedule := resolveTodaySchedule(bakery.Schedule, today)
	return dto.BakeryCardResponse{
		ID:            bakery.ID,
		Name:          bakery.Name,
		PhotoURL:      bakery.PhotoURL,
		TodaySchedule: schedule,
	}
}

// resolveTodaySchedule finds today's schedule entry and returns the response.
// If no schedule entry matches today, the bakery is considered closed.
func resolveTodaySchedule(schedule []domain.DaySchedule, today domain.DayOfWeek) dto.TodayScheduleResponse {
	for _, ds := range schedule {
		if ds.Day == today {
			if !ds.IsOpen {
				return dto.TodayScheduleResponse{IsOpen: false}
			}
			return dto.TodayScheduleResponse{
				OpenTime:  ds.OpenTime.String(),
				CloseTime: ds.CloseTime.String(),
				IsOpen:    true,
			}
		}
	}
	// No schedule entry for today means closed
	return dto.TodayScheduleResponse{IsOpen: false}
}

// timeToDayOfWeek converts a time.Time weekday to a domain.DayOfWeek.
func timeToDayOfWeek(t time.Time) domain.DayOfWeek {
	switch t.Weekday() {
	case time.Monday:
		return domain.Monday
	case time.Tuesday:
		return domain.Tuesday
	case time.Wednesday:
		return domain.Wednesday
	case time.Thursday:
		return domain.Thursday
	case time.Friday:
		return domain.Friday
	case time.Saturday:
		return domain.Saturday
	case time.Sunday:
		return domain.Sunday
	default:
		return domain.Monday
	}
}

// GetMenu handles GET /api/bakeries/{id}/menu.
// Returns products grouped by category for the given bakery.
func (h *BakeryHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	bakeryID := chi.URLParam(r, "id")
	if bakeryID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bakery ID is required",
		})
		return
	}

	menu, err := h.svc.GetMenu(r.Context(), bakeryID)
	if err != nil {
		if errors.Is(err, service.ErrBakeryNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "BAKERY_NOT_FOUND",
				Message: "bakery not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch menu",
		})
		return
	}

	writeJSON(w, http.StatusOK, menu)
}

// writeJSON marshals v as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
