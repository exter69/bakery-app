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

	// CapturePayment captures a previously authorized (manual-capture) payment.
	// Called when an order transitions to Delivered.
	CapturePayment(ctx context.Context, paymentIntentID string) error

	// VoidAuthorization cancels a previously authorized payment without capturing.
	// Called when an order is cancelled before delivery.
	VoidAuthorization(ctx context.Context, paymentIntentID string) error

	// RefundPayment issues a refund for a previously captured PaymentIntent.
	// amountCents = 0 means full refund; > 0 means partial refund of that amount.
	RefundPayment(ctx context.Context, paymentIntentID string, amountCents int64) error
}
