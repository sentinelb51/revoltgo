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
	Type  WebsocketMessageType `msg:"type" json:"type,omitzero"`
	Token string               `msg:"token" json:"token,omitzero"`
}

type WebsocketMessagePing struct {
	Type WebsocketMessageType `msg:"type" json:"type,omitzero"`
	Data int64                `msg:"data" json:"data,omitzero"`
}

type WebsocketChannelTyping struct {
	Type    WebsocketMessageType `msg:"type" json:"type,omitzero"`
	Channel string               `msg:"channel" json:"channel,omitzero"`
}

// EventErrorDataType is derived from
// https://developers.revolt.chat/developers/events/protocol.html#error
type EventErrorDataType string

const (
	EventErrorLabelMe               EventErrorDataType = "LabelMe"
	EventErrorInternalError         EventErrorDataType = "InternalError"
	EventErrorInvalidSession        EventErrorDataType = "InvalidSession"
	EventErrorOnboardingNotFinished EventErrorDataType = "OnboardingNotFinished"
	EventErrorAlreadyAuthenticated  EventErrorDataType = "AlreadyAuthenticated"
)

type EventErrorData struct {
	Type     EventErrorDataType `msg:"type" json:"type,omitzero"`
	Location string             `msg:"location" json:"location,omitzero"`
}

type EventError struct {
	Event
	Data EventErrorData `msg:"data" json:"data,omitzero"`
}

type EventBulk struct {
	Event
	V []msgp.Raw `msg:"v" json:"v,omitzero"`
}

type EventPong struct {
	Event
	Data int64 `msg:"data" json:"data,omitzero"`
}

// EventReadyPolicyChange is derived from:
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/policy_changes.rs
type EventReadyPolicyChange struct {
	CreatedTime   time.Time `msg:"created_time" json:"created_time,omitzero"`
	EffectiveTime time.Time `msg:"effective_time" json:"effective_time,omitzero"`
	Description   string    `msg:"description" json:"description,omitzero"`
	URL           string    `msg:"url" json:"url,omitzero"`
}

// EventReady provides information about objects relative to the user.
// This is used to populate the session's cache
type EventReady struct {
	Event
	Users          []*User                  `msg:"users" json:"users,omitzero"`
	Servers        []*Server                `msg:"servers" json:"servers,omitzero"`
	Channels       []*Channel               `msg:"channels" json:"channels,omitzero"`
	Members        []*ServerMember          `msg:"members" json:"members,omitzero"`
	Emojis         []*Emoji                 `msg:"emojis" json:"emojis,omitzero"`
	VoiceStates    []*ChannelVoiceState     `msg:"voice_states" json:"voice_states,omitzero"`
	UserSettings   map[string]any           `msg:"user_settings" json:"user_settings,omitzero"`
	ChannelUnreads []ChannelUnread          `msg:"channel_unreads" json:"channel_unreads,omitzero"`
	PolicyChanges  []EventReadyPolicyChange `msg:"policy_changes" json:"policy_changes,omitzero"`
}

type AuthType string

const (
	EventTypeAuthDeleteSession     AuthType = "DeleteSession"
	EventTypeAuthDeleteAllSessions AuthType = "DeleteAllSessions"
)

type EventAuth struct {
	Event
	EventType AuthType `msg:"event_type" json:"event_type,omitzero"`
	UserID    string   `msg:"user_id" json:"user_id,omitzero"`
	SessionID string   `msg:"session_id" json:"session_id,omitzero"`

	// Only present when... I forgot.
	ExcludeSessionID string `msg:"exclude_session_id" json:"exclude_session_id,omitzero"`
}

// EventAuthenticated is sent after the client has authenticated.
type EventAuthenticated struct {
	Event `msg:",flatten"`
}

type EventLogout struct {
	Event `msg:",flatten"`
}

type EventMessage struct {
	Event   `msg:",flatten"`
	Message `msg:",flatten"`
}

// EventServerUpdate is sent when a server is updated. Data will only contain fields that were modified.
type EventServerUpdate struct {
	Event `msg:",flatten"`
	ID    string        `msg:"id" json:"id,omitzero"`
	Data  PartialServer `msg:"data" json:"data,omitzero"`
	Clear []string      `msg:"clear" json:"clear,omitzero"`
}

// EventChannelUpdate is sent when a channel is updated. Data will only contain fields that were modified.
type EventChannelUpdate struct {
	Event `msg:",flatten"`
	ID    string         `msg:"id" json:"id,omitzero"`
	Data  PartialChannel `msg:"data" json:"data,omitzero"`
	Clear []string       `msg:"clear" json:"clear,omitzero"`
}

// EventServerRoleUpdate is sent when a role is updated. Data will only contain fields that were modified.
type EventServerRoleUpdate struct {
	Event  `msg:",flatten"`
	ID     string            `msg:"id" json:"id,omitzero"`
	RoleID string            `msg:"role_id" json:"role_id,omitzero"`
	Data   PartialServerRole `msg:"data" json:"data,omitzero"`
	Clear  []string          `msg:"clear" json:"clear,omitzero"`
}

// EventServerMemberUpdate is sent when a member is updated. Data will only contain fields that were modified.
type EventServerMemberUpdate struct {
	Event `msg:",flatten"`
	ID    MemberCompositeID   `msg:"id" json:"id,omitzero"`
	Data  PartialServerMember `msg:"data" json:"data,omitzero"`
	Clear []string            `msg:"clear" json:"clear,omitzero"`
}

type EventUserUpdate struct {
	Event   `msg:",flatten"`
	ID      string      `msg:"id" json:"id,omitzero"`
	Data    PartialUser `msg:"data" json:"data,omitzero"`
	Clear   []string    `msg:"clear" json:"clear,omitzero"`
	EventID *string     `msg:"event_id" json:"event_id,omitzero"`
}

type EventWebhookUpdate struct {
	Event  `msg:",flatten"`
	ID     string         `msg:"id" json:"id,omitzero"`
	Data   PartialWebhook `msg:"data" json:"data,omitzero"`
	Remove []string       `msg:"remove" json:"remove,omitzero"` // todo: why is this "remove" and not "clear"?
}

type EventMessageUpdate struct {
	Event   `msg:",flatten"`
	ID      string  `msg:"id" json:"id,omitzero"`
	Channel string  `msg:"channel" json:"channel,omitzero"`
	Data    Message `msg:"data" json:"data,omitzero"`
}

type EventMessageAppend struct {
	Event   `msg:",flatten"`
	ID      string  `msg:"id" json:"id,omitzero"`
	Channel string  `msg:"channel" json:"channel,omitzero"`
	Append  Message `msg:"append" json:"append,omitzero"`
}

type EventMessageDelete struct {
	Event   `msg:",flatten"`
	ID      string `msg:"id" json:"id,omitzero"`
	Channel string `msg:"channel" json:"channel,omitzero"`
}

type EventBulkMessageDelete struct {
	Event   `msg:",flatten"`
	Channel string   `msg:"channel" json:"channel,omitzero"`
	IDs     []string `msg:"ids" json:"ids,omitzero"`
}

// EventChannelStartTyping is sent when a user starts typing in a channel.
type EventChannelStartTyping struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitzero"`
	User  string `msg:"user" json:"user,omitzero"`
}

// EventChannelStopTyping is sent when a user stops typing in a channel.
type EventChannelStopTyping struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitzero"`
	User  string `msg:"user" json:"user,omitzero"`
}

type EventChannelAck struct {
	Event     `msg:",flatten"`
	ID        string `msg:"id" json:"id,omitzero"`
	User      string `msg:"user" json:"user,omitzero"`
	MessageID string `msg:"message_id" json:"message_id,omitzero"`
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
	ID    string `msg:"id" json:"id,omitzero"`
}

// EventServerMemberLeave is sent when a user leaves a server.
type EventServerMemberLeave struct {
	Event  `msg:",flatten"`
	ID     string `msg:"id" json:"id,omitzero"`     // Server ID
	User   string `msg:"user" json:"user,omitzero"` // User ID
	Reason string `msg:"reason" json:"reason,omitzero"`
}

// EventServerCreate is sent when a server is created (joined).
type EventServerCreate struct {
	Event       `msg:",flatten"`
	ID          string               `msg:"id" json:"id,omitzero"`
	Server      *Server              `msg:"server" json:"server,omitzero"`
	Channels    []*Channel           `msg:"channels" json:"channels,omitzero"`
	Emojis      []*Emoji             `msg:"emojis" json:"emojis,omitzero"`
	VoiceStates []*ChannelVoiceState `msg:"voice_states" json:"voice_states,omitzero"`
}

type EventServerRoleDelete struct {
	Event  `msg:",flatten"`
	ID     string `msg:"id" json:"id,omitzero"`
	RoleID string `msg:"role_id" json:"role_id,omitzero"`
}

type EventServerMemberJoin struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitzero"`
	User  string `msg:"user" json:"user,omitzero"`
}

type EventServerDelete struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitzero"`
}

type EventMessageReact struct {
	Event     `msg:",flatten"`
	ID        string `msg:"id" json:"id,omitzero"`
	ChannelID string `msg:"channel_id" json:"channel_id,omitzero"`
	UserID    string `msg:"user_id" json:"user_id,omitzero"`
	EmojiID   string `msg:"emoji_id" json:"emoji_id,omitzero"`
}

// EventMessageUnreact is sent when a user removes a singular reaction from a message.
type EventMessageUnreact struct {
	EventMessageReact `msg:",flatten"`
}

// EventMessageRemoveReaction is sent when all the reactions are removed from a message.
type EventMessageRemoveReaction struct {
	Event     `msg:",flatten"`
	ID        string `msg:"id" json:"id,omitzero"`
	ChannelID string `msg:"channel_id" json:"channel_id,omitzero"`
	EmojiID   string `msg:"emoji_id" json:"emoji_id,omitzero"`
}

type EventChannelGroupJoin struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitzero"`
	User  string `msg:"user" json:"user,omitzero"`
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
	ID    string `msg:"id" json:"id,omitzero"`
}

type EventUserRelationship struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitzero"`
	User  *User  `msg:"user" json:"user,omitzero"`
}

type EventUserPlatformWipe struct {
	Event
	UserID string `msg:"user_id" json:"user_id,omitzero"`
	Flags  uint32 `msg:"flags" json:"flags,omitzero"`
}

type EventUserSettingsUpdate struct {
	Event `msg:",flatten"`
	// Update is a tuple of (int, string); update time, and the data in JSON
	Update map[string]SyncSettingsParamsTuple `msg:"update" json:"update,omitzero"`
}

type EventWebhookCreate struct {
	Event   `msg:",flatten"`
	Webhook `msg:",flatten"`
}

type EventWebhookDelete struct {
	Event `msg:",flatten"`
	ID    string `msg:"id" json:"id,omitzero"`
}

// EventVoiceChannelJoin is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/events/client.rs#L352
type EventVoiceChannelJoin struct {
	Event `msg:",flatten"`
	// Channel.ID
	ID string `msg:"id" json:"id,omitzero"`
	// State.ID -> User.ID
	State UserVoiceState `msg:"state" json:"state,omitzero"`
}

// EventVoiceChannelLeave is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/events/client.rs#L356
type EventVoiceChannelLeave struct {
	Event `msg:",flatten"`
	// Channel.ID
	ID   string `msg:"id" json:"id,omitzero"`
	User string `msg:"user" json:"user"`
}

// EventVoiceChannelMove is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/events/client.rs#L360
// Sent instead of a leave/join pair
type EventVoiceChannelMove struct {
	Event `msg:",flatten"`
	User  string `msg:"user" json:"user,omitzero"`
	// Channel.ID
	From string `msg:"from" json:"from,omitzero"`
	// Channel.ID
	To    string         `msg:"to" json:"to,omitzero"`
	State UserVoiceState `msg:"state" json:"state,omitzero"`
}

// EventUserVoiceStateUpdate is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/events/client.rs#L366
type EventUserVoiceStateUpdate struct {
	Event `msg:",flatten"`
	// User.ID
	ID        string                `msg:"id" json:"id,omitzero"`
	ChannelID string                `msg:"channel_id" json:"channel_id,omitzero"`
	Data      PartialUserVoiceState `msg:"data" json:"data,omitzero"`
}

// EventUserMoveVoiceChannel is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/events/client.rs#L371
// Sent only to the user a moderator moved; everyone else sees EventVoiceChannelMove.
type EventUserMoveVoiceChannel struct {
	Event `msg:",flatten"`
	Node  string `msg:"node" json:"node,omitzero"`
	// Channel.ID
	From string `msg:"from" json:"from,omitzero"`
	// Channel.ID
	To string `msg:"to" json:"to,omitzero"`
	// Authenticates against Node
	Token string `msg:"token" json:"token,omitzero"`
}

type EventServerRoleRanksUpdate struct {
	Event `msg:",flatten"`
	ID    string   `msg:"id" json:"id,omitzero"`
	Ranks []string `msg:"ranks" json:"ranks,omitzero"`
}

type EventReportCreate struct {
	Event `msg:",flatten"`
	// todo: implement fields
}
