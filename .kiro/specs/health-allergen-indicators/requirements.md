# Requirements Document

## Introduction

This feature introduces health scores and allergen indicators on bakery products. Bakers assign a numeric health score and select allergens when creating or editing products in their dashboard. Customers see a health score displayed on product cards and in the product description, and an allergen icon on product cards that reveals allergen details on hover/click. Additionally, a floating allergen information icon at the bottom of the page provides general allergen education (EU-regulated allergens list, explanations) independent of any specific product.

## Glossary

- **Product_Card**: A visual card component that displays a product's image, name, price, and indicators in both the BakeryDetailPage and the ProductSelectionModal.
- **Allergen_Indicator**: A floating icon positioned at the bottom-right of a Product_Card that communicates allergen information. On hover it displays a text summary; on click it opens a detail modal.
- **Health_Score**: A numeric value from 1 to 5 assigned by the baker to indicate how healthy a product is (1 = least healthy, 5 = healthiest), displayed on Product_Cards and in the product description.
- **Allergen_Detail_Modal**: A modal dialog opened by clicking the Allergen_Indicator, showing the full list of allergens for that product with their names and descriptions.
- **Allergen_Info_Icon**: A floating icon fixed at the bottom of the page (sticky footer button) that provides general allergen education. On click it opens the Allergen_Info_Modal.
- **Allergen_Info_Modal**: A modal dialog opened by clicking the Allergen_Info_Icon, explaining what allergens are, listing the 14 EU-regulated allergens with descriptions, and providing general safety information. This modal is not product-specific.
- **Baker_Portal**: The dashboard application used by bakery owners to manage their bakery, products, and orders (system sans-serif theme).
- **Customer_Portal**: The customer-facing application where customers browse bakeries, view menus, and place orders (artisan design theme).
- **Allergen**: A substance that may cause an allergic reaction. The system supports a predefined set of EU-regulated allergens: gluten, crustaceans, eggs, fish, peanuts, soy, dairy, nuts, celery, mustard, sesame, sulphites, lupin, and molluscs.
- **Product_Management_Page**: The DashboardProducts page in the Baker_Portal where bakers create and edit products.

## Requirements

### Requirement 1: Product Data Model

**User Story:** As a baker, I want to associate allergen information and a health score with each product, so that customers with allergies or dietary preferences can make informed purchasing decisions.

#### Acceptance Criteria

1. THE Product data model SHALL include an allergens field that stores zero or more unique Allergen values from the predefined set of 14 EU-regulated allergens (gluten, crustaceans, eggs, fish, peanuts, soy, dairy, nuts, celery, mustard, sesame, sulphites, lupin, molluscs).
2. THE Product data model SHALL include a health_score field that stores a nullable integer value representing the product's health rating.
3. WHEN a product has no allergens specified, THE System SHALL treat the allergens field as an empty list and persist it as an empty array.
4. WHEN a product has no health score specified, THE System SHALL treat the health_score field as null and persist it as NULL in the database.
5. THE Database SHALL store allergen and health score data as part of the product record using a numbered database migration following the existing sequence.
6. IF a product record is created or updated with an allergen value not in the predefined set of 14 EU-regulated allergens, THEN THE System SHALL reject the operation and return an error indicating the invalid allergen value.
7. IF a product record is created or updated with a health_score value outside the range 1 to 5 (inclusive), THEN THE System SHALL reject the operation and return an error indicating the invalid health score value.
8. WHEN the migration is applied, THE System SHALL default the allergens field to an empty list and the health_score field to NULL for all existing product records.

### Requirement 2: Baker Product Input

**User Story:** As a baker, I want to select allergens and assign a health score when creating or editing a product, so that I can communicate accurate dietary information to customers.

#### Acceptance Criteria

1. WHEN a baker opens the product creation form in the Product_Management_Page, THE Baker_Portal SHALL display a multi-select input for allergens showing all 14 EU-regulated allergen options: gluten, crustaceans, eggs, fish, peanuts, soy, dairy, nuts, celery, mustard, sesame, sulphites, lupin, and molluscs.
2. WHEN a baker opens the product creation form in the Product_Management_Page, THE Baker_Portal SHALL display a numeric input field for health score that accepts integer values from 1 to 5, with a label explaining the scale (1 = least healthy, 5 = healthiest).
3. WHEN a baker edits an existing product, THE Baker_Portal SHALL pre-populate the allergen multi-select with the product's currently saved allergen values and the health score input with the product's current health score within 2 seconds of the form loading.
4. WHEN a baker saves a product with selected allergens and/or a health score, THE Baker_Portal SHALL persist the allergen selections and health score to the backend and display a success confirmation to the baker within 3 seconds.
5. WHEN a baker deselects all allergens from a product, THE Baker_Portal SHALL save the product with an empty allergens list.
6. WHEN a baker clears the health score input, THE Baker_Portal SHALL save the product with a null health score value.
7. IF the backend fails to save allergen or health score data, THEN THE Baker_Portal SHALL display an error message indicating the save failed and SHALL retain the baker's unsaved selections in the form so no input is lost.
8. THE Baker_Portal SHALL allow a baker to select any combination of allergens (0 to 14) independently of the health score, and the health score SHALL be optional (can be left blank).
9. IF a baker enters a health score value outside the range 1 to 5, THEN THE Baker_Portal SHALL display a client-side validation error before submitting the form.

### Requirement 3: Allergen Indicator Display on Product Cards

**User Story:** As a customer, I want to see an allergen indicator on product cards, so that I can quickly identify products that may contain allergens.

#### Acceptance Criteria

1. WHEN a product has one or more allergens in its allergen list, THE Customer_Portal SHALL display a single Allergen_Indicator icon positioned at the bottom-right corner of the Product_Card, overlapping the card edge by no more than 50% of the icon's dimensions.
2. WHEN a product has an empty allergen list or no allergen data, THE Customer_Portal SHALL not render the Allergen_Indicator icon on that Product_Card.
3. THE Allergen_Indicator SHALL be rendered at a fixed size of 24×24 pixels on desktop viewports and 20×20 pixels on viewports below 768px width.
4. THE Allergen_Indicator SHALL be visible on product cards displayed in the ProductSelectionModal grid cards.
5. THE Allergen_Indicator SHALL be visible on product cards displayed in the BakeryDetailPage desktop product grid and mobile product rows.
6. THE Allergen_Indicator SHALL include an accessible label (e.g., aria-label or tooltip) stating "Contains allergens" so that screen readers announce the presence of allergens.
7. WHEN the customer hovers over or focuses the Allergen_Indicator, THE Customer_Portal SHALL display a tooltip listing the product's allergen names, with each allergen separated by a comma, within 200 milliseconds of the hover or focus event.

### Requirement 4: Allergen Hover Summary

**User Story:** As a customer, I want to see a quick summary of allergens when hovering over the allergen indicator, so that I can get information without extra clicks.

#### Acceptance Criteria

1. WHEN a customer hovers over the Allergen_Indicator, THE Customer_Portal SHALL display a tooltip within 150 milliseconds containing a comma-separated list of the product's allergen names.
2. WHEN the customer moves the pointer away from the Allergen_Indicator, THE Customer_Portal SHALL hide the tooltip immediately (within one animation frame).
3. THE tooltip text SHALL be displayed in the customer's currently selected language (EN, FR, or NL).
4. THE tooltip SHALL be rendered above parent overflow boundaries so that no portion of its text is clipped or hidden by the Product_Card container.
5. WHEN a customer focuses the Allergen_Indicator via keyboard navigation, THE Customer_Portal SHALL display the same tooltip as on hover, and SHALL hide it when focus leaves the element.
6. WHILE the tooltip is displayed, THE Customer_Portal SHALL expose the tooltip content to assistive technologies using an appropriate accessible name or description associated with the Allergen_Indicator.

### Requirement 5: Allergen Detail Modal (Per-Product)

**User Story:** As a customer, I want to see detailed allergen information in a modal, so that I can review the full allergen breakdown for a product.

#### Acceptance Criteria

1. WHEN a customer clicks the Allergen_Indicator, THE Customer_Portal SHALL open the Allergen_Detail_Modal and move keyboard focus to the modal container.
2. THE Allergen_Detail_Modal SHALL display the product name as the modal title.
3. THE Allergen_Detail_Modal SHALL list each of the product's allergens by its translated name, ordered alphabetically in the active language.
4. WHEN the customer clicks outside the Allergen_Detail_Modal or presses Escape, THE Customer_Portal SHALL close the modal and return keyboard focus to the Allergen_Indicator that triggered it.
5. THE Allergen_Detail_Modal content SHALL be displayed in the customer's currently selected language (EN, FR, or NL).
6. WHEN a customer clicks the Allergen_Indicator, THE click event SHALL NOT propagate to the parent Product_Card (preventing unintended product selection or expansion).
7. WHILE the Allergen_Detail_Modal is open, THE Customer_Portal SHALL trap keyboard focus within the modal so that Tab and Shift+Tab cycle only through focusable elements inside it.
8. IF the Allergen_Detail_Modal is opened while the ProductSelectionModal is already open, THEN THE Customer_Portal SHALL display the Allergen_Detail_Modal above the ProductSelectionModal without closing it, and closing the Allergen_Detail_Modal SHALL return the customer to the ProductSelectionModal.

### Requirement 6: Health Score Display on Product Cards

**User Story:** As a customer, I want to see a health score on product cards, so that I can quickly identify how healthy a product is.

#### Acceptance Criteria

1. WHEN a product has a health_score value (not null), THE Customer_Portal SHALL display the Health_Score on the Product_Card as a numeric indicator (e.g., "3/5" or a visual scale from 1 to 5).
2. WHEN a product has no health_score (null value), THE Customer_Portal SHALL not render the Health_Score indicator on the Product_Card, leaving no empty space or placeholder.
3. THE Health_Score SHALL also be displayed in the product description area when the product is viewed in detail or in the ProductSelectionModal expanded state.
4. THE Health_Score display SHALL be visible on product cards in the ProductSelectionModal without requiring the card to be in expanded/hovered state.
5. THE Health_Score display SHALL be visible on product cards in the BakeryDetailPage.
6. THE Health_Score display SHALL include an accessible label (e.g., aria-label) conveying the score meaning, such as "Health score: 3 out of 5", so that screen readers can announce it.

### Requirement 7: Page-Level Allergen Information Icon

**User Story:** As a customer, I want access to general allergen education from any point on the page, so that I can learn about allergens without needing to find a specific product first.

#### Acceptance Criteria

1. THE Customer_Portal SHALL display a floating Allergen_Info_Icon fixed at the bottom of the viewport (sticky footer position), visible on all pages that display products (BakeryDetailPage, ProductSelectionModal context).
2. THE Allergen_Info_Icon SHALL remain visible while scrolling the page content and SHALL not overlap critical interactive elements such as "Add to cart" buttons or navigation.
3. WHEN a customer clicks the Allergen_Info_Icon, THE Customer_Portal SHALL open the Allergen_Info_Modal and move keyboard focus to the modal container.
4. THE Allergen_Info_Modal SHALL display a title such as "Allergen Information" (translated to the active language).
5. THE Allergen_Info_Modal SHALL contain an introductory paragraph explaining what allergens are and why they matter for food safety.
6. THE Allergen_Info_Modal SHALL list all 14 EU-regulated allergens with their translated names and descriptions.
7. WHEN the customer clicks outside the Allergen_Info_Modal or presses Escape, THE Customer_Portal SHALL close the modal and return keyboard focus to the Allergen_Info_Icon.
8. WHILE the Allergen_Info_Modal is open, THE Customer_Portal SHALL trap keyboard focus within the modal so that Tab and Shift+Tab cycle only through focusable elements inside it.
9. THE Allergen_Info_Icon SHALL include an accessible label (e.g., aria-label) stating "Allergen information" so that screen readers can identify its purpose.
10. THE Allergen_Info_Icon SHALL be rendered at a size of 40×40 pixels on desktop viewports and 36×36 pixels on viewports below 768px width.

### Requirement 8: Internationalization of Allergen and Health Data

**User Story:** As a customer browsing in my preferred language, I want allergen labels, health score labels, and allergen information to appear in my selected language, so that I can understand the information without language barriers.

#### Acceptance Criteria

1. THE System SHALL provide translations for all 14 allergen names (gluten, crustaceans, eggs, fish, peanuts, soy, dairy, nuts, celery, mustard, sesame, sulphites, lupin, molluscs) in EN, FR, and NL, totalling 42 translation entries.
2. THE System SHALL provide translations for allergen descriptions in EN, FR, and NL, where each description explains in no more than 150 characters what the allergen covers (e.g., which food products may contain it), totalling 42 translation entries.
3. THE System SHALL provide translations for the Health_Score label (e.g., "Health score") and scale explanation in EN, FR, and NL.
4. THE System SHALL provide translations for the Allergen_Info_Modal title, introductory paragraph, and all allergen entries in EN, FR, and NL.
5. WHEN the customer switches language, THE Customer_Portal SHALL update all displayed allergen names, health score labels, and allergen information modal content to the selected language without requiring a page reload or navigation.
6. IF a translation key is missing for the selected language, THEN THE System SHALL display the English (EN) translation as a fallback, and if the English translation is also unavailable, display the translation key identifier.
7. WHEN the customer selects a language, THE System SHALL persist the language preference in localStorage so that subsequent visits default to the last selected language.

### Requirement 9: Backend API for Allergen and Health Score Data

**User Story:** As a developer, I want the backend API to support allergen and health score data in product endpoints, so that the frontend can read and write this information.

#### Acceptance Criteria

1. WHEN a baker creates a product via the API, THE API SHALL accept an optional `allergens` field in the request body as a JSON array of strings with a maximum of 20 elements.
2. WHEN a baker creates a product via the API, THE API SHALL accept an optional `health_score` field in the request body as a JSON integer value between 1 and 5 inclusive, or null.
3. WHEN a baker updates a product via the API, THE API SHALL accept an optional `allergens` field in the request body as a JSON array of strings with a maximum of 20 elements.
4. WHEN a baker updates a product via the API, THE API SHALL accept an optional `health_score` field in the request body as a JSON integer value between 1 and 5 inclusive, or null.
5. IF the `allergens` field is omitted from an update request, THEN THE API SHALL leave the existing stored allergen values unchanged.
6. IF the `health_score` field is omitted from an update request, THEN THE API SHALL leave the existing stored health score value unchanged.
7. IF the `allergens` field is provided as an empty array in a create or update request, THEN THE API SHALL store an empty array for allergens, clearing any previously stored values.
8. IF the `health_score` field is provided as null in a create or update request, THEN THE API SHALL store NULL for the health score, clearing any previously stored value.
9. WHEN a client fetches a product or menu, THE API SHALL include the `allergens` field in the response as a JSON array of strings (empty array `[]` when no allergens) and the `health_score` field as a JSON integer or null.
10. IF a request contains an allergen value not in the predefined set defined in the Go domain package, THEN THE API SHALL return a 400 status code with an error response indicating which value is invalid.
11. IF a request contains a health_score value outside the range 1 to 5, THEN THE API SHALL return a 400 status code with an error response indicating the valid range.
12. IF a request contains more than 20 elements in the `allergens` array, THEN THE API SHALL return a 400 status code with an error response indicating the maximum allowed count was exceeded.
