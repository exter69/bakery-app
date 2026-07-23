# Implementation Plan: Surplus Boxes (Paniers du Soir)

## Overview

This plan implements the surplus bundle feature end-to-end: database migration, Go domain types and validation, repository layer, service layer with business logic, API handler with DTOs, expiration worker goroutine, WebSocket integration, frontend types/API/hooks, i18n translations, customer-facing components (BundlePage, BundleCard, HomeBundleCard, ReservationRail, ImpactCard, BundleMapView), and baker dashboard (BundleForm, DashboardBundlesPage).

## Tasks

- [x] 1. Database migration and domain types
  - [x] 1.1 Create database migration `018_create_surplus_bundles.sql`
    - Create `surplus_bundles` table with all columns: id, bakery_id, name, type, photo_url, description, estimated_value, original_price, discounted_price, quantity_total, quantity_remaining, pickup_start_time, pickup_end_time, published_date, expires_at, status, created_at, updated_at
    - Add CHECK constraints: type IN ('compose','surprise'), status IN ('draft','published','expired','sold_out'), original_price > 0, discounted_price > 0, discounted_price < original_price, quantity_total >= 1, quantity_remaining >= 0, quantity_remaining <= quantity_total, pickup_start_time < pickup_end_time
    - Create indexes: bakery_id, status, (status, expires_at) partial, published_date DESC partial
    - Create `surplus_bundle_items` table: id, bundle_id, product_id (nullable FK), description, quantity
    - Create `bundle_reservations` table: id, bundle_id, user_id, status, created_at, updated_at
    - Add CHECK constraint on reservation status IN ('pending','confirmed','picked_up','released','cancelled')
    - Create unique partial index enforcing max one active reservation per user per bundle
    - Write goose Down migration to drop all three tables
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

  - [x] 1.2 Create `internal/domain/bundle.go` with types, enums, and validation
    - Define `BundleType` enum constants: `BundleTypeCompose`, `BundleTypeSurprise`
    - Define `BundleStatus` enum constants: draft, published, expired, sold_out
    - Define `BundleReservationStatus` enum constants: pending, confirmed, picked_up, released, cancelled
    - Define `SurplusBundle` struct with all fields per design document
    - Define `BundleItem` struct: ID, BundleID, ProductID, Description, Quantity
    - Define `BundleReservation` struct: ID, BundleID, UserID, Status, CreatedAt, UpdatedAt
    - Define `BundleImpact` struct: TotalSaved, WeightAvoided
    - Implement `ValidateBundle(bundle SurplusBundle) error` — checks name non-empty (≤100 chars), valid type, prices > 0, discounted < original, quantity_total ≥ 1, pickup window order, items required for compose, description/estimated_value required for surprise
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

  - [x] 1.3 Add `BundleService` interface to `internal/domain/services.go`
    - Add `BundleFilters` struct: Type, PickupBefore
    - Add `BundleService` interface with methods: CreateBundle, PublishBundle, ListBundles, GetBundle, ReserveBundle, CancelReservation, ConfirmReservation, ExpireOverdueBundles, ReleaseOverdueReservations, GetImpact
    - _Requirements: 2.1, 2.2, 3.1, 4.1, 4.9, 6.4, 7.1_

  - [x] 1.4 Add `BundleRepository` interface to `internal/domain/repository.go`
    - Add `BundleRepository` interface with methods: CreateBundle, UpdateBundle, GetByID, ListPublished, GetExpiredBundles, CreateReservation, GetReservation, GetActiveReservation, UpdateReservation, GetOverdueReservations, CountPickedUpThisMonth, DecrementStock, IncrementStock
    - _Requirements: 1.7, 4.1, 4.8, 6.4_

  - [ ]* 1.5 Write property tests for bundle validation
    - **Property 2: Price ordering constraint**
    - **Property 3: Pickup window constraint**
    - Use `rapid` to generate random price pairs → assert rejection when discounted >= original or discounted <= 0
    - Use `rapid` to generate random time pairs → assert rejection when start >= end
    - **Validates: Requirements 1.4, 1.5, 10.4, 10.5**

- [x] 2. Backend repository layer
  - [x] 2.1 Implement `internal/repository/bundle_repository.go`
    - Implement `CreateBundle`: insert into surplus_bundles + bulk insert surplus_bundle_items within a transaction
    - Implement `UpdateBundle`: update surplus_bundles row
    - Implement `GetByID`: select bundle with LEFT JOIN on items, assemble struct
    - Implement `ListPublished`: select published bundles with optional type/pickup_before filters, with pagination and total count
    - Implement `GetExpiredBundles`: select published bundles where expires_at < NOW()
    - Implement `CreateReservation`: insert into bundle_reservations
    - Implement `GetReservation`: select by ID
    - Implement `GetActiveReservation`: select where user_id + bundle_id and status IN ('pending','confirmed')
    - Implement `UpdateReservation`: update status and updated_at
    - Implement `GetOverdueReservations`: select pending/confirmed reservations where bundle pickup_end_time has passed today
    - Implement `CountPickedUpThisMonth`: count reservations with status 'picked_up' in current month
    - Implement `DecrementStock`: atomic UPDATE with WHERE quantity_remaining > 0, return error if no rows affected
    - Implement `IncrementStock`: atomic UPDATE quantity_remaining + 1
    - Use parameterized queries throughout (no string concatenation)
    - _Requirements: 1.7, 4.1, 4.5, 4.6, 4.8, 6.4_

  - [ ]* 2.2 Write property test for bundle data round-trip
    - **Property 1: Bundle data round-trip**
    - Use `rapid` to generate random valid bundles (both compose and surprise types)
    - Create via repository → fetch by ID → assert all fields preserved
    - **Validates: Requirements 1.1, 1.2, 1.3**

- [x] 3. Backend service layer
  - [x] 3.1 Implement `internal/service/bundle_service.go`
    - Implement `CreateBundle`: validate bundle, set status to "draft", set quantity_remaining = quantity_total, call repo.CreateBundle
    - Implement `PublishBundle`: verify seller owns bakery, verify bundle is draft, set status to "published", compute expires_at from bakery closing time, set published_date, call repo.UpdateBundle
    - Implement `ListBundles`: delegate to repo.ListPublished with filters and pagination
    - Implement `GetBundle`: delegate to repo.GetByID
    - Implement `ReserveBundle`: check bundle is published, check quantity_remaining > 0, check no existing active reservation for user+bundle, call repo.DecrementStock, create reservation with status "pending", broadcast WebSocket stock update, check if sold_out and update status
    - Implement `CancelReservation`: verify ownership, verify cancellable status (pending/confirmed), update reservation to "cancelled", call repo.IncrementStock, if bundle was sold_out revert to "published", broadcast WebSocket stock update
    - Implement `ConfirmReservation`: verify ownership, verify pending status, update to "confirmed"
    - Implement `ExpireOverdueBundles`: get expired bundles, update status to "expired", broadcast bundle_expired events
    - Implement `ReleaseOverdueReservations`: get overdue reservations, update to "released", increment stock for each
    - Implement `GetImpact`: call repo.CountPickedUpThisMonth, compute weightAvoided = count * 0.5
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 4.1, 4.5, 4.6, 4.8, 4.9, 6.1, 6.2, 7.1, 7.2, 8.1_

  - [ ]* 3.2 Write property tests for service-level invariants
    - **Property 4: Stock invariant preserved through operations**
    - **Property 5: Draft status on creation**
    - **Property 9: Zero-stock reservation rejection**
    - **Property 10: One active reservation per customer per bundle**
    - Use `rapid` to generate random sequences of reserve/cancel operations → assert 0 <= quantity_remaining <= quantity_total
    - Generate valid creation requests → assert status is always "draft" and quantity_remaining == quantity_total
    - Generate bundles with quantity_remaining = 0 → assert reservation is rejected
    - Generate scenarios with existing active reservation → assert duplicate is rejected
    - Use mock repository for isolation
    - **Validates: Requirements 1.6, 2.1, 4.1, 4.5, 4.8, 4.9**

- [x] 4. Checkpoint - Backend domain, repo, and service complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Backend API handler and DTOs
  - [x] 5.1 Create `internal/api/dto/bundle.go` with request/response DTOs
    - Define `CreateBundleRequest` struct with validation tags
    - Define `BundleItemRequest` struct
    - Define `BundleResponse` struct including bakery name and coordinates
    - Define `BundleItemResponse` struct
    - Define `BundleReservationResponse` struct
    - Define `BundleImpactResponse` struct
    - _Requirements: 1.1, 3.8, 3.10, 10.1_

  - [x] 5.2 Implement `internal/api/bundle_handler.go`
    - Create `BundleHandler` struct with BundleService and WebSocket Hub dependencies
    - Implement `RegisterRoutes`: mount all routes on chi.Router with appropriate auth middleware
    - Implement `ListBundles` (GET /api/bundles): parse query params (page, type, pickupBefore), call service, map to DTOs
    - Implement `GetBundle` (GET /api/bundles/{id}): parse path param, call service, map to DTO
    - Implement `CreateBundle` (POST /api/bundles): parse body, validate DTO, extract seller from JWT, call service
    - Implement `PublishBundle` (POST /api/bundles/{id}/publish): extract seller from JWT, call service
    - Implement `ReserveBundle` (POST /api/bundles/{id}/reserve): extract customer from JWT, call service
    - Implement `ConfirmReservation` (POST /api/bundles/{id}/reserve/confirm): extract customer from JWT, call service
    - Implement `CancelReservation` (DELETE /api/bundle-reservations/{id}): extract customer from JWT, call service
    - Implement `GetImpact` (GET /api/bundles/impact): call service, map to DTO
    - Return appropriate HTTP status codes for errors (400, 401, 403, 404, 409, 500)
    - _Requirements: 2.6, 2.7, 4.7, 4.8, 6.4, 10.5_

  - [x] 5.3 Wire `BundleHandler` into `cmd/server/main.go`
    - Instantiate `BundleRepository` with DB pool
    - Instantiate `BundleService` with repository and WebSocket hub
    - Instantiate `BundleHandler` with service and hub
    - Call `RegisterRoutes` on the chi router
    - _Requirements: 2.6, 4.7_

- [ ] 6. Expiration worker and WebSocket integration
  - [x] 6.1 Implement `internal/service/bundle_expiration.go`
    - Implement `StartExpirationWorker(ctx, svc, hub)` goroutine with 60-second ticker
    - On each tick: call `svc.ExpireOverdueBundles()` and `svc.ReleaseOverdueReservations()`
    - Log counts when bundles expired or reservations released
    - Stop cleanly on context cancellation
    - _Requirements: 7.1, 7.2, 7.4_

  - [x] 6.2 Start expiration worker in `cmd/server/main.go`
    - Launch goroutine with server's context
    - Ensure graceful shutdown stops the worker
    - _Requirements: 7.4_

  - [x] 6.3 Add WebSocket event broadcasting in BundleService
    - Broadcast `bundle_stock_update` event (bundleId, quantityRemaining, status) on reserve/cancel
    - Broadcast `bundle_expired` event (bundleId) when expiration worker expires a bundle
    - _Requirements: 8.1, 4.10_

  - [ ]* 6.4 Write property test for expiration and filter logic
    - **Property 7: Expiration transitions and reservation release**
    - **Property 11: Filter correctness**
    - Generate bundles with various expires_at timestamps → run expiration → verify only past-due bundles transition
    - Generate random bundle sets with varying types/pickup times → apply filters → verify result contains exactly matching bundles
    - **Validates: Requirements 2.4, 3.2, 3.3, 3.4, 3.5, 7.1, 7.2**

- [x] 7. Checkpoint - Full backend complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8. Frontend types, API client, and hooks
  - [x] 8.1 Create `src/types/bundle.ts`
    - Define TypeScript types: BundleType, BundleStatus, BundleReservationStatus, BundleItem, Bundle, BundleReservation, BundleImpact, BundleFilters
    - Ensure strict typing (no `any`)
    - _Requirements: 1.1, 3.8_

  - [x] 8.2 Create `src/api/bundles.ts`
    - Implement `listBundles(params)` → GET /api/bundles
    - Implement `getBundle(id)` → GET /api/bundles/{id}
    - Implement `reserveBundle(id)` → POST /api/bundles/{id}/reserve
    - Implement `confirmReservation(bundleId)` → POST /api/bundles/{bundleId}/reserve/confirm
    - Implement `cancelBundleReservation(id)` → DELETE /api/bundle-reservations/{id}
    - Implement `getBundleImpact()` → GET /api/bundles/impact
    - Implement `createBundle(data)` → POST /api/bundles (seller)
    - Implement `publishBundle(id)` → POST /api/bundles/{id}/publish (seller)
    - _Requirements: 3.1, 4.1, 4.9, 6.4_

  - [x] 8.3 Create `src/hooks/useBundles.ts`
    - Implement `useBundles(filters)` — TanStack Query hook wrapping listBundles
    - Implement `useBundle(id)` — single bundle query
    - Implement `useBundleImpact()` — impact metrics query
    - Implement `useReserveBundle()` — mutation with optimistic stock decrement
    - Implement `useConfirmReservation()` — mutation
    - Implement `useCancelBundleReservation()` — mutation with optimistic stock increment
    - Handle WebSocket `bundle_stock_update` events to invalidate/update bundle queries
    - Handle WebSocket `bundle_expired` event to remove expired bundles from cache
    - _Requirements: 4.1, 4.9, 4.10, 8.1, 8.2, 8.3_

- [x] 9. Frontend i18n translations
  - [x] 9.1 Add bundle translations to `src/i18n/translations.ts`
    - Add keys for EN, FR, NL: page title, subtitle, anti-gaspi badge, filter labels ("Retrait avant 19h", "– de 500 m", "Surprise", "Composé"), type badges, status badges ("épuisé"), action buttons ("Réserver", "Confirmer"), reservation rail text, warning messages, impact card messages, footer note, payment method label, error messages
    - Ensure no empty values for any locale
    - _Requirements: 9.1, 9.2, 9.3, 9.4_

  - [ ]* 9.2 Write property test for translation completeness
    - **Property 13: Translation completeness for bundle UI**
    - Verify every bundle i18n key has non-empty translation for all 3 locales (EN, FR, NL)
    - Use vitest with exhaustive checks over all key × locale combinations
    - **Validates: Requirements 9.1, 9.2**

- [ ] 10. Customer portal - BundleCard and BundlePage
  - [x] 10.1 Create `BundleCard` component
    - Display: photo, name, type badge ("composé" or "surprise ★"), bakery name, distance, pickup window, contents list (composé) or estimated value (surprise), original price strikethrough, discounted price, stock badge ("reste N"), "Réserver" button
    - Sold-out state: dim card, show "épuisé" badge, disable button, show "revenez demain vers 17h"
    - Handle WebSocket stock updates reactively
    - Accessible: aria-labels on badges, button states
    - _Requirements: 3.7, 3.8, 8.2, 8.3, 10.1_

  - [x] 10.2 Create `BundlePage` at route `/paniers-du-soir`
    - Page header: "Paniers du soir" title, "anti-gaspi" badge, subtitle
    - Filter bar: "Retrait avant 19h", "– de 500 m", "Surprise", "Composé" toggle buttons
    - List/Map toggle
    - Bundle grid rendering BundleCards
    - Sort by proximity (when geolocation available) or published_date (fallback)
    - Footer note about expiration and daily publishing
    - Integrate ImpactCard
    - Handle empty state
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.9, 3.11, 7.3_

  - [x] 10.3 Create `BundleMapView` component
    - Leaflet/Mapbox map with markers for bundle bakery locations
    - Clicking a marker scrolls to the corresponding BundleCard in the list
    - Respect current filters
    - _Requirements: 3.6_

  - [ ]* 10.4 Write unit tests for BundleCard and BundlePage
    - Test BundleCard renders all elements for composé and surprise types
    - Test BundleCard renders sold-out state correctly
    - Test filter logic shows/hides appropriate bundles
    - Test geolocation fallback behavior
    - **Validates: Requirements 3.1, 3.7, 3.8, 3.11**

- [x] 11. Customer portal - ReservationRail and reservation flow
  - [x] 11.1 Create `ReservationRail` component
    - Display: bundle name, quantity (1), total price (discounted_price), pickup window, payment method "paiement au comptoir"
    - Show warning: "à récupérer ce soir — sinon le panier est libéré à [pickup_end_time]"
    - "Confirmer" button calling useConfirmReservation mutation
    - "Annuler" option calling useCancelBundleReservation mutation
    - Show loading/error states
    - _Requirements: 4.2, 4.3, 4.4, 4.9, 10.2, 10.3_

  - [x] 11.2 Wire reservation flow into BundleCard
    - "Réserver" button triggers useReserveBundle mutation
    - On success: show ReservationRail with reservation details
    - On 409 (sold out race condition): show error toast, refresh bundle stock
    - Disable button while mutation is pending
    - _Requirements: 4.1, 4.5, 8.4_

  - [ ]* 11.3 Write unit tests for ReservationRail
    - Test displays reservation details, warning, and payment method
    - Test confirm and cancel button interactions
    - Test error handling on race condition (409 response)
    - **Validates: Requirements 4.2, 4.3, 4.4, 8.4**

- [x] 12. Customer portal - HomeBundleCard and ImpactCard
  - [x] 12.1 Create `HomeBundleCard` component
    - Title "Paniers du soir" with "anti-gaspi" badge and subtitle
    - Expanded nearest bundle: full contents, pricing with strikethrough, stock badge
    - Up to 3 compact bundles: photo, bakery name, distance, type badge, discounted price
    - "Voir tous les paniers →" link to `/paniers-du-soir`
    - Do not render when no published bundles available
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

  - [x] 12.2 Create `ImpactCard` component
    - Display: "Déjà N paniers sauvés 🌱" with total count
    - Display: "soit ~X kg de pain et viennoiseries évités à la poubelle ce mois-ci"
    - Use useBundleImpact hook
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 12.3 Integrate `HomeBundleCard` into the home page
    - Conditionally render when published bundles exist
    - Position in appropriate section of home page layout
    - _Requirements: 5.1, 5.6_

  - [ ]* 12.4 Write unit tests for HomeBundleCard and ImpactCard
    - Test HomeBundleCard renders expanded + compact bundles
    - Test HomeBundleCard hides when no bundles available
    - Test ImpactCard renders metrics with correct formatting
    - **Property 12: Impact metrics consistency** (verify frontend display matches API data)
    - **Validates: Requirements 5.1, 5.6, 6.1, 6.2, 6.3**

- [x] 13. Checkpoint - Customer portal complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Baker dashboard - bundle creation
  - [x] 14.1 Create `BundleForm` component
    - Name input (required, max 100 chars)
    - Type selector: "composé" / "surprise" radio buttons
    - For composé: dynamic items list (product reference or free-text description + quantity) with add/remove
    - For surprise: description textarea (max 200 chars) + estimated value input (cents)
    - Photo URL input
    - Pricing: original price and discounted price inputs (cents, validation: discounted < original)
    - Quantity total input (min 1)
    - Pickup window: start time + end time inputs (HH:MM, validation: start < end)
    - Client-side validation with inline errors
    - Submit calls createBundle API
    - _Requirements: 2.1, 2.7, 1.4, 1.5_

  - [x] 14.2 Create `DashboardBundlesPage` in baker portal
    - List baker's bundles (all statuses) in a table/card layout
    - Show status badges (draft, published, expired, sold_out)
    - "Publier" button on draft bundles calling publishBundle API
    - "Nouveau panier" button opening BundleForm
    - _Requirements: 2.2, 2.6_

  - [x] 14.3 Wire baker bundle pages into dashboard routing
    - Add route for bundles management page in baker portal
    - Add navigation link in baker dashboard sidebar
    - _Requirements: 2.6_

  - [ ]* 14.4 Write unit tests for BundleForm
    - Test validation: missing fields, invalid prices, invalid time window
    - Test type switching shows/hides appropriate fields
    - Test form submission with valid data
    - **Validates: Requirements 2.1, 2.7, 1.4, 1.5**

- [x] 15. Integration and wiring
  - [x] 15.1 Add `/paniers-du-soir` route to frontend router
    - Register BundlePage component at the route
    - Add navigation link in appropriate customer portal navigation
    - _Requirements: 3.1_

  - [x] 15.2 Handle WebSocket bundle events in existing WS client
    - Listen for `bundle_stock_update` and `bundle_expired` event types
    - On `bundle_stock_update`: update bundle query cache with new quantityRemaining and status
    - On `bundle_expired`: remove bundle from published list cache
    - _Requirements: 8.1, 8.2, 8.3, 4.10_

  - [x] 15.3 Verify end-to-end data flow
    - Ensure bundle list endpoint returns bakery name + coordinates for distance computation
    - Ensure customer geolocation is requested and handled (permission denied fallback)
    - Verify language persistence across page visits for bundle content
    - _Requirements: 3.10, 3.11, 9.3_

- [x] 16. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties using `rapid` (Go) and vitest (frontend)
- Migration number is 018, following existing sequence (017_add_user_locale.sql)
- Backend uses Go with chi router; frontend uses React + TypeScript + TanStack Query
- WebSocket broadcasting uses existing `ws.Hub` infrastructure
- Prices are always in cents (BIGINT); distance is computed client-side from bakery lat/lng
- Baker-authored content (bundle names, descriptions, items) is not auto-translated — displayed as entered

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "1.3", "1.4"] },
    { "id": 1, "tasks": ["1.5", "2.1"] },
    { "id": 2, "tasks": ["2.2", "3.1"] },
    { "id": 3, "tasks": ["3.2", "5.1", "5.2"] },
    { "id": 4, "tasks": ["5.3", "6.1", "6.2", "6.3"] },
    { "id": 5, "tasks": ["6.4", "8.1", "8.2", "9.1"] },
    { "id": 6, "tasks": ["8.3", "9.2", "10.1"] },
    { "id": 7, "tasks": ["10.2", "10.3", "11.1"] },
    { "id": 8, "tasks": ["10.4", "11.2", "12.1", "12.2"] },
    { "id": 9, "tasks": ["11.3", "12.3", "12.4"] },
    { "id": 10, "tasks": ["14.1", "14.2"] },
    { "id": 11, "tasks": ["14.3", "14.4", "15.1", "15.2"] },
    { "id": 12, "tasks": ["15.3"] }
  ]
}
```
  