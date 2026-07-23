package domain

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

// RoleBusiness is the role for B2B professional clients.
const RoleBusiness UserRole = 3

// PaymentMethodOnInvoice is the payment method for B2B orders.
const PaymentMethodOnInvoice PaymentMethod = "on_invoice"

// B2BAccessStatus represents the state of a bakery-business access request.
type B2BAccessStatus string

const (
	B2BAccessPending  B2BAccessStatus = "pending"
	B2BAccessApproved B2BAccessStatus = "approved"
	B2BAccessRejected B2BAccessStatus = "rejected"
	B2BAccessRevoked  B2BAccessStatus = "revoked"
)

// BusinessProfile holds company-level details for a B2B user.
type BusinessProfile struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"userId"`
	CompanyName        string    `json:"companyName"`
	VATSiret           string    `json:"vatSiret"`
	IBAN               string    `json:"iban"`
	BillingEmail       string    `json:"billingEmail"`
	BillingContactName string    `json:"billingContactName"`
	ProDiscount        int       `json:"proDiscount"`        // per-account pro discount %
	CurrentMonthSpend  int64     `json:"currentMonthSpend"`  // cents HT spent this month
	SpendMonth         string    `json:"spendMonth"`         // YYYY-MM of current tracked month
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// DeliverySite represents a delivery address for a business user.
type DeliverySite struct {
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

// B2BAccess represents a bakery-to-business whitelisting record.
type B2BAccess struct {
	ID             string          `json:"id"`
	BakeryID       string          `json:"bakeryId"`
	BusinessUserID string          `json:"businessUserId"`
	Status         B2BAccessStatus `json:"status"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// B2BConfig holds per-bakery B2B ordering rules.
type B2BConfig struct {
	ID                  string    `json:"id"`
	BakeryID            string    `json:"bakeryId"`
	CutoffTime          TimeOfDay `json:"cutoffTime"`
	DeliveryWindowStart TimeOfDay `json:"deliveryWindowStart"`
	DeliveryWindowEnd   TimeOfDay `json:"deliveryWindowEnd"`
	OrderMinimum        int64     `json:"orderMinimum"`
	ProDiscount         int       `json:"proDiscount"`
	VATRate             int       `json:"vatRate"` // VAT rate %, default 6 for Belgium
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// SavedList holds a named product list for Commande Rapide.
type SavedList struct {
	ID        string          `json:"id"`
	UserID    string          `json:"userId"`
	BakeryID  string          `json:"bakeryId"`
	Name      string          `json:"name"`
	Items     []SavedListItem `json:"items"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// SavedListItem is a product-quantity pair within a SavedList.
type SavedListItem struct {
	ID        string `json:"id"`
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// B2BInvoice represents a generated invoice for a delivered B2B order.
type B2BInvoice struct {
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

// B2BOrderPricing computes and holds the pricing breakdown for a B2B order.
type B2BOrderPricing struct {
	SubtotalHT     int64 `json:"subtotalHt"`
	DiscountRate   int   `json:"discountRate"`
	DiscountAmount int64 `json:"discountAmount"`
	TVARate        int   `json:"tvaRate"`
	TVAAmount      int64 `json:"tvaAmount"`
	TotalTTC       int64 `json:"totalTtc"`
}

// VolumeTier defines a spend threshold and its associated discount.
type VolumeTier struct {
	ID              string `json:"id"`
	MinMonthlySpend int64  `json:"minMonthlySpend"` // cents HT
	DiscountPercent int    `json:"discountPercent"`  // 0-100
}

// B2BPricingResult holds the complete pricing breakdown including volume tiers.
type B2BPricingResult struct {
	SubtotalHT      int64       `json:"subtotalHt"`      // sum of (qty * unit_price)
	ProDiscountRate int         `json:"proDiscountRate"` // per-account %
	ProDiscountAmt  int64       `json:"proDiscountAmt"`  // subtotalHT * proDiscount / 100
	VolDiscountRate int         `json:"volDiscountRate"` // volume tier %
	VolDiscountAmt  int64       `json:"volDiscountAmt"`  // (subtotalHT - proDiscount) * volDiscount / 100
	TVARate         int         `json:"tvaRate"`         // e.g. 6
	TVAAmount       int64       `json:"tvaAmount"`       // computed on discounted subtotal
	TotalTTC        int64       `json:"totalTtc"`
	CurrentTier     *VolumeTier `json:"currentTier,omitempty"` // tier the account is currently on
	NextTier        *VolumeTier `json:"nextTier,omitempty"`    // next achievable tier
	MonthlySpend    int64       `json:"monthlySpend"`          // current rolling monthly spend
	SpendToNextTier int64       `json:"spendToNextTier"`       // how much more to reach next tier (0 if at max)
}

// RegisterBusinessRequest is the input for B2B registration.
type RegisterBusinessRequest struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	CompanyName        string `json:"companyName"`
	VATSiret           string `json:"vatSiret"`
	IBAN               string `json:"iban"`
	BillingEmail       string `json:"billingEmail"`
	BillingContactName string `json:"billingContactName"`
}

// UpdateProfileRequest is the input for profile updates (VAT/SIRET excluded).
type UpdateProfileRequest struct {
	CompanyName        string `json:"companyName"`
	IBAN               string `json:"iban"`
	BillingEmail       string `json:"billingEmail"`
	BillingContactName string `json:"billingContactName"`
}

// CheckoutRequest is the input for per-bakery B2B checkout.
type CheckoutRequest struct {
	BakeryID       string      `json:"bakeryId"`
	DeliverySiteID string      `json:"deliverySiteId"`
	Items          []OrderItem `json:"items"`
}

// EditOrderRequest is the input for editing a submitted B2B order.
type EditOrderRequest struct {
	Items []OrderItem `json:"items"`
}

// B2BOrderFilters extends OrderFilters for B2B-specific filtering.
type B2BOrderFilters struct {
	BakeryID string     `json:"bakeryId,omitempty"`
	Status   string     `json:"status,omitempty"`
	DateFrom *time.Time `json:"dateFrom,omitempty"`
	DateTo   *time.Time `json:"dateTo,omitempty"`
}

// ValidateBusinessRegistration validates a business registration request.
func ValidateBusinessRegistration(req RegisterBusinessRequest) error {
	var errs []string

	if strings.TrimSpace(req.Username) == "" {
		errs = append(errs, "username is required")
	}
	if len(req.Password) < 6 {
		errs = append(errs, "password must be at least 6 characters")
	}
	if strings.TrimSpace(req.CompanyName) == "" {
		errs = append(errs, "company name is required")
	}
	if len(req.CompanyName) > 200 {
		errs = append(errs, "company name must not exceed 200 characters")
	}
	if strings.TrimSpace(req.VATSiret) == "" {
		errs = append(errs, "VAT/SIRET number is required")
	}
	if len(req.VATSiret) > 20 {
		errs = append(errs, "VAT/SIRET must not exceed 20 characters")
	}
	if strings.TrimSpace(req.IBAN) == "" {
		errs = append(errs, "IBAN is required")
	}
	if len(req.IBAN) > 34 {
		errs = append(errs, "IBAN must not exceed 34 characters")
	}
	if strings.TrimSpace(req.BillingEmail) == "" {
		errs = append(errs, "billing email is required")
	} else if _, err := mail.ParseAddress(req.BillingEmail); err != nil {
		errs = append(errs, "billing email is not a valid email address")
	}
	if strings.TrimSpace(req.BillingContactName) == "" {
		errs = append(errs, "billing contact name is required")
	}
	if len(req.BillingContactName) > 100 {
		errs = append(errs, "billing contact name must not exceed 100 characters")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// B2BService handles all B2B-specific business logic.
type B2BService interface {
	// Registration & Profile
	RegisterBusiness(ctx context.Context, req RegisterBusinessRequest) (*BusinessProfile, string, error)
	GetProfile(ctx context.Context, userID string) (*BusinessProfile, error)
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*BusinessProfile, error)

	// Delivery Sites
	CreateSite(ctx context.Context, userID string, site DeliverySite) (*DeliverySite, error)
	ListSites(ctx context.Context, userID string) ([]DeliverySite, error)
	UpdateSite(ctx context.Context, userID string, siteID string, site DeliverySite) (*DeliverySite, error)
	DeleteSite(ctx context.Context, userID string, siteID string) error

	// Access Management
	RequestAccess(ctx context.Context, userID string, bakeryID string) (*B2BAccess, error)
	ApproveAccess(ctx context.Context, accessID string) error
	RejectAccess(ctx context.Context, accessID string) error
	RevokeAccess(ctx context.Context, accessID string) error
	ListAccessRequests(ctx context.Context, bakeryID string) ([]B2BAccess, error)
	ListApprovedBakeries(ctx context.Context, userID string) ([]Bakery, error)
	HasApprovedAccess(ctx context.Context, userID string, bakeryID string) (bool, error)

	// B2B Config
	GetConfig(ctx context.Context, bakeryID string) (*B2BConfig, error)
	SaveConfig(ctx context.Context, bakeryID string, config B2BConfig) (*B2BConfig, error)

	// Cart & Checkout
	CheckoutBakeryGroup(ctx context.Context, userID string, req CheckoutRequest) (*Order, error)
	EditOrder(ctx context.Context, userID string, orderID string, req EditOrderRequest) (*Order, error)
	ComputePricing(ctx context.Context, userID string, bakeryID string, items []OrderItem) (*B2BPricingResult, error)

	// Saved Lists
	CreateSavedList(ctx context.Context, userID string, list SavedList) (*SavedList, error)
	ListSavedLists(ctx context.Context, userID string, bakeryID string) ([]SavedList, error)
	DeleteSavedList(ctx context.Context, userID string, listID string) error

	// Deliveries & Invoices
	ListDeliveries(ctx context.Context, userID string, filters B2BOrderFilters, params PaginationParams) (*ListResult[Order], error)
	GetLastOrder(ctx context.Context, userID string, bakeryID string) (*Order, error)
	ListInvoices(ctx context.Context, userID string, params PaginationParams) (*ListResult[B2BInvoice], error)
	GenerateInvoice(ctx context.Context, orderID string) (*B2BInvoice, error)
	DownloadInvoicePDF(ctx context.Context, invoiceID string, userID string) ([]byte, error)
}

// B2BRepository provides data access for B2B-specific tables.
type B2BRepository interface {
	// Business Profiles
	CreateProfile(ctx context.Context, profile *BusinessProfile) error
	GetProfileByUserID(ctx context.Context, userID string) (*BusinessProfile, error)
	GetProfileByVAT(ctx context.Context, vatSiret string) (*BusinessProfile, error)
	UpdateProfile(ctx context.Context, profile *BusinessProfile) error

	// Delivery Sites
	CreateSite(ctx context.Context, site *DeliverySite) error
	GetSiteByID(ctx context.Context, id string) (*DeliverySite, error)
	ListSitesByUser(ctx context.Context, userID string) ([]DeliverySite, error)
	UpdateSite(ctx context.Context, site *DeliverySite) error
	DeleteSite(ctx context.Context, id string) error
	CountSitesByUser(ctx context.Context, userID string) (int, error)

	// Access Whitelisting
	CreateAccess(ctx context.Context, access *B2BAccess) error
	GetAccessByID(ctx context.Context, id string) (*B2BAccess, error)
	GetAccess(ctx context.Context, bakeryID string, userID string) (*B2BAccess, error)
	UpdateAccessStatus(ctx context.Context, id string, status B2BAccessStatus) error
	ListAccessByBakery(ctx context.Context, bakeryID string, status *B2BAccessStatus) ([]B2BAccess, error)
	ListApprovedBakeryIDs(ctx context.Context, userID string) ([]string, error)

	// B2B Config
	GetConfig(ctx context.Context, bakeryID string) (*B2BConfig, error)
	SaveConfig(ctx context.Context, config *B2BConfig) error

	// Volume Tiers
	ListVolumeTiers(ctx context.Context) ([]VolumeTier, error)

	// Monthly Spend Tracking
	UpdateMonthlySpend(ctx context.Context, profileID string, amount int64, month string) error

	// Saved Lists
	CreateSavedList(ctx context.Context, list *SavedList) error
	GetSavedListByID(ctx context.Context, id string) (*SavedList, error)
	ListSavedLists(ctx context.Context, userID string, bakeryID string) ([]SavedList, error)
	DeleteSavedList(ctx context.Context, id string) error

	// Invoices
	CreateInvoice(ctx context.Context, invoice *B2BInvoice) error
	GetInvoiceByID(ctx context.Context, id string) (*B2BInvoice, error)
	GetInvoiceByOrder(ctx context.Context, orderID string) (*B2BInvoice, error)
	ListInvoicesByUser(ctx context.Context, profileID string, params PaginationParams) ([]B2BInvoice, int, error)
	NextInvoiceNumber(ctx context.Context, bakeryID string) (int, error)

	// GDPR deletion
	DeleteProfile(ctx context.Context, userID string) error
	DeleteSavedListsByUser(ctx context.Context, userID string) error
}
