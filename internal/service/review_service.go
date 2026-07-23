package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// Review service sentinel errors.
var (
	// ErrNotVerifiedPurchaser is returned when a user has no delivered order at the bakery.
	ErrNotVerifiedPurchaser = errors.New("user has no delivered order at this bakery")

	// ErrReviewAlreadyExists is returned when a user already has a review for this bakery.
	ErrReviewAlreadyExists = errors.New("user already reviewed this bakery")

	// ErrReviewNotFound is returned when a review ID does not match any record.
	ErrReviewNotFound = errors.New("review not found")

	// ErrReviewTextTooLong is returned when review text exceeds the max length.
	ErrReviewTextTooLong = errors.New("review text must be 1000 characters or less")

	// ErrInvalidRating is returned when the rating is not between 1 and 5.
	ErrInvalidRating = errors.New("rating must be between 1 and 5")
)

// ReviewServiceConfig holds dependencies for the review service.
type ReviewServiceConfig struct {
	ReviewRepo domain.ReviewRepository
	OrderRepo  domain.OrderRepository
	BakeryRepo domain.BakeryRepository
}

type reviewService struct {
	reviewRepo domain.ReviewRepository
	orderRepo  domain.OrderRepository
	bakeryRepo domain.BakeryRepository
}

// NewReviewService creates a new ReviewService.
func NewReviewService(cfg ReviewServiceConfig) domain.ReviewService {
	return &reviewService{
		reviewRepo: cfg.ReviewRepo,
		orderRepo:  cfg.OrderRepo,
		bakeryRepo: cfg.BakeryRepo,
	}
}

func (s *reviewService) CreateReview(ctx context.Context, userID string, req domain.CreateReviewRequest) (*domain.Review, error) {
	// Validate rating
	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRating
	}

	// Validate text length
	if len(req.Text) > 1000 {
		return nil, ErrReviewTextTooLong
	}

	// Check that user has no existing review for this bakery
	existing, err := s.reviewRepo.GetByUserAndBakery(ctx, userID, req.BakeryID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrReviewAlreadyExists
	}

	// Check that user has a delivered order at this bakery
	deliveredStatus := domain.OrderStatusDelivered
	filters := domain.OrderFilters{
		Status:  &deliveredStatus,
		SortBy:  "createdAt",
		SortDir: "desc",
	}
	orders, _, err := s.orderRepo.ListByUser(ctx, userID, filters, domain.PaginationParams{Page: 1, PageSize: 100})
	if err != nil {
		return nil, err
	}

	// Find the most recent delivered order at this bakery
	var orderID string
	for _, o := range orders {
		if o.BakeryID == req.BakeryID {
			orderID = o.ID
			break
		}
	}
	if orderID == "" {
		return nil, ErrNotVerifiedPurchaser
	}

	review := &domain.Review{
		ID:        uuid.New().String(),
		BakeryID:  req.BakeryID,
		UserID:    userID,
		OrderID:   orderID,
		Rating:    req.Rating,
		Text:      req.Text,
		Hidden:    false,
		CreatedAt: time.Now(),
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *reviewService) ListReviews(ctx context.Context, bakeryID string, params domain.PaginationParams) (*domain.ListResult[domain.Review], error) {
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.Page < 1 {
		params.Page = 1
	}

	reviews, total, err := s.reviewRepo.ListByBakery(ctx, bakeryID, params)
	if err != nil {
		return nil, err
	}

	if reviews == nil {
		reviews = []domain.Review{}
	}

	return &domain.ListResult[domain.Review]{
		Items:    reviews,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

func (s *reviewService) HideReview(ctx context.Context, sellerID string, reviewID string) error {
	// Load the review
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return ErrReviewNotFound
	}

	// Load the bakery and verify ownership
	bakery, err := s.bakeryRepo.GetBakery(ctx, review.BakeryID)
	if err != nil {
		return err
	}
	if bakery == nil {
		return ErrBakeryNotFound
	}
	if bakery.OwnerID != sellerID {
		return ErrForbidden
	}

	// Toggle hidden flag
	return s.reviewRepo.SetHidden(ctx, reviewID, !review.Hidden)
}

func (s *reviewService) ReportReview(ctx context.Context, reporterID string, reviewID string, reason string) error {
	// Verify review exists
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return ErrReviewNotFound
	}

	report := &domain.ReviewReport{
		ID:         uuid.New().String(),
		ReviewID:   reviewID,
		ReporterID: reporterID,
		Reason:     reason,
		CreatedAt:  time.Now(),
	}

	return s.reviewRepo.CreateReport(ctx, report)
}
