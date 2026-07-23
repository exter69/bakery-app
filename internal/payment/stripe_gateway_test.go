package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStripeGateway_SetsConfig(t *testing.T) {
	cfg := StripeConfig{
		SecretKey:  "sk_test_abc123",
		SuccessURL: "http://localhost:5173/schedule?payment=success",
		CancelURL:  "http://localhost:5173/schedule?payment=cancelled",
	}

	gw := NewStripeGateway(cfg)

	assert.Equal(t, "sk_test_abc123", gw.secretKey)
	assert.Equal(t, "http://localhost:5173/schedule?payment=success", gw.successURL)
	assert.Equal(t, "http://localhost:5173/schedule?payment=cancelled", gw.cancelURL)
}

func TestStripeGateway_ImplementsInterface(t *testing.T) {
	// Compile-time check that StripeGateway implements PaymentGateway.
	var _ PaymentGateway = (*StripeGateway)(nil)
}

func TestStripeGateway_CreateCheckoutURL_InvalidKey(t *testing.T) {
	gw := NewStripeGateway(StripeConfig{
		SecretKey:  "sk_test_invalid",
		SuccessURL: "http://localhost/success",
		CancelURL:  "http://localhost/cancel",
	})

	// With an invalid key, Stripe's API will return an auth error.
	_, err := gw.CreateCheckoutURL(context.Background(), "order-123", 5000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stripe: failed to create checkout session")
}

func TestStripeGateway_VerifyPayment_InvalidKey(t *testing.T) {
	gw := NewStripeGateway(StripeConfig{
		SecretKey:  "sk_test_invalid",
		SuccessURL: "http://localhost/success",
		CancelURL:  "http://localhost/cancel",
	})

	// With an invalid key, Stripe's API will return an auth error.
	_, err := gw.VerifyPayment(context.Background(), "cs_test_fake_session")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stripe: failed to retrieve session")
}

func TestStripeWebhookHandler_InvalidSignature(t *testing.T) {
	handler := NewStripeWebhookHandler("whsec_test_secret", NewMockPaymentService(), nil)

	// Send request without a valid Stripe-Signature header
	req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", strings.NewReader(`{"type":"checkout.session.completed"}`))
	req.Header.Set("Stripe-Signature", "invalid_sig")
	w := httptest.NewRecorder()

	handler.HandleWebhook(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStripeWebhookHandler_EmptyBody(t *testing.T) {
	handler := NewStripeWebhookHandler("whsec_test_secret", NewMockPaymentService(), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", strings.NewReader(""))
	req.Header.Set("Stripe-Signature", "t=123,v1=abc")
	w := httptest.NewRecorder()

	handler.HandleWebhook(w, req)

	// Empty body with invalid signature should still fail signature verification
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStripeWebhookHandler_MissingSignatureHeader(t *testing.T) {
	handler := NewStripeWebhookHandler("whsec_test_secret", NewMockPaymentService(), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", strings.NewReader(`{}`))
	// No Stripe-Signature header
	w := httptest.NewRecorder()

	handler.HandleWebhook(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
