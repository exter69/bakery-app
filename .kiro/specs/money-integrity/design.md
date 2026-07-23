# Design Document: Money Integrity at the DB Boundary

## Overview

This design addresses three money-integrity bugs (MA-70) in the bakery marketplace backend:

1. **Float64 round-trip** — The persistence layer uses `centsToDecimal`/`decimalToCents` to convert between the domain's int64 cents and the database's DECIMAL(10,2) columns. This introduces floating-point arithmetic into the money path, violating the domain invariant.
2. **Silent 0-price items** — Order enrichment silently leaves `UnitPrice = 0` when a product lookup misses, causing either a CHECK constraint violation in Postgres or silent undercharging in memory mode.
3. **Capture-without-save** — The seller service captures payment before persisting the order status. If the save fails, the customer is charged but the order remains in `ready` state with no compensation path.

The fix converts all money columns to BIGINT (storing raw cents), adds explicit error handling for unknown product IDs, and introduces a `capturing` transitional state with retry/alert logic.

## Architecture

The changes are confined to three layers:

```
┌─────────────────────────────────────────────┐
│  Domain Layer (unchanged)                    │
│  int64 cents: Product.Price, OrderItem.*,   │
│  Order.TotalAmount, Reservation.TotalAmount │
└──────────────────────┬──────────────────────┘
                       │ direct int64 storage
┌──────────────────────▼──────────────────────┐
│  Repository Layer (modified)                 │
│  Remove centsToDecimal / decimalToCents      │
│  Write/read int64 directly to BIGINT cols    │
└──────────────────────┬──────────────────────┘
                       │
┌──────────────────────▼──────────────────────┐
│  Database (migrated)                         │
│  DECIMAL(10,2) → BIGINT for all money cols  │
└─────────────────────────────────────────────┘
```

For the order enrichment fix, the change is in the service layer:

```
Order_Service.CreateOrder
  └─ Step 4: Enrich items
       └─ if product not in catalog → return error (currently: silent skip)
```

For the capture atomicity fix:

```
Seller_Service.UpdateOrderStatus (newStatus == delivered)
  1. Save order with status = "capturing"        ← NEW
  2. Call paymentGateway.CapturePayment
  3. Save order with status = "delivered"
     └─ On failure: retry up to 3 times
     └─ On exhaustion: log alert, leave order in "capturing"
```

## Components and Interfaces

### Migration (025_money_columns_to_bigint.sql)

A new Goose migration that:
1. Alters `products.price` from `DECIMAL(10,2)` to `BIGINT` with data conversion (`price * 100`).
2. Alters `orders.total_amount` from `DECIMAL(10,2)` to `BIGINT` with conversion.
3. Alters `order_items.unit_price` and `order_items.subtotal` from `DECIMAL(10,2)` to `BIGINT`.
4. Alters `reservations.total_amount` from `DECIMAL(10,2)` to `BIGINT`.
5. Updates CHECK constraints to operate on integer values (e.g., `price > 0` remains valid for BIGINT).
6. Down migration reverses: `BIGINT` back to `DECIMAL(10,2)` with division.

### Repository Layer Changes

**Files affected:**
- `internal/repository/postgres/helpers.go` — delete `centsToDecimal` and `decimalToCents`
- `internal/repository/postgres/bakery_repo.go` — pass `product.Price` directly (int64 to BIGINT)
- `internal/repository/postgres/order_repo.go` — pass `order.TotalAmount`, `item.UnitPrice`, `item.Subtotal` directly
- `internal/repository/postgres/reservation_repo.go` — pass `reservation.TotalAmount`, item prices directly

All scan targets change from `float64` to `int64` since the DB column is now BIGINT.

### Order Enrichment Fix (order_service.go)

In `CreateOrder`, after building the `productMap`, the enrichment loop becomes:

```go
for i := range order.Items {
    p, ok := productMap[order.Items[i].ProductID]
    if !ok {
        return nil, nil, fmt.Errorf("unknown product ID %q in order items", order.Items[i].ProductID)
    }
    order.Items[i].ProductName = p.Name
    order.Items[i].UnitPrice = p.Price
}
```

This replaces the current silent-skip pattern. The error is returned before persistence.

### Capture Flow Fix (seller_service.go)

The `UpdateOrderStatus` method changes for `newStatus == delivered`:

```go
// 1. Transition to "capturing" and persist
if err := domain.TransitionOrder(order, domain.OrderStatusCapturing); err != nil {
    return nil, fmt.Errorf("transition to capturing: %w", err)
}
order.UpdatedAt = time.Now()
if err := s.orderRepo.Save(ctx, order); err != nil {
    return nil, fmt.Errorf("saving capturing state: %w", err)
}

// 2. Capture payment
if err := s.paymentGateway.CapturePayment(ctx, order.PaymentIntentID); err != nil {
    return nil, fmt.Errorf("capturing payment: %w", err)
}

// 3. Transition to delivered with retry
if err := domain.TransitionOrder(order, domain.OrderStatusDelivered); err != nil {
    return nil, fmt.Errorf("transition to delivered: %w", err)
}
order.UpdatedAt = time.Now()

var saveErr error
for attempt := 0; attempt < 3; attempt++ {
    if saveErr = s.orderRepo.Save(ctx, order); saveErr == nil {
        break
    }
}
if saveErr != nil {
    log.Printf("[ALERT] capture succeeded but save failed: orderID=%s paymentIntentID=%s err=%v",
        order.ID, order.PaymentIntentID, saveErr)
    return nil, fmt.Errorf("saving delivered state after capture: %w", saveErr)
}
```

### Domain State Machine Update

Add `OrderStatusCapturing` to the order status enum and update the transition map:

```go
const OrderStatusCapturing OrderStatus = "capturing"

// Valid transitions include:
// ready → capturing
// capturing → delivered
```

## Data Models

No new domain structs. Changes are to database column types only:

| Table | Column | Before | After |
|-------|--------|--------|-------|
| products | price | DECIMAL(10,2) | BIGINT |
| orders | total_amount | DECIMAL(10,2) | BIGINT |
| order_items | unit_price | DECIMAL(10,2) | BIGINT |
| order_items | subtotal | DECIMAL(10,2) | BIGINT |
| reservations | total_amount | DECIMAL(10,2) | BIGINT |

The domain model already uses `int64` for all these fields — no domain changes needed.

The `OrderStatus` type gains one new value: `"capturing"`.

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Money round-trip identity

*For any* valid int64 cents value (including odd-cent values like 1999, boundary values, and large values), persisting the value to the database and reading it back SHALL produce the exact same int64 value.

**Validates: Requirements 1.4, 1.2**

### Property 2: Unknown product ID rejection

*For any* order containing at least one item whose ProductID does not exist in the bakery's product catalog, the Order_Service SHALL return an error containing that product ID and SHALL NOT persist the order.

**Validates: Requirements 2.1, 2.2, 2.3**

### Property 3: Capturing state machine validity

*For any* order in the `ready` state, transitioning to `capturing` SHALL succeed. *For any* order NOT in the `ready` state (excluding `capturing` itself), transitioning to `capturing` SHALL fail. *For any* order in the `capturing` state, transitioning to `delivered` SHALL succeed.

**Validates: Requirements 3.5**

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Unknown product ID during enrichment | Return error with product ID; do not persist order |
| Capture succeeds, save fails | Retry save 3 times; on exhaustion log ALERT with order ID + payment intent ID |
| Order in `capturing` state on system recovery | Admin/operator can manually transition to `delivered` or investigate |
| Migration fails mid-way | Goose transaction rollback restores original DECIMAL columns |

## Testing Strategy

### Unit Tests (example-based)

- Capture flow ordering: verify save-with-capturing precedes gateway call (mock-based)
- Retry behavior: mock Save to fail N times, verify retry count
- Alert logging: mock Save to always fail, verify log contains order ID + payment intent
- Observable state: after save failure, order remains in `capturing`

### Property-Based Tests (using `pgregory.net/rapid`)

- **Property 1**: Generate random int64 cents values, round-trip through repository, assert identity
- **Property 2**: Generate random product catalogs and order items with missing IDs, assert error returned and no persistence
- **Property 3**: Generate random order states, attempt `capturing` transition, assert success/failure matches expected state machine

### Integration / Smoke Tests

- Migration smoke test: verify column types in information_schema after migration
- Grep/static analysis: verify no `centsToDecimal`/`decimalToCents` usages remain

### Test Configuration

- Property tests: minimum 100 iterations per property
- Test framework: Go standard `testing` package with `pgregory.net/rapid` for PBT
- Each property test tagged with: `Feature: money-integrity, Property N: <title>`
