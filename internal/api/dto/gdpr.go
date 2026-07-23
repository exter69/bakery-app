package dto

import "time"

// DataExportResponse contains all personal data for GDPR data portability (Article 20).
type DataExportResponse struct {
	ExportedAt time.Time              `json:"exportedAt"`
	Profile    DataExportProfile      `json:"profile"`
	Orders     []DataExportOrder      `json:"orders"`
	Reservations []DataExportReservation `json:"reservations"`
	Reviews    []DataExportReview     `json:"reviews"`
	RecurringOrders []DataExportRecurringOrder `json:"recurringOrders"`
	SocialLogins []DataExportSocialLogin `json:"socialLogins"`
	B2BProfile *DataExportB2BProfile  `json:"b2bProfile,omitempty"`
	DeliverySites []DataExportDeliverySite `json:"deliverySites,omitempty"`
}

// DataExportProfile is the user profile section of a data export.
type DataExportProfile struct {
	ID               string     `json:"id"`
	Username         string     `json:"username"`
	ContactEmail     string     `json:"contactEmail,omitempty"`
	Role             int        `json:"role"`
	Locale           string     `json:"locale,omitempty"`
	HolidayMode      bool       `json:"holidayMode"`
	HolidayFrom      *time.Time `json:"holidayFrom,omitempty"`
	HolidayTo        *time.Time `json:"holidayTo,omitempty"`
	FavoriteProducts []string   `json:"favoriteProducts"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// DataExportOrder is the order section of a data export.
type DataExportOrder struct {
	ID            string              `json:"id"`
	BakeryID      string              `json:"bakeryId"`
	Items         []OrderItemResponse `json:"items"`
	Status        string              `json:"status"`
	TotalAmount   int64               `json:"totalAmount"`
	PaymentMethod string              `json:"paymentMethod"`
	CreatedAt     time.Time           `json:"createdAt"`
}

// DataExportReservation is the reservation section of a data export.
type DataExportReservation struct {
	ID          string              `json:"id"`
	BakeryID    string              `json:"bakeryId"`
	Items       []OrderItemResponse `json:"items"`
	Status      string              `json:"status"`
	TotalAmount int64               `json:"totalAmount"`
	CreatedAt   time.Time           `json:"createdAt"`
}

// DataExportReview is the review section of a data export.
type DataExportReview struct {
	ID        string    `json:"id"`
	BakeryID  string    `json:"bakeryId"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

// DataExportRecurringOrder is the recurring order section of a data export.
type DataExportRecurringOrder struct {
	ID            string              `json:"id"`
	BakeryID      string              `json:"bakeryId"`
	Items         []OrderItemResponse `json:"items"`
	Frequency     string              `json:"frequency"`
	SelectionMode string              `json:"selectionMode"`
	Active        bool                `json:"active"`
	CreatedAt     time.Time           `json:"createdAt"`
}

// DataExportSocialLogin is the social login section of a data export.
type DataExportSocialLogin struct {
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

// DataExportB2BProfile is the B2B profile section of a data export.
type DataExportB2BProfile struct {
	CompanyName        string `json:"companyName"`
	VATSiret           string `json:"vatSiret"`
	IBAN               string `json:"iban"`
	BillingEmail       string `json:"billingEmail"`
	BillingContactName string `json:"billingContactName"`
}

// DataExportDeliverySite is the delivery site section of a data export.
type DataExportDeliverySite struct {
	Name                 string `json:"name"`
	StreetAddress        string `json:"streetAddress"`
	City                 string `json:"city"`
	PostalCode           string `json:"postalCode"`
	Country              string `json:"country"`
	DeliveryInstructions string `json:"deliveryInstructions,omitempty"`
}
