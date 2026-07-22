# 🥐 Mie & Beurre — App Documentation

> A bakery ordering & reservation platform connecting artisan bakers with hungry customers.

---

## 📚 Pages

| # | Page | Description |
|---|------|-------------|
| 01 | [Getting Started](./01-getting-started.md) | Setup & run instructions |
| 02 | [Architecture](./02-architecture.md) | System architecture overview |
| 03 | [Frontend](./03-frontend.md) | Frontend structure & components |
| 04 | [Backend](./04-backend.md) | Backend structure & API |
| 05 | [Database](./05-database.md) | Database schema & migrations |
| 06 | [Customer Features](./06-features-customer.md) | Customer features guide |
| 07 | [Baker Portal](./07-features-baker.md) | Baker portal features guide |
| 08 | [Auth & Security](./08-auth-and-security.md) | Authentication, roles, security |
| 09 | [Internationalization](./09-i18n.md) | i18n support (EN, FR, NL) |
| 10 | [Testing](./10-testing.md) | Testing strategy |
| 11 | [Design System](./11-design-system.md) | Design tokens, themes, CSS |
| 12 | [API Reference](./12-api-reference.md) | API endpoints reference |
| 13 | [Deployment](./13-deployment.md) | Deployment notes & env vars |
| 14 | [Roadmap](./14-roadmap.md) | What's implemented vs TODO |

---

## ⚡ Quick Stats

| Area | Details |
|------|---------|
| **Backend** | Go 1.26 · chi router · JWT auth |
| **Frontend** | React 19 · TypeScript · Vite 8 |
| **Database** | PostgreSQL (schema ready) · In-memory repos for dev |
| **Testing** | Go unit + property tests · Vitest + fast-check · Playwright E2E |
| **Languages** | English, French, Dutch (Flemish) |
| **Maps** | Leaflet + react-leaflet |
| **Charts** | Recharts (baker dashboard) |

---

## 🏪 What is Mie & Beurre?

Mie & Beurre is a full-stack bakery ordering platform that allows:

- **Customers** to browse bakeries, place delivery orders (paid online), make pickup reservations (paid on-spot), and set up recurring weekly/bi-weekly orders.
- **Bakers** to manage their bakery profile, products (with allergens & health scores), opening hours, and process incoming orders/reservations through a status lifecycle.

---

*Last updated: July 2025*
