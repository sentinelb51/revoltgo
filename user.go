package revoltgo

type UserRelationsType string

const (
	UserRelationsTypeNone         = "None"
	UserRelationsTypeUser         = "User"
	UserRelationsTypeFriend       = "Friend"
	UserRelationsTypeOutgoing     = "Outgoing"
	UserRelationsTypeIncoming     = "Incoming"
	UserRelationsTypeBlocked      = "Blocked"
	UserRelationsTypeBlockedOther = "BlockedOther"
)

type User struct {
	ID            string         `json:"_id"`
	Username      string         `json:"username"`
	Discriminator string         `json:"discriminator"`
	DisplayName   string         `json:"display_name"`
	Avatar        *Attachment    `json:"avatar"`
	Relations     *UserRelations `json:"relations"`

	// Bitfield of user badges
	Badges *int `json:"badges"`

	// User's active status
	Status *UserStatus `json:"status"`

	// User's profile
	Profile *UserProfile `json:"profile"`

	// Enum of user flags
	Flags *int `json:"flags"`

	// Racism?!1
	Privileged *bool `json:"privileged"`

	// Bot information, if the user is a bot
	Bot *Bot `json:"bot"`

	Relationship string `json:"relationship"`

	// Whether this user is currently online
	Online *bool `json:"online"`
}

type UserProfile struct {
	Content    string      `json:"content"`
	Background *Attachment `json:"background"`
}

type UserRelations struct {
	ID     string            `json:"_id"`
	Status UserRelationsType `json:"status"`
}

type UserStatus struct {
	Text     string `json:"text"`
	Presence string `json:"presence"`
}

type BotInformation struct {
	Owner string `json:"owner"`
}

type UserMutual struct {
	Users   []string `json:"users"`
	Servers []string `json:"servers"`
}
