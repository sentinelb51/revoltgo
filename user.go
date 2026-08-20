package revoltgo

import (
	"fmt"
	"log"
	"time"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false

type UserRelationshipType string

const (
	UserRelationshipTypeNone         = "None"
	UserRelationshipTypeUser         = "User"
	UserRelationshipTypeFriend       = "Friend"
	UserRelationshipTypeOutgoing     = "Outgoing"
	UserRelationshipTypeIncoming     = "Incoming"
	UserRelationshipTypeBlocked      = "Blocked"
	UserRelationshipTypeBlockedOther = "BlockedOther"
)

// User is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/users.rs#L24
type User struct {
	ID            string               `msg:"_id" json:"_id,omitzero"`
	Username      string               `msg:"username" json:"username,omitzero"`
	Discriminator string               `msg:"discriminator" json:"discriminator,omitzero"`
	Flags         uint32               `msg:"flags" json:"flags,omitzero"`
	Privileged    bool                 `msg:"privileged" json:"privileged,omitzero"`
	Badges        uint32               `msg:"badges" json:"badges,omitzero"`
	Online        bool                 `msg:"online" json:"online,omitzero"`
	Relations     []UserRelationship   `msg:"relations" json:"relations,omitzero"`
	Relationship  UserRelationshipType `msg:"relationship" json:"relationship,omitzero"`
	DisplayName   *string              `msg:"display_name" json:"display_name,omitzero"`
	Avatar        *File                `msg:"avatar" json:"avatar,omitzero"`
	Status        *UserStatus          `msg:"status" json:"status,omitzero"`
	Profile       *UserProfile         `msg:"profile" json:"profile,omitzero"` // todo: deprecated? not present in src
	Bot           *Bot                 `msg:"bot" json:"bot,omitzero"`
}

func (u *User) AvatarURL(size string) string {
	if u.Avatar == nil {
		return apiURL + EndpointUserDefaultAvatar(u.ID)
	}

	return u.Avatar.URL(size)
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
			log.Printf("User.Clear(): unknown field %s\n", field)
		}
	}
}

type PartialUser struct {
	ID            *string               `msg:"_id" json:"_id,omitzero"`
	Username      *string               `msg:"username" json:"username,omitzero"`
	Discriminator *string               `msg:"discriminator" json:"discriminator,omitzero"`
	Flags         *uint32               `msg:"flags" json:"flags,omitzero"`
	Privileged    *bool                 `msg:"privileged" json:"privileged,omitzero"`
	Badges        *uint32               `msg:"badges" json:"badges,omitzero"`
	Online        *bool                 `msg:"online" json:"online,omitzero"`
	Relations     *[]UserRelationship   `msg:"relations" json:"relations,omitzero"`
	Relationship  *UserRelationshipType `msg:"relationship" json:"relationship,omitzero"`
	DisplayName   *string               `msg:"display_name" json:"display_name,omitzero"`
	Avatar        *File                 `msg:"avatar" json:"avatar,omitzero"`
	Status        *UserStatus           `msg:"status" json:"status,omitzero"`
	Profile       *UserProfile          `msg:"profile" json:"profile,omitzero"` // todo: deprecated? not present in src
	Bot           *Bot                  `msg:"bot" json:"bot,omitzero"`
}

func (u *User) Mention() string {
	return fmt.Sprintf("<@%s>", u.ID)
}

type UserProfile struct {
	Content    string `msg:"content" json:"content,omitzero"`
	Background *File  `msg:"background" json:"background,omitzero"`
}

type UserRelationship struct {
	ID     string               `msg:"_id" json:"_id,omitzero"`
	Status UserRelationshipType `msg:"status" json:"status,omitzero"`
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
	Text     string             `msg:"text" json:"text,omitzero"`
	Presence UserStatusPresence `msg:"presence" json:"presence,omitzero"`
}

type BotInformation struct {
	Owner string `msg:"owner" json:"owner,omitzero"`
}

type MutualFriendsAndServersResponse struct {
	Users    []string `msg:"users" json:"users,omitzero"`
	Servers  []string `msg:"servers" json:"servers,omitzero"`
	Channels []string `msg:"channels" json:"channels,omitzero"`
}

// UserSettings TODO: This does not get decoded due to API sending tuples for some god-forsaken reason
type UserSettings struct {
	Updated int
	Data    msgp.Raw
}

// UserVoiceState is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/users.rs#L292
type UserVoiceState struct {
	ID            string    `msg:"id" json:"id,omitzero"`
	JoinedAt      time.Time `msg:"joined_at" json:"joined_at,omitzero"`
	IsReceiving   bool      `msg:"is_receiving" json:"is_receiving,omitzero"`
	IsPublishing  bool      `msg:"is_publishing" json:"is_publishing,omitzero"`
	Screensharing bool      `msg:"screensharing" json:"screensharing,omitzero"`
	Camera        bool      `msg:"camera" json:"camera,omitzero"`
}

// update applies a partial voice state; ID is skipped, as the voice cache keys on it
func (v *UserVoiceState) update(data PartialUserVoiceState) {
	if data.JoinedAt != nil {
		v.JoinedAt = *data.JoinedAt
	}

	if data.IsReceiving != nil {
		v.IsReceiving = *data.IsReceiving
	}

	if data.IsPublishing != nil {
		v.IsPublishing = *data.IsPublishing
	}

	if data.Screensharing != nil {
		v.Screensharing = *data.Screensharing
	}

	if data.Camera != nil {
		v.Camera = *data.Camera
	}
}

type PartialUserVoiceState struct {
	ID            *string    `msg:"id" json:"id,omitzero"`
	JoinedAt      *time.Time `msg:"joined_at" json:"joined_at,omitzero"`
	IsReceiving   *bool      `msg:"is_receiving" json:"is_receiving,omitzero"`
	IsPublishing  *bool      `msg:"is_publishing" json:"is_publishing,omitzero"`
	Screensharing *bool      `msg:"screensharing" json:"screensharing,omitzero"`
	Camera        *bool      `msg:"camera" json:"camera,omitzero"`
}
