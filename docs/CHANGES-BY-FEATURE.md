# Changes by Feature

History grouped by feature. Every ticket that adds/changes a feature appends one row to that feature's table (see `.kiro/steering/release-versioning.md` §4). Append only.

All rows are `pre-1.0`: the app is not deployed yet — the first production deployment becomes `v1.0.0`. Dates use git commit dates (2026).

## Core ordering & reservations (Ma Boulangerie)

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-17 | — | pre-1.0 | Initial app: browse bakeries, delivery orders, pickup reservations, order state machine |
| 2026-07-22 | MA-16 | pre-1.0 | Order history + re-order button |
| 2026-07-23 | MA-15 | pre-1.0 | Product search & filtering |
| 2026-07-23 | MA-17 | pre-1.0 | Customer reviews & ratings (backend + frontend) |
| 2026-07-28 | MA-61 | pre-1.0 | Portal rebranding: consumer portal renamed to "Ma Boulangerie" |

## Payments (B2C — Stripe)

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-22 | MA-10 | pre-1.0 | Stripe payment gateway (Payment Intents) replacing stub |
| 2026-07-22 | MA-11 | pre-1.0 | Payment confirmation emails + invoice generation |
| 2026-07-22 | MA-25 | pre-1.0 | Saved payment methods (cards on file) |
| 2026-07-22 | MA-33 | pre-1.0 | Delayed capture: authorize on order, capture on delivery |
| 2026-07-22 | MA-24 | pre-1.0 | Refunds & order cancellation handling |
| 2026-07-28 | MA-63 | pre-1.0 | Payment callback endpoint hardened: gated by mode, ownership + state checks, IDOR fix |
| 2026-07-28 | MA-65 | pre-1.0 | charge.refunded webhook now persists refund state idempotently via PI lookup |

## Marketplace payouts (Stripe Connect)

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-23 | MA-57 | pre-1.0 | Connect Express payouts: commission split, transfer on delivery, reversal on refund, seller payouts page |
| 2026-07-23 | MA-60 | pre-1.0 | Connect webhook handler (account.updated), dashboard onboarding banner, i18n keys |
| 2026-07-28 | MA-65 | pre-1.0 | Wire refund to payout reversal: updateRefundStatus persists state from charge.refunded webhook, OnOrderRefunded called on order cancellation refund, account.updated syncs bakery Connect status |

## Notifications

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-22 | MA-12 | pre-1.0 | Email notifications: orders, status changes, baker alerts |
| 2026-07-22 | MA-13 | pre-1.0 | Real-time notifications via WebSocket |
| 2026-07-22 | MA-14 | pre-1.0 | Push notifications (PWA, VAPID) |

## Auth & accounts

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-17 | — | pre-1.0 | JWT auth + roles (Admin/Seller/Customer) |
| 2026-07-23 | MA-55 | pre-1.0 | SSO / social login (Google, Apple) |
| 2026-07-23 | MA-58 | pre-1.0 | GDPR: data export, account deletion, cookie consent, privacy/terms pages |
| 2026-07-28 | MA-66 | pre-1.0 | Auth hardening: prod boot guard for secrets, server-signed OAuth state (CSRF), safe account linking, crypto/rand tokens, removed X-User-ID header fallback |

## Surplus boxes — "Panier du soir"

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-23 | MA-30 | pre-1.0 | End-of-day surplus boxes: bundles, reservations, real-time stock (WebSocket) |
| 2026-07-23 | MA-38 | pre-1.0 | Baker-side bundle composer (Pro portal) |
| 2026-07-23 | MA-49 | pre-1.0 | Guest preview of bundles, login gate on reservation |

## Baker portal (Votre Boulangerie)

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-23 | MA-34 | pre-1.0 | Rebrand + Pro design system |
| 2026-07-23 | MA-37 | pre-1.0 | Menu & stock: card inventory, inline stock editing, day toggles |
| 2026-07-23 | — | pre-1.0 | Orders as kanban board with drag-and-drop; KPI overview dashboard |
| 2026-07-28 | MA-61 | pre-1.0 | Portal rebranding: baker portal renamed to "Votre Boulangerie" (collapsed: "VB") |

## B2B portal (Notre Boulangerie)

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-23 | MA-26 | pre-1.0 | B2B account type & registration option |
| 2026-07-23 | MA-27 | pre-1.0 | Comptoir portal shell (nav, gating, business theme) |
| 2026-07-23 | MA-28 | pre-1.0 | Commande rapide: spreadsheet-style order sheet |
| 2026-07-23 | MA-47 | pre-1.0 | Multi-bakery basket |
| 2026-07-23 | MA-39 | pre-1.0 | Récurrences: standing orders per weekday |
| 2026-07-23 | MA-29 | pre-1.0 | Factures: monthly statements + SEPA billing |
| 2026-07-28 | MA-41 | pre-1.0 | B2B pricing & VAT: per-account pro discount, configurable VAT rate, volume tiers with rolling monthly spend, next-tier nudge |
| 2026-07-28 | MA-61 | pre-1.0 | Portal rebranding: B2B portal renamed to "Notre Boulangerie" |

## Platform & infrastructure

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| 2026-07-22 | MA-5 | pre-1.0 | PostgreSQL repository layer (goose migrations, DATABASE_URL) |
| 2026-07-23 | MA-9 | pre-1.0 | Production deployment: Vercel + Railway + HTTPS (first prod deploy) |
| 2026-07-23 | MA-31 | pre-1.0 | Security hardening & vulnerability review |
| 2026-07-23 | MA-22 | pre-1.0 | Error tracking (Sentry) |
| 2026-07-23 | MA-19 | pre-1.0 | Image upload for products & bakeries |
| 2026-07-23 | — | pre-1.0 | GitHub Actions CI/CD pipeline |
| 2026-07-23 | MA-20 | pre-1.0 | OpenAPI 3.0 specification |
| 2026-07-23 | MA-21 | pre-1.0 | Storybook component library |
| 2026-07-23 | MA-56 | pre-1.0 | UI performance: code splitting, lazy loading |
| 2026-07-23 | MA-18 | pre-1.0 | Dark mode |
| 2026-07-28 | MA-62 | pre-1.0 | Postgres fix: UUID IDs, role CHECK widened to 0..3, data-race elimination |
| 2026-07-28 | MA-66 | pre-1.0 | Rate limiter hardened: keyed on RemoteAddr, stale entry eviction, register endpoint rate-limited |
