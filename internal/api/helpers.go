package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/lucatorrekens/bakery-app/internal/validation"
)

// parseTimeOfDayStr parses a "HH:MM" string into a domain.TimeOfDay.
func parseTimeOfDayStr(s string) (domain.TimeOfDay, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return domain.TimeOfDay{}, fmt.Errorf("expected HH:MM format, got %q", s)
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return domain.TimeOfDay{}, fmt.Errorf("invalid hour in %q", s)
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return domain.TimeOfDay{}, fmt.Errorf("invalid minute in %q", s)
	}
	return domain.TimeOfDay{Hour: hour, Minute: minute}, nil
}

// parseTimeSlotDTO parses a TimeSlotRequest (HH:MM strings) into a domain.TimeSlot.
func parseTimeSlotDTO(ts dto.TimeSlotRequest) (domain.TimeSlot, error) {
	start, err := parseTimeOfDayStr(ts.StartTime)
	if err != nil {
		return domain.TimeSlot{}, fmt.Errorf("invalid startTime: %w", err)
	}
	end, err := parseTimeOfDayStr(ts.EndTime)
	if err != nil {
		return domain.TimeSlot{}, fmt.Errorf("invalid endTime: %w", err)
	}
	return domain.TimeSlot{StartTime: start, EndTime: end}, nil
}

// toValidationErrorDTOs converts validation errors to DTO format.
func toValidationErrorDTOs(errs []validation.ValidationError) []dto.ValidationErrorDetail {
	details := make([]dto.ValidationErrorDetail, len(errs))
	for i, ve := range errs {
		details[i] = dto.ValidationErrorDetail{
			Field:   ve.Field,
			Message: ve.Message,
		}
	}
	return details
}

// extractUserID extracts the user ID from the request.
// It checks the context first (set by JWT auth middleware), then falls back to
// the X-User-ID header for backward compatibility with tests.
func extractUserID(r *http.Request) string {
	if uid := middleware.GetUserIDFromContext(r.Context()); uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	return "anonymous"
}
