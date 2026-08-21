package revoltgo

//go:generate msgp -tests=false -io=false

type BotRemoveField string

const (
	BotRemoveToken           BotRemoveField = "Token"
	BotRemoveInteractionsURL BotRemoveField = "InteractionsURL"
)

type Bot struct {
	ID string `msg:"_id" json:"_id,omitzero"`

	// User ID of the bot owner
	Owner string `msg:"owner" json:"owner,omitzero"`

	// Token used to authenticate requests for this bot
	Token string `msg:"token" json:"token,omitzero"`

	// Whether the bot is public (may be invited by anyone)
	Public bool `msg:"public" json:"public,omitzero"`

	// Whether to enable analytics
	Analytics bool `msg:"analytics" json:"analytics,omitzero"`

	// Whether this bot should be publicly discoverable
	Discoverable bool `msg:"discoverable" json:"discoverable,omitzero"`

	// Reserved; URL for handling interactions
	InteractionsURL string `msg:"interactions_url" json:"interactions_url,omitzero"`

	// URL for terms of service
	TermsOfServiceURL string `msg:"terms_of_service_url" json:"terms_of_service_url,omitzero"`

	// URL for privacy policy
	PrivacyPolicyURL string `msg:"privacy_policy_url" json:"privacy_policy_url,omitzero"`

	// Enum of bot flags
	Flags uint32 `msg:"flags" json:"flags,omitzero"`
}

type PublicBot struct {
	ID          string `msg:"_id" json:"_id,omitzero"`
	Username    string `msg:"username" json:"username,omitzero"`
	Avatar      *File  `msg:"avatar" json:"avatar,omitzero"`
	Description string `msg:"description" json:"description,omitzero"`
}

type FetchedBot struct {
	Bot  *Bot  `msg:"bot" json:"bot,omitzero"`
	User *User `msg:"user" json:"user,omitzero"`
}

type FetchedBots struct {
	Bots  []*Bot  `msg:"bots" json:"bots,omitzero"`
	Users []*User `msg:"users" json:"users,omitzero"`
}
