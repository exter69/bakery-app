# Requirements: Frontend Session & Dashboard Quality

## Introduction

This spec addresses multiple quality issues in the frontend session management and dashboard internationalisation. The JWT decode function in `client.ts` uses standard `atob()` on base64url-encoded payloads without character translation, causing intermittent authentication failures. Token expiry is not checked client-side, TypeScript strict mode is disabled, role identifiers use magic numbers, and seven dashboard pages bypass the i18n system with hardcoded French/English strings.

## Linked Ticket

MA-73 — Frontend session & dashboard quality: JWT base64url bug, no expiry check, strict mode off, dashboard i18n bypass

## Glossary

- **JWT_Decoder**: The utility function responsible for parsing JWT token payloads on the client side.
- **Auth_Module**: The collection of functions in `frontend/src/api/client.ts` managing token storage, retrieval, and validation.
- **Role_Enum**: A shared TypeScript enum defining user roles (Admin, Seller, Customer, B2B) by name instead of magic numbers.
- **RoleRoute**: The route guard component that checks authentication and role before rendering protected routes.
- **Dashboard_Pages**: The set of seller portal pages under `frontend/src/pages/dashboard/`.
- **I18n_System**: The existing internationalisation system (`useI18n` hook, translation dictionaries for EN/FR/NL).

---

## Requirement 1: Fix JWT Base64url Decoding

**User Story:** As a baker or B2B user, I want my JWT token to always decode correctly, so that I am never incorrectly redirected away from my portal due to a decoding failure.

### Acceptance Criteria

1.1 WHEN the JWT payload contains base64url characters (`-` or `_`), THE JWT_Decoder SHALL translate them to standard base64 (`+` and `/`) and add padding before calling `atob()`.

1.2 WHEN the JWT payload requires padding to reach a multiple of 4 characters, THE JWT_Decoder SHALL append the correct number of `=` characters.

1.3 THE JWT_Decoder SHALL reuse the existing `urlBase64ToUint8Array` algorithm from `usePushNotifications.ts` (or an equivalent shared utility) for the base64url-to-standard conversion.

1.4 IF the token has fewer than 3 dot-separated parts or the payload is not valid JSON, THEN THE JWT_Decoder SHALL return `null` without throwing.

---

## Requirement 2: Client-Side Token Expiry Check

**User Story:** As any authenticated user, I want expired tokens to be treated as logged-out immediately, so that I am not surprised by a sudden 401 redirect mid-session.

### Acceptance Criteria

2.1 WHEN `isAuthenticated()` is called, THE Auth_Module SHALL decode the token's `exp` claim and return `false` if the current time (in seconds) is greater than or equal to `exp`.

2.2 WHEN the `exp` claim is missing or not a number, THE Auth_Module SHALL treat the token as valid (fall back to existing behaviour).

2.3 THE Auth_Module SHALL retain the existing `auth:unauthorized` event dispatch on 401 responses as a secondary fallback for server-side expiry enforcement.

---

## Requirement 3: Enable TypeScript Strict Mode

**User Story:** As a developer, I want TypeScript strict mode enabled, so that the type checker catches null-safety and implicit-any bugs before runtime.

### Acceptance Criteria

3.1 THE Build_System SHALL compile cleanly with `"strict": true` in `tsconfig.app.json`, or at minimum with `"strictNullChecks": true` and `"noImplicitAny": true` as a staged first step.

3.2 IF enabling full `"strict": true` produces more than 50 type errors, THEN the team SHALL enable `"strictNullChecks": true` and `"noImplicitAny": true` first and document remaining strict flags for a follow-up ticket.

3.3 THE Build_System SHALL not introduce new `// @ts-ignore` or `as any` casts to satisfy the stricter checks, except where explicitly justified with a comment referencing a ticket.

---

## Requirement 4: Shared Role Enum and Reactive Auth State

**User Story:** As a developer, I want role identifiers defined in a single shared enum, so that route guards and components reference roles by name instead of magic numbers.

### Acceptance Criteria

4.1 THE Role_Enum SHALL define named constants for all user roles: `Admin = 0`, `Seller = 1`, `Customer = 2`, `B2B = 3`.

4.2 WHEN route guards reference roles, THE RoleRoute component SHALL use Role_Enum members instead of numeric literals.

4.3 THE Auth_Module SHALL expose a reactive auth state (e.g., via React context or a shared hook) so that components do not re-read localStorage on every render to determine role.

---

## Requirement 5: Dashboard Pages Use i18n System

**User Story:** As a baker using the seller portal, I want all dashboard text displayed in my chosen locale, so that I have a consistent language experience.

### Acceptance Criteria

5.1 WHEN rendering user-visible text, THE Dashboard_Pages SHALL call `t()` with the appropriate translation key instead of using hardcoded strings.

5.2 THE I18n_System SHALL contain translation keys for all user-facing strings currently hardcoded in: DashboardOverview, DashboardSchedule, DashboardPayouts, DashboardBakery, DashboardProducts, DashboardOrders, DashboardBundles.

5.3 THE I18n_System SHALL provide translations in all three supported locales (EN, FR, NL) for every new key added.

5.4 WHEN a grep is run for hardcoded French or English strings in Dashboard_Pages source files, THE result SHALL be empty (excluding i18n key literals and non-user-facing code comments).

---

## Requirement 6: DashboardOverview Stats Date Filtering

**User Story:** As a baker, I want my "today" statistics to reflect only orders for today, so that the numbers are accurate.

### Acceptance Criteria

6.1 WHEN displaying "today" order stats, THE DashboardOverview SHALL filter orders by today's date in addition to `status=confirmed`, or rename the label to accurately describe what is shown (e.g., "Active orders").

