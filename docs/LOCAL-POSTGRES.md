# Running locally against PostgreSQL

By default the backend runs on **in-memory** repositories (data seeded at startup, lost on restart). This guide switches it to a real **Postgres** running in Docker, with dummy data.

> ⚠️ **Current limitation (MA-62).** On Postgres, **read/browse/login flows work**, but **create flows fail** — registering a user, placing an order, or making a reservation errors out because the app generates non-UUID IDs (`order-1`, `user-1`) that violate the `UUID` primary-key columns. Until MA-62 lands, use Postgres for browsing seeded data and keep in-memory mode for exercising the full write path. The seed data below is inserted directly with valid UUIDs, so it works regardless.

## Prerequisites

- Docker (Desktop or engine) with Compose v2
- Go (for the backend and the `goose` migration tool — auto-installed by the script if missing)

## One command

```bash
./scripts/db-setup.sh
```

This starts Postgres, runs all migrations, and loads dummy data. It's idempotent — safe to re-run.

Then tell the backend to use it by adding this line to your `.env`:

```
DATABASE_URL=postgres://bakery:bakery@localhost:5432/bakery_app?sslmode=disable
```

and start it:

```bash
make run          # backend now uses Postgres
cd frontend && npm run dev
```

Leaving `DATABASE_URL` unset returns you to in-memory mode — nothing else changes.

## What you get

Connection: `postgres://bakery:bakery@localhost:5432/bakery_app` (user `bakery`, password `bakery`, db `bakery_app`).

Dummy accounts (passwords in parentheses):

| Username | Password | Role | Notes |
|----------|----------|------|-------|
| `customer_demo` | `demo-customer` | Customer | Has 2 sample orders in history |
| `baker_jean` | `demo-baker` | Baker | Owns Bakery 1 & 3 |
| `baker_marie` | `demo-baker` | Baker | Owns Bakery 2 |
| `admin_demo` | `demo-admin` | Admin | — |
| `pro_demo` | `demo-b2b` | B2B / Comptoir | Reaches `/comptoir` (see note below) |

Plus 3 bakeries, 15 products, and 2 sample orders. Defined in [`db/seed/seed.sql`](../db/seed/seed.sql) — edit and re-run to change.

## Manual steps (if you prefer not to use the script)

```bash
# 1. Start the database
docker compose up -d db

# 2. Run migrations
goose -dir db/migrations postgres "postgres://bakery:bakery@localhost:5432/bakery_app?sslmode=disable" up

# 3. Load dummy data
docker exec -i bakery_db psql -U bakery -d bakery_app < db/seed/seed.sql
```

`make migrate-up` / `migrate-status` / `migrate-reset` also work if you pass the matching DSN:

```bash
make migrate-status DB_DSN="postgres://bakery:bakery@localhost:5432/bakery_app?sslmode=disable"
```

## Housekeeping

```bash
docker compose down       # stop the DB (data kept in the named volume)
docker compose down -v     # stop AND wipe all data
docker exec -it bakery_db psql -U bakery -d bakery_app   # open a psql shell
```

If you already run Postgres on `5432`, either stop it or change the host port in `docker-compose.yml` (e.g. `"5433:5432"`) and update the DSN to `:5433`.

## Notes & gaps

- **B2B profile:** `pro_demo` exists in `users` with role 3, which is enough to pass the `/comptoir` route guard and log in. The richer B2B profile (company, VAT, sites — migration `020`) is **not** seeded, so some Comptoir screens may show empty or error until that data exists. Extending the seed here is tracked in the enablement ticket.
- **No auto-seeding in DB mode:** the app's built-in `seedDemoData` only runs in in-memory mode, which is why dummy data is provided as SQL instead. Making the app seed (or not) on Postgres is part of the enablement ticket.
- **Migrations on deploy** are still manual (see `docs/DEPLOYMENT.md` and MA-72).
