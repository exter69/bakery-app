package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// UserRole represents the access level of a user.
type UserRole int

const (
	RoleAdmin    UserRole = 0
	RoleSeller   UserRole = 1
	RoleCustomer UserRole = 2
)

// User represents an authenticated user of the system.
type User struct {
	ID               string     `json:"id"`
	Username         string     `json:"username"`
	PasswordHash     string     `json:"-"`
	Role             UserRole   `json:"role"`
	ContactEmail     string     `json:"contactEmail,omitempty"`
	Locale           string     `json:"locale,omitempty"` // "en", "fr", "nl" — defaults to "en"
	HolidayMode      bool       `json:"holidayMode"`
	HolidayFrom      *time.Time `json:"holidayFrom,omitempty"`
	HolidayTo        *time.Time `json:"holidayTo,omitempty"`
	FavoriteProducts []string   `json:"favoriteProducts"`
	StripeCustomerID string     `json:"-"` // Stripe Customer ID — internal only, never exposed via API
	CreatedAt        time.Time  `json:"createdAt"`
}

// RegistrationToken represents a token used to register as a baker.
type RegistrationToken struct {
	ID         string    `json:"id"`
	Token      string    `json:"token"`
	Email      string    `json:"email"`
	BakeryName string    `json:"bakeryName"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Used       bool      `json:"used"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Bakery represents a bakery with its schedule.
type Bakery struct {
	ID          string        `json:"id"`
	OwnerID     string        `json:"ownerId"`
	Name        string        `json:"name"`
	PhotoURL    string        `json:"photoUrl"`
	Description string        `json:"description"`
	Address     string        `json:"address"`
	Latitude      float64       `json:"latitude"`
	Longitude     float64       `json:"longitude"`
	GooglePlaceID string        `json:"googlePlaceId,omitempty"`
	MinDelivery   int64         `json:"minDeliveryAmount"` // minimum delivery amount in cents, default 1000 (€10)
	Schedule      []DaySchedule `json:"schedule"`
	CreatedAt     time.Time     `json:"createdAt"`
}

// DaySchedule represents the operating hours for a single day.
type DaySchedule struct {
	Day       DayOfWeek `json:"day"`
	OpenTime  TimeOfDay `json:"openTime"`
	CloseTime TimeOfDay `json:"closeTime"`
	IsOpen    bool      `json:"isOpen"`
}

// TimeOfDay represents a time of day as hours and minutes (HH:MM).
type TimeOfDay struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// String returns the time formatted as HH:MM.
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

// Before reports whether t is before other.
func (t TimeOfDay) Before(other TimeOfDay) bool {
	if t.Hour != other.Hour {
		return t.Hour < other.Hour
	}
	return t.Minute < other.Minute
}

// After reports whether t is after other.
func (t TimeOfDay) After(other TimeOfDay) bool {
	return other.Before(t)
}

// Equal reports whether t equals other.
func (t TimeOfDay) Equal(other TimeOfDay) bool {
	return t.Hour == other.Hour && t.Minute == other.Minute
}

// BeforeOrEqual reports whether t is before or equal to other.
func (t TimeOfDay) BeforeOrEqual(other TimeOfDay) bool {
	return t.Before(other) || t.Equal(other)
}

// AfterOrEqual reports whether t is after or equal to other.
func (t TimeOfDay) AfterOrEqual(other TimeOfDay) bool {
	return t.After(other) || t.Equal(other)
}

// Product represents a menu item in a bakery.
type Product struct {
	ID          string   `json:"id"`
	BakeryID    string   `json:"bakeryId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       int64    `json:"price"` // price in cents
	PhotoURL    string   `json:"photoUrl"`
	Category    string   `json:"category"`
	IsAvailable bool     `json:"isAvailable"`
	Allergens   []string `json:"allergens"`   // subset of ValidAllergens; empty = no allergens
	HealthScore *int     `json:"healthScore"` // nullable; 1-5
}

// MarshalJSON ensures Allergens is serialized as an empty array (never null)
// and HealthScore is serialized as a number or null.
func (p Product) MarshalJSON() ([]byte, error) {
	type Alias Product
	a := Alias(p)
	if a.Allergens == nil {
		a.Allergens = []string{}
	}
	return json.Marshal(a)
}

// Order represents a delivery order with online payment.
type Order struct {
	ID              string        `json:"id"`
	BakeryID        string        `json:"bakeryId"`
	UserID          string        `json:"userId"`
	Items           []OrderItem   `json:"items"`
	ScheduledDay    DayOfWeek     `json:"scheduledDay"`
	ScheduledTime   TimeSlot      `json:"scheduledTime"`
	Status          OrderStatus   `json:"status"`
	TotalAmount     int64         `json:"totalAmount"` // total in cents
	PaymentMethod   PaymentMethod `json:"paymentMethod"`
	SelectionMode   SelectionMode `json:"selectionMode,omitempty"`
	PaymentIntentID string        `json:"paymentIntentId,omitempty"` // Stripe PaymentIntent ID for delayed capture
	RefundStatus    string        `json:"refundStatus,omitempty"`    // "", "pending", "refunded", "partial"
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

// OrderItem represents a single line item in an order or reservation.
type OrderItem struct {
	ProductID   string `json:"productId"`
	ProductName string `json:"productName"`
	Quantity    int    `json:"quantity"`    // must be >= 1 and <= 999
	UnitPrice   int64  `json:"unitPrice"`   // price per unit in cents, must be > 0
	Subtotal    int64  `json:"subtotal"`    // quantity * unitPrice in cents
}

// TimeSlot represents a time range within a bakery's operating hours.
type TimeSlot struct {
	StartTime TimeOfDay `json:"startTime"`
	EndTime   TimeOfDay `json:"endTime"`
}

// Reservation represents a pickup reservation with on-spot payment.
type Reservation struct {
	ID            string            `json:"id"`
	BakeryID      string            `json:"bakeryId"`
	UserID        string            `json:"userId"`
	Items         []OrderItem       `json:"items"`
	ScheduledDay  DayOfWeek         `json:"scheduledDay"`
	ScheduledTime TimeSlot          `json:"scheduledTime"`
	Status        ReservationStatus `json:"status"`
	TotalAmount   int64             `json:"totalAmount"` // total in cents
	PaymentMethod PaymentMethod     `json:"paymentMethod"` // always OnSpot
	CreatedAt     time.Time         `json:"createdAt"`
}

// ProductSearchParams holds search and filter parameters.
type ProductSearchParams struct {
	Query            string   // text search on product name (case-insensitive ILIKE)
	Category         string   // filter by category (exact match)
	ExcludeAllergens []string // exclude products containing any of these allergens
	MinHealthScore   *int     // filter: health_score >= this value
	OnlyAvailable    bool     // only return available products
	PaginationParams
}

// ProductSearchResult pairs a product with its bakery info for search results.
type ProductSearchResult struct {
	Product    Product `json:"product"`
	BakeryID   string  `json:"bakeryId"`
	BakeryName string  `json:"bakeryName"`
}

// RecurringFrequency represents how often a recurring order repeats.
type RecurringFrequency string

const (
	FrequencyWeekly   RecurringFrequency = "weekly"
	FrequencyBiWeekly RecurringFrequency = "bi_weekly"
)

// SelectionMode represents how items are chosen for a recurring order.
type SelectionMode string

const (
	SelectionFixed           SelectionMode = "fixed"
	SelectionBakeryChoice    SelectionMode = "bakery_choice"
	SelectionRandomFavorites SelectionMode = "random_favorites"
)

// RecurringOrder represents a scheduled repeating order.
type RecurringOrder struct {
	ID            string             `json:"id"`
	UserID        string             `json:"userId"`
	BakeryID      string             `json:"bakeryId"`
	Items         []OrderItem        `json:"items"`
	ScheduledDay  DayOfWeek          `json:"scheduledDay"`
	ScheduledTime TimeSlot           `json:"scheduledTime"`
	Frequency     RecurringFrequency `json:"frequency"`
	SelectionMode SelectionMode      `json:"selectionMode"`
	Active        bool               `json:"active"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}
