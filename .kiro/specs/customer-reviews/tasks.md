# Implementation Plan: Customer Reviews and Ratings

## Overview

This plan implements customer reviews and ratings for bakeries: database migration, domain types, repository and service layers, API handler, frontend components (StarRating, ReviewList, ReviewPrompt), bakery card/detail integration, i18n, and seed data.

## Tasks

- [x] 1. Database migration and domain model
  - [x] 1.1 Create migration `db/migrations/019_create_reviews.sql` with reviews table, review_reports table, and bakery rating columns
  - [x] 1.2 Add `Review` and `ReviewReport` structs to `internal/domain/models.go`
  - [x] 1.3 Add `RatingAvg *float64` and `RatingCount int` fields to the existing `Bakery` struct

- [x] 2. Repository layer
  - [x] 2.1 Add `ReviewRepository` interface to `internal/domain/repository.go`
  - [x] 2.2 Implement `PostgresReviewRepository` in `internal/repository/postgres/review_repo.go` with transactional aggregate recalculation
  - [x] 2.3 Update existing bakery repository queries to include rating_avg and rating_count

- [x] 3. Service layer
  - [x] 3.1 Add `ReviewService` interface and `CreateReviewRequest` struct to `internal/domain/services.go`
  - [x] 3.2 Implement `ReviewServiceImpl` in `internal/service/review_service.go` (CreateReview, ListReviews, HideReview, ReportReview)

- [x] 4. API handler
  - [x] 4.1 Create DTO structs in `internal/api/dto/review.go`
  - [x] 4.2 Create `ReviewHandler` in `internal/api/review_handler.go` with all endpoints
  - [x] 4.3 Register ReviewHandler routes in `cmd/server/main.go`

- [x] 5. Frontend — StarRating component
  - [x] 5.1 Create `StarRating.tsx` with SVG stars, display and interactive modes, 3 sizes, ARIA attributes
  - [x] 5.2 Create `StarRating.css` with size variants and hover states

- [x] 6. Frontend — ReviewList component
  - [x] 6.1 Create `frontend/src/api/reviews.ts` with fetchReviews, createReview, reportReview
  - [x] 6.2 Create `ReviewList.tsx` with paginated reviews, relative timestamps, empty state

- [x] 7. Frontend — ReviewPrompt modal
  - [x] 7.1 Create `ReviewPrompt.tsx` with interactive StarRating, textarea, submit, session dismiss
  - [x] 7.2 Integrate ReviewPrompt into BakeryDetailPage

- [x] 8. Frontend — Bakery card and detail integration
  - [x] 8.1 Update `types/bakery.ts` with ratingAvg and ratingCount fields
  - [x] 8.2 Update BakeriesPage with StarRating on bakery cards
  - [x] 8.3 Update BakeryDetailPage with StarRating in header and ReviewList below menu

- [x] 9. Internationalization
  - [x] 9.1 Add review-related i18n keys for EN, FR, NL
  - [x] 9.2 Use locale-aware relative time formatting in ReviewList

- [x] 10. Verification
  - [x] 10.1 Go build and vet pass
  - [x] 10.2 Frontend TypeScript compilation passes
  - [x] 10.3 Bakery list and detail endpoints return rating_avg and rating_count

## Notes

- Migration number is 019, following existing sequence (018_create_surplus_bundles.sql)
- Rating aggregates are recalculated in the same transaction as review insert/hide to maintain consistency
- StarRating uses SVG paths (no emoji) per the no-emoji steering rule
- ReviewPrompt dismiss is stored in sessionStorage (per bakery, per session)
- One review per user per bakery enforced via unique constraint

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "1.3"] },
    { "id": 1, "tasks": ["2.1", "2.2", "2.3"] },
    { "id": 2, "tasks": ["3.1", "3.2"] },
    { "id": 3, "tasks": ["4.1", "4.2", "4.3"] },
    { "id": 4, "tasks": ["5.1", "5.2", "6.1", "6.2"] },
    { "id": 5, "tasks": ["7.1", "7.2", "8.1", "8.2", "8.3"] },
    { "id": 6, "tasks": ["9.1", "9.2"] },
    { "id": 7, "tasks": ["10.1", "10.2", "10.3"] }
  ]
}
```
