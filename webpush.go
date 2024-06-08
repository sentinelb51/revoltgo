package revoltgo

type WebpushSubscription struct {

	// The URL to send the notification to
	Endpoint string `json:"endpoint"`

	// P-256 Diffie-Hellman public key
	P256DH string `json:"p256dh"`

	// Auth secret; used to authenticate the origin of the notification
	Auth string `json:"auth"`
}
