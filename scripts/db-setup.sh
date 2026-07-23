#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# Local Postgres setup: start the DB (Docker), run migrations, load dummy data.
# Usage:  ./scripts/db-setup.sh
# Re-runnable: migrations and seed are idempotent.
# ---------------------------------------------------------------------------
set -euo pipefail

DSN="postgres://bakery:bakery@localhost:5432/bakery_app?sslmode=disable"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> 1/4  Starting Postgres (docker compose up -d db)"
if docker compose version >/dev/null 2>&1; then
  docker compose up -d db
else
  docker-compose up -d db   # older Compose v1
fi

echo "==> 2/4  Waiting for Postgres to be healthy"
for i in $(seq 1 30); do
  if docker exec bakery_db pg_isready -U bakery -d bakery_app >/dev/null 2>&1; then
    echo "    ready."
    break
  fi
  sleep 1
  [ "$i" = "30" ] && { echo "    Postgres did not become ready in time"; exit 1; }
done

echo "==> 3/4  Running migrations (goose)"
if ! command -v goose >/dev/null 2>&1; then
  echo "    goose not found — installing..."
  go install github.com/pressly/goose/v3/cmd/goose@latest
  export PATH="$PATH:$(go env GOPATH)/bin"
fi
goose -dir db/migrations postgres "$DSN" up

echo "==> 4/4  Loading dummy data (db/seed/seed.sql)"
docker exec -i bakery_db psql -U bakery -d bakery_app < db/seed/seed.sql

cat <<'EOF'

Done. Postgres is running with dummy data.

Point the backend at it by setting in your .env (or shell):
  DATABASE_URL=postgres://bakery:bakery@localhost:5432/bakery_app?sslmode=disable

Then start the backend:  make run
(Leaving DATABASE_URL unset keeps the old in-memory mode.)

Demo logins (password in parentheses):
  customer_demo (demo-customer)   customer
  baker_jean    (demo-baker)      baker — owns Bakery 1 & 3
  baker_marie   (demo-baker)      baker — owns Bakery 2
  admin_demo    (demo-admin)      admin
  pro_demo      (demo-b2b)        B2B / Comptoir

Stop the DB:   docker compose down
Wipe it:       docker compose down -v   (then re-run this script)

HEADS-UP: read/browse/login flows work on Postgres, but CREATE flows
(register, place order, reservation) currently FAIL on Postgres — see MA-62.
EOF
