# Implementation Plan: Health & Allergen Indicators

## Overview

This plan implements health scores (numeric 1–5, nullable) and allergen indicators on bakery products. It covers the database migration, Go domain validation/API changes, baker portal form inputs, customer portal display components (per-product allergen indicator + page-level allergen info icon), and i18n translations for EN/FR/NL.

## Tasks

- [x] 1. Database migration and Go domain layer
  - [x] 1.1 Create database migration `012_add_health_allergens.sql`
    - Add `allergens TEXT[] NOT NULL DEFAULT '{}'` column to products table
    - Add `health_score INTEGER NULL` column to products table
    - Add CHECK constraint `chk_health_score` ensuring value is NULL or between 1 and 5
    - Write goose Down migration to drop columns and constraint
    - _Requirements: 1.1, 1.2, 1.5, 1.8_

  - [x] 1.2 Create `internal/domain/allergens.go` with allergen constants and validation
    - Define `Allergen` type and 14 EU-regulated allergen constants
    - Create `ValidAllergens` map for O(1) membership checks
    - Implement `ValidateAllergens(allergens []string) error` — checks membership, uniqueness, max 20 elements
    - Implement `ValidateHealthScore(score *int) error` — checks nil or 1–5
    - _Requirements: 1.6, 1.7, 9.10, 9.11, 9.12_

  - [x] 1.3 Update `internal/domain/models.go` Product struct
    - Add `Allergens []string` field with JSON tag `allergens`
    - Add `HealthScore *int` field with JSON tag `healthScore`
    - _Requirements: 1.1, 1.2_

  - [x]* 1.4 Write property tests for allergen and health score validation
    - **Property 3: Invalid allergens are rejected**
    - **Property 4: Invalid health scores are rejected**
    - Use `rapid` library in `internal/domain/` test file
    - Generate random strings not in valid set → assert ValidateAllergens returns error
    - Generate random integers outside [1,5] → assert ValidateHealthScore returns error
    - **Validates: Requirements 1.6, 1.7, 9.10, 9.11**

- [x] 2. Backend API changes (seller handler)
  - [x] 2.1 Update seller handler `CreateProduct` to accept and validate allergens + health_score
    - Parse `allergens` (optional, default `[]`) and `healthScore` (optional, default nil) from request body
    - Call `ValidateAllergens` and `ValidateHealthScore` before persistence
    - Return 400 with descriptive error on validation failure
    - _Requirements: 9.1, 9.2, 9.10, 9.11, 9.12_

  - [x] 2.2 Update seller handler `UpdateProduct` to accept and validate allergens + health_score
    - Check for `allergens` and `healthScore` keys in partial update payload
    - Only validate/update fields that are present in the request
    - Omitted fields remain unchanged in the database
    - Support explicit null for health_score and empty array for allergens
    - _Requirements: 9.3, 9.4, 9.5, 9.6, 9.7, 9.8_

  - [x] 2.3 Update repository layer to persist and retrieve allergens + health_score
    - Update SQL queries in bakery repository for INSERT and UPDATE operations
    - Use `pq.Array` or equivalent for PostgreSQL array column
    - Update SELECT queries to include new columns
    - Ensure menu/product fetch endpoints return allergens and healthScore in response
    - _Requirements: 9.9, 1.3, 1.4_

  - [x]* 2.4 Write property tests for API round-trip behavior
    - **Property 1: Allergen data round-trip**
    - **Property 2: Health score data round-trip**
    - **Property 5: Partial update preserves omitted fields**
    - Use `rapid` to generate valid allergen subsets → create product → fetch → assert equality
    - Use `rapid` to generate valid scores (nil, 1–5) → create → fetch → assert equality
    - Use `rapid` to generate partial updates → assert omitted fields unchanged
    - **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 9.1–9.9**

- [x] 3. Checkpoint - Backend complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Frontend type updates and i18n translations
  - [x] 4.1 Update `frontend/src/types/bakery.ts` Product interface
    - Add `allergens: string[]` field
    - Add `healthScore: number | null` field
    - _Requirements: 1.1, 1.2_

  - [x] 4.2 Add allergen and health score translations to `frontend/src/i18n/translations.ts`
    - Add 14 allergen name keys in EN, FR, NL (42 entries)
    - Add 14 allergen description keys in EN, FR, NL (42 entries, ≤150 chars each)
    - Add health score label and scale explanation in EN, FR, NL
    - Add allergen info modal title, intro paragraph in EN, FR, NL
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [x]* 4.3 Write property test for translation completeness
    - **Property 10: Translation completeness for allergens**
    - **Property 11: Translation fallback chain**
    - Verify every allergen × locale has non-empty name and description (≤150 chars)
    - Verify fallback to EN when locale key missing, then to raw key
    - **Validates: Requirements 8.1, 8.2, 8.4, 8.6**

- [x] 5. Baker portal - allergen and health score inputs
  - [x] 5.1 Create `AllergenMultiSelect` component in baker portal
    - Render 14 checkboxes (one per EU-regulated allergen) in a 2-column layout
    - Accept `selected: string[]` and `onChange: (allergens: string[]) => void` props
    - Use system sans-serif theme consistent with dashboard
    - _Requirements: 2.1, 2.8_

  - [x] 5.2 Create `HealthScoreInput` component in baker portal
    - Render number input with min=1, max=5, step=1, and a clear button
    - Show label: "Health score (1 = least healthy, 5 = healthiest)"
    - Client-side validation: display inline error for values outside 1–5
    - Accept `value: number | null`, `onChange: (score: number | null) => void`, `error?: string`
    - _Requirements: 2.2, 2.9_

  - [x] 5.3 Integrate allergen and health score inputs into `DashboardProducts.tsx` product form
    - Add `AllergenMultiSelect` and `HealthScoreInput` to create/edit product form
    - Pre-populate fields with existing product data when editing
    - Include allergens and healthScore in form submission payload to API
    - Handle save errors: show error toast, retain unsaved form state
    - _Requirements: 2.3, 2.4, 2.5, 2.6, 2.7, 2.8_

  - [x]* 5.4 Write unit tests for baker portal components
    - Test AllergenMultiSelect renders 14 options and toggles correctly
    - Test HealthScoreInput validates range and supports null/clear
    - Test form pre-population with existing product data
    - **Property 12: Baker form accepts all valid allergen/score combinations**
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.8, 2.9**

- [x] 6. Customer portal - health score display
  - [x] 6.1 Create `HealthScoreDisplay` component
    - Render numeric score as a small badge (e.g., "🌿 3/5")
    - Add `aria-label="Health score: {score} out of 5"`
    - Only rendered when score is non-null
    - Use artisan theme styling
    - _Requirements: 6.1, 6.2, 6.6_

  - [x] 6.2 Integrate `HealthScoreDisplay` into product cards and detail views
    - Add to ProductSelectionModal grid cards (visible without hover/expand)
    - Add to BakeryDetailPage product cards
    - Add to product description area in expanded/detail view
    - Do not render when healthScore is null (no empty space or placeholder)
    - _Requirements: 6.3, 6.4, 6.5_

- [x] 7. Customer portal - per-product allergen indicator
  - [x] 7.1 Create `AllergenIndicator` component
    - Render 24×24px icon (20×20px below 768px viewport)
    - Position at bottom-right of product card, overlapping edge
    - Add `aria-label="Contains allergens"`
    - On hover/focus: show tooltip with comma-separated translated allergen names (≤200ms)
    - On click: open AllergenDetailModal; call `event.stopPropagation()`
    - Only render when `allergens.length > 0`
    - _Requirements: 3.1, 3.2, 3.3, 3.6, 3.7, 4.1, 4.3, 4.5, 4.6, 5.6_

  - [x] 7.2 Create `AllergenDetailModal` component
    - Display product name as modal title
    - List allergens alphabetically sorted by translated name in active locale
    - Focus trap: Tab/Shift+Tab cycle within modal
    - Close on Escape or outside click; return focus to trigger element
    - Render above ProductSelectionModal via z-index layering
    - Display content in active language (EN/FR/NL)
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.7, 5.8_

  - [x] 7.3 Integrate `AllergenIndicator` into product cards
    - Add to ProductSelectionModal grid cards
    - Add to BakeryDetailPage desktop grid and mobile product rows
    - Ensure tooltip renders above parent overflow boundaries (no clipping)
    - Ensure tooltip hides immediately on pointer leave (within one animation frame)
    - _Requirements: 3.4, 3.5, 4.2, 4.4_

- [x] 8. Customer portal - page-level allergen info icon and modal
  - [x] 8.1 Create `AllergenInfoIcon` floating button component
    - Fixed position at bottom of viewport (sticky footer), e.g., `bottom: 16px; right: 16px`
    - 40×40px desktop, 36×36px below 768px
    - `aria-label="Allergen information"` (translated)
    - Ensure no overlap with critical interactive elements (Add to cart, navigation)
    - Remains visible while scrolling
    - _Requirements: 7.1, 7.2, 7.9, 7.10_

  - [x] 8.2 Create `AllergenInfoModal` component
    - Title: "Allergen Information" (translated)
    - Intro paragraph explaining allergens and food safety (translated)
    - List all 14 EU-regulated allergens with translated name and description
    - Focus trap: Tab/Shift+Tab cycle within modal
    - Close on Escape or outside click; return focus to AllergenInfoIcon
    - _Requirements: 7.3, 7.4, 7.5, 7.6, 7.7, 7.8_

  - [x] 8.3 Integrate `AllergenInfoIcon` into customer pages
    - Render on BakeryDetailPage
    - Render in ProductSelectionModal context (visible while modal open)
    - _Requirements: 7.1_

  - [x]* 8.4 Write unit tests for allergen info components
    - Test AllergenInfoIcon renders with correct size and aria-label
    - Test AllergenInfoModal displays all 14 allergens with translations
    - Test focus trap and close behavior
    - Test language switching updates content without page reload
    - **Validates: Requirements 7.1–7.10, 8.5**

- [x] 9. Checkpoint - Frontend complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 10. Integration and wiring
  - [x] 10.1 Update seller API client in frontend to send allergens + healthScore
    - Update `frontend/src/api/seller.ts` create/update product functions
    - Include allergens array and healthScore in request body
    - _Requirements: 9.1, 9.2, 9.3, 9.4_

  - [x] 10.2 Verify end-to-end data flow
    - Ensure bakery menu fetch includes allergens and healthScore in response
    - Ensure ProductSelectionModal and BakeryDetailPage receive updated Product objects
    - Verify language persistence in localStorage across page visits
    - _Requirements: 9.9, 8.5, 8.7_

- [x] 11. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties using `rapid` (Go)
- Frontend property-like tests (translation completeness, fallback) can use vitest with exhaustive checks
- Migration number is 012, following existing sequence (011_create_registration_tokens.sql)
- Baker portal uses English for allergen keys (admin-facing); customer portal uses i18n translations

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "1.3"] },
    { "id": 1, "tasks": ["1.4", "2.1", "2.2", "2.3", "4.1", "4.2"] },
    { "id": 2, "tasks": ["2.4", "4.3", "5.1", "5.2", "6.1"] },
    { "id": 3, "tasks": ["5.3", "6.2", "7.1"] },
    { "id": 4, "tasks": ["5.4", "7.2", "7.3", "8.1"] },
    { "id": 5, "tasks": ["8.2", "8.3"] },
    { "id": 6, "tasks": ["8.4", "10.1"] },
    { "id": 7, "tasks": ["10.2"] }
  ]
}
```
