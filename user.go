package revoltgo

import "github.com/goccy/go-json"

type UserRelationshipType string

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
	ID            string           `json:"_id"`
	Username      string           `json:"username"`
	Discriminator string           `json:"discriminator"`
	DisplayName   string           `json:"display_name"`
	Avatar        *Attachment      `json:"avatar"`
	Relations     []*UserRelations `json:"relations"`

	// Bitfield of user badges
	Badges int `json:"badges"`

	// User's active status
	Status *UserStatus `json:"status"`

	// User's profile
	Profile *UserProfile `json:"profile"`

	// Enum of user flags
	Flags *int `json:"flags"`

	// Racism?!1
	Privileged bool `json:"privileged"`

	// Bot information, if the user is a bot
	Bot *Bot `json:"bot"`

	// Your relationship to this user
	Relationship UserRelationshipType `json:"relationship"`

	// Whether this user is currently online
	Online bool `json:"online"`
}

type UserProfile struct {
	Content    string      `json:"content,omitempty"`
	Background *Attachment `json:"background,omitempty"`
}

type UserRelations struct {
	ID     string               `json:"_id"`
	Status UserRelationshipType `json:"status"`
}

type UserStatusPresence string

const (
	UserStatusPresenceOnline    UserStatusPresence = "Online"
	UserStatusPresenceIdle      UserStatusPresence = "Idle"
	UserStatusPresenceFocus     UserStatusPresence = "Focus"
	UserStatusPresenceBusy      UserStatusPresence = "Busy"
	UserStatusPresenceInvisible UserStatusPresence = "Invisible"
)

type UserStatus struct {
	Text     string             `json:"text,omitempty"`
	Presence UserStatusPresence `json:"presence"`
}

type BotInformation struct {
	Owner string `json:"owner"`
}

type MutualFriendsAndServersResponse struct {
	Users   []string `json:"users"`
	Servers []string `json:"servers"`
}

// UserSettings TODO: This does not get decoded due to API sending tuples for some god-forsaken reason
type UserSettings struct {
	Updated int
	Data    json.RawMessage
}
