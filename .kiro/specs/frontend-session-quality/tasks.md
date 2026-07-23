# Implementation Plan: Frontend Session & Dashboard Quality

## Overview

Fix JWT base64url decoding, add client-side token expiry checks, introduce a shared Role enum with reactive auth context, enable TypeScript strict mode, and migrate all dashboard pages to use the i18n system. All changes are frontend-only (TypeScript/React).

## Tasks

- [ ] 1. Fix JWT base64url decode and add expiry check
  - [ ] 1.1 Refactor `decodeTokenRole` in `frontend/src/api/client.ts` to use a new `decodeBase64Url` helper that translates `-`/`_` to `+`/`/` and pads with `=` before calling `atob()`; extract a `decodeTokenPayload` function returning the full payload object or null
    - Reuse the algorithm pattern from `usePushNotifications.ts:urlBase64ToUint8Array`
    - Keep `decodeTokenRole` as a thin wrapper calling `decodeTokenPayload`
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - [ ] 1.2 Update `isAuthenticated()` to decode `exp` from the token and return `false` (plus `clearToken()`) when expired; keep returning `true` when `exp` is missing or not a number
    - _Requirements: 2.1, 2.2, 2.3_
  - [ ]* 1.3 Write property tests for JWT decode round-trip (fast-check)
    - **Property 1: JWT Base64url Round-Trip Decode**
    - Generate random JSON objects, encode as base64url, verify `decodeBase64Url` + `JSON.parse` reproduces the original
    - **Validates: Requirements 1.1, 1.2**
  - [ ]* 1.4 Write property tests for malformed token handling
    - **Property 2: Malformed Tokens Return Null**
    - Generate strings without valid JWT structure, verify `decodeTokenPayload` returns null
    - **Validates: Requirements 1.4**
  - [ ]* 1.5 Write property tests for expiry logic
    - **Property 3: Expiry Determines Authentication**
    - Generate tokens with random `exp` values, mock `Date.now`, verify `isAuthenticated` correctness
    - **Validates: Requirements 2.1, 2.2**

- [ ] 2. Introduce Role enum and reactive auth context
  - [ ] 2.1 Create `frontend/src/auth/roles.ts` with `UserRole` enum (`Admin=0, Seller=1, Customer=2, B2B=3`)
    - _Requirements: 4.1_
  - [ ] 2.2 Create `frontend/src/auth/AuthProvider.tsx` and `useAuth.ts` hook exposing reactive auth state (`isLoggedIn`, `role`, `login`, `logout`) via React context
    - On mount, read token from localStorage and decode role
    - Listen for `storage` events and `auth:unauthorized` custom event
    - _Requirements: 4.3_
  - [ ] 2.3 Wrap the app in `AuthProvider` in `App.tsx`; update `RoleRoute.tsx` to use `useAuth()` instead of direct localStorage reads; replace magic numbers `[0, 1]` and `[3]` with `[UserRole.Admin, UserRole.Seller]` and `[UserRole.B2B]`
    - _Requirements: 4.2, 4.3_
  - [ ]* 2.4 Write unit tests for RoleRoute with auth context (render with mock provider, verify redirect behaviour)
    - _Requirements: 4.2, 4.3_

- [ ] 3. Checkpoint - Verify auth changes
  - Ensure all tests pass (`vitest --run`), ask the user if questions arise.

- [ ] 4. Enable TypeScript strict mode
  - [ ] 4.1 Add `"strict": true` to `frontend/tsconfig.app.json` `compilerOptions`; run `tsc --noEmit` and fix all resulting type errors
    - If >50 errors, fall back to `"strictNullChecks": true` + `"noImplicitAny": true` and document remaining flags
    - No new `@ts-ignore` or `as any` without a ticket reference in a comment
    - _Requirements: 3.1, 3.2, 3.3_

- [ ] 5. Checkpoint - Strict mode compiles clean
  - Run `tsc --noEmit` and ensure exit code 0, ask the user if questions arise.

- [ ] 6. Dashboard i18n migration
  - [ ] 6.1 Add all missing dashboard translation keys to `frontend/src/i18n/translations.ts` for EN, FR, and NL (schedule page, payouts page, bakery page, products page, orders page, bundles page, and remaining hardcoded strings in overview page)
    - _Requirements: 5.2, 5.3_
  - [ ] 6.2 Refactor `DashboardSchedule.tsx` to import `useI18n` and replace all hardcoded English strings with `t()` calls
    - _Requirements: 5.1_
  - [ ] 6.3 Refactor `DashboardPayouts.tsx` to import `useI18n` and replace all hardcoded English strings with `t()` calls
    - _Requirements: 5.1_
  - [ ] 6.4 Refactor `DashboardBakery.tsx` to import `useI18n` and replace all hardcoded strings with `t()` calls
    - _Requirements: 5.1_
  - [ ] 6.5 Refactor `DashboardProducts.tsx` to import `useI18n` and replace all hardcoded strings with `t()` calls
    - _Requirements: 5.1_
  - [ ] 6.6 Refactor `DashboardOrders.tsx` to import `useI18n` and replace all hardcoded strings with `t()` calls
    - _Requirements: 5.1_
  - [ ] 6.7 Refactor `DashboardBundles.tsx` to import `useI18n` and replace all hardcoded strings with `t()` calls
    - _Requirements: 5.1_
  - [ ] 6.8 Replace remaining hardcoded French strings in `DashboardOverview.tsx` (greeting, stat labels, section titles, empty states) with `t()` calls
    - _Requirements: 5.1, 5.4_
  - [ ]* 6.9 Write a translation-key coverage test verifying all `t()` keys used in dashboard pages exist in all three locale dictionaries
    - **Property 4: Translation Key Coverage**
    - **Validates: Requirements 5.2, 5.3**

- [ ] 7. Fix DashboardOverview "today" stats
  - [ ] 7.1 Update `DashboardOverview.tsx` to pass a `date` filter (today's date in ISO format) to `fetchBakeryOrders` and `fetchBakeryReservations`, or rename the "Commandes du jour" label to accurately describe unfiltered confirmed orders
    - _Requirements: 6.1_

- [ ] 8. Final checkpoint - Full verification
  - Run `tsc --noEmit` (strict), `vitest --run`, and grep for hardcoded dashboard strings. Ensure all pass, ask the user if questions arise.

## Notes

- The `urlBase64ToUint8Array` in `usePushNotifications.ts` already implements the correct base64url-to-standard algorithm; the new `decodeBase64Url` uses the same pattern but returns a string (for `JSON.parse`) rather than a `Uint8Array`.
- The `auth:unauthorized` event listener in `AuthRedirectListener` stays in place as a server-side fallback.
- Dashboard pages that already use `useI18n` (DashboardOverview partially, DashboardLayout, DashboardB2BPage) need only the hardcoded string cleanup — they don't need the import added.
- The i18n dictionary is already at 372-key parity across locales; only dashboard-specific keys are missing.
- Property tests use `fast-check` which is already available or trivially installable in the Vitest setup.

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "2.1"] },
    { "id": 1, "tasks": ["1.3", "1.4", "1.5", "2.2"] },
    { "id": 2, "tasks": ["2.3", "2.4"] },
    { "id": 3, "tasks": ["3"] },
    { "id": 4, "tasks": ["4.1"] },
    { "id": 5, "tasks": ["5"] },
    { "id": 6, "tasks": ["6.1"] },
    { "id": 7, "tasks": ["6.2", "6.3", "6.4", "6.5", "6.6", "6.7", "6.8"] },
    { "id": 8, "tasks": ["6.9", "7.1"] },
    { "id": 9, "tasks": ["8"] }
  ]
}
```
