# Product Selection Modal — Design Spec (Option 6a)

Companion to `design-direction.md` (same design language: warm artisan palette, Caveat + Patrick Hand, 1.5px ink borders `#3a2e22`, offset solid shadows). This REPLACES the darkened-page "selection mode" (`ProductSelectionOverlay`) described in `design-direction.md` §3 step 2. Kiro: implement as a new `ProductSelectionModal.tsx` + `.css`; the order/reservation side panels now open this modal via their "✚ Sélectionner des produits" button.

## 1. Modal shell

- Opens over the bakery detail page with a dim backdrop `rgba(58,46,34,.45)`; click backdrop or ✕ closes (state preserved while the parent panel stays open).
- Container: centered, `width: min(960px, 92vw)`, `height: min(680px, 88vh)`, background `#fffdf8`, `1.5px solid #3a2e22`, `border-radius: 16px`, shadow `4px 4px 0 rgba(58,46,34,.25)`, `overflow: hidden`.
- Header bar: `#f7f1e5` bg, ink bottom border. Title (Caveat ~22px): "Composer votre commande — {bakery name}". ✕ button right.
- Enter/exit animation: scale .97→1 + fade, 200–300ms.

## 2. Two-pane layout

Flex row filling the modal below the header:

```
┌──────────────────────────────┬──────────────┐
│ PRODUCTS (flex:1, scrolls)   │ SETUP (fixed)│
└──────────────────────────────┴──────────────┘
```

### 2.1 Products pane — INDEPENDENT SCROLL (required)

- `flex: 1; min-width: 0; overflow-y: auto; overscroll-behavior: contain;`
- Only THIS pane scrolls. The modal shell, header, and setup column never move; the page behind must not scroll (lock `body` overflow while the modal is open).
- Sticky inside the scroll area, top: category filter chips row (Viennoiseries / Pains / Pâtisseries…), active chip accent-filled `#e8b04b`, background `#fffdf8` with a soft bottom fade so cards scroll under it cleanly.
- Product grid: `display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 12px; padding: 14px; align-content: start;`

### 2.2 Product card — hover-expand behavior (core interaction)

At rest:
- Card: `#fffdf8` bg, ink border, `border-radius: 12px`, image on top (~64% of card height), then name + price.

On hover / focus-within:
- The top image SHRINKS (height ~64px → ~38px, `transition: 250ms ease`) to make room for details — the card grows slightly (`max-height` transition) but MUST NOT reflow the grid: expand via `transform: scale(1.04)` + internal height change with the card given `z-index: 2`, or reserve the expanded height with a fixed row height. Neighboring cards must not jump.
- Revealed details, sliding up under the image: short description ("Beurre AOP, cuit ce matin."), allergens line, then a quantity row: − / count / + steppers (pill chips, + accent-filled) and an "Ajouter" accent chip.
- Border becomes `2px solid #e8b04b` + offset shadow when hovered AND when the product has quantity > 0 (with a `×N` chip on the image corner at rest).
- Touch devices (no hover): first tap expands the card, second tap on +/Ajouter adds; tapping another card collapses the previous.

### 2.3 Setup column (right, fixed)

- `width: 280px; flex: none;` background `#f7f1e5`, `border-left: 1.5px solid #3a2e22`, padding 16px, flex column. NEVER scrolls with the products; if its own content overflows (long item list), only the line-items box gets `overflow-y: auto`.
- Top-to-bottom:
  1. "Votre commande" heading (Caveat ~20px)
  2. Mode toggle chips: **Livraison** / **Retrait** (active = accent fill). Livraison → pay online; Retrait → pay at counter (existing rules).
  3. Day + time-slot selector (ink-bordered dropdown on `#fffdf8`), validated against bakery opening hours.
  4. Line-items summary box: dashed `#a08b70` border, `#fffdf8` bg — "2× Pain au chocolat — €3.20" rows, dashed divider, bold Total. Rows removable (small ✕ on hover).
  5. Spacer, then pinned bottom CTA: "Valider →" accent-filled pill. Disabled until ≥1 product and a valid slot.

## 3. Behavior & validation

- Confirming closes the modal and returns the selection + slot to the parent order/reservation panel (same invariants as today: delivery = online payment; reservation = Confirmed status, OnSpot payment, ≥1 item, within opening hours).
- Esc closes; focus is trapped inside the modal; restore focus to the opener button on close.
- Keyboard: arrow keys move between cards, Enter expands, +/− adjust quantity.
- Scroll position of the products pane is preserved if the user closes and reopens during the same order draft.

## 4. Responsive

| Breakpoint | Layout |
|---|---|
| ≥900px | Two panes side by side as above |
| <900px | Modal goes full-screen; setup column collapses into a sticky bottom tray (dark ink `#3a2e22`, cream text — style of wireframe 1e) showing mode, slot, total, "Valider →"; tapping the tray expands the full setup |
| Grid | minmax drops to 140px; hover-expand becomes tap-expand |
