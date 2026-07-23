# OWASP Top 10 (2021) — Security Checklist

Status legend: [x] addressed, [ ] not yet addressed, [~] partially addressed

## A01: Broken Access Control

- [x] Role-Based Access Control (RBAC) — JWT claims carry role; middleware enforces per-route
- [x] Ownership checks on all seller mutations (SellerService.verifyBakeryOwnership)
- [x] Bundle publish/create verifies seller owns the associated bakery
- [x] Order/reservation delete verifies user owns the resource
- [x] Admin endpoints check role == 0 before proceeding
- [x] No wildcard CORS — restricted to FRONTEND_ORIGIN

## A02: Cryptographic Failures

- [x] Passwords hashed with bcrypt (cost 10)
- [x] JWT signed with HMAC-SHA256; secret loaded from env var
- [x] No secrets logged (redacted token from registration log)
- [x] `.env` file in `.gitignore`
- [x] HSTS header enforced in production (non-localhost)

## A03: Injection

- [x] All database queries use parameterized statements (pgx positional params)
- [x] HTML-stripping input sanitizer middleware on all inbound requests
- [x] No string concatenation in SQL anywhere in the repository layer

## A04: Insecure Design

- [x] Rate limiting on auth endpoints (5 req/min per IP)
- [x] Rate limiting on order/reservation creation (10 req/min per user)
- [x] Request body size limit (1 MB default, prevents denial-of-service via oversized payloads)
- [x] Input validation at both handler (DTO) and service (domain) layers
- [x] Token-gated baker registration (prevents unauthorized seller account creation)

## A05: Security Misconfiguration

- [x] Security headers middleware (CSP, X-Frame-Options, X-Content-Type-Options, etc.)
- [x] CORS restricted to single frontend origin; credentials allowed
- [x] Permissions-Policy disables camera/microphone; geolocation self-only
- [x] Environment-based configuration (no hardcoded secrets)
- [x] Default JWT secret triggers a warning log in dev mode

## A06: Vulnerable and Outdated Components

- [ ] Tracked in MA-32 (dependency audit and update)
- [~] go.mod pins exact versions for all dependencies

## A07: Identification and Authentication Failures

- [x] JWT expiry enforced (middleware rejects expired tokens)
- [x] Generic "invalid username or password" error (prevents username enumeration)
- [x] IP-based rate limiting on login endpoint (brute-force mitigation)
- [x] Bcrypt with cost factor 10 (resistant to offline attacks)
- [x] Registration token has expiry and single-use enforcement

## A08: Software and Data Integrity Failures

- [x] Stripe webhook signature verification (STRIPE_WEBHOOK_SECRET)
- [x] JWT signature validated on every protected request
- [~] No CI/CD pipeline integrity checks yet (out of scope for app layer)

## A09: Security Logging and Monitoring Failures

- [~] Tracked in MA-22/MA-23 (structured logging and monitoring)
- [x] Chi Logger middleware records every request (method, path, status, duration)
- [x] No sensitive data in logs (tokens redacted)
- [ ] Alerting on repeated auth failures (future work)

## A10: Server-Side Request Forgery (SSRF)

- [x] No outbound HTTP requests constructed from user input
- [x] Only outbound calls are to Stripe API and SMTP server (both from trusted config)
- [x] WebSocket connections are inbound-only; no user-controlled URL fetching
