package payment

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stripe/stripe-go/v82/webhook"
)

func TestConnectWebhookHandler_RejectsMissingSignature(t *testing.T) {
	handler := NewConnectWebhookHandler("whsec_test", nil)

	body := []byte(`{"type":"account.updated"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/stripe/connect-webhook", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestConnectWebhookHandler_RejectsInvalidSignature(t *testing.T) {
	handler := NewConnectWebhookHandler("whsec_test_secret", nil)

	body := []byte(`{"type":"account.updated","data":{"object":{}}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/stripe/connect-webhook", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", "t=1234567890,v1=invalidsig")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestConnectWebhookHandler_AcceptsValidAccountUpdatedEvent(t *testing.T) {
	secret := "whsec_test_secret_valid"
	handler := NewConnectWebhookHandler(secret, nil)

	payload := `{
		"id": "evt_test_123",
		"type": "account.updated",
		"api_version": "2025-08-27.basil",
		"data": {
			"object": {
				"id": "acct_123",
				"charges_enabled": true,
				"payouts_enabled": false,
				"requirements": {}
			}
		}
	}`

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/connect-webhook", bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Stripe-Signature", signedPayload.Header)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestConnectWebhookHandler_AcceptsUnknownEventType(t *testing.T) {
	secret := "whsec_test_secret_unknown"
	handler := NewConnectWebhookHandler(secret, nil)

	payload := `{
		"id": "evt_test_456",
		"type": "transfer.created",
		"api_version": "2025-08-27.basil",
		"data": {
			"object": {}
		}
	}`

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/connect-webhook", bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Stripe-Signature", signedPayload.Header)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for unhandled event type, got %d", rec.Code)
	}
}


// fakeBakeryRepo is a minimal BakeryRepository for connect webhook tests.
type fakeBakeryRepo struct {
	bakeries map[string]*domain.Bakery
	updated  []*domain.Bakery
}

func newFakeBakeryRepo() *fakeBakeryRepo {
	return &fakeBakeryRepo{
		bakeries: make(map[string]*domain.Bakery),
	}
}

func (r *fakeBakeryRepo) ListBakeries(_ context.Context, _ domain.PaginationParams) ([]domain.Bakery, int, error) {
	return nil, 0, nil
}
func (r *fakeBakeryRepo) GetBakery(_ context.Context, id string) (*domain.Bakery, error) {
	b, ok := r.bakeries[id]
	if !ok {
		return nil, nil
	}
	return b, nil
}
func (r *fakeBakeryRepo) GetBakeryByOwner(_ context.Context, _ string) (*domain.Bakery, error) {
	return nil, nil
}
func (r *fakeBakeryRepo) GetByStripeConnectID(_ context.Context, stripeConnectID string) (*domain.Bakery, error) {
	for _, b := range r.bakeries {
		if b.StripeConnectID == stripeConnectID {
			return b, nil
		}
	}
	return nil, nil
}
func (r *fakeBakeryRepo) UpdateBakery(_ context.Context, bakery *domain.Bakery) error {
	r.updated = append(r.updated, bakery)
	return nil
}
func (r *fakeBakeryRepo) GetProductsByBakery(_ context.Context, _ string) ([]domain.Product, error) {
	return nil, nil
}
func (r *fakeBakeryRepo) GetProductByID(_ context.Context, _ string) (*domain.Product, error) {
	return nil, nil
}
func (r *fakeBakeryRepo) CreateProduct(_ context.Context, _ *domain.Product) error { return nil }
func (r *fakeBakeryRepo) UpdateProduct(_ context.Context, _ *domain.Product) error { return nil }
func (r *fakeBakeryRepo) DeleteProduct(_ context.Context, _ string) error           { return nil }
func (r *fakeBakeryRepo) SearchProducts(_ context.Context, _ domain.ProductSearchParams) ([]domain.ProductSearchResult, int, error) {
	return nil, 0, nil
}

func TestConnectWebhookHandler_AccountUpdated_SyncsBakeryStatus(t *testing.T) {
	secret := "whsec_sync_test"
	bakeryRepo := newFakeBakeryRepo()
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{
		ID:              "bakery-1",
		OwnerID:         "seller-1",
		Name:            "Test Bakery",
		StripeConnectID: "acct_onboard_123",
	}

	handler := NewConnectWebhookHandler(secret, bakeryRepo)

	payload := `{
		"id": "evt_onboard",
		"type": "account.updated",
		"api_version": "2025-08-27.basil",
		"data": {
			"object": {
				"id": "acct_onboard_123",
				"charges_enabled": true,
				"payouts_enabled": true,
				"requirements": {}
			}
		}
	}`

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/connect-webhook", bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Stripe-Signature", signedPayload.Header)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Verify UpdateBakery was called
	if len(bakeryRepo.updated) != 1 {
		t.Fatalf("expected 1 UpdateBakery call, got %d", len(bakeryRepo.updated))
	}
	if bakeryRepo.updated[0].ID != "bakery-1" {
		t.Errorf("expected bakery-1 to be updated, got %s", bakeryRepo.updated[0].ID)
	}
	// Task 5.3: Assert ChargesEnabled and PayoutsEnabled are set on the updated bakery
	if !bakeryRepo.updated[0].ChargesEnabled {
		t.Errorf("expected ChargesEnabled=true on updated bakery, got false")
	}
	if !bakeryRepo.updated[0].PayoutsEnabled {
		t.Errorf("expected PayoutsEnabled=true on updated bakery, got false")
	}
}

func TestConnectWebhookHandler_AccountUpdated_NoopWhenBakeryNotFound(t *testing.T) {
	secret := "whsec_noop_test"
	bakeryRepo := newFakeBakeryRepo()
	// No bakery with this connect ID exists

	handler := NewConnectWebhookHandler(secret, bakeryRepo)

	payload := `{
		"id": "evt_unknown_acct",
		"type": "account.updated",
		"api_version": "2025-08-27.basil",
		"data": {
			"object": {
				"id": "acct_unknown_999",
				"charges_enabled": true,
				"payouts_enabled": true,
				"requirements": {}
			}
		}
	}`

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/connect-webhook", bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Stripe-Signature", signedPayload.Header)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (no error even if bakery not found), got %d", rec.Code)
	}

	// Verify UpdateBakery was NOT called
	if len(bakeryRepo.updated) != 0 {
		t.Errorf("expected 0 UpdateBakery calls, got %d", len(bakeryRepo.updated))
	}
}

// Task 5.1: Verify charges_enabled=true from Stripe sets the bakery's ChargesEnabled field.
func TestConnectWebhookHandler_AccountUpdated_SetsOnboardingFlags(t *testing.T) {
	secret := "whsec_onboard_flags"
	bakeryRepo := newFakeBakeryRepo()
	bakeryRepo.bakeries["bakery-flags"] = &domain.Bakery{
		ID:              "bakery-flags",
		OwnerID:         "seller-1",
		Name:            "Flag Bakery",
		StripeConnectID: "acct_flags_456",
		ChargesEnabled:  false,
		PayoutsEnabled:  false,
	}

	handler := NewConnectWebhookHandler(secret, bakeryRepo)

	payload := `{
		"id": "evt_flags_test",
		"type": "account.updated",
		"api_version": "2025-08-27.basil",
		"data": {
			"object": {
				"id": "acct_flags_456",
				"charges_enabled": true,
				"payouts_enabled": false,
				"requirements": {}
			}
		}
	}`

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/connect-webhook", bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Stripe-Signature", signedPayload.Header)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if len(bakeryRepo.updated) != 1 {
		t.Fatalf("expected 1 UpdateBakery call, got %d", len(bakeryRepo.updated))
	}

	updated := bakeryRepo.updated[0]
	if !updated.ChargesEnabled {
		t.Errorf("expected ChargesEnabled=true, got false")
	}
	if updated.PayoutsEnabled {
		t.Errorf("expected PayoutsEnabled=false, got true")
	}
}

// Task 5.2: Verify no DB write when stored values already match incoming webhook values.
func TestConnectWebhookHandler_AccountUpdated_IdempotentSkipsUpdate(t *testing.T) {
	secret := "whsec_idempotent"
	bakeryRepo := newFakeBakeryRepo()
	// Bakery already has the same values as what the webhook will send
	bakeryRepo.bakeries["bakery-idem"] = &domain.Bakery{
		ID:              "bakery-idem",
		OwnerID:         "seller-1",
		Name:            "Idempotent Bakery",
		StripeConnectID: "acct_idem_789",
		ChargesEnabled:  true,
		PayoutsEnabled:  true,
	}

	handler := NewConnectWebhookHandler(secret, bakeryRepo)

	payload := `{
		"id": "evt_idem_test",
		"type": "account.updated",
		"api_version": "2025-08-27.basil",
		"data": {
			"object": {
				"id": "acct_idem_789",
				"charges_enabled": true,
				"payouts_enabled": true,
				"requirements": {}
			}
		}
	}`

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/connect-webhook", bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Stripe-Signature", signedPayload.Header)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// No DB write should have happened since values already match
	if len(bakeryRepo.updated) != 0 {
		t.Errorf("expected 0 UpdateBakery calls (idempotent skip), got %d", len(bakeryRepo.updated))
	}
}
