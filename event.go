package revoltgo

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type EventType string

const (
	EventTypeReady         EventType = "Ready"
	EventTypePong          EventType = "Pong"
	EventTypeAuth          EventType = "Auth"
	EventTypeAuthenticated EventType = "Authenticated"

	EventTypeMessage        EventType = "Message"
	EventTypeMessageAppend  EventType = "MessageAppend"
	EventTypeMessageUpdate  EventType = "MessageUpdate"
	EventTypeMessageDelete  EventType = "MessageDelete"
	EventTypeMessageReact   EventType = "MessageReact"
	EventTypeMessageUnreact EventType = "MessageUnreact"

	EventTypeChannelCreate EventType = "ChannelCreate"
	EventTypeChannelUpdate EventType = "ChannelUpdate"
	EventTypeChannelDelete EventType = "ChannelDelete"

	EventTypeGroupJoin  EventType = "GroupCreate"
	EventTypeGroupLeave EventType = "GroupLeave"
	EventTypeChannelAck EventType = "ChannelAck"

	EventTypeChannelStartTyping EventType = "ChannelStartTyping"
	EventTypeChannelStopTyping  EventType = "ChannelStopTyping"

	EventTypeServerCreate EventType = "ServerCreate"
	EventTypeServerUpdate EventType = "ServerUpdate"
	EventTypeServerDelete EventType = "ServerDelete"

	EventTypeServerRoleDelete EventType = "ServerRoleDelete"
	EventTypeServerRoleUpdate EventType = "ServerRoleUpdate"

	EventTypeServerMemberUpdate EventType = "ServerMemberUpdate"
	EventTypeServerMemberJoin   EventType = "ServerMemberJoin"
	EventTypeServerMemberLeave  EventType = "ServerMemberLeave"

	EventTypeEmojiCreate EventType = "EmojiCreate"
	EventTypeEmojiDelete EventType = "EmojiDelete"

	EventTypeUserUpdate EventType = "UserUpdate"
)

func (et EventType) String() string {
	return string(et)
}

func (et EventType) Unmarshal(data []byte) (result any) {

	switch et {
	case EventTypeAuth:
		result = new(EventAuth)
	case EventTypeReady:
		result = new(EventReady)
	case EventTypeAuthenticated:
		return &EventAuthenticated{Event: Event{Type: et}}
	case EventTypePong:
		result = new(EventPong)

	case EventTypeMessage:
		result = new(EventMessage)
	case EventTypeMessageAppend:
		result = new(EventMessageAppend)
	case EventTypeMessageUpdate:
		result = new(EventMessageUpdate)
	case EventTypeMessageDelete:
		result = new(EventMessageDelete)
	case EventTypeMessageReact:
		result = new(EventMessageReact)
	case EventTypeMessageUnreact:
		result = new(EventMessageUnreact)

	case EventTypeChannelCreate:
		result = new(EventChannelCreate)
	case EventTypeChannelUpdate:
		result = new(EventChannelUpdate)
	case EventTypeChannelDelete:
		result = new(EventChannelDelete)
	case EventTypeChannelAck:
		result = new(EventChannelAck)

	case EventTypeServerUpdate:
		result = new(EventServerUpdate)
	case EventTypeServerCreate:
		result = new(EventServerCreate)
	case EventTypeServerDelete:
		result = new(EventServerDelete)

	case EventTypeServerRoleUpdate:
		result = new(EventServerRoleUpdate)
	case EventTypeServerRoleDelete:
		result = new(EventServerRoleDelete)

	case EventTypeServerMemberUpdate:
		result = new(EventServerMemberUpdate)
	case EventTypeServerMemberJoin:
		result = new(EventServerMemberJoin)
	case EventTypeServerMemberLeave:
		result = new(EventServerMemberLeave)

	case EventTypeChannelStartTyping:
		result = new(EventChannelStartTyping)
	case EventTypeChannelStopTyping:
		result = new(EventChannelStopTyping)

	case EventTypeEmojiCreate:
		result = new(EventEmojiCreate)
	case EventTypeEmojiDelete:
		result = new(EventEmojiDelete)

	case EventTypeUserUpdate:
		result = new(EventUserUpdate)

	default:
		panic(fmt.Errorf("unknown event type: %s", et))
	}

	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("%s: unmarshal: %s", et, err)
	}

	return
}

type Event struct {
	Type EventType `json:"type"`
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

type EventAuth struct {
	Event
	EventType EventAuthType `json:"event_type"`
	UserID    string        `json:"user_id"`
	SessionID string        `json:"session_id"`

	// Only present when
	ExcludeSessionID string `json:"exclude_session_id"`
}

type EventAuthType string

const (
	EventTypeAuthDeleteSession     EventAuthType = "DeleteSession"
	EventTypeAuthDeleteAllSessions EventAuthType = "DeleteAllSessions"
)

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

type EventMessageAppend struct {
	ID      string  `json:"id"`
	Channel string  `json:"channel"`
	Append  Message `json:"append"`
}

type EventMessageUpdateData struct {
	Content string         `json:"content"`
	Edited  time.Time      `json:"edited"`
	Embeds  []MessageEmbed `json:"embeds"`
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
