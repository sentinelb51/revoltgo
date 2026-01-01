package revoltgo

import "log"

//go:generate msgp -tests=false -io=false

type ChannelType string

const (
	ChannelTypeSavedMessages ChannelType = "SavedMessages"
	ChannelTypeText          ChannelType = "TextChannel"
	ChannelTypeVoice         ChannelType = "VoiceChannel"
	ChannelTypeDM            ChannelType = "DirectMessage"
	ChannelTypeGroup         ChannelType = "Group"
)

// Channel is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/channels.rs#L13
type Channel struct {
	ID          string      `msg:"_id" json:"_id,omitempty"`
	ChannelType ChannelType `msg:"channel_type" json:"channel_type,omitempty"`

	Name        string      `msg:"name" json:"name,omitempty"`
	Description *string     `msg:"description" json:"description,omitempty"`
	Icon        *Attachment `msg:"icon" json:"icon,omitempty"`
	NSFW        bool        `msg:"nsfw" json:"nsfw,omitempty"`
	Active      bool        `msg:"active" json:"active,omitempty"`

	Server          *string                        `msg:"server" json:"server,omitempty"`                     // Server channels only
	Voice           *ChannelVoiceInformation       `msg:"voice" json:"voice,omitempty"`                       // Server channels only
	RolePermissions map[string]PermissionOverwrite `msg:"role_permissions" json:"role_permissions,omitempty"` // Server channel only

	Recipients  []string `msg:"recipients" json:"recipients,omitempty"`   // DM or Group
	Permissions *int64   `msg:"permissions" json:"permissions,omitempty"` // Group only
	Owner       string   `msg:"owner" json:"owner,omitempty"`             // Group or SavedMessages ("user" in SavedMessages)

	LastMessageID      *string              `msg:"last_message_id" json:"last_message_id,omitempty"`
	DefaultPermissions *PermissionOverwrite `msg:"default_permissions" json:"default_permissions,omitempty"`
}

func (c *Channel) update(data PartialChannel) {
	if data.Name != nil {
		c.Name = *data.Name
	}

	if data.Owner != nil {
		c.Owner = *data.Owner
	}

	if data.Description != nil {
		// Description in Main is *string, so we copy the pointer (or value)
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
}

func (c *Channel) clear(fields []string) {
	for _, field := range fields {
		switch field {
		case "Icon":
			c.Icon = nil
		case "Description":
			c.Description = nil
		default:
			log.Printf("Channel.clear(): unknown field %s", field)
		}
	}
}

type PartialChannel struct {
	Name               *string                        `msg:"name" json:"name,omitempty"`
	Owner              *string                        `msg:"owner" json:"owner,omitempty"`
	Description        *string                        `msg:"description" json:"description,omitempty"`
	Icon               *Attachment                    `msg:"icon" json:"icon,omitempty"`
	NSFW               *bool                          `msg:"nsfw" json:"nsfw,omitempty"`
	Active             *bool                          `msg:"active" json:"active,omitempty"`
	Permissions        *int64                         `msg:"permissions" json:"permissions,omitempty"`
	RolePermissions    map[string]PermissionOverwrite `msg:"role_permissions" json:"role_permissions,omitempty"`
	DefaultPermissions *PermissionOverwrite           `msg:"default_permissions" json:"default_permissions,omitempty"`
	LastMessageID      *string                        `msg:"last_message_id" json:"last_message_id,omitempty"`
	Voice              *ChannelVoiceInformation       `msg:"voice" json:"voice,omitempty"`
}

type ChannelFetchedMessages struct {
	Messages []*Message      `msg:"messages" json:"messages,omitempty"`
	Users    []*User         `msg:"users" json:"users,omitempty"`
	Members  []*ServerMember `msg:"members" json:"members,omitempty"`
}

type ChannelJoinCall struct {
	// Token for authenticating with the voice server
	Token string `msg:"token" json:"token,omitempty"`

	// URL of the livekit server to connect to
	URL string `msg:"url" json:"url,omitempty"`
}

type ChannelVoiceInformation struct {
	MaxUsers *int `msg:"max_users" json:"max_users,omitempty"`
}
