# Code Review — Improvement Recommendations

A full-stack audit of the bakery ordering app. Organized by priority and area.

---

## 1. Backend (Go)

### Architecture

| Issue | File | Recommendation |
|-------|------|----------------|
| `GetMyBakery` fetches ALL bakeries then filters | `seller_service.go:103` | Add `GetBakeryByOwnerID(ctx, ownerID)` to the repository interface — O(1) instead of O(n) scan |
| Product IDs use `time.Now().UnixNano()` | `seller_service.go:143` | Use UUID generation (`uuid.New().String()`) for collision safety and consistency with DB schema |
| In-memory repo uses `map[string][]Product` keyed by bakeryID | `bakery_repo.go` | `GetProductByID` does O(bakeries × products) scan. Add a secondary index `map[string]*Product` keyed by product ID |
| `UpdateProduct` accepts `map[string]interface{}` | `seller_service.go:155` | Type-unsafe. Replace with a typed `ProductUpdateInput` struct using pointer fields for "was set" semantics (like `UpdateBakery` already does) |
| `handleSellerError` uses `errors.Is` chain | `seller_handler.go` | Fine for now but will become unwieldy. Consider a typed `AppError` with an HTTP status field |
| No request body size limit | `main.go` | Add `chimw.AllowContentType("application/json")` and a body size limiter (`http.MaxBytesReader`) to prevent abuse |
| JWT secret fallback to hardcoded value in dev | `main.go:30` | Log a WARNING when using the fallback so it's obvious in production |

### Readability

| Issue | File | Recommendation |
|-------|------|----------------|
| `seller_handler.go` is 500+ lines | — | Split into `seller_product_handler.go`, `seller_order_handler.go`, `seller_bakery_handler.go` |
| Duplicated ownership check pattern (fetch bakery → check owner) | `seller_service.go` | Extract `verifyBakeryOwnership(ctx, bakeryID, ownerID) (*Bakery, error)` helper |
| `seed.go` is 200+ lines of hardcoded data | — | Move seed data to a JSON/YAML file loaded at runtime, or use a table-driven approach |
| Custom `MarshalJSON` on Product for allergens nil-safety | `models.go` | Fragile — if anyone adds a field they must remember the alias trick. Consider an `EnsureDefaults()` method called before repo persist instead |

### Security

| Issue | Severity | Recommendation |
|-------|----------|----------------|
| Passwords hardcoded in seed (`admin123`, `baker123`, `customer123`) | Low (dev only) | Fine for dev, but add a comment warning not to use in production; or read from env/file |
| No HTTPS enforcement | Medium | Add a middleware that redirects HTTP→HTTPS in production (behind a flag/env) |
| No rate limiting on login endpoint | Medium | Add rate limiting to `/api/auth/login` (brute-force protection) |
| `decodeTokenRole` in `client.ts` doesn't verify signature | Low | It's only for UI routing, not auth (backend verifies JWT), but document this clearly |

---

## 2. Frontend (React + TypeScript)

### Component Size & Splitting

| Issue | File | Recommendation |
|-------|------|----------------|
| `ProductSelectionModal.tsx` is ~400 lines | — | Extract `SetupContent` into its own file (`OrderSetupPanel.tsx`). Extract the product card into `ProductCard.tsx` |
| `BakeryDetailPage.tsx` is ~280 lines | — | Extract the Haversine calculation into a `utils/geo.ts` module. Extract the product grid/row rendering into a `ProductGrid` component |
| `DashboardProducts.tsx` mixes table + modal + form | — | Extract the form into `ProductFormModal.tsx` and the table into `ProductsTable.tsx` |

### Performance

| Issue | Impact | Recommendation |
|-------|--------|----------------|
| `getQuantity` uses `.find()` on every render for every card | O(items × products) | Use a `Map<productId, quantity>` via `useMemo` instead of finding on each call |
| `ProductSelectionModal` re-renders all cards when any quantity changes | Re-render 50+ cards | Wrap individual card in `React.memo` with a custom comparator or move card to its own component |
| `translations.ts` is a 300+ line monolithic object loaded at startup | Bundle size | Split into `translations/en.ts`, `translations/fr.ts`, `translations/nl.ts` and lazy-load via dynamic `import()` |
| Haversine calculation runs on every render (in `useMemo` depending on `bakery`) | Negligible | Already memoized — fine |
| `overflow: visible` on `.psm__card` defeats the image `border-radius` clipping | Visual | Added `border-radius` to `.psm__card-image-wrap` — verify this looks right on all viewports |

### State Management

| Issue | Recommendation |
|-------|----------------|
| `orderItems` stored as `OrderItem[]` (contains full Product object) | Store just `{ productId, quantity }` and look up product from the `products` prop. Reduces state size and avoids stale product data |
| Multiple `useEffect` to trim days on mode change creates potential infinite loops | The `useEffect` for day trimming depends on `selectedDays` which it also sets. Add a guard or use a ref to break the cycle |
| `isAuthenticated()` is a function call (reads localStorage) in render | Memoize in context or use a `useAuth()` hook with state that syncs with storage events |

### Type Safety

| Issue | Recommendation |
|-------|----------------|
| Day names are `string` everywhere | Create a `type DayOfWeek = 'monday' | 'tuesday' | ...` union and use it in schedule types |
| `Product.allergens` typed as `string[]` | Use `type Allergen = 'gluten' | 'crustaceans' | ...` for compile-time safety |
| API responses are cast with `as Promise<T>` without runtime validation | Add a lightweight runtime check or use zod/valibot for response parsing on critical paths |
| `createProduct` parameter is an inline object type | Define a `CreateProductRequest` interface in `seller.ts` for reuse |

### Accessibility

| Issue | Recommendation |
|-------|----------------|
| `.psm-backdrop` has `aria-hidden="true"` but wraps the dialog | The backdrop click handler is fine but `aria-hidden` on a visible interactive container is incorrect. Remove it or move the dialog outside the backdrop div |
| Day chips use `aria-pressed` but are visually styled like radio buttons (pickup mode) | In reservation mode, use `role="radiogroup"` + `role="radio"` + `aria-checked` instead of `aria-pressed` for single-select |
| `AllergenInfoIcon` uses emoji 🛡️ as the visual | Emojis render inconsistently across platforms. Use an SVG icon with a consistent design |
| Form modals don't trap focus | The `ProductSelectionModal` traps focus but the `DashboardProducts` form modal and the quick reserve widget do not — add focus trap |
| Color contrast of `--ink-muted: #7a5c3e` on `--bg-card: #fffdf8` | Ratio ~4.2:1, barely meets AA for large text. For small body text (13-14px), bump to `#5e4730` or similar |

### Dead Code

| File | Issue |
|------|-------|
| `OrderSidePanel.tsx`, `ReservationSidePanel.tsx`, `ProductSelectionOverlay.tsx` | No longer imported anywhere — delete |
| `.psm__card-add-btn` CSS class | "Ajouter" button removed but CSS rule still exists — remove |
| `design-bakery-list.md`, `design-product-selection.md` | Design docs in repo root should move to `.kiro/` or a `docs/` folder |

---

## 3. Database Schema

### Missing Indexes

| Table | Column(s) | Why |
|-------|-----------|-----|
| `bakeries` | `owner_id` | `GetMyBakery` and ownership checks filter by owner — needs an index |
| `products` | `id` (covered by PK), but also `is_available` | Menu fetches filter available products |
| `recurring_orders` | `bakery_id` | Seller queries filter by bakery |

### Schema Consistency

| Issue | Recommendation |
|-------|----------------|
| `recurring_orders` uses `TEXT` for IDs but `bakeries`/`orders` use `UUID` | Migrate to UUID for consistency and index performance |
| `recurring_orders.items` uses `JSONB` but `order_items` is a normalized table | Pick one pattern — JSONB is simpler for read-heavy, denormalized data. But inconsistency is confusing |
| `day_schedules.day_of_week` is INT (0-6) but domain uses string ("monday") | The in-memory repo never actually uses the SQL schema — when switching to PostgreSQL, align the mapping |
| No `owner_id` column on `bakeries` table | Migration 001 lacks it. The in-memory repo has it on the struct but it's never migrated. Add a migration |
| `holiday_from` / `holiday_to` in migration 009 has no DOWN section | Add `-- +goose Down` block for reversibility |

### Data Integrity

| Issue | Recommendation |
|-------|----------------|
| `order_items.product_id` references nothing (no FK to products) | Products can be deleted while order_items reference them. Acceptable (keeps history) but add a comment explaining the intentional lack of FK |
| `orders.total_amount` can drift from sum of items | Add a trigger or application-level check on INSERT/UPDATE to verify `total_amount = SUM(subtotals)` |
| `allergens TEXT[]` has no element-level constraint in PostgreSQL | Application layer validates, but a GIN index on the array would enable `@>` queries ("find all products containing dairy") |

---

## 4. i18n

| Issue | Recommendation |
|-------|----------------|
| Translations file is 300+ lines — single file for all locales | Split into per-locale files, or per-feature chunks (`allergens.en.ts`, `allergens.fr.ts`) |
| No type-safety on translation keys | Generate a `TranslationKey` union type from the EN keys, then type `t(key: TranslationKey)` |
| Allergen names are duplicated (allergen constants in Go + translation keys in TS) | Generate the TS constants from the Go source (or share via a JSON file both consume) |
| Baker portal isn't translated (English-only) | Fine for MVP, but document this as a known limitation |

---

## 5. CSS & Styling

| Issue | Recommendation |
|-------|----------------|
| `BakeryDetailPage.css` is 500+ lines | Split into `product-card.css`, `panel.css`, etc. using CSS modules or component-scoped files |
| Duplicate box-shadow/border patterns across components | Define CSS custom properties: `--shadow-card`, `--shadow-lifted`, `--border-card` |
| No dark mode support | Low priority but adding `@media (prefers-color-scheme: dark)` with inverted variables would be a nice UX win |
| Font loading not optimized | Add `<link rel="preload">` for Caveat and Patrick Hand fonts in `index.html`, or use `font-display: swap` |

---

## 6. Configuration & DX

| Issue | Recommendation |
|-------|----------------|
| No `.env.example` file | Add one listing all required/optional env vars with descriptions |
| `Makefile` has no frontend targets | Add `make frontend-dev`, `make frontend-build`, `make frontend-test` |
| No Docker/docker-compose for local dev | Add `docker-compose.yml` with PostgreSQL + backend + frontend for one-command setup |
| Missing `.editorconfig` | Add for consistent formatting (indent size, line endings) across editors |
| `vitest.config.ts` duplicated from `vite.config.ts` | Use `mergeConfig` from vite to extend base config |

---

## 7. Testing

| Issue | Recommendation |
|-------|----------------|
| Frontend tests exist but no CI pipeline | Add GitHub Actions workflow running `make test` + `cd frontend && npm test` |
| No integration test that starts the real HTTP server | Add a `TestMain` that spins up the chi router and runs requests against it (Go `httptest.Server`) |
| Property tests use `rapid` defaults (100 iterations) | Increase to 1000 in CI for better coverage, keep 100 locally for speed |
| No test for the `MarshalJSON` nil-allergens behavior | Add a unit test ensuring `json.Marshal(Product{Allergens: nil})` produces `"allergens": []` |

---

## Priority Summary

**High (do first):**
1. Add `owner_id` column to bakeries migration + index
2. Fix `GetMyBakery` O(n) scan → repository method
3. Replace `time.UnixNano` product IDs with UUIDs
4. Split `ProductSelectionModal` into smaller components
5. Delete dead components (OrderSidePanel, ReservationSidePanel, ProductSelectionOverlay)
6. Add rate limiting to login endpoint

**Medium (quality of life):**
7. Type the `UpdateProduct` handler with a struct instead of `map[string]interface{}`
8. Extract duplicate ownership verification into a helper
9. Split translations into per-locale files
10. Add `.env.example` and docker-compose

**Low (polish):**
11. Move design docs to `docs/` folder
12. Add font preloading
13. Add dark mode CSS variables
14. Generate TypeScript allergen types from Go source
