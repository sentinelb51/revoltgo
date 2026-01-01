package revoltgo

import (
	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false -v=true

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
	Error EventErrorType `msg:"error"`
}

type EventBulk struct {
	Event
	V []msgp.Raw `msg:"v"`
}

type EventPong struct {
	Event
	Data int64 `msg:"data"`
}

// EventReady provides information about objects relative to the user.
// This is used to populate the session's cache
type EventReady struct {
	Event
	Users    []*User         `msg:"users"`
	Servers  []*Server       `msg:"servers"`
	Channels []*Channel      `msg:"channels"`
	Members  []*ServerMember `msg:"members"`
	Emojis   []*Emoji        `msg:"emojis"`
}

type AuthType string

const (
	EventTypeAuthDeleteSession     AuthType = "DeleteSession"
	EventTypeAuthDeleteAllSessions AuthType = "DeleteAllSessions"
)

type EventAuth struct {
	Event
	EventType AuthType `msg:"event_type"`
	UserID    string   `msg:"user_id"`
	SessionID string   `msg:"session_id"`

	// Only present when... I forgot.
	ExcludeSessionID string `msg:"exclude_session_id"`
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
	ID    string        `msg:"id"`
	Data  PartialServer `msg:"data"`
	Clear []string      `msg:"clear"`
}

// EventChannelUpdate is sent when a channel is updated. Data will only contain fields that were modified.
type EventChannelUpdate struct {
	Event `msg:",flatten"`
	ID    string         `msg:"id"`
	Data  PartialChannel `msg:"data"`
	Clear []string       `msg:"clear"`
}

// EventServerRoleUpdate is sent when a role is updated. Data will only contain fields that were modified.
type EventServerRoleUpdate struct {
	Event  `msg:",flatten"`
	ID     string            `msg:"id"`
	RoleID string            `msg:"role_id"`
	Data   PartialServerRole `msg:"data"`
	Clear  []string          `msg:"clear"`
}

// EventServerMemberUpdate is sent when a member is updated. Data will only contain fields that were modified.
type EventServerMemberUpdate struct {
	Event `msg:",flatten"`
	ID    MemberCompositeID   `msg:"id"`
	Data  PartialServerMember `msg:"data"`
	Clear []string            `msg:"clear"`
}

type EventUserUpdate struct {
	Event `msg:",flatten"`
	ID    string      `msg:"id"`
	Data  PartialUser `msg:"data"`
	Clear []string    `msg:"clear"`
}

type EventWebhookUpdate struct {
	Event  `msg:",flatten"`
	ID     string         `msg:"id"`
	Data   PartialWebhook `msg:"data"`
	Remove []string       `msg:"remove"` // todo: why is this "remove" and not "clear"?
}

type EventMessageUpdate struct {
	Event   `msg:",flatten"`
	ID      string  `msg:"id"`
	Channel string  `msg:"channel"`
	Data    Message `msg:"data"`
}

type EventMessageAppend struct {
	ID      string  `msg:"id"`
	Channel string  `msg:"channel"`
	Append  Message `msg:"append"`
}

type EventMessageDelete struct {
	Event   `msg:",flatten"`
	ID      string `msg:"id"`
	Channel string `msg:"channel"`
}

type EventBulkMessageDelete struct {
	Event   `msg:",flatten"`
	Channel string   `msg:"channel"`
	IDs     []string `msg:"ids"`
}

// EventChannelStartTyping is sent when a user starts typing in a channel.
type EventChannelStartTyping struct {
	Event `msg:",flatten"`
	ID    string `msg:"id"`
	User  string `msg:"user,omitempty"`
}

// EventChannelStopTyping is sent when a user stops typing in a channel.
type EventChannelStopTyping struct {
	EventChannelStartTyping `msg:",flatten"`
}

type EventChannelAck struct {
	Event     `msg:",flatten"`
	ID        string `msg:"id"`
	User      string `msg:"user"`
	MessageID string `msg:"message_id"`
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
	ID    string `msg:"id"`
}

// EventServerMemberLeave is sent when a user leaves a server.
type EventServerMemberLeave struct {
	Event  `msg:",flatten"`
	ID     string `msg:"id"`
	User   string `msg:"user"`
	Reason string `msg:"reason"`
}

// EventServerCreate is sent when a server is created (joined).
type EventServerCreate struct {
	Event    `msg:",flatten"`
	ID       string     `msg:"id"`
	Server   *Server    `msg:"server"`
	Channels []*Channel `msg:"channels"`
	Emojis   []*Emoji   `msg:"emojis"`
}

type EventServerRoleDelete struct {
	Event  `msg:",flatten"`
	ID     string `msg:"id"`
	RoleID string `msg:"role_id"`
}

type EventServerMemberJoin struct {
	Event `msg:",flatten"`
	ID    string `msg:"id"`
	User  string `msg:"user"`
}

type EventServerDelete struct {
	Event `msg:",flatten"`
	ID    string `msg:"id"`
}

type EventMessageReact struct {
	Event     `msg:",flatten"`
	ID        string `msg:"id"`
	ChannelID string `msg:"channel_id"`
	UserID    string `msg:"user_id"`
	EmojiID   string `msg:"emoji_id"`
}

// EventMessageUnreact is sent when a user removes a singular reaction from a message.
type EventMessageUnreact struct {
	EventMessageReact `msg:",flatten"`
}

// EventMessageRemoveReaction is sent when all the reactions are removed from a message.
type EventMessageRemoveReaction struct {
	ID        string `msg:"id"`
	ChannelID string `msg:"channel_id"`
	EmojiID   string `msg:"emoji_id"`
}

type EventChannelGroupJoin struct {
	Event `msg:",flatten"`
	ID    string `msg:"id"`
	User  string `msg:"user"`
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
	ID    string `msg:"id"`
}

type EventUserRelationship struct {
	Event `msg:",flatten"`
	ID    string `msg:"id"`
	User  *User  `msg:"user"`
}

type EventUserPlatformWipe struct {
	Event
	UserID string `msg:"user_id"`
	Flags  int    `msg:"flags"`
}

type EventUserSettingsUpdate struct {
	Event `msg:",flatten"`
	// Update is a tuple of (int, string); update time, and the data in JSON
	Update map[string]SyncSettingsDataTuple `msg:"update"`
}

type EventWebhookCreate struct {
	Event   `msg:",flatten"`
	Webhook `msg:",flatten"`
}

type EventWebhookDelete struct {
	Event `msg:",flatten"`
	ID    string `msg:"id"`
}
