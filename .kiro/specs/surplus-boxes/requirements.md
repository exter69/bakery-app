# Requirements Document

## Introduction

This feature enables bakeries to publish end-of-day surplus boxes ("paniers du soir") — discounted packages of unsold items available for customer pickup before closing. The goal is to reduce food waste ("anti-gaspi") while offering customers discounted products. Customers browse available bundles on a dedicated page or via a home page card, reserve a bundle for on-spot payment, and pick it up before the bakery closes. Bundles come in two types: "composé" (baker specifies exact items) and "surprise" (baker describes category/estimated value only). The feature includes real-time stock tracking, geolocation-based filtering, and community impact metrics.

## Glossary

- **Surplus_Bundle**: A discounted package of unsold bakery items published by a baker for end-of-day pickup. A bundle has a type (composé or surprise), a quantity, a pickup time window, and an expiration time tied to the bakery's closing.
- **Composé_Bundle**: A Surplus_Bundle where the baker specifies the exact items included (e.g., "2× croissant, 1× baguette tradition, 1 part de tarte du jour").
- **Surprise_Bundle**: A Surplus_Bundle where the baker provides only a category description and estimated value (e.g., "viennoiseries & pain du jour, valeur estimée €10+"), without listing specific items.
- **Bundle_Card**: A visual card component displaying a Surplus_Bundle's photo, name, type badge, bakery name, distance, pickup window, contents or value estimate, pricing, remaining stock, and a reservation button.
- **Bundle_Reservation**: A customer's claim on one Surplus_Bundle, confirmed for on-spot payment at pickup. Unretrieved reservations are released at the end of the pickup window.
- **Pickup_Window**: The time range during which a customer can collect a reserved Surplus_Bundle (e.g., 18:30–19:00). Defined by the baker when publishing.
- **Bundle_Page**: The dedicated full-page view at route `/paniers-du-soir` listing all available Surplus_Bundles with filtering, sorting, and map/list toggle.
- **Home_Bundle_Card**: A summary card on the home page showing nearby available bundles with a link to the Bundle_Page.
- **Reservation_Rail**: A sidebar panel on the Bundle_Page showing the customer's current Bundle_Reservation with pickup details and a confirmation button.
- **Impact_Card**: A community metrics card showing the total number of bundles saved and estimated waste avoided.
- **Customer_Portal**: The customer-facing application where customers browse bakeries, view menus, and place orders.
- **Baker_Portal**: The dashboard application used by bakery owners to manage their bakery, products, and orders.
- **Bundle_API**: The backend API endpoints for creating, listing, reserving, and managing Surplus_Bundles.

## Requirements

### Requirement 1: Surplus Bundle Data Model

**User Story:** As a baker, I want to define surplus bundles with specific attributes, so that customers can see what is available for end-of-day pickup.

#### Acceptance Criteria

1. THE Surplus_Bundle data model SHALL include: id (UUID), bakery_id (references bakery), name (string, max 100 chars), type (enum: "compose" or "surprise"), photo_url (string), original_price (integer, cents), discounted_price (integer, cents), quantity_total (integer ≥ 1), quantity_remaining (integer ≥ 0), pickup_start_time (time), pickup_end_time (time), published_date (date), expires_at (timestamp), status (enum: "draft", "published", "expired", "sold_out"), and created_at/updated_at timestamps.
2. WHEN a Surplus_Bundle has type "compose", THE Surplus_Bundle data model SHALL include an items list containing one or more entries, each with a product reference or free-text description and a quantity.
3. WHEN a Surplus_Bundle has type "surprise", THE Surplus_Bundle data model SHALL include a description field (max 200 chars) and an estimated_value field (integer, cents) representing the approximate retail value of the contents.
4. THE Surplus_Bundle data model SHALL enforce that discounted_price is strictly less than original_price.
5. THE Surplus_Bundle data model SHALL enforce that pickup_start_time is before pickup_end_time.
6. THE Surplus_Bundle data model SHALL enforce that quantity_remaining is less than or equal to quantity_total.
7. THE Database SHALL store Surplus_Bundle data using a numbered migration following the existing goose sequence.

### Requirement 2: Bundle Publishing

**User Story:** As a baker, I want to publish surplus bundles each afternoon, so that customers can see what is available for pickup before closing.

#### Acceptance Criteria

1. WHEN a baker creates a Surplus_Bundle via the Bundle_API, THE System SHALL validate all required fields and store the bundle with status "draft".
2. WHEN a baker publishes a draft Surplus_Bundle, THE System SHALL change the status to "published" and make the bundle visible to customers.
3. WHEN a baker publishes a Surplus_Bundle, THE System SHALL set the expires_at timestamp to the bakery's closing time on the published_date.
4. WHEN the current time passes a published bundle's expires_at timestamp, THE System SHALL automatically change the bundle status to "expired".
5. WHEN all units of a published bundle are reserved, THE System SHALL change the bundle status to "sold_out".
6. THE Bundle_API SHALL require JWT authentication with seller or admin role for creating and publishing bundles.
7. IF a baker attempts to publish a bundle with missing required fields, THEN THE Bundle_API SHALL return a 400 status code with descriptive error messages.

### Requirement 3: Bundle Listing and Filtering

**User Story:** As a customer, I want to browse available surplus bundles with filters, so that I can find relevant bundles near me.

#### Acceptance Criteria

1. WHEN a customer visits the Bundle_Page, THE Customer_Portal SHALL display all published Surplus_Bundles with status "published" sorted by proximity to the customer's location.
2. WHEN a customer selects the "Retrait avant 19h" filter, THE Customer_Portal SHALL display only bundles with a pickup_end_time before 19:00.
3. WHEN a customer selects the "– de 500 m" filter, THE Customer_Portal SHALL display only bundles from bakeries within 500 meters of the customer's location.
4. WHEN a customer selects the "Surprise" filter, THE Customer_Portal SHALL display only bundles with type "surprise".
5. WHEN a customer selects the "Composé" filter, THE Customer_Portal SHALL display only bundles with type "compose".
6. THE Bundle_Page SHALL support toggling between a list view and a map view (Carte).
7. WHEN a Surplus_Bundle has status "sold_out", THE Bundle_Card SHALL display an "épuisé" badge, dim the card visually, and show the text "revenez demain vers 17h".
8. THE Bundle_Card SHALL display: the bundle photo, name, type badge ("composé" or "surprise ★"), bakery name, distance from customer, pickup time window, contents list (for composé) or estimated value (for surprise), original price crossed out, discounted price, remaining stock badge ("reste N"), and a "Réserver" button.
9. THE Bundle_Page SHALL display a header with the title "Paniers du soir", an "anti-gaspi" badge, and subtitle "sauvez les invendus du jour — jusqu'à ~60%".
10. THE Bundle_API SHALL return bundles with bakery location data so the Customer_Portal can compute distance client-side.
11. WHEN a customer has no geolocation available, THE Customer_Portal SHALL sort bundles by published_date descending and disable the distance filter.

### Requirement 4: Bundle Reservation

**User Story:** As a customer, I want to reserve a surplus bundle, so that I can pick it up before the bakery closes.

#### Acceptance Criteria

1. WHEN a customer clicks "Réserver" on a Bundle_Card, THE System SHALL decrement the bundle's quantity_remaining by one and create a Bundle_Reservation linked to the customer and bundle.
2. WHEN a Bundle_Reservation is created, THE Reservation_Rail SHALL display the reservation details: bundle name, quantity, total price, pickup window, and payment method ("paiement au comptoir").
3. WHEN a Bundle_Reservation is created, THE Reservation_Rail SHALL display a warning: "à récupérer ce soir — sinon le panier est libéré à [pickup_end_time]".
4. WHEN the customer clicks "Confirmer" in the Reservation_Rail, THE System SHALL confirm the Bundle_Reservation and transition its status to "confirmed".
5. IF a customer attempts to reserve a bundle with quantity_remaining equal to zero, THEN THE System SHALL reject the reservation and display a message indicating the bundle is sold out.
6. WHEN the pickup_end_time passes and a Bundle_Reservation has not been picked up, THE System SHALL release the reservation by incrementing the bundle's quantity_remaining and changing the reservation status to "released".
7. THE Bundle_API SHALL require JWT authentication with customer role for creating reservations.
8. THE Bundle_API SHALL enforce a maximum of one active Bundle_Reservation per customer per bundle at any time.
9. WHEN a customer cancels a Bundle_Reservation before pickup_end_time, THE System SHALL release the reservation by incrementing the bundle's quantity_remaining and changing the reservation status to "cancelled".
10. THE System SHALL update quantity_remaining in real-time via the existing WebSocket infrastructure so all connected clients see current availability.

### Requirement 5: Home Page Integration

**User Story:** As a customer, I want to see available surplus bundles on the home page, so that I can discover deals without navigating to the dedicated page.

#### Acceptance Criteria

1. THE Customer_Portal SHALL display a Home_Bundle_Card on the home page when published Surplus_Bundles are available.
2. THE Home_Bundle_Card SHALL display the title "Paniers du soir" with an "anti-gaspi" badge and subtitle "Les invendus du jour, à petit prix. Retrait avant la fermeture."
3. THE Home_Bundle_Card SHALL display the nearest available bundle in an expanded format showing full contents, pricing with strikethrough original price, and remaining stock.
4. THE Home_Bundle_Card SHALL display up to three additional bundles in a compact format showing photo, bakery name, distance, type badge, and discounted price.
5. THE Home_Bundle_Card SHALL include a "Voir tous les paniers →" link that navigates to the Bundle_Page.
6. WHEN no published Surplus_Bundles are available, THE Customer_Portal SHALL not render the Home_Bundle_Card on the home page.

### Requirement 6: Community Impact Metrics

**User Story:** As a customer, I want to see the community impact of surplus bundles, so that I am motivated to participate in reducing food waste.

#### Acceptance Criteria

1. THE Bundle_Page SHALL display an Impact_Card showing the total number of bundles saved (picked up) in the current month.
2. THE Impact_Card SHALL display an estimated weight of food waste avoided, calculated as total_bundles_saved multiplied by 0.5 kg.
3. THE Impact_Card SHALL display the metrics with community-oriented messaging (e.g., "Déjà N paniers sauvés 🌱" and "soit ~X kg de pain et viennoiseries évités à la poubelle ce mois-ci").
4. THE Bundle_API SHALL provide an endpoint returning community impact metrics (total bundles picked up in current month).

### Requirement 7: Bundle Expiration and Lifecycle

**User Story:** As a system operator, I want bundles to expire automatically, so that outdated bundles are not shown to customers.

#### Acceptance Criteria

1. WHEN the current time exceeds a published bundle's expires_at timestamp, THE System SHALL transition the bundle status to "expired" and remove the bundle from customer-facing listings.
2. WHEN a bundle expires, THE System SHALL release all pending (unconfirmed) Bundle_Reservations associated with that bundle and increment the quantity_remaining accordingly.
3. THE Bundle_Page footer SHALL display a note: "les paniers expirent à la fermeture de chaque boulangerie · publication chaque jour en fin d'après-midi".
4. THE System SHALL run expiration checks at a frequency of at least once per minute to ensure timely status transitions.

### Requirement 8: Real-Time Stock Updates

**User Story:** As a customer, I want to see real-time stock availability, so that I do not attempt to reserve a bundle that is already sold out.

#### Acceptance Criteria

1. WHEN a bundle's quantity_remaining changes (due to reservation or cancellation), THE System SHALL broadcast an update via WebSocket to all connected clients viewing the Bundle_Page.
2. WHEN the Customer_Portal receives a stock update via WebSocket, THE Bundle_Card SHALL update the remaining stock badge ("reste N") without requiring a page reload.
3. WHEN quantity_remaining reaches zero via a WebSocket update, THE Bundle_Card SHALL transition to the sold-out state (dimmed card, "épuisé" badge, disabled "Réserver" button).
4. IF a customer clicks "Réserver" and the server rejects due to a race condition (stock reached zero between display and click), THEN THE Customer_Portal SHALL display an error message indicating the bundle is no longer available and refresh the bundle's stock display.

### Requirement 9: Internationalization

**User Story:** As a customer browsing in my preferred language, I want all surplus bundle UI content to appear in my selected language, so that I can understand the information.

#### Acceptance Criteria

1. THE System SHALL provide translations for all static UI text on the Bundle_Page, Home_Bundle_Card, Reservation_Rail, and Impact_Card in EN, FR, and NL.
2. THE System SHALL provide translations for bundle type labels ("composé", "surprise"), filter labels, status badges ("épuisé"), action buttons ("Réserver", "Confirmer"), and informational messages in EN, FR, and NL.
3. WHEN the customer switches language, THE Customer_Portal SHALL update all surplus bundle UI text to the selected language without requiring a page reload.
4. THE System SHALL display bundle names, descriptions, and item lists in the language provided by the baker (no automatic translation of baker-authored content).

### Requirement 10: Pricing and Payment

**User Story:** As a customer, I want to see clear pricing and pay on-spot, so that I understand the deal and can pay when I pick up.

#### Acceptance Criteria

1. THE Bundle_Card SHALL display the original_price with a strikethrough and the discounted_price prominently.
2. THE Reservation_Rail SHALL display the total amount as the discounted_price of the reserved bundle.
3. THE Reservation_Rail SHALL indicate the payment method as "paiement au comptoir" (on-spot payment at pickup).
4. THE Surplus_Bundle data model SHALL enforce that discounted_price is greater than zero.
5. IF a baker attempts to set a discounted_price equal to or greater than original_price, THEN THE Bundle_API SHALL return a 400 status code with an error indicating the discounted price must be less than the original price.

