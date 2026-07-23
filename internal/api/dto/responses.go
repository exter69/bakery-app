package dto

import "time"

// BakeryCardResponse is the response for a bakery in a list view.
type BakeryCardResponse struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	PhotoURL      string                `json:"photoUrl"`
	Latitude      float64               `json:"latitude"`
	Longitude     float64               `json:"longitude"`
	RatingAvg     *float64              `json:"ratingAvg"`
	RatingCount   int                   `json:"ratingCount"`
	TodaySchedule TodayScheduleResponse `json:"todaySchedule"`
	Distance      *float64              `json:"distance,omitempty"`
}

// TodayScheduleResponse represents the bakery's schedule for today.
type TodayScheduleResponse struct {
	OpenTime  string `json:"openTime,omitempty"`  // HH:MM format, empty if closed
	CloseTime string `json:"closeTime,omitempty"` // HH:MM format, empty if closed
	IsOpen    bool   `json:"isOpen"`
}

// OrderItemResponse represents a single item in an order or reservation response.
type OrderItemResponse struct {
	ProductID   string `json:"productId"`
	ProductName string `json:"productName"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unitPrice"`
	Subtotal    int64  `json:"subtotal"`
}

// OrderResponse is the full response for a single order.
type OrderResponse struct {
	ID            string              `json:"id"`
	BakeryID      string              `json:"bakeryId"`
	Items         []OrderItemResponse `json:"items"`
	ScheduledDay  string              `json:"scheduledDay"`
	ScheduledTime TimeSlotResponse    `json:"scheduledTime"`
	Status        string              `json:"status"`
	TotalAmount   int64               `json:"totalAmount"`
	PaymentMethod string              `json:"paymentMethod"`
	SelectionMode string              `json:"selectionMode,omitempty"`
	PaymentLink   string              `json:"paymentLink,omitempty"`
	CreatedAt     time.Time           `json:"createdAt"`
	UpdatedAt     time.Time           `json:"updatedAt"`
}

// TimeSlotResponse represents a time range in a response.
type TimeSlotResponse struct {
	StartTime string `json:"startTime"` // HH:MM format
	EndTime   string `json:"endTime"`   // HH:MM format
}

// ReservationResponse is the full response for a single reservation.
type ReservationResponse struct {
	ID            string              `json:"id"`
	BakeryID      string              `json:"bakeryId"`
	Items         []OrderItemResponse `json:"items"`
	ScheduledDay  string              `json:"scheduledDay"`
	ScheduledTime TimeSlotResponse    `json:"scheduledTime"`
	Status        string              `json:"status"`
	TotalAmount   int64               `json:"totalAmount"`
	PaymentMethod string              `json:"paymentMethod"`
	CreatedAt     time.Time           `json:"createdAt"`
}

// ListResponse is a generic paginated response wrapper.
type ListResponse[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// PaymentCallbackResponse is the response for a payment callback.
type PaymentCallbackResponse struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationErrorDetail represents a single field-level validation error.
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse is the response returned when validation fails.
type ValidationErrorResponse struct {
	Code    string                  `json:"code"`
	Message string                  `json:"message"`
	Errors  []ValidationErrorDetail `json:"errors"`
}

// RecurringOrderResponse is the full response for a single recurring order.
type RecurringOrderResponse struct {
	ID            string              `json:"id"`
	BakeryID      string              `json:"bakeryId"`
	Items         []OrderItemResponse `json:"items"`
	ScheduledDay  string              `json:"scheduledDay"`
	ScheduledTime TimeSlotResponse    `json:"scheduledTime"`
	Frequency     string              `json:"frequency"`
	SelectionMode string              `json:"selectionMode"`
	Active        bool                `json:"active"`
	CreatedAt     time.Time           `json:"createdAt"`
	UpdatedAt     time.Time           `json:"updatedAt"`
}

// UserProfileResponse is the response for the user profile endpoint.
type UserProfileResponse struct {
	ID               string     `json:"id"`
	Username         string     `json:"username"`
	Role             int        `json:"role"`
	HolidayMode      bool       `json:"holidayMode"`
	HolidayFrom      *time.Time `json:"holidayFrom,omitempty"`
	HolidayTo        *time.Time `json:"holidayTo,omitempty"`
	FavoriteProducts []string   `json:"favoriteProducts"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// FavoritesResponse is the response for the favorites endpoint.
type FavoritesResponse struct {
	ProductIDs []string `json:"productIds"`
}

// PushSubscribeResponse is returned after successfully subscribing to push notifications.
type PushSubscribeResponse struct {
	ID string `json:"id"`
}

// VAPIDKeyResponse returns the public VAPID key for push subscription.
type VAPIDKeyResponse struct {
	PublicKey string `json:"publicKey"`
}
