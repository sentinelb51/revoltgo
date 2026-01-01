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
	ID            string               `msg:"_id"`
	Username      string               `msg:"username"`
	Discriminator string               `msg:"discriminator"`
	Flags         uint32               `msg:"flags"`
	Privileged    bool                 `msg:"privileged"`
	Badges        uint32               `msg:"badges"`
	Online        bool                 `msg:"online"`
	Relations     []UserRelations      `msg:"relations"`
	Relationship  UserRelationshipType `msg:"relationship"`
	DisplayName   *string              `msg:"display_name"`
	Avatar        *Attachment          `msg:"avatar"`
	Status        *UserStatus          `msg:"status"`
	Profile       *UserProfile         `msg:"profile"` // todo: deprecated? not present in src
	Bot           *Bot                 `msg:"bot"`
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
	ID            *string               `msg:"_id"`
	Username      *string               `msg:"username"`
	Discriminator *string               `msg:"discriminator"`
	Flags         *uint32               `msg:"flags"`
	Privileged    *bool                 `msg:"privileged"`
	Badges        *uint32               `msg:"badges"`
	Online        *bool                 `msg:"online"`
	Relations     *[]UserRelations      `msg:"relations"`
	Relationship  *UserRelationshipType `msg:"relationship"`
	DisplayName   *string               `msg:"display_name"`
	Avatar        *Attachment           `msg:"avatar"`
	Status        *UserStatus           `msg:"status"`
	Profile       *UserProfile          `msg:"profile,omitempty"` // todo: deprecated? not present in src
	Bot           *Bot                  `msg:"bot"`
}

func (u *User) Mention() string {
	return fmt.Sprintf("<@%s>", u.ID)
}

type UserProfile struct {
	Content    string      `msg:"content,omitempty"`
	Background *Attachment `msg:"background,omitempty"`
}

type UserRelations struct {
	ID     string               `msg:"_id"`
	Status UserRelationshipType `msg:"status"`
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
	Text     string             `msg:"text,omitempty"`
	Presence UserStatusPresence `msg:"presence"`
}

type BotInformation struct {
	Owner string `msg:"owner"`
}

type MutualFriendsAndServersResponse struct {
	Users   []string `msg:"users"`
	Servers []string `msg:"servers"`
}

// UserSettings TODO: This does not get decoded due to API sending tuples for some god-forsaken reason
type UserSettings struct {
	Updated int
	Data    msgp.Raw
}
