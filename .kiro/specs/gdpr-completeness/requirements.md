# Requirements Document

## Introduction

The `DeleteAccount` function anonymizes the user row but leaves personally identifiable data in auxiliary tables (`social_logins`, `business_profiles`, push subscriptions) and does not revoke the Stripe Customer object. The data export omits reviews authored by the user. Issued JWTs remain valid after deletion. This spec addresses all gaps to achieve full GDPR Article 17 compliance.

## Linked Ticket

MA-71 — GDPR completeness: deletion leaves IBAN/social logins/push subs; export omits reviews; JWTs survive deletion

## Glossary

- **Deletion_Service**: The `UserService.DeleteAccount` function responsible for erasure/anonymization.
- **Auth_Middleware**: The `JWTAuth` middleware that validates Bearer tokens on every request.
- **Export_Service**: The `UserService.ExportData` function responsible for GDPR data portability.
- **Push_Store**: The in-memory `push.Store` holding Web Push subscription endpoints per user.
- **Stripe_Customer_Service**: The `payment.StripeCustomerService` managing Stripe Customer lifecycle.
- **User**: A row in the `users` table representing a platform account.
- **Social_Login**: A row in `social_logins` linking an OAuth provider identity to a User.
- **B2B_Profile**: A row in `business_profiles` containing company name, VAT, IBAN, billing contact.
- **Deleted_User**: A User whose `username` starts with `deleted-` and whose `contact_email` is empty.

---

## Requirements

### Requirement 1: Delete Social Login Records on Account Deletion

**User Story:** As a user exercising my right to erasure, I want my linked social login records removed when I delete my account, so that no provider IDs or emails remain in the system.

#### Acceptance Criteria

1.1 WHEN a user deletes their account, THE Deletion_Service SHALL delete all rows in `social_logins` where `user_id` matches the deleted user.

1.2 WHEN the deletion completes, THE system SHALL have zero rows in `social_logins` for that user ID.

---

### Requirement 2: Anonymize or Delete B2B Profile on Account Deletion

**User Story:** As a business user exercising my right to erasure, I want my B2B profile (company name, VAT number, IBAN, billing contact) removed or anonymized when I delete my account, so that no financial PII remains.

#### Acceptance Criteria

2.1 WHEN a user with a B2B profile deletes their account, THE Deletion_Service SHALL delete the row in `business_profiles` where `user_id` matches the deleted user.

2.2 WHEN a user with saved lists deletes their account, THE Deletion_Service SHALL delete all rows in `saved_lists` where `user_id` matches the deleted user.

2.3 WHEN the deletion completes, THE system SHALL have zero rows in `business_profiles` and `saved_lists` for that user ID.

---

### Requirement 3: Delete Push Subscriptions on Account Deletion

**User Story:** As a user exercising my right to erasure, I want my push notification subscriptions removed when I delete my account, so that no device endpoint data remains.

#### Acceptance Criteria

3.1 WHEN a user deletes their account, THE Deletion_Service SHALL remove all entries from the Push_Store for that user ID.

3.2 WHEN the deletion completes, THE Push_Store SHALL return an empty list for that user ID.

---

### Requirement 4: Delete Stripe Customer on Account Deletion

**User Story:** As a user exercising my right to erasure, I want my Stripe Customer object deleted when I delete my account, so that payment data is removed from the external processor.

#### Acceptance Criteria

4.1 WHEN a user with a non-empty `stripe_customer_id` deletes their account, THE Deletion_Service SHALL call the Stripe Customer delete API for that customer ID.

4.2 IF the Stripe API call fails, THEN THE Deletion_Service SHALL log the error and continue with the remaining deletion steps without failing the overall operation.

4.3 WHEN the deletion completes, THE user row SHALL have an empty `stripe_customer_id` field.

---

### Requirement 5: Invalidate JWTs After Account Deletion

**User Story:** As a platform operator, I want JWTs issued before account deletion to be rejected by the API, so that a deleted user cannot continue calling endpoints.

#### Acceptance Criteria

5.1 WHEN a request arrives with a valid JWT, THE Auth_Middleware SHALL verify that the user referenced by the token's `sub` claim is not a Deleted_User.

5.2 IF the user is a Deleted_User, THEN THE Auth_Middleware SHALL return HTTP 401 with code `ACCOUNT_DELETED`.

5.3 WHILE the user has not been deleted, THE Auth_Middleware SHALL allow the request to proceed normally (no behavior change for active users).

---

### Requirement 6: Include Reviews in Data Export

**User Story:** As a user exercising my right to data portability, I want the data export to include all reviews I have authored, so that my export is complete.

#### Acceptance Criteria

6.1 THE Export_Service SHALL include all reviews authored by the user in the data export response.

6.2 WHEN a user has authored reviews, THE export `reviews` array SHALL contain each review's ID, bakery ID, rating, text, and creation timestamp.

6.3 WHEN a user has authored zero reviews, THE export `reviews` array SHALL be an empty array.

---

### Requirement 7: Update DATA-INVENTORY.md to Reflect Implementation

**User Story:** As a compliance officer, I want the data inventory documentation to accurately describe the deletion behavior, so that audits reflect reality.

#### Acceptance Criteria

7.1 THE `docs/DATA-INVENTORY.md` data retention table SHALL state that social logins are hard-deleted on account deletion.

7.2 THE `docs/DATA-INVENTORY.md` data retention table SHALL state that B2B profiles are hard-deleted on account deletion.

7.3 THE `docs/DATA-INVENTORY.md` data retention table SHALL state that push subscriptions are cleared from memory on account deletion.

7.4 THE `docs/DATA-INVENTORY.md` data retention table SHALL state that the Stripe Customer object is deleted via API on account deletion.

7.5 THE `docs/DATA-INVENTORY.md` data subject rights section SHALL note that JWTs are invalidated post-deletion via middleware check.
