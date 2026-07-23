package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/push"
)

// PushHandler handles HTTP requests for push notification subscription management.
type PushHandler struct {
	sender *push.Sender
	store  *push.Store
}

// NewPushHandler creates a new PushHandler.
func NewPushHandler(sender *push.Sender, store *push.Store) *PushHandler {
	return &PushHandler{sender: sender, store: store}
}

// Subscribe handles POST /api/user/push/subscribe.
// Saves a new push subscription for the authenticated user.
func (h *PushHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" || userID == "anonymous" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return
	}

	var req dto.PushSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "endpoint, keys.p256dh, and keys.auth are required",
		})
		return
	}

	sub := push.Subscription{
		ID:       generateSubscriptionID(),
		UserID:   userID,
		Endpoint: req.Endpoint,
		P256dh:   req.Keys.P256dh,
		Auth:     req.Keys.Auth,
	}

	h.store.Save(sub)

	writeJSON(w, http.StatusCreated, dto.PushSubscribeResponse{
		ID: sub.ID,
	})
}

// Unsubscribe handles DELETE /api/user/push/unsubscribe.
// Removes a push subscription for the authenticated user by endpoint.
func (h *PushHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" || userID == "anonymous" {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "authentication required",
		})
		return
	}

	var req dto.PushUnsubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	if req.Endpoint == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "endpoint is required",
		})
		return
	}

	h.store.DeleteByEndpoint(userID, req.Endpoint)

	w.WriteHeader(http.StatusNoContent)
}

// GetVAPIDKey handles GET /api/push/vapid-key.
// Returns the public VAPID key needed by the browser to subscribe.
// This is a public endpoint (no auth required).
func (h *PushHandler) GetVAPIDKey(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, dto.VAPIDKeyResponse{
		PublicKey: h.sender.VAPIDPublicKey(),
	})
}

// generateSubscriptionID creates a random hex ID for a subscription.
func generateSubscriptionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback: won't happen in practice on a system with /dev/urandom
		return "sub-fallback"
	}
	return hex.EncodeToString(b)
}
