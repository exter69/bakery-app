# Implementation Plan: Auth & Secrets Hardening

## Overview

This plan verifies and completes the security hardening for MA-66. Much of the work is already implemented; the tasks focus on verifying existing behavior, adding missing tests, and closing any remaining gaps identified during verification.

## Tasks

- [ ] 1. Verify and harden production secret checks
  - [ ] 1.1 Audit `cmd/server/main.go` secret validation logic
    - Confirm `log.Fatal` fires for empty JWT_SECRET when APP_ENV=production
    - Confirm `log.Fatal` fires for empty STRIPE_WEBHOOK_SECRET when APP_ENV=production and PAYMENT_GATEWAY=stripe
    - Confirm no secret values are logged (only warnings about missing vars)
    - Extract secret validation into a testable `validateSecrets()` function if not already testable
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [ ]* 1.2 Write unit tests for secret validation
    - Test production + missing JWT_SECRET returns error
    - Test production + stripe + missing webhook secret returns error
    - Test non-production + missing JWT_SECRET returns default with warning
    - Test that secret values never appear in error messages
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [ ] 2. Verify OAuth state token implementation
  - [ ] 2.1 Audit `internal/api/oauth_handler.go` state token logic
    - Confirm `generateOAuthState` uses crypto/rand for nonce (16 bytes)
    - Confirm HMAC-SHA256 signing with `stateKey`
    - Confirm `verifyOAuthState` checks signature, checks TTL (10 min)
    - Confirm callback returns HTTP 403 for missing/invalid/expired state
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [ ]* 2.2 Write property test for OAuth state structural integrity
    - **Property 1: OAuth state token structural integrity**
    - Generate 100+ state tokens, verify 3-part structure, valid base64 nonce, valid hex sig, valid RFC3339 timestamp
    - **Validates: Requirements 2.1**

  - [ ]* 2.3 Write property test for OAuth state tamper detection
    - **Property 2: OAuth state token tamper detection**
    - Generate valid tokens, tamper with each component, verify all rejected
    - **Validates: Requirements 2.2**

  - [ ]* 2.4 Write property test for OAuth state expiry
    - **Property 3: OAuth state token expiry enforcement**
    - Generate tokens at time T, verify acceptance at T+delta for delta <= 10min and rejection for delta > 10min
    - **Validates: Requirements 2.3**

- [ ] 3. Verify OAuth account linking safety
  - [ ] 3.1 Audit `internal/service/oauth_service.go` linking logic
    - Confirm password-protected accounts reject auto-linking (returns ErrOAuthAccountLinkRequiresVerification)
    - Confirm social-only accounts (empty PasswordHash) allow auto-linking
    - Confirm new emails create new accounts
    - _Requirements: 3.1, 3.2, 3.3_

  - [ ]* 3.2 Write property test for account linking decision
    - **Property 4: OAuth account linking decision**
    - Generate random users (with/without password) and OAuth emails, verify linking outcome matches rules
    - **Validates: Requirements 3.1, 3.2, 3.3**

- [ ] 4. Verify and test rate limiter hardening
  - [ ] 4.1 Audit `internal/middleware/ratelimit.go` and `cmd/server/main.go` rate limiter config
    - Confirm auth rate limiter keys on `r.RemoteAddr`
    - Confirm `/api/auth/register` uses `authRateLimiter` middleware
    - Confirm `evictStale` logic removes entries after 5x window with no activity
    - Confirm HTTP 429 response with code "RATE_LIMITED"
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ]* 4.2 Write property test for rate limiter keying
    - **Property 5: Rate limiter keys on RemoteAddr**
    - Generate requests with random RemoteAddr + X-Forwarded-For combinations, verify independent bucketing by RemoteAddr
    - **Validates: Requirements 4.1**

  - [ ]* 4.3 Write property test for stale entry eviction
    - **Property 6: Rate limiter stale entry eviction**
    - Generate entries, advance time past eviction interval, verify all stale entries removed
    - **Validates: Requirements 4.3**

- [ ] 5. Verify cryptographic token generation
  - [ ] 5.1 Audit `internal/service/auth_service.go` token generation
    - Confirm `generateTokenCode` uses `crypto/rand.Int` (not math/rand)
    - Confirm 8-character output from 32-char alphabet (40+ bits entropy)
    - Run `grep -r "math/rand" internal/` to verify no math/rand in security paths
    - _Requirements: 5.1, 5.2, 5.3_

  - [ ]* 5.2 Write property test for token entropy
    - **Property 7: Registration token entropy**
    - Generate 100+ tokens, verify each is 8 chars from alphabet "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
    - **Validates: Requirements 5.1, 5.3**

- [ ] 6. Verify identity extraction safety
  - [ ] 6.1 Audit `internal/api/helpers.go` and `internal/middleware/auth.go`
    - Confirm `extractUserID` calls only `GetUserIDFromContext`
    - Confirm no X-User-ID header reading anywhere in the function
    - Confirm returns empty string when no context value set
    - _Requirements: 6.1, 6.2, 6.3_

  - [ ]* 6.2 Write property test for identity extraction
    - **Property 8: Identity extraction ignores headers**
    - Generate requests with random X-User-ID headers but no JWT context, verify empty string returned
    - **Validates: Requirements 6.1, 6.2, 6.3**

- [ ] 7. Static analysis and CI verification
  - [ ] 7.1 Run static analysis checks
    - Run `go vet ./...` and `staticcheck ./...`
    - Run `grep -rn "math/rand" internal/ cmd/` to verify no math/rand in auth/token paths
    - Verify all existing tests pass with `go test ./...`
    - _Requirements: 5.2_

- [ ] 8. Final checkpoint
  - Ensure all tests pass, ask the user if questions arise.
  - Verify no regressions in existing auth flows (login, register, OAuth callback)
  - Confirm the codebase matches all 6 requirements

## Notes

- Many fixes from MA-66 are already implemented in the current codebase. The primary value of this spec is verification and test coverage.
- Tasks marked with `*` are optional property-based tests that provide strong correctness guarantees.
- The Go stdlib `testing/quick` package or `github.com/leanovate/gopter` can be used for property tests.
- No database migrations are needed for this hardening ticket.
- No frontend changes are needed.

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "2.1", "3.1", "4.1", "5.1", "6.1"] },
    { "id": 1, "tasks": ["1.2", "2.2", "2.3", "2.4", "3.2", "4.2", "4.3", "5.2", "6.2"] },
    { "id": 2, "tasks": ["7.1"] },
    { "id": 3, "tasks": ["8"] }
  ]
}
```
