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
	ID           string               `msg:"_id" json:"_id,omitzero"`
	Author       string               `msg:"author" json:"author,omitzero"`
	Channel      string               `msg:"channel" json:"channel,omitzero"`
	Attachments  []*File              `msg:"attachments" json:"attachments,omitzero"`
	Content      string               `msg:"content" json:"content,omitzero"`
	Edited       *time.Time           `msg:"edited" json:"edited,omitzero"`
	Embeds       []*MessageEmbed      `msg:"embeds" json:"embeds,omitzero"`
	Flags        MessageFlagsType     `msg:"flags" json:"flags,omitzero"`
	Interactions *MessageInteractions `msg:"interactions" json:"interactions,omitzero"`
	Masquerade   *MessageMasquerade   `msg:"masquerade" json:"masquerade,omitzero"`
	Member       *ServerMember        `msg:"member" json:"member,omitzero"`
	Mentions     []string             `msg:"mentions" json:"mentions,omitzero"`
	Nonce        string               `msg:"nonce" json:"nonce,omitzero"`
	Pinned       bool                 `msg:"pinned" json:"pinned,omitzero"`

	// Emoji.ID -> []User.ID
	Reactions map[string][]string `msg:"reactions" json:"reactions,omitzero"`

	// []Message.ID's that this message replies to
	Replies []string `msg:"replies" json:"replies,omitzero"`

	// Roles that were mentioned
	RoleMentions []string        `msg:"role_mentions" json:"role_mentions,omitzero"`
	System       *MessageSystem  `msg:"system" json:"system,omitzero"`
	User         *User           `msg:"user" json:"user,omitzero"`
	Webhook      *MessageWebhook `msg:"webhook" json:"webhook,omitzero"`
}

// PartialMessage is derived from:
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/messages.rs#L17
type PartialMessage struct {
	ID           *string              `msg:"_id" json:"_id,omitzero"`
	Author       *string              `msg:"author" json:"author,omitzero"`
	Channel      *string              `msg:"channel" json:"channel,omitzero"`
	Attachments  []*File              `msg:"attachments" json:"attachments,omitzero"`
	Content      *string              `msg:"content" json:"content,omitzero"`
	Edited       *time.Time           `msg:"edited" json:"edited,omitzero"`
	Embeds       []*MessageEmbed      `msg:"embeds" json:"embeds,omitzero"`
	Flags        *MessageFlagsType    `msg:"flags" json:"flags,omitzero"`
	Interactions *MessageInteractions `msg:"interactions" json:"interactions,omitzero"`
	Masquerade   *MessageMasquerade   `msg:"masquerade" json:"masquerade,omitzero"`
	Member       *ServerMember        `msg:"member" json:"member,omitzero"`
	Mentions     []string             `msg:"mentions" json:"mentions,omitzero"`
	Nonce        *string              `msg:"nonce" json:"nonce,omitzero"`
	Pinned       *bool                `msg:"pinned" json:"pinned,omitzero"`

	// Emoji.ID -> []User.ID
	Reactions map[string][]string `msg:"reactions" json:"reactions,omitzero"`

	// []Message.ID's that this message replies to
	Replies []string `msg:"replies" json:"replies,omitzero"`

	// Roles that were mentioned
	RoleMentions []string        `msg:"role_mentions" json:"role_mentions,omitzero"`
	System       *MessageSystem  `msg:"system" json:"system,omitzero"`
	User         *User           `msg:"user" json:"user,omitzero"`
	Webhook      *MessageWebhook `msg:"webhook" json:"webhook,omitzero"`
}

type MessageAppend struct {
	Embeds []*MessageEmbed `msg:"embeds" json:"embeds,omitzero"`
}

// MessageWebhook is derived from:
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/channel_webhooks.rs#L36
type MessageWebhook struct {
	Name   string  `msg:"name" json:"name,omitzero"`
	Avatar *string `msg:"avatar" json:"avatar,omitzero"`
}

func (ms *MessageWebhook) AvatarURL(size string) string {
	if ms.Avatar == nil {
		return ""
	}

	return EndpointAutumnFile(FileTagAvatars, *ms.Avatar, size)
}

type MessageInteractions struct {
	Reactions []string `msg:"reactions" json:"reactions,omitzero"`

	// Whether reactions should be restricted to the given list
	RestrictReactions bool `msg:"restrict_reactions" json:"restrict_reactions,omitzero"`
}

type MessageSystem struct {
	Type MessageSystemType `msg:"type" json:"type,omitzero"`
	ID   string            `msg:"id" json:"id,omitzero"`
}

type MessageEdited struct {
	Date int `msg:"$date" json:"$date,omitzero"`
}

// MessageEmbed is derived from:
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/embeds.rs#L158
type MessageEmbed struct {
	Type        string               `msg:"type" json:"type,omitzero"`
	URL         string               `msg:"url" json:"url,omitzero"`
	OriginalURL string               `msg:"original_url" json:"original_url,omitzero"`
	Special     *MessageEmbedSpecial `msg:"special" json:"special,omitzero"`
	Title       string               `msg:"title" json:"title,omitzero"`
	Description string               `msg:"description" json:"description,omitzero"`
	Image       *MessageEmbedImage   `msg:"image" json:"image,omitzero"`
	Video       *MessageEmbedVideo   `msg:"video" json:"video,omitzero"`
	SiteName    string               `msg:"site_name" json:"site_name,omitzero"`
	IconURL     string               `msg:"icon_url" json:"icon_url,omitzero"`
	Colour      string               `msg:"colour" json:"colour,omitzero"`
	Media       *File                `msg:"media" json:"media,omitzero"`
}

type MessageEmbedSpecial struct {
	Type        MessageEmbedSpecialType `msg:"type" json:"type,omitzero"`
	ID          string                  `msg:"id" json:"id,omitzero"`
	Timestamp   string                  `msg:"timestamp" json:"timestamp,omitzero"`
	ContentType string                  `msg:"content_type" json:"content_type,omitzero"`
	AlbumID     string                  `msg:"album_id" json:"album_id,omitzero"`
	TrackID     string                  `msg:"track_id" json:"track_id,omitzero"`
}

const (
	MessageEmbedImageSizeLarge   MessageEmbedImageSizeType = "Large"
	MessageEmbedImageSizePreview MessageEmbedImageSizeType = "Preview"
)

type MessageEmbedImage struct {
	Size   MessageEmbedImageSizeType `msg:"size" json:"size,omitzero"`
	URL    string                    `msg:"url" json:"url,omitzero"`
	Width  int                       `msg:"width" json:"width,omitzero"`
	Height int                       `msg:"height" json:"height,omitzero"`
}

type MessageEmbedVideo struct {
	URL    string `msg:"url" json:"url,omitzero"`
	Width  int    `msg:"width" json:"width,omitzero"`
	Height int    `msg:"height" json:"height,omitzero"`
}

// MessageSend is used for sending messages to channels
// todo: move to http since this is a sendable request body
type MessageSend struct {
	Content      string               `msg:"content" json:"content,omitzero"`
	Attachments  []string             `msg:"attachments" json:"attachments,omitzero"`
	Replies      []*MessageReplies    `msg:"replies" json:"replies,omitzero"`
	Embeds       []*MessageEmbed      `msg:"embeds" json:"embeds,omitzero"`
	Masquerade   *MessageMasquerade   `msg:"masquerade" json:"masquerade,omitzero"`
	Interactions *MessageInteractions `msg:"interactions" json:"interactions,omitzero"`
}

type MessageMasquerade struct {
	Name   string `msg:"name" json:"name,omitzero"`
	Avatar string `msg:"avatar" json:"avatar,omitzero"`
	Colour string `msg:"colour" json:"colour,omitzero"`
}

type MessageReplies struct {
	ID      string `msg:"id" json:"id,omitzero"`
	Mention bool   `msg:"mention" json:"mention"`
}
