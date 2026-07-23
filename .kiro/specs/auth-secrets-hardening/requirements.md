# Requirements: Auth & Secrets Hardening

## Overview

Harden authentication and secrets management across the bakery platform: enforce required secrets at boot, strengthen OAuth CSRF protection and account linking, bound rate-limiter memory, eliminate weak randomness in token paths, and remove unsafe identity fallbacks.

## Linked Ticket

MA-66 — Auth & secrets hardening: JWT/webhook secret defaults, OAuth CSRF + account linking, rate limiting

---

## Requirement 1: Production Secret Enforcement

**User Story:** As a platform operator, I want the server to refuse to start in production when critical secrets are missing, so that the application never runs with predictable credentials.

### Acceptance Criteria

1.1 WHEN APP_ENV is "production" AND JWT_SECRET is empty or unset, THE Server SHALL exit immediately with a non-zero exit code and a clear error message naming the missing variable.

1.2 WHEN APP_ENV is "production" AND PAYMENT_GATEWAY is "stripe" AND STRIPE_WEBHOOK_SECRET is empty or unset, THE Server SHALL exit immediately with a non-zero exit code and a clear error message naming the missing variable.

1.3 WHEN APP_ENV is not "production" AND JWT_SECRET is empty, THE Server SHALL use a development-only default and log a warning indicating the secret is insecure.

1.4 THE Server SHALL NOT log the actual value of any secret during startup.

---

## Requirement 2: OAuth State Verification (CSRF Protection)

**User Story:** As a user, I want OAuth login flows to be protected against cross-site request forgery, so that attackers cannot trick me into linking their provider account.

### Acceptance Criteria

2.1 WHEN a client requests an OAuth authorization URL, THE OAuthHandler SHALL generate a server-signed state token containing a cryptographic nonce and an expiration timestamp.

2.2 WHEN the OAuth callback is received, THE OAuthHandler SHALL verify the state parameter signature and reject the request with HTTP 403 if the signature is invalid.

2.3 WHEN the OAuth callback state token has expired (older than 10 minutes), THE OAuthHandler SHALL reject the request with HTTP 403.

2.4 WHEN the OAuth callback state parameter is missing or empty, THE OAuthHandler SHALL reject the request with HTTP 403.

---

## Requirement 3: OAuth Account Linking Safety

**User Story:** As a user with an existing password-protected account, I want OAuth linking to require verification, so that an attacker with a matching email on a provider cannot take over my account.

### Acceptance Criteria

3.1 WHEN an OAuth callback returns an email matching an existing account that has a password, THE OAuthService SHALL reject auto-linking and return an error indicating password verification is required.

3.2 WHEN an OAuth callback returns an email matching an existing social-only account (no password), THE OAuthService SHALL auto-link the new provider to that account.

3.3 WHEN an OAuth callback returns an email that matches no existing account, THE OAuthService SHALL create a new user account and link the provider.

---

## Requirement 4: Rate Limiting Hardening

**User Story:** As a platform operator, I want rate limiting to use reliable client identification and bounded memory, so that brute-force attacks are mitigated and the server does not leak memory.

### Acceptance Criteria

4.1 THE AuthRateLimiter SHALL key rate-limit buckets on the TCP-level remote address (r.RemoteAddr) rather than any client-supplied header.

4.2 WHEN the register endpoint receives requests, THE Server SHALL enforce rate limiting with the same configuration as the login endpoint.

4.3 THE RateLimiter SHALL periodically evict stale entries that have no timestamps within the current window, bounding memory growth to active clients only.

4.4 WHEN a rate-limited request is rejected, THE RateLimiter SHALL return HTTP 429 with a JSON error body containing code "RATE_LIMITED".

---

## Requirement 5: Cryptographic Token Generation

**User Story:** As a platform operator, I want all security-sensitive tokens to use cryptographically secure randomness, so that tokens cannot be predicted by attackers.

### Acceptance Criteria

5.1 THE AuthService SHALL generate registration tokens using crypto/rand exclusively.

5.2 THE Server SHALL NOT import or use math/rand in any code path that generates tokens, secrets, or credentials.

5.3 THE registration token code SHALL contain at least 40 bits of entropy (8 alphanumeric characters from a 32-character alphabet or equivalent).

---

## Requirement 6: Identity Extraction Safety

**User Story:** As a developer, I want the user identity extraction to rely solely on the JWT-authenticated context, so that no header-based bypass is possible even if routes are misconfigured.

### Acceptance Criteria

6.1 THE extractUserID function SHALL read the user ID exclusively from the JWT-authenticated request context.

6.2 THE extractUserID function SHALL NOT read from the X-User-ID header or any other client-supplied header.

6.3 WHEN no user ID is present in the context, THE extractUserID function SHALL return an empty string (never fall back to "anonymous" or a header value).
