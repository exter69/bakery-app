# Implementation Plan: GDPR Completeness — Deletion, Export, and Token Invalidation

## Overview

Extend `DeleteAccount` to fully erase PII from all auxiliary tables and external services, add JWT invalidation for deleted users in the auth middleware, and include reviews in the data export. All changes are in Go (backend).

## Tasks

- [ ] 1. Add missing repository methods
  - [x] 1.1 Add `DeleteByUser(ctx, userID) error` to `SocialLoginRepository` interface in `internal/domain/social_login.go`
    - Implement in `internal/repository/postgres/social_login_repo.go`: `DELETE FROM social_logins WHERE user_id = $1`
    - Implement in `internal/repository/memory/social_login_repo.go` for tests
    - _Requirements: 1.1_

  - [x] 1.2 Add `DeleteProfile(ctx, userID) error` and `DeleteSavedListsByUser(ctx, userID) error` to `B2BRepository` interface in `internal/domain/b2b.go`
    - Implement in `internal/repository/postgres/b2b_repo.go`: `DELETE FROM business_profiles WHERE user_id = $1` and `DELETE FROM saved_lists WHERE user_id = $1`
    - Implement in-memory variants for tests
    - _Requirements: 2.1, 2.2_

  - [x] 1.3 Add `ListByUser(ctx, userID) ([]Review, error)` to `ReviewRepository` interface in `internal/domain/repository.go`
    - Implement in `internal/repository/postgres/review_repo.go`: `SELECT * FROM reviews WHERE user_id = $1 ORDER BY created_at DESC`
    - Implement in-memory variant for tests
    - _Requirements: 6.1_

  - [x] 1.4 Add `DeleteByUser(userID string)` method to `push.Store` in `internal/push/store.go`
    - Lock mutex, delete the user's key from the `subs` map
    - _Requirements: 3.1_

- [x] 2. Add Stripe Customer deletion method
  - [x] 2.1 Add `DeleteCustomer(ctx, customerID string) error` to `StripeCustomerService` in `internal/payment/stripe_customer.go`
    - If customerID is empty, return nil immediately
    - Call `customer.Del(customerID, nil)` from the Stripe SDK
    - Return wrapped error on failure
    - _Requirements: 4.1_

- [x] 3. Extend DeleteAccount with full cleanup
  - [x] 3.1 Update `UserService` struct to accept `PushStore *push.Store` and `StripeCustomerSvc *payment.StripeCustomerService` dependencies
    - Add fields to `UserServiceConfig` and `UserService`
    - Update `NewUserServiceFull` constructor
    - Update wiring in `cmd/server/main.go`
    - _Requirements: 3.1, 4.1_

  - [x] 3.2 Extend `DeleteAccount` in `internal/service/user_service.go` with new cleanup steps
    - After existing anonymization: call `socialLoginRepo.DeleteByUser`
    - Call `b2bRepo.DeleteProfile` (if b2bRepo != nil)
    - Call `b2bRepo.DeleteSavedListsByUser` (if b2bRepo != nil)
    - Call `pushStore.DeleteByUser` (if pushStore != nil)
    - Call `stripeCustomerSvc.DeleteCustomer` with the user's StripeCustomerID before clearing it (best-effort: log error, don't fail)
    - _Requirements: 1.1, 2.1, 2.2, 3.1, 4.1, 4.2_

  - [ ]* 3.3 Write property test for complete PII erasure (Property 1)
    - **Property 1: Complete PII erasure across all tables**
    - Use rapid to generate random user state with varying social logins, B2B profile, saved lists, push subs
    - Invoke DeleteAccount, verify zero residual records in all stores
    - **Validates: Requirements 1.1, 1.2, 2.1, 2.2, 2.3, 3.1, 3.2**

  - [ ]* 3.4 Write property test for Stripe deletion (Property 2)
    - **Property 2: Stripe Customer deletion is attempted for all users with a customer ID**
    - Use rapid to generate users with/without stripe_customer_id
    - Mock Stripe client, verify delete called iff ID non-empty
    - **Validates: Requirements 4.1, 4.3**

- [x] 4. JWT invalidation for deleted users
  - [x] 4.1 Modify `JWTAuth` middleware in `internal/middleware/auth.go` to accept a `UserRepository` parameter
    - After successful token validation, call `userRepo.GetByID(ctx, userID)`
    - If user is found and username starts with `deleted-` and contact_email is empty, return 401 with code `ACCOUNT_DELETED`
    - If user is nil (not found in DB), allow through (avoid blocking on transient failures)
    - _Requirements: 5.1, 5.2, 5.3_

  - [x] 4.2 Update all `JWTAuth` call sites in `cmd/server/main.go` to pass the `userRepo`
    - _Requirements: 5.1_

  - [ ]* 4.3 Write property test for deleted-user JWT rejection (Property 3)
    - **Property 3: Deleted-user JWT rejection**
    - Use rapid to generate deleted usernames, create JWTs, verify 401 ACCOUNT_DELETED
    - **Validates: Requirements 5.1, 5.2**

  - [ ]* 4.4 Write property test for active-user JWT passthrough (Property 4)
    - **Property 4: Active-user JWT passthrough**
    - Use rapid to generate active users with valid JWTs, verify request proceeds
    - **Validates: Requirements 5.3**

- [x] 5. Checkpoint
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Include reviews in data export
  - [x] 6.1 Update `ExportData` in `internal/service/user_service.go` to call `ReviewRepository.ListByUser`
    - Replace the existing comment/placeholder with actual review fetching
    - Map each `domain.Review` to `dto.DataExportReview` with ID, BakeryID, Rating, Text, CreatedAt
    - Handle nil reviewRepo gracefully (empty array)
    - _Requirements: 6.1, 6.2, 6.3_

  - [ ]* 6.2 Write property test for export review completeness (Property 5)
    - **Property 5: Export includes all user reviews**
    - Use rapid to generate users with 0-10 reviews, verify export array length and field completeness
    - **Validates: Requirements 6.1, 6.2, 6.3**

- [x] 7. Update DATA-INVENTORY.md
  - [x] 7.1 Update `docs/DATA-INVENTORY.md` data retention table
    - Change "Social logins" retention to: "Until account deletion | Hard delete"
    - Change "B2B profiles" retention to: "Until account deletion | Hard delete (profile + saved lists)"
    - Add "Push subscriptions" row: "Until account deletion | Cleared from memory"
    - Add "Stripe Customer" row: "Until account deletion | Deleted via Stripe API"
    - Update "Deleted accounts" row to mention JWT invalidation via middleware check
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 8. Final checkpoint
  - Ensure all tests pass, ask the user if questions arise.
  - Verify `go build ./...` and `go vet ./...` succeed
  - Run property tests with `go test -run Property -count=1 ./internal/service/... ./internal/middleware/...`

## Notes

- The push store is in-memory; no migration needed for push subscription cleanup.
- Stripe Customer deletion is best-effort to avoid blocking the user's erasure right on external failures.
- The JWT middleware check queries the DB on every request. At current scale this is acceptable; a short TTL cache can be added later if needed.
- `saved_list_items` are cleaned up automatically via `ON DELETE CASCADE` when `saved_lists` rows are removed.
- The `rapid` library (`pgregory.net/rapid`) should be added as a test dependency: `go get pgregory.net/rapid`
- Tasks marked with `*` are optional property-based test tasks that can be skipped for faster delivery.

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "1.3", "1.4"] },
    { "id": 1, "tasks": ["2.1"] },
    { "id": 2, "tasks": ["3.1", "3.2"] },
    { "id": 3, "tasks": ["3.3", "3.4", "4.1", "4.2"] },
    { "id": 4, "tasks": ["4.3", "4.4", "5"] },
    { "id": 5, "tasks": ["6.1"] },
    { "id": 6, "tasks": ["6.2", "7.1"] },
    { "id": 7, "tasks": ["8"] }
  ]
}
```
