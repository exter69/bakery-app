# Design: Ops/Deploy Fixes

## Overview

This design addresses six operational defects in the bakery-app backend. The changes span the Dockerfile, server startup/shutdown lifecycle, CORS middleware, the global JSON sanitizer, and a minor readability fix. Each fix is isolated and introduces no new external dependencies beyond what the project already uses (goose for migrations is already a CI dependency).

## Architecture

The fixes touch the following layers:

```
Dockerfile (build)
  |
cmd/server/main.go (startup lifecycle, middleware wiring)
  |
internal/middleware/cors.go (CORS policy)
internal/middleware/sanitize.go (removal from global chain; function retained)
internal/middleware/ratelimit.go (no change — usage site only)
```

No new services, repositories, or domain types are introduced. The migration runner is embedded in the server startup path using goose's library API.

## Components and Interfaces

### 1. Dockerfile Version Pinning

**Current state:** `FROM golang:1.22-alpine` while `go.mod` declares `go 1.26.5`.

**Change:** Update the FROM line to `golang:1.26-alpine` (matching the major.minor from go.mod). Alpine images use major.minor tags; the patch version is resolved by the image registry.

### 2. Migration Runner (`cmd/server/main.go`)

**Current state:** Migrations are bundled in the image (`COPY db/migrations`) but never executed.

**Change:** After connecting to PostgreSQL (pool creation), call goose programmatically:

```go
import "github.com/pressly/goose/v3"

func runMigrations(pool *pgxpool.Pool) error {
    db := stdlib.OpenDBFromPool(pool)
    goose.SetDialect("postgres")
    return goose.Up(db, "db/migrations")
}
```

- Called only when `DATABASE_URL` is set (Postgres mode).
- On failure: log error, close pool, `log.Fatalf(...)`.
- On success: log applied count and proceed.

**New dependency:** `github.com/jackc/pgx/v5/stdlib` (already transitively available via pgx).
**Existing dependency:** `github.com/pressly/goose/v3` (add to go.mod — already used in CI).

### 3. Graceful HTTP Shutdown (`cmd/server/main.go`)

**Current state:** `http.ListenAndServe` blocks forever; SIGTERM only cancels the bundle worker context.

**Change:** Replace with `http.Server` + goroutine pattern:

```go
srv := &http.Server{
    Addr:    ":" + port,
    Handler: r,
}

// Start HTTP server in a goroutine
go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("HTTP server error: %v", err)
    }
}()

log.Printf("Starting bakery-app server on :%s", port)

// Block until signal
<-ctx.Done()
log.Println("Shutting down gracefully...")

shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
defer shutdownCancel()

if err := srv.Shutdown(shutdownCtx); err != nil {
    log.Printf("Forced shutdown: %v", err)
}
log.Println("Server stopped")
```

The existing `signal.NotifyContext` already handles SIGINT/SIGTERM and cancels the bundle worker. The HTTP shutdown is added after `<-ctx.Done()`.

### 4. CORS Middleware Hardening (`internal/middleware/cors.go`)

**Current state:** Dev mode reflects any `Origin` header value back with `Allow-Credentials: true`.

**Change:**

```go
var devAllowedOrigins = map[string]bool{
    "http://localhost:5173":   true,
    "http://localhost:3000":   true,
    "http://127.0.0.1:5173":  true,
    "http://127.0.0.1:3000":  true,
}

func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
    isDev := strings.EqualFold(os.Getenv("APP_ENV"), "development")

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            allowed := false

            if isDev {
                allowed = devAllowedOrigins[origin]
            } else {
                allowed = origin != "" && origin == cfg.AllowedOrigin
            }

            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }

            // Common headers (set regardless — browsers ignore them without Allow-Origin)
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
            w.Header().Set("Access-Control-Expose-Headers", "Link")
            w.Header().Set("Access-Control-Max-Age", "300")

            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

Key invariant: `Allow-Credentials` is only sent when `Allow-Origin` is also sent.

### 5. Sanitizer Middleware Removal

**Current state:** `appmw.InputSanitizer` is in the global middleware chain. It regex-strips `<...>` from every JSON string value, corrupting passwords and other data.

**Change:**
- Remove `r.Use(appmw.InputSanitizer)` from `main.go`.
- Keep `sanitize.go` and its `SanitizeString` function for explicit per-field use (e.g., review text).
- The `InputSanitizer` HTTP middleware function remains in the file but is no longer wired globally.
- Existing call sites that explicitly sanitize (review text) continue to work.

### 6. Rate Limiter Readability

**Current state:** `Window: 60_000_000_000` with a comment explaining it's nanoseconds.

**Change:** Replace with `Window: time.Minute` in both rate limiter instantiations in `main.go`.

## Data Models

No new data models. The migration runner uses the existing `db/migrations/*.sql` files and the existing PostgreSQL connection pool.

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system -- essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: CORS allows only allowlisted origins (dev mode)

*For any* HTTP request with an arbitrary `Origin` header value, when `APP_ENV=development`, the CORS middleware SHALL set `Access-Control-Allow-Origin` if and only if the origin is in the fixed development allowlist.

**Validates: Requirements 4.1, 4.2**

### Property 2: Credentials header implies origin header

*For any* HTTP request processed by the CORS middleware (in any mode), `Access-Control-Allow-Credentials: true` is present in the response if and only if `Access-Control-Allow-Origin` is also present.

**Validates: Requirements 4.3**

### Property 3: Production CORS allows only configured origin

*For any* HTTP request with an arbitrary `Origin` header value, when `APP_ENV` is not `development`, the CORS middleware SHALL set `Access-Control-Allow-Origin` only when the origin exactly equals the configured `FRONTEND_ORIGIN`.

**Validates: Requirements 4.4**

### Property 4: Request body passthrough preserves all characters

*For any* valid JSON string value (including those containing angle-bracket characters), when the `InputSanitizer` middleware is not in the middleware chain, the request body arrives at the handler byte-for-byte identical to what the client sent.

**Validates: Requirements 5.2**

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Migration failure on startup | Log error with migration file name, call `log.Fatalf` (non-zero exit) |
| Database unreachable | Existing behavior: `log.Fatalf` on pool creation failure |
| Shutdown timeout exceeded | Log warning, force-close connections, exit 0 |
| CORS: disallowed origin | Omit CORS headers; request proceeds (browser enforces) |

## Testing Strategy

### Unit Tests

- **CORS middleware tests** (`internal/middleware/cors_test.go`):
  - Dev mode: verify each allowlisted origin gets headers; verify a random external origin does not.
  - Production mode: verify only `FRONTEND_ORIGIN` gets headers.
  - Verify credentials header invariant.
  
- **Sanitize function** (`internal/middleware/sanitize_test.go`):
  - Verify `SanitizeString` still works for explicit use.
  - Verify angle-bracket strings pass through when middleware is not applied.

- **Graceful shutdown** (integration-style in `cmd/server`):
  - Verify `http.Server.Shutdown` is called with correct timeout context.

### Property-Based Tests

Property-based testing applies to the CORS middleware (properties 1-3) and the body passthrough (property 4). The project already uses `pgregory.net/rapid` for PBT.

- **Library:** `pgregory.net/rapid` (already in go.mod)
- **Minimum iterations:** 100 per property
- **Tag format:** `Feature: ops-deploy-fixes, Property N: <title>`

### Integration Tests (CI)

- Docker build succeeds in CI (existing `ci.yml` pipeline — add `docker build` step).
- Fresh database with migrations applied via server startup.

