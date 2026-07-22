# 🗺️ Roadmap

## ✅ Implemented

### Core Features
- [x] Bakery browsing (list + map view)
- [x] Geolocation sorting by distance
- [x] Bakery detail page (schedule, menu, description)
- [x] Delivery orders (online payment)
- [x] Pickup reservations (on-spot payment)
- [x] Recurring orders (weekly / bi-weekly)
- [x] Selection modes (fixed, bakery choice, random favorites)
- [x] Order schedule management (view, cancel)
- [x] Holiday mode (pause recurring orders)
- [x] Favorite products

### Baker Portal
- [x] Token-based registration flow
- [x] Dashboard with stats and charts
- [x] Bakery profile management
- [x] Product CRUD (with categories)
- [x] Allergen management (14 EU allergens)
- [x] Health score per product (1-5)
- [x] Schedule/opening hours management
- [x] Order status lifecycle (Confirmed → Preparing → Ready → Delivered)
- [x] Reservation status lifecycle (Confirmed → Ready → Picked Up)

### Platform
- [x] JWT authentication + role-based access
- [x] Internationalization (EN, FR, NL)
- [x] Rate limiting (orders + auth)
- [x] Input sanitization
- [x] Guest browsing mode
- [x] Responsive design (mobile, tablet, desktop)
- [x] E2E test suite (Playwright, 11 specs)
- [x] Property-based tests (Go rapid + fast-check)
- [x] PostgreSQL schema (migrations ready)
- [x] Interactive map (Leaflet)

---

## 🚧 In Progress

- Nothing currently in active development

---

## 📋 Planned

### Payments & Transactions
- [ ] Stripe payment gateway integration
- [ ] Payment confirmation emails
- [ ] Invoice generation

### Notifications
- [ ] Real-time notifications (WebSocket)
- [ ] Email notifications (order confirmation, status change, baker alerts)
- [ ] Push notifications (PWA)

### Infrastructure
- [ ] PostgreSQL repository implementation (replace memory repos)
- [ ] Docker / docker-compose for local dev
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Automated deployment

### Features
- [ ] Delivery tracking & logistics
- [ ] Image upload (replace URL input with file upload + storage)
- [ ] Product search and filtering
- [ ] Customer reviews and ratings
- [ ] Dark mode
- [ ] Order history with re-order button
- [ ] Loyalty program / points system
- [ ] Multi-bakery order (cart spanning bakeries)

### DX & Quality
- [ ] OpenAPI/Swagger specification
- [ ] Storybook for component library
- [ ] Visual regression testing
- [ ] Performance monitoring (APM)
- [ ] Error tracking (Sentry)

---

## 💡 Ideas (Backlog)

- Bakery analytics dashboard (customer demographics, peak hours)
- Subscription boxes (curated weekly selection)
- B2B ordering (offices, cafes)
- Bakery-to-bakery product sharing
- Carbon footprint tracking per delivery
