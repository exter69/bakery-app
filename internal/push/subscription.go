package push

// Subscription holds a Web Push subscription endpoint and keys for a user.
type Subscription struct {
	ID       string `json:"id"`
	UserID   string `json:"userID"`
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"` // client public key
	Auth     string `json:"auth"`   // auth secret
}
