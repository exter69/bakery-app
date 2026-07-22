# 🚢 Deployment

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP server port |
| `JWT_SECRET` | **Yes** | dev-secret (warning) | JWT signing key |
| `FRONTEND_ORIGIN` | No | `http://localhost:5173` | CORS allowed origin |
| `CONTACT_EMAIL` | No | `admin@mieetbeurre.com` | Baker registration contact |

> ⚠️ Always set `JWT_SECRET` to a strong random value in production.

---

## Building

### Backend

```bash
make build
# Output: bin/bakery-app
```

The binary is self-contained — no runtime dependencies.

### Frontend

```bash
cd frontend
npm run build
# Output: frontend/dist/
```

Produces static files ready for any web server or CDN.

---

## Running in Production

```bash
# Set environment
export PORT=8080
export JWT_SECRET=<strong-random-secret>
export FRONTEND_ORIGIN=https://yourdomain.com

# Run binary
./bin/bakery-app
```

---

## Current Storage Mode

The application currently uses **in-memory repositories**:

- All data is seeded on startup
- Data resets when the server restarts
- No database connection required
- Perfect for development and demos

---

## Future: PostgreSQL

The database schema is fully defined (13 migrations). To switch to persistent storage:

1. Provision PostgreSQL instance
2. Run migrations: `make migrate-up`
3. Implement SQL repository (replacing `internal/repository/memory/`)
4. Add `DATABASE_URL` environment variable

---

## Serving Frontend

Options for serving the built frontend:

1. **Separate web server** (Nginx, Caddy) serving `frontend/dist/`
2. **Embed in Go binary** using `embed` package
3. **CDN** (Cloudflare Pages, Vercel, Netlify)

Ensure the frontend's API calls point to the correct backend URL.

---

## Docker (Planned)

```dockerfile
# Example structure (not yet implemented)
FROM golang:1.21-alpine AS backend
WORKDIR /app
COPY . .
RUN go build -o /bakery-app ./cmd/server

FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/ .
RUN npm ci && npm run build

FROM alpine:3.18
COPY --from=backend /bakery-app /usr/local/bin/
COPY --from=frontend /app/dist /static
EXPOSE 8080
CMD ["bakery-app"]
```

---

## Health Check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Use this endpoint for load balancer health checks and container orchestration.
