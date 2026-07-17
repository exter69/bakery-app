package payment

import "context"

// PaymentGateway abstracts the external payment provider (Stripe, PayPal, etc.).
// Implementations handle the actual interaction with third-party APIs.
type PaymentGateway interface {
	// CreateCheckoutURL generates a hosted payment page URL for the given order and amount.
	// The returned URL should direct the user to the payment provider's checkout page.
	CreateCheckoutURL(ctx context.Context, orderID string, amountCents int64) (string, error)

	// VerifyPayment confirms that a payment reference is valid and the payment succeeded.
	VerifyPayment(ctx context.Context, paymentRef string) (bool, error)
}
