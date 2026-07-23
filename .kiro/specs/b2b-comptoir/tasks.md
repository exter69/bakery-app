# Implementation Plan: B2B Comptoir Portal

## Overview

This plan implements the B2B Comptoir portal end-to-end: database migration (7 tables + ALTER orders), Go domain types and validation, repository interface and Postgres implementation, B2B service with pricing/cutoff/access logic, API handler with DTOs and routes, frontend types/API client, ComptoirLayout with navigation, CommanderPage (spreadsheet grid), RecurrencesPage, LivraisonsPage, FacturesPage, B2B cart logic (localStorage), Baker Portal B2B config page, i18n translations, and route registration.

## Tasks

- [x] 1. Database migration and domain types
  - [x] 1.1 Create database migration `020_create_b2b_tables.sql`
    - Create `business_profiles` table (id, user_id, company_name, vat_siret, iban, billing_email, billing_contact_name, created_at, updated_at)
    - Create `delivery_sites` table (id, user_id, name, street_address, city, postal_code, country, delivery_instructions, created_at, updated_at)
    - Create `bakery_b2b_access` table (id, bakery_id, business_user_id, status, created_at, updated_at) with CHECK constraint on status
    - Create `b2b_config` table (id, bakery_id, cutoff_time, delivery_window_start, delivery_window_end, order_minimum, pro_discount, created_at, updated_at)
    - Create `saved_lists` and `saved_list_items` tables
    - Create `b2b_invoices` table with sequential invoice_number per bakery
    - ALTER orders: add delivery_site_id, subtotal_ht, discount_amount, tva_amount columns
    - Add all indexes and unique constraints per design
    - Write goose Down migration to reverse all changes
    - _Requirements: 1.6, 1.7, 2.7, 3.9, 7.7, 8.6, 13.6_

  - [x] 1.2 Create `internal/domain/b2b_models.go` with types and enums
    - Define `RoleBusiness UserRole = 3` constant
    - Define `B2BAccessStatus` type with constants: pending, approved, rejected, revoked
    - Define `BusinessProfile` struct with all fields per design
    - Define `DeliverySite` struct
    - Define `B2BAccess` struct
    - Define `B2BConfig` struct (with TimeOfDay for time fields)
    - Define `SavedList` and `SavedListItem` structs
    - Define `B2BInvoice` struct
    - Define `B2BOrderPricing` struct with pricing breakdown fields
    - _Requirements: 1.1, 1.5, 2.1, 7.1, 8.6, 13.6, 14.4, 15.1_

  - [x] 1.3 Create `internal/domain/b2b_services.go` with B2BService interface
    - Define `B2BService` interface with all method signatures per design
    - Define request/response types: RegisterBusinessRequest, UpdateProfileRequest, CheckoutRequest, EditOrderRequest, B2BOrderFilters
    - _Requirements: 1.1, 2.1, 3.2, 6.2, 7.1, 8.6, 9.3, 12.5, 13.3_

  - [x] 1.4 Create `internal/domain/b2b_repository.go` with B2BRepository interface
    - Define `B2BRepository` interface with methods for: business profiles, delivery sites, access whitelisting, B2B config, saved lists, invoices
    - _Requirements: 1.6, 2.7, 3.9, 7.7, 8.6, 13.6_

  - [x]* 1.5 Write property tests for B2B registration validation
    - **Property 1: B2B Registration Validation**
    - Use `rapid` to generate random registration payloads with systematically invalid fields (missing company name, invalid email, oversized VAT)
    - Assert validation rejects all invalid payloads
    - **Validates: Requirements 1.1, 1.3**

- [x] 2. Backend repository layer
  - [x] 2.1 Implement `internal/repository/b2b_repository.go`
    - Implement `CreateProfile` / `GetProfileByUserID` / `GetProfileByVAT` / `UpdateProfile`
    - Implement `CreateSite` / `GetSiteByID` / `ListSitesByUser` / `UpdateSite` / `DeleteSite` / `CountSitesByUser`
    - Implement `CreateAccess` / `GetAccessByID` / `GetAccess` / `UpdateAccessStatus` / `ListAccessByBakery` / `ListApprovedBakeryIDs`
    - Implement `GetConfig` / `SaveConfig` (upsert using ON CONFLICT)
    - Implement `CreateSavedList` / `GetSavedListByID` / `ListSavedLists` / `DeleteSavedList`
    - Implement `CreateInvoice` / `GetInvoiceByID` / `GetInvoiceByOrder` / `ListInvoicesByUser` / `NextInvoiceNumber`
    - Use parameterized queries throughout (no string concatenation)
    - Use pgx transactions where atomicity is required
    - _Requirements: 1.6, 2.7, 3.9, 7.7, 8.6, 13.6_

  - [x]* 2.2 Write property test for profile data round-trip
    - **Property 2: VAT/SIRET Immutability**
    - Use `rapid` to generate random valid profiles, create via repository, update with random payloads including vat_siret changes
    - Assert VAT/SIRET remains unchanged after any update operation
    - **Validates: Requirements 1.5**

- [x] 3. Backend service layer
  - [x] 3.1 Implement `internal/service/b2b_service.go` - Registration and Profile
    - Implement `RegisterBusiness`: validate all fields, check VAT uniqueness, create User with role 3, create BusinessProfile in a single transaction
    - Implement `GetProfile`: fetch by user ID
    - Implement `UpdateProfile`: fetch existing profile, update mutable fields (not VAT/SIRET), persist
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 3.2 Implement `internal/service/b2b_service.go` - Delivery Sites and Access
    - Implement `CreateSite` / `ListSites` / `UpdateSite` (verify ownership) / `DeleteSite` (reject if only site)
    - Implement `RequestAccess`: check no existing request, create with status "pending"
    - Implement `ApproveAccess` / `RejectAccess` / `RevokeAccess`: update status
    - Implement `ListAccessRequests` / `ListApprovedBakeries` / `HasApprovedAccess`
    - _Requirements: 2.1, 2.2, 2.5, 2.6, 3.2, 3.3, 3.4, 3.5, 3.10_

  - [x] 3.3 Implement `internal/service/b2b_service.go` - B2B Config, Checkout, and Pricing
    - Implement `GetConfig` / `SaveConfig`
    - Implement `CheckoutBakeryGroup`: verify access, verify delivery site, check minimum, check cutoff time, compute pricing, create order with payment_method="on_invoice"
    - Implement `EditOrder`: verify ownership, verify before cutoff, validate minimum, recalculate pricing, persist
    - Implement `ComputePricing`: pure computation with pro_discount and 6% TVA
    - _Requirements: 6.2, 6.6, 7.1, 7.5, 9.3, 9.5, 9.6, 14.2, 14.3, 14.4_

  - [x] 3.4 Implement `internal/service/b2b_service.go` - Saved Lists, Deliveries, and Invoices
    - Implement `CreateSavedList` / `ListSavedLists` / `DeleteSavedList` (verify ownership)
    - Implement `ListDeliveries` with filters and pagination
    - Implement `GetLastOrder` for repeat-last-order functionality
    - Implement `ListInvoices` with pagination
    - Implement `GenerateInvoice`: create invoice with sequential number per bakery on delivery
    - Implement `DownloadInvoicePDF`: generate PDF with billing details, line items, pricing breakdown
    - _Requirements: 8.6, 8.7, 12.1, 12.4, 12.5, 13.3, 13.4, 13.5_

  - [x]* 3.5 Write property tests for pricing computation
    - **Property 5: Pricing Computation Correctness**
    - Use `rapid` to generate random item lists (quantity 1-999, price 1-100000) and random discount rates (0-100)
    - Assert: discount_amount = subtotal_ht * pro_discount / 100, tva = (subtotal_ht - discount) * 6 / 100, total_ttc = subtotal_ht - discount + tva
    - **Validates: Requirements 14.2, 14.3, 14.4**

  - [x]* 3.6 Write property tests for cutoff and minimum enforcement
    - **Property 6: Cutoff Time Enforcement**
    - **Property 7: Order Minimum Validation**
    - Use `rapid` to generate checkout requests with random times before/after cutoff
    - Assert: after-cutoff requests return error, before-cutoff requests proceed
    - Generate random item totals below/at/above minimum, assert rejection only when below
    - **Validates: Requirements 6.6, 6.7, 7.3, 7.4, 7.5, 9.5**

- [x] 4. Checkpoint - Backend domain, repository, and service complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Backend API handler and DTOs
  - [x] 5.1 Create `internal/api/dto/b2b.go` with request/response DTOs
    - Define `RegisterBusinessRequest` DTO with validation tags
    - Define `UpdateProfileRequest` DTO
    - Define `DeliverySiteRequest` / `DeliverySiteResponse` DTOs
    - Define `CheckoutRequest` / `EditOrderRequest` DTOs
    - Define `SavedListRequest` / `SavedListResponse` DTOs
    - Define `B2BConfigRequest` / `B2BConfigResponse` DTOs
    - Define `B2BInvoiceResponse` / `B2BOrderPricingResponse` DTOs
    - Define `BusinessProfileResponse` DTO
    - _Requirements: 1.1, 2.1, 7.1, 8.6, 14.1_

  - [x] 5.2 Implement `internal/api/b2b_handler.go`
    - Create `B2BHandler` struct with B2BService dependency
    - Implement `RegisterRoutes`: mount all routes on chi.Router with role middleware
    - Implement registration endpoint (POST /api/comptoir/register) - no auth required
    - Implement profile endpoints (GET/PUT /api/comptoir/profile)
    - Implement delivery site CRUD endpoints
    - Implement access request endpoint and approved bakeries listing
    - Implement product catalog endpoint (with access check)
    - Implement checkout, edit order, and pricing endpoints
    - Implement saved list CRUD endpoints
    - Implement deliveries listing and last order endpoint
    - Implement invoice listing and PDF download endpoints
    - Implement baker-facing B2B access management endpoints (approve/reject/revoke)
    - Implement baker-facing B2B config endpoints (GET/PUT)
    - Return structured error responses per design error table
    - _Requirements: 1.1, 1.3, 1.4, 2.1, 2.8, 3.2, 3.7, 3.8, 4.5, 6.2, 6.6, 7.5, 8.6, 8.7, 9.5, 12.5, 13.5, 15.2_

  - [x] 5.3 Wire `B2BHandler` into `cmd/server/main.go`
    - Instantiate `B2BRepository` with DB pool
    - Instantiate `B2BService` with repository
    - Instantiate `B2BHandler` with service
    - Call `RegisterRoutes` on the chi router
    - _Requirements: 15.2, 15.4_

  - [x]* 5.4 Write property test for access control enforcement
    - **Property 3: Access Control Enforcement**
    - **Property 8: B2B Role Authorization Gate**
    - Use `rapid` to generate random user/bakery pairs with various access states
    - Assert: non-approved pairs get 403 on product/checkout endpoints
    - Assert: requests without role 3 JWT get 403 on all /comptoir/* endpoints (except register)
    - **Validates: Requirements 3.6, 15.2, 2.8, 4.5**

- [x] 6. Checkpoint - Full backend complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Frontend types and API client
  - [x] 7.1 Create `src/types/b2b.ts`
    - Define TypeScript interfaces: BusinessProfile, DeliverySite, B2BAccessStatus, B2BAccess, B2BConfig, SavedList, SavedListItem, B2BInvoice, B2BCartItem, B2BCartGroup, B2BCart, B2BOrderPricing
    - Ensure strict typing (no `any`)
    - _Requirements: 1.1, 2.1, 5.1, 7.1, 8.6, 13.1, 14.1_

  - [x] 7.2 Create `src/api/b2b-client.ts`
    - Implement all API client functions per design: registerBusiness, getProfile, updateProfile
    - Implement delivery site CRUD: listSites, createSite, updateSite, deleteSite
    - Implement access: requestAccess, listApprovedBakeries
    - Implement catalog: getProducts, getConfig
    - Implement checkout: checkout, editOrder, computePricing
    - Implement saved lists: listSavedLists, createSavedList, deleteSavedList
    - Implement deliveries: listDeliveries, getLastOrder
    - Implement invoices: listInvoices, downloadInvoicePDF
    - _Requirements: 1.1, 2.1, 3.10, 5.6, 6.2, 8.6, 12.5, 13.5_

  - [x] 7.3 Create `src/hooks/useB2BCart.ts` - B2B cart logic with localStorage
    - Implement multi-bakery cart state: addItem, removeItem, updateQuantity, clearGroup, clearAll
    - Group items by bakery with per-group subtotal calculation
    - Persist to localStorage with key per authenticated user
    - Restore cart on page load
    - Remove empty bakery groups automatically
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

  - [x]* 7.4 Write property test for cart consistency invariant
    - **Property 4: Cart Consistency Invariant**
    - Use vitest with fast-check to generate random sequences of add/remove/update operations
    - Assert: no empty groups, group subtotals equal sum of (qty * price), overall total equals sum of group subtotals
    - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [x] 8. Frontend i18n translations
  - [x] 8.1 Add B2B translations to `src/i18n/translations.ts`
    - Add keys for EN, FR, NL: Comptoir nav tabs (Commander, Recurrences, Livraisons, Factures)
    - Add keys for site switcher, profile page labels, registration form
    - Add keys for Commande Rapide: grid headers, save list action, repeat last order action
    - Add keys for pricing summary: Sous-total HT, Remise pro, TVA, Total TTC
    - Add keys for delivery list: status labels, date headers, empty state
    - Add keys for invoice list: column headers, payment status badges
    - Add keys for B2B config page labels (baker portal)
    - Add keys for error messages: below minimum, cutoff passed, access denied
    - Ensure no empty values for any locale
    - _Requirements: 10.3, 10.4, 14.1, 14.5_

- [x] 9. Comptoir layout and navigation
  - [x] 9.1 Create `src/pages/comptoir/ComptoirLayout.tsx`
    - Implement layout wrapper with business-blue theme
    - Render ComptoirNav at top
    - Render `<Outlet />` for nested route content
    - Apply typographic design (no product photos in listings)
    - _Requirements: 10.1, 10.2, 10.7_

  - [x] 9.2 Create `src/pages/comptoir/ComptoirNav.tsx`
    - Render navigation tabs: Commander, Recurrences, Livraisons, Factures
    - Highlight active tab based on current route
    - Include SiteSwitcher component (delivery site dropdown)
    - Include AccountMenu showing company name and settings link
    - _Requirements: 10.3, 10.4, 10.5_

  - [x] 9.3 Create `src/components/comptoir/SiteSwitcher.tsx`
    - Fetch delivery sites via API
    - Show dropdown when user has multiple sites
    - Show single site name (no dropdown) when only one site exists
    - Store selected site in React context for use by checkout
    - _Requirements: 2.3, 2.4, 10.5_

- [x] 10. CommanderPage - Commande Rapide interface
  - [x] 10.1 Create `src/pages/comptoir/CommanderPage.tsx`
    - Render bakery selector (list of approved bakeries)
    - Render cutoff time and delivery window for selected bakery
    - Render Commande Rapide grid for selected bakery
    - Show "Commander" button per bakery group with order minimum indicator
    - Show cart summary with pricing breakdown (Sous-total HT, Remise pro, TVA, Total TTC)
    - Disable ordering when cutoff has passed with message
    - _Requirements: 4.1, 4.2, 5.6, 6.1, 7.2, 7.3, 7.4, 14.1, 14.5_

  - [x] 10.2 Create `src/components/comptoir/CommandeRapide.tsx`
    - Render spreadsheet grid: rows = products grouped by category, columns = name, unit price, quantity input
    - Quantity input updates B2B cart on change (add item when > 0, remove when 0)
    - Display unavailable products in disabled state with input disabled
    - Show SavedListPicker dropdown for quick list loading
    - Show "Repasser la derniere commande" button that populates grid from last order
    - Show "Sauvegarder la liste" button to save current quantities as named list
    - _Requirements: 4.2, 4.4, 8.1, 8.2, 8.3, 8.4, 8.5, 8.8_

  - [x] 10.3 Create `src/components/comptoir/SavedListPicker.tsx`
    - Dropdown listing saved lists for current bakery
    - On select: populate CommandeRapide quantities from saved list items
    - Handle unavailable products in saved list (skip and show disabled)
    - _Requirements: 8.3, 8.4, 8.8_

  - [x] 10.4 Create `src/components/comptoir/B2BCartSummary.tsx`
    - Display per-bakery pricing breakdown: Sous-total HT, Remise pro (hidden when 0%), TVA 6%, Total TTC
    - Call computePricing API to get server-computed breakdown
    - Show order minimum warning when total is below minimum
    - _Requirements: 5.6, 6.7, 14.1, 14.5_

  - [x]* 10.5 Write unit tests for CommandeRapide and cart integration
    - Test grid renders products by category with quantity inputs
    - Test quantity change adds/removes items from cart
    - Test saved list loading populates quantities
    - Test repeat last order populates quantities
    - Test disabled state for unavailable products
    - **Validates: Requirements 8.1, 8.2, 8.4, 8.5, 8.8**

- [x] 11. Checkpoint - Commander page complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 12. RecurrencesPage
  - [x] 12.1 Create `src/pages/comptoir/RecurrencesPage.tsx`
    - List existing recurring order templates with: bakery name, frequency, item count, estimated total, active status
    - Provide "Nouvelle recurrence" button to create a template
    - Provide edit and deactivate actions per template
    - Create template form: select bakery, select products with quantities, select frequency (daily weekdays, weekly, custom days), select delivery site
    - _Requirements: 11.1, 11.2, 11.5, 11.6_

  - [x]* 12.2 Write unit tests for RecurrencesPage
    - Test template list renders with frequency and status
    - Test create form validates required fields
    - Test deactivate action updates template status
    - **Validates: Requirements 11.1, 11.2, 11.5**

- [x] 13. LivraisonsPage
  - [x] 13.1 Create `src/pages/comptoir/LivraisonsPage.tsx`
    - Display chronological list of B2B orders grouped by delivery date
    - Separate upcoming deliveries from past deliveries
    - Show per entry: bakery name, order date, delivery window, item summary, total TTC, status
    - Show "Editer" button on orders before cutoff
    - Implement edit mode: modify quantities, add/remove items, save changes
    - Implement pagination with 20 entries per page
    - Implement filters: date range, bakery, status
    - _Requirements: 9.1, 9.2, 9.4, 12.1, 12.2, 12.3, 12.4_

  - [x]* 13.2 Write unit tests for LivraisonsPage
    - Test delivery list renders upcoming and past sections
    - Test edit button visibility based on cutoff time
    - Test filter application shows correct subset
    - Test pagination controls
    - **Validates: Requirements 9.1, 9.4, 12.1, 12.3, 12.4**

- [x] 14. FacturesPage
  - [x] 14.1 Create `src/pages/comptoir/FacturesPage.tsx`
    - Display list of invoices grouped by month
    - Show per entry: invoice number, bakery name, period, total HT, TVA, total TTC, payment status badge
    - Provide PDF download link per invoice
    - Implement pagination
    - _Requirements: 13.1, 13.2, 13.4, 13.5_

  - [x]* 14.2 Write unit tests for FacturesPage
    - Test invoice list renders grouped by month
    - Test payment status badges display correctly
    - Test PDF download link triggers download
    - **Validates: Requirements 13.1, 13.2, 13.4**

- [x] 15. Baker Portal B2B config page
  - [x] 15.1 Create `src/pages/dashboard/DashboardB2BPage.tsx`
    - Display B2B access requests list with approve/reject/revoke buttons
    - Display B2B config form: cutoff_time, delivery_window_start, delivery_window_end, order_minimum, pro_discount
    - Config form saves via PUT /api/dashboard/b2b/config
    - Access management calls approve/reject/revoke endpoints
    - _Requirements: 3.1, 3.3, 3.4, 3.5, 3.7, 7.6_

  - [x] 15.2 Wire Baker B2B page into dashboard routing
    - Add route for B2B management page in baker dashboard
    - Add navigation link in baker dashboard sidebar
    - _Requirements: 3.1, 7.6_

  - [x]* 15.3 Write unit tests for DashboardB2BPage
    - Test access request list renders with action buttons
    - Test approve/reject/revoke button interactions
    - Test config form validation and save
    - **Validates: Requirements 3.1, 3.3, 7.6**

- [x] 16. Checkpoint - All pages complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 17. Route registration and integration
  - [x] 17.1 Register Comptoir routes in `App.tsx`
    - Add `/comptoir` route with RoleRoute guard (allowedRoles=[3])
    - Nest ComptoirLayout with child routes: index (CommanderPage), recurrences, livraisons, factures, profile
    - Ensure unauthenticated/wrong-role users redirect to login
    - _Requirements: 10.1, 10.6, 10.7, 15.2, 15.5_

  - [x] 17.2 Create `src/pages/comptoir/ComptoirProfilePage.tsx`
    - Display and edit business profile (company name, IBAN, billing email, billing contact)
    - Show VAT/SIRET as read-only
    - Manage delivery sites (CRUD) inline on profile page
    - _Requirements: 1.5, 2.1, 2.6_

  - [x] 17.3 Verify end-to-end data flow
    - Ensure product catalog endpoint returns products grouped by category
    - Ensure checkout creates order with payment_method="on_invoice" and delivery_site_id
    - Verify pricing computation matches frontend display
    - Verify role-based route guards work for role 3
    - _Requirements: 4.1, 6.2, 14.4, 15.2, 15.4_

- [x] 18. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties using `rapid` (Go) and vitest + fast-check (frontend)
- Migration number is 019, following existing sequence (018_create_surplus_bundles.sql)
- Backend uses Go with chi router and pgx; frontend uses React + TypeScript + TanStack Query
- Prices are always in cents (BIGINT); TVA rate is 6% for Belgian bakery products
- B2B cart is a client-side concern (localStorage); backend only sees per-bakery checkout requests
- The recurring orders system (Requirement 11) is defined in the UI but the auto-generation cron/worker is a follow-up if scope needs trimming
- Baker-authored content (product names, descriptions) is not auto-translated

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "1.3", "1.4"] },
    { "id": 1, "tasks": ["1.5", "2.1"] },
    { "id": 2, "tasks": ["2.2", "3.1", "3.2"] },
    { "id": 3, "tasks": ["3.3", "3.4"] },
    { "id": 4, "tasks": ["3.5", "3.6", "5.1"] },
    { "id": 5, "tasks": ["5.2", "5.3"] },
    { "id": 6, "tasks": ["5.4", "7.1", "7.2", "8.1"] },
    { "id": 7, "tasks": ["7.3", "7.4", "9.1", "9.2"] },
    { "id": 8, "tasks": ["9.3", "10.1", "10.2"] },
    { "id": 9, "tasks": ["10.3", "10.4", "10.5"] },
    { "id": 10, "tasks": ["12.1", "13.1", "14.1"] },
    { "id": 11, "tasks": ["12.2", "13.2", "14.2", "15.1"] },
    { "id": 12, "tasks": ["15.2", "15.3", "17.1", "17.2"] },
    { "id": 13, "tasks": ["17.3"] }
  ]
}
```
