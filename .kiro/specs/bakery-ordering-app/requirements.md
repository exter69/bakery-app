# Requirements Document

## Introduction

This document specifies the functional requirements for the Bakery Ordering & Reservation App — a modern web application that enables users to browse bakeries, explore menus, place delivery orders with online payment, and make reservations with on-spot payment. The requirements are derived from the technical design and cover bakery browsing, ordering workflows, reservation workflows, schedule management, and UI behavior.

## Glossary

- **App**: The Bakery Ordering & Reservation web application
- **Bakery_List_Page**: The page displaying all bakeries as rounded cards
- **Bakery_Detail_Page**: The page showing a single bakery's menu and information
- **Order_Side_Panel**: The sliding panel used to create delivery orders
- **Reservation_Side_Panel**: The sliding panel used to create reservations
- **Schedule_Orders_Page**: The table-view page for managing orders and reservations
- **Product_Selection_Overlay**: The darkened overlay that enables product clicking during order/reservation creation
- **Order**: A delivery request with online payment containing items, schedule, and payment info
- **Reservation**: A pickup request with on-spot payment containing items and schedule
- **Time_Slot**: A time range (start and end) within a bakery's operating hours
- **Order_Status**: The lifecycle state of an order (PendingPayment, Confirmed, Preparing, Ready, Delivered, Cancelled)
- **Reservation_Status**: The lifecycle state of a reservation (Confirmed, Ready, PickedUp, Cancelled)

## Requirements

### Requirement 1: Browse Bakeries

**User Story:** As a user, I want to browse a list of bakeries displayed as rounded cards, so that I can quickly identify bakeries and their current schedules.

#### Acceptance Criteria

1. WHEN the user navigates to the Bakery_List_Page, THE App SHALL display all bakeries as rounded cards showing the bakery photo, bakery name, and today's schedule including open and close times if the bakery is open, or a closed indicator if the bakery is closed today
2. WHEN a bakery card is clicked, THE App SHALL navigate the user to the Bakery_Detail_Page for that bakery
3. THE Bakery_List_Page SHALL load bakery data within 500ms for up to 50 bakeries
4. WHEN the bakery list contains more than 50 entries, THE App SHALL paginate the results displaying 50 bakeries per page
5. IF no bakeries are available in the system, THEN THE App SHALL display an empty state message indicating that no bakeries are currently listed

### Requirement 2: Explore Bakery Menu

**User Story:** As a user, I want to view a bakery's menu organized by category with modern decoration, so that I can find products I want to order or reserve.

#### Acceptance Criteria

1. WHEN the user opens the Bakery_Detail_Page, THE App SHALL display the bakery information including name, photo, description, and address, along with menu items grouped by category
2. THE Bakery_Detail_Page SHALL provide a button to open the Order_Side_Panel for delivery orders
3. THE Bakery_Detail_Page SHALL provide a button to open the Reservation_Side_Panel for reservations
4. WHEN product images are present, THE App SHALL lazy-load them so that only images within or near the viewport are fetched
5. IF the menu fails to load, THEN THE App SHALL display an error message and provide a retry option
6. IF a bakery has no products in its menu, THEN THE App SHALL display a message indicating the menu is empty

### Requirement 3: Create Delivery Order

**User Story:** As a user, I want to place a delivery order by selecting a day, time slot, and products, so that I can receive bakery items at a scheduled time.

#### Acceptance Criteria

1. WHEN the user opens the Order_Side_Panel, THE App SHALL slide the panel in from the right side and display a day and time slot selector
2. WHEN the user selects a day and time slot, THE App SHALL validate that the selected Time_Slot falls within the bakery's operating hours for that day
3. IF the user selects a day when the bakery is closed, THEN THE App SHALL display an error message and highlight available days
4. IF the user selects a time outside operating hours, THEN THE App SHALL display an error and show the valid time range
5. WHEN the user initiates product selection, THE App SHALL activate the Product_Selection_Overlay by darkening the background and enabling product clicks
6. WHEN the user clicks a product during selection mode, THE App SHALL add one unit of that product to the order items; IF the product is already in the order, THEN THE App SHALL increment the existing item's quantity by one and update the order summary in the side panel
7. WHILE the order contains no products, THE App SHALL keep the submit button disabled and display a validation message indicating that at least one product is required
8. WHEN a valid order is submitted, THE App SHALL send the order to the backend and redirect the user to online payment
9. IF the order submission to the backend fails, THEN THE App SHALL display an error message indicating the submission failure and retain the user's selections in the Order_Side_Panel
10. WHEN the order is created on the backend, THE App SHALL set the initial Order_Status to PendingPayment

### Requirement 4: Online Payment for Delivery Orders

**User Story:** As a user, I want to pay online for my delivery order, so that the bakery can confirm and prepare my order.

#### Acceptance Criteria

1. WHEN an order is submitted successfully, THE App SHALL generate a single-use payment link that expires after 30 minutes
2. WHEN payment succeeds, THE App SHALL update the Order_Status from PendingPayment to Confirmed
3. IF payment fails, THEN THE App SHALL keep the Order_Status as PendingPayment and display an error message indicating the reason for the payment failure
4. WHEN a payment failure occurs, THE App SHALL present the user with options to retry payment (up to 3 attempts) or cancel the order
5. IF the payment link expires before the user completes payment, THEN THE App SHALL set the Order_Status to Cancelled and display a message indicating the payment window has expired

### Requirement 5: Create Reservation

**User Story:** As a user, I want to make a reservation by selecting a day, time slot, and products, so that I can pick up bakery items at a scheduled time and pay on arrival.

#### Acceptance Criteria

1. WHEN the user opens the Reservation_Side_Panel, THE App SHALL slide the panel in from the right side and display a day and time slot selector
2. WHEN the user selects a day and time slot for a reservation, THE App SHALL validate that the selected Time_Slot falls within the bakery's operating hours for that day
3. IF the user selects a closed day for reservation, THEN THE App SHALL display an error message and highlight available days
4. IF the user selects a time outside operating hours for reservation, THEN THE App SHALL display an error message and show the valid time range for the selected day
5. WHEN the user initiates product selection for reservation, THE App SHALL activate the Product_Selection_Overlay by darkening the background and enabling product clicks
6. WHEN the user clicks a product during reservation selection, THE App SHALL add that product to the reservation items with a quantity of 1, or increment the quantity by 1 if the product is already in the list, and update the summary to display the item names, quantities, and estimated total amount
7. WHEN the user submits the reservation, THE App SHALL validate that at least one product is selected before sending the request
8. IF the user attempts to submit a reservation with no products selected, THEN THE App SHALL disable the submit button and display a validation message indicating at least one product is required
9. WHEN a valid reservation is submitted, THE App SHALL create the reservation with payment method set to on-spot without initiating online payment
10. WHEN a reservation is created, THE App SHALL set the initial Reservation_Status to Confirmed

### Requirement 6: Order and Reservation Data Integrity

**User Story:** As a system operator, I want orders and reservations to maintain data integrity, so that financial calculations are always correct and scheduling constraints are enforced.

#### Acceptance Criteria

1. THE App SHALL calculate the total amount of an order or reservation as the sum of all item subtotals, rounded to 2 decimal places
2. THE App SHALL calculate each item subtotal as the product of quantity and unit price, rounded to 2 decimal places
3. THE App SHALL enforce that every order or reservation item has a quantity between 1 and 999 inclusive
4. THE App SHALL enforce that every order or reservation item has a unit price greater than 0
5. THE App SHALL enforce that all products in a submitted order or reservation are available at the time of submission
6. IF a product becomes unavailable between selection and submission, THEN THE App SHALL reject the submission and return an error identifying each unavailable product
7. IF the selected day is a day the bakery is closed, THEN THE App SHALL reject the order or reservation and return an error indicating the bakery is closed on that day
8. THE App SHALL enforce that the scheduled time slot for any order or reservation starts at or after the bakery's opening time and ends at or before the bakery's closing time on the selected day
9. IF an order or reservation item fails quantity or unit price validation, THEN THE App SHALL reject the submission and return an error identifying the invalid item and the violated constraint

### Requirement 7: Manage Orders and Reservations

**User Story:** As a user, I want to view and manage my orders and reservations in a table view, so that I can track their status and cancel them if needed.

#### Acceptance Criteria

1. WHEN the user navigates to the Schedule_Orders_Page, THE App SHALL display all user orders and reservations in a table format showing items, scheduled time, and status
2. THE Schedule_Orders_Page SHALL support filtering by Order_Status and Reservation_Status, and sorting by scheduled time and creation date in ascending or descending order
3. THE Schedule_Orders_Page SHALL paginate results at 20 items per page
4. WHEN the user requests to delete an order, THE App SHALL verify that the order belongs to the requesting user
5. IF a user attempts to delete an order that does not belong to them, THEN THE App SHALL reject the request with a forbidden error
6. WHEN the user deletes an order that is in PendingPayment or Confirmed or Preparing status, THE App SHALL set the Order_Status to Cancelled
7. IF the user attempts to delete an order that is already Delivered or Cancelled, THEN THE App SHALL reject the deletion with an error message indicating that completed or already-cancelled orders cannot be deleted
8. WHEN a confirmed or preparing order is cancelled, THE App SHALL initiate a refund
9. WHEN the user deletes a reservation that is in Confirmed or Ready status, THE App SHALL set the Reservation_Status to Cancelled
10. IF the user attempts to delete a reservation that is already PickedUp or Cancelled, THEN THE App SHALL reject the deletion with an error message indicating that completed or already-cancelled reservations cannot be deleted
11. IF the user attempts to delete an order or reservation that does not exist, THEN THE App SHALL reject the request with an error message indicating the record was not found

### Requirement 8: Order State Transitions

**User Story:** As a system operator, I want orders to follow valid state transitions, so that the order lifecycle is predictable and consistent.

#### Acceptance Criteria

1. THE App SHALL allow Order_Status transitions only as follows: PendingPayment to Confirmed, PendingPayment to Cancelled, Confirmed to Preparing, Confirmed to Cancelled, Preparing to Ready, Preparing to Cancelled, Ready to Delivered
2. IF an invalid state transition is attempted, THEN THE App SHALL reject the transition, maintain the current status, and return an error indicating the attempted transition is not allowed
3. THE App SHALL treat Delivered and Cancelled as terminal states from which no further transitions are permitted

### Requirement 9: UI Presentation and Interactions

**User Story:** As a user, I want a round and modern UI with smooth transitions, so that the app feels polished and easy to use.

#### Acceptance Criteria

1. THE App SHALL render all bakery cards with rounded corners of at least 8px border-radius and display photo, name, and schedule
2. WHEN a side panel opens, THE App SHALL animate it sliding in from the right side over a duration of 200ms to 400ms at 60 frames per second
3. WHEN a side panel closes, THE App SHALL animate it sliding out over a duration of 200ms to 400ms and reset its internal state to the initial empty state, clearing selected day, time slot, and all product items
4. WHEN the Product_Selection_Overlay is activated, THE App SHALL apply a semi-transparent dark layer over the background and display a visible hover or focus state on each product to indicate it is clickable
5. WHEN the Product_Selection_Overlay is deactivated, THE App SHALL remove the dark overlay layer and restore the page to its pre-overlay visual state
6. WHEN the user clicks the "Start Selecting Products" button in a side panel, THE App SHALL activate the Product_Selection_Overlay within 100ms
7. THE App SHALL run all side panel and overlay animations at a minimum of 60 frames per second using CSS transforms

### Requirement 10: Security and Authentication

**User Story:** As a system operator, I want all operations to be authenticated and authorized, so that user data is protected and only owners can modify their orders.

#### Acceptance Criteria

1. IF a request to any API endpoint does not include a valid, non-expired JWT token, THEN THE App SHALL reject the request with an authentication error and not process the operation
2. IF a user attempts a mutation operation on an order or reservation they do not own, THEN THE App SHALL reject the request with a forbidden error and leave the resource unchanged
3. THE App SHALL strip HTML tags and script content from all user-provided text input fields before processing or storing the data
4. IF a single authenticated user exceeds 10 order or reservation submission requests within a 1-minute window, THEN THE App SHALL reject subsequent requests with a rate-limit error until the window resets
5. THE App SHALL never store card numbers, CVVs, or bank account details, delegating all payment data handling to the payment gateway
6. WHEN the App receives a request with a JWT token that has been tampered with or has an invalid signature, THE App SHALL reject the request with an authentication error
