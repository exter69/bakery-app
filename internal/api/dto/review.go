package dto

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
