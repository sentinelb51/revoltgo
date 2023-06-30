package revoltgo

import (
	"time"
)

type (
	MessageSystemType       string
	MessageEmbedSpecialType string
)

const (
	MessageSystemTypeText                      MessageSystemType = "text"
	MessageSystemTypeUserAdded                 MessageSystemType = "user_added"
	MessageSystemTypeUserRemove                MessageSystemType = "user_remove"
	MessageSystemTypeUserJoined                MessageSystemType = "user_joined"
	MessageSystemTypeUserLeft                  MessageSystemType = "user_left"
	MessageSystemTypeUserKicked                MessageSystemType = "user_kicked"
	MessageSystemTypeUserBanned                MessageSystemType = "user_banned"
	MessageSystemTypeChannelRenamed            MessageSystemType = "channel_renamed"
	MessageSystemTypeChannelDescriptionChanged MessageSystemType = "channel_description_changed"
	MessageSystemTypeChannelIconChanged        MessageSystemType = "channel_icon_changed"
	MessageSystemTypeChannelOwnershipChanged   MessageSystemType = "channel_ownership_changed"
)

const (
	MessageEmbedSpecialTypeNone       = "None"
	MessageEmbedSpecialTypeGIF        = "GIF"
	MessageEmbedSpecialTypeYouTube    = "YouTube"
	MessageEmbedSpecialTypeLightspeed = "Lightspeed"
	MessageEmbedSpecialTypeTwitch     = "Twitch"
	MessageEmbedSpecialTypeSpotify    = "Spotify"
	MessageEmbedSpecialTypeSoundcloud = "Soundcloud"
	MessageEmbedSpecialTypeBandcamp   = "Bandcamp"
	MessageEmbedSpecialTypeStreamable = "Streamable"
)

// Message contains information about a message.
type Message struct {
	ID          string          `json:"_id"`
	Nonce       string          `json:"nonce"`
	Channel     string          `json:"channel"`
	Author      string          `json:"author"`
	Webhook     *MessageWebhook `json:"webhook"`
	Content     string          `json:"content"`
	System      *MessageSystem  `json:"system"`
	Attachments []*Attachment   `json:"attachments"`
	Edited      time.Time       `json:"edited"`
	Embeds      []*MessageEmbed `json:"embeds"`
	Mentions    []string        `json:"mentions"`
	Replies     []string        `json:"replies"`

	// Map of emoji ID to array of user ID who reacted to it
	Reactions map[string][]string `json:"reactions"`

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
	ContentType string `json:"content_type"`
}

type MessageEmbedImage struct {
	Size   string `json:"size"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type MessageEmbedVideo struct {
	Url    string `json:"url"`
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
