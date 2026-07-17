package payment

import (
	"context"
	"fmt"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// MockPaymentService is a mock implementation of domain.PaymentService
// that returns dummy payment links for testing/development.
type MockPaymentService struct{}

// NewMockPaymentService creates a new MockPaymentService.
func NewMockPaymentService() *MockPaymentService {
	return &MockPaymentService{}
}

// InitiatePayment generates a dummy single-use payment link.
func (s *MockPaymentService) InitiatePayment(_ context.Context, orderID string, amount int64) (*domain.PaymentLink, error) {
	return &domain.PaymentLink{
		URL:       fmt.Sprintf("https://pay.example.com/checkout/%s?amount=%d", orderID, amount),
		ExpiresIn: 1800, // 30 minutes
	}, nil
}

// ProcessPaymentCallback is a no-op for the mock.
func (s *MockPaymentService) ProcessPaymentCallback(_ context.Context, _ string, _ string) error {
	return nil
}

// InitiateRefund is a no-op for the mock. In production this would call the payment gateway.
func (s *MockPaymentService) InitiateRefund(_ context.Context, _ string, _ int64) error {
	return nil
}
