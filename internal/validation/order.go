package validation

import (
	"fmt"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// ValidationError represents a single field-level validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult holds the outcome of a validation pass.
type ValidationResult struct {
	Errors []ValidationError `json:"errors,omitempty"`
}

// HasErrors reports whether any validation errors were found.
func (r ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// Merge combines another ValidationResult into this one.
func (r *ValidationResult) Merge(other ValidationResult) {
	r.Errors = append(r.Errors, other.Errors...)
}

// ValidateOrderItemsQuantity validates that the items list is non-empty and each
// item has a valid quantity (1-999). This is used before price enrichment from
// the product catalog.
func ValidateOrderItemsQuantity(items []domain.OrderItem) ValidationResult {
	result := ValidationResult{}

	if len(items) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "items",
			Message: "at least one item is required",
		})
		return result
	}

	for i, item := range items {
		field := fmt.Sprintf("items[%d]", i)

		if item.Quantity < 1 || item.Quantity > 999 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   field + ".quantity",
				Message: fmt.Sprintf("quantity must be between 1 and 999, got %d", item.Quantity),
			})
		}
	}

	return result
}

// ValidateOrderItems validates that the items list is non-empty and each item
// has a valid quantity (1-999) and positive unit price (> 0).
func ValidateOrderItems(items []domain.OrderItem) ValidationResult {
	result := ValidationResult{}

	if len(items) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "items",
			Message: "at least one item is required",
		})
		return result
	}

	for i, item := range items {
		field := fmt.Sprintf("items[%d]", i)

		if item.Quantity < 1 || item.Quantity > 999 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   field + ".quantity",
				Message: fmt.Sprintf("quantity must be between 1 and 999, got %d", item.Quantity),
			})
		}

		if item.UnitPrice <= 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   field + ".unitPrice",
				Message: fmt.Sprintf("unit price must be greater than 0, got %d", item.UnitPrice),
			})
		}
	}

	return result
}

// ValidateSchedule validates that the bakery is open on the selected day and
// that the time slot falls within operating hours.
func ValidateSchedule(scheduledDay domain.DayOfWeek, scheduledTime domain.TimeSlot, schedule []domain.DaySchedule) ValidationResult {
	result := ValidationResult{}

	// Find the schedule entry for the selected day.
	var daySchedule *domain.DaySchedule
	for i := range schedule {
		if schedule[i].Day == scheduledDay {
			daySchedule = &schedule[i]
			break
		}
	}

	if daySchedule == nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "scheduledDay",
			Message: fmt.Sprintf("no schedule found for %s", scheduledDay),
		})
		return result
	}

	if !daySchedule.IsOpen {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "scheduledDay",
			Message: fmt.Sprintf("bakery is closed on %s", scheduledDay),
		})
		return result
	}

	// Check that the time slot falls within operating hours.
	if scheduledTime.StartTime.Before(daySchedule.OpenTime) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "scheduledTime.startTime",
			Message: fmt.Sprintf("start time %s is before opening time %s", scheduledTime.StartTime, daySchedule.OpenTime),
		})
	}

	if scheduledTime.EndTime.After(daySchedule.CloseTime) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "scheduledTime.endTime",
			Message: fmt.Sprintf("end time %s is after closing time %s", scheduledTime.EndTime, daySchedule.CloseTime),
		})
	}

	return result
}

// ValidateProductAvailability validates that all products referenced in the
// order items are available at submission time.
func ValidateProductAvailability(items []domain.OrderItem, products []domain.Product) ValidationResult {
	result := ValidationResult{}

	// Build a lookup map of product ID → Product.
	productMap := make(map[string]domain.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	for i, item := range items {
		field := fmt.Sprintf("items[%d]", i)

		product, exists := productMap[item.ProductID]
		if !exists {
			result.Errors = append(result.Errors, ValidationError{
				Field:   field + ".productId",
				Message: fmt.Sprintf("product %q not found", item.ProductID),
			})
			continue
		}

		if !product.IsAvailable {
			result.Errors = append(result.Errors, ValidationError{
				Field:   field + ".productId",
				Message: fmt.Sprintf("product %q is not available", product.Name),
			})
		}
	}

	return result
}
