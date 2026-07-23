package payment

import (
	"context"
	"fmt"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/refund"
)

// StripeGateway implements PaymentGateway using Stripe Checkout Sessions.
type StripeGateway struct {
	secretKey  string
	successURL string // URL to redirect after successful payment
	cancelURL  string // URL to redirect on payment cancellation
	userRepo   domain.UserRepository  // optional — enables saved card association
	orderRepo  domain.OrderRepository // optional — used to look up order's user for customer linking
}

// StripeConfig holds configuration for the Stripe gateway.
type StripeConfig struct {
	SecretKey  string // Stripe secret key (sk_test_... or sk_live_...)
	SuccessURL string // e.g., "http://localhost:5173/schedule?payment=success"
	CancelURL  string // e.g., "http://localhost:5173/schedule?payment=cancelled"
	UserRepo   domain.UserRepository  // optional — if set, checkout sessions link to the Stripe Customer
	OrderRepo  domain.OrderRepository // optional — used to look up order's user for customer linking
}

// NewStripeGateway creates a new Stripe-backed payment gateway.
func NewStripeGateway(cfg StripeConfig) *StripeGateway {
	return &StripeGateway{
		secretKey:  cfg.SecretKey,
		successURL: cfg.SuccessURL,
		cancelURL:  cfg.CancelURL,
		userRepo:   cfg.UserRepo,
		orderRepo:  cfg.OrderRepo,
	}
}

// CreateCheckoutURL creates a Stripe Checkout Session and returns its URL.
// If the order's user has a Stripe Customer ID, the session is linked to that customer
// and new cards are saved for future use via setup_future_usage.
func (g *StripeGateway) CreateCheckoutURL(ctx context.Context, orderID string, amountCents int64) (string, error) {
	stripe.Key = g.secretKey

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("eur"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("Order %s", orderID)),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(fmt.Sprintf("%s&order_id=%s", g.successURL, orderID)),
		CancelURL:  stripe.String(fmt.Sprintf("%s&order_id=%s", g.cancelURL, orderID)),
		Metadata: map[string]string{
			"order_id": orderID,
		},
	}

	// If repos are available, link the checkout session to the Stripe Customer.
	// This enables Stripe Checkout to display saved cards and save new ones.
	if g.orderRepo != nil && g.userRepo != nil {
		if order, err := g.orderRepo.GetByID(ctx, orderID); err == nil && order != nil {
			if user, err := g.userRepo.GetByID(ctx, order.UserID); err == nil && user != nil && user.StripeCustomerID != "" {
				params.Customer = stripe.String(user.StripeCustomerID)
			}
		}
	}

	// Delayed capture: authorize funds at checkout, capture on delivery
	if params.PaymentIntentData == nil {
		params.PaymentIntentData = &stripe.CheckoutSessionPaymentIntentDataParams{}
	}
	params.PaymentIntentData.CaptureMethod = stripe.String("manual")
	params.PaymentIntentData.SetupFutureUsage = stripe.String("on_session")

	s, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe: failed to create checkout session: %w", err)
	}

	return s.URL, nil
}

// VerifyPayment verifies a Stripe Checkout Session by its ID.
func (g *StripeGateway) VerifyPayment(ctx context.Context, sessionID string) (bool, error) {
	stripe.Key = g.secretKey

	s, err := session.Get(sessionID, nil)
	if err != nil {
		return false, fmt.Errorf("stripe: failed to retrieve session: %w", err)
	}

	return s.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid, nil
}

// CapturePayment captures a previously authorized (manual-capture) PaymentIntent.
func (g *StripeGateway) CapturePayment(_ context.Context, paymentIntentID string) error {
	stripe.Key = g.secretKey

	_, err := paymentintent.Capture(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("stripe: failed to capture payment %s: %w", paymentIntentID, err)
	}
	return nil
}

// VoidAuthorization cancels a previously authorized PaymentIntent without capturing.
func (g *StripeGateway) VoidAuthorization(_ context.Context, paymentIntentID string) error {
	stripe.Key = g.secretKey

	_, err := paymentintent.Cancel(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("stripe: failed to void authorization %s: %w", paymentIntentID, err)
	}
	return nil
}

// RefundPayment issues a refund for a previously captured PaymentIntent.
// If amountCents is 0, a full refund is issued; otherwise a partial refund of that amount.
func (g *StripeGateway) RefundPayment(_ context.Context, paymentIntentID string, amountCents int64) error {
	stripe.Key = g.secretKey

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}
	if amountCents > 0 {
		params.Amount = stripe.Int64(amountCents)
	}

	_, err := refund.New(params)
	if err != nil {
		return fmt.Errorf("stripe: failed to refund payment %s: %w", paymentIntentID, err)
	}
	return nil
}
