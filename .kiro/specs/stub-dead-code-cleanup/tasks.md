# Implementation Plan: Stub & Dead-Code Cleanup (MA-68)

## Overview

Systematic cleanup of stubs, bypassed state machine calls, tracked binaries, and route naming inconsistencies. Several items from the original ticket have already been resolved (S3 fake upload, cookie consent, recharts dep, dead page files, pkg/ dir). This plan addresses the remaining issues.

## Tasks

- [x] 1. Remove Apple SSO button from login page
  - [x] 1.1 Remove the `VITE_APPLE_OAUTH_ENABLED` conditional block and Apple button JSX from `frontend/src/pages/LoginPage.tsx`
    - Remove the `import AppleIcon` statement
    - Remove the entire `{import.meta.env.VITE_APPLE_OAUTH_ENABLED === 'true' && (...)}` block
    - Keep `AppleProvider` in backend and `AppleIcon` component on disk for future implementation
    - _Requirements: 1.1, 1.2, 1.3_
  - [x] 1.2 Remove `login.signInWithApple` i18n key usage (leave key in translation files for future use)
    - Verify no remaining references to `VITE_APPLE_OAUTH_ENABLED` in frontend source
    - _Requirements: 1.2_

- [x] 2. Add explicit S3 case to upload storage config
  - [x] 2.1 In `cmd/server/main.go`, add `case "s3":` before the default that fatals with message referencing `upload.ErrS3NotImplemented`
    - Ensures someone setting `UPLOAD_STORAGE=s3` gets a targeted message rather than the generic "unknown value" error
    - _Requirements: 2.1, 2.2, 2.3_
  - [ ]* 2.2 Write unit test for storage backend selection logic
    - Test that `""` and `"local"` succeed, `"s3"` fatals with S3-specific message, unknown values fatal with generic message
    - _Requirements: 2.1, 2.2, 2.3_

- [x] 3. Enforce state machine in payment service
  - [x] 3.1 In `internal/payment/service.go`, replace `order.Status = domain.OrderStatusConfirmed` with `domain.TransitionOrder(order, domain.OrderStatusConfirmed)` and handle the error
    - The preceding guard `if order.Status != domain.OrderStatusPendingPayment` already validates the "from" state, but using TransitionOrder makes the contract explicit
    - Return error if transition fails (should never happen given the guard, but defense in depth)
    - _Requirements: 3.1, 3.4_
  - [ ]* 3.2 Update payment service tests to verify TransitionOrder is used
    - Test that confirming an order in `pending_payment` state succeeds
    - Test that confirming an order in any other state fails
    - _Requirements: 3.1_

- [x] 4. Enforce state machine in reservation service
  - [x] 4.1 In `internal/service/reservation_service.go` `DeleteReservation`, replace `reservation.Status = domain.ReservationStatusCancelled` with `domain.TransitionReservation(reservation, domain.ReservationStatusCancelled)` and handle the error
    - Remove the manual terminal-state check (`if reservation.Status == PickedUp || Cancelled`) since TransitionReservation handles this
    - _Requirements: 3.3, 3.4_
  - [x] 4.2 In `internal/service/reservation_service.go` creation path, add a comment documenting that direct assignment of initial `ReservationStatusConfirmed` is the sole permitted direct-status-set (no "from" state exists for a new entity)
    - This is the only case where direct assignment is acceptable -- document the exception clearly
    - _Requirements: 3.2, 3.4_
  - [ ]* 4.3 Update reservation service tests to verify state machine enforcement
    - Test that cancelling a reservation in `confirmed` or `ready` state succeeds via TransitionReservation
    - Test that cancelling a reservation in `picked_up` or `cancelled` state fails
    - _Requirements: 3.3, 3.4_

- [x] 5. Checkpoint -- Verify state machine enforcement
  - Ensure all tests pass, ask the user if questions arise.
  - Run `go build ./...` and `go test ./...` to confirm no regressions
  - Grep for direct `.Status =` assignments in service/payment packages; only the documented creation exception should remain

- [x] 6. Remove tracked server binary and fix route naming
  - [x] 6.1 Add `server` (root-level binary) to `.gitignore`
    - Add line `/server` to the existing `.gitignore` file
    - _Requirements: 5.1_
  - [x] 6.2 Remove the binary from git index with `git rm --cached server`
    - This untracks the file without deleting it from disk
    - _Requirements: 5.1_
  - [x] 6.3 Rename `/dashboard/stats` route to `/dashboard/schedule` in `frontend/src/App.tsx`
    - Change `<Route path="stats"` to `<Route path="schedule"`
    - _Requirements: 5.2_
  - [x] 6.4 Update sidebar link in `frontend/src/pages/dashboard/DashboardLayout.tsx`
    - Change `to: '/dashboard/stats'` to `to: '/dashboard/schedule'`
    - Change label from `'Statistiques'` to `'Planning'`
    - _Requirements: 5.2_

- [x] 7. Verify dead files are already removed
  - [x] 7.1 Confirm no `BakeryListPage.tsx`, `DashboardReservations.tsx`, duplicate `DashboardBundlesPage.tsx`, `pkg/` dir, or `recharts` dep exist
    - Run verification commands: `find` for files, `grep` package.json for recharts
    - If any exist, remove them
    - _Requirements: 5.3, 5.4, 5.5_

- [x] 8. Final checkpoint -- Full build verification
  - Ensure all tests pass, ask the user if questions arise.
  - Run `go build ./...` -- must succeed
  - Run `cd frontend && npm run build` -- must succeed
  - Run `git ls-files server` -- must return empty
  - Run `go test ./...` -- all pass
  - Run `cd frontend && npm run test` -- all pass
  - _Requirements: 5.6_

## Notes

- Several items from the original MA-68 ticket are already resolved in the current codebase:
  - S3 fake upload: replaced with `ErrS3NotImplemented` sentinel and boot-time fatal for unknown values
  - Cookie consent: correctly gates Sentry via `initSentry()` checking `getConsentValue()`
  - `recharts` dependency: not present in `package.json`
  - Dead page files (`BakeryListPage.tsx`, `DashboardReservations.tsx`, `DashboardBundlesPage.tsx`): do not exist
  - `pkg/` directory: does not exist
- The Apple button is currently behind `VITE_APPLE_OAUTH_ENABLED` env var but we remove it entirely for safety
- Tasks marked with `*` are optional test sub-tasks
- State machine property tests already exist in `internal/domain/statemachine_property_test.go`
- The `AppleProvider` struct and `AppleIcon` component remain on disk for future Apple SSO implementation

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "2.1", "6.1", "6.2", "7.1"], "description": "Independent removals and config changes" },
    { "id": 1, "tasks": ["2.2", "3.1", "4.1", "4.2", "6.3", "6.4"], "description": "Code changes requiring wave 0 context" },
    { "id": 2, "tasks": ["3.2", "4.3"], "description": "Tests for state machine enforcement" },
    { "id": 3, "tasks": ["5"], "description": "Mid-point checkpoint" },
    { "id": 4, "tasks": ["8"], "description": "Final build verification" }
  ]
}
```

