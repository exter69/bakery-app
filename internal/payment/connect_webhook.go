package payment

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// ConnectWebhookHandler handles Stripe Connect account.updated webhooks.
type ConnectWebhookHandler struct {
	webhookSecret string
	bakeryRepo    domain.BakeryRepository
}

// NewConnectWebhookHandler creates a new ConnectWebhookHandler.
func NewConnectWebhookHandler(webhookSecret string, bakeryRepo domain.BakeryRepository) *ConnectWebhookHandler {
	return &ConnectWebhookHandler{
		webhookSecret: webhookSecret,
		bakeryRepo:    bakeryRepo,
	}
}

// HandleWebhook processes Stripe Connect webhook events.
// Route: POST /api/stripe/connect-webhook
func (h *ConnectWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), h.webhookSecret)
	if err != nil {
		http.Error(w, "signature verification failed", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "account.updated":
		var acct stripe.Account
		if err := json.Unmarshal(event.Data.Raw, &acct); err != nil {
			log.Printf("[CONNECT-WEBHOOK] failed to unmarshal account: %v", err)
			w.WriteHeader(http.StatusOK)
			return
		}
		h.handleAccountUpdated(r.Context(), &acct)
	}

	w.WriteHeader(http.StatusOK)
}

// handleAccountUpdated syncs the bakery's Connect onboarding status from Stripe.
// When charges_enabled becomes true, the bakery is considered fully onboarded.
func (h *ConnectWebhookHandler) handleAccountUpdated(ctx context.Context, acct *stripe.Account) {
	log.Printf("[CONNECT-WEBHOOK] account %s updated: charges_enabled=%t, payouts_enabled=%t",
		acct.ID, acct.ChargesEnabled, acct.PayoutsEnabled)

	if h.bakeryRepo == nil {
		return
	}

	bakery, err := h.bakeryRepo.GetByStripeConnectID(ctx, acct.ID)
	if err != nil {
		log.Printf("[CONNECT-WEBHOOK] failed to look up bakery for account %s: %v", acct.ID, err)
		return
	}
	if bakery == nil {
		log.Printf("[CONNECT-WEBHOOK] no bakery found for account %s", acct.ID)
		return
	}

	// Idempotency: skip the DB write if values already match
	if bakery.ChargesEnabled == acct.ChargesEnabled && bakery.PayoutsEnabled == acct.PayoutsEnabled {
		log.Printf("[CONNECT-WEBHOOK] bakery %s already up to date, skipping write", bakery.ID)
		return
	}

	bakery.ChargesEnabled = acct.ChargesEnabled
	bakery.PayoutsEnabled = acct.PayoutsEnabled

	if err := h.bakeryRepo.UpdateBakery(ctx, bakery); err != nil {
		log.Printf("[CONNECT-WEBHOOK] failed to update bakery %s: %v", bakery.ID, err)
	}
}
