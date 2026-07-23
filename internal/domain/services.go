package domain

import "context"

// PaginationParams holds pagination parameters for list queries.
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// OrderFilters holds filtering and sorting options for order queries.
type OrderFilters struct {
	Status  *OrderStatus `json:"status,omitempty"`
	SortBy  string       `json:"sortBy,omitempty"`  // "scheduledTime" or "createdAt"
	SortDir string       `json:"sortDir,omitempty"` // "asc" or "desc"
}

// ReservationFilters holds filtering and sorting options for reservation queries.
type ReservationFilters struct {
	Status  *ReservationStatus `json:"status,omitempty"`
	SortBy  string             `json:"sortBy,omitempty"`
	SortDir string             `json:"sortDir,omitempty"`
}

// ListResult is a generic paginated response wrapper.
type ListResult[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// PaymentLink holds a generated payment URL with metadata.
type PaymentLink struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expiresIn"` // seconds until expiry
}

// BakeryListParams extends pagination with optional location-based filtering.
type BakeryListParams struct {
	PaginationParams
	Lat    *float64 // user latitude (optional)
	Lng    *float64 // user longitude (optional)
	Radius float64  // max distance in km (default 100)
}

// BakeryWithDistance pairs a bakery with its computed distance from the user.
type BakeryWithDistance struct {
	Bakery   Bakery
	Distance *float64 // distance in km, nil if no location provided
}

// BakeryService handles bakery listing and menu retrieval.
type BakeryService interface {
	// ListBakeries returns a paginated list of bakeries with today's schedule.
	// If lat/lng are provided in params, results are filtered by radius and sorted by distance.
	ListBakeries(ctx context.Context, params BakeryListParams) (*ListResult[BakeryWithDistance], error)

	// GetMenu returns products grouped by category for a given bakery.
	GetMenu(ctx context.Context, bakeryID string) (map[string][]Product, error)

	// SearchProducts searches products across bakeries with filters.
	SearchProducts(ctx context.Context, params ProductSearchParams) (*ListResult[ProductSearchResult], error)
}

// OrderService handles order creation, retrieval, and deletion.
type OrderService interface {
	// CreateOrder validates and creates a new delivery order.
	CreateOrder(ctx context.Context, userID string, order Order) (*Order, *PaymentLink, error)

	// GetOrders returns a paginated list of orders for a user.
	GetOrders(ctx context.Context, userID string, filters OrderFilters, params PaginationParams) (*ListResult[Order], error)

	// DeleteOrder cancels an order after verifying ownership and state.
	DeleteOrder(ctx context.Context, orderID string, userID string) error
}

// ReservationService handles reservation creation and deletion.
type ReservationService interface {
	// CreateReservation validates and creates a new reservation with on-spot payment.
	CreateReservation(ctx context.Context, userID string, reservation Reservation) (*Reservation, error)

	// GetReservations returns a paginated list of reservations for a user.
	GetReservations(ctx context.Context, userID string, filters ReservationFilters, params PaginationParams) (*ListResult[Reservation], error)

	// DeleteReservation cancels a reservation after verifying ownership and state.
	DeleteReservation(ctx context.Context, reservationID string, userID string) error
}

// PaymentService handles payment initiation and callback processing.
type PaymentService interface {
	// InitiatePayment generates a single-use payment link for an order.
	InitiatePayment(ctx context.Context, orderID string, amount int64) (*PaymentLink, error)

	// ProcessPaymentCallback handles the payment gateway callback to confirm payment.
	ProcessPaymentCallback(ctx context.Context, orderID string, paymentRef string) error

	// InitiateRefund initiates a refund for a cancelled order that had already been paid.
	InitiateRefund(ctx context.Context, orderID string, amount int64) error
}

// BundleFilters holds filtering options for bundle list queries.
type BundleFilters struct {
	Type         *BundleType `json:"type,omitempty"`
	PickupBefore *TimeOfDay  `json:"pickupBefore,omitempty"` // filter: pickup_end_time before this
}

// BundleService handles surplus bundle lifecycle and reservations.
type BundleService interface {
	// CreateBundle validates and stores a new bundle in draft status.
	CreateBundle(ctx context.Context, sellerID string, bundle SurplusBundle) (*SurplusBundle, error)

	// PublishBundle transitions a draft bundle to published status.
	PublishBundle(ctx context.Context, sellerID string, bundleID string) (*SurplusBundle, error)

	// ListBundles returns published bundles, optionally filtered.
	ListBundles(ctx context.Context, filters BundleFilters, params PaginationParams) (*ListResult[SurplusBundle], error)

	// GetBundle returns a single bundle by ID.
	GetBundle(ctx context.Context, bundleID string) (*SurplusBundle, error)

	// ReserveBundle creates a reservation for a customer, decrementing stock.
	ReserveBundle(ctx context.Context, customerID string, bundleID string) (*BundleReservation, error)

	// CancelReservation releases the reservation and increments stock.
	CancelReservation(ctx context.Context, customerID string, reservationID string) error

	// ConfirmReservation transitions a pending reservation to confirmed.
	ConfirmReservation(ctx context.Context, customerID string, reservationID string) (*BundleReservation, error)

	// ExpireOverdueBundles finds and expires bundles past their expires_at time.
	// Called by the expiration worker goroutine.
	ExpireOverdueBundles(ctx context.Context) (int, error)

	// ReleaseOverdueReservations releases unconfirmed reservations past pickup_end_time.
	ReleaseOverdueReservations(ctx context.Context) (int, error)

	// GetImpact returns community impact metrics for the current month.
	GetImpact(ctx context.Context) (*BundleImpact, error)
}
