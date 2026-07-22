# 🧪 Testing

## Overview

| Layer | Framework | Location |
|-------|-----------|----------|
| Backend unit | Go test + testify | `internal/**/*_test.go` |
| Backend property | pgregory.net/rapid | `internal/**/*_property_test.go` |
| Frontend unit | Vitest + @testing-library/react | `frontend/src/__tests__/` |
| Frontend property | fast-check | `frontend/src/__tests__/` |
| E2E | Playwright | `e2e/tests/` |

---

## Backend Unit Tests

Standard Go tests with `stretchr/testify` assertions:

```bash
# Run all backend tests
make test

# With race detection
make test-race

# With coverage report
make test-cover
```

Test files:
- `internal/api/auth_handler_test.go`
- `internal/api/bakery_handler_test.go`
- `internal/api/order_handler_test.go`
- `internal/api/reservation_handler_test.go`
- `internal/api/seller_handler_test.go`
- `internal/api/payment_handler_test.go`
- `internal/api/integration_test.go`
- `internal/middleware/auth_test.go`
- `internal/middleware/ratelimit_test.go`
- `internal/middleware/sanitize_test.go`
- `internal/domain/models_test.go`
- `internal/domain/statemachine_test.go`
- `internal/domain/allergens_test.go`
- `internal/domain/calculations_test.go`
- `internal/validation/order_test.go`
- `internal/payment/service_test.go`

---

## Backend Property Tests

Using `pgregory.net/rapid` for property-based testing:

- `internal/domain/allergens_property_test.go`
- `internal/domain/calculations_property_test.go`
- `internal/domain/panel_state_property_test.go`
- `internal/domain/selection_property_test.go`
- `internal/domain/statemachine_property_test.go`
- `internal/service/order_service_property_test.go`
- `internal/service/reservation_property_test.go`
- `internal/validation/availability_property_test.go`
- `internal/validation/order_property_test.go`

Properties verified:
- Order total equals sum of item subtotals
- State machine transitions are valid and irreversible
- Allergen validation accepts only valid allergens
- Schedule availability checks are consistent

---

## Frontend Unit Tests

```bash
cd frontend
npm test
```

Using Vitest + @testing-library/react + jsdom:

- `allergen-info.test.tsx` — Allergen component rendering
- `allergens-translations.test.ts` — All allergens translated
- `baker-portal-allergens.test.tsx` — Baker product allergen editing

---

## Frontend Property Tests

Using `fast-check`:

- Price calculation invariants
- Translation key completeness
- Allergen data consistency

---

## E2E Tests (Playwright)

```bash
cd e2e
npx playwright test
```

11 spec files covering 37+ test cases:

| Spec File | Coverage |
|-----------|----------|
| `auth.spec.ts` | Login, logout, invalid credentials |
| `guest-access.spec.ts` | Browse without account |
| `customer-order.spec.ts` | Full order placement flow |
| `customer-reservation.spec.ts` | Reservation flow |
| `customer-management.spec.ts` | Schedule page, cancel orders |
| `baker-dashboard.spec.ts` | Dashboard overview, bakery editing |
| `baker-orders.spec.ts` | Order status progression |
| `allergens.spec.ts` | Allergen display and filtering |
| `i18n.spec.ts` | Language switching |
| `responsive.spec.ts` | Mobile/tablet layouts |
| `full-journey.spec.ts` | End-to-end customer → baker flow |

---

## Running All Tests

```bash
# Backend
make test

# Frontend
cd frontend && npm test

# E2E (start both servers first)
cd e2e && npx playwright test
```
