# ⚛️ Frontend

## Directory Structure

```
frontend/src/
├── api/                 # API client modules
│   ├── client.ts        # Base fetch wrapper, auth headers, guest mode
│   ├── bakeries.ts      # Bakery & menu endpoints
│   ├── orders.ts        # Order & reservation endpoints
│   ├── recurring.ts     # Recurring order endpoints
│   └── seller.ts        # Seller portal endpoints
├── components/          # Shared components
│   ├── dashboard/       # Baker dashboard components
│   ├── AllergenIndicator.tsx
│   ├── AllergenInfoIcon.tsx
│   ├── AllergenInfoModal.tsx
│   ├── AllergenDetailModal.tsx
│   ├── BakerCard.tsx
│   ├── BakeryMap.tsx
│   ├── CustomerLayout.tsx
│   ├── Footer.tsx
│   ├── HealthScoreDisplay.tsx
│   ├── LanguageSwitcher.tsx
│   ├── ProductSelectionModal.tsx
│   ├── ProtectedRoute.tsx
│   └── RoleRoute.tsx
├── i18n/                # Internationalization
│   ├── I18nContext.tsx   # React context provider
│   ├── translations.ts  # All translation strings
│   └── index.ts         # Exports
├── pages/               # Route-level page components
│   ├── dashboard/       # Baker portal pages
│   ├── HomePage.tsx
│   ├── BakeriesPage.tsx
│   ├── BakeryDetailPage.tsx
│   ├── ScheduleOrdersPage.tsx
│   ├── RecurringOrdersPage.tsx
│   ├── LoginPage.tsx
│   ├── RegisterPage.tsx
│   ├── AboutPage.tsx
│   ├── GuidePage.tsx
│   └── NotFoundPage.tsx
├── types/               # TypeScript interfaces
│   ├── bakery.ts        # Bakery, Product, Menu types
│   └── order.ts         # Order, Reservation types
└── __tests__/           # Unit & property tests
```

---

## Routing (App.tsx)

| Path | Component | Auth |
|------|-----------|------|
| `/` | HomePage | Public |
| `/login` | LoginPage | Public |
| `/register` | RegisterPage | Public |
| `/bakeries` | BakeriesPage | Public |
| `/bakeries/:id` | BakeryDetailPage | Public |
| `/schedule` | ScheduleOrdersPage | Protected |
| `/recurring` | RecurringOrdersPage | Protected |
| `/about` | AboutPage | Public |
| `/guide` | GuidePage | Public |
| `/dashboard` | DashboardOverview | Seller/Admin |
| `/dashboard/bakery` | DashboardBakery | Seller/Admin |
| `/dashboard/products` | DashboardProducts | Seller/Admin |
| `/dashboard/schedule` | DashboardSchedule | Seller/Admin |
| `/dashboard/orders` | DashboardOrders | Seller/Admin |
| `/dashboard/reservations` | DashboardReservations | Seller/Admin |

---

## Key Components

### ProductSelectionModal
Modal for selecting products when placing an order or reservation. Shows categories, quantities, allergen badges, and running total.

### AllergenIndicator
Small colored pill showing allergen names. Used in product cards and menus.

### HealthScoreDisplay
Visual 1–5 score indicator for product healthiness.

### CustomerLayout
Wraps customer pages with the pill-style navigation bar and footer.

### BakeryMap
Leaflet map showing bakery locations with interactive markers.

---

## State Management

- No Redux or external state library
- React hooks (`useState`, `useEffect`) for local component state
- `useNavigate` for programmatic routing
- Custom `useI18n` hook for translations
- Auth state: JWT in localStorage, decoded for role/user info

---

## Design Approach

- **Customer portal**: Warm artisan style with handwritten fonts (Caveat, Patrick Hand)
- **Baker portal**: Clean system sans-serif with indigo accent
- See [Design System](./11-design-system.md) for full details
