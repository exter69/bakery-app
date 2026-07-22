# 🚀 Getting Started

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Backend server |
| Node.js | 20+ | Frontend dev server |
| npm | 10+ | Package management |

---

## Clone & Setup

```bash
git clone <repo-url>
cd mb
```

## Environment Variables

Copy the example env file:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend server port |
| `JWT_SECRET` | `your-secret-here` | JWT signing key |
| `FRONTEND_ORIGIN` | `http://localhost:5173` | CORS allowed origin |
| `CONTACT_EMAIL` | `admin@mieetbeurre.com` | Baker registration contact |

---

## Running the Backend

```bash
# From project root
go run ./cmd/server
```

The server starts on **http://localhost:8080** with seeded demo data.

---

## Running the Frontend

```bash
cd frontend
npm install
npm run dev
```

The dev server starts on **http://localhost:5173**.

---

## Test Accounts

The backend seeds these demo accounts on startup:

| Username | Password | Role |
|----------|----------|------|
| `admin` | `admin123` | Admin |
| `baker_jean` | `baker123` | Seller |
| `baker_marie` | `baker123` | Seller |
| `alice` | `customer123` | Customer |
| `bob` | `customer123` | Customer |

---

## Demo Registration Code

To test baker registration flow, use code: **`DEMO1234`**

---

## Build & Test

```bash
# Build backend binary
make build

# Run all backend tests
make test

# Run frontend tests
cd frontend && npm test

# Run E2E tests (requires both servers running)
cd e2e && npx playwright test
```

---

## Quick Start Checklist

- [ ] Clone repository
- [ ] Copy `.env.example` → `.env`
- [ ] Start backend: `go run ./cmd/server`
- [ ] Start frontend: `cd frontend && npm run dev`
- [ ] Open http://localhost:5173
- [ ] Login as `alice` / `customer123`
