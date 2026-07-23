package dto

import "time"

// --- B2B Registration ---

// RegisterBusinessRequest is the request body for POST /api/comptoir/register.
type RegisterBusinessRequest struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	CompanyName        string `json:"companyName"`
	VATSiret           string `json:"vatSiret"`
	IBAN               string `json:"iban"`
	BillingEmail       string `json:"billingEmail"`
	BillingContactName string `json:"billingContactName"`
}

// RegisterBusinessResponse is the response for successful B2B registration.
type RegisterBusinessResponse struct {
	Token   string                  `json:"token"`
	Profile BusinessProfileResponse `json:"profile"`
}

// --- Business Profile ---

// BusinessProfileResponse is the response for business profile endpoints.
type BusinessProfileResponse struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"userId"`
	CompanyName        string    `json:"companyName"`
	VATSiret           string    `json:"vatSiret"`
	IBAN               string    `json:"iban"`
	BillingEmail       string    `json:"billingEmail"`
	BillingContactName string    `json:"billingContactName"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// UpdateProfileRequest is the request body for PUT /api/comptoir/profile.
type UpdateProfileRequest struct {
	CompanyName        string `json:"companyName"`
	IBAN               string `json:"iban"`
	BillingEmail       string `json:"billingEmail"`
	BillingContactName string `json:"billingContactName"`
}

// --- Delivery Sites ---

// DeliverySiteRequest is the request body for creating/updating a delivery site.
type DeliverySiteRequest struct {
	Name                 string `json:"name"`
	StreetAddress        string `json:"streetAddress"`
	City                 string `json:"city"`
	PostalCode           string `json:"postalCode"`
	Country              string `json:"country"`
	DeliveryInstructions string `json:"deliveryInstructions,omitempty"`
}

// DeliverySiteResponse is the response for delivery site endpoints.
type DeliverySiteResponse struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"userId"`
	Name                 string    `json:"name"`
	StreetAddress        string    `json:"streetAddress"`
	City                 string    `json:"city"`
	PostalCode           string    `json:"postalCode"`
	Country              string    `json:"country"`
	DeliveryInstructions string    `json:"deliveryInstructions,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// --- B2B Access ---

// B2BAccessResponse is the response for access management endpoints.
type B2BAccessResponse struct {
	ID             string    `json:"id"`
	BakeryID       string    `json:"bakeryId"`
	BusinessUserID string    `json:"businessUserId"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// --- B2B Config ---

// B2BConfigRequest is the request body for PUT /api/dashboard/b2b/config.
type B2BConfigRequest struct {
	CutoffTime          string `json:"cutoffTime"`          // HH:MM
	DeliveryWindowStart string `json:"deliveryWindowStart"` // HH:MM
	DeliveryWindowEnd   string `json:"deliveryWindowEnd"`   // HH:MM
	OrderMinimum        int64  `json:"orderMinimum"`        // cents HT
	ProDiscount         int    `json:"proDiscount"`         // 0-100
	VATRate             int    `json:"vatRate"`             // 0-100, default 6
}

// B2BConfigResponse is the response for B2B config endpoints.
type B2BConfigResponse struct {
	ID                  string `json:"id"`
	BakeryID            string `json:"bakeryId"`
	CutoffTime          string `json:"cutoffTime"`          // HH:MM
	DeliveryWindowStart string `json:"deliveryWindowStart"` // HH:MM
	DeliveryWindowEnd   string `json:"deliveryWindowEnd"`   // HH:MM
	OrderMinimum        int64  `json:"orderMinimum"`
	ProDiscount         int    `json:"proDiscount"`
	VATRate             int    `json:"vatRate"`
}

// --- Checkout & Orders ---

// B2BCheckoutRequest is the request body for POST /api/comptoir/checkout.
type B2BCheckoutRequest struct {
	BakeryID       string             `json:"bakeryId"`
	DeliverySiteID string             `json:"deliverySiteId"`
	Items          []OrderItemRequest `json:"items"`
}

// B2BEditOrderRequest is the request body for PUT /api/comptoir/orders/{orderId}.
type B2BEditOrderRequest struct {
	Items []OrderItemRequest `json:"items"`
}

// B2BPricingRequest is the request body for POST /api/comptoir/pricing.
type B2BPricingRequest struct {
	BakeryID string             `json:"bakeryId"`
	Items    []OrderItemRequest `json:"items"`
}

// B2BOrderPricingResponse is the response for pricing computation.
type B2BOrderPricingResponse struct {
	SubtotalHT     int64 `json:"subtotalHt"`
	DiscountRate   int   `json:"discountRate"`
	DiscountAmount int64 `json:"discountAmount"`
	TVARate        int   `json:"tvaRate"`
	TVAAmount      int64 `json:"tvaAmount"`
	TotalTTC       int64 `json:"totalTtc"`
}

// VolumeTierResponse is the response for a volume tier.
type VolumeTierResponse struct {
	MinMonthlySpend int64 `json:"minMonthlySpend"`
	DiscountPercent int   `json:"discountPercent"`
}

// B2BPricingResultResponse is the full pricing response with volume tier info.
type B2BPricingResultResponse struct {
	SubtotalHT      int64               `json:"subtotalHt"`
	ProDiscountRate int                  `json:"proDiscountRate"`
	ProDiscountAmt  int64               `json:"proDiscountAmt"`
	VolDiscountRate int                  `json:"volDiscountRate"`
	VolDiscountAmt  int64               `json:"volDiscountAmt"`
	TVARate         int                  `json:"tvaRate"`
	TVAAmount       int64               `json:"tvaAmount"`
	TotalTTC        int64               `json:"totalTtc"`
	CurrentTier     *VolumeTierResponse  `json:"currentTier,omitempty"`
	NextTier        *VolumeTierResponse  `json:"nextTier,omitempty"`
	MonthlySpend    int64               `json:"monthlySpend"`
	SpendToNextTier int64               `json:"spendToNextTier"`
}

// --- Saved Lists ---

// SavedListRequest is the request body for creating a saved list.
type SavedListRequest struct {
	BakeryID string                 `json:"bakeryId"`
	Name     string                 `json:"name"`
	Items    []SavedListItemRequest `json:"items"`
}

// SavedListItemRequest is a product-quantity pair in a saved list request.
type SavedListItemRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// SavedListResponse is the response for saved list endpoints.
type SavedListResponse struct {
	ID        string                  `json:"id"`
	UserID    string                  `json:"userId"`
	BakeryID  string                  `json:"bakeryId"`
	Name      string                  `json:"name"`
	Items     []SavedListItemResponse `json:"items"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

// SavedListItemResponse is a product-quantity pair in a saved list response.
type SavedListItemResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// --- Invoices ---

// B2BInvoiceResponse is the response for invoice endpoints.
type B2BInvoiceResponse struct {
	ID                string     `json:"id"`
	OrderID           string     `json:"orderId"`
	BakeryID          string     `json:"bakeryId"`
	BusinessProfileID string     `json:"businessProfileId"`
	InvoiceNumber     int        `json:"invoiceNumber"`
	SubtotalHT        int64      `json:"subtotalHt"`
	DiscountAmount    int64      `json:"discountAmount"`
	TVAAmount         int64      `json:"tvaAmount"`
	TotalTTC          int64      `json:"totalTtc"`
	PaymentStatus     string     `json:"paymentStatus"`
	IssuedAt          time.Time  `json:"issuedAt"`
	PaidAt            *time.Time `json:"paidAt,omitempty"`
}
