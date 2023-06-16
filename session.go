package revoltgo

import (
	"net/http"
	"time"

	"github.com/sacOO7/gowebsocket"
)

// Session struct.
type Session struct {
	SelfBot *SelfBot
	Token   string
	Socket  gowebsocket.Socket
	HTTP    *http.Client
	State   *State

	// The user agent used for REST APIs
	UserAgent string

	// Indicates whether the session is connected (received Authenticated event)
	Connected bool

	// Last time a ping was sent
	LastHeartbeatSent time.Time

	// Last time a ping was received
	LastHeartbeatAck time.Time

	/* Event handlers */

	// Authentication-related handlers
	OnReadyHandlers         []func(*Session, *EventReady)
	OnPongHandlers          []func(*Session, *EventPong)
	OnAuthenticatedHandlers []func(*Session, *EventAuthenticated)

	// Message-related handlers
	OnMessageHandlers        []func(*Session, *EventMessage)
	OnMessageUpdateHandlers  []func(*Session, *EventMessageUpdate)
	OnMessageDeleteHandlers  []func(*Session, *EventMessageDelete)
	OnMessageReactHandlers   []func(*Session, *EventMessageReact)
	OnMessageUnreactHandlers []func(*Session, *EventMessageUnreact)

	// Channel-related handlers
	OnChannelCreateHandlers      []func(*Session, *EventChannelCreate)
	OnChannelUpdateHandlers      []func(*Session, *EventChannelUpdate)
	OnChannelDeleteHandlers      []func(*Session, *EventChannelDelete)
	OnChannelStartTypingHandlers []func(*Session, *EventChannelStartTyping)
	OnChannelStopTypingHandlers  []func(*Session, *EventChannelStopTyping)

	// Group-related handlers
	OnChannelGroupJoinHandlers  []func(*Session, *EventChannelGroupJoin)
	OnChannelGroupLeaveHandlers []func(*Session, *EventChannelGroupLeave)

	// Server-related handlers
	OnServerCreateHandlers []func(*Session, *EventServerCreate)
	OnServerUpdateHandlers []func(*Session, *EventServerUpdate)
	OnServerDeleteHandlers []func(*Session, *EventServerDelete)

	// Member-related handlers
	OnServerMemberUpdateHandlers []func(*Session, *EventServerMemberUpdate)
	OnServerMemberJoinHandlers   []func(*Session, *EventServerMemberJoin)
	OnServerMemberLeaveHandlers  []func(*Session, *EventServerMemberLeave)

	// Unknown event handler. Useful for debugging purposes
	OnUnknownEventHandlers []func(session *Session, message string)
}

type SelfBot struct {
	ID           string `json:"id"`
	UserID       string `json:"uid"`
	SessionToken string `json:"token"`
}

// OnReady accepts a function that handles EventReady.
func (s *Session) OnReady(fn func(session *Session, ready *EventReady)) {
	s.OnReadyHandlers = append(s.OnReadyHandlers, fn)
}

// OnMessage accepts a function that handles EventMessage.
func (s *Session) OnMessage(fn func(*Session, *EventMessage)) {
	s.OnMessageHandlers = append(s.OnMessageHandlers, fn)
}

// OnMessageUpdate accepts a function that handles EventMessageUpdate.
func (s *Session) OnMessageUpdate(fn func(*Session, *EventMessageUpdate)) {
	s.OnMessageUpdateHandlers = append(s.OnMessageUpdateHandlers, fn)
}

// OnMessageDelete accepts a function that handles EventMessageDelete.
func (s *Session) OnMessageDelete(fn func(*Session, *EventMessageDelete)) {
	s.OnMessageDeleteHandlers = append(s.OnMessageDeleteHandlers, fn)
}

// OnChannelCreate accepts a function that handles EventChannelCreate.
func (s *Session) OnChannelCreate(fn func(*Session, *EventChannelCreate)) {
	s.OnChannelCreateHandlers = append(s.OnChannelCreateHandlers, fn)
}

// OnChannelUpdate accepts a function that handles EventChannelUpdate.
func (s *Session) OnChannelUpdate(fn func(*Session, *EventChannelUpdate)) {
	s.OnChannelUpdateHandlers = append(s.OnChannelUpdateHandlers, fn)
}

// OnChannelDelete accepts a function that handles EventChannelDelete.
func (s *Session) OnChannelDelete(fn func(*Session, *EventChannelDelete)) {
	s.OnChannelDeleteHandlers = append(s.OnChannelDeleteHandlers, fn)
}

// OnChannelGroupJoin accepts a function that handles EventChannelGroupJoin.
func (s *Session) OnChannelGroupJoin(fn func(*Session, *EventChannelGroupJoin)) {
	s.OnChannelGroupJoinHandlers = append(s.OnChannelGroupJoinHandlers, fn)
}

// OnChannelGroupLeave accepts a function that handles EventChannelGroupLeave.
func (s *Session) OnChannelGroupLeave(fn func(*Session, *EventChannelGroupLeave)) {
	s.OnChannelGroupLeaveHandlers = append(s.OnChannelGroupLeaveHandlers, fn)
}

// OnUnknownEvent accepts a function that handles unknown events. The messages are raw JSON strings.
func (s *Session) OnUnknownEvent(fn func(session *Session, message string)) {
	s.OnUnknownEventHandlers = append(s.OnUnknownEventHandlers, fn)
}

// OnChannelStartTyping accepts a function that handles EventChannelStartTyping.
func (s *Session) OnChannelStartTyping(fn func(*Session, *EventChannelStartTyping)) {
	s.OnChannelStartTypingHandlers = append(s.OnChannelStartTypingHandlers, fn)
}

// OnChannelStopTyping accepts a function that handles EventChannelStopTyping.
func (s *Session) OnChannelStopTyping(fn func(*Session, *EventChannelStopTyping)) {
	s.OnChannelStopTypingHandlers = append(s.OnChannelStopTypingHandlers, fn)
}

// OnServerCreate accepts a function that handles EventServerCreate.
func (s *Session) OnServerCreate(fn func(*Session, *EventServerCreate)) {
	s.OnServerCreateHandlers = append(s.OnServerCreateHandlers, fn)
}

// OnServerUpdate accepts a function that handles EventServerUpdate.
func (s *Session) OnServerUpdate(fn func(*Session, *EventServerUpdate)) {
	s.OnServerUpdateHandlers = append(s.OnServerUpdateHandlers, fn)
}

// OnServerDelete accepts a function that handles EventServerDelete.
func (s *Session) OnServerDelete(fn func(*Session, *EventServerDelete)) {
	s.OnServerDeleteHandlers = append(s.OnServerDeleteHandlers, fn)
}

// OnServerMemberUpdate accepts a function that handles EventServerMemberUpdate.
func (s *Session) OnServerMemberUpdate(fn func(*Session, *EventServerMemberUpdate)) {
	s.OnServerMemberUpdateHandlers = append(s.OnServerMemberUpdateHandlers, fn)
}

// OnServerMemberJoin accepts a function that handles EventServerMemberJoin.
func (s *Session) OnServerMemberJoin(fn func(*Session, *EventServerMemberJoin)) {
	s.OnServerMemberJoinHandlers = append(s.OnServerMemberJoinHandlers, fn)
}

// OnServerMemberLeave accepts a function that handles EventServerMemberLeave.
func (s *Session) OnServerMemberLeave(fn func(*Session, *EventServerMemberLeave)) {
	s.OnServerMemberLeaveHandlers = append(s.OnServerMemberLeaveHandlers, fn)
}

// Channel fetches a channel using an API call
func (s *Session) Channel(id string) (channel *ServerChannel, err error) {
	url := EndpointChannels(id)
	err = s.request(http.MethodGet, url, nil, &channel)
	return
}

// User fetches a user by their ID
func (s *Session) User(id string) (user *User, err error) {
	url := EndpointUsers(id)
	err = s.request(http.MethodGet, url, nil, &user)
	return
}

// Server fetches a server by its ID
func (s *Session) Server(id string) (server *Server, err error) {
	url := EndpointServers(id)
	err = s.request(http.MethodGet, url, nil, &server)
	return
}

// ServerCreate creates a server based on the data provided
func (s *Session) ServerCreate(data *ServerCreateData) (server *Server, err error) {

	if data.Nonce == "" {
		data.Nonce = ULID()
	}

	url := EndpointServers("create")
	err = s.request(http.MethodPost, url, data, &server)

	return
}

func (s *Session) ChannelMessageSend(cID string, ms *MessageSend) (message *Message, err error) {

	if ms.Nonce == "" {
		ms.Nonce = ULID()
	}

	url := EndpointChannelMessages(cID)
	err = s.request(http.MethodPost, url, ms, &message)
	return
}

func (s *Session) ChannelMessageDelete(cID, mID string) error {
	url := EndpointChannelMessagesMessage(cID, mID)
	_, err := s.handleRequest(http.MethodDelete, url, nil)
	return err
}

// Login as a regular user; this is for self-bots only
// friendlyName is an optional parameter that helps identify the session later
func (s *Session) Login(email, password, friendlyName string) (sb *SelfBot, err error) {

	if s.SelfBot == nil {
		panic("method restricted to self-bots")
	}

	payload := &LoginData{
		Email:        email,
		Password:     password,
		FriendlyName: friendlyName,
	}

	url := URLAuthSessionsLogin
	err = s.request(http.MethodPost, url, payload, &sb)
	return
}

// DirectMessages returns a list of direct message channels.
func (s *Session) DirectMessages() (channels []*ServerChannel, err error) {
	url := EndpointUsers("dms")
	err = s.request(http.MethodGet, url, nil, &channels)
	return
}

// Edit client user.
func (s *Session) Edit(eu *EditUser) error {
	url := EndpointUsers("@me")
	return s.request(http.MethodPatch, url, eu, nil)
}

// GroupCreate creates a group based on the data provided
// "Users" field is a list of users that will be in the group
func (s *Session) GroupCreate(data GroupCreateData) (channel *ServerChannel, err error) {

	if data.Nonce == "" {
		data.Nonce = ULID()
	}

	url := EndpointChannels("create")
	err = s.request(http.MethodPost, url, data, &channel)
	return
}

// Relationships returns a list of relationships for the current user
func (s *Session) Relationships() (relationships []*UserRelations, err error) {
	url := URLUsersRelationships
	err = s.request(http.MethodGet, url, nil, &relationships)
	return
}

// FriendAdd sends or accepts a friend request.
func (s *Session) FriendAdd(username string) (relations *UserRelations, err error) {
	url := EndpointUsersFriend(username)
	err = s.request(http.MethodPut, url, nil, &relations)
	return
}

// FriendDelete removes a friend or declines a friend request.
func (s *Session) FriendDelete(username string) (relations *UserRelations, err error) {
	url := EndpointUsersFriend(username)
	err = s.request(http.MethodDelete, url, nil, &relations)
	return
}

// BotCreate creates a bot based on the data provided
func (s *Session) BotCreate(data BotCreateData) (bot *Bot, err error) {
	url := EndpointBots("create")
	err = s.request(http.MethodPost, url, data, &bot)
	return
}

// Bots returns a list of bots for the current user
func (s *Session) Bots() (bots *FetchedBots, err error) {
	url := EndpointBots("@me")
	err = s.request(http.MethodGet, url, nil, &bots)
	return
}

// BotsPublic fetches a public bot by its ID
func (s *Session) BotsPublic(id string) (bot *Bot, err error) {
	url := EndpointBots(id)
	err = s.request(http.MethodGet, url, nil, &bot)
	return
}
