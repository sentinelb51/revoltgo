package revoltgo

//go:generate msgp -tests=false -io=false

type WebpushSubscription struct {

	// The URL to send the notification to
	Endpoint string `msg:"endpoint" json:"endpoint,omitempty"`

	// P-256 Diffie-Hellman public key
	P256DH string `msg:"p256dh" json:"p256dh,omitempty"`

	// Auth secret; used to authenticate the origin of the notification
	Auth string `msg:"auth" json:"auth,omitempty"`
}
