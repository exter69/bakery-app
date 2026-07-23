package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// ReviewHandler handles HTTP requests for customer reviews.
type ReviewHandler struct {
	svc      domain.ReviewService
	userRepo domain.UserRepository
}

// NewReviewHandler creates a new ReviewHandler.
func NewReviewHandler(svc domain.ReviewService, userRepo domain.UserRepository) *ReviewHandler {
	return &ReviewHandler{svc: svc, userRepo: userRepo}
}

// RegisterPublicRoutes registers public review routes (no auth required).
func (h *ReviewHandler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/api/bakeries/{id}/reviews", h.ListReviews)
}

// RegisterRoutes registers authenticated review routes.
func (h *ReviewHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/bakeries/{id}/reviews", h.CreateReview)
	r.Post("/api/reviews/{id}/report", h.ReportReview)
}

// RegisterSellerRoutes registers seller-only review routes.
func (h *ReviewHandler) RegisterSellerRoutes(r chi.Router) {
	r.Put("/api/seller/reviews/{id}/hide", h.HideReview)
}

// CreateReview handles POST /api/bakeries/{id}/reviews.
func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return
	}

	bakeryID := chi.URLParam(r, "id")
	if bakeryID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bakery ID is required",
		})
		return
	}

	var req dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	// Explicitly sanitize review text to strip HTML/script tags (per-field sanitization).
	sanitizedText := middleware.SanitizeString(req.Text)

	review, err := h.svc.CreateReview(r.Context(), userID, domain.CreateReviewRequest{
		BakeryID: bakeryID,
		Rating:   req.Rating,
		Text:     sanitizedText,
	})
	if err != nil {
		h.handleReviewError(w, err)
		return
	}

	// Look up author name
	authorName := h.getAuthorName(r.Context(), review.UserID)

	writeJSON(w, http.StatusCreated, toReviewResponse(review, authorName))
}

// ListReviews handles GET /api/bakeries/{id}/reviews.
func (h *ReviewHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	bakeryID := chi.URLParam(r, "id")
	if bakeryID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "bakery ID is required",
		})
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil || parsed < 1 {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_PAGE",
				Message: "page must be a positive integer",
			})
			return
		}
		page = parsed
	}

	pageSize := 20
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		parsed, err := strconv.Atoi(ps)
		if err != nil || parsed < 1 || parsed > 100 {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_PAGE_SIZE",
				Message: "pageSize must be between 1 and 100",
			})
			return
		}
		pageSize = parsed
	}

	result, err := h.svc.ListReviews(r.Context(), bakeryID, domain.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to list reviews",
		})
		return
	}

	// Map reviews to response DTOs with author names
	items := make([]dto.ReviewResponse, 0, len(result.Items))
	for i := range result.Items {
		authorName := h.getAuthorName(r.Context(), result.Items[i].UserID)
		items = append(items, toReviewResponse(&result.Items[i], authorName))
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ReviewResponse]{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

// ReportReview handles POST /api/reviews/{id}/report.
func (h *ReviewHandler) ReportReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return
	}

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "review ID is required",
		})
		return
	}

	var req dto.ReportReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	err := h.svc.ReportReview(r.Context(), userID, reviewID, req.Reason)
	if err != nil {
		h.handleReviewError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HideReview handles PUT /api/seller/reviews/{id}/hide.
func (h *ReviewHandler) HideReview(w http.ResponseWriter, r *http.Request) {
	sellerID := requireSeller(w, r)
	if sellerID == "" {
		return
	}

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "review ID is required",
		})
		return
	}

	err := h.svc.HideReview(r.Context(), sellerID, reviewID)
	if err != nil {
		h.handleReviewError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "review visibility updated",
	})
}

// handleReviewError maps review service errors to HTTP responses.
func (h *ReviewHandler) handleReviewError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNotVerifiedPurchaser) {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "NOT_VERIFIED_PURCHASER",
			Message: "you must have a delivered order at this bakery to leave a review",
		})
		return
	}
	if errors.Is(err, service.ErrReviewAlreadyExists) {
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "REVIEW_ALREADY_EXISTS",
			Message: "you have already reviewed this bakery",
		})
		return
	}
	if errors.Is(err, service.ErrReviewNotFound) {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "REVIEW_NOT_FOUND",
			Message: "review not found",
		})
		return
	}
	if errors.Is(err, service.ErrInvalidRating) {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_RATING",
			Message: "rating must be between 1 and 5",
		})
		return
	}
	if errors.Is(err, service.ErrReviewTextTooLong) {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "TEXT_TOO_LONG",
			Message: "review text must be 1000 characters or less",
		})
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "you do not own this resource",
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "an unexpected error occurred",
	})
}

// getAuthorName looks up a user's username by ID. Returns "Unknown" on failure.
func (h *ReviewHandler) getAuthorName(ctx context.Context, userID string) string {
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return "Unknown"
	}
	return user.Username
}

// toReviewResponse converts a domain Review to a response DTO.
func toReviewResponse(review *domain.Review, authorName string) dto.ReviewResponse {
	var text *string
	if review.Text != "" {
		text = &review.Text
	}
	return dto.ReviewResponse{
		ID:         review.ID,
		Rating:     review.Rating,
		Text:       text,
		AuthorName: authorName,
		CreatedAt:  review.CreatedAt.Format(time.RFC3339),
	}
}
