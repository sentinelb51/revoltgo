package revoltgo

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

func New(token string) *Session {
	return &Session{
		Token:             token,
		HeartbeatInterval: 15 * time.Second,
		ReconnectInterval: 5 * time.Second,
		Ratelimiter:       newRatelimiter(),
		UserAgent:         "RevoltGo/1.0.0",
		HTTP:              &http.Client{Timeout: 10 * time.Second},
	}
}

// NewWithLogin exchanges an email and password for a session token, and then creates a new session.
// You are expected to store and re-use the token for future sessions.
func NewWithLogin(data LoginData) (*Session, LoginResponse, error) {
	session := New("")
	mfa, err := session.Login(data)
	if err == nil {
		session.Token = mfa.Token
	}

	return session, mfa, err
}

// Session struct.
type Session struct {
	Token  string
	Socket net.Conn

	// HTTP client used for the REST API
	HTTP *http.Client

	// Ratelimiter for the REST API
	Ratelimiter *Ratelimiter

	// State is a central store for all data received from the API
	State *State

	// The user agent used for REST APIs
	UserAgent string

	// Indicates whether the session is connected
	Connected bool

	// Interval between sending heartbeats
	HeartbeatInterval time.Duration

	// Heartbeat counter
	heartbeatCount int

	// Interval between reconnecting, if connection fails
	ReconnectInterval time.Duration

	// Last time a ping was sent
	LastHeartbeatSent time.Time

	// Last time a ping was received
	LastHeartbeatAck time.Time

	/* Event handlers */

	// Authentication-related handlers
	HandlersReady         []func(*Session, *EventReady)
	HandlersAuth          []func(*Session, *EventAuth)
	HandlersPong          []func(*Session, *EventPong)
	HandlersAuthenticated []func(*Session, *EventAuthenticated)

	// User-related handlers
	HandlersUserUpdate         []func(*Session, *EventUserUpdate)
	HandlersUserSettingsUpdate []func(*Session, *EventUserSettingsUpdate)
	HandlersUserRelationship   []func(*Session, *EventUserRelationship)
	HandlersUserPlatformWipe   []func(*Session, *EventUserPlatformWipe)

	// Message-related handlers
	HandlersMessage        []func(*Session, *EventMessage)
	HandlersMessageAppend  []func(*Session, *EventMessageAppend)
	HandlersMessageUpdate  []func(*Session, *EventMessageUpdate)
	HandlersMessageDelete  []func(*Session, *EventMessageDelete)
	HandlersMessageReact   []func(*Session, *EventMessageReact)
	HandlersMessageUnreact []func(*Session, *EventMessageUnreact)

	// Channel-related handlers
	HandlersChannelCreate      []func(*Session, *EventChannelCreate)
	HandlersChannelUpdate      []func(*Session, *EventChannelUpdate)
	HandlersChannelDelete      []func(*Session, *EventChannelDelete)
	HandlersChannelStartTyping []func(*Session, *EventChannelStartTyping)
	HandlersChannelStopTyping  []func(*Session, *EventChannelStopTyping)
	HandlersChannelAck         []func(*Session, *EventChannelAck)

	// Group-related handlers
	HandlersGroupJoin  []func(*Session, *EventGroupJoin)
	HandlersGroupLeave []func(*Session, *EventGroupLeave)

	// Server-related handlers
	HandlersServerCreate []func(*Session, *EventServerCreate)
	HandlersServerUpdate []func(*Session, *EventServerUpdate)
	HandlersServerDelete []func(*Session, *EventServerDelete)

	// ServerRole-related handlers
	HandlersServerRoleUpdate []func(*Session, *EventServerRoleUpdate)
	HandlersServerRoleDelete []func(*Session, *EventServerRoleDelete)

	// ServerMember-related handlers
	HandlersServerMemberUpdate []func(*Session, *EventServerMemberUpdate)
	HandlersServerMemberJoin   []func(*Session, *EventServerMemberJoin)
	HandlersServerMemberLeave  []func(*Session, *EventServerMemberLeave)

	// Emoji-related handlers
	HandlersEmojiCreate []func(*Session, *EventEmojiCreate)
	HandlersEmojiDelete []func(*Session, *EventEmojiDelete)

	// Unknown event handler. Useful for debugging purposes
	HandlersUnknown []func(session *Session, message string)
}

// WriteSocket writes data to the websocket in JSON
func (s *Session) WriteSocket(data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return wsutil.WriteClientText(s.Socket, payload)
}

func (s *Session) Emoji(eID string) (emoji *Emoji, err error) {
	endpoint := EndpointEmoji(eID)
	err = s.request(http.MethodGet, endpoint, nil, &emoji)
	return
}

func (s *Session) EmojiCreate(eID string, data EmojiCreateData) (emoji *Emoji, err error) {
	endpoint := EndpointEmoji(eID)
	err = s.request(http.MethodPut, endpoint, data, &emoji)
	return
}

func (s *Session) EmojiDelete(eID string) error {
	endpoint := EndpointEmoji(eID)
	return s.request(http.MethodDelete, endpoint, nil, nil)
}

// Channel fetches a channel using an API call
func (s *Session) Channel(cID string) (channel *Channel, err error) {
	endpoint := EndpointChannels(cID)
	err = s.request(http.MethodGet, endpoint, nil, &channel)
	return
}

// User fetches a user by their ID
// To fetch self, supply "@me" as the ID
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

func (s *Session) ChannelWebhooks(cID string) (webhooks []*Webhook, err error) {
	endpoint := EndpointChannelsWebhooks(cID)
	err = s.request(http.MethodGet, endpoint, nil, &webhooks)
	return
}

func (s *Session) ChannelWebhooksCreate(cID string, data WebhookCreate) (webhook *Webhook, err error) {
	endpoint := EndpointChannelsWebhooks(cID)
	err = s.request(http.MethodPost, endpoint, data, &webhook)
	return
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
	endpoint := EndpointChannelsInvites(cID)
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
	endpoint := EndpointChannelsMessagesMessage(cID, mID)
	err = s.request(http.MethodGet, endpoint, nil, &message)
	return
}

// ChannelMessageReactionCreate adds a reaction (emoji ID) to a message
func (s *Session) ChannelMessageReactionCreate(cID, mID, eID string) (err error) {
	endpoint := EndpointChannelsMessageReaction(cID, mID, eID)
	err = s.request(http.MethodPut, endpoint, nil, nil)
	return
}

// ChannelMessageReactionDelete deletes a singular reaction on a message
func (s *Session) ChannelMessageReactionDelete(cID, mID, eID string) (err error) {
	endpoint := EndpointChannelsMessageReaction(cID, mID, eID)
	err = s.request(http.MethodDelete, endpoint, nil, nil)
	return
}

// ChannelMessageReactionClear clears all reactions on a message
func (s *Session) ChannelMessageReactionClear(cID, mID string) (err error) {
	endpoint := EndpointChannelsMessageReactions(cID, mID)
	err = s.request(http.MethodDelete, endpoint, nil, nil)
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
	endpoint := EndpointChannelsMessages(cID)

	if len(params) > 0 {
		endpoint += params[0].Encode()
	}

	err = s.request(http.MethodGet, endpoint, nil, &messages)
	return
}

func (s *Session) ChannelMessageEdit(cID, mID string, data MessageEditData) (message *Message, err error) {
	endpoint := EndpointChannelsMessagesMessage(cID, mID)
	err = s.request(http.MethodPatch, endpoint, data, &message)
	return
}

func (s *Session) ChannelMessageSend(cID string, data MessageSend) (message *Message, err error) {
	endpoint := EndpointChannelsMessages(cID)
	err = s.request(http.MethodPost, endpoint, data, &message)
	return
}

func (s *Session) ChannelMessageDelete(cID, mID string) error {
	endpoint := EndpointChannelsMessagesMessage(cID, mID)
	return s.request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) AccountCreate(data AccountCreateData) error {
	endpoint := EndpointAuthAccount("create")
	return s.request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) AccountReverify(data AccountReverifyData) error {
	endpoint := EndpointAuthAccount("reverify")
	return s.request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) AccountDeleteConfirm(data AccountDeleteConfirmData) error {
	endpoint := EndpointAuthAccount("delete")
	return s.request(http.MethodPut, endpoint, data, nil)
}

func (s *Session) AccountDelete() error {
	endpoint := EndpointAuthAccount("delete")
	return s.request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) Account() (account *Account, err error) {
	endpoint := EndpointAuthAccount("")
	err = s.request(http.MethodGet, endpoint, nil, &account)
	return
}

func (s *Session) AccountDisable() error {
	endpoint := EndpointAuthAccount("disable")
	return s.request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) AccountChangePassword(data AccountChangePasswordData) error {
	endpoint := EndpointAuthAccountChange("password")
	return s.request(http.MethodPatch, endpoint, data, nil)
}

func (s *Session) AccountChangeEmail(data AccountChangeEmailData) error {
	endpoint := EndpointAuthAccountChange("email")
	return s.request(http.MethodPatch, endpoint, data, nil)
}

func (s *Session) VerifyEmail(code string) (ticket ChangeEmail, err error) {
	endpoint := EndpointAuthAccountVerify(code)
	err = s.request(http.MethodPost, endpoint, nil, &ticket)
	return
}

// PasswordReset requests a password reset, which is sent to the email provided
func (s *Session) PasswordReset(data AccountReverifyData) error {
	endpoint := EndpointAuthAccount("reset_password")
	return s.request(http.MethodPost, endpoint, data, nil)
}

// PasswordResetConfirm confirms a password reset
func (s *Session) PasswordResetConfirm(data PasswordResetConfirmData) error {
	endpoint := EndpointAuthAccount("reset_password")
	return s.request(http.MethodPatch, endpoint, data, nil)
}

// Login as a regular user instead of bot. Friendly name is used to identify the session via MFA
func (s *Session) Login(data LoginData) (mfa LoginResponse, err error) {
	endpoint := EndpointAuthSession("login")
	err = s.request(http.MethodPost, endpoint, data, &mfa)
	return
}

func (s *Session) Sessions() (sessions []*Sessions, err error) {
	endpoint := EndpointAuthSession("all")
	err = s.request(http.MethodGet, endpoint, nil, &sessions)
	return
}

func (s *Session) SessionEdit(id string, data SessionEditData) (session SessionEditData, err error) {
	endpoint := EndpointAuthSession(id)
	err = s.request(http.MethodPatch, endpoint, data, &session)
	return
}

// Onboarding returns whether the current account requires onboarding or whether you can continue to send requests as usual
func (s *Session) Onboarding() (onboarding Onboarding, err error) {
	endpoint := EndpointOnboard("hello")
	err = s.request(http.MethodGet, endpoint, nil, &onboarding)
	return
}

// OnboardingComplete sets a new username, completes onboarding and allows a user to start using Revolt.
func (s *Session) OnboardingComplete(data OnboardingCompleteData) error {
	endpoint := EndpointOnboard("complete")
	return s.request(http.MethodPost, endpoint, data, nil)
}

// SessionsDelete invalidates a session with the provided ID
func (s *Session) SessionsDelete(id string) error {
	endpoint := EndpointAuthSession(id)
	return s.request(http.MethodDelete, endpoint, nil, nil)
}

// SessionsDeleteAll invalidates all sessions, including this one if revokeSelf is true
func (s *Session) SessionsDeleteAll(revokeSelf bool) error {
	endpoint := EndpointAuthSession("all")
	if revokeSelf {
		values := url.Values{}
		values.Set("revoke_self", "true")
		endpoint += fmt.Sprintf("?%s", values.Encode())
	}

	return s.request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) Logout() error {
	endpoint := EndpointAuthSession("logout")
	return s.request(http.MethodPost, endpoint, nil, nil)
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
