# Requirements: Refund-to-Payout Reversal and Onboarding Sync

## Overview

Wire the existing but uncalled `PayoutService.OnOrderRefunded` into the refund flow so that refunding a delivered order reverses the bakery's Stripe Connect transfer. Implement the `updateRefundStatus` function to persist refund state idempotently. Handle `account.updated` Connect webhooks to sync bakery onboarding status automatically.

## Linked Ticket

MA-65 — Wire refund to payout reversal; implement updateRefundStatus (currently a log line)

---

## Requirement 1: Transfer Reversal on Refund

### Acceptance Criteria

1.1 WHEN an order with status "delivered" is fully refunded, THE PayoutService SHALL reverse the Stripe Connect transfer associated with that order and mark the payout record as "refunded".

1.2 WHEN `OnOrderRefunded` is called for an order that has no payout record, THE PayoutService SHALL return nil (no-op) without error.

1.3 WHEN `OnOrderRefunded` is called for a payout that is in "pending" or "failed" status (no transfer was made), THE PayoutService SHALL mark the payout as "refunded" without calling the Stripe transfer reversal API.

1.4 WHEN a cancellation of a delivered order triggers a Stripe refund, THE OrderService SHALL call `OnOrderRefunded` to reverse the bakery transfer.

1.5 WHEN a `charge.refunded` webhook indicates a full refund, THE StripeWebhookHandler SHALL call `OnOrderRefunded` to reverse the bakery transfer.

---

## Requirement 2: Persist Refund Status from Webhook

### Acceptance Criteria

2.1 WHEN a `charge.refunded` webhook event is received, THE StripeWebhookHandler SHALL look up the order by PaymentIntent ID and persist the refund status ("refunded" for full, "partial" for partial).

2.2 WHEN the order already has the same refund status as the incoming event, THE StripeWebhookHandler SHALL skip the update (idempotent no-op).

2.3 WHEN the PaymentIntent ID does not match any order, THE StripeWebhookHandler SHALL log a warning and return HTTP 200 (no retry).

2.4 WHEN a `charge.refunded` event is replayed, THE system SHALL produce the same persisted state as the first delivery of that event.

---

## Requirement 3: Connect Onboarding Status Sync

### Acceptance Criteria

3.1 WHEN a Stripe `account.updated` Connect webhook is received with `charges_enabled=true`, THE ConnectWebhookHandler SHALL update the bakery's `charges_enabled` field to true.

3.2 WHEN a Stripe `account.updated` Connect webhook is received with `payouts_enabled=true`, THE ConnectWebhookHandler SHALL update the bakery's `payouts_enabled` field to true.

3.3 WHEN the `account.updated` webhook reports a Stripe Connect ID that matches no bakery in the database, THE ConnectWebhookHandler SHALL log the event and return HTTP 200.

3.4 WHEN the bakery's stored `charges_enabled` and `payouts_enabled` already match the incoming values, THE ConnectWebhookHandler SHALL skip the database update (idempotent).

---

## Requirement 4: Wiring and Integration

### Acceptance Criteria

4.1 THE `PayoutService.OnOrderRefunded` method SHALL be reachable from both the order cancellation flow and the `charge.refunded` webhook handler.

4.2 THE `StripeWebhookHandler` SHALL expose a `SetPayoutReverser` method so that `PayoutService` can be injected without import cycles.

4.3 THE `ConnectWebhookHandler` SHALL use the injected `BakeryRepository` to persist onboarding status changes.

