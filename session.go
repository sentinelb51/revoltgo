package revoltgo

import (
	"encoding/json"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/http"
	"time"
)

func New(token string) *Session {
	return &Session{
		Token:             token,
		HeartbeatInterval: 30 * time.Second,
		UserAgent:         "RevoltGo/pre-release",
		HTTP:              &http.Client{Timeout: 10 * time.Second},
	}
}

// Session struct.
type Session struct {
	SelfBot *SelfBot
	Token   string
	Socket  net.Conn
	HTTP    *http.Client
	State   *State

	// The user agent used for REST APIs
	UserAgent string

	// Indicates whether the session is connected (received Authenticated event)
	Connected bool

	// Interval between sending heartbeats
	HeartbeatInterval time.Duration

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

	// ServerMember-related handlers
	OnServerMemberUpdateHandlers []func(*Session, *EventServerMemberUpdate)
	OnServerMemberJoinHandlers   []func(*Session, *EventServerMemberJoin)
	OnServerMemberLeaveHandlers  []func(*Session, *EventServerMemberLeave)

	// Unknown event handler. Useful for debugging purposes
	OnUnknownEventHandlers []func(session *Session, message string)
}

//todo: remove SelfBot and make Login() function accept such parameters

type SelfBot struct {
	ID           string `json:"id"`
	UID          string `json:"uid"`
	SessionToken string `json:"token"`
}

// WriteSocket writes data to the websocket in JSON
func (s *Session) WriteSocket(data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return wsutil.WriteClientText(s.Socket, payload)
}

// Channel fetches a channel using an API call
func (s *Session) Channel(cID string) (channel *Channel, err error) {
	endpoint := EndpointChannels(cID)
	err = s.request(http.MethodGet, endpoint, nil, &channel)
	return
}

// User fetches a user by their ID
// To fetch self, supply "me" as the ID
func (s *Session) User(uID string) (user *User, err error) {
	endpoint := EndpointUsers(uID)
	err = s.request(http.MethodGet, endpoint, nil, &user)
	return
}

func (s *Session) UserBlock(uID string) (user *User, err error) {
	endpoint := EndpointUsersBlock(uID)
	err = s.request(http.MethodPut, endpoint, nil, user)
	return
}

func (s *Session) UserUnblock(uID string) (user *User, err error) {
	endpoint := EndpointUsersBlock(uID)
	err = s.request(http.MethodDelete, endpoint, nil, user)
	return
}

func (s *Session) UserProfile(uID string) (profile *UserProfile, err error) {
	endpoint := EndpointUsersProfile(uID)
	err = s.request(http.MethodGet, endpoint, nil, &profile)
	return
}

func (s *Session) UserDefaultAvatar(uID string) (binary []byte, err error) {
	endpoint := EndpointUsersDefaultAvatar(uID)
	err = s.request(http.MethodGet, endpoint, nil, &binary)
	return
}

func (s *Session) SetUsername(data UsernameData) (user *User, err error) {
	err = s.request(http.MethodPatch, URLUsersUsername, data, &user)
	return
}

func (s *Session) UserFlags(uID string) (flags int, err error) {
	endpoint := EndpointUsersFlags(uID)
	err = s.request(http.MethodGet, endpoint, nil, &flags)
	return
}

func (s *Session) UserEdit(uID string, data UserEditData) (user *User, err error) {
	endpoint := EndpointUsers(uID)
	err = s.request(http.MethodPatch, endpoint, data, &user)
	return
}

// Server fetches a server by its ID
func (s *Session) Server(id string) (server *Server, err error) {
	endpoint := EndpointServers(id)
	err = s.request(http.MethodGet, endpoint, nil, &server)
	return
}

func (s *Session) ServerEdit(id string, data ServerEditData) (server *Server, err error) {
	endpoint := EndpointServers(id)
	err = s.request(http.MethodPatch, endpoint, data, &server)
	return
}

// ServerCreate creates a server based on the data provided
func (s *Session) ServerCreate(data ServerCreateData) (server *Server, err error) {
	endpoint := EndpointServers("create")
	err = s.request(http.MethodPost, endpoint, data, &server)
	return
}

// ChannelBeginTyping is a websocket method to start typing in a channel
func (s *Session) ChannelBeginTyping(cID string) (err error) {
	data := WebsocketChannelTyping{Channel: cID, Type: WebsocketMessageTypeBeginTyping}
	return s.WriteSocket(data)
}

// ChannelEndTyping is a websocket method to stop typing in a channel
func (s *Session) ChannelEndTyping(cID string) (err error) {
	data := WebsocketChannelTyping{Channel: cID, Type: WebsocketMessageTypeEndTyping}
	return s.WriteSocket(data)
}

// GroupCreate creates a group based on the data provided
// "Users" field is a list of user IDs that will be in the group
func (s *Session) GroupCreate(data GroupCreateData) (group *Group, err error) {
	endpoint := EndpointChannels("create")
	err = s.request(http.MethodPost, endpoint, data, &group)
	return
}

func (s *Session) GroupMemberAdd(cID, mID string) (err error) {
	endpoint := EndpointChannelsRecipients(cID, mID)
	err = s.request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) GroupMemberDelete(cID, mID string) (err error) {
	endpoint := EndpointChannelsRecipients(cID, mID)
	err = s.request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) GroupMembers(cID string) (users []*User, err error) {
	endpoint := EndpointChannels(cID)
	err = s.request(http.MethodGet, endpoint, nil, &users)
	return
}

func (s *Session) ChannelInviteCreate(cID string) (invite *InviteCreate, err error) {
	endpoint := EndpointChannelInvites(cID)
	err = s.request(http.MethodPost, endpoint, nil, &invite)
	return
}

func (s *Session) ChannelDelete(cID string) (err error) {
	endpoint := EndpointChannels(cID)
	err = s.request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerAck(serverID string) (err error) {
	endpoint := EndpointServersAck(serverID)
	err = s.request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) ServerBans(sID string) (bans []*ServerBans, err error) {
	endpoint := EndpointServersBans(sID)
	err = s.request(http.MethodGet, endpoint, nil, &bans)
	return
}

func (s *Session) ServersRole(sID, rID string) (role *ServerRole, err error) {
	endpoint := EndpointServersRole(sID, rID)
	err = s.request(http.MethodGet, endpoint, nil, &role)
	return
}

func (s *Session) Invite(iID string) (invite *Invite, err error) {
	endpoint := EndpointInvites(iID)
	err = s.request(http.MethodGet, endpoint, nil, &invite)
	return
}

func (s *Session) InviteJoin(iID string) (invite *Invite, err error) {
	endpoint := EndpointInvites(iID)
	err = s.request(http.MethodPost, endpoint, nil, &invite)
	return
}

func (s *Session) InviteDelete(iID string) (err error) {
	endpoint := EndpointInvites(iID)
	err = s.request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerRoleDelete(sID, rID string) (err error) {
	endpoint := EndpointServersRole(sID, rID)
	err = s.request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServersRoleEdit(sID, rID string, data ServerRoleEditData) (role *ServerRole, err error) {
	endpoint := EndpointServersRole(sID, rID)
	err = s.request(http.MethodPatch, endpoint, data, &role)
	return
}

func (s *Session) ServersRoleCreate(sID string, data ServerRoleCreateData) (role *ServerRole, err error) {
	endpoint := EndpointServersRoles(sID)
	err = s.request(http.MethodPost, endpoint, data, &role)
	return
}

func (s *Session) PermissionsSet(sID, rID string, data PermissionAD) (err error) {
	endpoint := EndpointPermissions(sID, rID)
	err = s.request(http.MethodPut, endpoint, data, nil)
	return
}

func (s *Session) ChannelPermissionsSet(sID, cID string, data PermissionAD) (err error) {
	endpoint := EndpointChannelsPermissions(sID, cID)
	err = s.request(http.MethodPut, endpoint, data, nil)
	return
}

func (s *Session) ChannelPermissionsSetDefault(sID string, data PermissionAD) (err error) {
	endpoint := EndpointChannelsPermissions(sID, "default")
	err = s.request(http.MethodPut, endpoint, data, nil)
	return
}

// PermissionsSetDefault sets the permissions of a role in a server
func (s *Session) PermissionsSetDefault(sID string, data PermissionsSetDefaultData) (err error) {
	endpoint := EndpointPermissions(sID, "default")
	err = s.request(http.MethodPut, endpoint, data, nil)
	return
}

func (s *Session) ChannelEdit(cID string, data ChannelEditData) (channel *Channel, err error) {
	endpoint := EndpointChannels(cID)
	err = s.request(http.MethodPatch, endpoint, data, &channel)
	return
}

func (s *Session) ServerMemberUnban(sID, mID string) (err error) {
	endpoint := EndpointServersBan(sID, mID)
	err = s.request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberBan(sID, mID string) (err error) {
	endpoint := EndpointServersBan(sID, mID)
	err = s.request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberDelete(sID, mID string) (err error) {
	endpoint := EndpointServersMember(sID, mID)
	err = s.request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberEdit(sID, mID string, data ServerMemberEditData) (member *ServerMember, err error) {
	endpoint := EndpointServersMember(sID, mID)
	err = s.request(http.MethodPatch, endpoint, data, &member)
	return
}

func (s *Session) ServerMember(sID string, mID string) (members []*ServerMember, err error) {
	endpoint := EndpointServersMember(sID, mID)
	err = s.request(http.MethodGet, endpoint, nil, &members)
	return
}

func (s *Session) ServerMembers(sID string) (members []*ServerMembers, err error) {
	endpoint := EndpointServersMembers(sID)
	err = s.request(http.MethodGet, endpoint, nil, &members)
	return
}

func (s *Session) ChannelMessage(cID, mID string) (message *Message, err error) {
	endpoint := EndpointChannelMessagesMessage(cID, mID)
	err = s.request(http.MethodGet, endpoint, nil, &message)
	return
}

func (s *Session) ChannelCreate(sID string, data ChannelCreateData) (channel *Channel, err error) {
	endpoint := EndpointServersChannels(sID)
	err = s.request(http.MethodPost, endpoint, data, &channel)
	return
}

func (s *Session) ServerDelete(sID string) error {
	endpoint := EndpointServers(sID)
	return s.request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) ChannelMessages(cID string, params ...ChannelMessagesParams) (messages []*Message, err error) {
	endpoint := EndpointChannelMessages(cID)

	if len(params) > 0 {
		endpoint += params[0].Encode()
	}

	err = s.request(http.MethodGet, endpoint, nil, &messages)
	return
}

func (s *Session) ChannelMessageEdit(cID, mID string, data MessageEditData) (message *Message, err error) {
	endpoint := EndpointChannelMessagesMessage(cID, mID)
	err = s.request(http.MethodPatch, endpoint, data, &message)
	return
}

func (s *Session) ChannelMessageSend(cID string, ms MessageSend) (message *Message, err error) {
	endpoint := EndpointChannelMessages(cID)
	err = s.request(http.MethodPost, endpoint, ms, &message)
	return
}

func (s *Session) ChannelMessageDelete(cID, mID string) error {
	endpoint := EndpointChannelMessagesMessage(cID, mID)
	return s.request(http.MethodDelete, endpoint, nil, nil)
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

	endpoint := URLAuthSessionsLogin
	err = s.request(http.MethodPost, endpoint, payload, &sb)
	return
}

func (s *Session) UserMutual(uID string) (mutual []*UserMutual, err error) {
	endpoint := EndpointUsersMutual(uID)
	err = s.request(http.MethodGet, endpoint, nil, &mutual)
	return
}

// DirectMessages returns a list of direct message channels.
func (s *Session) DirectMessages() (channels []*Channel, err error) {
	endpoint := EndpointUsers("dms")
	err = s.request(http.MethodGet, endpoint, nil, &channels)
	return
}

func (s *Session) DirectMessageCreate(uID string) (channel *Channel, err error) {
	endpoint := EndpointUsersDM(uID)
	err = s.request(http.MethodPost, endpoint, nil, &channel)
	return
}

// Relationships returns a list of relationships for the current user
func (s *Session) Relationships() (relationships []*UserRelations, err error) {
	endpoint := URLUsersRelationships
	err = s.request(http.MethodGet, endpoint, nil, &relationships)
	return
}

// FriendAdd sends or accepts a friend request.
func (s *Session) FriendAdd(username string) (user *User, err error) {
	endpoint := EndpointUsersFriend(username)
	err = s.request(http.MethodPut, endpoint, nil, &user)
	return
}

// FriendDelete removes a friend or declines a friend request.
func (s *Session) FriendDelete(username string) (user *User, err error) {
	endpoint := EndpointUsersFriend(username)
	err = s.request(http.MethodDelete, endpoint, nil, &user)
	return
}

// Bot fetches details of a bot you own by its ID
func (s *Session) Bot(bID string) (bot *FetchedBot, err error) {
	endpoint := EndpointBots(bID)
	err = s.request(http.MethodGet, endpoint, nil, &bot)
	return
}

// Bots returns a list of bots for the current user
func (s *Session) Bots() (bots *FetchedBots, err error) {
	endpoint := EndpointBots("@me")
	err = s.request(http.MethodGet, endpoint, nil, &bots)
	return
}

// BotCreate creates a bot based on the data provided
func (s *Session) BotCreate(data BotCreateData) (bot *Bot, err error) {
	endpoint := EndpointBots("create")
	err = s.request(http.MethodPost, endpoint, data, &bot)
	return
}

func (s *Session) BotEdit(id string, data BotEditData) (bot *Bot, err error) {
	endpoint := EndpointBots(id)
	err = s.request(http.MethodPatch, endpoint, data, &bot)
	return
}

func (s *Session) BotDelete(bID string) error {
	endpoint := EndpointBots(bID)
	return s.request(http.MethodDelete, endpoint, nil, nil)
}

// BotPublic fetches a public bot by its ID
func (s *Session) BotPublic(bID string) (bot *PublicBot, err error) {
	endpoint := EndpointBotsInvite(bID)
	err = s.request(http.MethodGet, endpoint, nil, &bot)
	return
}

// BotInvite invites a bot by its ID to a server or group
func (s *Session) BotInvite(bID string, data BotInviteData) (err error) {
	endpoint := EndpointBotsInvite(bID)
	err = s.request(http.MethodPost, endpoint, data, nil)
	return
}
