# Frontend Performance Guide

## Route-Level Code Splitting

All page components are loaded via `React.lazy()` with dynamic imports in `src/App.tsx`.
Each route resolves to its own JS chunk, downloaded only when the user navigates to that page.

A `<Suspense fallback={<LoadingSpinner />}>` wrapper around the `<Routes>` tree shows a
centered spinner while a chunk loads.

**Why this matters:** The initial bundle only contains the framework, router, and shared
layout code. Heavy pages (dashboard, B2B comptoir) never load for anonymous visitors.

## Lazy-Loaded Heavy Dependencies

### Leaflet / react-leaflet

`BakeryMap` (used in `HomePage`) imports `react-leaflet`, `leaflet`, and its CSS/images.
This component is lazy-loaded separately from the HomePage module:

```tsx
const BakeryMap = lazy(() => import('../components/BakeryMap'));
```

Wrapped in its own `<Suspense>` with a map-height placeholder spinner so the page layout
doesn't shift while the map loads.

### Recharts

`recharts` is available as a dependency for future dashboard charts. When adding chart
components, lazy-load them the same way to keep the seller portal chunks lean.

## Font Loading Strategy

Fonts (Caveat, Patrick Hand) are loaded via Google Fonts with `&display=swap`:

```html
<link href="https://fonts.googleapis.com/css2?family=Caveat:wght@500;600&family=Patrick+Hand&display=swap" rel="stylesheet">
```

`font-display: swap` ensures text is visible immediately with a system fallback, then
swaps to the custom font once downloaded. This prevents invisible text (FOIT) and
keeps LCP fast.

Preconnect hints (`fonts.googleapis.com` and `fonts.gstatic.com`) are also included
to reduce DNS/TLS overhead.

## Image Optimization

- All below-the-fold images use `loading="lazy"` (bakery cards, product photos, bundle cards).
- Above-the-fold images (hero photos on Login/Register, bakery detail header) intentionally
  omit `loading="lazy"` to avoid hurting Largest Contentful Paint.
- All `<img>` elements include `alt` text for accessibility.
- External images (Unsplash) use width parameters (`?w=400`, `?w=1200`) to request
  appropriately sized files rather than full-resolution originals.

## Performance Targets

| Metric | Target | Description |
|--------|--------|-------------|
| LCP    | < 2.5s | Largest Contentful Paint — hero image or main heading |
| CLS    | < 0.1  | Cumulative Layout Shift — stable layout during load |
| INP    | < 200ms| Interaction to Next Paint — responsive to user input |

## Running a Local Lighthouse Audit

```bash
# Start the dev server
cd frontend && npm run dev

# In another terminal, run Lighthouse
npx lighthouse http://localhost:5173 --output html --output-path ./lighthouse-report.html

# Or for a production build:
npm run build && npm run preview
npx lighthouse http://localhost:4173 --output html --output-path ./lighthouse-report.html
```

## Vite Build — Code-Split Chunks

After running `npm run build`, inspect `dist/assets/` to verify multiple JS chunks exist.
Each lazy-loaded route produces its own chunk file. The output should show:

- A main entry chunk (framework + router)
- Separate chunks for each page group (Home, Bakeries, Dashboard, Comptoir, etc.)
- A dedicated chunk for `leaflet`/`react-leaflet`

## Future Improvements

- **List virtualization**: For bakery lists or product grids with 50+ items, add
  `react-window` or `@tanstack/virtual` to only render visible rows.
- **Lighthouse CI**: Integrate `@lhci/cli` in the GitHub Actions pipeline to fail PRs
  that regress performance budgets.
- **Image CDN**: Move from Unsplash direct links to a CDN with automatic format
  negotiation (WebP/AVIF) and responsive `srcset`.
- **Prefetch**: Add `<link rel="prefetch">` hints for likely next-page navigations
  (e.g., prefetch BakeryDetailPage when hovering a bakery card).
