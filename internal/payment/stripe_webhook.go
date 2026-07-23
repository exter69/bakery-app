package payment

import (
	"context"
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
	orderRepo     domain.OrderRepository
}

// NewStripeWebhookHandler creates a new webhook handler.
func NewStripeWebhookHandler(webhookSecret string, paymentSvc domain.PaymentService, orderRepo domain.OrderRepository) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		webhookSecret: webhookSecret,
		paymentSvc:    paymentSvc,
		orderRepo:     orderRepo,
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

		// Pass the PaymentIntent ID so it can be stored on the order for delayed capture
		paymentRef := session.PaymentIntent.ID
		if paymentRef == "" {
			// Fallback to session ID if PaymentIntent is not expanded
			paymentRef = session.ID
		}

		// Process the payment callback (updates order status to Confirmed).
		// Log errors but return 200 to prevent Stripe from retrying for business logic failures.
		if err := h.paymentSvc.ProcessPaymentCallback(r.Context(), orderID, paymentRef); err != nil {
			fmt.Printf("webhook: failed to process payment for order %s: %v\n", orderID, err)
		}
	}

	// Handle charge.refunded — update refund_status idempotently
	if event.Type == "charge.refunded" {
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			http.Error(w, "failed to parse charge.refunded event", http.StatusBadRequest)
			return
		}

		piID := ""
		if charge.PaymentIntent != nil {
			piID = charge.PaymentIntent.ID
		}
		if piID != "" && h.orderRepo != nil {
			h.updateRefundStatus(r.Context(), piID, charge.AmountRefunded, charge.Amount)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// updateRefundStatus logs the refund event from Stripe.
// The order_service already sets RefundStatus="refunded" when it issues the refund;
// this webhook acts as an idempotent confirmation. A PI→order lookup can be added later.
func (h *StripeWebhookHandler) updateRefundStatus(_ context.Context, paymentIntentID string, refundedAmount int64, totalAmount int64) {
	status := "refunded"
	if refundedAmount > 0 && refundedAmount < totalAmount {
		status = "partial"
	}
	fmt.Printf("webhook: charge.refunded for PI %s — status=%s (refunded=%d, total=%d)\n",
		paymentIntentID, status, refundedAmount, totalAmount)
}
