package revoltgo

import (
	"encoding/json"
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
	Email        string `json:"-"`
	Password     string `json:"-"`
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
	response, err := s.request(http.MethodGet, url, nil)
	if err == nil {
		err = json.Unmarshal(response, &channel)
	}

	return
}

// User fetches a user by their ID
func (s *Session) User(id string) (user *User, err error) {

	url := EndpointUsers(id)
	response, err := s.request(http.MethodGet, url, nil)
	if err == nil {
		err = json.Unmarshal(response, &user)
	}

	return
}

// Server fetches a server by its ID
func (s *Session) Server(id string) (server *Server, err error) {

	url := EndpointServers(id)
	response, err := s.request(http.MethodGet, url, nil)
	if err == nil {
		err = json.Unmarshal(response, &server)
	}

	return
}

// ServerCreate creates a server based on the data provided
func (s *Session) ServerCreate(data *ServerCreateData) (server *Server, err error) {

	if data.Nonce == "" {
		data.Nonce = ULID()
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	url := URLServersCreate
	response, err := s.request(http.MethodPost, url, payload)
	if err == nil {
		err = json.Unmarshal(response, &server)
	}

	return
}

// Login as a regular user; this is for self-bots only
func (s *Session) Login(friendlyName string) (sb *SelfBot, err error) {

	if s.SelfBot == nil {
		panic("method restricted to self-bots")
	}

	url := URLAuthSessionLogin
	response, err := s.request(http.MethodPost, url, []byte("{\"email\":\""+s.SelfBot.Email+"\",\"password\":\""+s.SelfBot.Password+"\",\"friendly_name\":\""+friendlyName+"\"}"))
	if err == nil {
		err = json.Unmarshal(response, &sb)
	}

	return
}

// Fetch all of the DMs.
func (s *Session) DirectMessages() (channels []*ServerChannel, err error) {

	url := EndpointUsers("dms")
	response, err := s.request(http.MethodGet, url, nil)
	if err == nil {
		err = json.Unmarshal(response, &channels)
	}

	return
}

// Edit client user.
func (s *Session) Edit(eu *EditUser) error {

	data, err := json.Marshal(eu)
	if err != nil {
		return err
	}

	url := EndpointUsers("@me")
	_, err = s.request(http.MethodPatch, url, data)
	return err
}

// GroupCreate creates a group based on the data provided
// "Users" field is a list of users that will be in the group
func (s *Session) GroupCreate(data *GroupCreateData) (channel *ServerChannel, err error) {

	if data.Nonce == "" {
		data.Nonce = ULID()
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	url := URLChannelsCreate
	response, err := s.request(http.MethodPost, url, payload)
	if err == nil {
		err = json.Unmarshal(response, &channel)
	}

	return
}

// Fetch relationships.
func (s *Session) Relationships() (relationships []*UserRelations, err error) {

	url := URLUsersRelationships
	response, err := s.request(http.MethodGet, url, nil)
	if err == nil {
		err = json.Unmarshal(response, &relationships)
	}

	return
}

// Send friend request. / Accept friend request.
// User relations struct only will have status. id is not defined for this function.
func (s *Session) AddFriend(username string) (*UserRelations, error) {
	relationshipData := &UserRelations{}

	response, err := s.request("PUT", "/users/"+username+"/friend", nil)

	if err != nil {
		return relationshipData, err
	}

	err = json.Unmarshal(response, relationshipData)
	return relationshipData, err
}

// Deny friend request. / Remove friend.
// User relations struct only will have status. id is not defined for this function.
func (s *Session) RemoveFriend(username string) (*UserRelations, error) {
	relationshipData := &UserRelations{}

	response, err := s.request("DELETE", "/users/"+username+"/friend", nil)

	if err != nil {
		return relationshipData, err
	}

	err = json.Unmarshal(response, relationshipData)
	return relationshipData, err
}

// Create a new bot.
func (s *Session) BotCreate(name string) (*Bot, error) {
	botData := &Bot{}
	botData.Client = s

	response, err := s.request(http.MethodPost, "/bots/create", []byte("{\"name\":\""+name+"\"}"))

	if err != nil {
		return botData, err
	}

	err = json.Unmarshal(response, botData)
	return botData, err

}

// Fetch client bots.
func (s *Session) Bots() (*FetchedBots, error) {
	bots := &FetchedBots{}

	response, err := s.request(http.MethodGet, "/bots/@me", nil)

	if err != nil {
		return bots, err
	}

	err = json.Unmarshal(response, bots)

	if err != nil {
		return bots, err
	}

	// Add client for bots.
	for _, i := range bots.Bots {
		i.Client = s
	}

	return bots, nil
}

// Fetch a bot.
func (s *Session) FetchBot(id string) (*Bot, error) {
	bot := &struct {
		Bot *Bot `json:"bot"`
	}{
		Bot: &Bot{
			Client: s,
		},
	}

	response, err := s.request(http.MethodGet, "/bots/"+id, nil)

	if err != nil {
		return bot.Bot, err
	}

	err = json.Unmarshal(response, bot)
	return bot.Bot, err
}
