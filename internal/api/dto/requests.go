package dto

import (
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// OrderItemRequest represents a single item in an order or reservation request.
type OrderItemRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// CreateOrderRequest is the request body for creating a delivery order.
type CreateOrderRequest struct {
	BakeryID      string             `json:"bakeryId"`
	Items         []OrderItemRequest `json:"items"`
	ScheduledDay  domain.DayOfWeek   `json:"scheduledDay"`
	ScheduledTime TimeSlotRequest    `json:"scheduledTime"`
	SelectionMode string             `json:"selectionMode,omitempty"` // "fixed" (default), "bakery_choice", or "random_favorites"
}

// TimeSlotRequest represents a time range in a request.
type TimeSlotRequest struct {
	StartTime string `json:"startTime"` // HH:MM format
	EndTime   string `json:"endTime"`   // HH:MM format
}

// CreateReservationRequest is the request body for creating a reservation.
type CreateReservationRequest struct {
	BakeryID      string             `json:"bakeryId"`
	Items         []OrderItemRequest `json:"items"`
	ScheduledDay  domain.DayOfWeek   `json:"scheduledDay"`
	ScheduledTime TimeSlotRequest    `json:"scheduledTime"`
}

// DeleteOrderRequest is the request for cancelling an order.
// The orderId comes from the URL path; userId comes from the auth context.
type DeleteOrderRequest struct {
	OrderID string `json:"orderId"`
}

// DeleteReservationRequest is the request for cancelling a reservation.
// The reservationId comes from the URL path; userId comes from the auth context.
type DeleteReservationRequest struct {
	ReservationID string `json:"reservationId"`
}

// PaymentCallbackRequest is the request body for the payment webhook callback.
type PaymentCallbackRequest struct {
	OrderID    string `json:"orderId"`
	PaymentRef string `json:"paymentRef"`
	Status     string `json:"status"` // "success" or "failed"
}

// CreateRecurringOrderRequest is the request body for creating a recurring order.
type CreateRecurringOrderRequest struct {
	BakeryID      string                     `json:"bakeryId"`
	Items         []OrderItemRequest         `json:"items"`
	ScheduledDay  domain.DayOfWeek           `json:"scheduledDay"`
	ScheduledTime TimeSlotRequest            `json:"scheduledTime"`
	Frequency     domain.RecurringFrequency  `json:"frequency"`
	SelectionMode domain.SelectionMode       `json:"selectionMode"`
}

// UpdateHolidayRequest is the request body for setting holiday mode.
type UpdateHolidayRequest struct {
	HolidayMode bool       `json:"holidayMode"`
	HolidayFrom *time.Time `json:"holidayFrom,omitempty"`
	HolidayTo   *time.Time `json:"holidayTo,omitempty"`
}

// UpdateFavoritesRequest is the request body for updating favorite products.
type UpdateFavoritesRequest struct {
	ProductIDs []string `json:"productIds"`
}

// PushSubscribeRequest is the request body for subscribing to push notifications.
type PushSubscribeRequest struct {
	Endpoint string          `json:"endpoint"`
	Keys     PushKeysRequest `json:"keys"`
}

// PushKeysRequest holds the cryptographic keys from the browser push subscription.
type PushKeysRequest struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}

// PushUnsubscribeRequest is the request body for unsubscribing from push notifications.
type PushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}
