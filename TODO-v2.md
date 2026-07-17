# TODO v2 — Customer Experience & Feature Roadmap

## Navigation & Layout

- [x] 1. Top navigation bar (Home, Bakeries, About, auth buttons, hero strip)
- [x] 2. Navbar persists across all customer pages (CustomerLayout with Outlet)

## Home Page

- [x] 3. Redesign home page (hero, how it works, delivery/reservation sections, nearby bakeries)

## Bakeries Page

- [x] 4. Friendly empty state message
- [x] 5. Location-based sorting (geolocation, haversine distance, radius filter, distance on cards)

## Recurring Delivery System

- [x] 6. "Recurring order" option with frequency selection
- [x] 7. Backend: recurring_orders table, service, handler, API endpoints
- [x] 8. Frontend: recurring orders page with pause/resume/delete

## Holiday Mode

- [x] 9. Holiday mode toggle with date range (backend + frontend)

## Random Items Feature

- [x] 10. "Surprise me" option: bakery_choice and random_favorites modes
- [x] 11. Backend: selection_mode on orders, favorite_products on user, random picking logic

## Product Classification System (Bonus — Later)

- [ ] 12. Product "weight unit" system (portion_units decimal field)
- [ ] 13. Statistics/recommendations during selection (people count, health rating, balanced suggestions)
- [ ] 14. Backend classification endpoint (portion_units, health_score, taste_score)

## Priority

**Now:** 1 → 2 → 3 → 4 → 5
**Next sprint:** 6 → 7 → 8 → 9
**Then:** 10 → 11
**Bonus/Later:** 12 → 13 → 14
