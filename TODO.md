# Development TODO

## Authentication & User Roles

- [x] 1. Add `users` table with columns: id (UUID), username, password_hash, role (int), created_at
- [x] 2. Implement `POST /api/auth/register` endpoint (username, password, role)
- [x] 3. Implement `POST /api/auth/login` endpoint (returns JWT with user ID + role in claims)
- [x] 4. Update JWT middleware to include role in token claims and expose `GetUserRoleFromContext()`

## Login Page Redesign

- [x] 5. Redesign login page (background image, username/password, sign in, register, guest visit)
- [x] 6. Create registration page (username, password, confirm password, role selector)
- [x] 7. Common login for both customer and client — role determines redirect

## Guest / Anonymous Access

- [x] 8. Allow browsing bakeries and menus without authentication (public endpoints)
- [x] 9. Frontend: "Visit without account" → guest mode (read-only, no orders)

## Role-Based Routing

- [x] 10. After login, redirect based on role
- [x] 11. Protect seller routes — only role 0/1 can access `/dashboard/*`
- [x] 12. Protect admin routes — only role 0 can access `/admin/*`

## Client/Seller Portal (New)

- [x] 13. Create seller dashboard layout (sidebar navigation)
- [x] 14. Seller: manage bakery info (name, description, address, photo) — frontend form
- [x] 15. Seller: manage products (CRUD — add/edit/remove products, set availability) — frontend
- [x] 16. Seller: manage schedule (set open/close times per day) — frontend
- [x] 17. Seller: view incoming orders (table with status, items) — frontend
- [x] 18. Seller: view incoming reservations — frontend
- [x] 19. Seller: update order status (confirm, preparing, ready, delivered) — frontend
- [x] 20. Seller: update reservation status (ready, picked up) — frontend

## Backend Endpoints for Seller Portal

- [x] 21. `PUT /api/bakeries/:id` — update bakery info (owner only)
- [x] 22. `POST /api/bakeries/:id/products` — add product
- [x] 23. `PUT /api/products/:id` — update product
- [x] 24. `DELETE /api/products/:id` — remove product
- [x] 25. `PUT /api/bakeries/:id/schedule` — update schedule
- [x] 26. `GET /api/bakeries/:id/orders` — list orders for a bakery (seller view)
- [x] 27. `GET /api/bakeries/:id/reservations` — list reservations for a bakery
- [x] 28. `PUT /api/orders/:id/status` — update order status (seller only)
- [x] 29. `PUT /api/reservations/:id/status` — update reservation status (seller only)

## Status

- Backend: COMPLETE ✓
- Frontend: COMPLETE ✓
- All 29 tasks done
