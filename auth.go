package revoltgo

//go:generate msgp -tests=false -io=false

type AuthMFAMethod string

const (
	AuthMFAMethodPassword AuthMFAMethod = "Password"
	AuthMFAMethodRecovery AuthMFAMethod = "Recovery"
	AuthMFAMethodTOTP     AuthMFAMethod = "Totp"
)

type AuthMFAResponse struct {
	EmailOTP        bool `msg:"email_otp" json:"email_otp"`
	TrustedHandover bool `msg:"trusted_handover" json:"trusted_handover"`
	EmailMFA        bool `msg:"email_mfa" json:"email_mfa"`
	TotpMFA         bool `msg:"totp_mfa" json:"totp_mfa"`
	SecurityKeyMFA  bool `msg:"security_key_mfa" json:"security_key_mfa"`
	RecoveryActive  bool `msg:"recovery_active" json:"recovery_active"`
}

type AuthMFATicketResponse struct {
	MFATicket
}

type AuthMFATOTPSecretResponse struct {
	Secret string `msg:"secret" json:"secret,omitzero"`
}

type LoginResponse struct {
	Result       string              `msg:"result" json:"result,omitzero"`
	ID           string              `msg:"_id" json:"_id,omitzero"`
	UserID       string              `msg:"user_id" json:"user_id,omitzero"`
	Token        string              `msg:"token" json:"token,omitzero"`
	Name         string              `msg:"name" json:"name,omitzero"`
	Subscription WebpushSubscription `msg:"subscription" json:"subscription,omitzero"`
}

type Sessions struct {
	ID   string `msg:"_id" json:"_id,omitzero"`
	Name string `msg:"name" json:"name,omitzero"`
}
type Account struct {
	ID    string `msg:"_id" json:"_id,omitzero"`
	Email string `msg:"email" json:"email,omitzero"`
}

type Onboarding struct {
	Onboarding bool `msg:"onboarding" json:"onboarding,omitzero"`
}

type MFA struct {
	// Unvalidated or authorised MFA ticket; used to resolve the correct account
	MfaTicket string `msg:"mfa_ticket" json:"mfa_ticket,omitzero"`

	// MFA response
	MfaResponse MFAResponse `msg:"mfa_response" json:"mfa_response,omitzero"`

	// Friendly name used for the session
	FriendlyName string `msg:"friendly_name" json:"friendly_name,omitzero"`
}

type MFAResponse struct {
	Password string `msg:"password" json:"password,omitzero"`
}

type ChangeEmail struct {
	Ticket MFATicket `msg:"ticket" json:"ticket,omitzero"` // Why is this nested (seriously, look at AuthMFATicketResponse)
}

type MFATicket struct {
	ID           string `msg:"_id" json:"_id,omitzero"`
	AccountID    string `msg:"account_id" json:"account_id,omitzero"`
	Token        string `msg:"token" json:"token,omitzero"`
	Validated    bool   `msg:"validated" json:"validated,omitzero"`
	Authorised   bool   `msg:"authorised" json:"authorised,omitzero"`
	LastTOTPCode string `msg:"last_totp_code" json:"last_totp_code,omitzero"`
}
