# Requirements Document

## Introduction

End-to-end testing suite using Playwright for the bakery ordering application (Mie et Beurre). The test suite validates full user flows across the React frontend and Go backend running together, covering customer ordering journeys, baker management workflows, authentication, responsive behavior, allergen information display, and internationalization.

## Glossary

- **Test_Suite**: The Playwright-based end-to-end test framework configured to run against the full application stack
- **Application**: The bakery ordering app consisting of the React frontend (port 5173) and Go backend (port 8080) running together
- **Customer**: A user with role "customer" (seeded accounts: alice/customer123, bob/customer123)
- **Baker**: A user with role "seller" (seeded accounts: baker_jean/baker123, baker_marie/baker123)
- **Admin**: A user with role "admin" (seeded account: admin/admin123)
- **Guest**: An unauthenticated visitor browsing the application without logging in
- **Dashboard**: The baker portal at /dashboard with sub-pages for bakery, products, schedule, orders, and reservations management
- **Product_Selection_Modal**: The modal dialog on the bakery detail page where customers select products for an order
- **Registration_Code**: The code required for baker registration (seeded value: DEMO1234)
- **Seeded_Data**: Demo data loaded on application startup including 3 bakeries, 15 products, 3 orders, 2 reservations, 2 recurring orders, and test user accounts

## Requirements

### Requirement 1: Playwright Test Infrastructure Setup

**User Story:** As a developer, I want Playwright configured for the bakery app, so that I can run end-to-end tests against the full stack.

#### Acceptance Criteria

1. THE Test_Suite SHALL be configured with Playwright installed as a dev dependency in the frontend project
2. THE Test_Suite SHALL include a Playwright configuration file that targets the Application running on localhost:5173
3. THE Test_Suite SHALL provide a test script in package.json to execute all end-to-end tests
4. THE Test_Suite SHALL use a global setup mechanism to ensure the Application backend and frontend are running before tests execute
5. THE Test_Suite SHALL include reusable authentication helper utilities that log in as Customer, Baker, or Admin via the API and store session state

### Requirement 2: Customer Order Flow

**User Story:** As a QA engineer, I want to verify the full customer ordering journey, so that I can confirm customers can successfully place orders.

#### Acceptance Criteria

1. WHEN a Customer navigates to /bakeries, THE Test_Suite SHALL verify that the bakery list page displays all seeded bakeries
2. WHEN a Customer clicks on a bakery card, THE Test_Suite SHALL verify navigation to the bakery detail page at /bakeries/:id
3. WHEN a Customer clicks the order action on the bakery detail page, THE Test_Suite SHALL verify that the Product_Selection_Modal opens
4. WHEN a Customer selects products and submits the order form with a valid date and time, THE Test_Suite SHALL verify that the order is created and the Customer receives confirmation
5. WHEN a Customer views the /schedule page after placing an order, THE Test_Suite SHALL verify the new order appears in the order list

### Requirement 3: Customer Reservation Flow

**User Story:** As a QA engineer, I want to verify the reservation pickup flow, so that I can confirm customers can make reservations.

#### Acceptance Criteria

1. WHEN a Customer initiates a reservation on a bakery detail page, THE Test_Suite SHALL verify the reservation form is displayed
2. WHEN a Customer submits a reservation with valid product selections and a pickup time, THE Test_Suite SHALL verify that the reservation is created successfully
3. WHEN a Customer views the /schedule page after making a reservation, THE Test_Suite SHALL verify the reservation appears in the schedule

### Requirement 4: Customer Order Management

**User Story:** As a QA engineer, I want to verify customers can manage their orders, so that I can confirm the cancellation flow works.

#### Acceptance Criteria

1. WHEN a logged-in Customer navigates to /schedule, THE Test_Suite SHALL verify existing orders are displayed
2. WHEN a Customer cancels an order from the /schedule page, THE Test_Suite SHALL verify the order is removed from the list
3. WHEN a Customer cancels an order, THE Test_Suite SHALL verify a confirmation prompt is shown before deletion

### Requirement 5: Guest Browsing Restrictions

**User Story:** As a QA engineer, I want to verify guest access controls, so that I can confirm unauthenticated users have limited access.

#### Acceptance Criteria

1. WHEN a Guest navigates to /bakeries, THE Test_Suite SHALL verify the bakery list is visible without authentication
2. WHEN a Guest navigates to /bakeries/:id, THE Test_Suite SHALL verify the bakery detail page is visible without authentication
3. WHEN a Guest attempts to access /schedule, THE Test_Suite SHALL verify the Guest is redirected to /login
4. WHEN a Guest attempts to access /recurring, THE Test_Suite SHALL verify the Guest is redirected to /login
5. WHEN a Guest attempts to access /dashboard, THE Test_Suite SHALL verify the Guest is redirected to /login

### Requirement 6: Baker Dashboard and Product Management

**User Story:** As a QA engineer, I want to verify baker dashboard functionality, so that I can confirm bakers can manage their products.

#### Acceptance Criteria

1. WHEN a Baker logs in and navigates to /dashboard, THE Test_Suite SHALL verify the dashboard overview page loads with bakery statistics
2. WHEN a Baker navigates to /dashboard/products, THE Test_Suite SHALL verify the product list for their bakery is displayed
3. WHEN a Baker adds a new product with name, price, and category, THE Test_Suite SHALL verify the product appears in the product list
4. WHEN a Baker edits an existing product's price, THE Test_Suite SHALL verify the updated price is reflected in the product list
5. WHEN a Baker deletes a product, THE Test_Suite SHALL verify the product is removed from the list

### Requirement 7: Baker Order Processing

**User Story:** As a QA engineer, I want to verify bakers can process orders, so that I can confirm order status management works.

#### Acceptance Criteria

1. WHEN a Baker navigates to /dashboard/orders, THE Test_Suite SHALL verify pending orders are displayed
2. WHEN a Baker changes an order status (e.g., from pending to confirmed), THE Test_Suite SHALL verify the status update is reflected in the order list

### Requirement 8: Baker Registration Flow

**User Story:** As a QA engineer, I want to verify the baker registration process, so that I can confirm new bakers can sign up with a valid code.

#### Acceptance Criteria

1. WHEN a user navigates to /register, THE Test_Suite SHALL verify the registration form is displayed
2. WHEN a user submits the registration form with valid credentials and the Registration_Code DEMO1234, THE Test_Suite SHALL verify the account is created and the user is redirected to the login page or dashboard
3. WHEN a user submits the registration form with an invalid registration code, THE Test_Suite SHALL verify an error message is displayed

### Requirement 9: Authentication Error Handling

**User Story:** As a QA engineer, I want to verify authentication error cases, so that I can confirm the app handles invalid credentials gracefully.

#### Acceptance Criteria

1. WHEN a user submits the login form with an invalid username or password, THE Test_Suite SHALL verify an error message is displayed on the login page
2. WHEN a user submits the login form with empty fields, THE Test_Suite SHALL verify form validation prevents submission or shows validation errors
3. WHEN a user submits the login form with valid credentials, THE Test_Suite SHALL verify the user is redirected away from the login page

### Requirement 10: Protected Route Redirects

**User Story:** As a QA engineer, I want to verify protected routes redirect unauthenticated users, so that I can confirm route guards function correctly.

#### Acceptance Criteria

1. WHEN an unauthenticated user navigates directly to /schedule, THE Test_Suite SHALL verify a redirect to /login occurs
2. WHEN an unauthenticated user navigates directly to /dashboard, THE Test_Suite SHALL verify a redirect to /login occurs
3. WHEN an unauthenticated user navigates directly to /recurring, THE Test_Suite SHALL verify a redirect to /login occurs

### Requirement 11: Responsive Layout Verification

**User Story:** As a QA engineer, I want to verify the app works on mobile viewports, so that I can confirm responsive design functions correctly.

#### Acceptance Criteria

1. WHILE the viewport is set to a mobile width (375px), THE Test_Suite SHALL verify the home page renders without horizontal overflow
2. WHILE the viewport is set to a mobile width (375px), THE Test_Suite SHALL verify the bakery list page displays bakery cards in a single-column layout
3. WHILE the viewport is set to a mobile width (375px), THE Test_Suite SHALL verify navigation is accessible (via a hamburger menu or equivalent mobile navigation pattern)
4. WHILE the viewport is set to a mobile width (375px), THE Test_Suite SHALL verify the Customer order flow can be completed from bakery selection through order submission

### Requirement 12: Allergen Indicator Interactions

**User Story:** As a QA engineer, I want to verify allergen indicator UI interactions, so that I can confirm tooltip and modal behaviors work.

#### Acceptance Criteria

1. WHEN a Customer hovers over an allergen indicator icon on a product, THE Test_Suite SHALL verify a tooltip with allergen name is displayed
2. WHEN a Customer clicks an allergen indicator icon, THE Test_Suite SHALL verify an allergen detail modal opens with full allergen information
3. WHEN a Customer closes the allergen detail modal, THE Test_Suite SHALL verify the modal is no longer visible

### Requirement 13: Internationalization Language Switching

**User Story:** As a QA engineer, I want to verify language switching works, so that I can confirm i18n content updates correctly.

#### Acceptance Criteria

1. WHEN a user switches the language using the language selector, THE Test_Suite SHALL verify that page content updates to the selected language
2. WHEN a user switches language on the home page, THE Test_Suite SHALL verify that navigation labels, headings, and button text change to the target language
3. WHEN a user switches language and navigates to another page, THE Test_Suite SHALL verify the selected language persists across navigation
