package push

import (
	"encoding/json"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// PushMessage is the payload sent in a push notification.
type PushMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"` // click target URL
}

// Sender handles sending Web Push notifications to users via their subscriptions.
type Sender struct {
	vapidPublicKey  string
	vapidPrivateKey string
	contactEmail    string
	store           *Store
}

// NewSender creates a new push notification sender.
func NewSender(publicKey, privateKey, contactEmail string, store *Store) *Sender {
	return &Sender{
		vapidPublicKey:  publicKey,
		vapidPrivateKey: privateKey,
		contactEmail:    contactEmail,
		store:           store,
	}
}

// VAPIDPublicKey returns the public VAPID key for client subscription requests.
func (s *Sender) VAPIDPublicKey() string {
	return s.vapidPublicKey
}

// SendToUser sends a push notification to all subscriptions for a given user.
// Expired subscriptions (HTTP 410) are automatically removed.
func (s *Sender) SendToUser(userID string, msg PushMessage) {
	subs := s.store.GetByUser(userID)
	if len(subs) == 0 {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[PUSH] failed to marshal message for user %s: %v", userID, err)
		return
	}

	for _, sub := range subs {
		resp, err := webpush.SendNotification(payload, &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}, &webpush.Options{
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivateKey,
			Subscriber:      s.contactEmail,
		})
		if err != nil {
			log.Printf("[PUSH] failed to send to user %s endpoint %s: %v", userID, sub.Endpoint, err)
			continue
		}
		resp.Body.Close()

		// HTTP 410 Gone means the subscription expired — clean it up
		if resp.StatusCode == 410 {
			log.Printf("[PUSH] subscription expired for user %s, removing", userID)
			s.store.Delete(sub.ID)
		}
	}
}
