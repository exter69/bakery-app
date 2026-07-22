# 🎨 Design System

## Two Portals, Two Identities

The app has distinct visual identities for each user type:

| Portal | Audience | Vibe |
|--------|----------|------|
| Customer | End users | Warm, artisanal, playful |
| Baker | Shop owners | Clean, professional, functional |

---

## Customer Portal

### Fonts

| Usage | Font | Style |
|-------|------|-------|
| Headings | Caveat | Handwritten, warm |
| Body | Patrick Hand | Casual, friendly |

### Color Palette (CSS Variables)

| Variable | Value | Usage |
|----------|-------|-------|
| `--bg-page` | `#faf6f1` | Page background (warm cream) |
| `--bg-card` | `#ffffff` | Card backgrounds |
| `--ink` | `#2d2a26` | Primary text |
| `--accent` | `#c0785a` | Buttons, links, highlights |
| `--accent-hover` | `#a8634a` | Button hover state |
| `--muted` | `#8c857d` | Secondary text |
| `--border` | `#e8e0d8` | Card borders |
| `--success` | `#4a9a5b` | Status: ready, confirmed |
| `--warning` | `#d4a843` | Status: preparing |
| `--error` | `#c54b4b` | Error states |

### Component Patterns

- **Cards**: White background, 1px ink border, 3px offset shadow
- **Buttons**: Pill shape (border-radius: 999px), solid accent color
- **Badges**: Small pills for allergens, status indicators
- **Shadows**: Offset ink shadows (2px 2px) for depth without realism

---

## Baker Portal

### Fonts

| Usage | Font |
|-------|------|
| All text | System sans-serif (`-apple-system, BlinkMacSystemFont, 'Segoe UI'`) |

### Color Palette

| Variable | Value | Usage |
|----------|-------|-------|
| `--dash-bg` | `#f8f9fa` | Dashboard background |
| `--dash-card` | `#ffffff` | Panel backgrounds |
| `--dash-accent` | `#4f46e5` | Indigo — primary actions |
| `--dash-accent-hover` | `#4338ca` | Button hover |
| `--dash-text` | `#1f2937` | Primary text |
| `--dash-muted` | `#6b7280` | Secondary text |
| `--dash-border` | `#e5e7eb` | Borders |

### Component Patterns

- **Layout**: Sidebar navigation + main content area
- **Tables**: Striped rows, sortable headers
- **Forms**: Standard inputs with label above
- **Cards**: Subtle shadow, rounded corners (8px)

---

## Responsive Breakpoints

| Breakpoint | Target |
|-----------|--------|
| `< 720px` | Mobile — single column, stacked layout |
| `720px – 900px` | Tablet — reduced padding, collapsible nav |
| `> 900px` | Desktop — full layout with sidebar |

---

## Shared Patterns

- Prices displayed as `€X.XX` (cents → euros formatting)
- Time slots displayed as `HH:MM – HH:MM`
- Status badges: colored pills matching status (green, amber, blue)
- Loading states: skeleton placeholders
- Empty states: illustrated message + CTA
