package revoltgo

import (
	"encoding/json"
	"time"
)

type Event struct {
	Type string `json:"type"`
}

var eventToStruct = map[string]func() any{
	"Authenticated": func() any { return new(EventAuthenticated) },
	"Ready":         func() any { return new(EventReady) },
	"Pong":          func() any { return new(EventPong) },
	"Auth":          func() any { return new(EventAuth) },

	"Message":        func() any { return new(EventMessage) },
	"MessageAppend":  func() any { return new(EventMessageAppend) },
	"MessageUpdate":  func() any { return new(EventMessageUpdate) },
	"MessageDelete":  func() any { return new(EventMessageDelete) },
	"MessageReact":   func() any { return new(EventMessageReact) },
	"MessageUnreact": func() any { return new(EventMessageUnreact) },

	"ChannelCreate":      func() any { return new(EventChannelCreate) },
	"ChannelUpdate":      func() any { return new(EventChannelUpdate) },
	"ChannelDelete":      func() any { return new(EventChannelDelete) },
	"ChannelAck":         func() any { return new(EventChannelAck) },
	"ChannelStartTyping": func() any { return new(EventChannelStartTyping) },
	"ChannelStopTyping":  func() any { return new(EventChannelStopTyping) },

	"GroupJoin":  func() any { return new(EventGroupJoin) },
	"GroupLeave": func() any { return new(EventGroupLeave) },

	"ServerCreate":       func() any { return new(EventServerCreate) },
	"ServerUpdate":       func() any { return new(EventServerUpdate) },
	"ServerDelete":       func() any { return new(EventServerDelete) },
	"ServerRoleDelete":   func() any { return new(EventServerRoleDelete) },
	"ServerRoleUpdate":   func() any { return new(EventServerRoleUpdate) },
	"ServerMemberUpdate": func() any { return new(EventServerMemberUpdate) },
	"ServerMemberJoin":   func() any { return new(EventServerMemberJoin) },
	"ServerMemberLeave":  func() any { return new(EventServerMemberLeave) },

	"EmojiCreate": func() any { return new(EventEmojiCreate) },
	"EmojiDelete": func() any { return new(EventEmojiDelete) },

	"UserUpdate":         func() any { return new(EventUserUpdate) },
	"UserSettingsUpdate": func() any { return new(EventUserSettingsUpdate) },
	"UserRelationship":   func() any { return new(EventUserRelationship) },
	"UserPlatformWipe":   func() any { return new(EventUserPlatformWipe) },

	"WebhookCreate": func() any { return new(EventWebhookCreate) },
	"WebhookUpdate": func() any { return new(EventWebhookUpdate) },
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
	ChannelType ChannelType `json:"channel_type"`
	ID          string      `json:"_id"`
	Server      string      `json:"server"`
	Name        string      `json:"name"`
}

// EventServerUpdate is sent when a server is updated. Data will only contain fields that were modified.
type EventServerUpdate struct {
	Event
	ID    string   `json:"id"`
	Data  *Server  `json:"data"`
	Clear []string `json:"clear"`
}

// EventChannelUpdate is sent when a channel is updated. Data will only contain fields that were modified.
type EventChannelUpdate struct {
	Event
	ID    string   `json:"id"`
	Data  *Channel `json:"data"`
	Clear []string `json:"clear"`
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

// EventServerRoleUpdate is sent when a role is updated. Data will only contain fields that were modified.
type EventServerRoleUpdate struct {
	Event
	ID     string      `json:"id"`
	RoleID string      `json:"role_id"`
	Data   *ServerRole `json:"data"`
	Clear  []string    `json:"clear"`
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

// EventServerMemberUpdate is sent when a member is updated. Data will only contain fields that were modified.
type EventServerMemberUpdate struct {
	Event
	ID    MemberCompoundID `json:"id"`
	Data  *ServerMember    `json:"data"`
	Clear []string         `json:"clear"`
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

type EventUserUpdate struct {
	Event
	ID    string   `json:"id"`
	Data  *User    `json:"data"`
	Clear []string `json:"clear"`
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

type EventWebhookUpdate struct {
	Event
	ID     string   `json:"id"`
	Data   *Webhook `json:"data"`
	Remove []string `json:"remove"`
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
