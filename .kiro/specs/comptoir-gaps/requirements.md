# Requirements Document

## Introduction

This specification addresses four gaps in the Comptoir (B2B portal) identified in ticket MA-69: the placeholder Recurrences page that is not wired to the existing backend API, the dead "Editer" button on upcoming deliveries, swallowed errors rendering false empty states, and broken French accents / hardcoded English strings.

## Linked Ticket

MA-69 — Comptoir gaps: Recurrences placeholder, dead Editer button, swallowed errors, broken FR accents

## Glossary

- **Comptoir**: The B2B ordering portal within the bakery application, used by professional buyers.
- **Recurring_Order**: A standing order template stored in the `recurring_orders` table, defining items, frequency, bakery, and schedule.
- **RecurrencesPage**: The frontend page at `/comptoir/recurrences` for managing recurring orders.
- **LivraisonsPage**: The frontend page at `/comptoir/livraisons` listing upcoming and past deliveries.
- **FacturesPage**: The frontend page at `/comptoir/factures` listing invoices.
- **Error_State**: A UI state that informs the user that data could not be loaded due to a system error, distinct from an empty-data state.
- **i18n_System**: The existing internationalization module at `frontend/src/i18n/translations.ts` providing EN, FR, and NL translations.

---

## Requirements

### Requirement 1: Recurrences Page — Real CRUD Operations

**User Story:** As a B2B buyer, I want to list, create, edit, pause, resume, and delete recurring orders from the Recurrences page, so that I can manage my standing orders without contacting support.

#### Acceptance Criteria

1. WHEN the RecurrencesPage loads, THE RecurrencesPage SHALL fetch the user's recurring orders from `GET /api/recurring-orders` and display them in a table.
2. WHEN the user clicks "New", THE RecurrencesPage SHALL display a creation form with fields for bakery selection, product items, frequency, scheduled day, and time slot.
3. WHEN the user submits a valid creation form, THE RecurrencesPage SHALL call `POST /api/recurring-orders` and append the new recurring order to the list on success.
4. WHEN the user toggles an active recurring order to paused, THE RecurrencesPage SHALL call `PUT /api/recurring-orders/{id}/pause` and update the row status on success.
5. WHEN the user toggles a paused recurring order to active, THE RecurrencesPage SHALL call `PUT /api/recurring-orders/{id}/resume` and update the row status on success.
6. WHEN the user deletes a recurring order, THE RecurrencesPage SHALL call `DELETE /api/recurring-orders/{id}` and remove the row from the list on success.
7. IF any API call on the RecurrencesPage fails, THEN THE RecurrencesPage SHALL display an Error_State message and retain the previous data.

### Requirement 2: Livraisons "Editer" Button Wiring

**User Story:** As a B2B buyer, I want to click the "Editer" button on an upcoming delivery to edit that order, so that I can modify quantities before the delivery is dispatched.

#### Acceptance Criteria

1. WHEN the user clicks "Editer" on an upcoming delivery, THE LivraisonsPage SHALL navigate to the order edit flow passing the delivery ID.
2. THE LivraisonsPage SHALL call `editOrder(orderId, items)` from `b2b-client.ts` when the user saves edits.
3. IF the edit API call fails, THEN THE LivraisonsPage SHALL display an Error_State indicating the edit could not be saved.

### Requirement 3: Error States Instead of Swallowed Errors

**User Story:** As a B2B buyer, I want to see a clear error message when data cannot be loaded, so that I can distinguish a system outage from having no data.

#### Acceptance Criteria

1. WHEN `listInvoices` fails on FacturesPage, THE FacturesPage SHALL display an Error_State message instead of an empty list.
2. WHEN `downloadInvoicePDF` fails on FacturesPage, THE FacturesPage SHALL display a visible error notification to the user.
3. WHEN `listDeliveries` fails on LivraisonsPage, THE LivraisonsPage SHALL display an Error_State message instead of an empty list.
4. THE Error_State SHALL be visually distinct from the empty-data state, using a different icon or color and an explicit error message.
5. THE Error_State SHALL include a retry action allowing the user to re-attempt the failed request.

### Requirement 4: French Accents and Hardcoded Strings

**User Story:** As a French-speaking B2B buyer, I want all interface text to use correct French accents and no hardcoded English fragments, so that I can use the portal comfortably in my language.

#### Acceptance Criteria

1. THE i18n_System SHALL use correct French accents in all comptoir delivery status strings: "Confirmee" becomes "Confirmee" corrected to "Confirmee" with accent "Confirmée", "A venir" becomes "À venir", "Passees" becomes "Passées", "Editer" becomes "Éditer".
2. THE i18n_System SHALL use correct French accents in all comptoir recurrence strings: "recurrentes" becomes "récurrentes", "recurrence" becomes "récurrence", "Frequence" becomes "Fréquence", "Desactiver" becomes "Désactiver".
3. THE i18n_System SHALL use correct French accents in remaining comptoir strings: "Confirmee" (invoice paid) becomes "Payée" is already correct but "depassee" becomes "dépassée", "Creneau" becomes "Créneau", "sauvegardees" becomes "sauvegardées", "derniere" becomes "dernière".
4. THE LivraisonsPage SHALL replace the hardcoded English string `"{n} items"` with a translated, pluralization-aware i18n key `comptoir.deliveries.itemCount`.
5. THE i18n_System SHALL provide the `comptoir.deliveries.itemCount` key in EN, FR, and NL with correct pluralization.
