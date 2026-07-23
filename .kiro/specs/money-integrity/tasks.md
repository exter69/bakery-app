# Implementation Plan: Money Integrity at the DB Boundary

## Overview

Fix three money-integrity bugs: eliminate float64 from the persistence path by migrating money columns to BIGINT, reject orders with unknown product IDs, and make the payment capture flow atomic with retry/alert logic. All implementation is in Go with PostgreSQL migrations.

## Tasks

- [x] 1. Database migration: DECIMAL to BIGINT
  - [x] 1.1 Create migration `db/migrations/025_money_columns_to_bigint.sql`
    - Up: For each money column (`products.price`, `orders.total_amount`, `order_items.unit_price`, `order_items.subtotal`, `reservations.total_amount`), alter from DECIMAL(10,2) to BIGINT with data conversion (`SET column = column * 100` before type change or use `USING (column * 100)::BIGINT`)
    - Preserve CHECK constraints (e.g. `price > 0`, `unit_price > 0`, `subtotal > 0`, `total_amount >= 0`)
    - Down: reverse BIGINT back to DECIMAL(10,2) with `column / 100.0`
    - _Requirements: 1.1, 1.2_

- [x] 2. Remove float64 conversion helpers and update repository layer
  - [x] 2.1 Delete `centsToDecimal` and `decimalToCents` from `internal/repository/postgres/helpers.go`
    - Remove the `math` import if no longer needed
    - _Requirements: 1.3, 1.5_
  - [x] 2.2 Update `internal/repository/postgres/bakery_repo.go`
    - Replace all `centsToDecimal(product.Price)` with `product.Price` in INSERT/UPDATE queries
    - Replace all `decimalToCents(priceDecimal)` scans: change scan target from `float64` to `int64`, assign directly to `p.Price`
    - _Requirements: 1.3_
  - [x] 2.3 Update `internal/repository/postgres/order_repo.go`
    - Replace `centsToDecimal(order.TotalAmount)` with `order.TotalAmount` in Save
    - Replace `centsToDecimal(item.UnitPrice)` and `centsToDecimal(item.Subtotal)` with direct int64 values
    - Replace `decimalToCents(...)` scans with direct int64 scans for `totalDec`, `unitPriceDec`, `subtotalDec`
    - _Requirements: 1.3_
  - [x] 2.4 Update `internal/repository/postgres/reservation_repo.go`
    - Same pattern: remove centsToDecimal/decimalToCents, use direct int64 read/write
    - _Requirements: 1.3_
  - [ ]* 2.5 Write property test for money round-trip
    - **Property 1: Money round-trip identity**
    - Generate random int64 cent values (including odd values like 1999, 1, large values near max)
    - Persist via repository, read back, assert identity
    - Use `pgregory.net/rapid` with minimum 100 iterations
    - **Validates: Requirements 1.4, 1.2**

- [x] 3. Checkpoint
  - Ensure `go build ./...` and `go vet ./...` pass with no compilation errors after removing the float helpers
  - Ensure all existing tests pass (the repository tests may need scan-type adjustments)
  - Ask the user if questions arise

- [ ] 4. Reject orders with unknown product IDs
  - [x] 4.1 Update enrichment logic in `internal/service/order_service.go`
    - In `CreateOrder`, Step 4 (enrich items loop): if `productMap[order.Items[i].ProductID]` lookup returns `!ok`, return an error `fmt.Errorf("unknown product ID %q in order items", order.Items[i].ProductID)`
    - Ensure the error is returned before any call to `orderRepo.Save`
    - _Requirements: 2.1, 2.2, 2.3_
  - [x] 4.2 Update `internal/service/reservation_service.go` with the same pattern
    - The reservation enrichment has the same silent-skip issue
    - _Requirements: 2.1, 2.2_
  - [ ]* 4.3 Write property test for unknown product rejection
    - **Property 2: Unknown product ID rejection**
    - Generate random product catalogs (1-10 products) and order item lists with at least one ID not in catalog
    - Verify: error is returned, error contains the unknown product ID, repository Save not called
    - Use `pgregory.net/rapid` with minimum 100 iterations
    - **Validates: Requirements 2.1, 2.2, 2.3**

- [x] 5. Make payment capture atomic with retry
  - [x] 5.1 Add `OrderStatusCapturing` to the domain state machine
    - Add `const OrderStatusCapturing OrderStatus = "capturing"` in `internal/domain/order_status.go` (or equivalent file where order statuses are defined)
    - Update the transition map: allow `ready → capturing` and `capturing → delivered`
    - Update the orders table CHECK constraint in migration to include `'capturing'`
    - _Requirements: 3.5_
  - [x] 5.2 Refactor `UpdateOrderStatus` in `internal/service/seller_service.go`
    - When `newStatus == OrderStatusDelivered` and payment capture is needed:
      1. Transition to `capturing`, save order
      2. Call `CapturePayment`
      3. Transition to `delivered`, save with retry (3 attempts)
      4. On retry exhaustion: log `[ALERT]` with order ID + payment intent ID
    - When capture is not needed (no payment intent): proceed as before
    - _Requirements: 3.1, 3.2, 3.3_
  - [ ]* 5.3 Write unit tests for capture flow
    - Test: save-with-capturing is called before CapturePayment (mock ordering)
    - Test: on save failure after capture, retries 3 times
    - Test: on retry exhaustion, alert log contains order ID and payment intent ID
    - Test: after all retries fail, order remains in `capturing` state
    - _Requirements: 3.1, 3.2, 3.3, 3.4_
  - [ ]* 5.4 Write property test for capturing state machine
    - **Property 3: Capturing state machine validity**
    - Generate random order states, attempt transition to `capturing`, verify only `ready` succeeds
    - From `capturing`, verify only `delivered` transition succeeds
    - Use `pgregory.net/rapid` with minimum 100 iterations
    - **Validates: Requirements 3.5**

- [x] 6. Final checkpoint
  - Ensure `go build ./...` and `go vet ./...` pass
  - Ensure all tests pass (existing + new property tests)
  - Verify via grep: no remaining usages of `centsToDecimal` or `decimalToCents`
  - Ask the user if questions arise

## Notes

- Migration number is 025, following existing sequence (024_widen_user_role.sql)
- The `capturing` status must also be added to the orders CHECK constraint in the migration
- Property tests use `pgregory.net/rapid` which is already used in this project (see `internal/validation/availability_property_test.go`)
- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- The reservation_service.go has the same silent-skip enrichment bug and should be fixed alongside order_service.go

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["2.1", "2.2", "2.3", "2.4", "5.1"] },
    { "id": 2, "tasks": ["2.5", "3"] },
    { "id": 3, "tasks": ["4.1", "4.2", "5.2"] },
    { "id": 4, "tasks": ["4.3", "5.3", "5.4"] },
    { "id": 5, "tasks": ["6"] }
  ]
}
```
