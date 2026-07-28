package revoltgo

import "time"

//go:generate msgp -tests=false -io=false

type (
	MessageSystemType         string
	MessageEmbedSpecialType   string
	MessageEmbedImageSizeType string
	MessageFlagsType          uint32
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
	MessageSystemMessagePinned             MessageSystemType = "message_pinned"
	MessageSystemMessageUnpinned           MessageSystemType = "message_unpinned"
	MessageSystemCallStarted               MessageSystemType = "call_started"
)

// Derived from:
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/embeds.rs#L158
const (
	MessageEmbedSpecialNone       MessageEmbedSpecialType = "None"
	MessageEmbedSpecialGIF        MessageEmbedSpecialType = "GIF"
	MessageEmbedSpecialYouTube    MessageEmbedSpecialType = "YouTube"
	MessageEmbedSpecialLightspeed MessageEmbedSpecialType = "Lightspeed"
	MessageEmbedSpecialTwitch     MessageEmbedSpecialType = "Twitch"
	MessageEmbedSpecialSpotify    MessageEmbedSpecialType = "Spotify"
	MessageEmbedSpecialSoundcloud MessageEmbedSpecialType = "Soundcloud"
	MessageEmbedSpecialBandcamp   MessageEmbedSpecialType = "Bandcamp"
	MessageEmbedSpecialAppleMusic MessageEmbedSpecialType = "AppleMusic"
	MessageEmbedSpecialStreamable MessageEmbedSpecialType = "Streamable"
)

const (
	MessageFlagsSuppressNotifications MessageFlagsType = 1 // Will not send push / desktop notifications
	MessageFlagsMentionsEveryone      MessageFlagsType = 2 // will mention all users who can see the channel
	MessageFlagsMentionsOnline        MessageFlagsType = 3 // will mention all users who are online and can see the channel. This cannot be true if MentionsEveryone is true
)

// Message contains information about a message.
type Message struct {
	ID           string               `msg:"_id" json:"_id,omitempty"`
	Author       string               `msg:"author" json:"author,omitempty"`
	Channel      string               `msg:"channel" json:"channel,omitempty"`
	Attachments  []*File              `msg:"attachments" json:"attachments,omitempty"`
	Content      string               `msg:"content" json:"content,omitempty"`
	Edited       *time.Time           `msg:"edited" json:"edited,omitempty"`
	Embeds       []*MessageEmbed      `msg:"embeds" json:"embeds,omitempty"`
	Flags        MessageFlagsType     `msg:"flags" json:"flags,omitempty"`
	Interactions *MessageInteractions `msg:"interactions" json:"interactions,omitempty"`
	Masquerade   *MessageMasquerade   `msg:"masquerade" json:"masquerade,omitempty"`
	Member       *ServerMember        `msg:"member" json:"member,omitempty"`
	Mentions     []string             `msg:"mentions" json:"mentions,omitempty"`
	Nonce        string               `msg:"nonce" json:"nonce,omitempty"`
	Pinned       bool                 `msg:"pinned" json:"pinned,omitempty"`

	// Emoji.ID -> []User.ID
	Reactions map[string][]string `msg:"reactions" json:"reactions,omitempty"`

	// []Message.ID's that this message replies to
	Replies []string `msg:"replies" json:"replies,omitempty"`

	// Server that were mentioned
	RoleMentions []string        `msg:"role_mentions" json:"role-mentions,omitempty"`
	System       *MessageSystem  `msg:"system" json:"system,omitempty"`
	User         *User           `msg:"user" json:"user,omitempty"`
	Webhook      *MessageWebhook `msg:"webhook" json:"webhook,omitempty"`
}

// MessageWebhook is derived from:
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/channel_webhooks.rs#L36
type MessageWebhook struct {
	Name   string  `msg:"name" json:"name,omitempty"`
	Avatar *string `msg:"avatar" json:"avatar,omitempty"`
}

func (ms MessageWebhook) AvatarURL(size string) string {
	if ms.Avatar == nil {
		return ""
	}

	return EndpointAutumnFile("avatars", *ms.Avatar, size)
}

type MessageInteractions struct {
	Reactions []string `msg:"reactions" json:"reactions,omitempty"`

	// Whether reactions should be restricted to the given list
	RestrictReactions bool `msg:"restrict_reactions" json:"restrict_reactions,omitempty"`
}

type MessageSystem struct {
	Type MessageSystemType `msg:"type" json:"type,omitempty"`
	ID   string            `msg:"id" json:"id,omitempty"`
}

type MessageEdited struct {
	Date int `msg:"$date" json:"$date,omitempty"`
}

// MessageEmbed is derived from:
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/embeds.rs#L158
type MessageEmbed struct {
	Type        string               `msg:"type" json:"type,omitempty"`
	URL         string               `msg:"url" json:"url,omitempty"`
	OriginalURL string               `msg:"original_url" json:"original_url,omitempty"`
	Special     *MessageEmbedSpecial `msg:"special" json:"special,omitempty"`
	Title       string               `msg:"title" json:"title,omitempty"`
	Description string               `msg:"description" json:"description,omitempty"`
	Image       *MessageEmbedImage   `msg:"image" json:"image,omitempty"`
	Video       *MessageEmbedVideo   `msg:"video" json:"video,omitempty"`
	SiteName    string               `msg:"site_name" json:"site_name,omitempty"`
	IconURL     string               `msg:"icon_url" json:"icon_url,omitempty"`
	Colour      string               `msg:"colour" json:"colour,omitempty"`
	Media       *File                `msg:"media" json:"media,omitempty"`
}

type MessageEmbedSpecial struct {
	Type        MessageEmbedSpecialType `msg:"type" json:"type,omitempty"`
	ID          string                  `msg:"id" json:"id,omitempty"`
	Timestamp   string                  `msg:"timestamp" json:"timestamp,omitempty"`
	ContentType string                  `msg:"content_type" json:"content_type,omitempty"`
	AlbumID     string                  `msg:"album_id" json:"album_id,omitempty"`
	TrackID     string                  `msg:"track_id" json:"track_id,omitempty"`
}

const (
	MessageEmbedImageSizeLarge   MessageEmbedImageSizeType = "Large"
	MessageEmbedImageSizePreview MessageEmbedImageSizeType = "Preview"
)

type MessageEmbedImage struct {
	Size   MessageEmbedImageSizeType `msg:"size" json:"size,omitempty"`
	URL    string                    `msg:"url" json:"url,omitempty"`
	Width  int                       `msg:"width" json:"width,omitempty"`
	Height int                       `msg:"height" json:"height,omitempty"`
}

type MessageEmbedVideo struct {
	URL    string `msg:"url" json:"url,omitempty"`
	Width  int    `msg:"width" json:"width,omitempty"`
	Height int    `msg:"height" json:"height,omitempty"`
}

// MessageSend is used for sending messages to channels
// todo: move to http since this is a sendable request body
type MessageSend struct {
	Content      string               `msg:"content" json:"content,omitempty"`
	Attachments  []string             `msg:"attachments" json:"attachments,omitempty"`
	Replies      []*MessageReplies    `msg:"replies" json:"replies,omitempty"`
	Embeds       []*MessageEmbed      `msg:"embeds" json:"embeds,omitempty"`
	Masquerade   *MessageMasquerade   `msg:"masquerade" json:"masquerade,omitempty"`
	Interactions *MessageInteractions `msg:"interactions" json:"interactions,omitempty"`
}

type MessageMasquerade struct {
	Name   string `msg:"name" json:"name,omitempty"`
	Avatar string `msg:"avatar" json:"avatar,omitempty"`
	Colour string `msg:"colour" json:"colour,omitempty"`
}

type MessageReplies struct {
	ID      string `msg:"id" json:"id,omitempty"`
	Mention bool   `msg:"mention" json:"mention"`
}
