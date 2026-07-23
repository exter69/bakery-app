# Implementation Plan: Ops/Deploy Fixes

## Overview

This plan implements six operational fixes: Docker Go version pinning, automated migrations on deploy, graceful HTTP shutdown, CORS hardening, sanitizer middleware removal, and rate-limiter readability. All changes are in Go and Dockerfile; no frontend changes required.

## Tasks

- [x] 1. Pin Dockerfile Go version to match go.mod
  - [x] 1.1 Update `Dockerfile` FROM line from `golang:1.22-alpine` to `golang:1.26-alpine`
    - Match the major.minor version declared in `go.mod` (`go 1.26.5`)
    - _Requirements: 1.1, 1.2_
  - [x] 1.2 Add a Docker build step to `.github/workflows/ci.yml`
    - Add a job or step that runs `docker build .` to catch image build failures
    - _Requirements: 1.3_

- [x] 2. Implement automated database migrations on server startup
  - [x] 2.1 Add `github.com/pressly/goose/v3` and `github.com/jackc/pgx/v5/stdlib` to `go.mod`
    - Run `go get github.com/pressly/goose/v3 github.com/jackc/pgx/v5/stdlib`
    - _Requirements: 2.1_
  - [x] 2.2 Create migration runner function in `cmd/server/main.go`
    - Add a `runMigrations(pool *pgxpool.Pool) error` function that uses `goose.Up` with dialect "postgres" and dir "db/migrations"
    - Use `stdlib.OpenDBFromPool(pool)` to get a `*sql.DB` from the pgx pool
    - _Requirements: 2.1, 2.2, 2.3_
  - [x] 2.3 Call `runMigrations` after pool creation, before HTTP listener
    - Only call when `databaseURL != ""` (skip in in-memory mode)
    - On error: `log.Fatalf("Migration failed: %v", err)`
    - On success: `log.Println("Database migrations applied")`
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [x] 3. Implement graceful HTTP shutdown
  - [x] 3.1 Replace `http.ListenAndServe` with `http.Server` + goroutine pattern in `cmd/server/main.go`
    - Create `srv := &http.Server{Addr: ":"+port, Handler: r}`
    - Start `srv.ListenAndServe()` in a goroutine
    - After `<-ctx.Done()`: log shutdown start, create 15s timeout context, call `srv.Shutdown(shutdownCtx)`, log completion
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 4. Checkpoint
  - Ensure `go build ./...` and `go vet ./...` pass. Verify the server starts and shuts down cleanly. Ask the user if questions arise.

- [x] 5. Harden CORS middleware
  - [x] 5.1 Refactor `internal/middleware/cors.go` to use a fixed dev origin allowlist
    - Add `var devAllowedOrigins = map[string]bool{...}` with localhost:5173, localhost:3000, 127.0.0.1:5173, 127.0.0.1:3000
    - In dev mode: check origin against map instead of reflecting
    - Only set `Access-Control-Allow-Credentials: true` when `Access-Control-Allow-Origin` is also being set
    - In production: only match `cfg.AllowedOrigin` exactly
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  - [ ]* 5.2 Write property tests for CORS middleware (`internal/middleware/cors_property_test.go`)
    - **Property 1: CORS allows only allowlisted origins (dev mode)**
    - **Property 2: Credentials header implies origin header**
    - **Property 3: Production CORS allows only configured origin**
    - Use `pgregory.net/rapid` with minimum 100 iterations per property
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4**
  - [ ]* 5.3 Write unit tests for CORS middleware (`internal/middleware/cors_test.go`)
    - Test each dev-allowed origin gets headers
    - Test arbitrary external origin does not get headers
    - Test production mode only allows configured origin
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 6. Remove global JSON sanitizer middleware
  - [x] 6.1 Remove `r.Use(appmw.InputSanitizer)` from `cmd/server/main.go`
    - Keep the `sanitize.go` file and `SanitizeString` function intact for explicit per-field use
    - _Requirements: 5.1, 5.3, 5.4_
  - [ ]* 6.2 Write a test verifying JSON bodies with angle brackets pass through unchanged
    - **Property 4: Request body passthrough preserves all characters**
    - Use `pgregory.net/rapid` to generate strings containing `<` and `>`, send as JSON, verify handler receives them unchanged
    - **Validates: Requirements 5.2**
  - [x] 6.3 Verify existing explicit sanitization call sites still work
    - Confirm review handler still calls `SanitizeString` on review text body
    - _Requirements: 5.4_

- [x] 7. Fix rate limiter readability
  - [x] 7.1 Replace `60_000_000_000` with `time.Minute` in both rate limiter configs in `cmd/server/main.go`
    - Remove the now-unnecessary nanosecond comments
    - _Requirements: 6.1_

- [x] 8. Final checkpoint
  - Run `go build ./...`, `go vet ./...`, and `go test ./...`. Ensure all pass. Verify Docker build succeeds locally if possible. Ask the user if questions arise.

## Notes

- The project already has `pgregory.net/rapid` in go.mod for property-based testing.
- `goose` is already used in CI (installed via `go install`); adding it as a library dependency makes the server self-contained for migrations.
- The `InputSanitizer` middleware function is kept in `sanitize.go` for backward compatibility but is no longer in the global chain.
- Tasks marked with `*` are optional test tasks and can be skipped for faster delivery.
- No version bump (pre-1.0 per release-versioning steering).

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "2.1"] },
    { "id": 1, "tasks": ["2.2", "2.3"] },
    { "id": 2, "tasks": ["3.1"] },
    { "id": 3, "tasks": ["4"] },
    { "id": 4, "tasks": ["5.1", "6.1", "7.1"] },
    { "id": 5, "tasks": ["5.2", "5.3", "6.2", "6.3"] },
    { "id": 6, "tasks": ["8"] }
  ]
}
```

