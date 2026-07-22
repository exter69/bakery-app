# 🗄️ Database

## Overview

The project uses **PostgreSQL** as the target database. Schema is defined via 13 goose migrations. At runtime, the app currently uses **in-memory repositories** — PostgreSQL repos are ready to implement using the existing schema.

---

## Migrations

| # | File | Description |
|---|------|-------------|
| 001 | `create_bakeries.sql` | bakeries + day_schedules tables |
| 002 | `create_products.sql` | products table |
| 003 | `create_orders.sql` | orders table |
| 004 | `create_reservations.sql` | reservations table |
| 005 | `create_order_items.sql` | order_items table |
| 006 | `create_users.sql` | users table |
| 007 | `add_bakery_location.sql` | lat/lng + google_place_id columns |
| 008 | `create_recurring_orders.sql` | recurring_orders table |
| 009 | `add_user_holiday.sql` | holiday_mode, holiday_from, holiday_to |
| 010 | `add_user_favorites.sql` | favorite_products array |
| 011 | `create_registration_tokens.sql` | registration_tokens table |
| 012 | `add_health_allergens.sql` | allergens[], health_score on products |
| 013 | `add_bakery_owner.sql` | owner_id on bakeries |

---

## ER Diagram (Simplified)

```
┌──────────────────┐       ┌──────────────────┐
│     users        │       │    bakeries      │
├──────────────────┤       ├──────────────────┤
│ id (PK)          │◄──┐   │ id (PK)          │
│ username         │   │   │ owner_id (FK)    │──► users
│ password_hash    │   │   │ name             │
│ role             │   │   │ address          │
│ holiday_mode     │   │   │ latitude         │
│ favorite_products│   │   │ longitude        │
└──────────────────┘   │   │ min_delivery     │
                       │   └──────────────────┘
                       │          │
                       │          ▼
                       │   ┌──────────────────┐
                       │   │  day_schedules   │
                       │   ├──────────────────┤
                       │   │ bakery_id (FK)   │
                       │   │ day_of_week      │
                       │   │ open_time        │
                       │   │ close_time       │
                       │   │ is_open          │
                       │   └──────────────────┘
                       │
                       │   ┌──────────────────┐
                       │   │    products      │
                       │   ├──────────────────┤
                       │   │ id (PK)          │
                       │   │ bakery_id (FK)   │──► bakeries
                       │   │ name             │
                       │   │ price            │
                       │   │ category         │
                       │   │ allergens[]      │
                       │   │ health_score     │
                       │   └──────────────────┘
                       │
    ┌──────────────────┤   ┌──────────────────┐
    │    orders        │   │  reservations    │
    ├──────────────────┤   ├──────────────────┤
    │ id (PK)          │   │ id (PK)          │
    │ bakery_id (FK)   │   │ bakery_id (FK)   │
    │ user_id (FK)  ───┤   │ user_id (FK)  ───┘
    │ scheduled_day    │   │ scheduled_day    │
    │ scheduled_time   │   │ scheduled_time   │
    │ status           │   │ status           │
    │ total_amount     │   │ total_amount     │
    │ payment_method   │   │ payment_method   │
    └──────────────────┘   └──────────────────┘
            │
            ▼
    ┌──────────────────┐
    │   order_items    │
    ├──────────────────┤
    │ order_id (FK)    │
    │ product_id (FK)  │
    │ quantity         │
    │ unit_price       │
    │ subtotal         │
    └──────────────────┘
```

---

## Migration Commands

Requires [goose](https://github.com/pressly/goose):

```bash
# Apply all migrations
make migrate-up

# Rollback one migration
make migrate-down

# Check status
make migrate-status

# Reset all
make migrate-reset
```

Default DSN: `postgres://localhost:5432/bakery_app?sslmode=disable`

Override with: `DB_DSN=<your-dsn> make migrate-up`

---

## Key Constraints

- `day_schedules`: UNIQUE(bakery_id, day_of_week), CHECK(open < close when open)
- `products.price`: stored in cents (integer)
- `orders.total_amount`: stored in cents
- `users.role`: integer enum (0=Admin, 1=Seller, 2=Customer)
