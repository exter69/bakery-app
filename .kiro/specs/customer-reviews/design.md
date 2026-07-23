# Design: Customer Reviews and Ratings

## Overview

This design adds a reviews subsystem to the bakery platform. It introduces a `reviews` table, denormalized rating columns on `bakeries`, a new `ReviewService`/`ReviewRepository` layer, API handlers for CRUD and moderation, and frontend components for display and submission.

---

## Database Schema

### New Table: `reviews`

```sql
-- Migration: 019_create_reviews.sql
CREATE TABLE reviews (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bakery_id   UUID NOT NULL REFERENCES bakeries(id),
    user_id     UUID NOT NULL REFERENCES users(id),
    order_id    UUID NOT NULL REFERENCES orders(id),
    rating      SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text        TEXT,
    hidden      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_reviews_user_bakery UNIQUE (user_id, bakery_id)
);

CREATE INDEX idx_reviews_bakery_id ON reviews(bakery_id) WHERE hidden = FALSE;
```

### New Table: `review_reports`

```sql
CREATE TABLE review_reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id   UUID NOT NULL REFERENCES reviews(id),
    reporter_id UUID NOT NULL REFERENCES users(id),
    reason      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_reports_user_review UNIQUE (reporter_id, review_id)
);
```

### Alter Table: `bakeries`

```sql
-- Migration: 019_create_reviews.sql (same migration)
ALTER TABLE bakeries
    ADD COLUMN rating_avg  NUMERIC(2,1),
    ADD COLUMN rating_count INTEGER NOT NULL DEFAULT 0;
```

---

## Domain Model

### New Types (internal/domain/models.go)

```go
// Review represents a customer review of a bakery.
type Review struct {
    ID        string    `json:"id"`
    BakeryID  string    `json:"bakeryId"`
    UserID    string    `json:"userId"`
    OrderID   string    `json:"orderId"`
    Rating    int       `json:"rating"`    // 1-5
    Text      string    `json:"text"`      // optional, max 1000 chars
    Hidden    bool      `json:"hidden"`
    CreatedAt time.Time `json:"createdAt"`
}

// ReviewReport represents a report filed against a review.
type ReviewReport struct {
    ID         string    `json:"id"`
    ReviewID   string    `json:"reviewId"`
    ReporterID string    `json:"reporterId"`
    Reason     string    `json:"reason"`
    CreatedAt  time.Time `json:"createdAt"`
}
```

### Updated Type (Bakery struct)

Add two fields to the existing `Bakery` struct:

```go
RatingAvg   *float64 `json:"ratingAvg"`   // nil when no reviews
RatingCount int      `json:"ratingCount"`
```

And correspondingly to `BakeryCard` in the frontend types.

---

## Repository Interface

### New Interface (internal/domain/repository.go)

```go
// ReviewRepository provides data access for reviews.
type ReviewRepository interface {
    // Create persists a new review and updates bakery rating aggregates.
    Create(ctx context.Context, review *Review) error

    // GetByID returns a review by ID, or nil if not found.
    GetByID(ctx context.Context, id string) (*Review, error)

    // GetByUserAndBakery returns the user's review for a bakery, or nil.
    GetByUserAndBakery(ctx context.Context, userID, bakeryID string) (*Review, error)

    // ListByBakery returns non-hidden reviews for a bakery, paginated, newest first.
    ListByBakery(ctx context.Context, bakeryID string, params PaginationParams) ([]Review, int, error)

    // SetHidden toggles the hidden flag and recalculates bakery rating aggregates.
    SetHidden(ctx context.Context, reviewID string, hidden bool) error

    // CreateReport persists a review report.
    CreateReport(ctx context.Context, report *ReviewReport) error
}
```

---

## Service Layer

### New Interface (internal/domain/services.go)

```go
// ReviewService handles review creation, listing, and moderation.
type ReviewService interface {
    // CreateReview validates purchaser status and creates a review.
    CreateReview(ctx context.Context, userID string, req CreateReviewRequest) (*Review, error)

    // ListReviews returns paginated reviews for a bakery.
    ListReviews(ctx context.Context, bakeryID string, params PaginationParams) (*ListResult[Review], error)

    // HideReview toggles the hidden flag; only the bakery owner can do this.
    HideReview(ctx context.Context, sellerID string, reviewID string) error

    // ReportReview files a report against a review.
    ReportReview(ctx context.Context, reporterID string, reviewID string, reason string) error
}
```

### CreateReviewRequest

```go
type CreateReviewRequest struct {
    BakeryID string
    Rating   int
    Text     string
}
```

### Service Implementation Logic (internal/service/review_service.go)

`CreateReview`:
1. Validate rating is 1-5.
2. Validate text length <= 1000 chars (post-sanitization).
3. Query `OrderRepository.ListByUser` for the user's orders at that bakery with status "delivered". If none, return a "not a verified purchaser" error.
4. Check `ReviewRepository.GetByUserAndBakery`. If exists, return conflict error.
5. Pick the most recent delivered order ID as `order_id`.
6. Call `ReviewRepository.Create` (which atomically updates `bakeries.rating_avg` and `rating_count`).

`HideReview`:
1. Load review by ID.
2. Load bakery by review's `bakery_id`.
3. Verify `bakery.OwnerID == sellerID`. If not, return forbidden error.
4. Call `ReviewRepository.SetHidden`.

---

## API Endpoints

### Customer Routes (require authenticated customer)

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | `/api/bakeries/{id}/reviews` | `CreateReview` | Submit a review |
| GET | `/api/bakeries/{id}/reviews` | `ListReviews` | List visible reviews |
| POST | `/api/reviews/{id}/report` | `ReportReview` | Report a review |

### Seller Routes (require authenticated seller)

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| PUT | `/api/seller/reviews/{id}/hide` | `HideReview` | Toggle hidden flag |

### Handler Structure (internal/api/review_handler.go)

```go
type ReviewHandler struct {
    svc domain.ReviewService
}

func NewReviewHandler(svc domain.ReviewService) *ReviewHandler { ... }

func (h *ReviewHandler) RegisterRoutes(r chi.Router) {
    r.Post("/api/bakeries/{id}/reviews", h.CreateReview)
    r.Get("/api/bakeries/{id}/reviews", h.ListReviews)
    r.Post("/api/reviews/{id}/report", h.ReportReview)
}

func (h *ReviewHandler) RegisterSellerRoutes(r chi.Router) {
    r.Put("/api/seller/reviews/{id}/hide", h.HideReview)
}
```

### Request/Response DTOs (internal/api/dto/)

```go
// CreateReviewRequest is the request body for POST /api/bakeries/{id}/reviews.
type CreateReviewRequest struct {
    Rating int    `json:"rating"`
    Text   string `json:"text"`
}

// ReviewResponse is the public representation of a review.
type ReviewResponse struct {
    ID         string  `json:"id"`
    Rating     int     `json:"rating"`
    Text       *string `json:"text"`
    AuthorName string  `json:"authorName"`
    CreatedAt  string  `json:"createdAt"`
}

// ReportReviewRequest is the request body for POST /api/reviews/{id}/report.
type ReportReviewRequest struct {
    Reason string `json:"reason"`
}
```

---

## Rating Aggregate Update Strategy

On review creation or hide/unhide, recalculate using SQL:

```sql
UPDATE bakeries SET
    rating_avg = (SELECT AVG(rating)::NUMERIC(2,1) FROM reviews WHERE bakery_id = $1 AND hidden = FALSE),
    rating_count = (SELECT COUNT(*) FROM reviews WHERE bakery_id = $1 AND hidden = FALSE)
WHERE id = $1;
```

This runs inside the same transaction as the review insert/update to maintain consistency.

---

## Frontend Components

### StarRating Component

- **Location**: `frontend/src/components/StarRating.tsx`
- **Props**: `rating: number` (display mode), `onChange?: (rating: number) => void` (interactive mode), `size?: 'sm' | 'md' | 'lg'`
- **Behavior**: Renders 5 SVG star icons. In interactive mode, hovering highlights stars and clicking calls `onChange`. Uses `aria-label` for accessibility.

### ReviewList Component

- **Location**: `frontend/src/components/ReviewList.tsx`
- **Props**: `bakeryId: string`
- **Behavior**: Fetches paginated reviews from `/api/bakeries/{id}/reviews`. Renders each review with StarRating (display mode), author name, relative timestamp, and text. Shows empty state when no reviews.

### ReviewPrompt Modal

- **Location**: `frontend/src/components/ReviewPrompt.tsx`
- **Props**: `bakeryId: string`, `onClose: () => void`, `onSubmitted: () => void`
- **Behavior**: Displays StarRating (interactive), optional textarea, and submit button. Calls `POST /api/bakeries/{id}/reviews`. On success calls `onSubmitted`.

### Integration Points

- **BakeriesPage.tsx**: Add StarRating + count to bakery card (both grid and ledger views).
- **BakeryDetailPage.tsx**: Add StarRating + count in header. Add ReviewList below menu. Conditionally render ReviewPrompt if user has a delivered order and no existing review.
- **BakeryCard type** (`frontend/src/types/bakery.ts`): Add `ratingAvg?: number | null` and `ratingCount: number`.

---

## API Client (Frontend)

New file: `frontend/src/api/reviews.ts`

```typescript
export async function fetchReviews(bakeryId: string, page: number): Promise<ListResponse<Review>> { ... }
export async function createReview(bakeryId: string, data: { rating: number; text?: string }): Promise<Review> { ... }
export async function reportReview(reviewId: string, reason: string): Promise<void> { ... }
```

---

## Internationalization

Add keys to all three locale files (`en.json`, `fr.json`, `nl.json`):

- `reviews.title`: "Reviews" / "Avis" / "Beoordelingen"
- `reviews.writeReview`: "Write a review" / "Laisser un avis" / "Schrijf een beoordeling"
- `reviews.ratingLabel`: "Your rating" / "Votre note" / "Jouw beoordeling"
- `reviews.textPlaceholder`: "Share your experience (optional)" / ...
- `reviews.submit`: "Submit" / "Envoyer" / "Verzenden"
- `reviews.empty`: "No reviews yet" / "Pas encore d'avis" / "Nog geen beoordelingen"
- `reviews.thankYou`: "Thanks for your review!" / ...
- `reviews.reported`: "Review reported" / ...
- `reviews.count`: "{count} reviews" / "{count} avis" / "{count} beoordelingen"

---

## Security Considerations

- Review creation requires JWT auth + verified purchaser check (delivered order at bakery).
- Hide endpoint restricted to seller role + bakery ownership verification.
- Text input sanitized via existing `InputSanitizer` middleware.
- `review_reports` stores audit trail; no immediate action taken on report (future admin feature).
- Rate limiting on review creation (existing rate limiter applies).
