# Design: CI/Test Integrity

## Overview

This spec addresses six interrelated problems that collectively prevent the CI pipeline from providing meaningful signal:

1. **ThemeSwitcher tests target removed UI** — 6 tests fail deterministically, blocking the E2E job (which `needs: [frontend]`).
2. **CI E2E wiring issues** — the workflow already starts servers manually; Playwright config already conditionally skips `webServer` in CI; the backend runs in in-memory mode with seed data. The remaining issue is the health check probing `/api/health` instead of `/health` and the `|| true` swallowing failures.
3. **Flaky bundle-utils property test** — `calculateBundlePrice` with a 1-cent item and 1% discount produces `discountedPrice == originalPrice` (rounding to 0 cents off), violating the "strict decrease" assertion.
4. **Test theater** — `full-journey.spec.ts` has TODO stubs; the refund test lacked a spy (already fixed); DashboardBundles test already imports the correct component; StockStepper tests already render the real component.
5. **CI health check** — probes wrong endpoint and uses `|| true`.
6. **Missing E2E coverage** — Comptoir and baker portal specs already exist (`comptoir.spec.ts`, `baker-portal.spec.ts`).

## Architecture

No architectural changes are required. This is a test/CI debt cleanup — all changes are to test files, CI configuration, and one utility function's rounding logic.

```
ci.yml (workflow)
  ├─ backend job → go test
  ├─ frontend job → vitest run (must be green)
  └─ e2e job (needs: [backend, frontend])
       ├─ starts backend (in-memory, seed data)
       ├─ builds + serves frontend
       ├─ waits for /health (no || true)
       └─ runs playwright test
```

## Components and Interfaces

### 1. ThemeSwitcher Test (`ThemeSwitcher.test.tsx`)

**Current state:** Tests already updated to match the single-button cycling implementation. They query by `role('button')` with the correct labels ("System theme", "Dark mode", "Light mode") and verify the cycle order.

**Analysis:** The existing test file at `frontend/src/components/ThemeSwitcher.test.tsx` already correctly tests the current implementation. The 6 tests pass. The ticket's description of "6/6 target old radiogroup UI" appears to have been addressed in a prior commit. No changes needed unless tests are still failing.

**Action:** Verify suite is green. If any residual radiogroup references exist elsewhere, remove them.

### 2. CI Workflow (`ci.yml`)

**Current state:** The e2e job correctly:
- Omits `DATABASE_URL` (uses in-memory mode)
- Builds and starts the backend manually
- Builds frontend and serves via `vite preview`
- Conditionally skips Playwright `webServer` when `CI=true`

**Issues to fix:**
- "Wait for services" step: uses correct `/health` endpoint and no `|| true` (already fixed in current file). The current ci.yml shows `curl -sf http://localhost:8080/health` — this is correct.
- Verify no `|| true` exists.

**Analysis:** Looking at the current `ci.yml`, the wait step already uses:
```yaml
timeout 30 bash -c 'until curl -sf http://localhost:8080/health; do sleep 1; done'
timeout 30 bash -c 'until curl -sf http://localhost:5173; do sleep 1; done'
```
No `|| true` present. The endpoint is `/health` (correct). This appears already fixed.

### 3. Playwright Config (`e2e/playwright.config.ts`)

**Current state:** Already conditionally applies `webServer` only when `!process.env.CI`:
```typescript
...(!process.env.CI && { webServer: [...] })
```

**Analysis:** No conflict exists. In CI, servers are started by the workflow; locally, Playwright starts them. This is coherent.

### 4. Bundle Utils (`bundle-utils.ts`)

**Current state:** `calculateBundlePrice` uses `Math.round(originalPrice * (1 - discountFactor))`. For `originalPrice = 1` (1 cent) and `discountFactor = 0.01` (1%), the discount is `0.01` cents, which rounds to 0, so `discountedPrice = 1 = originalPrice`. The "strict decrease" assertion fails.

**Fix options:**
- **Option A (fix the invariant):** Only assert strict decrease when `rawDiscount >= 0.5` (i.e., when the discount rounds to at least 1 cent). This is mathematically sound — you can't have a strict decrease when the discount is sub-cent and rounds to zero.
- **Option B (fix the rounding):** Use `Math.floor` instead of `Math.round` to always round down, guaranteeing at least 1 cent off. But this changes business behavior for edge cases.

**Decision:** Option A — fix the test invariant. The test at `bundle-utils.test.ts:22-32` already implements this guard (`if (rawDiscount >= 0.5)`). The existing test file shows the corrected invariant. This appears already fixed.

### 5. Full Journey Spec (`full-journey.spec.ts`)

**Current state:** The spec has 11 steps with real assertions. The ticket mentioned steps 8-14 being TODO, but the current file shows:
- Step 8: Verifies order count in schedule (`expect(orderCount).toBeGreaterThan(0)`)
- Step 9: Baker sees orders (`expect(bakerOrderCount).toBeGreaterThan(0)`)
- Step 10: Baker advances status
- Step 11: Verifies status changed

**Analysis:** The spec has real assertions at every step. No TODO comments remain. This appears already fixed. The original claim of "15 steps with 8-14 as TODO" no longer matches the current state (11 steps, all with assertions).

### 6. StockStepper Test (`StockStepper.test.tsx`)

**Current state:** The property tests render the actual `<StockStepper>` component, click its increment/decrement buttons via `screen.getByLabelText('Augmenter'/'Diminuer')`, track onChange calls, and rerender with updated values. The assertions check bounds on the actual component output.

**Analysis:** The ticket claimed "property tests re-implement the clamping inline and never render the component." Looking at the actual code, this is incorrect — the tests DO render the component and interact through the UI. No changes needed.

### 7. DashboardBundles Test (`DashboardBundles.test.tsx`)

**Current state:** Imports `DashboardBundles` from `./DashboardBundles` — this IS the routed component (the default export from `pages/dashboard/DashboardBundles.tsx`). The test mocks `../../api/seller` and tests real rendering behavior.

**Analysis:** The ticket claimed the test "imports the unrouted DashboardBundlesPage" but no such component exists. The import points to the correct file. No changes needed.

### 8. Refund Test (`service_test.go`)

**Current state:** `TestInitiateRefund_CallsGatewayRefundPayment` uses a `spyGateway` and asserts:
```go
assert.True(t, gateway.refundCalled)
assert.Equal(t, "pi_captured_123", gateway.refundPaymentIntentID)
assert.Equal(t, int64(5000), gateway.refundAmountCents)
```

**Analysis:** Already fixed with proper spy verification.

### 9. E2E Portal Coverage

**Current state:** Both `comptoir.spec.ts` (5 tests) and `baker-portal.spec.ts` (5 tests) already exist with meaningful assertions about navigation, content visibility, URL routing, and access control.

## Data Models

No data model changes required. This is a test/CI-only change.

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Bundle discount invariant (non-strict)

*For any* non-empty list of selected items with positive prices (in cents) and any discount factor in (0, 1), `calculateBundlePrice` SHALL produce a `discountedPrice` that satisfies `0 <= discountedPrice <= originalPrice`.

**Validates: Requirements 3.1**

### Property 2: Bundle discount strict decrease (guarded)

*For any* non-empty list of selected items and any discount factor where `originalPrice * discountFactor >= 0.5` (discount rounds to at least 1 cent), `calculateBundlePrice` SHALL produce `discountedPrice < originalPrice`.

**Validates: Requirements 3.2**

### Property 3: StockStepper bounds invariant

*For any* min/max bounds and any sequence of increment/decrement operations on the rendered StockStepper component, the value passed to `onChange` SHALL always satisfy `min <= value <= max`.

**Validates: Requirements 4.2**

## Error Handling

- CI "Wait for services" timeout: job fails immediately with non-zero exit (no silent fallback).
- If frontend suite has failures, the e2e job is skipped (by `needs: [frontend]` dependency).
- Property tests that find counterexamples fail the suite with the shrunk example logged.

## Testing Strategy

### Unit Tests (Vitest)
- ThemeSwitcher: 6 example-based tests covering all cycle transitions and persistence
- DashboardBundles: 5 example-based tests covering rendering, interactions, loading/empty states
- Bundle-utils: property tests (200 iterations) + empty-selection edge case
- StockStepper: property tests (50 iterations each, 3 properties) + 1 example-based boundary test

### Property Tests (fast-check via Vitest)
- `bundle-utils.test.ts`: Properties 1 and 2 (discount invariants) — 200 iterations
- `StockStepper.test.tsx`: Property 3 (bounds invariant) — 50 iterations per sub-property

### E2E Tests (Playwright)
- `full-journey.spec.ts`: Complete order lifecycle with assertions at every step
- `comptoir.spec.ts`: B2B portal navigation and access control (5 tests)
- `baker-portal.spec.ts`: Pro baker dashboard navigation and access control (5 tests)

### Backend Tests (Go)
- `service_test.go`: Refund test with spy gateway verifying actual gateway calls

### CI Verification
- All tests green locally (`npx vitest run`, `go test ./...`)
- Push to branch and verify GitHub Actions e2e job executes (not skipped)
- Verify health check catches boot failures (remove server, see job fail)
