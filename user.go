package revoltgo

import (
	"fmt"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false

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

// User is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/users.rs#L24
type User struct {
	ID            string               `msg:"_id" json:"_id,omitempty"`
	Username      string               `msg:"username" json:"username,omitempty"`
	Discriminator string               `msg:"discriminator" json:"discriminator,omitempty"`
	Flags         uint32               `msg:"flags" json:"flags,omitempty"`
	Privileged    bool                 `msg:"privileged" json:"privileged,omitempty"`
	Badges        uint32               `msg:"badges" json:"badges,omitempty"`
	Online        bool                 `msg:"online" json:"online,omitempty"`
	Relations     []UserRelations      `msg:"relations" json:"relations,omitempty"`
	Relationship  UserRelationshipType `msg:"relationship" json:"relationship,omitempty"`
	DisplayName   *string              `msg:"display_name" json:"display_name,omitempty"`
	Avatar        *Attachment          `msg:"avatar" json:"avatar,omitempty"`
	Status        *UserStatus          `msg:"status" json:"status,omitempty"`
	Profile       *UserProfile         `msg:"profile" json:"profile,omitempty"` // todo: deprecated? not present in src
	Bot           *Bot                 `msg:"bot" json:"bot,omitempty"`
}

func (u *User) update(data PartialUser) {
	if data.Username != nil {
		u.Username = *data.Username
	}

	if data.Discriminator != nil {
		u.Discriminator = *data.Discriminator
	}

	if data.DisplayName != nil {
		u.DisplayName = data.DisplayName
	}

	if data.Avatar != nil {
		u.Avatar = data.Avatar
	}

	if data.Relations != nil {
		u.Relations = *data.Relations
	}

	if data.Badges != nil {
		u.Badges = *data.Badges
	}

	if data.Status != nil {
		u.Status = data.Status
	}

	if data.Flags != nil {
		u.Flags = *data.Flags
	}

	if data.Privileged != nil {
		u.Privileged = *data.Privileged
	}

	if data.Bot != nil {
		u.Bot = data.Bot
	}

	if data.Relationship != nil {
		u.Relationship = *data.Relationship
	}

	if data.Online != nil {
		u.Online = *data.Online
	}
}

func (u *User) clear(fields []string) {
	for _, field := range fields {
		switch field {
		case "ProfileContent":
			if u.Profile != nil {
				u.Profile.Content = ""
			}
		case "ProfileBackground":
			if u.Profile != nil {
				u.Profile.Background = nil
			}
		case "StatusText":
			if u.Status != nil {
				u.Status.Text = ""
			}
		case "Avatar":
			u.Avatar = nil
		case "DisplayName":
			u.DisplayName = nil
		default:
			fmt.Printf("User.Clear(): unknown field %s\n", field)
		}
	}
}

type PartialUser struct {
	ID            *string               `msg:"_id" json:"_id,omitempty"`
	Username      *string               `msg:"username" json:"username,omitempty"`
	Discriminator *string               `msg:"discriminator" json:"discriminator,omitempty"`
	Flags         *uint32               `msg:"flags" json:"flags,omitempty"`
	Privileged    *bool                 `msg:"privileged" json:"privileged,omitempty"`
	Badges        *uint32               `msg:"badges" json:"badges,omitempty"`
	Online        *bool                 `msg:"online" json:"online,omitempty"`
	Relations     *[]UserRelations      `msg:"relations" json:"relations,omitempty"`
	Relationship  *UserRelationshipType `msg:"relationship" json:"relationship,omitempty"`
	DisplayName   *string               `msg:"display_name" json:"display_name,omitempty"`
	Avatar        *Attachment           `msg:"avatar" json:"avatar,omitempty"`
	Status        *UserStatus           `msg:"status" json:"status,omitempty"`
	Profile       *UserProfile          `msg:"profile" json:"profile,omitempty"` // todo: deprecated? not present in src
	Bot           *Bot                  `msg:"bot" json:"bot,omitempty"`
}

func (u *User) Mention() string {
	return fmt.Sprintf("<@%s>", u.ID)
}

type UserProfile struct {
	Content    string      `msg:"content" json:"content,omitempty"`
	Background *Attachment `msg:"background" json:"background,omitempty"`
}

type UserRelations struct {
	ID     string               `msg:"_id" json:"_id,omitempty"`
	Status UserRelationshipType `msg:"status" json:"status,omitempty"`
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
	Text     string             `msg:"text" json:"text,omitempty"`
	Presence UserStatusPresence `msg:"presence" json:"presence,omitempty"`
}

type BotInformation struct {
	Owner string `msg:"owner" json:"owner,omitempty"`
}

type MutualFriendsAndServersResponse struct {
	Users   []string `msg:"users" json:"users,omitempty"`
	Servers []string `msg:"servers" json:"servers,omitempty"`
}

// UserSettings TODO: This does not get decoded due to API sending tuples for some god-forsaken reason
type UserSettings struct {
	Updated int
	Data    msgp.Raw
}
