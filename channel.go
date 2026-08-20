package revoltgo

import "log"

//go:generate msgp -tests=false -io=false

type ChannelType string

const (
	ChannelTypeSavedMessages ChannelType = "SavedMessages"
	ChannelTypeText          ChannelType = "TextChannel"
	ChannelTypeDM            ChannelType = "DirectMessage"
	ChannelTypeGroup         ChannelType = "Group"
)

// Channel is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/channels.rs#L13
type Channel struct {
	ID          string      `msg:"_id" json:"_id,omitzero"`
	ChannelType ChannelType `msg:"channel_type" json:"channel_type,omitzero"`

	Name        string  `msg:"name" json:"name,omitzero"`
	Description *string `msg:"description" json:"description,omitzero"`
	Icon        *File   `msg:"icon" json:"icon,omitzero"`
	NSFW        bool    `msg:"nsfw" json:"nsfw,omitzero"`
	Active      bool    `msg:"active" json:"active,omitzero"`

	Server          *string                        `msg:"server" json:"server,omitzero"` // Server channels only
	Voice           *ChannelVoiceInformation       `msg:"voice" json:"voice,omitzero"`   // Server channels only
	Slowmode        *int                           `msg:"slowmode" json:"slowmode,omitzero"`
	RolePermissions map[string]PermissionOverwrite `msg:"role_permissions" json:"role_permissions,omitzero"` // Server channel only

	Recipients  []string `msg:"recipients" json:"recipients,omitzero"`   // DM or Group
	Permissions *int64   `msg:"permissions" json:"permissions,omitzero"` // Group only
	Owner       string   `msg:"owner" json:"owner,omitzero"`             // Group or SavedMessages ("user" in SavedMessages)

	LastMessageID      *string              `msg:"last_message_id" json:"last_message_id,omitzero"`
	DefaultPermissions *PermissionOverwrite `msg:"default_permissions" json:"default_permissions,omitzero"`
}

func (c *Channel) update(data PartialChannel) {
	if data.Name != nil {
		c.Name = *data.Name
	}

	if data.Owner != nil {
		c.Owner = *data.Owner
	}

	if data.Description != nil {
		c.Description = data.Description
	}

	if data.Icon != nil {
		c.Icon = data.Icon
	}

	if data.NSFW != nil {
		c.NSFW = *data.NSFW
	}

	if data.Active != nil {
		c.Active = *data.Active
	}

	if data.Permissions != nil {
		c.Permissions = data.Permissions
	}

	if data.RolePermissions != nil {
		c.RolePermissions = data.RolePermissions
	}

	if data.DefaultPermissions != nil {
		c.DefaultPermissions = data.DefaultPermissions
	}

	if data.LastMessageID != nil {
		c.LastMessageID = data.LastMessageID
	}

	if data.Voice != nil {
		c.Voice = data.Voice
	}

	if data.Slowmode != nil {
		c.Slowmode = data.Slowmode
	}
}

func (c *Channel) clear(fields []string) {
	for _, field := range fields {
		switch field {
		case "Icon":
			c.Icon = nil
		case "Description":
			c.Description = nil
		case "DefaultPermissions":
			c.DefaultPermissions = nil
		case "Voice":
			c.Voice = nil
		case "Slowmode":
			c.Slowmode = nil
		default:
			log.Printf("Channel.clear(): unknown field %s", field)
		}
	}
}

type PartialChannel struct {
	Name        *string `msg:"name" json:"name,omitzero"`
	Owner       *string `msg:"owner" json:"owner,omitzero"`
	Description *string `msg:"description" json:"description,omitzero"`
	Icon        *File   `msg:"icon" json:"icon,omitzero"`
	NSFW        *bool   `msg:"nsfw" json:"nsfw,omitzero"`

	// Whether the channel is listed in direct messages. False means hidden; the DM was closed.
	Active             *bool                          `msg:"active" json:"active,omitzero"`
	Permissions        *int64                         `msg:"permissions" json:"permissions,omitzero"`
	RolePermissions    map[string]PermissionOverwrite `msg:"role_permissions" json:"role_permissions,omitzero"`
	DefaultPermissions *PermissionOverwrite           `msg:"default_permissions" json:"default_permissions,omitzero"`
	LastMessageID      *string                        `msg:"last_message_id" json:"last_message_id,omitzero"`
	Voice              *ChannelVoiceInformation       `msg:"voice" json:"voice,omitzero"`
	Slowmode           *int                           `msg:"slowmode" json:"slowmode,omitzero"`
}

type CompositeChannelID struct {
	Channel string `msg:"channel" json:"channel,omitzero"`
	User    string `msg:"user" json:"user,omitzero"`
}

type ChannelFetchedMessages struct {
	Messages []*Message      `msg:"messages" json:"messages,omitzero"`
	Users    []*User         `msg:"users" json:"users,omitzero"`
	Members  []*ServerMember `msg:"members" json:"members,omitzero"`
}

type ChannelJoinCall struct {
	// Token for authenticating with the voice server
	Token string `msg:"token" json:"token,omitzero"`

	// URL of the livekit server to connect to
	URL string `msg:"url" json:"url,omitzero"`
}

type ChannelVoiceInformation struct {
	MaxUsers *int `msg:"max_users" json:"max_users,omitzero"`
}

type ChannelVoiceState struct {
	ID           string            `msg:"id" json:"id,omitzero"`
	Participants []*UserVoiceState `msg:"participants" json:"participants,omitzero"`
}

type ChannelUnreadCompositeID struct {
	Channel string `msg:"channel" json:"channel"`
	User    string `msg:"user" json:"user"`
}

type ChannelUnread struct {
	ID            ChannelUnreadCompositeID `msg:"_id" json:"_id"`
	LastMessageID *string                  `msg:"last_id" json:"last_id,omitzero"`
	MentionIDs    []string                 `msg:"mentions" json:"mentions,omitzero"`
}

type ChannelMessages struct {
	Messages []*Message      `msg:"messages" json:"messages,omitzero"`
	Users    []*User         `msg:"users" json:"users,omitzero"`
	Members  []*ServerMember `msg:"members" json:"members,omitzero"`
}
