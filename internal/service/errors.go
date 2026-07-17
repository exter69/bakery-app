package service

import (
	"errors"
	"strings"

	"github.com/lucatorrekens/bakery-app/internal/validation"
)

// Sentinel errors for service-layer failures.
var (
	// ErrBakeryNotFound is returned when a bakery ID does not match any record.
	ErrBakeryNotFound = errors.New("bakery not found")

	// ErrOrderNotFound is returned when an order ID does not match any record.
	ErrOrderNotFound = errors.New("order not found")

	// ErrReservationNotFound is returned when a reservation ID does not match any record.
	ErrReservationNotFound = errors.New("reservation not found")

	// ErrProductNotFound is returned when a product ID does not match any record.
	ErrProductNotFound = errors.New("product not found")

	// ErrForbidden is returned when a user tries to access a resource they don't own.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidStatusTransition is returned when an order/reservation status change is not allowed.
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrOrderNotCancellable is returned when trying to cancel a delivered or already-cancelled order.
	ErrOrderNotCancellable = errors.New("order cannot be cancelled in its current state")

	// ErrReservationNotCancellable is returned when trying to cancel a picked-up or already-cancelled reservation.
	ErrReservationNotCancellable = errors.New("reservation cannot be cancelled in its current state")

	// ErrUsernameRequired is returned when username is empty.
	ErrUsernameRequired = errors.New("username is required")

	// ErrPasswordTooShort is returned when password is less than 6 characters.
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")

	// ErrUsernameExists is returned when a username is already taken.
	ErrUsernameExists = errors.New("username already exists")

	// ErrInvalidCredentials is returned when login credentials are wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrRecurringOrderNotFound is returned when a recurring order ID does not match any record.
	ErrRecurringOrderNotFound = errors.New("recurring order not found")

	// ErrRecurringOrderNoItems is returned when a recurring order has no items.
	ErrRecurringOrderNoItems = errors.New("recurring order must have at least one item")

	// ErrInvalidFrequency is returned when an invalid recurring frequency is provided.
	ErrInvalidFrequency = errors.New("frequency must be 'weekly' or 'bi_weekly'")

	// ErrInvalidSelectionMode is returned when an invalid selection mode is provided.
	ErrInvalidSelectionMode = errors.New("selection mode must be 'fixed', 'bakery_choice', or 'random_favorites'")

	// ErrUserNotFound is returned when a user ID does not match any record.
	ErrUserNotFound = errors.New("user not found")
)

// ValidationErrors wraps a list of validation errors from the validation package.
type ValidationErrors struct {
	Errors []validation.ValidationError
}

func (e *ValidationErrors) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, ve := range e.Errors {
		msgs[i] = ve.Error()
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}
