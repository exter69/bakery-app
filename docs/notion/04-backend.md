# 🔧 Backend

## Directory Structure

```
internal/
├── api/                  # HTTP handlers
│   ├── dto/              # Request/response DTOs
│   ├── auth_handler.go   # Login, register, request-access
│   ├── bakery_handler.go # Public bakery endpoints
│   ├── order_handler.go  # Order CRUD
│   ├── reservation_handler.go
│   ├── seller_handler.go # Seller portal endpoints
│   ├── recurring_handler.go
│   ├── payment_handler.go
│   ├── user_handler.go   # Profile, holiday, favorites
│   └── helpers.go        # Shared response utilities
├── domain/               # Core domain
│   ├── models.go         # User, Bakery, Product, Order, Reservation, RecurringOrder
│   ├── enums.go          # DayOfWeek, OrderStatus, PaymentMethod
│   ├── repository.go     # Repository interfaces
│   ├── services.go       # Service interfaces
│   ├── statemachine.go   # Order/Reservation status transitions
│   ├── calculations.go   # Price calculations
│   ├── allergens.go      # Allergen constants & validation
│   ├── selection.go      # Selection mode logic
│   ├── geo.go            # Distance calculations
│   └── panel_state.go    # UI panel state helpers
├── service/              # Business logic
│   ├── auth_service.go
│   ├── bakery_service.go
│   ├── order_service.go
│   ├── reservation_service.go
│   ├── seller_service.go
│   ├── recurring_order_service.go
│   ├── user_service.go
│   └── errors.go
├── repository/
│   └── memory/           # In-memory implementations
├── middleware/
│   ├── auth.go           # JWT authentication middleware
│   ├── ratelimit.go      # Per-user rate limiting
│   └── sanitize.go       # HTML input stripping
├── payment/
│   ├── gateway.go        # Payment gateway interface
│   ├── service.go        # Payment service
│   └── stub_gateway.go   # Stub (always succeeds)
└── validation/
    └── order.go          # Order validation rules
```

---

## Domain Models

| Model | Key Fields |
|-------|-----------|
| `User` | ID, Username, PasswordHash, Role, HolidayMode, FavoriteProducts |
| `Bakery` | ID, OwnerID, Name, Address, Lat/Lng, GooglePlaceID, Schedule, MinDelivery |
| `Product` | ID, BakeryID, Name, Price, Category, Allergens[], HealthScore |
| `Order` | ID, BakeryID, UserID, Items, ScheduledDay/Time, Status, TotalAmount, PaymentMethod |
| `Reservation` | ID, BakeryID, UserID, Items, ScheduledDay/Time, Status, PaymentMethod (OnSpot) |
| `RecurringOrder` | ID, UserID, BakeryID, Items, Frequency, SelectionMode, Active |
| `RegistrationToken` | ID, Token, Email, BakeryName, ExpiresAt, Used |

---

## Service Layer

| Service | Responsibility |
|---------|---------------|
| `AuthService` | Login, register, token validation, JWT generation |
| `BakeryService` | List bakeries, get details, menu |
| `OrderService` | Create/list/delete orders, payment orchestration |
| `ReservationService` | Create/delete reservations |
| `SellerService` | Bakery management, product CRUD, order status updates |
| `RecurringOrderService` | CRUD recurring orders |
| `UserService` | Profile, holiday mode, favorites |

---

## Middleware Stack

```
Request → Logger → Recoverer → RequestID → CORS → InputSanitizer → [JWTAuth] → [RateLimit] → Handler
```

| Middleware | Scope | Behavior |
|-----------|-------|----------|
| Logger | Global | chi request logging |
| Recoverer | Global | Panic recovery |
| CORS | Global | Allows frontend origin |
| InputSanitizer | Global | Strips HTML from request bodies |
| JWTAuth | Protected routes | Validates Bearer token, injects user context |
| RateLimit | Order/Reservation POST | 10 req/min per user |
| AuthRateLimit | Login | 5 req/min per IP |

---

## Error Handling

- Service errors are typed (see `service/errors.go`)
- Handlers map service errors to appropriate HTTP status codes
- Consistent JSON error format: `{"error": "message"}`
- Panic recovery via chi middleware
