package revoltgo

import (
	"encoding/json"
	"log"
	"time"
)

type EventType string

const (
	EventTypeReady         EventType = "Ready"
	EventTypePong          EventType = "Pong"
	EventTypeAuthenticated EventType = "Authenticated"

	EventTypeMessage        EventType = "Message"
	EventTypeMessageUpdate  EventType = "MessageUpdate"
	EventTypeMessageDelete  EventType = "MessageDelete"
	EventTypeMessageReact   EventType = "MessageReact"
	EventTypeMessageUnreact EventType = "MessageUnreact"

	EventTypeChannelCreate EventType = "ChannelCreate"
	EventTypeChannelUpdate EventType = "ChannelUpdate"
	EventTypeChannelDelete EventType = "ChannelDelete"

	EventTypeGroupCreate        EventType = "GroupCreate"
	EventTypeGroupMemberAdded   EventType = "GroupMemberAdded"
	EventTypeGroupMemberRemoved EventType = "GroupMemberRemoved"

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
)

func (et EventType) String() string {
	return string(et)
}

func (et EventType) Unmarshal(data []byte) (result any) {

	switch et {
	case EventTypeReady:
		result = new(EventReady)
	case EventTypeAuthenticated:
		// maybe: optimise by skipping unmarshal, cast instead?
		result = new(EventAuthenticated)
	case EventTypePong:
		result = new(EventPong)

	case EventTypeMessage:
		result = new(EventMessage)
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

	default:
		log.Printf("unknown event type: %s", et)
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
	Users    []*User          `json:"users"`
	Servers  []*Server        `json:"servers"`
	Channels []*ServerChannel `json:"channels"`
	Members  []*Member        `json:"members"`
	Emojis   []*Emoji         `json:"emojis"`
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

type EventMessageDelete struct {
	Event
	ID      string `json:"id"`
	Channel string `json:"channel"`
}

// EventChannelStartTyping is sent when a user starts typing in a channel.
type EventChannelStartTyping struct {
	Event
	ID   string `json:"id"`
	User string `json:"user"`
}

// EventChannelStopTyping is sent when a user stops typing in a channel.
type EventChannelStopTyping struct {
	EventChannelStartTyping
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
	ID    string `json:"id"`
	Data  Server `json:"data"`
	Clear any    `json:"clear"` // TODO: what is this?
}

// EventChannelUpdate is sent when a channel is updated. Data will only contain fields that were modified.
type EventChannelUpdate struct {
	Event
	ID    string        `json:"id"`
	Data  ServerChannel `json:"data"`
	Clear any           `json:"clear"` // TODO: what is this?
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
	ID     string     `json:"id"`
	RoleID string     `json:"role_id"`
	Data   ServerRole `json:"data"`
	Clear  any        `json:"clear"` // TODO: what is this?
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
	ID    string  `json:"id"`
	Data  *Member `json:"data"`
	Clear any     `json:"clear"` // TODO: what is this?
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

type EventChannelGroupJoin struct {
	Event
	ID   string `json:"id"`
	User string `json:"user"`
}

type EventChannelGroupLeave struct {
	EventChannelGroupJoin
}
