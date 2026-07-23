package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucatorrekens/bakery-app/internal/push"
)

func TestPushHandler_Subscribe_success(t *testing.T) {
	store := push.NewStore()
	sender := push.NewSender("pub-key", "priv-key", "test@example.com", store)
	handler := NewPushHandler(sender, store)

	body := `{"endpoint":"https://push.example.com/sub1","keys":{"p256dh":"test-p256dh","auth":"test-auth"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/push/subscribe", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	handler.Subscribe(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp["id"])

	// Verify subscription was stored
	subs := store.GetByUser("user-1")
	require.Len(t, subs, 1)
	assert.Equal(t, "https://push.example.com/sub1", subs[0].Endpoint)
	assert.Equal(t, "test-p256dh", subs[0].P256dh)
	assert.Equal(t, "test-auth", subs[0].Auth)
}

func TestPushHandler_Subscribe_rejectsUnauthenticated(t *testing.T) {
	store := push.NewStore()
	sender := push.NewSender("pub-key", "priv-key", "test@example.com", store)
	handler := NewPushHandler(sender, store)

	body := `{"endpoint":"https://push.example.com/sub1","keys":{"p256dh":"key","auth":"auth"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/push/subscribe", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	// No X-User-ID or JWT context
	w := httptest.NewRecorder()

	handler.Subscribe(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPushHandler_Subscribe_rejectsMissingFields(t *testing.T) {
	store := push.NewStore()
	sender := push.NewSender("pub-key", "priv-key", "test@example.com", store)
	handler := NewPushHandler(sender, store)

	tests := []struct {
		name string
		body string
	}{
		{"missing endpoint", `{"endpoint":"","keys":{"p256dh":"key","auth":"auth"}}`},
		{"missing p256dh", `{"endpoint":"https://push.example.com","keys":{"p256dh":"","auth":"auth"}}`},
		{"missing auth", `{"endpoint":"https://push.example.com","keys":{"p256dh":"key","auth":""}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/push/subscribe", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "user-1")
			w := httptest.NewRecorder()

			handler.Subscribe(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestPushHandler_Subscribe_rejectsInvalidJSON(t *testing.T) {
	store := push.NewStore()
	sender := push.NewSender("pub-key", "priv-key", "test@example.com", store)
	handler := NewPushHandler(sender, store)

	req := httptest.NewRequest(http.MethodPost, "/api/user/push/subscribe", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	handler.Subscribe(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushHandler_Unsubscribe_success(t *testing.T) {
	store := push.NewStore()
	store.Save(push.Subscription{
		ID:       "sub-1",
		UserID:   "user-1",
		Endpoint: "https://push.example.com/sub1",
		P256dh:   "key",
		Auth:     "auth",
	})
	sender := push.NewSender("pub-key", "priv-key", "test@example.com", store)
	handler := NewPushHandler(sender, store)

	body := `{"endpoint":"https://push.example.com/sub1"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/user/push/unsubscribe", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	handler.Unsubscribe(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify subscription was removed
	subs := store.GetByUser("user-1")
	assert.Nil(t, subs)
}

func TestPushHandler_Unsubscribe_rejectsUnauthenticated(t *testing.T) {
	store := push.NewStore()
	sender := push.NewSender("pub-key", "priv-key", "test@example.com", store)
	handler := NewPushHandler(sender, store)

	body := `{"endpoint":"https://push.example.com/sub1"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/user/push/unsubscribe", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Unsubscribe(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPushHandler_Unsubscribe_rejectsMissingEndpoint(t *testing.T) {
	store := push.NewStore()
	sender := push.NewSender("pub-key", "priv-key", "test@example.com", store)
	handler := NewPushHandler(sender, store)

	body := `{"endpoint":""}`
	req := httptest.NewRequest(http.MethodDelete, "/api/user/push/unsubscribe", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	handler.Unsubscribe(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushHandler_GetVAPIDKey_returnsPublicKey(t *testing.T) {
	store := push.NewStore()
	sender := push.NewSender("my-vapid-public-key", "priv-key", "test@example.com", store)
	handler := NewPushHandler(sender, store)

	req := httptest.NewRequest(http.MethodGet, "/api/push/vapid-key", nil)
	w := httptest.NewRecorder()

	handler.GetVAPIDKey(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "my-vapid-public-key", resp["publicKey"])
}
