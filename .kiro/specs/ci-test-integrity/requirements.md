# Requirements: CI/Test Integrity

## Introduction

Restore CI pipeline integrity by fixing deterministically failing tests that block E2E execution, repairing CI E2E wiring, fixing flaky property tests, replacing test theater (tests that pass without testing anything real), and adding missing E2E coverage for the Comptoir and pro baker portals.

## Linked Ticket

MA-67 — CI/test integrity: red suite blocks E2E, broken CI wiring, test theater cleanup

## Glossary

- **CI_Pipeline**: The GitHub Actions workflow (`ci.yml`) that runs backend tests, frontend tests, and E2E tests in sequence.
- **Frontend_Suite**: The Vitest-based unit/property test suite under `frontend/`.
- **E2E_Suite**: The Playwright-based end-to-end test suite under `e2e/`.
- **Test_Theater**: Tests that pass without meaningfully asserting the behavior they claim to verify (TODO stubs, no-spy assertions, wrong import targets).
- **Seed_Data**: Demo users, bakeries, products, and orders loaded by `seedDemoData()` in in-memory mode.
- **ThemeSwitcher**: The single-button cycling theme toggle component (`ThemeSwitcher.tsx`).
- **Bundle_Utils**: Pure pricing functions (`calculateBundlePrice`, `capQuantity`) in `bundle-utils.ts`.
- **StockStepper**: The numeric increment/decrement component for stock quantities.
- **DashboardBundles**: The routed bundle composer page at `/dashboard/bundles`.
- **Comptoir_Portal**: The B2B ordering portal accessible at `/comptoir`.
- **Baker_Portal**: The pro baker dashboard accessible at `/dashboard`.

---

## Requirement 1: Frontend Suite Passes (ThemeSwitcher Fix)

**User Story:** As a developer, I want the frontend test suite to be fully green, so that the CI E2E job is no longer blocked by upstream failures.

### Acceptance Criteria

1.1 THE Frontend_Suite SHALL pass all tests with zero failures when run via `npx vitest run`.

1.2 WHEN the ThemeSwitcher tests execute, THE Frontend_Suite SHALL test against the current single-button cycling implementation (not the removed radiogroup UI).

1.3 WHEN the ThemeSwitcher button is clicked, THE ThemeSwitcher tests SHALL verify the theme cycles through light, system, and dark in the correct order.

1.4 WHEN a theme preference is persisted in localStorage, THE ThemeSwitcher tests SHALL verify the component renders the matching label on mount.

---

## Requirement 2: CI E2E Job Executes Successfully

**User Story:** As a developer, I want the E2E job to actually run in CI (not be skipped due to upstream failures), so that regressions are caught before merge.

### Acceptance Criteria

2.1 WHEN the Frontend_Suite passes, THE CI_Pipeline SHALL proceed to run the E2E_Suite without being blocked.

2.2 THE CI_Pipeline SHALL use a coherent E2E strategy where either Playwright `webServer` starts services OR the workflow starts them manually, but not both conflicting approaches.

2.3 WHEN the CI E2E job starts services manually, THE CI_Pipeline SHALL omit the `reuseExistingServer` / `webServer` config from Playwright in CI mode.

2.4 THE CI_Pipeline SHALL run the backend in in-memory mode (no `DATABASE_URL`) so that Seed_Data is loaded automatically and test users exist.

2.5 WHEN the CI "Wait for services" step probes readiness, THE CI_Pipeline SHALL use the correct health endpoint path (`/health`) and SHALL fail the job if services do not start within the timeout (no `|| true` swallowing).

---

## Requirement 3: Bundle Pricing Property Test Fix

**User Story:** As a developer, I want the bundle-utils property test to be non-flaky, so that CI results are trustworthy.

### Acceptance Criteria

3.1 WHEN a bundle contains items totaling 1 cent and a 1% discount factor is applied, THE Bundle_Utils `calculateBundlePrice` function SHALL produce a `discountedPrice` that is less than or equal to `originalPrice`.

3.2 THE Bundle_Utils property test SHALL only assert strict decrease (`discountedPrice < originalPrice`) when the raw discount (`originalPrice * discountFactor`) rounds to at least 1 cent (i.e., `rawDiscount >= 0.5`).

3.3 THE Bundle_Utils property test SHALL pass 200 iterations of fast-check without counterexamples.

---

## Requirement 4: Test Theater Cleanup

**User Story:** As a developer, I want every test to assert the behavior it claims, so that passing tests provide real confidence in correctness.

### Acceptance Criteria

4.1 WHEN the full-journey E2E spec runs, THE E2E_Suite SHALL execute real assertions for all steps of the order lifecycle (no TODO placeholders remaining).

4.2 THE StockStepper property tests SHALL render the actual StockStepper component and verify bounds through the rendered UI (not by re-implementing clamping logic inline).

4.3 THE DashboardBundles test SHALL import and render the routed `DashboardBundles` component (the default export from `pages/dashboard/DashboardBundles.tsx`), not an unrouted page component.

4.4 WHEN `TestInitiateRefund_CallsGatewayRefundPayment` runs, THE test SHALL use a spy on the payment gateway and assert that `RefundPayment` was called with the correct payment intent ID and amount.

---

## Requirement 5: CI Health Check Correctness

**User Story:** As a developer, I want the CI health check to detect boot failures, so that E2E tests don't run against a dead backend.

### Acceptance Criteria

5.1 THE CI_Pipeline "Wait for services" step SHALL probe `/health` (not `/api/health`) for backend readiness.

5.2 IF the backend or frontend does not become ready within the timeout, THEN THE CI_Pipeline SHALL fail the job (no silent `|| true` fallback).

5.3 THE CI_Pipeline SHALL use `curl -sf` with a `timeout` wrapper that exits non-zero on failure.

---

## Requirement 6: E2E Coverage for Comptoir and Baker Portals

**User Story:** As a developer, I want at least one E2E spec each for the Comptoir portal and the pro baker portal, so that critical user flows have automated regression coverage.

### Acceptance Criteria

6.1 THE E2E_Suite SHALL include at least one spec file exercising the Comptoir_Portal with a B2B user login and page navigation.

6.2 THE E2E_Suite SHALL include at least one spec file exercising the Baker_Portal with a seller user login, dashboard navigation, and at least one meaningful interaction.

6.3 WHEN E2E specs for Comptoir and Baker portals run, THE E2E_Suite SHALL assert visible content and correct URL routing, not just absence of errors.

6.4 THE E2E specs SHALL use the existing auth fixture pattern (`b2bPage`, `bakerPage`) from `auth.fixture.ts`.
