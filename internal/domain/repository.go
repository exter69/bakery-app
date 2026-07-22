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
