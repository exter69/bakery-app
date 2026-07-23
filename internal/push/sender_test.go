package push

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSender_VAPIDPublicKey(t *testing.T) {
	store := NewStore()
	sender := NewSender("public-key", "private-key", "admin@example.com", store)

	assert.Equal(t, "public-key", sender.VAPIDPublicKey())
}

func TestSender_SendToUser_noSubscriptions(t *testing.T) {
	store := NewStore()
	sender := NewSender("public-key", "private-key", "admin@example.com", store)

	// Should not panic with no subscriptions
	sender.SendToUser("unknown-user", PushMessage{
		Title: "Test",
		Body:  "Hello",
	})
}

func TestSender_SendToUser_invalidEndpointDoesNotPanic(t *testing.T) {
	store := NewStore()
	store.Save(Subscription{
		ID:       "sub-1",
		UserID:   "user-1",
		Endpoint: "https://invalid-endpoint.example.com/push",
		P256dh:   "invalid-key",
		Auth:     "invalid-auth",
	})

	sender := NewSender("public-key", "private-key", "admin@example.com", store)

	// Should log the error but not panic
	sender.SendToUser("user-1", PushMessage{
		Title: "Test",
		Body:  "Hello",
	})
}
