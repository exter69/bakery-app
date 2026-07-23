# Implementation Plan: CI/Test Integrity

## Overview

This plan fixes CI pipeline integrity by ensuring the frontend suite is green (unblocking E2E), verifying CI E2E wiring is coherent, fixing the flaky bundle-utils property test, cleaning up test theater, and confirming E2E portal coverage exists. Based on code inspection, several issues described in the ticket have already been resolved — the tasks focus on verification and any remaining gaps.

## Tasks

- [ ] 1. Verify and fix ThemeSwitcher tests
  - [ ] 1.1 Run `npx vitest run` in `frontend/` and confirm ThemeSwitcher tests pass
    - If any tests still reference radiogroup or old UI patterns, rewrite them to test the single-button cycling component
    - Tests should query by `role('button')` with labels "System theme", "Dark mode", "Light mode"
    - Verify cycle order: system → dark → light → system
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - [ ] 1.2 Confirm full frontend suite is green (226+ tests, 0 failures)
    - Run `npx vitest run` and verify exit code 0
    - Record pass/fail counts
    - _Requirements: 1.1_

- [ ] 2. Fix CI workflow health check and verify E2E wiring
  - [ ] 2.1 Audit `.github/workflows/ci.yml` e2e job for correctness
    - Verify "Wait for services" uses `/health` (not `/api/health`)
    - Verify no `|| true` or error-swallowing in the wait step
    - Verify `curl -sf` is used (fails on HTTP errors and silent mode)
    - Verify `timeout` wrapper will fail the step on timeout
    - Fix any issues found
    - _Requirements: 5.1, 5.2, 5.3_
  - [ ] 2.2 Verify E2E strategy coherence
    - Confirm e2e job does NOT set `DATABASE_URL` (in-memory mode with seed data)
    - Confirm `playwright.config.ts` conditionally omits `webServer` when `CI=true`
    - Confirm workflow starts backend and frontend manually before Playwright runs
    - Fix any inconsistencies
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [ ] 3. Fix bundle-utils property test (flaky counterexample)
  - [ ] 3.1 Verify `calculateBundlePrice` rounding behavior and fix test invariant
    - The counterexample: 1-cent item, 1% discount → `Math.round(1 * 0.99) = 1` → no strict decrease
    - Ensure the "strict decrease" assertion is guarded: only assert `discountedPrice < originalPrice` when `originalPrice * discountFactor >= 0.5`
    - The non-strict invariant (`discountedPrice <= originalPrice`) must always hold
    - Run `npx vitest run bundle-utils` and confirm 200 iterations pass with no counterexamples
    - _Requirements: 3.1, 3.2, 3.3_

- [ ] 4. Checkpoint — Frontend suite fully green
  - Ensure all tests pass (`npx vitest run`), ask the user if questions arise.

- [ ] 5. Verify and fix test theater issues
  - [ ] 5.1 Verify `full-journey.spec.ts` has real assertions at every step
    - Confirm no TODO comments or placeholder steps remain
    - Every step must have an `expect()` assertion or meaningful user action
    - If TODOs remain, implement the assertions (baker advances status, customer sees update)
    - _Requirements: 4.1_
  - [ ] 5.2 Verify StockStepper property tests render the actual component
    - Confirm tests use `render(<StockStepper ... />)` and interact via `screen.getByLabelText`
    - Confirm assertions check `onChange` callback values, not inline re-implementations of clamping
    - If tests re-implement clamping logic without rendering, rewrite to render and click
    - _Requirements: 4.2_
  - [ ] 5.3 Verify DashboardBundles test imports the correct routed component
    - Confirm import is `from './DashboardBundles'` (the default export, routed component)
    - Confirm no `DashboardBundlesPage` unrouted import exists
    - If pointing at wrong component, fix the import
    - _Requirements: 4.3_
  - [ ] 5.4 Verify refund test has spy gateway assertion
    - Confirm `TestInitiateRefund_CallsGatewayRefundPayment` in `internal/payment/service_test.go` uses a spy
    - Confirm it asserts `gateway.refundCalled`, correct payment intent ID, and correct amount
    - If spy is missing, add `spyGateway` struct and assertions
    - _Requirements: 4.4_

- [ ] 6. Verify E2E portal coverage exists
  - [ ] 6.1 Confirm Comptoir portal E2E spec (`e2e/tests/comptoir.spec.ts`)
    - Verify spec uses `b2bPage` fixture for authentication
    - Verify spec navigates to `/comptoir` and asserts content visibility
    - Verify spec tests at least navigation links and page loads
    - If spec is missing or trivial, add meaningful assertions
    - _Requirements: 6.1, 6.3, 6.4_
  - [ ] 6.2 Confirm Baker portal E2E spec (`e2e/tests/baker-portal.spec.ts`)
    - Verify spec uses `bakerPage` fixture for authentication
    - Verify spec navigates to `/dashboard` and asserts sidebar, navigation, content
    - Verify spec tests sub-page navigation (products, bundles)
    - If spec is missing or trivial, add meaningful assertions
    - _Requirements: 6.2, 6.3, 6.4_

- [ ] 7. Final verification checkpoint
  - [ ] 7.1 Run full frontend suite and confirm green
    - `cd frontend && npx vitest run` — expect 0 failures
    - _Requirements: 1.1_
  - [ ] 7.2 Run Go test suite and confirm green
    - `go test -race ./...` — expect all pass including refund spy test
    - _Requirements: 4.4_
  - [ ] 7.3 Verify CI YAML is syntactically valid
    - Check for YAML parse errors in the workflow file
    - _Requirements: 2.1, 5.1_

## Notes

- Based on code inspection, many issues described in MA-67 appear to have already been addressed in a prior commit (ThemeSwitcher tests updated, refund spy added, bundle-utils guard condition added, full-journey TODOs replaced, portal E2E specs exist).
- The primary remaining risks are: (a) any residual test failures not visible in the current files, (b) CI workflow health check correctness.
- Tasks are written as "verify and fix" — if the code is already correct, the task completes quickly with verification. If not, the fix is specified.
- No new test frameworks or dependencies are introduced.
- Property tests use fast-check (already a dependency).

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2"] },
    { "id": 1, "tasks": ["2.1", "2.2", "3.1"] },
    { "id": 2, "tasks": ["4"] },
    { "id": 3, "tasks": ["5.1", "5.2", "5.3", "5.4"] },
    { "id": 4, "tasks": ["6.1", "6.2"] },
    { "id": 5, "tasks": ["7.1", "7.2", "7.3"] }
  ]
}
```
