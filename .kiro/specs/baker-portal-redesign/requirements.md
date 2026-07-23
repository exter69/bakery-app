# Requirements Document

## Introduction

Ce document décrit les exigences pour la refonte du portail Pro du boulanger dans l'application "Mie & Beurre". La refonte transforme l'interface existante (tableaux HTML, labels anglais) en un portail français moderne avec des cartes, un kanban, et une gestion visuelle du stock et des paniers anti-gaspi. Aucun nouvel endpoint backend n'est nécessaire — seuls les composants frontend sont reconstruits.

## Glossary

- **Portal**: L'application web React du boulanger (pages sous `/dashboard`)
- **Sidebar**: Le panneau de navigation latéral gauche (composant DashboardLayout)
- **KanbanBoard**: Le tableau à 4 colonnes pour la gestion des commandes par statut
- **OrderCard**: La carte visuelle représentant une commande individuelle dans le kanban
- **ProductCard**: La carte visuelle représentant un produit dans l'inventaire
- **StockStepper**: Le contrôle −/+ permettant de modifier la quantité en stock
- **FilterChips**: Les boutons-pilules de filtrage par catégorie ou type
- **BundleComposer**: L'interface split-panel de composition des paniers du soir
- **StatCard**: La carte KPI affichant une métrique clé sur le tableau de bord
- **DayToggles**: Les boutons circulaires L/M/M/J/V/S/D pour la disponibilité par jour

## Requirements

### Requirement 1: Navigation sidebar en français

**User Story:** En tant que boulanger, je veux voir la navigation en français avec les bons libellés, afin que l'interface corresponde à ma langue de travail.

#### Acceptance Criteria

1. THE Sidebar SHALL display navigation labels in French: "Tableau de bord", "Commandes", "Menu & stock", "Paniers du soir", "Statistiques", "Boutique"
2. WHEN the "Commandes" nav item has pending orders, THE Sidebar SHALL display a numeric badge showing the count of confirmed orders
3. THE Sidebar SHALL display the bakery name, avatar, and current open/closed status with hours in the footer
4. WHEN a nav item is active, THE Sidebar SHALL highlight it with a blue pill shape using the accent color (#4b8fe8)

---

### Requirement 2: Tableau de bord du matin (Overview)

**User Story:** En tant que boulanger, je veux voir un résumé de ma journée dès l'ouverture du portail, afin de savoir immédiatement combien de commandes préparer, quel est mon stock, et si des paniers du soir sont à créer.

#### Acceptance Criteria

1. WHEN the overview page loads, THE Portal SHALL display a personalized greeting with the baker's name and today's date in French
2. WHEN the overview page loads, THE Portal SHALL display KPI stat cards showing: today's order count, next scheduled pickup/delivery time, and today's revenue
3. WHEN there are orders with status "confirmed", THE Portal SHALL display them in an "À préparer maintenant" section as OrderCards
4. WHEN any product has stock at or below a low threshold, THE Portal SHALL display a "Stock faible" warning section listing those product names
5. THE Portal SHALL display a golden anti-gaspi CTA card showing estimated unsold value with a link to the bundle composer
6. THE Portal SHALL include a shop open/closed toggle in the overview header

---

### Requirement 3: Gestion des commandes en kanban

**User Story:** En tant que boulanger, je veux gérer mes commandes du jour sur un tableau kanban à 4 colonnes, afin de suivre visuellement l'avancement de chaque commande.

#### Acceptance Criteria

1. WHEN the orders page loads, THE KanbanBoard SHALL display exactly 4 columns labeled: "À PRÉPARER", "EN PRÉPARATION", "PRÊT", "REMIS / LIVRÉ"
2. WHEN orders are fetched, THE KanbanBoard SHALL assign each order to the column matching its status: confirmed → "À préparer", preparing → "En préparation", ready → "Prêt", delivered → "Remis / Livré"
3. WHEN a user drags an OrderCard to an adjacent column, THE KanbanBoard SHALL update the order status via the API and move the card to the new column
4. IF a user attempts to drag an OrderCard to a non-adjacent column, THEN THE KanbanBoard SHALL reject the move, snap the card back, and display a toast explaining valid transitions
5. WHEN a filter chip is selected (Livraison, Retrait, or Toutes), THE KanbanBoard SHALL display only orders matching that delivery type
6. THE OrderCard SHALL display: order time, item summary, delivery type badge ("livraison"/"retrait"), and an action button matching the next valid status transition
7. WHILE an order is in "En préparation" status, THE OrderCard SHALL display a blue left-border accent
8. WHEN an order moves to "Prêt" status, THE Portal SHALL trigger a customer notification via the existing API

---

### Requirement 4: Gestion des produits en cartes

**User Story:** En tant que boulanger, je veux gérer mon menu et mon stock via des cartes produit avec édition inline, afin de mettre à jour mes quantités rapidement sans ouvrir de formulaire séparé.

#### Acceptance Criteria

1. WHEN the products page loads, THE Portal SHALL display products as cards grouped by category (viennoiseries, pains, pâtisseries)
2. THE ProductCard SHALL display: photo, product name, description, price in euros, allergen tags, and a stock stepper
3. WHEN a user clicks the + or − button on a StockStepper, THE Portal SHALL immediately update the stock quantity via the API
4. WHILE a product's stock is at or below the low threshold, THE StockStepper SHALL display in red (danger styling)
5. WHEN a user clicks the visibility toggle on a ProductCard, THE Portal SHALL toggle the product between "en vente" and "masqué" states via the API
6. WHILE a product is in "masqué" state, THE ProductCard SHALL appear visually dimmed
7. WHEN a category filter chip is selected, THE Portal SHALL display only products belonging to that category
8. WHEN a user clicks "+ Nouveau produit", THE Portal SHALL open a product creation form
9. THE Portal SHALL display day-availability toggles (L, M, M, J, V, S, D) for scheduling product availability
10. THE Portal SHALL display a note: "le stock se remet à zéro chaque soir ↺"

---

### Requirement 5: Composition des paniers du soir (Bundles)

**User Story:** En tant que boulanger, je veux composer et publier des paniers anti-gaspi à partir de mes invendus du jour, afin de réduire le gaspillage tout en générant du chiffre d'affaires en fin de journée.

#### Acceptance Criteria

1. WHEN the bundles page loads, THE BundleComposer SHALL display a list of today's products with their remaining stock ("reste X")
2. WHEN a user checks a product checkbox, THE BundleComposer SHALL add that product to the bundle and update the live preview
3. WHEN a user adjusts the quantity stepper for a selected product, THE BundleComposer SHALL update the preview and price calculation
4. IF a user sets a product quantity exceeding its remaining stock, THEN THE BundleComposer SHALL reject the value and cap it at the remaining stock
5. THE BundleComposer SHALL display a live client preview panel showing: bundle name, pickup time, selected items with quantities, and crossed-out original price with discounted price
6. WHEN at least one product is selected, THE BundleComposer SHALL ensure the discounted price is strictly less than the sum of original item prices
7. WHEN a user adjusts the pickup time window, THE BundleComposer SHALL update the preview with the new pickup time
8. WHEN a user clicks "Publier les paniers", THE BundleComposer SHALL publish the bundle via the existing reservation API
9. WHILE no products are selected, THE BundleComposer SHALL disable the "Publier les paniers" button
10. THE BundleComposer SHALL allow setting a basket count (number of identical baskets to publish) with a minimum of 1

---

### Requirement 6: Composants partagés (FilterChips, StatCard, StockStepper)

**User Story:** En tant que développeur, je veux des composants UI réutilisables et bien testés, afin d'assurer la cohérence visuelle et de faciliter la maintenance.

#### Acceptance Criteria

1. WHEN a FilterChips component renders, THE FilterChips SHALL display a horizontal row of chip buttons with the active chip filled in blue and inactive chips outlined
2. WHEN a user clicks a chip, THE FilterChips SHALL select that chip and deselect the previously active one (single selection)
3. THE StatCard SHALL display a large value, a muted label above, a subtitle below, and an optional colored badge
4. WHEN the StockStepper receives increment or decrement actions, THE StockStepper SHALL keep the value within the configured [min, max] bounds
5. WHEN the StockStepper `danger` prop is true, THE StockStepper SHALL render with red styling

---

### Requirement 7: Gestion des erreurs

**User Story:** En tant que boulanger, je veux voir des messages d'erreur clairs en français quand quelque chose ne fonctionne pas, afin de savoir quoi faire sans perdre mon travail en cours.

#### Acceptance Criteria

1. IF an API call fails during page load, THEN THE Portal SHALL display an inline error message in French and a "Réessayer" button
2. IF an API call fails during page load, THEN THE Portal SHALL retain any previously loaded data on screen
3. IF a stock update conflicts with a concurrent modification, THEN THE Portal SHALL display a stale-data warning and reload the product data from the API
4. IF a drag-and-drop operation fails at the API level, THEN THE Portal SHALL revert the card to its original column and display a toast error

---

### Requirement 8: Performance et réactivité

**User Story:** En tant que boulanger, je veux que le portail soit rapide et réactif même avec beaucoup de commandes, afin de ne pas perdre de temps pendant le rush du matin.

#### Acceptance Criteria

1. WHEN the orders page loads, THE KanbanBoard SHALL fetch only today's orders (not historical data)
2. THE Portal SHALL lazy-load product images with placeholder elements to avoid layout shift
3. WHILE a drag operation is in progress on the KanbanBoard, THE Portal SHALL use React.memo on OrderCard and ProductCard to prevent unnecessary re-renders
4. WHEN the bundle composer recalculates prices on quantity change, THE BundleComposer SHALL debounce the calculation to avoid excessive recomputation
