# Implementation Plan: E2E Playwright Tests

## Overview

Set up a Playwright-based E2E test suite in a root-level `e2e/` directory with its own package.json. The suite uses the Page Object Model, custom fixtures for pre-authenticated contexts, and covers all 13 requirements through 10 spec files. Tests are example-based (no property-based testing).

## Tasks

- [x] 1. Set up Playwright infrastructure
  - [x] 1.1 Create `e2e/package.json` and install Playwright
    - Create `e2e/package.json` with `@playwright/test` as a dev dependency
    - Add scripts: `"test": "playwright test"`, `"test:headed": "playwright test --headed"`, `"report": "playwright show-report"`
    - _Requirements: 1.1, 1.3_

  - [x] 1.2 Create `e2e/tsconfig.json`
    - Configure TypeScript for the e2e directory with strict mode
    - Set `module: "ESNext"`, `moduleResolution: "bundler"`, target ES2020+
    - Include all `.ts` files under the e2e directory
    - _Requirements: 1.1_

  - [x] 1.3 Create `e2e/playwright.config.ts`
    - Configure testDir to `./tests`, timeout 30s, retries 1
    - Set baseURL to `http://localhost:5173`
    - Enable trace on first retry, screenshot on failure
    - Configure single Chromium project
    - Add webServer array: Go backend (`go run ./cmd/server` from `..`, port 8080) and Vite frontend (`npm run dev` from `../frontend`, port 5173)
    - Use `reuseExistingServer: !process.env.CI`
    - _Requirements: 1.2, 1.4_

  - [x] 1.4 Create `e2e/helpers/auth.ts` — authentication helper
    - Implement `loginAsUser(page, credentials)` that calls `POST /api/auth/login` and injects JWT into localStorage under key `auth_token`
    - Implement `createAuthenticatedContext(browser, credentials, baseURL)` that creates a new BrowserContext with the token pre-set
    - Export `LoginCredentials` and `AuthToken` interfaces
    - _Requirements: 1.5_

  - [x] 1.5 Create `e2e/helpers/test-data.ts` — test data constants
    - Define `USERS` object with admin, baker (baker_jean), baker2 (baker_marie), customer (alice), customer2 (bob) credentials and roles
    - Define `BAKERIES` with ids and names for the 3 seeded bakeries
    - Define `REGISTRATION_CODE = 'DEMO1234'`
    - Define `SEEDED_COUNTS` with bakeries: 3, products: 15, orders: 3, reservations: 2, recurringOrders: 2
    - _Requirements: 1.5_

  - [x] 1.6 Create `e2e/fixtures/auth.fixture.ts` — custom authenticated fixtures
    - Extend `@playwright/test` base test with `customerPage`, `bakerPage`, `adminPage` fixtures
    - Each fixture uses `createAuthenticatedContext` to provide a pre-authenticated Page
    - `customerPage` authenticates as alice/customer123
    - `bakerPage` authenticates as baker_jean/baker123
    - `adminPage` authenticates as admin/admin123
    - Export the extended `test` and `expect` from this module
    - _Requirements: 1.5_

- [x] 2. Create page objects
  - [x] 2.1 Create `e2e/page-objects/LoginPage.ts`
    - Encapsulate selectors for username input, password input, submit button, error message
    - Implement `goto()`, `login(username, password)`, `getErrorMessage()` methods
    - Use accessible locators (`getByRole`, `getByLabel`, `getByText`) where possible
    - _Requirements: 8.1, 9.1_

  - [x] 2.2 Create `e2e/page-objects/BakeriesPage.ts`
    - Encapsulate selectors for bakery cards, search/filter input
    - Implement `goto()`, `getBakeryCount()`, `clickBakery(name)` methods
    - _Requirements: 2.1, 5.1_

  - [x] 2.3 Create `e2e/page-objects/BakeryDetailPage.ts`
    - Encapsulate selectors for bakery info, order button, reservation button, product list, allergen indicators
    - Implement `goto(id)`, `clickOrder()`, `clickReservation()`, `getProductNames()`, `hoverAllergenIcon(productName)`, `clickAllergenIcon(productName)` methods
    - _Requirements: 2.2, 2.3, 3.1, 5.2, 12.1, 12.2_

  - [x] 2.4 Create `e2e/page-objects/SchedulePage.ts`
    - Encapsulate selectors for order list, reservation list, cancel buttons
    - Implement `goto()`, `getOrders()`, `getReservations()`, `cancelOrder(index)`, `confirmCancellation()` methods
    - _Requirements: 2.5, 3.3, 4.1, 4.2, 4.3_

  - [x] 2.5 Create `e2e/page-objects/DashboardPage.ts`
    - Encapsulate selectors for dashboard overview stats and navigation links
    - Implement `goto()`, `getStats()`, `navigateToProducts()`, `navigateToOrders()` methods
    - _Requirements: 6.1_

  - [x] 2.6 Create `e2e/page-objects/DashboardProductsPage.ts`
    - Encapsulate selectors for product table/list, add product form, edit/delete buttons
    - Implement `goto()`, `getProductCount()`, `addProduct(name, price, category)`, `editProductPrice(name, newPrice)`, `deleteProduct(name)`, `hasProduct(name)` methods
    - _Requirements: 6.2, 6.3, 6.4, 6.5_

  - [x] 2.7 Create `e2e/page-objects/DashboardOrdersPage.ts`
    - Encapsulate selectors for orders list, status dropdowns/buttons
    - Implement `goto()`, `getPendingOrders()`, `changeOrderStatus(orderId, newStatus)` methods
    - _Requirements: 7.1, 7.2_

- [x] 3. Checkpoint - Verify infrastructure compiles
  - Ensure TypeScript compiles without errors (`npx tsc --noEmit` from e2e/), ask the user if questions arise.

- [x] 4. Write authentication and access control spec files
  - [x] 4.1 Create `e2e/tests/auth.spec.ts`
    - Test: registration form is displayed at /register
    - Test: successful registration with valid credentials and code DEMO1234 redirects to login/dashboard
    - Test: registration with invalid code shows error message
    - Test: login with invalid credentials shows error message
    - Test: login with empty fields shows validation errors
    - Test: login with valid credentials redirects away from login page
    - Use LoginPage page object
    - _Requirements: 8.1, 8.2, 8.3, 9.1, 9.2, 9.3_

  - [x] 4.2 Create `e2e/tests/guest-access.spec.ts`
    - Test: guest can view /bakeries without login
    - Test: guest can view /bakeries/:id without login
    - Test: guest navigating to /schedule is redirected to /login
    - Test: guest navigating to /recurring is redirected to /login
    - Test: guest navigating to /dashboard is redirected to /login
    - Use BakeriesPage and BakeryDetailPage page objects
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 10.1, 10.2, 10.3_

- [x] 5. Write customer flow spec files
  - [x] 5.1 Create `e2e/tests/customer-order.spec.ts`
    - Use `customerPage` fixture for authenticated customer context
    - Test: bakery list displays all seeded bakeries
    - Test: clicking bakery card navigates to /bakeries/:id
    - Test: clicking order action opens the Product Selection Modal
    - Test: selecting products and submitting order with date/time creates order and shows confirmation
    - Test: new order appears on /schedule page
    - Use BakeriesPage, BakeryDetailPage, SchedulePage page objects
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [x] 5.2 Create `e2e/tests/customer-reservation.spec.ts`
    - Use `customerPage` fixture for authenticated customer context
    - Test: reservation form is displayed on bakery detail page
    - Test: submitting reservation with valid selections and pickup time creates it successfully
    - Test: new reservation appears on /schedule page
    - Use BakeryDetailPage, SchedulePage page objects
    - _Requirements: 3.1, 3.2, 3.3_

  - [x] 5.3 Create `e2e/tests/customer-management.spec.ts`
    - Use `customerPage` fixture for authenticated customer context
    - Test: existing orders are displayed on /schedule
    - Test: cancelling an order shows confirmation prompt
    - Test: confirming cancellation removes the order from the list
    - Use SchedulePage page object
    - _Requirements: 4.1, 4.2, 4.3_

- [x] 6. Write baker flow spec files
  - [x] 6.1 Create `e2e/tests/baker-dashboard.spec.ts`
    - Use `bakerPage` fixture for authenticated baker context
    - Test: dashboard overview loads with bakery statistics
    - Test: product list is displayed at /dashboard/products
    - Test: adding a new product with name, price, category shows it in the list
    - Test: editing product price reflects the update in the list
    - Test: deleting a product removes it from the list
    - Use DashboardPage, DashboardProductsPage page objects
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

  - [x] 6.2 Create `e2e/tests/baker-orders.spec.ts`
    - Use `bakerPage` fixture for authenticated baker context
    - Test: pending orders are displayed at /dashboard/orders
    - Test: changing order status (pending → confirmed) updates the display
    - Use DashboardOrdersPage page object
    - _Requirements: 7.1, 7.2_

- [x] 7. Checkpoint - Verify core specs
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8. Write responsive, allergen, and i18n spec files
  - [x] 8.1 Create `e2e/tests/responsive.spec.ts`
    - Set viewport to 375px width for all tests in this file
    - Test: home page renders without horizontal overflow
    - Test: bakery list displays cards in single-column layout
    - Test: navigation is accessible via mobile menu (hamburger or equivalent)
    - Test: full customer order flow completes on mobile viewport
    - Use BakeriesPage, BakeryDetailPage page objects and `customerPage` fixture
    - _Requirements: 11.1, 11.2, 11.3, 11.4_

  - [x] 8.2 Create `e2e/tests/allergens.spec.ts`
    - Use `customerPage` fixture for authenticated customer context
    - Test: hovering allergen indicator icon shows tooltip with allergen name
    - Test: clicking allergen icon opens detail modal with full information
    - Test: closing allergen modal removes it from the DOM
    - Use BakeryDetailPage page object
    - _Requirements: 12.1, 12.2, 12.3_

  - [x] 8.3 Create `e2e/tests/i18n.spec.ts`
    - Test: switching language via language selector updates page content
    - Test: navigation labels, headings, and button text change to target language
    - Test: language selection persists across page navigation
    - _Requirements: 13.1, 13.2, 13.3_

- [x] 9. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- No property-based tests — E2E tests are inherently example-based against seeded data
- Each spec file maps directly to one or more requirements for full traceability
- The `webServer` config auto-starts backend and frontend; no manual server management needed
- Tests use API-based login (not UI login) for speed and reliability in fixtures
- The backend uses in-memory repos reset on restart, ensuring consistent seeded data per test run

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2"] },
    { "id": 1, "tasks": ["1.3", "1.5"] },
    { "id": 2, "tasks": ["1.4", "1.6"] },
    { "id": 3, "tasks": ["2.1", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7"] },
    { "id": 4, "tasks": ["4.1", "4.2", "5.1", "5.2", "5.3", "6.1", "6.2"] },
    { "id": 5, "tasks": ["8.1", "8.2", "8.3"] }
  ]
}
```
