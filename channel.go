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
	ID          string      `msg:"_id"`
	ChannelType ChannelType `msg:"channel_type"`

	Name        string      `msg:"name"`
	Description *string     `msg:"description"`
	Icon        *Attachment `msg:"icon"`
	NSFW        bool        `msg:"nsfw"`
	Active      bool        `msg:"active"`

	Server          *string                        `msg:"server"`           // Server channels only
	Voice           *ChannelVoiceInformation       `msg:"voice"`            // Server channels only
	RolePermissions map[string]PermissionOverwrite `msg:"role_permissions"` // Server channel only

	Recipients  []string `msg:"recipients"`  // DM or Group
	Permissions *int64   `msg:"permissions"` // Group only
	Owner       string   `msg:"owner"`       // Group or SavedMessages ("user" in SavedMessages)

	LastMessageID      *string              `msg:"last_message_id"`
	DefaultPermissions *PermissionOverwrite `msg:"default_permissions"`
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
	Name               *string                        `msg:"name,omitempty"`
	Owner              *string                        `msg:"owner,omitempty"`
	Description        *string                        `msg:"description,omitempty"`
	Icon               *Attachment                    `msg:"icon,omitempty"`
	NSFW               *bool                          `msg:"nsfw,omitempty"`
	Active             *bool                          `msg:"active,omitempty"`
	Permissions        *int64                         `msg:"permissions,omitempty"`
	RolePermissions    map[string]PermissionOverwrite `msg:"role_permissions,omitempty"`
	DefaultPermissions *PermissionOverwrite           `msg:"default_permissions,omitempty"`
	LastMessageID      *string                        `msg:"last_message_id,omitempty"`
	Voice              *ChannelVoiceInformation       `msg:"voice,omitempty"`
}

type ChannelFetchedMessages struct {
	Messages []*Message      `msg:"messages"`
	Users    []*User         `msg:"users"`
	Members  []*ServerMember `msg:"members"`
}

type ChannelJoinCall struct {
	// Token for authenticating with the voice server
	Token string `msg:"token"`

	// URL of the livekit server to connect to
	URL string `msg:"url"`
}

type ChannelVoiceInformation struct {
	MaxUsers *int `msg:"max_users"`
}
