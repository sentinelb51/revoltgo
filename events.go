package revoltgo

import (
	"time"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false -v=true

type WebsocketMessageType string

const (
	WebsocketKeepAlivePeriod = 60 * time.Second

	WebsocketMessageTypeAuthenticate WebsocketMessageType = "Authenticate"
	WebsocketMessageTypeHeartbeat    WebsocketMessageType = "Ping"
	WebsocketMessageTypeBeginTyping  WebsocketMessageType = "BeginTyping"
	WebsocketMessageTypeEndTyping    WebsocketMessageType = "EndTyping"
)

type WebsocketMessageAuthenticate struct {
	Type  WebsocketMessageType `msg:"type" json:"type,omitempty"`
	Token string               `msg:"token" json:"token,omitempty"`
}

type WebsocketMessagePing struct {
	Type WebsocketMessageType `msg:"type" json:"type,omitempty"`
	Data int64                `msg:"data" json:"data,omitempty"`
}

type WebsocketChannelTyping struct {
	Type    WebsocketMessageType `msg:"type" json:"type,omitempty"`
	Channel string               `msg:"channel" json:"channel,omitempty"`
}

type EventErrorType string

const (
	EventErrorTypeLabelMe               EventErrorType = "LabelMe"
	EventErrorTypeInternalError         EventErrorType = "InternalError"
	EventErrorTypeInvalidSession        EventErrorType = "InvalidSession"
	EventErrorTypeOnboardingNotFinished EventErrorType = "OnboardingNotFinished"
	EventErrorTypeAlreadyAuthenticated  EventErrorType = "AlreadyAuthenticated"
)

type EventError struct {
	Event
	// https://developers.revolt.chat/developers/events/protocol.html#error
	Error EventErrorType `msg:"error" json:"error,omitempty"`
}

type EventBulk struct {
	Event
	V []msgp.Raw `msg:"v" json:"v,omitempty"`
}

type EventPong struct {
	Event
	Data int64 `msg:"data" json:"data,omitempty"`
}

// EventReady provides information about objects relative to the user.
// This is used to populate the session's cache
type EventReady struct {
	Event
	Users    []*User         `msg:"users" json:"users,omitempty"`
	Servers  []*Server       `msg:"servers" json:"servers,omitempty"`
	Channels []*Channel      `msg:"channels" json:"channels,omitempty"`
	Members  []*ServerMember `msg:"members" json:"members,omitempty"`
	Emojis   []*Emoji        `msg:"emojis" json:"emojis,omitempty"`
}

type AuthType string

const (
	EventTypeAuthDeleteSession     AuthType = "DeleteSession"
	EventTypeAuthDeleteAllSessions AuthType = "DeleteAllSessions"
)

type EventAuth struct {
	Event
	EventType AuthType `msg:"event_type" json:"event_type,omitempty"`
	UserID    string   `msg:"user_id" json:"user_id,omitempty"`
	SessionID string   `msg:"session_id" json:"session_id,omitempty"`

	// Only present when... I forgot.
	ExcludeSessionID string `msg:"exclude_session_id" json:"exclude_session_id,omitempty"`
}

// EventAuthenticated is sent after the client has authenticated.
type EventAuthenticated struct {
	Event `msg:",flatten"`
}

type EventMessage struct {
	Event   `msg:",flatten"`
	Message `msg:",flatten"`
}

// EventServerUpdate is sent when a server is updated. Data will only contain fields that were modified.
type EventServerUpdate struct {
	Event `msg:",flatten"`
	ID    string        `msg:"id" json:"id,omitempty"`
	Data  PartialServer `msg:"data" json:"data,omitempty"`
	Clear []string      `msg:"clear" json:"clear,omitempty"`
}

// EventChannelUpdate is sent when a channel is updated. Data will only contain fields that were modified.
type EventChannelUpdate struct {
	Event `msg:",flatten"`
	ID    string         `msg:"id" json:"id,omitempty"`
	Data  PartialChannel `msg:"data" json:"data,omitempty"`
	Clear []string       `msg:"clear" json:"clear,omitempty"`
}

// EventServerRoleUpdate is sent when a role is updated. Data will only contain fields that were modified.
type EventServerRoleUpdate struct {
	Event  `msg:",flatten"`
	ID     string            `msg:"id" json:"id,omitempty"`
	RoleID string            `msg:"role_id" json:"role_id,omitempty"`
	Data   PartialServerRole `msg:"data" json:"data,omitempty"`
	Clear  []string          `msg:"clear" json:"clear,omitempty"`
}

// EventServerMemberUpdate is sent when a member is updated. Data will only contain fields that were modified.
type EventServerMemberUpdate struct {
	Event `msg:",flatten"`
	ID    MemberCompositeID   `msg:"id" json:"id,omitempty"`
	Data  PartialServerMember `msg:"data" json:"data,omitempty"`
	Clear []string            `msg:"clear" json:"clear,omitempty"`
}

type EventUserUpdate struct {
	Event `msg:",flatten"`
	ID    string      `msg:"id" json:"id,omitempty"`
	Data  PartialUser `msg:"data" json:"data,omitempty"`
	Clear []string    `msg:"clear" json:"clear,omitempty"`
}

type EventWebhookUpdate struct {
	Event  `msg:",flatten"`
	ID     string         `msg:"id" json:"id,omitempty"`
	Data   PartialWebhook `msg:"data" json:"data,omitempty"`
	Remove []string       `msg:"remove" json:"remove,omitempty"` // todo: why is this "remove" and not "clear"?
}

type EventMessageUpdate struct {
	Event   `msg:",flatten"`
	ID      string  `msg:"id" json:"id,omitempty"`
	Channel string  `msg:"channel" json:"channel,omitempty"`
	Data    Message `msg:"data" json:"data,omitempty"`
}

type EventMessageAppend struct {
	ID      string  `msg:"id" json:"id,omitempty"`
	Channel string  `msg:"channel" json:"channel,omitempty"`
	Append  Message `msg:"append" json:"append,omitempty"`
}

type EventMessageDelete struct {
	Event   `msg:",flatten"`
	ID      string `msg:"id" json:"id,omitempty"`
	Channel string `msg:"channel" json:"channel,omitempty"`
}

type EventBulkMessageDelete struct {
	Event   `msg:",flatten"`
	Channel string   `msg:"channel" json:"channel,omitempty"`
	IDs     []string `msg:"ids" json:"ids,omitempty"`
}

// EventChannelStartTyping is sent when a user starts typing in a channel.
type EventChannelStartTyping struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
	User  string `msg:"user" json:"user,omitempty"`
}

// EventChannelStopTyping is sent when a user stops typing in a channel.
type EventChannelStopTyping struct {
	EventChannelStartTyping `msg:",flatten"`
}

type EventChannelAck struct {
	Event     `msg:",flatten"`
	ID        string `msg:"id" json:"id,omitempty"`
	User      string `msg:"user" json:"user,omitempty"`
	MessageID string `msg:"message_id" json:"message_id,omitempty"`
}

// EventChannelCreate is sent when a channel is created.
// This is dispatched in conjunction with EventServerUpdate
type EventChannelCreate struct {
	Event   `msg:",flatten"`
	Channel `msg:",flatten"`
}

// EventChannelDelete is sent when a channel is deleted.
type EventChannelDelete struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
}

// EventServerMemberLeave is sent when a user leaves a server.
type EventServerMemberLeave struct {
	Event  `msg:",flatten"`
	ID     string `msg:"id" json:"id,omitempty"`
	User   string `msg:"user" json:"user,omitempty"`
	Reason string `msg:"reason" json:"reason,omitempty"`
}

// EventServerCreate is sent when a server is created (joined).
type EventServerCreate struct {
	Event    `msg:",flatten"`
	ID       string     `msg:"id" json:"id,omitempty"`
	Server   *Server    `msg:"server" json:"server,omitempty"`
	Channels []*Channel `msg:"channels" json:"channels,omitempty"`
	Emojis   []*Emoji   `msg:"emojis" json:"emojis,omitempty"`
}

type EventServerRoleDelete struct {
	Event  `msg:",flatten"`
	ID     string `msg:"id" json:"id,omitempty"`
	RoleID string `msg:"role_id" json:"role_id,omitempty"`
}

type EventServerMemberJoin struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
	User  string `msg:"user" json:"user,omitempty"`
}

type EventServerDelete struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
}

type EventMessageReact struct {
	Event     `msg:",flatten"`
	ID        string `msg:"id" json:"id,omitempty"`
	ChannelID string `msg:"channel_id" json:"channel_id,omitempty"`
	UserID    string `msg:"user_id" json:"user_id,omitempty"`
	EmojiID   string `msg:"emoji_id" json:"emoji_id,omitempty"`
}

// EventMessageUnreact is sent when a user removes a singular reaction from a message.
type EventMessageUnreact struct {
	EventMessageReact `msg:",flatten"`
}

// EventMessageRemoveReaction is sent when all the reactions are removed from a message.
type EventMessageRemoveReaction struct {
	ID        string `msg:"id" json:"id,omitempty"`
	ChannelID string `msg:"channel_id" json:"channel_id,omitempty"`
	EmojiID   string `msg:"emoji_id" json:"emoji_id,omitempty"`
}

type EventChannelGroupJoin struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
	User  string `msg:"user" json:"user,omitempty"`
}

type EventChannelGroupLeave struct {
	EventChannelGroupJoin `msg:",flatten"`
}

type EventEmojiCreate struct {
	Event `msg:",flatten"`
	Emoji `msg:",flatten"`
}

type EventEmojiDelete struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
}

type EventUserRelationship struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
	User  *User  `msg:"user" json:"user,omitempty"`
}

type EventUserPlatformWipe struct {
	Event
	UserID string `msg:"user_id" json:"user_id,omitempty"`
	Flags  int    `msg:"flags" json:"flags,omitempty"`
}

type EventUserSettingsUpdate struct {
	Event `msg:",flatten"`
	// Update is a tuple of (int, string); update time, and the data in JSON
	Update map[string]SyncSettingsDataTuple `msg:"update" json:"update,omitempty"`
}

type EventWebhookCreate struct {
	Event   `msg:",flatten"`
	Webhook `msg:",flatten"`
}

type EventWebhookDelete struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
}

type EventVoiceChannelJoin struct {
	Event `msg:",flatten"`
	ID    string         `msg:"id" json:"id,omitempty"`
	State UserVoiceState `msg:"state" json:"state,omitempty"`
}

type EventVoiceChannelLeave struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitempty"`
	User  string `msg:"user" json:"user"`
}

type EventVoiceChannelMove struct {
	Event `msg:",flatten"`
	User  string         `msg:"user" json:"user,omitempty"`
	From  string         `msg:"from" json:"from,omitempty"`
	To    string         `msg:"to" json:"to,omitempty"`
	State UserVoiceState `msg:"state" json:"state,omitempty"`
}

type EventUserVoiceStateUpdate struct {
	Event     `msg:",flatten"`
	ID        string                `msg:"id" json:"id,omitempty"`
	ChannelID string                `msg:"channel_id" json:"channel_id,omitempty"`
	Data      PartialUserVoiceState `msg:"data" json:"data,omitempty"`
}

type EventUserMoveVoiceChannel struct {
	Event `msg:",flatten"`
	Node  string `msg:"node" json:"node,omitempty"`
	Token string `msg:"token" json:"token,omitempty"`
}
