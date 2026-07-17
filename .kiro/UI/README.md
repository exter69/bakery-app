# UI Kit — Navigation Guide

> Tailwind Plus (Tailwind UI) component library. This README is written so an AI agent (or a human) can locate the right component fast without scanning every folder.

## What's here

Four kits, ~684 copy-paste components + 27 packaged React components.

| Kit | Purpose | Components | Formats |
|-----|---------|-----------:|---------|
| `application-ui-v4/` | Dashboards, admin panels, app interiors (forms, tables, nav, overlays) | 364 | React (`.jsx`), HTML (`.html`), Vue (`.vue`) |
| `ecommerce-v4/` | Storefronts, product pages, carts, checkout | 114 | React, HTML, Vue |
| `marketing-v4/` | Landing pages, marketing sections (heroes, pricing, footers) | 179 | React, HTML, Vue |
| `catalyst-ui-kit/` | Ready-made React component library (importable, not copy-paste) | 27 | TypeScript (`.tsx`), JavaScript (`.jsx`) |

## How the files are organized

Each of the three `*-v4` kits mirrors the same structure across three framework folders:

```
<kit>/
  react/   <category>/ <subcategory>/ NN-variant-name.jsx
  html/    <category>/ <subcategory>/ NN-variant-name.html
  vue/     <category>/ <subcategory>/ NN-variant-name.vue
```

Rules to rely on:
- **The three format folders are identical in structure** — same categories, subcategories, and filenames (only the extension differs). Pick your framework folder, ignore the other two.
- Files are numbered (`01-…`, `02-…`); numbers are non-contiguous (some variants were removed upstream). Don't assume `01` exists.
- Each React/Vue file is a self-contained `export default function Example()` — drop it in, no props required.
- Styling is Tailwind utility classes inline, with dark-mode variants (`dark:…`) already included. Default accent color is `indigo`.

`catalyst-ui-kit/` is different — it's a real component package (`<Button>`, `<Dialog>`, etc.) with `typescript/` and `javascript/` source, plus `demo/` usage examples. Use it when you want importable components rather than pasted markup.

## Fast index — where to find things

### application-ui-v4 — app & dashboard UI
Path prefix: `application-ui-v4/{react|html|vue}/`

- **application-shells/** — full page skeletons: `multi-column` (6), `sidebar` (8), `stacked` (9)
- **data-display/** — `calendars` (8), `description-lists` (6), `stats` (5)
- **elements/** — `avatars` (11), `badges` (16), `button-groups` (5), `buttons` (8), `dropdowns` (5)
- **feedback/** — `alerts` (6), `empty-states` (6)
- **forms/** — `action-panels` (8), `checkboxes` (4), `comboboxes` (4), `form-layouts` (4), `input-groups` (21), `radio-groups` (12), `select-menus` (7), `sign-in-forms` (4), `textareas` (5), `toggles` (5)
- **headings/** — `card-headings` (6), `page-headings` (9), `section-headings` (10)
- **layout/** — `cards` (10), `containers` (5), `dividers` (8), `list-containers` (7), `media-objects` (8)
- **lists/** — `feeds` (3), `grid-lists` (7), `stacked-lists` (15), `tables` (19)
- **navigation/** — `breadcrumbs` (4), `command-palettes` (8), `navbars` (11), `pagination` (3), `progress-bars` (8), `sidebar-navigation` (5), `tabs` (9), `vertical-navigation` (6)
- **overlays/** — `drawers` (12), `modal-dialogs` (6), `notifications` (6)
- **page-examples/** — `detail-screens` (2), `home-screens` (2), `settings-screens` (2)

### ecommerce-v4 — store UI
Path prefix: `ecommerce-v4/{react|html|vue}/`

- **components/** — `category-filters` (5), `category-previews` (6), `checkout-forms` (5), `incentives` (8), `order-history` (4), `order-summaries` (4), `product-features` (9), `product-lists` (11), `product-overviews` (5), `product-quickviews` (4), `promo-sections` (8), `reviews` (4), `shopping-carts` (6), `store-navigation` (5)
- **page-examples/** — `category-pages` (5), `checkout-pages` (5), `order-detail-pages` (3), `order-history-pages` (5), `product-pages` (5), `shopping-cart-pages` (3), `storefront-pages` (4)

### marketing-v4 — marketing & landing UI
Path prefix: `marketing-v4/{react|html|vue}/`

- **elements/** — `banners` (13), `flyout-menus` (7), `headers` (11)
- **feedback/** — `404-pages` (5)
- **page-examples/** — `about-pages` (3), `landing-pages` (4), `pricing-pages` (3)
- **sections/** — `bento-grids` (3), `blog-sections` (7), `contact-sections` (7), `content-sections` (7), `cta-sections` (11), `faq-sections` (7), `feature-sections` (15), `footers` (7), `header` (8), `heroes` (12), `logo-clouds` (6), `newsletter-sections` (6), `pricing` (12), `stats-sections` (8), `team-sections` (9), `testimonials` (8)

### catalyst-ui-kit — importable React components
Path: `catalyst-ui-kit/{typescript|javascript}/`

alert, auth-layout, avatar, badge, button, checkbox, combobox, description-list, dialog, divider, dropdown, fieldset, heading, input, link, listbox, navbar, pagination, radio, select, sidebar, sidebar-layout, stacked-layout, switch, table, text, textarea

Usage examples live in `catalyst-ui-kit/demo/{typescript|javascript}/`. See `catalyst-ui-kit/README.md`.

## Recipes for common searches

- **"I need a login page"** → `application-ui-v4/react/forms/sign-in-forms/` or `catalyst/auth-layout`
- **"A pricing table"** → `marketing-v4/react/sections/pricing/` (12 variants)
- **"A data table"** → `application-ui-v4/react/lists/tables/` (19 variants)
- **"A product page"** → `ecommerce-v4/react/page-examples/product-pages/`
- **"A hero section"** → `marketing-v4/react/sections/heroes/` (12 variants)
- **"A dashboard shell with sidebar"** → `application-ui-v4/react/application-shells/sidebar/`
- **"A modal / dialog"** → `application-ui-v4/react/overlays/modal-dialogs/` or `catalyst/dialog`

## Grep tips for agents

```bash
# List every variant in a subcategory (any format):
ls application-ui-v4/react/lists/tables/

# Find components by keyword across a kit:
grep -rl "checkout" ecommerce-v4/react/

# Find every subcategory anywhere:
find . -type d -mindepth 3 -maxdepth 3

# Same component, different framework: swap the folder + extension.
# react/.../buttons/01-primary-buttons.jsx  <->  html/.../buttons/01-primary-buttons.html
```

## Conventions summary

- Framework: React JSX / plain HTML / Vue SFC — same content, pick one.
- CSS: Tailwind v4 utility classes, inline. Dark mode built in via `dark:` variants.
- Accent color: `indigo` throughout (find-and-replace to re-theme).
- Icons: Heroicons (imported in React/Vue files; inline SVG in HTML).
- No build step needed for the markup itself — it's plain Tailwind.
