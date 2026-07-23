# Requirements Document

## Introduction

This specification addresses six operational and security defects in the bakery-app backend that prevent reliable, secure deployments on Railway (or any container-based PaaS): a Go toolchain mismatch, missing database migration step, lack of graceful HTTP shutdown, an overly permissive CORS development mode, a data-corrupting JSON sanitizer middleware, and a minor readability issue in rate-limiter configuration.

## Linked Ticket

MA-72 -- Ops/deploy fixes: Go version mismatch, no migrations on deploy, no graceful shutdown, CORS dev reflection, JSON sanitizer

## Glossary

- **Server**: The Go HTTP server process started by `cmd/server/main.go`.
- **Dockerfile**: The multi-stage container image build definition at repository root.
- **Goose**: The SQL migration tool (`github.com/pressly/goose/v3`) used for database schema management.
- **CORS_Middleware**: The `internal/middleware/cors.go` middleware that sets Cross-Origin Resource Sharing headers.
- **Sanitizer_Middleware**: The `internal/middleware/sanitize.go` middleware that modifies JSON request bodies.
- **Rate_Limiter**: The `internal/middleware/ratelimit.go` per-user/IP rate limiter.
- **Railway**: The PaaS deployment target that sends SIGTERM before killing processes.

---

## Requirements

### Requirement 1: Docker Image Go Version Pinning

**User Story:** As a DevOps engineer, I want the Docker build image to match the Go version declared in go.mod, so that builds are reproducible without relying on automatic toolchain downloads.

#### Acceptance Criteria

1. THE Dockerfile SHALL use a builder base image whose Go version matches the version declared in `go.mod`.

2. WHEN `go.mod` declares a new Go version, THE Dockerfile SHALL be updated to reference the corresponding `golang:<version>-alpine` image.

3. WHEN CI runs the `docker build` step, THE build SHALL succeed without downloading a newer Go toolchain at runtime.

---

### Requirement 2: Automated Database Migrations on Deploy

**User Story:** As a platform engineer, I want database migrations to run automatically when the application starts, so that a fresh Railway database boots with the correct schema without manual intervention.

#### Acceptance Criteria

1. WHEN the Server starts and `DATABASE_URL` is set, THE Server SHALL run all pending Goose migrations from `db/migrations/` before accepting HTTP traffic.

2. IF a migration fails, THEN THE Server SHALL log the error and exit with a non-zero status code.

3. WHEN the database is already up-to-date, THE Server SHALL proceed to start the HTTP listener without error.

4. WHEN `DATABASE_URL` is not set (in-memory mode), THE Server SHALL skip the migration step.

---

### Requirement 3: Graceful HTTP Shutdown

**User Story:** As a platform engineer, I want the HTTP server to drain in-flight requests on SIGTERM, so that deployments do not kill active client connections.

#### Acceptance Criteria

1. WHEN a SIGTERM or SIGINT signal is received, THE Server SHALL stop accepting new connections and wait for in-flight requests to complete.

2. THE Server SHALL enforce a maximum shutdown grace period of 15 seconds; IF in-flight requests have not completed within that period, THEN THE Server SHALL force-close remaining connections and exit.

3. WHEN the signal is received, THE Server SHALL also cancel background workers (bundle expiration) via the existing context cancellation.

4. THE Server SHALL log when graceful shutdown begins and when it completes.

---

### Requirement 4: Secure CORS in Development Mode

**User Story:** As a security engineer, I want development CORS to use a fixed allowlist rather than reflecting arbitrary origins with credentials, so that local development does not train developers to ignore credential-bearing cross-origin requests from untrusted origins.

#### Acceptance Criteria

1. WHILE `APP_ENV=development`, THE CORS_Middleware SHALL allow requests only from a fixed set of local development origins (`http://localhost:5173`, `http://localhost:3000`, `http://127.0.0.1:5173`, `http://127.0.0.1:3000`).

2. WHEN the request origin is not in the allowed set, THE CORS_Middleware SHALL omit the `Access-Control-Allow-Origin` header entirely from the response.

3. THE CORS_Middleware SHALL set `Access-Control-Allow-Credentials: true` only when it also sets a matching `Access-Control-Allow-Origin` header.

4. WHEN `APP_ENV` is not `development`, THE CORS_Middleware SHALL allow only the single origin specified in `FRONTEND_ORIGIN`.

---

### Requirement 5: Remove Global JSON Sanitizer

**User Story:** As a developer, I want the global JSON sanitizer middleware removed, so that passwords and other legitimate data containing angle-bracket characters are not corrupted, and so that XSS prevention is handled correctly at the output layer.

#### Acceptance Criteria

1. THE Server SHALL NOT apply the `InputSanitizer` middleware globally.

2. WHEN a user submits a JSON body containing angle-bracket characters (e.g., a password like `p<a>ss`), THE Server SHALL preserve the value unchanged through to the service layer.

3. THE Server SHALL rely on output encoding (Go template escaping, JSON marshaling) and targeted input validation (per-field, in DTOs/service layer) for XSS prevention instead of blanket regex stripping.

4. THE `SanitizeString` function SHALL remain available for explicit per-field use (e.g., review text body) but SHALL NOT be applied to all request bodies.

---

### Requirement 6: Rate Limiter Readability

**User Story:** As a developer, I want the rate limiter window duration expressed using `time.Duration` constants, so that the configuration is self-documenting.

#### Acceptance Criteria

1. THE rate limiter configuration in `cmd/server/main.go` SHALL express the window duration using `time.Minute` instead of a raw nanosecond literal.
