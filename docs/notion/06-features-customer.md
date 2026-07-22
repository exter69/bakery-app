# 🛒 Customer Features

## Browse Bakeries

- View all bakeries in a list or map view
- **Geolocation sorting**: bakeries sorted by distance when location is shared
- Filter by currently open
- Each card shows: photo, name, today's schedule, distance
- Interactive Leaflet map with markers

---

## View Bakery Detail

- Full bakery info: description, address, schedule, map
- **Menu** grouped by category (Breads, Viennoiseries, Pastries)
- Product cards show: name, description, price, allergen badges, health score
- Travel time indicator (based on geolocation)
- Minimum delivery amount displayed

---

## Place Delivery Order

1. Click "Order" on a bakery detail page
2. **ProductSelectionModal** opens:
   - Browse products by category
   - Set quantity per item
   - See allergen indicators per product
   - Running total displayed
3. Choose delivery day (from bakery schedule)
4. Choose time slot within opening hours
5. Select one-time or weekly/bi-weekly (recurring)
6. Confirm — pays online
7. Must meet minimum delivery amount (€10 default)

---

## Make Reservation (Pickup)

1. Click "Reserve" on bakery detail
2. Select products (same modal)
3. Choose pickup day and time slot
4. Confirm — payment at counter (on-spot)
5. No minimum amount for reservations

---

## Manage Orders (Schedule Page)

- View all upcoming orders and reservations
- Grouped by day of week
- See status: Confirmed → Preparing → Ready → Delivered
- Cancel pending orders (before preparation starts)

---

## Recurring Orders

- Set up weekly or bi-weekly orders
- **Selection modes**:
  - **Fixed**: same items every time
  - **Bakery Choice**: baker picks for you
  - **Random Favorites**: system picks from your favorites
- Pause/resume recurring orders
- **Holiday Mode**: pause all recurring orders during vacation
  - Set holiday start/end dates
  - All orders auto-paused during range

---

## Quick Reserve Widget

- Available on the home page
- Select a bakery from dropdown
- Quickly reserve without navigating to bakery detail

---

## Allergen Information

- Per-product allergen badges (color-coded pills)
- Floating allergen info icon on bakery detail page
- **AllergenInfoModal**: explains all 14 allergens with descriptions
- **AllergenDetailModal**: detailed view per allergen
- Translated to all 3 supported languages

---

## Guest Mode

- Browse bakeries and menus without an account
- Must login/register to place orders or reservations
