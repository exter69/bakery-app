# Ma Boulangerie

A marketplace connecting artisan bakeries with customers and business buyers — delivery orders with online payment, pickup reservations, end-of-day surplus boxes, and wholesale B2B ordering.

## The three portals

One SPA serves three distinct portals. The names are **invariant across all languages** (EN/FR/NL always show the French name — see [MA-61](https://linear.app/ma-boulangerie/issue/MA-61/portal-rebranding-ma-notre-votre-boulangerie) for the rename rollout):

| Portal | Name | Audience | Routes |
|--------|------|----------|--------|
| Consumer app | **Ma Boulangerie** | Customers (B2C) | `/` |
| B2B buyer portal | **Notre Boulangerie** | Restaurants, cafés, offices | `/comptoir/*` |
| Baker back-office | **Votre Boulangerie** | Bakers (sellers) | `/dashboard/*` |

Consumers buy on *Ma*, businesses buy on *Notre*, bakers sell on *Votre*.

## Features

- **Ordering & payments** — delivery orders paid online via Stripe (authorize at order, capture on delivery), refunds on cancellation, saved cards, confirmation emails + invoices
- **Reservations** — pickup reservations, always paid on the spot
- **Paniers du soir** — end-of-day surplus boxes with real-time stock (WebSocket), guest preview
- **Recurring orders** — weekly/bi-weekly standing orders; per-weekday récurrences for B2B
- **B2B portal** — spreadsheet-style order sheet, multi-bakery basket, deliveries, monthly statements + SEPA billing
- **Baker portal** — inventory with day toggles, kanban order board, bundle composer, Stripe Connect payouts
- **Accounts** — JWT auth, Google/Apple SSO, roles (Admin/Seller/Customer + B2B account type)
- **GDPR** — data export, account deletion, cookie consent, privacy/terms pages
- **Platform** — i18n (EN/FR/NL), dark mode, PWA push notifications, Sentry, GitHub Actions CI

## Stack

| Layer | Tech |
|-------|------|
| Backend | Go 1.26 · chi · JWT · goose migrations |
| Frontend | React 19 · TypeScript · Vite |
| Database | PostgreSQL (in-memory repos available for dev) |
| Payments | Stripe (Payment Intents, webhooks, Connect payouts) |
| Testing | Go unit + property tests · Vitest · Playwright E2E · Storybook |
| Hosting | Vercel (frontend) · Railway (backend + Postgres) |

## Quick Start

Prerequisites: Go 1.22+, Node.js 18+, (optional) PostgreSQL.

```bash
# Backend — http://localhost:8080
make run

# Frontend — http://localhost:5173
cd frontend
npm install
npm run dev
```

Without `DATABASE_URL`, the backend runs on in-memory repositories seeded at startup (dev only). With `DATABASE_URL` set, it uses PostgreSQL — run migrations first:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
DB_DSN="postgres://localhost:5432/bakery_app?sslmode=disable" make migrate-up
```

## Configuration

Core variables (full reference incl. Stripe, SMTP, OAuth, VAPID: [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md)):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend port |
| `DATABASE_URL` | *(unset → in-memory)* | PostgreSQL connection string |
| `JWT_SECRET` | dev value | JWT signing secret — set in production |
| `FRONTEND_ORIGIN` | `http://localhost:5173` | Allowed CORS origin |
| `PAYMENT_GATEWAY` | stub | `stripe` in production |
| `VITE_API_BASE_URL` | — | (frontend) backend URL **including `/api`** |

## API

The API is documented in the OpenAPI 3.0 spec under [`api/`](api/) and in the Notion docs (12 · API Reference). Auth endpoints (`/api/auth/*`) issue JWTs; authenticated routes expect `Authorization: Bearer <token>`.

## Commands

```bash
make build / run / test / test-race / test-cover
make fmt / vet / lint / tidy / check
make migrate-up / migrate-status

cd frontend
npm run dev / build / preview / lint / storybook
npx playwright test   # E2E (e2e/)
```

## Project Structure

```
├── cmd/server/          # Entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── domain/          # Models, interfaces, order state machine
│   ├── middleware/      # JWT, rate limiting, sanitization
│   ├── payment/         # Stripe gateway, webhooks, Connect payouts
│   ├── repository/      # Postgres + in-memory implementations
│   ├── service/         # Business logic
│   └── validation/
├── db/migrations/       # PostgreSQL schema (goose)
├── frontend/src/
│   ├── pages/           # Consumer pages · comptoir/ (B2B) · dashboard/ (baker)
│   ├── i18n/            # EN/FR/NL translations
│   └── api/ components/ hooks/ theme/
├── api/                 # OpenAPI spec
├── e2e/                 # Playwright tests
└── docs/                # CHANGELOG, DEPLOYMENT, SECURITY-CHECKLIST, DATA-INVENTORY
```

## Key Design Decisions

- **Money as `int64` cents** — no floating point
- **Order state machine**: PendingPayment → Confirmed → Preparing → Ready → Delivered; bakery payout triggers on Delivered
- **Authorize-then-capture**: customers are charged on delivery, not at order time
- **Repository interfaces**: Postgres and in-memory are swappable without touching business logic
- **Rate limiting** (10 submissions/user/min) and input sanitization on all JSON fields

## Documentation

- **Notion** — full docs hub "Mie & Beurre — App Documentation": business (15), quick technical (16), deployment & versions (17), plus pages 01–14
- [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) — production setup (Vercel + Railway)
- [`docs/CHANGELOG.md`](docs/CHANGELOG.md) — per-change log with Linear IDs
- [`docs/SECURITY-CHECKLIST.md`](docs/SECURITY-CHECKLIST.md) · [`docs/DATA-INVENTORY.md`](docs/DATA-INVENTORY.md) (GDPR)
