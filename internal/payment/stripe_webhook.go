package payment

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

// PayoutReverser is the interface for reversing payouts on refund.
type PayoutReverser interface {
	OnOrderRefunded(ctx context.Context, orderID string) error
}

// StripeWebhookHandler processes incoming Stripe webhook events.
type StripeWebhookHandler struct {
	webhookSecret  string
	paymentSvc     domain.PaymentService
	orderRepo      domain.OrderRepository
	payoutReverser PayoutReverser
}

// NewStripeWebhookHandler creates a new webhook handler.
func NewStripeWebhookHandler(webhookSecret string, paymentSvc domain.PaymentService, orderRepo domain.OrderRepository) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		webhookSecret: webhookSecret,
		paymentSvc:    paymentSvc,
		orderRepo:     orderRepo,
	}
}

// SetPayoutReverser sets the payout reversal callback (called from main wiring to avoid import cycles).
func (h *StripeWebhookHandler) SetPayoutReverser(reverser PayoutReverser) {
	h.payoutReverser = reverser
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
			log.Printf("[WEBHOOK] failed to process payment for order %s: %v", orderID, err)
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

// updateRefundStatus looks up the order by PaymentIntent ID and persists the refund state.
// Idempotent: if the order already has the same refund status, it's a no-op.
func (h *StripeWebhookHandler) updateRefundStatus(ctx context.Context, paymentIntentID string, refundedAmount int64, totalAmount int64) {
	status := "refunded"
	if refundedAmount > 0 && refundedAmount < totalAmount {
		status = "partial"
	}

	order, err := h.orderRepo.GetByPaymentIntentID(ctx, paymentIntentID)
	if err != nil {
		log.Printf("[WEBHOOK] failed to look up order for PI %s: %v", paymentIntentID, err)
		return
	}
	if order == nil {
		log.Printf("[WEBHOOK] no order found for PI %s", paymentIntentID)
		return
	}

	// Idempotent: skip if already at the target status
	if order.RefundStatus == status {
		return
	}

	order.RefundStatus = status
	if err := h.orderRepo.Save(ctx, order); err != nil {
		log.Printf("[WEBHOOK] failed to update refund status for order %s: %v", order.ID, err)
		return
	}

	// Trigger payout reversal for fully refunded orders
	if status == "refunded" && h.payoutReverser != nil {
		if err := h.payoutReverser.OnOrderRefunded(ctx, order.ID); err != nil {
			log.Printf("[WEBHOOK] payout reversal failed for order %s: %v", order.ID, err)
		}
	}
}
