package payment

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

// StripeWebhookHandler processes incoming Stripe webhook events.
type StripeWebhookHandler struct {
	webhookSecret string
	paymentSvc    domain.PaymentService
}

// NewStripeWebhookHandler creates a new webhook handler.
func NewStripeWebhookHandler(webhookSecret string, paymentSvc domain.PaymentService) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		webhookSecret: webhookSecret,
		paymentSvc:    paymentSvc,
	}
}

// HandleWebhook processes the webhook HTTP request from Stripe.
// It verifies the signature, parses the event, and triggers payment confirmation
// when a checkout.session.completed event is received.
func (h *StripeWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature to ensure the event is from Stripe
	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), h.webhookSecret)
	if err != nil {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Handle checkout.session.completed event
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}

		orderID := session.Metadata["order_id"]
		if orderID == "" {
			http.Error(w, "missing order_id in metadata", http.StatusBadRequest)
			return
		}

		// Process the payment callback (updates order status to Confirmed).
		// Log errors but return 200 to prevent Stripe from retrying for business logic failures.
		if err := h.paymentSvc.ProcessPaymentCallback(r.Context(), orderID, session.ID); err != nil {
			fmt.Printf("webhook: failed to process payment for order %s: %v\n", orderID, err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
