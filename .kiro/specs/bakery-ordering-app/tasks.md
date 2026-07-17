# Implementation Plan: Bakery Ordering & Reservation App

## Overview

Implement a modern bakery ordering and reservation web application using Go for the backend API and a component-based frontend. The backend follows a RESTful architecture with PostgreSQL for persistence, JWT authentication, and integration with an external payment gateway. The frontend features rounded card-based UI, side-panel workflows, and overlay-driven product selection.

## Tasks

- [x] 1. Set up project structure and core interfaces
  - [x] 1.1 Initialize Go project with module structure and dependencies
    - Create Go module with `go mod init`
    - Set up directory structure: `cmd/`, `internal/`, `pkg/`, `api/`, `db/`, `frontend/`
    - Add dependencies: HTTP router (chi or gin), PostgreSQL driver, JWT library, testing libraries
    - Create Makefile with build, test, and run targets
    - _Requirements: N/A (infrastructure)_

  - [x] 1.2 Define core domain models and types
    - Create `internal/domain/` package with Go structs for Bakery, Product, Order, Reservation, OrderItem, TimeSlot, DaySchedule
    - Define enums as typed constants: OrderStatus, ReservationStatus, PaymentMethod, DayOfWeek
    - Add JSON tags and validation tags to all structs
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [x] 1.3 Define service interfaces and API request/response types
    - Create interfaces: BakeryService, OrderService, ReservationService, PaymentService
    - Define request DTOs: CreateOrderRequest, CreateReservationRequest, DeleteOrderRequest
    - Define response DTOs: OrderResponse, ReservationResponse, BakeryCardResponse
    - _Requirements: 3.8, 5.9, 7.1_

  - [x] 1.4 Set up database schema and migrations
    - Create SQL migration files for bakeries, products, orders, reservations, order_items tables
    - Define indexes for user lookups, bakery lookups, and status filtering
    - Add foreign key constraints and check constraints for data integrity
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 2. Implement bakery browsing and menu exploration
  - [x] 2.1 Implement bakery listing endpoint
    - Create `GET /api/bakeries` handler with pagination (50 per page)
    - Implement `BakeryService.ListBakeries()` with today's schedule resolution
    - Return bakery cards with photo, name, and today's open/close times or closed indicator
    - _Requirements: 1.1, 1.3, 1.4, 1.5_

  - [x] 2.2 Implement bakery menu endpoint
    - Create `GET /api/bakeries/:id/menu` handler
    - Implement `BakeryService.GetMenu(bakeryId)` returning products grouped by category
    - Handle bakery not found and empty menu cases
    - _Requirements: 2.1, 2.5, 2.6_

  - [x] 2.3 Write unit tests for bakery listing and menu
    - Test pagination logic (exactly 50 per page, page boundaries)
    - Test today's schedule resolution for open/closed bakeries
    - Test empty bakery list returns empty state
    - Test menu grouping by category
    - _Requirements: 1.1, 1.3, 1.4, 1.5, 2.1_

- [x] 3. Implement order creation and validation
  - [x] 3.1 Implement order validation logic
    - Create `internal/validation/` package with order validation functions
    - Validate: non-empty items, quantity range (1-999), positive unit price
    - Validate schedule: bakery open on selected day, time within operating hours
    - Validate product availability at submission time
    - _Requirements: 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9_

  - [x] 3.2 Implement order total calculation
    - Create `calculateOrderTotal(items []OrderItem) decimal` function
    - Calculate each item subtotal as quantity × unitPrice rounded to 2 decimal places
    - Calculate total as sum of all subtotals rounded to 2 decimal places
    - _Requirements: 6.1, 6.2_

  - [x] 3.3 Implement order creation endpoint
    - Create `POST /api/orders` handler
    - Implement `OrderService.CreateOrder()` with full validation pipeline
    - Set initial status to PendingPayment
    - Return order with payment link on success, error details on failure
    - _Requirements: 3.2, 3.3, 3.4, 3.7, 3.8, 3.9, 3.10_

  - [x] 3.4 Write property test for order total integrity
    - **Property 1: Order total integrity**
    - Generate random item lists with varying quantities and prices
    - Verify totalAmount always equals sum of (quantity × unitPrice) for all items
    - **Validates: Requirements 6.1, 6.2**

  - [x] 3.5 Write property test for schedule validity
    - **Property 3: Schedule validity**
    - Generate random time slots and bakery schedules
    - Verify orders are only accepted when bakery is open and time is within hours
    - **Validates: Requirements 3.2, 3.3, 5.2, 6.7**

  - [x] 3.6 Write property test for item quantity and price positivity
    - **Property 4: Item quantity and price positivity**
    - Generate random items with boundary values
    - Verify quantity >= 1 and unitPrice > 0 for all accepted items
    - **Validates: Requirements 6.3, 6.4**

  - [x] 3.7 Write property test for subtotal correctness
    - **Property 5: Subtotal correctness**
    - Generate random items
    - Verify each item subtotal equals quantity × unitPrice
    - **Validates: Requirements 6.2**

  - [x] 3.8 Write property test for non-empty order invariant
    - **Property 9: Non-empty order invariant**
    - Generate random order submissions
    - Verify orders with empty item lists are always rejected
    - **Validates: Requirements 3.7, 3.8**

  - [x] 3.9 Write property test for initial order status
    - **Property 10: Initial order status**
    - Generate random valid order requests
    - Verify all newly created orders have status PendingPayment
    - **Validates: Requirements 3.10**

- [x] 4. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Implement reservation creation
  - [x] 5.1 Implement reservation validation and creation endpoint
    - Create `POST /api/reservations` handler
    - Implement `ReservationService.CreateReservation()` with validation
    - Enforce paymentMethod = OnSpot, initial status = Confirmed
    - Validate schedule and product availability same as orders
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9, 5.10_

  - [x] 5.2 Write property test for reservation payment invariant
    - **Property 2: Reservation payment invariant**
    - Generate random reservation requests
    - Verify paymentMethod is always OnSpot and no payment link is generated
    - **Validates: Requirements 5.6, 5.7**

  - [x] 5.3 Write property test for product availability at order time
    - **Property 8: Product availability at order time**
    - Generate scenarios with available/unavailable products
    - Verify submissions with unavailable products are always rejected
    - **Validates: Requirements 6.5**

- [x] 6. Implement payment flow
  - [x] 6.1 Implement payment service integration
    - Create `internal/payment/` package with payment gateway client
    - Implement `PaymentService.InitiatePayment(orderId, amount)` returning single-use link with 30-minute expiry
    - Implement `PaymentService.ProcessPaymentCallback(orderId, paymentRef)` to confirm payment
    - _Requirements: 4.1, 4.2, 4.3_

  - [x] 6.2 Implement payment status handling
    - Create `POST /api/payments/callback` webhook handler
    - Update order status from PendingPayment to Confirmed on success
    - Handle payment failure: keep PendingPayment status, track retry count (max 3)
    - Handle payment link expiry: set status to Cancelled
    - _Requirements: 4.2, 4.3, 4.4, 4.5_

  - [x] 6.3 Write unit tests for payment flow
    - Test successful payment updates order to Confirmed
    - Test failed payment keeps PendingPayment, allows retry up to 3 times
    - Test expired link sets order to Cancelled
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 7. Implement order state machine and management
  - [x] 7.1 Implement order state transition logic
    - Create `internal/domain/statemachine.go` with valid transition map
    - Valid transitions: PendingPayment→Confirmed, PendingPayment→Cancelled, Confirmed→Preparing, Confirmed→Cancelled, Preparing→Ready, Preparing→Cancelled, Ready→Delivered
    - Reject invalid transitions with descriptive error
    - _Requirements: 8.1, 8.2, 8.3_

  - [x] 7.2 Implement order and reservation management endpoints
    - Create `GET /api/orders` with filtering by status, sorting by time/creation date, pagination (20/page)
    - Create `DELETE /api/orders/:id` with ownership verification and state validation
    - Create `DELETE /api/reservations/:id` with ownership verification and state validation
    - Initiate refund on cancellation of Confirmed/Preparing orders
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8, 7.9, 7.10, 7.11_

  - [x] 7.3 Write property test for order state transitions
    - **Property 6: Order state transitions**
    - Generate random sequences of state transitions
    - Verify only valid transitions succeed and invalid ones are rejected preserving current state
    - **Validates: Requirements 8.1, 8.2**

  - [x] 7.4 Write property test for ownership enforcement
    - **Property 7: Ownership enforcement**
    - Generate random user/order combinations
    - Verify mutation operations are rejected when userId != order.userId
    - **Validates: Requirements 7.4, 7.5, 10.2**

  - [x] 7.5 Write property test for deletion restricted for terminal states
    - **Property 14: Deletion restricted for terminal states**
    - Generate orders in Delivered and Cancelled states
    - Verify all delete attempts are rejected and state remains unchanged
    - **Validates: Requirements 7.7**

  - [x] 7.6 Write property test for cancellation refund invariant
    - **Property 11: Cancellation refund invariant**
    - Generate orders transitioning from Confirmed/Preparing to Cancelled
    - Verify refund is always initiated
    - **Validates: Requirements 7.8**

- [x] 8. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. Implement authentication and security middleware
  - [x] 9.1 Implement JWT authentication middleware
    - Create `internal/middleware/auth.go` with JWT validation
    - Extract user ID from token claims
    - Reject requests with missing, expired, or tampered tokens
    - _Requirements: 10.1, 10.6_

  - [x] 9.2 Implement rate limiting middleware
    - Create `internal/middleware/ratelimit.go` with per-user rate limiting
    - Limit to 10 order/reservation submissions per user per minute
    - Return rate-limit error when threshold exceeded
    - _Requirements: 10.4_

  - [x] 9.3 Implement input sanitization
    - Create `internal/middleware/sanitize.go` to strip HTML tags and script content
    - Apply to all user-provided text fields before processing/storing
    - _Requirements: 10.3_

  - [x] 9.4 Write unit tests for security middleware
    - Test valid JWT passes through
    - Test expired/tampered JWT is rejected
    - Test rate limiting triggers after 10 requests in 1 minute
    - Test HTML/script stripping from input fields
    - _Requirements: 10.1, 10.3, 10.4, 10.6_

- [x] 10. Implement frontend - bakery browsing
  - [x] 10.1 Create bakery list page with rounded cards
    - Build responsive grid of bakery cards with rounded corners (≥ 8px border-radius)
    - Display bakery photo, name, and today's schedule (open/close times or closed indicator)
    - Implement click handler to navigate to bakery detail
    - Handle empty state when no bakeries are available
    - _Requirements: 1.1, 1.2, 1.5, 9.1_

  - [x] 10.2 Create bakery detail page with menu
    - Display bakery info (name, photo, description, address)
    - Render menu items grouped by category with modern decoration
    - Implement lazy-loading for product images
    - Add buttons for "Place Order" and "Make Reservation"
    - Handle error state with retry option and empty menu state
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [x] 11. Implement frontend - order and reservation side panels
  - [x] 11.1 Create order side panel component
    - Implement sliding panel from right side with 200-400ms animation at 60fps using CSS transforms
    - Build day selector showing available/unavailable days
    - Build time slot selector within bakery operating hours
    - Display validation errors for closed days and invalid times
    - Implement submit with disabled state when no products selected
    - Reset all state on panel close
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.7, 3.8, 3.9, 9.2, 9.3, 9.7_

  - [x] 11.2 Create reservation side panel component
    - Implement sliding panel matching order panel behavior
    - Build day and time slot selectors with validation
    - Display order summary with item names, quantities, and estimated total
    - Submit without payment redirect, confirm with "pay on arrival" message
    - Reset all state on panel close
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.6, 5.7, 5.8, 9.2, 9.3, 9.7_

  - [x] 11.3 Create product selection overlay component
    - Implement semi-transparent dark overlay on background
    - Enable product click handling to add items to active panel
    - Show hover/focus state on products to indicate clickability
    - Increment quantity for already-selected products
    - Activate within 100ms of button click
    - Deactivate and restore normal view when done
    - _Requirements: 3.5, 3.6, 5.5, 5.6, 9.4, 9.5, 9.6_

  - [x] 11.4 Write property test for product selection adds to items
    - **Property 12: Product selection adds to items**
    - Generate random product click sequences
    - Verify each click increases item list length by 1 or increments existing quantity
    - **Validates: Requirements 3.6, 5.5**

  - [x] 11.5 Write property test for panel state reset on close
    - **Property 13: Panel state reset on close**
    - Generate random panel states then close
    - Verify state always resets to initial empty state
    - **Validates: Requirements 9.3**

- [x] 12. Implement frontend - schedule and orders management page
  - [x] 12.1 Create schedule/orders page with table view
    - Build table displaying orders and reservations with items, scheduled time, and status
    - Implement filtering by status (OrderStatus and ReservationStatus)
    - Implement sorting by scheduled time and creation date (ascending/descending)
    - Add pagination at 20 items per page
    - Implement delete functionality with confirmation
    - Display appropriate error messages for rejected deletions
    - _Requirements: 7.1, 7.2, 7.3, 7.6, 7.7, 7.9, 7.10, 7.11_

- [x] 13. Wire API routes and integrate frontend with backend
  - [x] 13.1 Register all API routes with middleware chain
    - Set up HTTP router with all endpoint handlers
    - Apply auth middleware to all protected routes
    - Apply rate limiting to order/reservation submission endpoints
    - Apply input sanitization middleware globally
    - Configure CORS for frontend domain
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

  - [x] 13.2 Connect frontend to backend API
    - Implement API client module with JWT token management
    - Wire bakery list and detail pages to backend endpoints
    - Wire order and reservation submission to backend
    - Wire schedule/orders page to management endpoints
    - Handle payment redirect flow
    - _Requirements: 3.8, 4.1, 5.9, 7.1_

  - [x] 13.3 Write integration tests for full order flow
    - Test: browse bakeries → view menu → create order → payment → confirmation
    - Test: browse bakeries → view menu → create reservation → confirmed
    - Test: create order → view in schedule → cancel → refund
    - Test: unauthorized access rejected
    - _Requirements: 1.1, 2.1, 3.8, 4.2, 5.9, 7.6, 10.1_

- [x] 14. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties from the design document
- Unit tests validate specific examples and edge cases
- Go is used for the backend; frontend framework choice (React, Vue, etc.) should be determined during implementation
- The payment gateway integration uses an interface to allow swapping providers (Stripe, PayPal, etc.)
- All monetary calculations use decimal types to avoid floating-point errors

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "1.4"] },
    { "id": 2, "tasks": ["1.3"] },
    { "id": 3, "tasks": ["2.1", "2.2", "3.1", "3.2"] },
    { "id": 4, "tasks": ["2.3", "3.3", "5.1"] },
    { "id": 5, "tasks": ["3.4", "3.5", "3.6", "3.7", "3.8", "3.9", "5.2", "5.3"] },
    { "id": 6, "tasks": ["6.1", "7.1"] },
    { "id": 7, "tasks": ["6.2", "7.2"] },
    { "id": 8, "tasks": ["6.3", "7.3", "7.4", "7.5", "7.6"] },
    { "id": 9, "tasks": ["9.1", "9.2", "9.3"] },
    { "id": 10, "tasks": ["9.4", "10.1", "10.2"] },
    { "id": 11, "tasks": ["11.1", "11.2", "11.3"] },
    { "id": 12, "tasks": ["11.4", "11.5", "12.1"] },
    { "id": 13, "tasks": ["13.1", "13.2"] },
    { "id": 14, "tasks": ["13.3"] }
  ]
}
```
