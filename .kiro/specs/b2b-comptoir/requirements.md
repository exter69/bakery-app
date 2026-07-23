# Requirements Document

## Introduction

This feature introduces a dedicated B2B ordering portal ("Comptoir") for professional clients (restaurants, hotels, cafeterias) to place bulk orders with bakeries. Business users register with a new `RoleBusiness` (role 3), provide company details (VAT/SIRET, IBAN, billing contact), and gain access to a streamlined, typographic portal at `/comptoir/*`. The portal supports multi-bakery carts, per-bakery checkout with invoice-based payment, spreadsheet-style rapid ordering, recurring order templates, delivery site management, and per-bakery business rules (minimum order amounts, cutoff times, delivery windows). Bakeries whitelist individual B2B accounts before they can order.

## Glossary

- **Comptoir_Portal**: The dedicated B2B frontend application served at `/comptoir/*` with a business-blue theme, typographic design (no product photos), and navigation tabs: Commander, Recurrences, Livraisons, Factures.
- **Business_User**: A user with `RoleBusiness` (role 3) representing a professional client (restaurant, hotel, etc.) who orders in bulk from bakeries.
- **Business_Profile**: The company-level record attached to a Business_User containing company name, VAT/SIRET number, IBAN, and billing contact information. Stored in the `business_profiles` table.
- **Delivery_Site**: A physical address where a Business_User receives deliveries. A Business_User can manage multiple Delivery_Sites. Stored in the `delivery_sites` table.
- **B2B_Access**: A bakery-to-Business_User whitelisting record in the `bakery_b2b_access` table. A Business_User can only browse and order from a bakery after that bakery has approved access.
- **B2B_Cart**: A multi-bakery cart domain object grouping cart items by bakery. Each bakery group is checked out independently.
- **Commande_Rapide**: The spreadsheet-style ordering interface allowing Business_Users to enter quantities for multiple products in a grid layout, supporting saved product lists and repeat-last-order functionality.
- **Saved_List**: A named collection of products that a Business_User saves for quick reordering in the Commande_Rapide interface.
- **Cutoff_Time**: The per-bakery deadline before which a B2B order must be placed (or last edited) for a given delivery window.
- **Delivery_Window**: The per-bakery time range during which deliveries are made to Business_Users.
- **Order_Minimum**: The minimum order amount (in cents, HT) that a bakery requires for a B2B order.
- **Pro_Discount**: A per-bakery percentage discount applied to B2B orders, displayed as "Remise pro" in the order summary.
- **B2B_Order**: An order placed by a Business_User via the Comptoir_Portal, using the existing Order model with `payment_method` set to "on_invoice".
- **B2B_API**: The backend API endpoints for B2B registration, access management, cart operations, and order placement, secured with JWT authentication requiring `RoleBusiness`.
- **Baker_Portal**: The existing dashboard application (role 0 or 1) where bakeries manage B2B access whitelisting and view B2B orders.

## Requirements

### Requirement 1: B2B User Registration and Profile

**User Story:** As a professional client, I want to register a business account with my company details, so that I can access the B2B ordering portal.

#### Acceptance Criteria

1. WHEN a user registers with role "business", THE B2B_API SHALL require: company name (string, max 200 chars), VAT or SIRET number (string, max 20 chars), IBAN (string, max 34 chars), billing contact email (valid email format), and billing contact name (string, max 100 chars).
2. WHEN a business registration is submitted with valid fields, THE B2B_API SHALL create a User record with role 3 and a linked Business_Profile record in a single transaction.
3. IF a business registration is submitted with missing or invalid required fields, THEN THE B2B_API SHALL return a 400 status code with descriptive validation error messages per field.
4. IF a business registration is submitted with a VAT/SIRET number already associated with an existing Business_Profile, THEN THE B2B_API SHALL return a 409 status code indicating the company is already registered.
5. WHEN a Business_User is authenticated, THE B2B_API SHALL provide an endpoint to retrieve and update the Business_Profile (company name, IBAN, billing contact email, billing contact name). VAT/SIRET number is read-only after creation.
6. THE B2B_API SHALL store Business_Profile data in the `business_profiles` table linked to the user via `user_id` foreign key.
7. THE Database SHALL store Business_Profile data using a numbered migration following the existing goose sequence.

### Requirement 2: Delivery Site Management

**User Story:** As a Business_User, I want to manage multiple delivery sites, so that I can receive orders at different locations.

#### Acceptance Criteria

1. THE B2B_API SHALL provide CRUD endpoints for Delivery_Sites, each containing: site name (string, max 100 chars), street address (string, max 300 chars), city (string, max 100 chars), postal code (string, max 10 chars), country (string, max 2 chars ISO), and optional delivery instructions (string, max 500 chars).
2. WHEN a Business_User creates a Delivery_Site, THE B2B_API SHALL associate the site with the authenticated Business_User and return the created record with a generated UUID.
3. WHEN a Business_User has exactly one Delivery_Site, THE Comptoir_Portal SHALL select that site as the active delivery destination by default.
4. WHEN a Business_User has multiple Delivery_Sites, THE Comptoir_Portal SHALL display a site switcher in the top navigation allowing the user to select the active delivery destination.
5. THE B2B_API SHALL require at least one Delivery_Site before a Business_User can place an order.
6. IF a Business_User attempts to delete the only remaining Delivery_Site, THEN THE B2B_API SHALL return a 422 status code indicating at least one site is required.
7. THE Database SHALL store Delivery_Site data in the `delivery_sites` table using a numbered migration following the existing goose sequence.
8. THE B2B_API SHALL require JWT authentication with `RoleBusiness` for all Delivery_Site endpoints.

### Requirement 3: Bakery B2B Access Whitelisting

**User Story:** As a bakery owner, I want to approve which business clients can order from me, so that I control my B2B customer base.

#### Acceptance Criteria

1. THE Baker_Portal SHALL display a B2B access management section listing all Business_Users who have requested access, with options to approve or reject each request.
2. WHEN a Business_User requests access to a bakery, THE B2B_API SHALL create a record in the `bakery_b2b_access` table with status "pending".
3. WHEN a bakery owner approves a B2B access request, THE B2B_API SHALL update the `bakery_b2b_access` record status to "approved" and grant the Business_User permission to view products and place orders with that bakery.
4. WHEN a bakery owner rejects a B2B access request, THE B2B_API SHALL update the `bakery_b2b_access` record status to "rejected".
5. WHEN a bakery owner revokes an approved B2B access, THE B2B_API SHALL update the status to "revoked" and prevent the Business_User from placing new orders with that bakery.
6. IF a Business_User attempts to browse products or place an order with a bakery where access is not approved, THEN THE B2B_API SHALL return a 403 status code.
7. THE B2B_API SHALL require JWT authentication with seller or admin role for access approval and revocation endpoints.
8. THE B2B_API SHALL require JWT authentication with `RoleBusiness` for access request endpoints.
9. THE Database SHALL store B2B access records in the `bakery_b2b_access` table (bakery_id, business_user_id, status, created_at, updated_at) using a numbered migration following the existing goose sequence.
10. WHEN a Business_User is authenticated, THE B2B_API SHALL provide an endpoint listing all bakeries where the user has approved access.

### Requirement 4: B2B Product Catalog

**User Story:** As a Business_User, I want to browse available products from my approved bakeries, so that I can build my order.

#### Acceptance Criteria

1. WHEN a Business_User requests the product catalog for an approved bakery, THE B2B_API SHALL return all available products grouped by category, including: product name, unit price (HT), category, and availability status.
2. THE Comptoir_Portal SHALL display products in a text-based table layout (no product photos) organized by category, with columns for product name, unit price, and a quantity input field.
3. THE Comptoir_Portal SHALL only display bakeries where the Business_User has approved B2B access.
4. WHEN a product is marked as unavailable by the bakery, THE Comptoir_Portal SHALL display the product in a disabled state with the quantity input field disabled.
5. THE B2B_API SHALL require JWT authentication with `RoleBusiness` and verified bakery access for product catalog endpoints.

### Requirement 5: Multi-Bakery Cart

**User Story:** As a Business_User, I want to add products from multiple bakeries to a single cart grouped by bakery, so that I can manage all my orders in one place.

#### Acceptance Criteria

1. THE Comptoir_Portal SHALL maintain a B2B_Cart that groups items by bakery, displaying each bakery group as a distinct section with its own subtotal.
2. WHEN a Business_User adds a product to the B2B_Cart, THE Comptoir_Portal SHALL place the item under the corresponding bakery group and update the group subtotal.
3. WHEN a Business_User modifies the quantity of a cart item, THE Comptoir_Portal SHALL recalculate the bakery group subtotal and the overall cart total.
4. WHEN a Business_User removes the last item from a bakery group, THE Comptoir_Portal SHALL remove that bakery group from the B2B_Cart.
5. THE Comptoir_Portal SHALL persist the B2B_Cart state in browser local storage so that cart contents survive page reloads and browser sessions.
6. THE Comptoir_Portal SHALL display a cart summary showing per-bakery breakdown: Sous-total HT, Remise pro (percentage and amount), TVA (calculated amount), and Total TTC per bakery, plus an overall total across all bakeries.

### Requirement 6: Per-Bakery Checkout

**User Story:** As a Business_User, I want to check out each bakery group independently, so that partial success does not block my entire order.

#### Acceptance Criteria

1. THE Comptoir_Portal SHALL provide a "Commander" button per bakery group, allowing the Business_User to submit each bakery's items as an independent B2B_Order.
2. WHEN a Business_User submits a bakery group for checkout, THE B2B_API SHALL create a B2B_Order using the existing Order model with `payment_method` set to "on_invoice" and a reference to the active Delivery_Site.
3. WHEN a bakery group checkout succeeds, THE Comptoir_Portal SHALL remove that bakery group from the B2B_Cart and display a confirmation message.
4. IF a bakery group checkout fails (validation error, server error), THEN THE Comptoir_Portal SHALL retain the items in the B2B_Cart and display the error message to the Business_User.
5. WHEN multiple bakery groups exist in the B2B_Cart, THE Comptoir_Portal SHALL allow the Business_User to submit each independently without requiring all groups to be submitted simultaneously.
6. THE B2B_API SHALL validate that the order total (HT) meets or exceeds the bakery's Order_Minimum before accepting the order.
7. IF a bakery group total is below the bakery's Order_Minimum, THEN THE B2B_API SHALL return a 422 status code with the minimum amount required and THE Comptoir_Portal SHALL display the shortage amount.

### Requirement 7: Cutoff Times and Delivery Windows

**User Story:** As a Business_User, I want to see cutoff times and delivery windows per bakery, so that I know when to place my order and when to expect delivery.

#### Acceptance Criteria

1. THE B2B_API SHALL expose per-bakery B2B configuration: cutoff_time (time of day), delivery_window_start (time of day), delivery_window_end (time of day), and order_minimum (integer, cents HT).
2. WHEN a Business_User views the Commande_Rapide interface, THE Comptoir_Portal SHALL display the cutoff time and delivery window for each bakery.
3. WHILE the current time is before a bakery's Cutoff_Time, THE Comptoir_Portal SHALL allow the Business_User to create or edit orders for that bakery's next delivery window.
4. WHILE the current time is after a bakery's Cutoff_Time, THE Comptoir_Portal SHALL disable order creation and editing for that bakery and display a message indicating the cutoff has passed.
5. WHEN a Business_User attempts to submit an order after the bakery's Cutoff_Time, THE B2B_API SHALL return a 422 status code indicating the cutoff deadline has passed.
6. THE Baker_Portal SHALL provide an interface for bakery owners to configure their B2B cutoff_time, delivery_window_start, delivery_window_end, order_minimum, and pro_discount percentage.
7. THE Database SHALL store per-bakery B2B configuration in a dedicated table or column set using a numbered migration following the existing goose sequence.

### Requirement 8: Commande Rapide (Spreadsheet Ordering)

**User Story:** As a Business_User, I want a spreadsheet-style ordering interface with saved lists, so that I can place orders quickly without navigating product pages.

#### Acceptance Criteria

1. THE Comptoir_Portal SHALL display a Commande_Rapide view with a grid layout where rows represent products and columns include: product name, unit price, and a numeric quantity input field.
2. WHEN a Business_User enters a quantity greater than zero for a product, THE Comptoir_Portal SHALL add that product to the B2B_Cart with the specified quantity.
3. THE Comptoir_Portal SHALL provide a "Sauvegarder la liste" action allowing the Business_User to save the current product selection (list of product IDs with quantities) as a named Saved_List.
4. WHEN a Business_User selects a Saved_List, THE Comptoir_Portal SHALL populate the quantity fields in the Commande_Rapide grid with the saved quantities.
5. THE Comptoir_Portal SHALL provide a "Repasser la derniere commande" action that populates the Commande_Rapide grid with the items and quantities from the Business_User's most recent B2B_Order for that bakery.
6. THE B2B_API SHALL provide CRUD endpoints for Saved_Lists, storing them per Business_User with a name (string, max 100 chars) and a list of product-quantity pairs.
7. THE B2B_API SHALL require JWT authentication with `RoleBusiness` for all Saved_List endpoints.
8. WHEN a product in a Saved_List is no longer available, THE Comptoir_Portal SHALL display that row in a disabled state and exclude the product from auto-populated quantities.

### Requirement 9: Order Editing Before Cutoff

**User Story:** As a Business_User, I want to edit my submitted orders until the bakery's cutoff time, so that I can adjust quantities based on last-minute needs.

#### Acceptance Criteria

1. WHILE the current time is before the bakery's Cutoff_Time for a submitted B2B_Order, THE Comptoir_Portal SHALL display an "Editer" button on the order.
2. WHEN a Business_User clicks "Editer" on an order before cutoff, THE Comptoir_Portal SHALL allow modification of item quantities and addition or removal of items.
3. WHEN a Business_User saves edits to a B2B_Order, THE B2B_API SHALL validate the updated order (minimum amount, product availability) and persist the changes.
4. WHILE the current time is after the bakery's Cutoff_Time, THE Comptoir_Portal SHALL hide the "Editer" button and display the order as read-only.
5. IF a Business_User attempts to edit an order via the B2B_API after the Cutoff_Time, THEN THE B2B_API SHALL return a 422 status code indicating the order can no longer be modified.
6. WHEN an order is edited, THE B2B_API SHALL recalculate the order total including Pro_Discount and update the stored total_amount.

### Requirement 10: Comptoir Portal Layout and Navigation

**User Story:** As a Business_User, I want a dedicated professional portal with clear navigation, so that I can efficiently manage my orders and account.

#### Acceptance Criteria

1. THE Comptoir_Portal SHALL be served at the `/comptoir/*` route prefix, separate from the consumer-facing routes.
2. THE Comptoir_Portal SHALL apply a business-blue color theme distinct from the consumer portal, with typographic design (no product photos in listings).
3. THE Comptoir_Portal SHALL display a top navigation bar with tabs: Commander, Recurrences, Livraisons, Factures.
4. THE Comptoir_Portal SHALL display an account menu in the top navigation showing the Business_User's company name and providing access to profile settings.
5. THE Comptoir_Portal SHALL display a site switcher in the top navigation when the Business_User has multiple Delivery_Sites, showing the active site name and allowing selection of a different site.
6. THE Comptoir_Portal SHALL require JWT authentication with `RoleBusiness` (role 3); unauthenticated users or users with other roles accessing `/comptoir/*` SHALL be redirected to the login page.
7. THE Comptoir_Portal SHALL integrate with the existing React Router configuration in the application.

### Requirement 11: Recurrences (Recurring B2B Orders)

**User Story:** As a Business_User, I want to set up recurring orders, so that my regular deliveries are placed automatically without manual reordering.

#### Acceptance Criteria

1. THE Comptoir_Portal SHALL provide a "Recurrences" tab where Business_Users can view, create, edit, and deactivate recurring order templates.
2. WHEN a Business_User creates a recurring order template, THE B2B_API SHALL store: bakery_id, delivery_site_id, list of product-quantity pairs, frequency (daily weekdays, weekly, custom days), and active status.
3. WHEN a recurring order template is active, THE System SHALL automatically generate a B2B_Order from the template before the bakery's Cutoff_Time on each scheduled day.
4. WHEN a product in a recurring order template is no longer available, THE System SHALL skip that product in the generated order and notify the Business_User via the Comptoir_Portal.
5. THE Comptoir_Portal SHALL display each recurring template with its frequency, bakery name, item count, and estimated total.
6. THE B2B_API SHALL require JWT authentication with `RoleBusiness` for all recurring order template endpoints.
7. IF a generated recurring order fails validation (below minimum amount due to skipped products), THEN THE System SHALL not place the order and notify the Business_User.

### Requirement 12: Livraisons (Delivery Tracking)

**User Story:** As a Business_User, I want to view my delivery history and upcoming deliveries, so that I can track what has been ordered and received.

#### Acceptance Criteria

1. THE Comptoir_Portal SHALL provide a "Livraisons" tab displaying a chronological list of B2B_Orders grouped by delivery date.
2. THE Comptoir_Portal SHALL display each delivery entry with: bakery name, order date, delivery window, item summary (product names and quantities), total amount TTC, and order status.
3. THE Comptoir_Portal SHALL separate upcoming deliveries (orders with status before "delivered") from past deliveries (status "delivered" or "cancelled").
4. THE Comptoir_Portal SHALL support pagination for the delivery list with a default page size of 20 entries.
5. THE B2B_API SHALL provide a filtered order listing endpoint accepting date range, bakery, and status filters, returning orders for the authenticated Business_User.

### Requirement 13: Factures (Invoice Management)

**User Story:** As a Business_User, I want to view and download invoices for my orders, so that I can manage my accounting.

#### Acceptance Criteria

1. THE Comptoir_Portal SHALL provide a "Factures" tab displaying a list of invoices grouped by month.
2. THE Comptoir_Portal SHALL display each invoice entry with: invoice number, bakery name, period, total amount HT, TVA amount, total TTC, and payment status.
3. THE B2B_API SHALL generate invoice records when a B2B_Order is marked as delivered, referencing the order, bakery, and Business_Profile billing details.
4. THE B2B_API SHALL provide an endpoint to download an invoice as a PDF document containing: company details (from Business_Profile), bakery details, line items with quantities and unit prices, subtotal HT, Pro_Discount, TVA breakdown, and total TTC.
5. THE B2B_API SHALL require JWT authentication with `RoleBusiness` for invoice listing and download endpoints.
6. THE Database SHALL store invoice records with a sequential invoice number per bakery using a numbered migration following the existing goose sequence.

### Requirement 14: Order Summary and Pricing

**User Story:** As a Business_User, I want a clear breakdown of pricing including professional discounts and taxes, so that I understand the cost before confirming.

#### Acceptance Criteria

1. THE Comptoir_Portal SHALL display an order summary per bakery group showing: Sous-total HT (sum of line items before discount), Remise pro (percentage and computed discount amount), subtotal after discount, TVA (computed at applicable rate on the discounted subtotal), and Total TTC.
2. THE B2B_API SHALL compute the Pro_Discount as a percentage of the Sous-total HT, using the bakery's configured pro_discount rate.
3. THE B2B_API SHALL compute TVA at the standard rate (currently 6% for bakery products in Belgium) on the discounted subtotal.
4. THE B2B_API SHALL store the computed totals (subtotal_ht, discount_amount, tva_amount, total_ttc) on each B2B_Order record.
5. WHEN the bakery's Pro_Discount rate is zero, THE Comptoir_Portal SHALL omit the "Remise pro" line from the order summary.

### Requirement 15: B2B Role Authorization

**User Story:** As the system, I want to enforce role-based access so that only authorized Business_Users access the B2B portal and APIs.

#### Acceptance Criteria

1. THE System SHALL introduce `RoleBusiness` with integer value 3 in the existing UserRole enum.
2. THE B2B_API SHALL reject requests from users without `RoleBusiness` to any `/comptoir/*` or B2B-specific endpoint with a 403 status code.
3. THE existing consumer portal routes SHALL remain inaccessible to users with `RoleBusiness` unless explicitly permitted.
4. THE JWT token issued at login for a Business_User SHALL include role value 3, compatible with the existing JWT middleware.
5. THE RoleRoute component SHALL be extended to support role 3 for the Comptoir_Portal routes.
