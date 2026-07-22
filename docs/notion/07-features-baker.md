# 🏪 Baker Portal

## Registration Flow

```
Baker → "Request Access" form → Admin receives email
Admin → generates registration token (CLI or API)
Baker → enters token code at /register → account created (role: Seller)
```

- Registration tokens have expiry dates
- Demo token for testing: `DEMO1234`
- Token is single-use

---

## Dashboard Overview

Route: `/dashboard`

- Summary statistics (total orders, revenue, active products)
- Charts (Recharts): orders over time, revenue breakdown
- Quick status overview of pending orders/reservations

---

## Manage Bakery

Route: `/dashboard/bakery`

Edit bakery profile:
- Name
- Description
- Address
- Photo URL
- Google Place ID (for map integration)
- Minimum delivery amount

---

## Manage Schedule

Route: `/dashboard/schedule`

- Set opening hours per day of week
- Toggle days open/closed
- Google Maps mode (derive hours from Google Place data)

---

## Manage Products

Route: `/dashboard/products`

Full CRUD for products:

| Field | Type | Notes |
|-------|------|-------|
| Name | string | Required |
| Description | string | Optional |
| Price | cents | Integer, displayed as €X.XX |
| Photo URL | string | Image link |
| Category | string | Breads, Viennoiseries, Pastries, etc. |
| Available | boolean | Toggle on/off |
| Allergens | string[] | Pick from 14 standard allergens |
| Health Score | 1-5 | Optional nutritional indicator |

---

## Process Orders

Route: `/dashboard/orders`

Order status lifecycle:

```
Confirmed → Preparing → Ready → Delivered
```

- Baker advances status through each step
- View order details: customer, items, scheduled time, total
- Filter by status

---

## Process Reservations

Route: `/dashboard/reservations`

Reservation status lifecycle:

```
Confirmed → Ready → Picked Up
```

- Simpler than orders (no delivery step)
- Payment collected at pickup

---

## Seller API Endpoints

All seller endpoints require JWT with role 0 (Admin) or 1 (Seller):

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/seller/bakery` | Get own bakery |
| PUT | `/api/bakeries/:id` | Update bakery |
| POST | `/api/bakeries/:id/products` | Create product |
| PUT | `/api/bakeries/:id/products/:pid` | Update product |
| DELETE | `/api/bakeries/:id/products/:pid` | Delete product |
| GET | `/api/seller/orders` | List bakery orders |
| PUT | `/api/seller/orders/:id/status` | Update order status |
| GET | `/api/seller/reservations` | List bakery reservations |
| PUT | `/api/seller/reservations/:id/status` | Update reservation status |
