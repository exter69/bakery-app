package payment

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82/webhook"
)

// mockPayoutReverser tracks payout reversal calls.
type mockPayoutReverser struct {
	reversedOrders []string
}

func (m *mockPayoutReverser) OnOrderRefunded(_ context.Context, orderID string) error {
	m.reversedOrders = append(m.reversedOrders, orderID)
	return nil
}

// --- updateRefundStatus tests ---

func TestUpdateRefundStatus_PersistsRefundedStatus(t *testing.T) {
	orderRepo := newFakeOrderRepo()
	handler := NewStripeWebhookHandler("whsec_test", nil, orderRepo)

	order := &domain.Order{
		ID:              "order-1",
		PaymentIntentID: "pi_test_123",
		Status:          domain.OrderStatusDelivered,
		RefundStatus:    "",
	}
	_ = orderRepo.Save(context.Background(), order)

	handler.updateRefundStatus(context.Background(), "pi_test_123", 5000, 5000)

	updated, _ := orderRepo.GetByID(context.Background(), "order-1")
	assert.Equal(t, "refunded", updated.RefundStatus)
}

func TestUpdateRefundStatus_PersistsPartialStatus(t *testing.T) {
	orderRepo := newFakeOrderRepo()
	handler := NewStripeWebhookHandler("whsec_test", nil, orderRepo)

	order := &domain.Order{
		ID:              "order-1",
		PaymentIntentID: "pi_test_456",
		Status:          domain.OrderStatusDelivered,
		RefundStatus:    "",
	}
	_ = orderRepo.Save(context.Background(), order)

	handler.updateRefundStatus(context.Background(), "pi_test_456", 2500, 5000)

	updated, _ := orderRepo.GetByID(context.Background(), "order-1")
	assert.Equal(t, "partial", updated.RefundStatus)
}

func TestUpdateRefundStatus_IsIdempotent(t *testing.T) {
	orderRepo := newFakeOrderRepo()
	handler := NewStripeWebhookHandler("whsec_test", nil, orderRepo)

	order := &domain.Order{
		ID:              "order-1",
		PaymentIntentID: "pi_test_789",
		Status:          domain.OrderStatusDelivered,
		RefundStatus:    "refunded", // already set
	}
	_ = orderRepo.Save(context.Background(), order)

	// Replaying the event should be a no-op
	handler.updateRefundStatus(context.Background(), "pi_test_789", 5000, 5000)

	updated, _ := orderRepo.GetByID(context.Background(), "order-1")
	assert.Equal(t, "refunded", updated.RefundStatus)
}

func TestUpdateRefundStatus_NoopWhenOrderNotFound(t *testing.T) {
	orderRepo := newFakeOrderRepo()
	handler := NewStripeWebhookHandler("whsec_test", nil, orderRepo)

	// Should not panic or error — just log
	handler.updateRefundStatus(context.Background(), "pi_nonexistent", 5000, 5000)
}

func TestUpdateRefundStatus_TriggersPayoutReversal(t *testing.T) {
	orderRepo := newFakeOrderRepo()
	reverser := &mockPayoutReverser{}
	handler := NewStripeWebhookHandler("whsec_test", nil, orderRepo)
	handler.SetPayoutReverser(reverser)

	order := &domain.Order{
		ID:              "order-1",
		PaymentIntentID: "pi_reversal_test",
		Status:          domain.OrderStatusDelivered,
		RefundStatus:    "",
	}
	_ = orderRepo.Save(context.Background(), order)

	handler.updateRefundStatus(context.Background(), "pi_reversal_test", 5000, 5000)

	assert.Equal(t, []string{"order-1"}, reverser.reversedOrders)
}

func TestUpdateRefundStatus_SkipsPayoutReversalForPartialRefund(t *testing.T) {
	orderRepo := newFakeOrderRepo()
	reverser := &mockPayoutReverser{}
	handler := NewStripeWebhookHandler("whsec_test", nil, orderRepo)
	handler.SetPayoutReverser(reverser)

	order := &domain.Order{
		ID:              "order-1",
		PaymentIntentID: "pi_partial_test",
		Status:          domain.OrderStatusDelivered,
		RefundStatus:    "",
	}
	_ = orderRepo.Save(context.Background(), order)

	handler.updateRefundStatus(context.Background(), "pi_partial_test", 2500, 5000)

	// Partial refund should not trigger payout reversal
	assert.Empty(t, reverser.reversedOrders)
}

// --- Webhook endpoint integration tests ---

func TestStripeWebhook_ChargeRefunded_UpdatesOrder(t *testing.T) {
	secret := "whsec_charge_refund_test"
	orderRepo := newFakeOrderRepo()
	handler := NewStripeWebhookHandler(secret, nil, orderRepo)

	order := &domain.Order{
		ID:              "order-webhook-1",
		PaymentIntentID: "pi_1234567890",
		Status:          domain.OrderStatusDelivered,
		RefundStatus:    "",
	}
	_ = orderRepo.Save(context.Background(), order)

	payload := `{
		"id": "evt_charge_refunded",
		"type": "charge.refunded",
		"api_version": "2025-08-27.basil",
		"data": {
			"object": {
				"id": "ch_test_123",
				"amount": 5000,
				"amount_refunded": 5000,
				"payment_intent": "pi_1234567890"
			}
		}
	}`

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Stripe-Signature", signedPayload.Header)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	updated, _ := orderRepo.GetByID(context.Background(), "order-webhook-1")
	assert.Equal(t, "refunded", updated.RefundStatus)
}

func TestStripeWebhook_ChargeRefunded_IdempotentReplay(t *testing.T) {
	secret := "whsec_idempotent_test"
	orderRepo := newFakeOrderRepo()
	reverser := &mockPayoutReverser{}
	handler := NewStripeWebhookHandler(secret, nil, orderRepo)
	handler.SetPayoutReverser(reverser)

	order := &domain.Order{
		ID:              "order-replay",
		PaymentIntentID: "pi_replay_test",
		Status:          domain.OrderStatusDelivered,
		RefundStatus:    "refunded", // already refunded
	}
	_ = orderRepo.Save(context.Background(), order)

	payload := `{
		"id": "evt_replay",
		"type": "charge.refunded",
		"api_version": "2025-08-27.basil",
		"data": {
			"object": {
				"id": "ch_replay",
				"amount": 3000,
				"amount_refunded": 3000,
				"payment_intent": "pi_replay_test"
			}
		}
	}`

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Stripe-Signature", signedPayload.Header)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	// Payout reversal should NOT be triggered since it was already refunded (idempotent)
	assert.Empty(t, reverser.reversedOrders)
}

func TestStripeWebhook_RejectsInvalidSignature(t *testing.T) {
	handler := NewStripeWebhookHandler("whsec_real_secret", nil, nil)

	body := []byte(`{"type":"charge.refunded"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", "t=1234567890,v1=invalidsig")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
