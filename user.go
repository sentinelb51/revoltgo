package revoltgo

import (
	"fmt"
	"time"
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
	Pronouns      *string              `msg:"pronouns" json:"pronouns,omitzero"`
	Avatar        *File                `msg:"avatar" json:"avatar,omitzero"`
	Status        *UserStatus          `msg:"status" json:"status,omitzero"`
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

	if data.Pronouns != nil {
		u.Pronouns = data.Pronouns
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

type UserRemoveField string

const (
	UserRemoveAvatar            UserRemoveField = "Avatar"
	UserRemoveDisplayName       UserRemoveField = "DisplayName"
	UserRemovePronouns          UserRemoveField = "Pronouns"
	UserRemoveStatusText        UserRemoveField = "StatusText"
	UserRemoveStatusPresence    UserRemoveField = "StatusPresence"
	UserRemoveProfileContent    UserRemoveField = "ProfileContent"
	UserRemoveProfileBackground UserRemoveField = "ProfileBackground"
	UserRemoveInternal          UserRemoveField = "Internal"
)

// clear covers every UserRemoveField. ProfileContent and ProfileBackground
// are among them and are no-ops here: a profile is not part of the user record,
// only of the response to Session.UserProfile, so there is nothing cached to
// clear. They are still listed, or the default arm would log them as unknown.
func (u *User) clear(fields []UserRemoveField) {
	for _, field := range fields {
		switch field {
		case UserRemoveProfileContent, UserRemoveProfileBackground, UserRemoveInternal:
		case UserRemoveStatusText:
			if u.Status != nil {
				u.Status.Text = ""
			}
		case UserRemoveStatusPresence:
			if u.Status != nil {
				u.Status.Presence = ""
			}
		case UserRemoveAvatar:
			u.Avatar = nil
		case UserRemoveDisplayName:
			u.DisplayName = nil
		case UserRemovePronouns:
			u.Pronouns = nil
		default:
			logf("User.clear(): unknown field %s", field)
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
	Pronouns      *string               `msg:"pronouns" json:"pronouns,omitzero"`
	Avatar        *File                 `msg:"avatar" json:"avatar,omitzero"`
	Status        *UserStatus           `msg:"status" json:"status,omitzero"`
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
