package payment

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// StripeGateway implements PaymentGateway using Stripe Checkout Sessions.
type StripeGateway struct {
	secretKey  string
	successURL string // URL to redirect after successful payment
	cancelURL  string // URL to redirect on payment cancellation
}

// StripeConfig holds configuration for the Stripe gateway.
type StripeConfig struct {
	SecretKey  string // Stripe secret key (sk_test_... or sk_live_...)
	SuccessURL string // e.g., "http://localhost:5173/schedule?payment=success"
	CancelURL  string // e.g., "http://localhost:5173/schedule?payment=cancelled"
}

// NewStripeGateway creates a new Stripe-backed payment gateway.
func NewStripeGateway(cfg StripeConfig) *StripeGateway {
	return &StripeGateway{
		secretKey:  cfg.SecretKey,
		successURL: cfg.SuccessURL,
		cancelURL:  cfg.CancelURL,
	}
}

// CreateCheckoutURL creates a Stripe Checkout Session and returns its URL.
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
