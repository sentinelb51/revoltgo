package revoltgo

type Bot struct {
	ID string `json:"_id"`

	// User ID of the bot owner
	Owner string `json:"owner"`

	// Token used to authenticate requests for this bot
	Token string `json:"token"`

	// Whether the bot is public (may be invited by anyone)
	Public bool `json:"public"`

	// Whether to enable analytics
	Analytics bool `json:"analytics"`

	// Whether this bot should be publicly discoverable
	Discoverable bool `json:"discoverable"`

	// Reserved; URL for handling interactions
	InteractionsURL string `json:"interactions_url"`

	// URL for terms of service
	TermsOfServiceURL string `json:"terms_of_service_url"`

	// URL for privacy policy
	PrivacyPolicyURL string `json:"privacy_policy_url"`

	// Enum of bot flags
	Flags int `json:"flags"`
}

type PublicBot struct {
	ID          string      `json:"_id"`
	Username    string      `json:"username"`
	Avatar      *Attachment `json:"avatar"`
	Description string      `json:"description"`
}

type FetchedBot struct {
	Bot  *Bot  `json:"bot"`
	User *User `json:"user"`
}

type FetchedBots struct {
	Bots  []*Bot  `json:"bots"`
	Users []*User `json:"users"`
}
