package revoltgo

//go:generate msgp -tests=false -io=false

type Bot struct {
	ID string `msg:"_id" json:"_id,omitempty"`

	// User ID of the bot owner
	Owner string `msg:"owner" json:"owner,omitempty"`

	// Token used to authenticate requests for this bot
	Token string `msg:"token" json:"token,omitempty"`

	// Whether the bot is public (may be invited by anyone)
	Public bool `msg:"public" json:"public,omitempty"`

	// Whether to enable analytics
	Analytics bool `msg:"analytics" json:"analytics,omitempty"`

	// Whether this bot should be publicly discoverable
	Discoverable bool `msg:"discoverable" json:"discoverable,omitempty"`

	// Reserved; URL for handling interactions
	InteractionsURL string `msg:"interactions_url" json:"interactions_url,omitempty"`

	// URL for terms of service
	TermsOfServiceURL string `msg:"terms_of_service_url" json:"terms_of_service_url,omitempty"`

	// URL for privacy policy
	PrivacyPolicyURL string `msg:"privacy_policy_url" json:"privacy_policy_url,omitempty"`

	// Enum of bot flags
	Flags int `msg:"flags" json:"flags,omitempty"`
}

type PublicBot struct {
	ID          string      `msg:"_id" json:"_id,omitempty"`
	Username    string      `msg:"username" json:"username,omitempty"`
	Avatar      *Attachment `msg:"avatar" json:"avatar,omitempty"`
	Description string      `msg:"description" json:"description,omitempty"`
}

type FetchedBot struct {
	Bot  *Bot  `msg:"bot" json:"bot,omitempty"`
	User *User `msg:"user" json:"user,omitempty"`
}

type FetchedBots struct {
	Bots  []*Bot  `msg:"bots" json:"bots,omitempty"`
	Users []*User `msg:"users" json:"users,omitempty"`
}
