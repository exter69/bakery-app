# Implementation Plan: Refund-to-Payout Reversal and Onboarding Sync

## Overview

Wire existing `OnOrderRefunded` logic into the refund flow, verify `updateRefundStatus` implementation, and extend `account.updated` handling to persist `charges_enabled`/`payouts_enabled` on the bakery. Most of the core logic already exists — the work is adding the onboarding fields, updating the Connect webhook handler, and extending the postgres repository.

## Tasks

- [x] 1. Database migration and domain model for onboarding status
  - [x] 1.1 Create migration `db/migrations/025_add_bakery_onboarding_status.sql` adding `charges_enabled BOOLEAN NOT NULL DEFAULT FALSE` and `payouts_enabled BOOLEAN NOT NULL DEFAULT FALSE` to bakeries
    - _Requirements: 3.1, 3.2_
  - [x] 1.2 Add `ChargesEnabled bool` and `PayoutsEnabled bool` fields to the `Bakery` struct in `internal/domain/models.go`
    - _Requirements: 3.1, 3.2_

- [x] 2. Update postgres bakery repository to read/write onboarding fields
  - [x] 2.1 Update `UpdateBakery` in `internal/repository/postgres/bakery_repo.go` to include `stripe_connect_id`, `charges_enabled`, `payouts_enabled` in the UPDATE query
    - _Requirements: 3.1, 3.2, 4.3_
  - [x] 2.2 Update all bakery scan/select queries in the postgres repo to read `charges_enabled` and `payouts_enabled` columns
    - _Requirements: 3.1, 3.2_
  - [x] 2.3 Update the in-memory bakery repo (`internal/repository/memory/bakery_repo.go`) to store and return the new fields
    - _Requirements: 3.1, 3.2_

- [x] 3. Implement Connect webhook onboarding sync
  - [x] 3.1 Update `handleAccountUpdated` in `internal/payment/connect_webhook.go` to set `bakery.ChargesEnabled` and `bakery.PayoutsEnabled` from the Stripe account data, and add idempotency check (skip update if values already match)
    - _Requirements: 3.1, 3.2, 3.4_
  - [x] 3.2 Simplify `PayoutService.GetConnectStatus` to read from the persisted bakery fields instead of making a live Stripe API call (optional improvement, reduces latency)
    - _Requirements: 3.1, 3.2_

- [x] 4. Verify existing refund wiring is complete
  - [x] 4.1 Verify `cmd/server/main.go` wires `payoutSvc` as `PayoutReverser` on the webhook handler (already done — confirm with a read)
    - _Requirements: 4.1, 4.2_
  - [x] 4.2 Verify `OrderService.DeleteOrder` calls `onOrderRefunded` when a refund is issued for a delivered order (already wired — confirm the callback is set)
    - _Requirements: 1.4, 4.1_
  - [x] 4.3 Verify `updateRefundStatus` calls `payoutReverser.OnOrderRefunded` for full refunds (already implemented — confirm)
    - _Requirements: 1.5, 2.1_

- [x] 5. Extend tests for Connect webhook onboarding sync
  - [x] 5.1 Add `TestConnectWebhookHandler_AccountUpdated_SetsOnboardingFlags` — verify `ChargesEnabled=true` from Stripe sets the bakery field
    - _Requirements: 3.1, 3.2_
  - [x] 5.2 Add `TestConnectWebhookHandler_AccountUpdated_IdempotentSkipsUpdate` — verify no DB write if values already match
    - _Requirements: 3.4_
  - [x] 5.3 Extend existing `TestConnectWebhookHandler_AccountUpdated_SyncsBakeryStatus` to assert `ChargesEnabled`/`PayoutsEnabled` on the updated bakery
    - _Requirements: 3.1, 3.2_

- [x] 6. Add test for payout reversal of non-transferred payout
  - [x] 6.1 Add `TestPayoutService_OnOrderRefunded_NonTransferredPayout` — verify a pending/failed payout is marked "refunded" without calling Stripe
    - _Requirements: 1.3_
  - [x] 6.2 Add `TestPayoutService_OnOrderRefunded_NoPayout` — verify nil return when no payout exists
    - _Requirements: 1.2_

- [x] 7. Checkpoint
  - Ensure all tests pass (`go test ./...`), ask the user if questions arise.

## Notes

- The core refund/reversal logic already exists and is properly wired. The main code change is adding onboarding fields to the bakery model and persisting them from the Connect webhook.
- `updateRefundStatus` is fully implemented (was the previous log-line ticket already completed). Tests verify lookup, persist, idempotency, and payout reversal triggering.
- `OnOrderRefunded` is called from two places: (1) `OrderService.DeleteOrder` when a captured payment is refunded on cancellation, and (2) `StripeWebhookHandler.updateRefundStatus` when a `charge.refunded` event arrives for a full refund.
- The postgres `UpdateBakery` currently doesn't write `stripe_connect_id` or the new boolean fields — this must be fixed.
- Migration follows the existing sequence: `025_add_bakery_onboarding_status.sql`.

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2"] },
    { "id": 1, "tasks": ["2.1", "2.2", "2.3"] },
    { "id": 2, "tasks": ["3.1", "3.2"] },
    { "id": 3, "tasks": ["4.1", "4.2", "4.3"] },
    { "id": 4, "tasks": ["5.1", "5.2", "5.3", "6.1", "6.2"] },
    { "id": 5, "tasks": ["7"] }
  ]
}
```
