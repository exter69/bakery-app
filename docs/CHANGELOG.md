# Changelog

## [2025-07-22] Payment confirmation emails + invoice generation

- **Module/App**: Backend (Go) — `internal/email`, `internal/invoice`, `internal/notification`, `internal/payment`, `internal/api`, `cmd/server`
- **Purpose**: After a payment is confirmed, automatically generate an HTML invoice and send a confirmation email to the customer. Implements MA-11.
- **Features/Areas**: Payment confirmation, invoice generation, email notifications, invoice retrieval API
- **Summary**: Added provider-agnostic email service (LogSender for dev, SMTPSender for production via SMTP env vars). Invoice generator renders HTML invoices from order/bakery/user data using Go templates. In-memory invoice store holds generated invoices. Notification service orchestrates the flow: fetch order → generate invoice → store → send email. Wired into payment service via `OnOrderConfirmed` callback (non-blocking — logs errors without failing the payment). Added `GET /api/orders/:id/invoice` endpoint (auth required, owner-only). Updated `.env.example` with SMTP config vars.
- **Tests**: 10 new tests across 3 packages (email: 3, invoice: 4, notification: 3). All 46 project tests pass.

## [2025-07-16] Stripe payment gateway integration

- **Module/App**: Backend (Go) — `internal/payment`, `cmd/server`
- **Purpose**: Add real payment processing via Stripe Checkout Sessions, replacing the stub gateway in production while keeping the stub for local development.
- **Features/Areas**: Payment processing, webhook handling, gateway selection
- **Summary**: Implemented `StripeGateway` (Checkout Sessions for creating payment URLs, session verification) and `StripeWebhookHandler` (signature-verified webhook endpoint for `checkout.session.completed` events). Updated `cmd/server/main.go` to select the gateway based on `PAYMENT_GATEWAY` env var ("stripe" or "stub"). Added Stripe-related env vars to `.env.example`. The webhook endpoint is registered as a public route (no JWT) since Stripe calls it directly.
- **Tests**: Added `stripe_gateway_test.go` with 6 tests covering config initialization, interface compliance, API error handling, and webhook signature validation. All 14 payment tests pass.
