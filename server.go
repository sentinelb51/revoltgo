package revoltgo

import (
	"time"
)

// Server holds information about a server.
type Server struct {
	ID             string                `json:"_id"`
	Owner          string                `json:"owner"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Channels       []string              `json:"channels"`
	Categories     []*ServerCategory     `json:"categories"`
	SystemMessages *ServerSystemMessages `json:"system_messages"`

	// Roles is a map of role ID to ServerRole structs.
	Roles map[string]*ServerRole `json:"roles"`

	DefaultPermissions uint        `json:"default_permissions"`
	Icon               *Attachment `json:"icon"`
	Banner             *Attachment `json:"banner"`
	Flags              uint        `json:"flags"`
	NSFW               bool        `json:"nsfw"`
	Analytics          bool        `json:"analytics"`
	Discoverable       bool        `json:"discoverable"`
}

type ServerRole struct {
	Name        string       `json:"name"`
	Permissions PermissionAD `json:"permissions"`
	Colour      string       `json:"colour"`
	Hoist       bool         `json:"hoist"`
	Rank        uint         `json:"rank"`
}

// ServerCategory Server categories struct.
type ServerCategory struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Channels []string `json:"channels"`
}

// ServerSystemMessages System messages struct.
type ServerSystemMessages struct {
	UserJoined string `json:"user_joined,omitempty"`
	UserLeft   string `json:"user_left,omitempty"`
	UserKicked string `json:"user_kicked,omitempty"`
	UserBanned string `json:"user_banned,omitempty"`
}

type ServerMember struct {
	ID       MemberCompoundID `json:"_id"`
	Nickname string           `json:"nickname"`
	Avatar   *Attachment      `json:"avatar"`
	Roles    []string         `json:"roles"`
	Timeout  time.Time        `json:"timeout"`
}

type MemberCompoundID struct {
	User   string `json:"user"`
	Server string `json:"server"`
}

type ServerMembers struct {
	Members []*ServerMember `json:"members"`
	Users   []*User         `json:"users"`
}

type ServerBans struct {
	Users []*User      `json:"users"`
	Bans  []*ServerBan `json:"bans"`
}

type ServerBan struct {
	ID     MemberCompoundID `json:"_id"`
	Reason string           `json:"reason"`
}
