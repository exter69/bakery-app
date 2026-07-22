# 📡 API Reference

## Base URL

```
http://localhost:8080/api
```

All endpoints prefixed with `/api`. Protected endpoints require `Authorization: Bearer <jwt>` header.

---

## Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/login` | No | Login, returns JWT |
| POST | `/auth/register` | No | Register new user |
| POST | `/auth/request-access` | No | Baker requests registration token |
| POST | `/admin/tokens` | Admin | Generate registration token |

### POST /auth/login

```json
// Request
{ "username": "alice", "password": "customer123" }

// Response 200
{ "token": "eyJhbG...", "user": { "id": "...", "username": "alice", "role": 2 } }
```

### POST /auth/register

```json
// Request
{ "username": "newuser", "password": "pass123", "role": "customer" }
// For bakers, include token:
{ "username": "newbaker", "password": "pass123", "role": "seller", "token": "DEMO1234" }
```

---

## Bakeries (Public)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/bakeries` | List all bakeries |
| GET | `/bakeries/:id` | Get bakery details |
| GET | `/bakeries/:id/menu` | Get bakery menu |

### GET /bakeries

Query params: `?lat=48.87&lng=2.33` (optional, for distance sorting)

```json
// Response 200
{
  "items": [
    { "id": "bakery-1", "name": "La Boulangerie du Coin", "distance": 1.2, ... }
  ]
}
```

---

## Orders (Protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/orders` | Create order (rate limited) |
| GET | `/orders` | List user's orders |
| DELETE | `/orders/:id` | Cancel order |

### POST /orders

```json
// Request
{
  "bakeryId": "bakery-1",
  "items": [
    { "productId": "prod-1-1", "productName": "Croissant", "quantity": 3 }
  ],
  "scheduledDay": "wednesday",
  "scheduledTime": { "startTime": { "hour": 8, "minute": 0 }, "endTime": { "hour": 8, "minute": 30 } },
  "paymentMethod": "online",
  "selectionMode": "fixed"
}
```

---

## Reservations (Protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/reservations` | Create reservation (rate limited) |
| DELETE | `/reservations/:id` | Cancel reservation |

---

## Seller Portal (Seller/Admin)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/seller/bakery` | Get own bakery |
| PUT | `/bakeries/:id` | Update bakery info |
| POST | `/bakeries/:id/products` | Create product |
| PUT | `/bakeries/:id/products/:pid` | Update product |
| DELETE | `/bakeries/:id/products/:pid` | Delete product |
| GET | `/seller/orders` | List bakery's orders |
| PUT | `/seller/orders/:id/status` | Update order status |
| GET | `/seller/reservations` | List bakery's reservations |
| PUT | `/seller/reservations/:id/status` | Update reservation status |

### PUT /seller/orders/:id/status

```json
// Request
{ "status": "preparing" }
// Valid transitions: confirmed → preparing → ready → delivered
```

---

## Recurring Orders (Protected)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/recurring-orders` | List user's recurring orders |
| POST | `/recurring-orders` | Create recurring order |
| PUT | `/recurring-orders/:id` | Update recurring order |
| DELETE | `/recurring-orders/:id` | Delete recurring order |

---

## User Profile (Protected)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/user/profile` | Get current user profile |
| PUT | `/user/holiday` | Update holiday mode |
| GET | `/user/favorites` | Get favorite products |
| PUT | `/user/favorites` | Update favorite products |

---

## Payments

| Method | Path | Description |
|--------|------|-------------|
| POST | `/payments/callback` | Payment gateway callback |

---

## Common Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Validation error |
| 401 | Unauthorized (missing/invalid JWT) |
| 403 | Forbidden (wrong role) |
| 404 | Not found |
| 429 | Rate limit exceeded |
| 500 | Server error |
