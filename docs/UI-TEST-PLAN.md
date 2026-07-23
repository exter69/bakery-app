# UI Test Plan — Ma Boulangerie platform

Comprehensive manual + E2E test catalogue for the three-portal SPA, derived from the frontend source (routes, pages, components, API client) on 2026-07-23. Use it for manual QA passes and as the specification for the Playwright suite (see §14).

**Portals under test**
- **Ma Boulangerie** — consumer app (routes `/`, `/bakeries`, …)
- **Votre Boulangerie** — baker back-office (`/dashboard/*`, roles 0/1)
- **Notre Boulangerie** — B2B Comptoir (`/comptoir/*`, role 3)

> Names above are the agreed rebrand (MA-61); the code still shows "Mie & Beurre / Pro / Comptoir" until that ticket lands. Test against whatever the build renders and note mismatches.

---

## 1. How to use this document

Each test case has an **ID**, **preconditions**, **steps**, and **expected result**. Priority: **P0** = critical path / money / auth, **P1** = core feature, **P2** = secondary/cosmetic.

Result legend when executing: ✅ pass · ❌ fail (file/link a ticket) · ⚠️ pass-with-issue · ⛔ blocked · ➖ N/A.

Known-broken cases are marked **[KNOWN: MA-xx]** and cross-referenced in §13 — a failure there is expected and should match the linked ticket, not be re-filed.

---

## 2. Environment & preconditions

| Item | Value |
|------|-------|
| Frontend | `http://localhost:5173` (Vite dev) |
| Backend | `http://localhost:8080` (`make run`) |
| DB mode | In-memory (default, seeded) unless `DATABASE_URL` set |
| Payment | Stripe test mode — card `4242 4242 4242 4242`, any future expiry, any CVC. Never use a real card. |
| Browsers | Chromium, Firefox, WebKit (see §12) |
| Languages | EN / FR / NL |

Before each full pass: hard-reload with cleared `localStorage` (token, guest flag, theme, consent, language all persist there).

---

## 3. Test accounts

Do **not** hard-code credentials in this document. Test accounts come from one of two places depending on DB mode:

- **In-memory mode** — the backend auto-seeds demo accounts at startup (`cmd/server/seed.go`): one Admin, two Bakers (one owning two bakeries), two Customers. No Business (role 3) account and no register-UI path — see TC-AUTH-11.
- **Postgres mode** — accounts (incl. a Business/role-3 account) are provided by the dummy-data script. See `docs/LOCAL-POSTGRES.md` for the current usernames/passwords and how to run it.

Throughout this plan, accounts are referenced by **role** (Customer, Baker, Admin, Business) rather than by name. Substitute whichever seeded account matches. Seed data includes 3 bakeries and their products (prices shown in € in the UI; stored as cents internally).

---

## 4. Cross-cutting: routing & auth gating (P0)

Route protection is defined in `App.tsx`, `ProtectedRoute.tsx`, `RoleRoute.tsx`. Role is decoded client-side from the JWT (`decodeTokenRole`); the backend re-checks on every API call, so client gating is UX only.

| ID | Precondition | Steps | Expected |
|----|--------------|-------|----------|
| TC-ROUTE-01 | Logged out, no guest | Visit `/schedule` | Redirect to `/login`, `from` state = `/schedule` |
| TC-ROUTE-02 | Logged out, no guest | Visit `/history`, `/recurring`, `/settings` each | Each redirects to `/login` |
| TC-ROUTE-03 | Guest mode on | Visit `/schedule` | Allowed (ProtectedRoute permits guest); API calls may 401 → handled |
| TC-ROUTE-04 | Logged in as customer (role 2) | Visit `/dashboard` | Redirect to `/` (role not in [0,1]) |
| TC-ROUTE-05 | Logged in as customer | Visit `/comptoir` | Redirect to `/` (role ≠ 3) |
| TC-ROUTE-06 | Logged in as baker (role 1) | Visit `/dashboard` | Allowed; dashboard shell renders |
| TC-ROUTE-07 | Logged in as baker | Visit `/comptoir` | Redirect to `/` |
| TC-ROUTE-08 | Admin (role 0) | Visit `/dashboard` | Allowed |
| TC-ROUTE-09 | Any | Visit unknown path `/zzz` | NotFoundPage (404) |
| TC-ROUTE-10 | Logged in, token present | Directly deep-link `/dashboard/payouts` | Loads (no redirect); lazy chunk resolves |
| TC-ROUTE-11 | Token with `-`/`_` in base64url payload | Load any RoleRoute page | **[KNOWN: MA-73]** decode may throw → role null → bounced to `/`. Verify whether legit bakers are locked out |
| TC-ROUTE-12 | Session expired (old token) | Trigger any authed API call | `auth:unauthorized` event → redirect to `/login` (unless guest). Expiry only detected on first 401 **[KNOWN: MA-73]** |

---

## 5. Auth flows (P0)

Pages: `LoginPage.tsx`, `RegisterPage.tsx`, `OAuthCallbackPage.tsx`. Token stored in `localStorage`.

| ID | Precondition | Steps | Expected |
|----|--------------|-------|----------|
| TC-AUTH-01 | Logged out | `/register`, role **Customer** (default), username, password ≥6, matching confirm, accept terms, submit | Account created, auto-login, redirect to `/` (customer) |
| TC-AUTH-02 | — | Register with password < 6 chars | Inline error "password ≥ 6", no submit |
| TC-AUTH-03 | — | Register with mismatched confirm password | Inline error, no submit |
| TC-AUTH-04 | — | Register with empty username | Error "Username is required." |
| TC-AUTH-05 | — | Register a duplicate/existing username | Backend 409 → error surfaced, no crash |
| TC-AUTH-06 | — | Register with terms checkbox **unticked** | Verify: is submit blocked? Terms is UI-only; body sends `{username,password,role}` — confirm gate actually enforced ⚠️ |
| TC-AUTH-07 | — | `/register` role **Bakery Owner** (role 1); note extra "code"/token field appears | With valid seller code → seller account, redirect to `/dashboard`. Without/invalid code → backend rejects |
| TC-AUTH-08 | `/register?role=bakery` | Load URL | Page opens pre-set to Bakery mode (title/subtitle switch) |
| TC-AUTH-09 | Logged out | `/login` with valid **Customer** credentials | Login, redirect (customer → `/` or `from`) |
| TC-AUTH-10 | Logged out | `/login` with wrong password | Error `login.error`, stays on page, `role="alert"` shown |
| TC-AUTH-11 | — | Attempt to create/login a **B2B (role 3)** account | **[GAP]** Register UI offers only Customer/Bakery; no seed B2B user. `/comptoir` is unreachable by normal signup. Confirm intended entry path (admin-provisioned? `registerBusiness` API only?) and file if it's a real gap |
| TC-AUTH-12 | Logged out | Login page → "Continue as guest" | Guest mode set, redirect to `/`; protected consumer pages accessible, authed API calls degrade gracefully |
| TC-AUTH-13 | Logged out | Click "Sign in with Google" | Redirects to backend `/auth/oauth/google?state=…`; if Google env unset, expect graceful error not crash |
| TC-AUTH-14 | Logged out | Click "Sign in with Apple" | **[KNOWN: MA-68]** Apple `ExchangeCode` is a stub → flow cannot complete. Button should ideally be hidden |
| TC-AUTH-15 | OAuth return | Hit `/auth/callback?code=…` | Exchanges code, stores token, redirects by role; error param → error UI |
| TC-AUTH-16 | Logged in | Trigger logout (from nav/settings) | Token cleared, redirect to `/` or `/login`; protected routes now redirect |
| TC-AUTH-17 | Logged in as baker | Login redirect | Lands in `/dashboard` (role 1 → dashboard) |

---

## 6. Consumer portal — Ma Boulangerie

### 6.1 Home & navigation (P1)
Page: `HomePage.tsx` (renders own floating nav), `CustomerLayout.tsx` (pill navbar), `Footer.tsx`.

| ID | Steps | Expected |
|----|-------|----------|
| TC-HOME-01 | Load `/` | Hero + floating nav render; no console errors; lazy chunks load |
| TC-HOME-02 | Home → map section | `BakeryMap` (Leaflet) lazy-loads only when scrolled into view; markers render |
| TC-HOME-03 | Home bundle cards (`HomeBundleCard`) | Show surplus bundles or empty state; clicking routes to `/paniers-du-soir` |
| TC-HOME-04 | Nav links (Bakeries, Paniers du soir, About, Guide) | Each routes correctly; active state highlights |
| TC-HOME-05 | Footer links (Privacy, Terms) | Route to `/privacy`, `/terms` |

### 6.2 Bakeries list, search, map (P1)
Pages: `BakeriesPage.tsx`, `BakeryListPage.tsx` (⚠️ unrouted duplicate — MA-68), `SearchBar.tsx`, `BakeryMap.tsx`.

| ID | Steps | Expected |
|----|-------|----------|
| TC-BAK-01 | `/bakeries` | Paginated list (50/pg) of seed bakeries with cards (`BakerCard`), images lazy-load |
| TC-BAK-02 | Search bar: type "crois" | Product/bakery search filters results (`searchProducts`); debounced; clear resets |
| TC-BAK-03 | Search with no matches | Empty-state message, no crash |
| TC-BAK-04 | Toggle list/map view | Map shows bakery markers; clicking marker links to detail |
| TC-BAK-05 | Click a bakery card | Routes to `/bakeries/:id` |

### 6.3 Bakery detail + ordering (P0)
Page: `BakeryDetailPage.tsx` (`createOrder`, `createReservation`), `ProductSelectionModal.tsx`, `AllergenIndicator/Modal`, `HealthScoreDisplay`, `ReviewList`, `StarRating`.

| ID | Precondition | Steps | Expected |
|----|--------------|-------|----------|
| TC-ORD-01 | Any | Open `/bakeries/bakery-1` | Menu grouped by category (Viennoiseries/Breads/Pastries); prices in € |
| TC-ORD-02 | — | Open product → allergen indicator → allergen detail modal | Modal lists allergens; close works; `aria-modal` present |
| TC-ORD-03 | — | Health score display on products | Renders score 1–5 consistently |
| TC-ORD-04 | Logged in as a **Customer** | Add items (delivery mode), set schedule, submit order | `createOrder` called; response `paymentUrl` → `window.location` redirect to payment |
| TC-ORD-05 | On Stripe test page | Pay with `4242…` | Payment succeeds; return to app; order should become Confirmed. **Verify capture is on delivery, not now (MA-33)** |
| TC-ORD-06 | Stub/dev gateway | Submit delivery order | Confirm what `paymentUrl` points to in stub mode (stub page vs real Stripe) — document actual behavior |
| TC-ORD-07 | Logged in | Submit **reservation** (pickup) mode | `createReservation`; reservations are pay-on-spot (no online payment) — no redirect |
| TC-ORD-08 | Guest | Try to place a delivery order | Expected: prompted to log in / 401 handled gracefully (no silent failure) |
| TC-ORD-09 | — | Submit order with empty cart | Blocked with validation message |
| TC-ORD-10 | — | Reviews section: read reviews (`ReviewList`), star rating | Renders; pagination if many |
| TC-ORD-11 | Logged in, past delivered order | Leave a review (`ReviewPrompt`/`createReview`) | Review submits; appears in list; rating constrained 1–5 |
| TC-ORD-12 | — | **Security:** attempt to confirm/cancel an order you don't own via API | **[KNOWN: MA-63]** payment callback lacks ownership + verification — free goods / IDOR. Confirm still exploitable |

### 6.4 Paniers du soir (surplus bundles) (P1)
Page: `BundlePage.tsx`, `BundleCard.tsx`, `BundleMapView.tsx`, `ReservationRail.tsx`, `ImpactCard.tsx`.

| ID | Precondition | Steps | Expected |
|----|--------------|-------|----------|
| TC-BUN-01 | Guest/logged out | `/paniers-du-soir` | Bundles are **visible** to logged-out users (MA-49) as preview |
| TC-BUN-02 | Logged out | Click "Reserve" on a bundle | Redirect to `/login` with `from=/paniers-du-soir` (gate on submit) |
| TC-BUN-03 | Logged in | Reserve an available bundle | `reserveBundle`; reservation rail shows active reservation; stock decrements live (WebSocket) |
| TC-BUN-04 | Logged in | Confirm the reservation | `confirmReservation`; status updates |
| TC-BUN-05 | Two sessions | Reserve the last unit simultaneously | Loser gets 409 → alert `bundles.error.unavailable`; no oversell |
| TC-BUN-06 | — | Filter bundles (map/list) | Filters apply; empty state when none |
| TC-BUN-07 | — | Impact card figures | Renders (waste saved etc.) consistently |

### 6.5 Schedule, history, recurring (P1)
Pages: `ScheduleOrdersPage.tsx`, `OrderHistoryPage.tsx`, `RecurringOrdersPage.tsx`.

| ID | Precondition | Steps | Expected |
|----|--------------|-------|----------|
| TC-SCH-01 | Logged in | `/schedule` | Upcoming scheduled orders list; empty state otherwise |
| TC-HIS-01 | a **Customer** | `/history` | Past orders (20/pg), status badges, filter/sort |
| TC-HIS-02 | a **Customer** | Click "Re-order" on a past order | `storeReorderData` → routes to bakery with cart prefilled (`consumeReorderData`) |
| TC-HIS-03 | a **Customer** | Cancel a cancellable order | Order → cancelled; delivered/ready orders not cancellable |
| TC-REC-01 | Logged in | `/recurring` create a weekly recurring order | Created; appears in list |
| TC-REC-02 | — | Pause / resume / delete recurring order | State changes reflect (`pause/resume/deleteRecurringOrder`) |
| TC-REC-03 | — | Verify a recurring order actually generates real orders over time | **[KNOWN]** recurring service is CRUD-only; no scheduler materializes orders (noted in MA-69). Expect no auto-generated orders |

### 6.6 Account settings & GDPR (P0)
Page: `AccountSettingsPage.tsx`.

| ID | Precondition | Steps | Expected |
|----|--------------|-------|----------|
| TC-ACC-01 | Logged in | `/settings` | Profile view/edit; save persists |
| TC-ACC-02 | Logged in | Click "Export my data" | Downloads JSON with profile/orders/reservations/etc. **Verify reviews included (MA-71 says omitted)** |
| TC-ACC-03 | Logged in | Click "Delete account", confirm dialog | Account anonymized, logged out. **[KNOWN: MA-71]** IBAN/social logins/push subs may persist; JWT stays valid |
| TC-ACC-04 | — | Cancel the delete confirmation | No deletion |

### 6.7 Static pages, consent, push (P2)
| ID | Steps | Expected |
|----|-------|----------|
| TC-STAT-01 | Visit `/about`, `/guide`, `/privacy`, `/terms` | Render localized content; no raw i18n keys |
| TC-CON-01 | First visit | Cookie consent banner appears (`CookieConsent`); non-blocking |
| TC-CON-02 | Accept / decline | Choice stored in localStorage; banner gone on reload. **[KNOWN: MA-68]** consent gates nothing — Sentry loads regardless |
| TC-PUSH-01 | Logged in, supported browser | Toggle push notifications (`PushNotificationToggle`) | Permission prompt → VAPID subscribe; toggle reflects state |
| TC-PUSH-02 | Deny permission | Toggle | Graceful handling, no crash |

---

## 7. Baker portal — Votre Boulangerie (`/dashboard`, roles 0/1) (P0/P1)

Layout `DashboardLayout.tsx` sidebar: Tableau de bord, Commandes, Menu & stock, Paniers du soir, Paiements, Statistiques, Boutique, B2B Comptoir. ⚠️ Sidebar labels are hardcoded French, bypassing i18n (MA-73).

| ID | Precondition | Steps | Expected |
|----|--------------|-------|----------|
| TC-DASH-01 | a **Baker** | Login → `/dashboard` | Overview with KPI cards; data for owned bakery |
| TC-DASH-02 | — | KPI "today" figures | **[KNOWN: MA-73]** labelled "today" but actually `status=confirmed`, no date filter — verify wording |
| TC-DASH-03 | — | `/dashboard/orders` kanban | 4 columns (confirmed → preparing → ready → delivered); seed orders present |
| TC-DASH-04 | — | Drag an order across columns | Status updates (`updateOrderStatus`); persists on reload |
| TC-DASH-05 | — | Move order to a backward/invalid state | **[KNOWN: MA-68]** no unified state-machine guard; verify invalid transitions (e.g. Ready→Cancelled) are prevented or file |
| TC-DASH-06 | — | `/dashboard/products` (Menu & stock) | Card inventory; inline stock edit (`StockStepper`); day toggles |
| TC-DASH-07 | — | Edit a product price/availability | Persists (`updateProduct`); reflected on consumer menu |
| TC-DASH-08 | — | Create a new product | `createProduct`; appears in list and on storefront |
| TC-DASH-09 | — | Upload a product image | `uploadImage`. **[KNOWN: MA-68]** with `UPLOAD_STORAGE=s3` upload is faked → image 404s. In local/disk mode verify it works |
| TC-DASH-10 | — | Delete a product | `deleteProduct`; removed |
| TC-DASH-11 | — | `/dashboard/bundles` composer | Create/publish a surplus bundle (`createBundle`/`publishBundle`); appears on consumer `/paniers-du-soir` |
| TC-DASH-12 | — | `/dashboard/payouts` | Connect status; if not onboarded, "Start onboarding" (`startOnboarding` → Stripe). Payout history table with pagination |
| TC-DASH-13 | — | Complete Stripe Connect onboarding (test) | Returns; status flips. **[KNOWN: MA-65]** `account.updated` only logged — status may not sync |
| TC-DASH-14 | — | `/dashboard/stats` | ⚠️ Route labelled "Statistiques" actually renders the schedule editor (`DashboardSchedule`) — MA-68. Verify content vs label |
| TC-DASH-15 | — | `/dashboard/bakery` (Boutique) | Edit bakery profile, hours, holidays (`updateBakery`/`updateHoliday`); saves |
| TC-DASH-16 | — | `/dashboard/b2b` | B2B settings page loads |
| TC-DASH-17 | a **Baker** owning 2 bakeries | Switch active bakery (if selector present) | Data scopes to selected bakery |
| TC-DASH-18 | a second **Baker** | Try to view/mutate `bakery-1` orders (not hers) via API | Backend ownership check blocks (verifyBakeryOwnership) |

---

## 8. B2B Comptoir — Notre Boulangerie (`/comptoir`, role 3) (P0/P1)

**Blocker:** no UI path creates a role-3 account (TC-AUTH-11). To test, provision a B2B user (admin/API/DB) or via `registerBusiness`. Layout `ComptoirLayout.tsx`, nav: Commander, Récurrences, Livraisons, Factures, Profil. `SiteProvider`/`SiteSwitcher` for multi-site.

| ID | Precondition | Steps | Expected |
|----|--------------|-------|----------|
| TC-B2B-01 | Role-3 user | Login → `/comptoir` | Comptoir shell (business blue, no marketing chrome) |
| TC-B2B-02 | — | Commander order sheet (`CommanderPage`/`CommandeRapide`) | Spreadsheet-style rows; type quantities; multi-bakery basket |
| TC-B2B-03 | — | HT/TTC summary (`B2BCartSummary`) | Sous-total HT → remise pro (**1%**, MA-41) → TVA 5.5% → Total TTC computed correctly |
| TC-B2B-04 | — | Submit an order sheet | Order placed across bakeries; confirmation |
| TC-B2B-05 | — | Saved lists (`SavedListPicker`, `createSavedList`) | Create/apply/delete a saved order list |
| TC-B2B-06 | — | Site switcher (multi-site account) | Switching site scopes deliveries/invoices |
| TC-B2B-07 | — | `/comptoir/recurrences` | **[KNOWN: MA-69]** placeholder shell — empty list, "New" form has only a label+cancel, no persistence. Expect non-functional |
| TC-B2B-08 | — | `/comptoir/livraisons` | Deliveries list (`listDeliveries`); pagination |
| TC-B2B-09 | — | Livraisons "Éditer" on an upcoming delivery | **[KNOWN: MA-69]** button has no handler — does nothing |
| TC-B2B-10 | — | `/comptoir/factures` | Monthly statements (`listInvoices`); volume-tier nudge shown |
| TC-B2B-11 | — | Download an invoice PDF (`downloadInvoicePDF`) | PDF downloads; on failure **[KNOWN: MA-69]** it "silently fails" — verify error surfaced |
| TC-B2B-12 | Backend down | Load Factures/Livraisons | **[KNOWN: MA-69]** errors swallowed → false "Aucune facture" empty state instead of error UI |
| TC-B2B-13 | — | `/comptoir/profile` | Company profile, sites, access requests (`listAccessRequests`, approve/reject) |
| TC-B2B-14 | FR locale | Inspect Comptoir strings | **[KNOWN: MA-69]** accent-stripped French ("Confirmee", "A venir", "Editer") + hardcoded English "{n} items" |
| TC-B2B-15 | — | Bakery access flow (`requestAccess`/`approveAccess`/`revokeAccess`) | Request → appears for baker → approve → bakery listed in `listApprovedBakeries` |

---

## 9. Critical components matrix (P1)

Reusable components to test in isolation (Storybook stories exist for several).

| Component | Checks |
|-----------|--------|
| `SearchBar` | Debounce, clear, empty query, keyboard submit, aria-label |
| `StarRating` | Display vs interactive, 1–5 bounds, empty/full, keyboard |
| `AllergenIndicator` / modals | Single & multiple allergens, modal open/close, aria-modal, focus |
| `HealthScoreDisplay` | Scores 1–5 render, color coding |
| `BundleCard` / `HomeBundleCard` | Compose vs surprise, sold-out state, reserve button loading |
| `StockStepper` (pro) | Min/max clamp, no negative, disabled at bounds |
| `OrderCard` (pro) | Status pill, actions per status |
| `FilterChips` (pro) | Toggle, active state, multi-select |
| `StatCard` (pro) | Value/label render |
| `ErrorBanner` (pro) | Shows message, dismiss |
| `B2BCartSummary` | HT/TVA/TTC math, loading, max-tier state |
| `ThemeSwitcher` | Cycles light/dark/system; persists. ⚠️ tests stale (MA-67) |
| `LanguageSwitcher` | EN/FR/NL switch persists across reload |
| `PushNotificationToggle` | Subscribe/unsubscribe, permission-denied path |
| `LoadingSpinner` | Renders during lazy chunk / Suspense |
| `Footer` | Links, localized |

---

## 10. Internationalization (P1)

`i18n/translations.ts` — verified at **372 keys per locale (EN/FR/NL), zero missing/extra**. Sections: about, allergen, baker, bakeries, bundles, common, comptoir, consent, dashboard, footer, guide, health, history, home, login, nav, privacy, pro, register, reviews, search, settings, terms, theme.

| ID | Steps | Expected |
|----|-------|----------|
| TC-I18N-01 | Switch language EN→FR→NL on consumer app | All visible strings translate; no raw `key.path` rendered |
| TC-I18N-02 | Reload after switching | Language persists (localStorage) |
| TC-I18N-03 | Baker dashboard pages in FR/NL | **[KNOWN: MA-73]** 7 dashboard pages bypass i18n — mixed FR/EN hardcoded strings remain |
| TC-I18N-04 | Comptoir in FR | **[KNOWN: MA-69]** accent-stripped FR strings |
| TC-I18N-05 | Every routed page, each locale | No missing-key fallbacks (t() renders key verbatim on miss) |

---

## 11. Theming, responsive, accessibility (P2)

| ID | Steps | Expected |
|----|-------|----------|
| TC-THEME-01 | Toggle dark mode | `[data-theme]` variable system applies; 53/56 CSS files consume vars; persists |
| TC-THEME-02 | System theme = dark, first load | Pre-JS `prefers-color-scheme` fallback applies (no flash) |
| TC-RESP-01 | Consumer app at 375px / 768px / 1280px | Layout adapts; nav usable; no overflow |
| TC-RESP-02 | Dashboard kanban on narrow viewport | Usable / horizontally scrollable |
| TC-A11Y-01 | Modals (allergen, product, reservation) | `role="dialog"`, `aria-modal`, focus trap, Esc closes |
| TC-A11Y-02 | Toggles/icon buttons | `aria-pressed` / `aria-label` present |
| TC-A11Y-03 | Error messages | `role="alert"` announced |
| TC-A11Y-04 | Keyboard-only nav of a full order flow | All actions reachable via keyboard |

---

## 12. Cross-browser matrix (P1)

Run the P0 paths (auth, order+pay, baker order-status, B2B order sheet) on each engine.

| Flow | Chromium | Firefox | WebKit |
|------|----------|---------|--------|
| Register + login | | | |
| Guest browse | | | |
| Delivery order + Stripe test pay | | | |
| Bundle reserve + confirm (WebSocket) | | | |
| Baker order kanban drag | | | |
| B2B order sheet + HT/TTC | | | |
| Language switch EN/FR/NL | | | |
| Dark mode | | | |

WebSocket (bundle live stock, notifications) and Leaflet map are the most likely to differ across engines — watch those.

---

## 13. Known-broken / expected failures

These come from the 2026-07-23 code review. A failure matching the description is **expected**; link the ticket rather than re-filing.

| Area | Symptom | Ticket |
|------|---------|--------|
| Payment callback | Forged "success" confirms order; any user can cancel any order | **MA-63** |
| Postgres mode | Register/order/reservation/B2B fail (non-UUID IDs, role CHECK) — in-memory only works | **MA-62** |
| Apple SSO | Button present, backend stub → cannot complete | **MA-68** |
| S3 upload | Fake upload → image 404s when `UPLOAD_STORAGE=s3` | **MA-68** |
| Cookie consent | Gates nothing; Sentry loads regardless | **MA-68** |
| Order state machine | No unified guard; invalid transitions possible | **MA-68** |
| Comptoir Récurrences | Placeholder shell, no persistence | **MA-69** |
| Livraisons "Éditer" | Dead button | **MA-69** |
| Factures/Livraisons errors | Swallowed → false empty state; silent PDF fail | **MA-69** |
| Comptoir FR | Accent-stripped + hardcoded English | **MA-69** |
| JWT decode | base64url `-`/`_` can throw → user bounced from portal | **MA-73** |
| Session expiry | Only detected on first 401 | **MA-73** |
| Dashboard i18n | 7 pages hardcoded mixed-language | **MA-73** |
| Refund → payout | Reversal never wired; bakery keeps money | **MA-65** |
| GDPR delete | IBAN/social/push persist; JWT still valid | **MA-71** |
| Recurring orders | Never materialize into real orders | MA-69 (note) |
| B2B signup | No UI path to create role-3 account | file if confirmed |
| Unrouted duplicates | `BakeryListPage`, `DashboardReservations`, `DashboardBundlesPage` unused | **MA-68** |

---

## 14. Mapping to the Playwright suite

Target spec files under `e2e/tests/` (see MA-67 for why the current suite never runs in CI — fix that first):

| Spec file | Covers test IDs |
|-----------|-----------------|
| `auth.spec.ts` | TC-AUTH-01…17, TC-ROUTE-01…12 |
| `customer-browse.spec.ts` | TC-HOME-*, TC-BAK-* |
| `customer-order.spec.ts` | TC-ORD-01…09, TC-ORD-11 |
| `customer-reservation.spec.ts` | TC-ORD-07, TC-SCH-01 |
| `bundles.spec.ts` | TC-BUN-01…07 |
| `order-history.spec.ts` | TC-HIS-01…03, TC-REC-01…03 |
| `account-gdpr.spec.ts` | TC-ACC-01…04 |
| `i18n.spec.ts` | TC-I18N-01…05 |
| `theme-responsive.spec.ts` | TC-THEME-*, TC-RESP-*, TC-A11Y-* |
| `baker-portal.spec.ts` **(new — MA-67)** | TC-DASH-01…18 |
| `comptoir.spec.ts` **(new — MA-67)** | TC-B2B-01…15 |
| `security.spec.ts` **(new)** | TC-ORD-12, TC-ROUTE-04/05/07 (negative authz) |

Guidance for the specs:
- Run against **in-memory backend** with seed data (fixtures in `e2e/helpers/test-data.ts` assume a **Customer**/a **Baker**/`bakery-1`), OR add Postgres seeding — pick one and make CI match (MA-67).
- Use Stripe **test** cards; assert capture-on-delivery (MA-33), not capture-on-order.
- Add `chromium`, `firefox`, `webkit` projects in `playwright.config.ts`.
- For known-broken cases, write the test to assert the **correct** behavior and mark `test.fixme()` with the ticket ID, so the suite turns green automatically when the fix lands.

---

*Generated from source on 2026-07-23. Update alongside routing/feature changes. Pairs with `docs/CHANGES-BY-FEATURE.md` and the Notion docs (pages 06–07 features, 15 business).*
