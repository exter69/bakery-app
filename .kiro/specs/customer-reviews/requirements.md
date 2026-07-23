# Requirements: Customer Reviews and Ratings

## Overview

Allow verified customers to leave ratings and optional text reviews for bakeries they have ordered from. Display aggregate ratings on bakery cards and detail pages, list individual reviews, and provide basic moderation controls for bakery owners.

## Linked Ticket

MA-17 — Customer reviews & ratings

---

## Requirement 1: Review Creation (Verified Purchasers Only)

### Acceptance Criteria

1.1 A logged-in customer who has at least one order with status "delivered" at a given bakery can submit a review for that bakery.

1.2 A review consists of a rating (integer 1-5, required) and an optional text body (max 1000 characters after sanitization).

1.3 A customer can leave at most one review per bakery. Attempting to submit a second review returns HTTP 409 Conflict.

1.4 A customer who has no delivered order at the target bakery receives HTTP 403 Forbidden when attempting to submit a review.

1.5 The review text is sanitized using the existing `InputSanitizer` middleware (HTML/script stripping) before persistence.

1.6 Upon successful creation the API returns HTTP 201 with the created review resource.

---

## Requirement 2: Average Rating Display

### Acceptance Criteria

2.1 The bakery list endpoint (`GET /api/bakeries`) includes `ratingAvg` (float, one decimal) and `ratingCount` (integer) for each bakery in the response.

2.2 The bakery detail endpoint (`GET /api/bakeries/{id}`) includes `ratingAvg` and `ratingCount` in the response.

2.3 When a bakery has zero reviews, `ratingAvg` is `null` and `ratingCount` is `0`.

2.4 The frontend bakery card displays the average star rating and review count next to the bakery name.

2.5 The frontend bakery detail page displays the average star rating and review count in the header area.

---

## Requirement 3: Review Listing

### Acceptance Criteria

3.1 `GET /api/bakeries/{id}/reviews` returns a paginated list of non-hidden reviews for the bakery, sorted by most recent first.

3.2 Each review in the list includes: `id`, `rating`, `text` (nullable), `authorName` (username of the reviewer), `createdAt`.

3.3 Pagination follows the existing pattern: `page` and `pageSize` query parameters, response includes `items`, `page`, `pageSize`, `total`.

3.4 The frontend bakery detail page renders a `ReviewList` component below the menu showing reviews with star rating, optional text, author name, and relative timestamp.

---

## Requirement 4: Moderation (Baker Hides/Reports)

### Acceptance Criteria

4.1 A bakery owner (seller role) can hide a review on their bakery via `PUT /api/seller/reviews/{id}/hide`. The endpoint toggles the `hidden` flag.

4.2 Hidden reviews are excluded from the public review list (`GET /api/bakeries/{id}/reviews`) and are not counted in `ratingAvg`/`ratingCount`.

4.3 Only the owner of the bakery that the review belongs to can hide it; otherwise HTTP 403 is returned.

4.4 A customer can report a review via `POST /api/reviews/{id}/report`. This stores a report record for future admin review. Returns HTTP 204 on success.

---

## Requirement 5: Post-Order Review Prompt (Frontend)

### Acceptance Criteria

5.1 After an order transitions to "delivered" status, the next time the customer visits the bakery detail page they see a review prompt modal if they have not yet reviewed that bakery.

5.2 The prompt displays a `StarRating` interactive component and an optional text area.

5.3 Submitting the prompt calls the review creation endpoint and dismisses the modal on success.

5.4 The user can dismiss the prompt without submitting; it does not reappear for the same bakery during the session.

---

## Requirement 6: Internationalization

### Acceptance Criteria

6.1 All user-facing strings introduced by this feature (review prompt title, labels, empty states, error messages) are available in EN, FR, and NL via the existing i18n system.

6.2 Relative timestamps in the review list use locale-aware formatting.
