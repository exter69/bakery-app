# 🔐 Auth & Security

## Authentication

- **JWT-based** authentication via `Authorization: Bearer <token>` header
- Tokens issued on login via `POST /api/auth/login`
- Signed with `JWT_SECRET` environment variable
- Library: `github.com/golang-jwt/jwt/v5`

---

## User Roles

| Role | Value | Access |
|------|-------|--------|
| Admin | 0 | Full access, token generation |
| Seller | 1 | Baker portal, own bakery management |
| Customer | 2 | Browsing, ordering, reservations |

Role is embedded in the JWT payload and validated server-side.

---

## Registration Token System

Bakers cannot self-register without a token:

1. Baker fills "Request Access" form (name, email, bakery name)
2. Admin generates a token via `POST /api/admin/tokens` or CLI (`cmd/gentoken`)
3. Token is sent to baker (email or manual)
4. Baker registers at `/register` with the token code
5. Token is single-use and has an expiry date

---

## Rate Limiting

| Endpoint | Limit | Key |
|----------|-------|-----|
| `POST /api/auth/login` | 5 req/min | IP address |
| `POST /api/orders` | 10 req/min | User ID |
| `POST /api/reservations` | 10 req/min | User ID |

Implementation: sliding window counter in `internal/middleware/ratelimit.go`

---

## Input Sanitization

- Global middleware strips HTML tags from all request bodies
- Prevents XSS via stored data
- Located in `internal/middleware/sanitize.go`

---

## Frontend Route Protection

| Component | Purpose |
|-----------|---------|
| `ProtectedRoute` | Redirects to `/login` if no JWT present |
| `RoleRoute` | Restricts access by role (e.g., seller dashboard) |

```tsx
// Usage example
<Route path="/schedule" element={
  <ProtectedRoute><ScheduleOrdersPage /></ProtectedRoute>
} />

<Route path="/dashboard" element={
  <RoleRoute allowedRoles={[0, 1]}><DashboardLayout /></RoleRoute>
} />
```

---

## Password Hashing

- Algorithm: **bcrypt** (default cost)
- Library: `golang.org/x/crypto/bcrypt`
- Passwords never stored or transmitted in plain text
- `PasswordHash` field excluded from JSON serialization (`json:"-"`)

---

## CORS Configuration

```go
cors.Options{
    AllowedOrigins:   []string{frontendOrigin},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           300,
}
```

---

## Security Checklist

- [x] JWT authentication
- [x] Role-based access control
- [x] Rate limiting (brute-force + abuse prevention)
- [x] Input sanitization (HTML stripping)
- [x] CORS restricted to frontend origin
- [x] Password hashing (bcrypt)
- [x] Token-based baker registration
- [ ] HTTPS (deployment concern)
- [ ] CSRF protection (not needed — JWT in header, not cookies)
