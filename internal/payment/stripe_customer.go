package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"github.com/stripe/stripe-go/v82/setupintent"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrPaymentMethodNotOwned = errors.New("payment method does not belong to this user")
)

// SavedPaymentMethod represents a stored card on file (metadata only — no sensitive data).
type SavedPaymentMethod struct {
	ID        string `json:"id"`        // Stripe payment_method ID (pm_...)
	Brand     string `json:"brand"`     // card brand: visa, mastercard, etc.
	Last4     string `json:"last4"`     // last 4 digits
	ExpMonth  int    `json:"expMonth"`
	ExpYear   int    `json:"expYear"`
	IsDefault bool   `json:"isDefault"` // whether this is the customer's default method
}

// StripeCustomerService manages the Stripe Customer ↔ app user mapping and payment methods.
type StripeCustomerService struct {
	secretKey string
	userRepo  domain.UserRepository
}

// NewStripeCustomerService creates a new service for managing saved payment methods.
func NewStripeCustomerService(secretKey string, userRepo domain.UserRepository) *StripeCustomerService {
	return &StripeCustomerService{
		secretKey: secretKey,
		userRepo:  userRepo,
	}
}

// GetOrCreateCustomer ensures the app user has a Stripe Customer, creating one if needed.
// Returns the Stripe Customer ID.
func (s *StripeCustomerService) GetOrCreateCustomer(ctx context.Context, userID string) (string, error) {
	stripe.Key = s.secretKey

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	// If the user already has a Stripe Customer ID, return it
	if user.StripeCustomerID != "" {
		return user.StripeCustomerID, nil
	}

	// Create a new Stripe Customer
	params := &stripe.CustomerParams{
		Email: stripe.String(user.ContactEmail),
		Metadata: map[string]string{
			"app_user_id": userID,
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe: failed to create customer: %w", err)
	}

	// Persist the Stripe Customer ID on the user
	user.StripeCustomerID = cust.ID
	if err := s.userRepo.Save(ctx, user); err != nil {
		return "", fmt.Errorf("failed to save stripe customer id: %w", err)
	}

	return cust.ID, nil
}

// ListPaymentMethods returns saved cards for a user (only exposes brand, last4, exp — never full card data).
func (s *StripeCustomerService) ListPaymentMethods(ctx context.Context, userID string) ([]SavedPaymentMethod, error) {
	stripe.Key = s.secretKey

	customerID, err := s.GetOrCreateCustomer(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Fetch the customer to check default payment method
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe: failed to get customer: %w", err)
	}

	defaultPMID := ""
	if cust.InvoiceSettings != nil && cust.InvoiceSettings.DefaultPaymentMethod != nil {
		defaultPMID = cust.InvoiceSettings.DefaultPaymentMethod.ID
	}

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}

	iter := paymentmethod.List(params)
	var methods []SavedPaymentMethod

	for iter.Next() {
		pm := iter.PaymentMethod()
		if pm.Card == nil {
			continue
		}
		methods = append(methods, SavedPaymentMethod{
			ID:        pm.ID,
			Brand:     string(pm.Card.Brand),
			Last4:     pm.Card.Last4,
			ExpMonth:  int(pm.Card.ExpMonth),
			ExpYear:   int(pm.Card.ExpYear),
			IsDefault: pm.ID == defaultPMID,
		})
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("stripe: failed to list payment methods: %w", err)
	}

	if methods == nil {
		methods = []SavedPaymentMethod{}
	}

	return methods, nil
}

// CreateSetupIntent creates a Stripe SetupIntent for adding a new payment method.
// Returns the client_secret for the frontend to complete setup via Stripe.js.
func (s *StripeCustomerService) CreateSetupIntent(ctx context.Context, userID string) (string, error) {
	stripe.Key = s.secretKey

	customerID, err := s.GetOrCreateCustomer(ctx, userID)
	if err != nil {
		return "", err
	}

	params := &stripe.SetupIntentParams{
		Customer:           stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	}

	si, err := setupintent.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe: failed to create setup intent: %w", err)
	}

	return si.ClientSecret, nil
}

// DetachPaymentMethod removes a saved payment method from the user's Stripe Customer.
// Verifies ownership before detaching.
func (s *StripeCustomerService) DetachPaymentMethod(ctx context.Context, userID string, paymentMethodID string) error {
	stripe.Key = s.secretKey

	customerID, err := s.GetOrCreateCustomer(ctx, userID)
	if err != nil {
		return err
	}

	// Verify the payment method belongs to this customer
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("stripe: failed to get payment method: %w", err)
	}
	if pm.Customer == nil || pm.Customer.ID != customerID {
		return ErrPaymentMethodNotOwned
	}

	_, err = paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("stripe: failed to detach payment method: %w", err)
	}

	return nil
}

// SetDefaultPaymentMethod sets a payment method as the default for the user's invoices/charges.
// Verifies ownership before updating.
func (s *StripeCustomerService) SetDefaultPaymentMethod(ctx context.Context, userID string, paymentMethodID string) error {
	stripe.Key = s.secretKey

	customerID, err := s.GetOrCreateCustomer(ctx, userID)
	if err != nil {
		return err
	}

	// Verify the payment method belongs to this customer
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("stripe: failed to get payment method: %w", err)
	}
	if pm.Customer == nil || pm.Customer.ID != customerID {
		return ErrPaymentMethodNotOwned
	}

	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}

	_, err = customer.Update(customerID, params)
	if err != nil {
		return fmt.Errorf("stripe: failed to set default payment method: %w", err)
	}

	return nil
}

// ChargeWithSavedMethod charges a saved payment method directly (no Checkout Session needed).
// Creates an off-session PaymentIntent with confirm=true.
// Returns the PaymentIntent status (e.g., "succeeded", "requires_action").
func (s *StripeCustomerService) ChargeWithSavedMethod(ctx context.Context, userID string, paymentMethodID string, amountCents int64, orderID string) (string, error) {
	stripe.Key = s.secretKey

	customerID, err := s.GetOrCreateCustomer(ctx, userID)
	if err != nil {
		return "", err
	}

	// Verify the payment method belongs to this customer
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return "", fmt.Errorf("stripe: failed to get payment method: %w", err)
	}
	if pm.Customer == nil || pm.Customer.ID != customerID {
		return "", ErrPaymentMethodNotOwned
	}

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amountCents),
		Currency:      stripe.String("eur"),
		Customer:      stripe.String(customerID),
		PaymentMethod: stripe.String(paymentMethodID),
		OffSession:    stripe.Bool(true),
		Confirm:       stripe.Bool(true),
		Metadata: map[string]string{
			"order_id": orderID,
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe: failed to charge saved method: %w", err)
	}

	return string(pi.Status), nil
}
