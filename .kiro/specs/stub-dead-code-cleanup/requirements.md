# Requirements Document

## Introduction

Clean up stubs, dead code, and user-facing lies across the bakery-app codebase. The goals are: no visible UI for unimplemented features, a single enforced state machine for order/reservation transitions, correct GDPR consent gating, and removal of tracked binaries and route naming inconsistencies.

## Linked Ticket

MA-68 -- Stub & dead-code cleanup: Apple SSO stub, fake S3 upload, dead state machine, consent theater

## Glossary

- **State_Machine**: The `internal/domain/statemachine.go` module containing `TransitionOrder` and `TransitionReservation`, which enforce valid status transitions.
- **Apple_SSO_Stub**: The `AppleProvider.ExchangeCode` method that returns an "not yet implemented" error while a UI button could still render.
- **Upload_Storage**: The backend abstraction (`internal/upload/Storage` interface) that selects between local filesystem and cloud-based storage at boot.
- **Cookie_Consent_Banner**: The `CookieConsent.tsx` component that captures user consent preferences before initializing analytics/error-tracking.
- **Server_Binary**: The compiled `server` Mach-O binary at the repo root, currently tracked in git.

---

## Requirements

### Requirement 1: Apple SSO Cleanup

**User Story:** As a user, I want the login page to only show authentication options that work, so that I am not misled by non-functional buttons.

#### Acceptance Criteria

1. WHILE the Apple OAuth exchange is not implemented, THE Login_Page SHALL NOT render the Apple Sign-In button regardless of environment variable values.
2. WHEN the `VITE_APPLE_OAUTH_ENABLED` environment variable is removed from all configuration, THE Frontend_Build SHALL compile without errors.
3. THE Apple_SSO_Stub code (provider struct, `GetAuthURL`, `ExchangeCode`, icon component, i18n keys) SHALL remain in the codebase but be unreachable from production UI until fully implemented.

---

### Requirement 2: Upload Storage Fail-Loud

**User Story:** As an operator, I want the server to fail at boot when an unsupported storage backend is configured, so that silent data loss never occurs.

#### Acceptance Criteria

1. WHEN `UPLOAD_STORAGE` is set to an unrecognized value, THE Server SHALL terminate at startup with a fatal log message naming the invalid value and listing supported options.
2. THE Server SHALL support exactly one storage backend: `local` (or unset, defaulting to `local`).
3. WHEN `UPLOAD_STORAGE` is set to `"s3"`, THE Server SHALL terminate at startup with a clear message that S3 is not yet implemented.

---

### Requirement 3: Unified State Machine Enforcement

**User Story:** As a developer, I want every order and reservation status change to go through one transition table, so that invalid state transitions are impossible regardless of the code path.

#### Acceptance Criteria

1. WHEN the payment service confirms an order, THE Payment_Service SHALL use `domain.TransitionOrder` to change status from `pending_payment` to `confirmed`.
2. WHEN a reservation is created, THE Reservation_Service SHALL use `domain.TransitionReservation` (or set initial status via a dedicated `NewReservation` factory) to set the initial `confirmed` status.
3. WHEN a customer cancels a reservation, THE Reservation_Service SHALL use `domain.TransitionReservation` to transition to `cancelled`.
4. THE State_Machine SHALL be the sole mechanism for changing `Order.Status` and `Reservation.Status` outside of repository scan/hydration (reading from database).
5. IF a caller attempts to assign a status directly (bypassing the state machine), THEN THE Go_Compiler or linter SHALL flag it through an unexported status field or equivalent enforcement mechanism.

---

### Requirement 4: Cookie Consent Gates Analytics

**User Story:** As a user in the EU, I want my consent choice to actually control whether tracking initializes, so that my privacy preferences are respected.

#### Acceptance Criteria

1. WHEN the user has not granted "all" consent, THE Frontend SHALL NOT initialize Sentry or any future analytics library.
2. WHEN the user grants "all" consent, THE Frontend SHALL initialize Sentry immediately.
3. WHEN the page loads with a previously stored "all" consent, THE Frontend SHALL initialize Sentry on load.
4. WHEN the page loads with a previously stored "essential" consent or no consent, THE Frontend SHALL NOT initialize Sentry.

---

### Requirement 5: Dead Weight Removal

**User Story:** As a developer, I want the repository to contain only production-relevant code and no tracked build artifacts, so that the codebase stays lean and clones remain fast.

#### Acceptance Criteria

1. THE Repository SHALL NOT track the compiled `server` binary (add to `.gitignore`, remove from git history or index).
2. THE `/dashboard/stats` route SHALL be renamed to `/dashboard/schedule` to match the component it renders (`DashboardSchedule`), and the sidebar label SHALL read "Planning" instead of "Statistiques".
3. WHEN `DashboardBundlesPage.tsx`, `BakeryListPage.tsx`, or `DashboardReservations.tsx` files exist, THE Repository SHALL remove them if they are not routed or imported.
4. WHEN a `pkg/` directory exists with no source files, THE Repository SHALL remove it.
5. WHEN the `recharts` dependency is listed in `package.json` with zero imports, THE Repository SHALL remove it.
6. THE Frontend_Build and Go_Build SHALL compile cleanly after all removals.

