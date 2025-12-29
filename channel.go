package revoltgo

import "log"

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
	ID          string      `json:"_id"`
	ChannelType ChannelType `json:"channel_type"`

	Name        string      `json:"name"`
	Description *string     `json:"description"`
	Icon        *Attachment `json:"icon"`
	NSFW        bool        `json:"nsfw"`
	Active      bool        `json:"active"`

	Server          *string                        `json:"server"`           // Server channels only
	Voice           *ChannelVoiceInformation       `json:"voice"`            // Server channels only
	RolePermissions map[string]PermissionOverwrite `json:"role_permissions"` // Server channel only

	Recipients  []string `json:"recipients"`  // DM or Group
	Permissions *int64   `json:"permissions"` // Group only
	Owner       string   `json:"owner"`       // Group or SavedMessages ("user" in SavedMessages)

	LastMessageID      *string              `json:"last_message_id"`
	DefaultPermissions *PermissionOverwrite `json:"default_permissions"`
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
		// Replaces the entire map if provided
		c.RolePermissions = *data.RolePermissions
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
	Name               *string                         `json:"name,omitempty"`
	Owner              *string                         `json:"owner,omitempty"`
	Description        *string                         `json:"description,omitempty"`
	Icon               *Attachment                     `json:"icon,omitempty"`
	NSFW               *bool                           `json:"nsfw,omitempty"`
	Active             *bool                           `json:"active,omitempty"`
	Permissions        *int64                          `json:"permissions,omitempty"`
	RolePermissions    *map[string]PermissionOverwrite `json:"role_permissions,omitempty"`
	DefaultPermissions *PermissionOverwrite            `json:"default_permissions,omitempty"`
	LastMessageID      *string                         `json:"last_message_id,omitempty"`
	Voice              *ChannelVoiceInformation        `json:"voice,omitempty"`
}

type ChannelFetchedMessages struct {
	Messages []*Message      `json:"messages"`
	Users    []*User         `json:"users"`
	Members  []*ServerMember `json:"members"`
}

type ChannelJoinCall struct {
	// Token for authenticating with the voice server
	Token string `json:"token"`

	// URL of the livekit server to connect to
	URL string `json:"url"`
}

type ChannelVoiceInformation struct {
	MaxUsers *int `json:"max_users"`
}
