# Bakery Ordering & Reservation App

A modern web application for browsing bakeries, placing delivery orders with online payment, and making reservations with on-spot payment.

## Quick Start

### Prerequisites

- Go 1.22+ ([install](https://go.dev/doc/install))
- Node.js 18+ and npm ([install](https://nodejs.org/))

### Run the backend

```bash
make run
```

The API server starts on `http://localhost:8080`.

### Run the frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend starts on `http://localhost:5173`.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend server port |
| `JWT_SECRET` | `dev-secret-do-not-use-in-production` | Secret key for signing/verifying JWT tokens |
| `FRONTEND_ORIGIN` | `http://localhost:5173` | Allowed CORS origin for the frontend |

## API Endpoints

All `/api/*` endpoints require a valid JWT token in the `Authorization: Bearer <token>` header.

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |

### Bakeries

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/bakeries` | List bakeries (paginated, 50/page) |
| GET | `/api/bakeries/:id/menu` | Get bakery menu grouped by category |

### Orders

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/orders` | Create a delivery order (rate limited) |
| GET | `/api/orders` | List user's orders (filtered, sorted, 20/page) |
| DELETE | `/api/orders/:id` | Cancel an order |

### Reservations

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/reservations` | Create a reservation (rate limited) |
| DELETE | `/api/reservations/:id` | Cancel a reservation |

### Payments

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/payments/callback` | Payment gateway webhook |

## Authentication

The app uses JWT tokens. To interact with the API during development, generate a token:

```bash
# Using the dev secret (don't use this in production)
# Include "sub" claim with your user ID
# Example with jwt-cli or any JWT tool:

# Payload: {"sub": "user-123", "exp": <future_timestamp>}
```

The frontend stores tokens in `localStorage` and includes them automatically in all API requests.

## Project Structure

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── domain/          # Domain models, interfaces, state machine
│   ├── middleware/      # JWT auth, rate limiting, input sanitization
│   ├── payment/         # Payment gateway integration
│   ├── repository/      # Data access (in-memory implementation)
│   ├── service/         # Business logic
│   └── validation/      # Input validation
├── db/migrations/       # PostgreSQL schema (goose format)
├── frontend/            # React + TypeScript + Vite
│   └── src/
│       ├── api/         # API client with JWT management
│       ├── components/  # Side panels, overlay
│       ├── pages/       # Bakery list, detail, schedule
│       └── types/       # TypeScript interfaces
├── Makefile
└── go.mod
```

## Available Commands

```bash
make build        # Build the binary to bin/bakery-app
make run          # Run the server
make test         # Run all Go tests
make test-race    # Run tests with race detection
make test-cover   # Run tests with coverage report
make fmt          # Format Go code
make vet          # Run go vet
make lint         # Run golangci-lint
make tidy         # Tidy go.mod
make check        # Run fmt + vet + test
make clean        # Remove build artifacts
```

### Frontend commands

```bash
cd frontend
npm run dev       # Start dev server (hot reload)
npm run build     # TypeScript check + production build
npm run preview   # Preview production build
npm run lint      # Run ESLint
```

## Database Migrations

The app currently uses in-memory storage. To switch to PostgreSQL:

```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
DB_DSN="postgres://localhost:5432/bakery_app?sslmode=disable" make migrate-up

# Check status
make migrate-status
```

## Key Design Decisions

- **Monetary values** are stored as `int64` cents (no floating-point precision issues)
- **State machine** enforces valid order transitions (PendingPayment → Confirmed → Preparing → Ready → Delivered)
- **Payment links** are single-use with 30-minute expiry, max 3 retry attempts
- **Reservations** always use on-spot payment (no online payment flow)
- **Rate limiting** at 10 submissions per user per minute
- **Input sanitization** strips HTML/script tags from all JSON string fields
- **Repository interfaces** allow swapping in-memory for PostgreSQL without changing business logic

## Testing

The project includes 141 Go tests:

- Unit tests for all handlers, services, and validation logic
- 14 property-based tests (using [rapid](https://pkg.go.dev/pgregory.net/rapid)) validating correctness invariants
- Integration tests covering full order and reservation flows with JWT auth

```bash
make test         # Run all tests
make test-cover   # Generate coverage report
```
