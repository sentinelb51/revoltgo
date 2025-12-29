package revoltgo

import (
	"fmt"

	"github.com/goccy/go-json"
)

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
	ID            string               `json:"_id"`
	Username      string               `json:"username"`
	Discriminator string               `json:"discriminator"`
	Flags         uint32               `json:"flags"`
	Privileged    bool                 `json:"privileged"`
	Badges        uint32               `json:"badges"`
	Online        bool                 `json:"online"`
	Relations     []UserRelations      `json:"relations"`
	Relationship  UserRelationshipType `json:"relationship"`
	DisplayName   *string              `json:"display_name"`
	Avatar        *Attachment          `json:"avatar"`
	Status        *UserStatus          `json:"status"`
	Profile       *UserProfile         `json:"profile"` // todo: deprecated? not present in src
	Bot           *Bot                 `json:"bot"`
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
	ID            *string               `json:"_id"`
	Username      *string               `json:"username"`
	Discriminator *string               `json:"discriminator"`
	Flags         *uint32               `json:"flags"`
	Privileged    *bool                 `json:"privileged"`
	Badges        *uint32               `json:"badges"`
	Online        *bool                 `json:"online"`
	Relations     *[]UserRelations      `json:"relations"`
	Relationship  *UserRelationshipType `json:"relationship"`
	DisplayName   *string               `json:"display_name"`
	Avatar        *Attachment           `json:"avatar"`
	Status        *UserStatus           `json:"status"`
	Profile       *UserProfile          `json:"profile,omitempty"` // todo: deprecated? not present in src
	Bot           *Bot                  `json:"bot"`
}

func (u *User) Mention() string {
	return fmt.Sprintf("<@%s>", u.ID)
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
