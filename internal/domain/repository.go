package domain

import "context"

// BakeryRepository provides data access for bakeries and their products.
type BakeryRepository interface {
	// ListBakeries returns a paginated list of all bakeries.
	ListBakeries(ctx context.Context, params PaginationParams) ([]Bakery, int, error)

	// GetBakery returns a single bakery by ID, or nil if not found.
	GetBakery(ctx context.Context, id string) (*Bakery, error)

	// GetBakeryByOwner returns the bakery owned by the given user, or nil if not found.
	GetBakeryByOwner(ctx context.Context, ownerID string) (*Bakery, error)

	// GetByStripeConnectID returns the bakery with the given Stripe Connect account ID, or nil if not found.
	GetByStripeConnectID(ctx context.Context, stripeConnectID string) (*Bakery, error)

	// UpdateBakery persists changes to a bakery.
	UpdateBakery(ctx context.Context, bakery *Bakery) error

	// GetProductsByBakery returns all available products for a given bakery.
	GetProductsByBakery(ctx context.Context, bakeryID string) ([]Product, error)

	// GetProductByID returns a single product by ID, or nil if not found.
	GetProductByID(ctx context.Context, id string) (*Product, error)

	// CreateProduct persists a new product.
	CreateProduct(ctx context.Context, product *Product) error

	// UpdateProduct persists changes to a product.
	UpdateProduct(ctx context.Context, product *Product) error

	// DeleteProduct removes a product by ID.
	DeleteProduct(ctx context.Context, id string) error

	// SearchProducts searches products by name across all bakeries with optional filters.
	SearchProducts(ctx context.Context, params ProductSearchParams) ([]ProductSearchResult, int, error)
}

// ReservationRepository provides data access for reservations.
type ReservationRepository interface {
	// Save persists a reservation (create or update).
	Save(ctx context.Context, reservation Reservation) error

	// Get returns a reservation by ID, or nil if not found.
	Get(ctx context.Context, id string) (*Reservation, error)

	// ListByUser returns reservations for a user with optional filters and pagination.
	ListByUser(ctx context.Context, userID string, filters ReservationFilters, params PaginationParams) ([]Reservation, int, error)

	// ListByBakery returns reservations for a bakery with optional filters and pagination.
	ListByBakery(ctx context.Context, bakeryID string, filters ReservationFilters, params PaginationParams) ([]Reservation, int, error)
}

// OrderRepository provides data access for orders.
type OrderRepository interface {
	// Save persists an order (create or update).
	Save(ctx context.Context, order *Order) error

	// GetByID returns an order by ID, or nil if not found.
	GetByID(ctx context.Context, id string) (*Order, error)

	// GetByPaymentIntentID returns an order by its Stripe PaymentIntent ID, or nil if not found.
	GetByPaymentIntentID(ctx context.Context, paymentIntentID string) (*Order, error)

	// ListByUser returns orders for a user with optional filters and pagination.
	ListByUser(ctx context.Context, userID string, filters OrderFilters, params PaginationParams) ([]Order, int, error)

	// ListByBakery returns orders for a bakery with optional filters and pagination.
	ListByBakery(ctx context.Context, bakeryID string, filters OrderFilters, params PaginationParams) ([]Order, int, error)
}

// UserRepository provides data access for users.
type UserRepository interface {
	// Save persists a user (create or update).
	Save(ctx context.Context, user *User) error

	// GetByID returns a user by ID, or nil if not found.
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByUsername returns a user by username, or nil if not found.
	GetByUsername(ctx context.Context, username string) (*User, error)
}

// RecurringOrderRepository provides data access for recurring orders.
type RecurringOrderRepository interface {
	// Save persists a recurring order (create or update).
	Save(ctx context.Context, order *RecurringOrder) error

	// GetByID returns a recurring order by ID, or nil if not found.
	GetByID(ctx context.Context, id string) (*RecurringOrder, error)

	// ListByUser returns recurring orders for a user with pagination.
	ListByUser(ctx context.Context, userID string, params PaginationParams) ([]RecurringOrder, int, error)

	// ListActive returns all active recurring orders.
	ListActive(ctx context.Context) ([]RecurringOrder, error)

	// Delete removes a recurring order by ID.
	Delete(ctx context.Context, id string) error
}

// RegistrationTokenRepository provides data access for registration tokens.
type RegistrationTokenRepository interface {
	// Save persists a registration token.
	Save(ctx context.Context, token *RegistrationToken) error

	// GetByToken returns a registration token by its token string, or nil if not found.
	GetByToken(ctx context.Context, tokenStr string) (*RegistrationToken, error)

	// MarkUsed marks a registration token as used.
	MarkUsed(ctx context.Context, tokenStr string) error
}

// ReviewRepository provides data access for reviews.
type ReviewRepository interface {
	// Create persists a new review and updates bakery rating aggregates.
	Create(ctx context.Context, review *Review) error

	// GetByID returns a review by ID, or nil if not found.
	GetByID(ctx context.Context, id string) (*Review, error)

	// GetByUserAndBakery returns the user's review for a bakery, or nil.
	GetByUserAndBakery(ctx context.Context, userID, bakeryID string) (*Review, error)

	// ListByBakery returns non-hidden reviews for a bakery, paginated, newest first.
	ListByBakery(ctx context.Context, bakeryID string, params PaginationParams) ([]Review, int, error)

	// ListByUser returns all reviews authored by a user (regardless of hidden status).
	ListByUser(ctx context.Context, userID string) ([]Review, error)

	// SetHidden toggles the hidden flag and recalculates bakery rating aggregates.
	SetHidden(ctx context.Context, reviewID string, hidden bool) error

	// CreateReport persists a review report.
	CreateReport(ctx context.Context, report *ReviewReport) error
}

// BundleRepository provides data access for surplus bundles and reservations.
type BundleRepository interface {
	// CreateBundle persists a new surplus bundle with its items.
	CreateBundle(ctx context.Context, bundle *SurplusBundle) error

	// UpdateBundle persists changes to a surplus bundle.
	UpdateBundle(ctx context.Context, bundle *SurplusBundle) error

	// GetByID returns a surplus bundle by ID with items, or nil if not found.
	GetByID(ctx context.Context, id string) (*SurplusBundle, error)

	// ListPublished returns published bundles with optional filters and pagination.
	ListPublished(ctx context.Context, filters BundleFilters, params PaginationParams) ([]SurplusBundle, int, error)

	// GetExpiredBundles returns published bundles whose expires_at is in the past.
	GetExpiredBundles(ctx context.Context) ([]SurplusBundle, error)

	// CreateReservation persists a new bundle reservation.
	CreateReservation(ctx context.Context, reservation *BundleReservation) error

	// GetReservation returns a bundle reservation by ID, or nil if not found.
	GetReservation(ctx context.Context, id string) (*BundleReservation, error)

	// GetActiveReservation returns the active reservation (pending/confirmed) for a user+bundle, or nil.
	GetActiveReservation(ctx context.Context, userID string, bundleID string) (*BundleReservation, error)

	// UpdateReservation persists changes to a bundle reservation.
	UpdateReservation(ctx context.Context, reservation *BundleReservation) error

	// GetOverdueReservations returns pending/confirmed reservations past their bundle's pickup_end_time.
	GetOverdueReservations(ctx context.Context) ([]BundleReservation, error)

	// CountPickedUpThisMonth returns the number of picked-up reservations in the current month.
	CountPickedUpThisMonth(ctx context.Context) (int, error)

	// DecrementStock atomically decrements quantity_remaining by 1. Returns error if already 0.
	DecrementStock(ctx context.Context, bundleID string) error

	// IncrementStock atomically increments quantity_remaining by 1.
	IncrementStock(ctx context.Context, bundleID string) error
}
