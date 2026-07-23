# Implementation Plan: Baker Portal Redesign

## Overview

Rebuild the baker Pro portal pages from table-based English views to card-based, kanban-style French views. Work proceeds component-by-component: shared components first, then each page (Overview, Orders, Products, Bundles). All UI text is in French. No new API endpoints — reuse existing `seller.ts` functions.

## Tasks

- [x] 1. Create shared UI components (FilterChips, StatCard, StockStepper)
  - [x] 1.1 Create FilterChips component
    - Create `frontend/src/components/pro/FilterChips.tsx` and `FilterChips.css`
    - Generic single-selection chip row with active (filled blue) and inactive (outlined) states
    - Accept `options`, `selected`, `onChange` props with TypeScript generics
    - _Requirements: 6.1, 6.2_
  - [x] 1.2 Create StatCard component
    - Create `frontend/src/components/pro/StatCard.tsx` and `StatCard.css`
    - Display large value, muted label, subtitle, optional colored badge
    - Support badge variants: positive (green), neutral (gray), negative (red)
    - _Requirements: 6.3_
  - [x] 1.3 Create StockStepper component
    - Create `frontend/src/components/pro/StockStepper.tsx` and `StockStepper.css`
    - Inline −/+ with current value, respects min/max bounds, danger prop for red styling
    - _Requirements: 6.4, 6.5_
  - [x]* 1.4 Write property test for StockStepper bounds
    - **Property 4: Stock stepper bounds**
    - Generate random sequences of increment/decrement with random [min, max] bounds, verify value always within bounds
    - **Validates: Requirements 6.4, 5.10**
  - [x]* 1.5 Write unit tests for FilterChips, StatCard, StockStepper
    - Test FilterChips: renders all options, active chip highlighted, click selects new chip
    - Test StatCard: renders label, value, subtitle, optional badge
    - Test StockStepper: renders value, danger styling, prevents out-of-bounds
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 2. Create OrderCard component
  - [x] 2.1 Create OrderCard component
    - Create `frontend/src/components/pro/OrderCard.tsx` and `OrderCard.css`
    - Display order time, item summary, type badge (livraison/retrait), action button
    - Blue left-border accent when status is "preparing"
    - Action button text: "Commencer", "Prêt ✓", "Remis ✓" based on current status
    - _Requirements: 3.6, 3.7_
  - [x]* 2.2 Write unit tests for OrderCard
    - Test renders time, items, type badge
    - Test action button matches next valid transition
    - Test blue border when status is "preparing"
    - _Requirements: 3.6, 3.7_

- [x] 3. Update DashboardLayout sidebar
  - [x] 3.1 Update nav items to French labels and correct routes
    - Change labels to: "Tableau de bord", "Commandes", "Menu & stock", "Paniers du soir", "Statistiques", "Boutique"
    - Update route from `/dashboard/reservations` to `/dashboard/bundles`
    - Ensure badge on "Commandes" shows confirmed order count
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - [x]* 3.2 Update DashboardLayout tests
    - Verify French labels rendered
    - Verify badge displayed when orders exist
    - _Requirements: 1.1, 1.2_

- [x] 4. Checkpoint - Shared components complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Rebuild DashboardOverview page
  - [x] 5.1 Rebuild DashboardOverview with KPI cards and 2-column grid
    - Personalized greeting: "Bonjour [name] ☀️" with French-formatted date
    - 3 StatCard KPIs: today's order count, next pickup/delivery time, today's revenue
    - "À préparer maintenant" section with OrderCards for confirmed orders
    - "Stock faible ⚠️" section listing products at/below low stock threshold
    - Golden anti-gaspi CTA card with estimated unsold value and link to `/dashboard/bundles`
    - Shop open/closed toggle in header
    - Update CSS in `DashboardOverview.css` for the 2-column grid layout
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_
  - [x]* 5.2 Write unit tests for DashboardOverview
    - Test greeting contains baker name and French date
    - Test KPI stat cards render with correct values from mock data
    - Test confirmed orders appear in "À préparer" section
    - Test low-stock products listed when below threshold
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [x] 6. Rebuild DashboardOrders as kanban board
  - [x] 6.1 Implement kanban column grouping logic
    - Create `frontend/src/pages/dashboard/kanban-utils.ts`
    - Export `groupOrdersByStatus(orders): Map<OrderStatus, Order[]>` — partitions orders into 4 columns
    - Export `isAdjacentTransition(from, to): boolean` — validates column transitions
    - _Requirements: 3.2, 3.4_
  - [x]* 6.2 Write property tests for kanban grouping logic
    - **Property 1: Kanban order conservation** — no orders lost or duplicated across columns
    - **Property 2: Kanban column assignment by status** — each order in correct column
    - **Validates: Requirements 3.2**
  - [x] 6.3 Implement DashboardOrders kanban UI
    - Rebuild `DashboardOrders.tsx` with 4-column kanban layout
    - Header: "Commandes — [jour]" with date picker
    - FilterChips for delivery type: Livraison, Retrait, Toutes
    - Each column: label, count badge, scrollable list of OrderCards
    - Action buttons on cards trigger `updateOrderStatus` API call
    - _Requirements: 3.1, 3.5, 3.6, 3.8_
  - [x] 6.4 Implement HTML5 drag-and-drop between columns
    - Drag handlers on OrderCard (draggable, onDragStart, onDragEnd)
    - Drop handlers on KanbanColumn (onDragOver, onDrop)
    - Validate adjacent-column constraint; snap back + toast on invalid drop
    - Call `updateOrderStatus` on valid drop
    - _Requirements: 3.3, 3.4_
  - [x]* 6.5 Write property test for filter chip consistency
    - **Property 3: Filter chip consistency** — filtered orders are correct subset matching filter
    - **Validates: Requirements 3.5, 4.7**
  - [x]* 6.6 Write unit tests for DashboardOrders
    - Test 4 columns rendered with correct labels
    - Test filter chips show/hide orders by type
    - Test drag to adjacent column triggers API call
    - Test drag to non-adjacent column shows toast and reverts
    - _Requirements: 3.1, 3.3, 3.4, 3.5_

- [x] 7. Checkpoint - Orders kanban complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8. Rebuild DashboardProducts as card-based inventory
  - [x] 8.1 Create ProductCard component
    - Create `frontend/src/components/pro/ProductCard.tsx` and `ProductCard.css`
    - Display photo (lazy-loaded), name, description, price (€), allergen chips, StockStepper
    - "en vente"/"masqué" toggle badge; dimmed styling when hidden
    - _Requirements: 4.2, 4.5, 4.6, 8.2_
  - [x] 8.2 Rebuild DashboardProducts page
    - Replace table with card grid grouped by category
    - Category FilterChips at top: Viennoiseries, Pains, Pâtisseries
    - "+ Nouveau produit" button opens product creation form
    - Inline stock editing via StockStepper calls `updateProduct` API
    - Day-availability toggles (L, M, M, J, V, S, D) at page bottom
    - "le stock se remet à zéro chaque soir ↺" note
    - _Requirements: 4.1, 4.3, 4.7, 4.8, 4.9, 4.10_
  - [x]* 8.3 Write unit tests for DashboardProducts
    - Test products rendered as cards grouped by category
    - Test category filter shows only matching products
    - Test stock stepper calls API on +/−
    - Test visibility toggle updates product and dims card
    - _Requirements: 4.1, 4.2, 4.3, 4.5, 4.6, 4.7_

- [x] 9. Rebuild DashboardBundles as bundle composer
  - [x] 9.1 Implement bundle price calculation logic
    - Create `frontend/src/pages/dashboard/bundle-utils.ts`
    - Export `calculateBundlePrice(items): { originalPrice, discountedPrice }` — computes bundle pricing
    - Export `capQuantity(requested, remaining): number` — caps at remaining stock
    - _Requirements: 5.6, 5.4_
  - [x]* 9.2 Write property tests for bundle logic
    - **Property 5: Bundle price discount invariant** — discounted < original when items selected
    - **Property 6: Bundle item quantity bounds** — quantity never exceeds remaining stock
    - **Validates: Requirements 5.4, 5.6**
  - [x] 9.3 Implement DashboardBundles page UI
    - Rename from DashboardReservations; update route registration
    - Left panel: product checklist with checkboxes, "reste X" labels, quantity steppers
    - Right panel: live client preview card (warm cream palette)
    - Basket count stepper (min 1), pickup time window selectors
    - "Publier les paniers" button (disabled when no items selected)
    - Golden anti-gaspi badge in header
    - _Requirements: 5.1, 5.2, 5.3, 5.5, 5.7, 5.8, 5.9, 5.10_
  - [x]* 9.4 Write unit tests for DashboardBundles
    - Test product list renders with remaining stock
    - Test checking product adds to preview
    - Test quantity capped at remaining stock
    - Test "Publier" disabled with no selection, enabled with selection
    - Test publish calls reservation API
    - _Requirements: 5.1, 5.2, 5.4, 5.8, 5.9_

- [x] 10. Implement error handling patterns
  - [x] 10.1 Add error states and retry logic to all pages
    - Create shared error display component with French message and "Réessayer" button
    - Wrap API calls with error boundaries that retain stale data
    - Add conflict detection on stock updates (409 handling)
    - Add drag-and-drop revert on API failure with toast
    - _Requirements: 7.1, 7.2, 7.3, 7.4_
  - [x]* 10.2 Write unit tests for error handling
    - Test error message displayed on API failure
    - Test previously loaded data retained on error
    - Test card reverts on drag-drop API failure
    - _Requirements: 7.1, 7.2, 7.4_

- [x] 11. Update route configuration
  - [x] 11.1 Update router to register new/renamed routes
    - Change `/dashboard/reservations` to `/dashboard/bundles`
    - Ensure all nav links match updated routes
    - Verify lazy-loading of page components
    - _Requirements: 1.1_

- [x] 12. Final checkpoint - All pages complete
  - Ensure all tests pass, ask the user if questions arise.

## Task Dependency Graph

```json
{
  "waves": [
    {
      "name": "Wave 1 - Shared Components",
      "tasks": ["1", "2", "3"],
      "description": "Build shared UI components, OrderCard, and update sidebar"
    },
    {
      "name": "Wave 2 - Page Rebuilds",
      "tasks": ["5", "6", "8", "9"],
      "description": "Rebuild all dashboard pages using shared components",
      "dependsOn": ["1", "2", "3"]
    },
    {
      "name": "Wave 3 - Polish",
      "tasks": ["10", "11"],
      "description": "Error handling and route configuration",
      "dependsOn": ["5", "6", "8", "9"]
    }
  ]
}
```

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- All UI text is in French — no i18n keys needed for new components (hardcoded French per design)
- No new API endpoints required — reuse existing `seller.ts` functions
- CSS uses custom properties from `pro-theme.css` (already exists)
- Drag-and-drop uses HTML5 native API — no additional library needed
- Property tests use `fast-check` (already in devDependencies)
- Test files co-located next to source: `Component.test.tsx`
