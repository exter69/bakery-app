package payment

import (
	"context"
	"fmt"

	"github.com/lucatorrekens/bakery-app/internal/payment/token"
)

// StubGateway is a development/testing implementation of PaymentGateway
// that generates deterministic URLs without calling external services.
type StubGateway struct{}

// NewStubGateway creates a new StubGateway.
func NewStubGateway() *StubGateway {
	return &StubGateway{}
}

// CreateCheckoutURL generates a fake payment URL with a unique token.
func (g *StubGateway) CreateCheckoutURL(_ context.Context, orderID string, _ int64) (string, error) {
	tok := token.Generate()
	return fmt.Sprintf("https://pay.example.com/checkout/%s?token=%s", orderID, tok), nil
}

// VerifyPayment always returns true for the stub gateway.
func (g *StubGateway) VerifyPayment(_ context.Context, _ string) (bool, error) {
	return true, nil
}

// CapturePayment is a no-op for the stub gateway.
func (g *StubGateway) CapturePayment(_ context.Context, _ string) error {
	return nil
}

// VoidAuthorization is a no-op for the stub gateway.
func (g *StubGateway) VoidAuthorization(_ context.Context, _ string) error {
	return nil
}

// RefundPayment is a no-op for the stub gateway.
func (g *StubGateway) RefundPayment(_ context.Context, _ string, _ int64) error {
	return nil
}
