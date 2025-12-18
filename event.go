package revoltgo

import (
	"log"

	"github.com/goccy/go-json"
)

type Event struct {
	Type string `json:"type"`
}

type AbstractEventUpdateID struct {
	StringID string
	MemberID MemberCompositeID
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

	// ID can either be a simple string or a MemberCompositeID.
	ID AbstractEventUpdateID `json:"id"`

	// RoleID is only present in ServerRoleUpdate events.
	RoleID string `json:"role_id"`

	// The updated data for a specific event
	Data json.RawMessage `json:"data"`

	// Clear is a list of keys to clear from the cache.
	Clear []string `json:"clear"`

	// Remove is a list of keys to remove from the cache.
	// Ideally the API should just stick to "clear"...
	Remove []string `json:"remove"`
}

// standardise is used to migrate "remove" fields to "clear" if they exist.
func (aeu *AbstractEventUpdate) standardise() {

	if len(aeu.Remove) == 0 {
		return
	}

	if len(aeu.Clear) > 0 {
		log.Printf("AbstractEventUpdate has both 'clear' and 'remove' fields?")
		return
	}

	aeu.Clear = aeu.Remove
}

func (aeu *AbstractEventUpdate) EventServerUpdate() *EventServerUpdate {
	// Re-populate normal data
	event := EventServerUpdate{
		Event: aeu.Event,
		ID:    aeu.ID.StringID,
		Clear: aeu.Clear,
	}

	// Decode variable data into a struct
	err := json.Unmarshal(aeu.Data, &event.Data)
	if err != nil {
		log.Printf("Error unmarshalling %T from abstract data: %s\n", event, err)
	}

	return &event
}

func (aeu *AbstractEventUpdate) EventChannelUpdate() *EventChannelUpdate {
	// Re-populate normal data
	event := EventChannelUpdate{
		Event: aeu.Event,
		ID:    aeu.ID.StringID,
		Clear: aeu.Clear,
	}

	// Decode variable data into a struct
	err := json.Unmarshal(aeu.Data, &event.Data)
	if err != nil {
		log.Printf("Error unmarshalling %T from abstract data: %s\n", event, err)
	}

	return &event
}

func (aeu *AbstractEventUpdate) EventServerRoleUpdate() *EventServerRoleUpdate {
	// Re-populate normal data
	event := EventServerRoleUpdate{
		Event:  aeu.Event,
		ID:     aeu.ID.StringID,
		RoleID: aeu.RoleID,
		Clear:  aeu.Clear,
	}

	// Decode variable data into a struct
	err := json.Unmarshal(aeu.Data, &event.Data)
	if err != nil {
		log.Printf("Error unmarshalling %T from abstract data: %s\n", event, err)
	}

	return &event
}

func (aeu *AbstractEventUpdate) EventServerMemberUpdate() *EventServerMemberUpdate {
	// Re-populate normal data
	event := EventServerMemberUpdate{
		Event: aeu.Event,
		ID:    aeu.ID.MemberID,
		Clear: aeu.Clear,
	}

	// Decode variable data into a struct
	err := json.Unmarshal(aeu.Data, &event.Data)
	if err != nil {
		log.Printf("Error unmarshalling %T from abstract data: %s\n", event, err)
	}

	return &event
}

func (aeu *AbstractEventUpdate) EventUserUpdate() *EventUserUpdate {
	// Re-populate normal data
	event := EventUserUpdate{
		Event: aeu.Event,
		ID:    aeu.ID.StringID,
		Clear: aeu.Clear,
	}

	// Decode variable data into a struct
	err := json.Unmarshal(aeu.Data, &event.Data)
	if err != nil {
		log.Printf("Error unmarshalling %T from abstract data: %s\n", event, err)
	}

	return &event
}

func (aeu *AbstractEventUpdate) EventWebhookUpdate() *EventWebhookUpdate {
	// Re-populate normal data
	event := EventWebhookUpdate{
		Event:  aeu.Event,
		ID:     aeu.ID.StringID,
		Remove: aeu.Clear,
	}

	// Decode variable data into a struct
	err := json.Unmarshal(aeu.Data, &event.Data)
	if err != nil {
		log.Printf("Error unmarshalling %T from abstract data: %s\n", event, err)
	}

	return &event
}

func (aeu *AbstractEventUpdate) EventMessageUpdate() *EventMessageUpdate {
	// Re-populate normal data
	event := EventMessageUpdate{
		Event: aeu.Event,
		ID:    aeu.ID.StringID,
	}

	// Decode variable data into a struct
	err := json.Unmarshal(aeu.Data, &event.Data)
	if err != nil {
		log.Printf("Error unmarshalling %T from abstract data: %s\n", event, err)
	}

	return &event
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
	Error EventErrorType `json:"error"`
}

type EventBulk struct {
	Event
	V []json.RawMessage `json:"v"`
}

type EventPong struct {
	Event
	Data int64 `json:"data"`
}

// EventReady provides information about objects relative to the user.
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

// EventServerRoleUpdate is sent when a role is updated. Data will only contain fields that were modified.
type EventServerRoleUpdate struct {
	Event
	ID     string      `json:"id"`
	RoleID string      `json:"role_id"`
	Data   *ServerRole `json:"data"`
	Clear  []string    `json:"clear"`
}

// EventServerMemberUpdate is sent when a member is updated. Data will only contain fields that were modified.
type EventServerMemberUpdate struct {
	Event
	ID    MemberCompositeID `json:"id"`
	Data  *ServerMember     `json:"data"`
	Clear []string          `json:"clear"`
}

type EventUserUpdate struct {
	Event
	ID    string   `json:"id"`
	Data  *User    `json:"data"`
	Clear []string `json:"clear"`
}

type EventWebhookUpdate struct {
	Event
	ID     string   `json:"id"`
	Data   *Webhook `json:"data"`
	Remove []string `json:"remove"`
}

type EventMessageUpdate struct {
	Event
	ID      string  `json:"id"`
	Channel string  `json:"channel"`
	Data    Message `json:"data"`
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

type EventBulkMessageDelete struct {
	Event
	Channel string   `json:"channel"`
	IDs     []string `json:"ids"`
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
	*Channel
}

// EventChannelDelete is sent when a channel is deleted.
type EventChannelDelete struct {
	Event
	ID string `json:"id"`
}

// EventServerMemberLeave is sent when a user leaves a server.
type EventServerMemberLeave struct {
	Event
	ID     string `json:"id"`
	User   string `json:"user"`
	Reason string `json:"reason"`
}

// EventServerCreate is sent when a server is created (joined).
type EventServerCreate struct {
	Event
	ID       string     `json:"id"`
	Server   *Server    `json:"server"`
	Channels []*Channel `json:"channels"`
	Emojis   []*Emoji   `json:"emojis"`
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

// EventMessageUnreact is sent when a user removes a singular reaction from a message.
type EventMessageUnreact struct {
	EventMessageReact
}

// EventMessageRemoveReaction is sent when all the reactions are removed from a message.
type EventMessageRemoveReaction struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	EmojiID   string `json:"emoji_id"`
}

type EventChannelGroupJoin struct {
	Event
	ID   string `json:"id"`
	User string `json:"user"`
}

type EventChannelGroupLeave struct {
	EventChannelGroupJoin
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
	Update map[string]UpdateTuple `json:"update"`
}

type EventWebhookCreate struct {
	Event
	*Webhook
}

type EventWebhookDelete struct {
	Event
	ID string `json:"id"`
}
