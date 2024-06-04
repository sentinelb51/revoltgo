package revoltgo

import (
	"github.com/goccy/go-json"
	"time"
)

type Event struct {
	Type string `json:"type"`
}

type AbstractEventUpdateID struct {
	StringID string
	MemberID MemberCompoundID
}

func (id *AbstractEventUpdateID) UnmarshalJSON(data []byte) (err error) {

	if err = json.Unmarshal(data, &id.StringID); err != nil {
		err = json.Unmarshal(data, &id.MemberID)
	}

	return
}

// AbstractEventUpdate is a generic event for all update events.
// This is mainly used to update the cache, and is not a low-level event.
type AbstractEventUpdate struct {
	Event

	// ID can either be a simple string or a MemberCompoundID.
	ID AbstractEventUpdateID `json:"id"`

	// RoleID is only present in ServerRoleUpdate events.
	RoleID string `json:"role_id"`

	// The updated data for a specific event
	Data map[string]any `json:"data"`

	// Clear is a list of keys to clear from the cache.
	Clear []string `json:"clear"`
}

var eventToStruct = map[string]func() any{
	"Authenticated": func() any { return new(EventAuthenticated) },
	"Ready":         func() any { return new(EventReady) },
	"Pong":          func() any { return new(EventPong) },
	"Auth":          func() any { return new(EventAuth) },

	/* All update events are abstracted away. */
	"MessageUpdate":    func() any { return new(AbstractEventUpdate) },
	"ServerUpdate":     func() any { return new(AbstractEventUpdate) },
	"ChannelUpdate":    func() any { return new(AbstractEventUpdate) },
	"ServerRoleUpdate": func() any { return new(AbstractEventUpdate) },
	"WebhookUpdate":    func() any { return new(AbstractEventUpdate) },
	"UserUpdate":       func() any { return new(AbstractEventUpdate) },
	"ServerMemberUpdate": func() any {
		return new(AbstractEventUpdate)
	},

	"Message":        func() any { return new(EventMessage) },
	"MessageAppend":  func() any { return new(EventMessageAppend) },
	"MessageDelete":  func() any { return new(EventMessageDelete) },
	"MessageReact":   func() any { return new(EventMessageReact) },
	"MessageUnreact": func() any { return new(EventMessageUnreact) },

	"ChannelCreate":      func() any { return new(EventChannelCreate) },
	"ChannelDelete":      func() any { return new(EventChannelDelete) },
	"ChannelAck":         func() any { return new(EventChannelAck) },
	"ChannelStartTyping": func() any { return new(EventChannelStartTyping) },
	"ChannelStopTyping":  func() any { return new(EventChannelStopTyping) },

	"GroupJoin":  func() any { return new(EventGroupJoin) },
	"GroupLeave": func() any { return new(EventGroupLeave) },

	"ServerCreate":      func() any { return new(EventServerCreate) },
	"ServerDelete":      func() any { return new(EventServerDelete) },
	"ServerRoleDelete":  func() any { return new(EventServerRoleDelete) },
	"ServerMemberJoin":  func() any { return new(EventServerMemberJoin) },
	"ServerMemberLeave": func() any { return new(EventServerMemberLeave) },

	"EmojiCreate": func() any { return new(EventEmojiCreate) },
	"EmojiDelete": func() any { return new(EventEmojiDelete) },

	"UserSettingsUpdate": func() any { return new(EventUserSettingsUpdate) },
	"UserRelationship":   func() any { return new(EventUserRelationship) },
	"UserPlatformWipe":   func() any { return new(EventUserPlatformWipe) },

	"WebhookCreate": func() any { return new(EventWebhookCreate) },
	"WebhookDelete": func() any { return new(EventWebhookDelete) },

	"ReportCreate": func() any { return new(EventReportCreate) },
}

type EventPong struct {
	Event
	Data int `json:"data"`
}

// EventReady stores the data from the websocket ready event.
// This is used to populate the session's cache
type EventReady struct {
	Event
	Users    []*User         `json:"users"`
	Servers  []*Server       `json:"servers"`
	Channels []*Channel      `json:"channels"`
	Members  []*ServerMember `json:"members"`
	Emojis   []*Emoji        `json:"emojis"`
}

type AuthType string

const (
	EventTypeAuthDeleteSession     AuthType = "DeleteSession"
	EventTypeAuthDeleteAllSessions AuthType = "DeleteAllSessions"
)

type EventAuth struct {
	Event
	EventType AuthType `json:"event_type"`
	UserID    string   `json:"user_id"`
	SessionID string   `json:"session_id"`

	// Only present when... I forgot.
	ExcludeSessionID string `json:"exclude_session_id"`
}

// EventAuthenticated is sent after the client has authenticated.
type EventAuthenticated struct {
	Event
}

type EventMessage struct {
	Event
	Message
}

type EventMessageUpdate struct {
	Event
	ID      string                 `json:"id"`
	Channel string                 `json:"channel"`
	Data    EventMessageUpdateData `json:"data"`
}

type EventMessageUpdateData struct {
	Content string         `json:"content"`
	Edited  time.Time      `json:"edited"`
	Embeds  []MessageEmbed `json:"embeds"`
}

type EventMessageAppend struct {
	ID      string  `json:"id"`
	Channel string  `json:"channel"`
	Append  Message `json:"append"`
}

type EventMessageDelete struct {
	Event
	ID      string `json:"id"`
	Channel string `json:"channel"`
}

// EventChannelStartTyping is sent when a user starts typing in a channel.
type EventChannelStartTyping struct {
	Event
	ID   string `json:"id"`
	User string `json:"user,omitempty"`
}

// EventChannelStopTyping is sent when a user stops typing in a channel.
type EventChannelStopTyping struct {
	EventChannelStartTyping
}

type EventChannelAck struct {
	Event
	ID        string `json:"id"`
	User      string `json:"user"`
	MessageID string `json:"message_id"`
}

// EventChannelCreate is sent when a channel is created.
// This is dispatched in conjunction with EventServerUpdate
type EventChannelCreate struct {
	Event
	Type   ChannelType `json:"channel_type"`
	ID     string      `json:"_id"`
	Server string      `json:"server"`
	Name   string      `json:"name"`
}

// EventChannelDelete is sent when a channel is deleted.
type EventChannelDelete struct {
	Event
	ID string `json:"id"`
}

// EventServerMemberLeave is sent when a user leaves a server.
type EventServerMemberLeave struct {
	Event
	ID   string `json:"id"`
	User string `json:"user"`
}

// EventServerCreate is sent when a server is created (joined).
type EventServerCreate struct {
	Event
	ID     string  `json:"id"`
	Server *Server `json:"server"`
}

type EventServerRoleDelete struct {
	Event
	ID     string `json:"id"`
	RoleID string `json:"role_id"`
}

type EventServerMemberJoin struct {
	Event
	ID   string `json:"id"`
	User string `json:"user"`
}

type EventServerDelete struct {
	Event
	ID string `json:"id"`
}

type EventMessageReact struct {
	Event
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	EmojiID   string `json:"emoji_id"`
}

type EventMessageUnreact struct {
	EventMessageReact
}

type EventGroupJoin struct {
	Event
	ID   string `json:"id"`
	User string `json:"user"`
}

type EventGroupLeave struct {
	EventGroupJoin
}

type EventEmojiCreate struct {
	Event
	*Emoji
}

type EventEmojiDelete struct {
	Event
	ID string `json:"id"`
}

type EventUserRelationship struct {
	Event
	ID   string `json:"id"`
	User *User  `json:"user"`
}

type EventUserPlatformWipe struct {
	Event
	UserID string `json:"user_id"`
	Flags  int    `json:"flags"`
}

type EventUserSettingsUpdate struct {
	Event
	// Update is a tuple of (int, string); update time, and the data in JSON
	Update map[string][2]json.RawMessage `json:"update"`
}

type EventWebhookCreate struct {
	Event
	*Webhook
}

type EventWebhookDelete struct {
	Event
	ID string `json:"id"`
}

// EventReportCreate might not be broadcasted in the websocket for everyone
type EventReportCreate struct {
	Event
	*Report
}
