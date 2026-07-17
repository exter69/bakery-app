---
inclusion: auto
---

# UI Library — Tailwind UI Components

When building or modifying any frontend page or component, **always check the UI library first** before writing custom markup or CSS.

## Location

The UI library lives at `.kiro/UI/` and contains ~684 Tailwind UI components across four kits. See `.kiro/UI/README.md` for the full index.

## Rules

1. **Use React JSX variants** from the `react/` subfolder of each kit (the project uses React + TypeScript).
2. **Copy and adapt** — components are self-contained `export default function Example()` blocks. Rename, add types, and wire up props as needed.
3. **Prefer these components over writing custom CSS** — they already include dark mode, responsive design, and accessibility.
4. **Maintain the indigo accent color** already used in the project (find-replace if theming changes).
5. **Icons come from Heroicons** (already used in the library files) — install `@heroicons/react` if not already present.

## Quick Lookup

| Need | Path |
|------|------|
| Login / sign-in form | `application-ui-v4/react/forms/sign-in-forms/` |
| Data tables | `application-ui-v4/react/lists/tables/` |
| Cards | `application-ui-v4/react/layout/cards/` |
| Side panels / drawers | `application-ui-v4/react/overlays/drawers/` |
| Modals / dialogs | `application-ui-v4/react/overlays/modal-dialogs/` |
| Buttons | `application-ui-v4/react/elements/buttons/` |
| Form inputs | `application-ui-v4/react/forms/input-groups/` |
| Navigation / navbar | `application-ui-v4/react/navigation/navbars/` |
| Sidebar layout | `application-ui-v4/react/application-shells/sidebar/` |
| Alerts / feedback | `application-ui-v4/react/feedback/alerts/` |
| Empty states | `application-ui-v4/react/feedback/empty-states/` |
| Product lists | `ecommerce-v4/react/components/product-lists/` |
| Stats / metrics | `application-ui-v4/react/data-display/stats/` |
| Pagination | `application-ui-v4/react/navigation/pagination/` |
| 404 pages | `marketing-v4/react/feedback/404-pages/` |

## Reference

#[[file:.kiro/UI/README.md]]
