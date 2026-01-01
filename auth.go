package revoltgo

//go:generate msgp -tests=false -io=false

type LoginResponse struct {
	Result       string              `msg:"result" json:"result,omitempty"`
	ID           string              `msg:"_id" json:"_id,omitempty"`
	UserID       string              `msg:"user_id" json:"user_id,omitempty"`
	Token        string              `msg:"token" json:"token,omitempty"`
	Name         string              `msg:"name" json:"name,omitempty"`
	Subscription WebpushSubscription `msg:"subscription" json:"subscription,omitempty"`
}

type Sessions struct {
	ID   string `msg:"_id" json:"_id,omitempty"`
	Name string `msg:"name" json:"name,omitempty"`
}
type Account struct {
	ID    string `msg:"_id" json:"_id,omitempty"`
	Email string `msg:"email" json:"email,omitempty"`
}

type Onboarding struct {
	Onboarding bool `msg:"onboarding" json:"onboarding,omitempty"`
}

type MFA struct {
	// Unvalidated or authorised MFA ticket; used to resolve the correct account
	MfaTicket string `msg:"mfa_ticket" json:"mfa_ticket,omitempty"`

	// MFA response
	MfaResponse MFAResponse `msg:"mfa_response" json:"mfa_response,omitempty"`

	// Friendly name used for the session
	FriendlyName string `msg:"friendly_name" json:"friendly_name,omitempty"`
}

type MFAResponse struct {
	Password string `msg:"password" json:"password,omitempty"`
}

type ChangeEmail struct {
	Ticket Ticket `msg:"ticket" json:"ticket,omitempty"` // Why is this nested
}

type Ticket struct {
	ID           string `msg:"_id" json:"_id,omitempty"`
	AccountID    string `msg:"account_id" json:"account_id,omitempty"`
	Token        string `msg:"token" json:"token,omitempty"`
	Validated    bool   `msg:"validated" json:"validated,omitempty"`
	Authorised   bool   `msg:"authorised" json:"authorised,omitempty"`
	LastTOTPCode string `msg:"last_totp_code" json:"last_totp_code,omitempty"`
}
