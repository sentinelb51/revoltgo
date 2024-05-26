package revoltgo

type ChannelType string

const (
	ChannelTypeSavedMessages ChannelType = "SavedMessages"
	ChannelTypeText          ChannelType = "TextChannel"
	ChannelTypeVoice         ChannelType = "VoiceChannel"
	ChannelTypeDM            ChannelType = "DirectMessage"
	ChannelTypeGroup         ChannelType = "Group"
)

// Channel holds information about a channel.
type Channel struct {
	ID                 string        `json:"_id"`
	ChannelType        ChannelType   `json:"channel_type"`
	Server             string        `json:"server"`
	Name               string        `json:"name"`
	Description        string        `json:"description"`
	Icon               *Attachment   `json:"icon"`
	DefaultPermissions *PermissionAD `json:"default_permissions"`
	NSFW               bool          `json:"nsfw"`

	// ID of the last message sent in this channel
	LastMessageID string `json:"last_message_id"`

	// RolePermissions is a map of role ID to PermissionAD structs.
	RolePermissions map[string]*PermissionAD `json:"role_permissions"`

	/* Direct messages/groups only */

	// Permissions assigned to members of this group (does not apply to the owner of the group)
	Permissions *uint `json:"permissions"`

	// User ID of the owner of the group
	Owner string `json:"owner"`

	// [2-tuple of] user IDs participating in this channel
	Recipients []string `json:"recipients"`

	// Whether this direct message channel is currently open on both sides
	Active bool `json:"active"`
}

type ChannelFetchedMessages struct {
	Messages []*Message      `json:"messages"`
	Users    []*User         `json:"users"`
	Members  []*ServerMember `json:"members"`
}
