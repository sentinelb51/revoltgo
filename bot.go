package revoltgo

//go:generate msgp -tests=false -io=false

type Bot struct {
	ID string `msg:"_id"`

	// User ID of the bot owner
	Owner string `msg:"owner"`

	// Token used to authenticate requests for this bot
	Token string `msg:"token"`

	// Whether the bot is public (may be invited by anyone)
	Public bool `msg:"public"`

	// Whether to enable analytics
	Analytics bool `msg:"analytics"`

	// Whether this bot should be publicly discoverable
	Discoverable bool `msg:"discoverable"`

	// Reserved; URL for handling interactions
	InteractionsURL string `msg:"interactions_url"`

	// URL for terms of service
	TermsOfServiceURL string `msg:"terms_of_service_url"`

	// URL for privacy policy
	PrivacyPolicyURL string `msg:"privacy_policy_url"`

	// Enum of bot flags
	Flags int `msg:"flags"`
}

type PublicBot struct {
	ID          string      `msg:"_id"`
	Username    string      `msg:"username"`
	Avatar      *Attachment `msg:"avatar"`
	Description string      `msg:"description"`
}

type FetchedBot struct {
	Bot  *Bot  `msg:"bot"`
	User *User `msg:"user"`
}

type FetchedBots struct {
	Bots  []*Bot  `msg:"bots"`
	Users []*User `msg:"users"`
}
