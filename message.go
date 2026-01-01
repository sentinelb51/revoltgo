package revoltgo

//go:generate msgp -tests=false -io=false

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
	ID           string               `msg:"_id"`
	Nonce        string               `msg:"nonce"`
	Channel      string               `msg:"channel"`
	Author       string               `msg:"author"`
	Content      string               `msg:"content"`
	Mentions     []string             `msg:"mentions"`
	Replies      []string             `msg:"replies"`
	Reactions    map[string][]string  `msg:"reactions"` // Emoji ID to array of users IDs that reacted
	Pinned       bool                 `msg:"pinned"`
	Flags        MessageFlagsType     `msg:"flags"`
	Webhook      *MessageWebhook      `msg:"webhook"`
	System       *MessageSystem       `msg:"system"`
	Embeds       []*MessageEmbed      `msg:"embeds"`
	Attachments  []*Attachment        `msg:"attachments"`
	Edited       Timestamp            `msg:"edited"`
	Interactions *MessageInteractions `msg:"interactions"`
	Masquerade   *MessageMasquerade   `msg:"masquerade"`
}

type MessageWebhook struct {
	Name   string `msg:"name"`
	Avatar string `msg:"avatar"`
}

type MessageInteractions struct {
	Reactions []string `msg:"reactions"`

	// Whether reactions should be restricted to the given list
	RestrictReactions bool `msg:"restrict_reactions"`
}

type MessageSystem struct {
	Type MessageSystemType `msg:"type"`
	ID   string            `msg:"id"`
}

type MessageEdited struct {
	Date int `msg:"$date"`
}

type MessageEmbed struct {
	Type        string               `msg:"type"`
	URL         string               `msg:"url,omitempty"`
	OriginalURL string               `msg:"original_url,omitempty"`
	Special     *MessageEmbedSpecial `msg:"special,omitempty"`
	Title       string               `msg:"title,omitempty"`
	Description string               `msg:"description,omitempty"`
	Image       *MessageEmbedImage   `msg:"image,omitempty"`
	Video       *MessageEmbedVideo   `msg:"video,omitempty"`
	SiteName    string               `msg:"site_name,omitempty"`
	IconURL     string               `msg:"icon_url,omitempty"`
	Colour      string               `msg:"colour,omitempty"`
}

type MessageEmbedSpecial struct {
	Type      MessageEmbedSpecialType `msg:"type"`
	ID        string                  `msg:"id"`
	Timestamp Timestamp               `msg:"timestamp,omitempty"`

	// Identifies the type of content for types: Lightspeed, Twitch, Spotify, and Bandcamp
	ContentType string `msg:"content_type"` // todo: make enums
}

type MessageEmbedImage struct {
	Size   string `msg:"size"`
	URL    string `msg:"url"`
	Width  int    `msg:"width"`
	Height int    `msg:"height"`
}

type MessageEmbedVideo struct {
	URL    string `msg:"url"`
	Width  int    `msg:"width"`
	Height int    `msg:"height"`
}

// MessageSend is used for sending messages to channels
type MessageSend struct {
	Content      string               `msg:"content"`
	Attachments  []string             `msg:"attachments,omitempty"`
	Replies      []*MessageReplies    `msg:"replies,omitempty"`
	Embeds       []*MessageEmbed      `msg:"embeds,omitempty"`
	Masquerade   *MessageMasquerade   `msg:"masquerade,omitempty"`
	Interactions *MessageInteractions `msg:"interactions,omitempty"`
}

type MessageMasquerade struct {
	Name   string `msg:"name,omitempty"`
	Avatar string `msg:"avatar,omitempty"`
	Colour string `msg:"colour,omitempty"`
}

type MessageReplies struct {
	ID      string `msg:"id"`
	Mention bool   `msg:"mention"`
}
