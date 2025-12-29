package revoltgo

import (
	"time"
)

type (
	MessageSystemType       string
	MessageEmbedSpecialType string
	MessageFlagsType        uint32
)

const (
	MessageSystemText                      MessageSystemType = "text"
	MessageSystemUserAdded                 MessageSystemType = "user_added"
	MessageSystemUserRemove                MessageSystemType = "user_remove"
	MessageSystemUserJoined                MessageSystemType = "user_joined"
	MessageSystemUserLeft                  MessageSystemType = "user_left"
	MessageSystemUserKicked                MessageSystemType = "user_kicked"
	MessageSystemUserBanned                MessageSystemType = "user_banned"
	MessageSystemChannelRenamed            MessageSystemType = "channel_renamed"
	MessageSystemChannelDescriptionChanged MessageSystemType = "channel_description_changed"
	MessageSystemChannelIconChanged        MessageSystemType = "channel_icon_changed"
	MessageSystemChannelOwnershipChanged   MessageSystemType = "channel_ownership_changed"
)

const (
	MessageEmbedSpecialNone       MessageEmbedSpecialType = "None"
	MessageEmbedSpecialGIF        MessageEmbedSpecialType = "GIF"
	MessageEmbedSpecialYouTube    MessageEmbedSpecialType = "YouTube"
	MessageEmbedSpecialLightspeed MessageEmbedSpecialType = "Lightspeed"
	MessageEmbedSpecialTwitch     MessageEmbedSpecialType = "Twitch"
	MessageEmbedSpecialSpotify    MessageEmbedSpecialType = "Spotify"
	MessageEmbedSpecialSoundcloud MessageEmbedSpecialType = "Soundcloud"
	MessageEmbedSpecialBandcamp   MessageEmbedSpecialType = "Bandcamp"
	MessageEmbedSpecialStreamable MessageEmbedSpecialType = "Streamable"
)

const (
	// MessageFlagsSuppressNotifications  will not send push / desktop notifications
	MessageFlagsSuppressNotifications MessageFlagsType = 1
	// MessageFlagsMentionsEveryone will mention all users who can see the channel
	MessageFlagsMentionsEveryone MessageFlagsType = 2
	// MessageFlagsMentionsOnline will mention all users who are online and can see the channel.
	// This cannot be true if MentionsEveryone is true
	MessageFlagsMentionsOnline MessageFlagsType = 3
)

// Message contains information about a message.
type Message struct {
	ID           string               `json:"_id"`
	Nonce        string               `json:"nonce"`
	Channel      string               `json:"channel"`
	Author       string               `json:"author"`
	Content      string               `json:"content"`
	Mentions     []string             `json:"mentions"`
	Replies      []string             `json:"replies"`
	Reactions    map[string][]string  `json:"reactions"` // Emoji ID to array of users IDs that reacted
	Pinned       bool                 `json:"pinned"`
	Flags        MessageFlagsType     `json:"flags"`
	Webhook      *MessageWebhook      `json:"webhook"`
	System       *MessageSystem       `json:"system"`
	Embeds       []*MessageEmbed      `json:"embeds"`
	Attachments  []*Attachment        `json:"attachments"`
	Edited       time.Time            `json:"edited"`
	Interactions *MessageInteractions `json:"interactions"`
	Masquerade   *MessageMasquerade   `json:"masquerade"`
}

type MessageWebhook struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type MessageInteractions struct {
	Reactions []string `json:"reactions"`

	// Whether reactions should be restricted to the given list
	RestrictReactions bool `json:"restrict_reactions"`
}

type MessageSystem struct {
	Type MessageSystemType `json:"type"`
	ID   string            `json:"id"`
}

type MessageEdited struct {
	Date int `json:"$date"`
}

type MessageEmbed struct {
	Type        string               `json:"type"`
	URL         string               `json:"url,omitempty"`
	OriginalURL string               `json:"original_url,omitempty"`
	Special     *MessageEmbedSpecial `json:"special,omitempty"`
	Title       string               `json:"title,omitempty"`
	Description string               `json:"description,omitempty"`
	Image       *MessageEmbedImage   `json:"image,omitempty"`
	Video       *MessageEmbedVideo   `json:"video,omitempty"`
	SiteName    string               `json:"site_name,omitempty"`
	IconURL     string               `json:"icon_url,omitempty"`
	Colour      string               `json:"colour,omitempty"`
}

type MessageEmbedSpecial struct {
	Type      MessageEmbedSpecialType `json:"type"`
	ID        string                  `json:"id"`
	Timestamp time.Time               `json:"timestamp,omitempty"`

	// Identifies the type of content for types: Lightspeed, Twitch, Spotify, and Bandcamp
	ContentType string `json:"content_type"` // todo: make enums
}

type MessageEmbedImage struct {
	Size   string `json:"size"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type MessageEmbedVideo struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// MessageSend is used for sending messages to channels
type MessageSend struct {
	Content      string               `json:"content"`
	Attachments  []string             `json:"attachments,omitempty"`
	Replies      []*MessageReplies    `json:"replies,omitempty"`
	Embeds       []*MessageEmbed      `json:"embeds,omitempty"`
	Masquerade   *MessageMasquerade   `json:"masquerade,omitempty"`
	Interactions *MessageInteractions `json:"interactions,omitempty"`
}

type MessageMasquerade struct {
	Name   string `json:"name,omitempty"`
	Avatar string `json:"avatar,omitempty"`
	Colour string `json:"colour,omitempty"`
}

type MessageReplies struct {
	ID      string `json:"id"`
	Mention bool   `json:"mention"`
}
