package revoltgo

//go:generate msgp -tests=false -io=false

type LoginResponse struct {
	Result       string              `msg:"result"`
	ID           string              `msg:"_id"`
	UserID       string              `msg:"user_id"`
	Token        string              `msg:"token"`
	Name         string              `msg:"name"`
	Subscription WebpushSubscription `msg:"subscription"`
}

type Sessions struct {
	ID   string `msg:"_id"`
	Name string `msg:"name"`
}
type Account struct {
	ID    string `msg:"_id"`
	Email string `msg:"email"`
}

type Onboarding struct {
	Onboarding bool `msg:"onboarding"`
}

type MFA struct {
	// Unvalidated or authorised MFA ticket; used to resolve the correct account
	MfaTicket string `msg:"mfa_ticket"`

	// MFA response
	MfaResponse MFAResponse `msg:"mfa_response"`

	// Friendly name used for the session
	FriendlyName string `msg:"friendly_name"`
}

type MFAResponse struct {
	Password string `msg:"password"`
}

type ChangeEmail struct {
	Ticket Ticket `msg:"ticket"` // Why is this nested
}

type Ticket struct {
	ID           string `msg:"_id"`
	AccountID    string `msg:"account_id"`
	Token        string `msg:"token"`
	Validated    bool   `msg:"validated"`
	Authorised   bool   `msg:"authorised"`
	LastTOTPCode string `msg:"last_totp_code"`
}
