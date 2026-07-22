# Implement PostgreSQL repository layer

**Milestone:** Foundations & Infra · **Priority:** High · **Type:** Feature

## Context

The app currently runs on in-memory repositories (`internal/repository/memory/`). Data is seeded on startup and lost on restart. The PostgreSQL schema is already fully defined via 13 goose migrations (`db/migrations/`), and the domain layer already defines repository interfaces (`internal/domain/repository.go`). This task implements those interfaces against PostgreSQL so the app can run with persistent storage, selectable at startup.

## Goal

A Postgres-backed implementation of every domain repository interface, wired behind the same interfaces the service layer already uses, selectable via configuration — with zero changes to the service or handler layers.

## Scope

- New package `internal/repository/postgres/` implementing each interface in `internal/domain/repository.go` (users, bakeries + day_schedules, products, orders + order_items, reservations, recurring_orders, registration_tokens).
- Database connection setup: `DATABASE_URL` env var, `pgxpool` (or `database/sql` + `pq`) connection pool, sensible pool limits and timeouts, ping on startup.
- Startup wiring in `cmd/server`: if `DATABASE_URL` is set, use Postgres repos; otherwise fall back to in-memory (preserves current demo behavior).
- Run migrations before serving when using Postgres (or document `make migrate-up` as a prerequisite — pick one and be explicit).
- Map DB rows ↔ domain models, including: prices/amounts stored as integer cents, `allergens[]` as a Postgres array, `role` as integer enum, order↔order_items composition, and the bakery↔day_schedules relation.
- Transactions where an operation spans multiple tables (e.g. creating an order + its order_items must be atomic).

## Out of scope

- Removing the in-memory repos (keep them for tests/demo).
- New schema/migrations (schema is considered complete; add a migration only if a genuine gap is found — flag it separately).
- Connection ret/resilience beyond basic pool config (circuit breakers, read replicas, etc.).

## Acceptance criteria

- [ ] `internal/repository/postgres/` implements all domain repository interfaces; project compiles with no changes to service/handler signatures.
- [ ] With `DATABASE_URL` set against a migrated DB, the full app runs: login, browse bakeries, place order, make reservation, seller status updates, recurring orders, holiday mode, favorites — all persist across a server restart.
- [ ] With `DATABASE_URL` unset, the app still boots on in-memory repos exactly as today.
- [ ] Order creation writes order + order_items atomically; a failure mid-write rolls back (no orphan rows).
- [ ] Money persists and reads back as integer cents; `allergens[]` round-trips correctly; `role` enum maps correctly.
- [ ] Integration tests run the existing repository/service test expectations against a real Postgres (Dockerized or testcontainers) and pass.
- [ ] No SQL injection surface: all queries parameterized.
- [ ] README/docs updated: how to run against Postgres (`DATABASE_URL`, `make migrate-up`).

## Suggested approach

1. Add config + pooled connection (`internal/config` or existing config path); ping on boot.
2. Implement repos one aggregate at a time, starting with `users` and `bakeries` (used by auth + listing), then `products`, then `orders`/`order_items`, `reservations`, `recurring_orders`, `registration_tokens`.
3. Introduce a small transaction helper so multi-table writes share one tx.
4. Wire selection in `cmd/server` behind `DATABASE_URL`.
5. Add integration tests against a disposable Postgres; verify persistence across a simulated restart.

## Dependencies / unblocks

- Pairs with **Wire `DATABASE_URL` config + connection pooling** (can be folded in or done first).
- Unblocks: real deployments, Docker/compose task (needs a DB service), and CI running integration tests against Postgres.

## Open questions

- Driver choice: `pgx`/`pgxpool` (recommended) vs `database/sql` + `lib/pq`?
- Migrations at app startup vs. explicit `make migrate-up` step in deploy?
- Should seed demo data optionally load into Postgres (e.g. `SEED=true`) for parity with the in-memory demo?
