# Deployment Guide

Production deployment for Mie & Beurre: frontend on Vercel, backend + PostgreSQL on Railway.

## Architecture

```
                    HTTPS                         HTTPS
  [Browser] ──────────────> [Vercel CDN] ──── Static SPA (React/Vite)
      │
      │  API calls (CORS)
      └──────────────────> [Railway] ──── Go binary + PostgreSQL
```

- **Frontend**: Vercel serves the built Vite SPA with client-side routing.
- **Backend**: Railway runs the Go binary in a Docker container.
- **Database**: Railway Postgres plugin, provisioned alongside the backend service.
- **HTTPS**: Both Vercel and Railway provide automatic TLS certificates.

## Required Environment Variables

### Railway (backend)

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Auto | Provided by Railway Postgres plugin |
| `JWT_SECRET` | Yes | Strong random secret for JWT signing |
| `FRONTEND_ORIGIN` | Yes | Vercel production URL (e.g., `https://mieetbeurre.com`) |
| `PAYMENT_GATEWAY` | Yes | `stripe` for production |
| `STRIPE_SECRET_KEY` | Yes | Stripe live secret key |
| `STRIPE_WEBHOOK_SECRET` | Yes | Stripe webhook signing secret |
| `PORT` | No | Defaults to `8080` (Railway sets this automatically) |
| `CONTACT_EMAIL` | No | Defaults to `admin@mieetbeurre.com` |
| `SMTP_HOST` | No | Email sending (logs to stdout if unset) |
| `SMTP_PORT` | No | Defaults to 587 |
| `SMTP_USERNAME` | No | SMTP auth |
| `SMTP_PASSWORD` | No | SMTP auth |
| `SMTP_FROM` | No | Sender address |
| `VAPID_PUBLIC_KEY` | No | Web push (disabled if unset) |
| `VAPID_PRIVATE_KEY` | No | Web push (disabled if unset) |

### Vercel (frontend)

| Variable | Required | Description |
|----------|----------|-------------|
| `VITE_API_BASE_URL` | Yes | Railway backend URL **including `/api`** |

Example: `VITE_API_BASE_URL=https://bakery-app-production.up.railway.app/api`

The frontend client does `fetch(VITE_API_BASE_URL + path)` where path is like `/bakeries`, `/orders`, etc. The backend registers routes under `/api/...`, so the env var must include the `/api` suffix.

## Health Check

The backend exposes a public health endpoint:

```
GET /health → 200 {"status":"ok"}
```

Railway is configured to probe this endpoint (see `railway.toml`).

## Database Migrations

Migrations are managed with [goose](https://github.com/pressly/goose). The migration files are bundled in the Docker image at `db/migrations/`.

Run migrations manually against the Railway Postgres instance:

```bash
# Install goose locally
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations (get DATABASE_URL from Railway dashboard)
goose -dir db/migrations postgres "$DATABASE_URL" up
```

For automated migrations on deploy, add a Railway deploy command or use a one-off service:

```bash
# Option A: Run before the main service starts (Railway deploy command)
# In Railway dashboard → Service → Settings → Deploy command:
# goose -dir db/migrations postgres "$DATABASE_URL" up && ./server

# Option B: Separate one-off Railway service that runs migrations only
```

Note: The Dockerfile does NOT run migrations automatically. This is intentional — migrations should be a deliberate step to avoid issues with concurrent deploys.

## CORS Configuration

The backend restricts CORS to the origin specified in `FRONTEND_ORIGIN`. This must exactly match the Vercel production domain including protocol:

- Correct: `https://mieetbeurre.com`
- Wrong: `https://mieetbeurre.com/` (trailing slash)
- Wrong: `mieetbeurre.com` (missing protocol)

If you use a Vercel preview URL for staging, set `FRONTEND_ORIGIN` to that URL in your Railway staging environment.

## Setup from Scratch

### 1. Railway (backend + database)

1. Create a new Railway project
2. Add a **PostgreSQL** plugin — this auto-provides `DATABASE_URL`
3. Add a new service from your GitHub repo
4. Railway will detect the `Dockerfile` and `railway.toml` automatically
5. Set environment variables in the Railway dashboard:
   - `JWT_SECRET` — generate with `openssl rand -base64 32`
   - `FRONTEND_ORIGIN` — your Vercel domain (set after Vercel deploy)
   - `PAYMENT_GATEWAY=stripe`
   - `STRIPE_SECRET_KEY` — from Stripe dashboard
   - `STRIPE_WEBHOOK_SECRET` — from Stripe webhook config
   - Optional: SMTP and VAPID vars
6. Run migrations (see above)
7. Deploy — Railway builds the Docker image and starts the service
8. Note the Railway public URL (e.g., `https://bakery-app-production.up.railway.app`)

### 2. Vercel (frontend)

1. Import the GitHub repo in Vercel
2. Set the **Root Directory** to `frontend` (or leave as repo root — `vercel.json` handles it)
3. Framework preset: **Vite**
4. Set environment variable:
   - `VITE_API_BASE_URL=https://bakery-app-production.up.railway.app/api`
5. Deploy

### 3. Post-deploy verification

1. Check Railway health: `curl https://your-railway-url.up.railway.app/health`
2. Check Vercel serves the SPA: visit the Vercel URL
3. Verify CORS: open browser devtools, confirm API calls succeed without CORS errors
4. Set up Stripe webhook endpoint: `https://your-railway-url.up.railway.app/api/stripe/webhook`

## Custom Domain (optional)

### Vercel
1. Go to Vercel project settings → Domains
2. Add your domain (e.g., `mieetbeurre.com`)
3. Configure DNS (CNAME to `cname.vercel-dns.com` or A records)
4. Vercel auto-provisions a TLS certificate

### Railway
1. Go to Railway service settings → Networking → Custom Domain
2. Add your API domain (e.g., `api.mieetbeurre.com`)
3. Configure DNS (CNAME to the Railway-provided value)
4. Railway auto-provisions a TLS certificate
5. Update `VITE_API_BASE_URL` on Vercel to `https://api.mieetbeurre.com/api`
6. Update `FRONTEND_ORIGIN` on Railway if you changed the frontend domain

## Stripe Webhook

Configure the Stripe webhook in the Stripe dashboard:
- Endpoint URL: `https://your-railway-url.up.railway.app/api/stripe/webhook`
- Events: `payment_intent.succeeded`, `payment_intent.payment_failed`
- Copy the webhook signing secret to `STRIPE_WEBHOOK_SECRET` on Railway

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| CORS errors in browser | `FRONTEND_ORIGIN` mismatch | Ensure it matches the exact Vercel URL |
| 502 on Railway | App crashed on startup | Check Railway logs, likely missing env var |
| API returns 401 | JWT_SECRET differs between deploys | Ensure consistent secret across redeploys |
| Frontend shows no data | `VITE_API_BASE_URL` wrong | Must include `/api` suffix |
| Health check failing | Port mismatch | Ensure `PORT` env var matches `railway.toml` internalPort |

## Sentry Source Maps (future)

For production debugging with readable stack traces, the Vercel build should upload source maps to Sentry. Two options:

1. **Sentry Vercel Integration** (recommended): Install the Sentry integration from the Vercel marketplace. It auto-uploads source maps during each deploy without modifying the build script.

2. **Manual via sentry-cli**: Add a post-build step in the Vercel build command:
   ```bash
   npx @sentry/cli releases new $VITE_APP_VERSION
   npx @sentry/cli releases files $VITE_APP_VERSION upload-sourcemaps ./dist
   npx @sentry/cli releases finalize $VITE_APP_VERSION
   ```
   Requires `SENTRY_AUTH_TOKEN`, `SENTRY_ORG`, and `SENTRY_PROJECT` env vars in the Vercel dashboard.

Note: Source maps should NOT be served publicly. Configure Vite to generate them as hidden (`build.sourcemap: 'hidden'`) and let Sentry ingest them during the build step only.
