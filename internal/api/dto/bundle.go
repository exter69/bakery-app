package dto

// CreateBundleRequest is the request body for POST /api/bundles.
type CreateBundleRequest struct {
	Name            string              `json:"name"`
	Type            string              `json:"type"`            // "compose" or "surprise"
	PhotoURL        string              `json:"photoUrl"`
	Description     string              `json:"description"`     // required for surprise
	EstimatedValue  int64               `json:"estimatedValue"`  // required for surprise, cents
	OriginalPrice   int64               `json:"originalPrice"`   // cents
	DiscountedPrice int64               `json:"discountedPrice"` // cents
	QuantityTotal   int                 `json:"quantityTotal"`
	PickupStartTime string              `json:"pickupStartTime"` // "HH:MM"
	PickupEndTime   string              `json:"pickupEndTime"`   // "HH:MM"
	Items           []BundleItemRequest `json:"items"`           // required for compose
}

// BundleItemRequest represents a single item in a create-bundle request.
type BundleItemRequest struct {
	ProductID   string `json:"productId,omitempty"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

// BundleResponse is the response body for bundle endpoints.
type BundleResponse struct {
	ID                string               `json:"id"`
	BakeryID          string               `json:"bakeryId"`
	BakeryName        string               `json:"bakeryName"`
	BakeryLatitude    float64              `json:"bakeryLatitude"`
	BakeryLongitude   float64              `json:"bakeryLongitude"`
	Name              string               `json:"name"`
	Type              string               `json:"type"`
	PhotoURL          string               `json:"photoUrl"`
	Description       string               `json:"description"`
	EstimatedValue    int64                `json:"estimatedValue"`
	OriginalPrice     int64                `json:"originalPrice"`
	DiscountedPrice   int64                `json:"discountedPrice"`
	QuantityTotal     int                  `json:"quantityTotal"`
	QuantityRemaining int                  `json:"quantityRemaining"`
	PickupStartTime   string               `json:"pickupStartTime"`
	PickupEndTime     string               `json:"pickupEndTime"`
	PublishedDate     string               `json:"publishedDate"`
	ExpiresAt         string               `json:"expiresAt"`
	Status            string               `json:"status"`
	Items             []BundleItemResponse `json:"items"`
	CreatedAt         string               `json:"createdAt"`
}

// BundleItemResponse represents a single item in a bundle response.
type BundleItemResponse struct {
	ProductID   string `json:"productId,omitempty"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

// BundleReservationResponse is the response for bundle reservation endpoints.
type BundleReservationResponse struct {
	ID         string `json:"id"`
	BundleID   string `json:"bundleId"`
	BundleName string `json:"bundleName"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

// BundleImpactResponse is the response for GET /api/bundles/impact.
type BundleImpactResponse struct {
	TotalSaved    int     `json:"totalSaved"`
	WeightAvoided float64 `json:"weightAvoided"`
}
