# Requirements Document

## Introduction

This spec addresses three money-integrity bugs in the bakery marketplace backend (MA-70). The domain uses int64 cents for all monetary values, but the persistence layer converts to/from float64 for DECIMAL columns, order enrichment silently leaves price at 0 for unknown products, and the payment capture flow lacks atomicity guarantees. These bugs can cause rounding errors, silent undercharging, and irrecoverable payment states.

## Linked Ticket

MA-70 — Money integrity at the DB boundary: float64 round-trip, silent 0-price items, capture-without-save

## Glossary

- **Money_Value**: An int64 representing an amount in the smallest currency unit (cents). All monetary arithmetic in the domain uses this representation exclusively.
- **Helpers**: The conversion functions in `internal/repository/postgres/helpers.go` (`centsToDecimal`, `decimalToCents`) that bridge domain int64 cents to the database DECIMAL(10,2) columns.
- **Order_Enrichment**: The step in `order_service.go` where order items are populated with product names and prices from the catalog before persistence.
- **Capture_Flow**: The sequence in `seller_service.go` where a payment authorization is captured (charged) and the order status is then saved to the database.
- **Migration**: A Goose SQL migration file in `db/migrations/` that alters the database schema.

---

## Requirements

### Requirement 1: Eliminate float64 from the money persistence path

**User Story:** As a platform operator, I want all monetary values stored as integer cents in the database, so that no floating-point rounding error can corrupt financial data.

#### Acceptance Criteria

1. THE Migration SHALL alter columns `products.price`, `orders.total_amount`, `order_items.unit_price`, `order_items.subtotal`, and `reservations.total_amount` from DECIMAL(10,2) to BIGINT.
2. THE Migration SHALL include a data-conversion step that transforms existing DECIMAL values to integer cents (`value * 100`) without data loss.
3. THE Helpers SHALL be removed or replaced so that no float64 arithmetic exists in any code path between domain Money_Value fields and database columns.
4. WHEN a Money_Value with an odd-cent amount (e.g. 1999 representing 19.99 EUR) is persisted and re-read, THE Repository SHALL return the exact same int64 value without any rounding.
5. THE codebase SHALL contain zero usages of `centsToDecimal` or `decimalToCents` after the fix (verifiable via grep).

---

### Requirement 2: Reject orders referencing unknown product IDs

**User Story:** As a customer, I want the system to reject my order immediately if it references a product that does not exist, so that I am never silently undercharged.

#### Acceptance Criteria

1. WHEN Order_Enrichment encounters an order item whose ProductID is not found in the bakery's product catalog, THE Order_Service SHALL return an error indicating the unknown product ID.
2. WHEN an order is rejected due to an unknown product ID, THE Order_Service SHALL not persist the order.
3. THE error response SHALL include the specific product ID that was not found, so that the caller can identify the problem.

---

### Requirement 3: Make payment capture and order-status save atomic

**User Story:** As a platform operator, I want the system to handle capture-then-save failures gracefully, so that a customer is never charged without the order reaching a terminal state.

#### Acceptance Criteria

1. WHEN the Capture_Flow begins, THE Seller_Service SHALL save the order with a transitional status marker (e.g. `capturing`) before calling the payment gateway.
2. WHEN `CapturePayment` succeeds but the subsequent status save to `delivered` fails, THE Seller_Service SHALL retry the save operation up to 3 times.
3. IF all retries fail after a successful capture, THEN THE Seller_Service SHALL log an alert-level message containing the order ID and payment intent ID for manual reconciliation.
4. WHEN a simulated Save failure occurs after a successful capture in tests, THE system SHALL leave the order in the `capturing` state, which is observable and recoverable.
5. THE Order status state machine SHALL include `capturing` as a valid transitional state between `ready` and `delivered`.
