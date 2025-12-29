package revoltgo

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/lxzan/gws"
)

func New(token string) *Session {
	session := &Session{
		Token:       token,
		State:       newState(),
		Ratelimiter: newRatelimiter(),
		UserAgent:   fmt.Sprintf("RevoltGo/%s (github.com/sentinelb51/revoltgo)", VERSION),
		HTTP:        &http.Client{Timeout: 10 * time.Second},
	}

	session.addDefaultHandlers()

	// There may be some validation/boundary checks in the future.

	return session
}

// NewWithLogin exchanges an email and password for a session token, and then creates a new session.
// You are expected to store and re-use the token for future sessions.
func NewWithLogin(data LoginData) (*Session, LoginResponse, error) {
	session := New("")
	mfa, err := session.Login(data)
	if err == nil {
		session.Token = mfa.Token
		session.selfbot = true
	}

	return session, mfa, err
}

// todo: NewWithLoginExpress which automatically saves the token to a file

// Session represents a connection to the Revolt API.
type Session struct {
	Token       string       // Authorisation token
	WS          *Websocket   // The Websocket connection
	HTTP        *http.Client // HTTP client used for the REST API
	Ratelimiter *Ratelimiter // Ratelimiter for the REST API
	State       *State       // State is a central store for all data received from the API
	UserAgent   string       // The user agent used for REST APIs

	/* Private fields */

	selfbot bool   // Whether the session is a user or bot
	wsURL   string // Websocket URL obtained from root node

	/* Event handlers */

	handlers map[string][]func(*Session, any)
}

// Selfbot returns whether the session is a selfbot
func (s *Session) Selfbot() bool {
	return s.selfbot
}

// addDefaultHandlers adds mission-critical handlers to keep the library working
// todo: redo this; it's not dynamic
func (s *Session) addDefaultHandlers() {

	// The Websocket's first response if authentication was unsuccessful
	AddHandler(s, func(s *Session, e *EventError) {
		log.Printf("Authentication error: %s\n", e.Error)
	})

	AddHandler(s, func(s *Session, e *EventBulk) {
		for _, event := range e.V {
			go s.WS.handle(event)
		}
	})

	AddHandler(s, func(s *Session, e *EventReady) {
		s.State.populate(e)
		s.selfbot = s.State.self != nil && s.State.self.Bot == nil
	})

	// If state is disabled, none of these handlers are required
	if s.State == nil {
		return
	}

	if s.State.TrackUsers() {
		AddHandler(s, func(s *Session, e *EventUserPlatformWipe) {
			s.State.platformWipe(e)
		})
	}

	if s.State.TrackChannels() {
		AddHandler(s, func(s *Session, e *EventChannelCreate) {
			s.State.createChannel(e)
		})

		AddHandler(s, func(s *Session, e *EventChannelDelete) {
			s.State.deleteChannel(e)
		})

		AddHandler(s, func(s *Session, e *EventChannelGroupJoin) {
			s.State.addGroupParticipant(e)
		})

		AddHandler(s, func(s *Session, e *EventChannelGroupLeave) {
			s.State.removeGroupParticipant(e)
		})
	}

	if s.State.TrackServers() {
		AddHandler(s, func(s *Session, e *EventServerCreate) {
			s.State.createServer(e)
		})

		AddHandler(s, func(s *Session, e *EventServerDelete) {
			s.State.deleteServer(e)
		})

		AddHandler(s, func(s *Session, e *EventServerRoleDelete) {
			s.State.deleteServerRole(e)
		})
	}

	if s.State.TrackMembers() {
		AddHandler(s, func(s *Session, e *EventServerMemberJoin) {
			s.State.createServerMember(e)
		})

		AddHandler(s, func(s *Session, e *EventServerMemberLeave) {
			s.State.deleteServerMember(e)
		})
	}

	if s.State.TrackEmojis() {
		AddHandler(s, func(s *Session, e *EventEmojiCreate) {
			s.State.createEmoji(e)
		})

		AddHandler(s, func(s *Session, e *EventEmojiDelete) {
			s.State.deleteEmoji(e)
		})
	}

	if s.State.TrackWebhooks() {
		AddHandler(s, func(s *Session, e *EventWebhookCreate) {
			s.State.createWebhook(e)
		})

		AddHandler(s, func(s *Session, e *EventWebhookDelete) {
			s.State.deleteWebhook(e)
		})
	}

	AddHandler(s, func(s *Session, e *EventServerUpdate) {
		s.State.updateServer(e)
	})

	AddHandler(s, func(s *Session, e *EventServerMemberUpdate) {
		s.State.updateServerMember(e)
	})

	AddHandler(s, func(s *Session, e *EventChannelUpdate) {
		s.State.updateChannel(e)
	})

	AddHandler(s, func(s *Session, e *EventUserUpdate) {
		s.State.updateUser(e)
	})

	AddHandler(s, func(s *Session, e *EventServerRoleUpdate) {
		s.State.updateServerRole(e)
	})

	AddHandler(s, func(s *Session, e *EventWebhookUpdate) {
		s.State.updateWebhook(e)
	})
}

// AddHandler registers an event handler using generics.
// It infers the event name from the handler's argument type, removing the need for a type switch.
func AddHandler[T any](s *Session, handler func(*Session, T)) {

	if s.handlers == nil {
		s.handlers = make(map[string][]func(*Session, any))
	}

	// Optimization: Get the type of T without allocating a zero value using *new(T)
	t := reflect.TypeOf((*T)(nil)).Elem()

	// Fix: Drill down through pointers to find the underlying struct.
	// reflect.Type.Name() returns an empty string for pointers (e.g., *EventMessage),
	// so we must find the struct type (EventMessage) to get the correct name.
	// todo: look if this is needed
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Strip "Event" prefix (e.g., "EventMessage" -> "Message")
	name := strings.TrimPrefix(t.Name(), "Event")

	// Safety check (assuming eventConstructors is defined elsewhere)
	if _, found := eventConstructors[name]; !found {
		log.Fatalf("attempting to bind handler for unsupported event type: %s", name)
	}

	s.handlers[name] = append(s.handlers[name], func(s *Session, e any) {
		handler(s, e.(T))
	})
}

func (s *Session) IsConnected() bool {
	return s.WS != nil && s.WS.IsConnected()
}

// Open determines the Websocket URL and establishes a connection.
// It also detects if you are logged in as a user or a bot.
func (s *Session) Open() (err error) {

	if s.IsConnected() {
		return fmt.Errorf("already connected")
	}

	if s.Token == "" {
		return fmt.Errorf("no token provided")
	}

	// Determine the Websocket URL
	var query RootData
	err = s.Request(http.MethodGet, apiURL, nil, &query)
	if err != nil {
		return
	}

	parameters := url.Values{}
	parameters.Set("token", s.Token)
	parameters.Set("format", "json")
	parameters.Set("version", "1")

	wsURL, err := url.Parse(query.WS)
	if err != nil {
		return
	}

	wsURL.RawQuery = parameters.Encode()

	s.WS = newWebsocket(s, wsURL.String())
	s.WS.connect()

	return
}

// WriteSocket writes data to the Websocket in JSON
func (s *Session) WriteSocket(data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Should we use WriteAsync?
	return s.WS.WriteMessage(gws.OpcodeText, payload)
}

func (s *Session) AttachmentUpload(file *File) (attachment *FileAttachment, err error) {

	if file.Name == "" {
		log.Printf("Warning: uploading files without names may cause the media to not load on the client")
	}

	endpoint := EndpointAutumn("attachments")
	err = s.Request(http.MethodPost, endpoint, file, &attachment)
	return
}

func (s *Session) Emoji(eID string) (emoji *Emoji, err error) {
	endpoint := EndpointCustomEmoji(eID)
	err = s.Request(http.MethodGet, endpoint, nil, &emoji)
	s.State.addEmoji(emoji)
	return
}

func (s *Session) EmojiCreate(eID string, data EmojiCreateData) (emoji *Emoji, err error) {
	endpoint := EndpointCustomEmoji(eID)
	err = s.Request(http.MethodPut, endpoint, data, &emoji)
	return
}

func (s *Session) EmojiDelete(eID string) error {
	endpoint := EndpointCustomEmoji(eID)
	return s.Request(http.MethodDelete, endpoint, nil, nil)
}

// Channel fetches a channel using an API call
func (s *Session) Channel(cID string) (channel *Channel, err error) {
	endpoint := EndpointChannels(cID)
	err = s.Request(http.MethodGet, endpoint, nil, &channel)
	return
}

// User fetches a user by their ID
// To fetch self, supply "@me" as the ID
func (s *Session) User(uID string) (user *User, err error) {
	endpoint := EndpointUsers(uID)
	err = s.Request(http.MethodGet, endpoint, nil, &user)
	s.State.addUser(user)
	return
}

func (s *Session) UserBlock(uID string) (user *User, err error) {
	endpoint := EndpointUsersBlock(uID)
	err = s.Request(http.MethodPut, endpoint, nil, user)
	return
}

func (s *Session) UserUnblock(uID string) (user *User, err error) {
	endpoint := EndpointUsersBlock(uID)
	err = s.Request(http.MethodDelete, endpoint, nil, user)
	return
}

func (s *Session) UserProfile(uID string) (profile *UserProfile, err error) {
	endpoint := EndpointUsersProfile(uID)
	err = s.Request(http.MethodGet, endpoint, nil, &profile)
	return
}

func (s *Session) UserDefaultAvatar(uID string) (binary []byte, err error) {
	endpoint := EndpointUsersDefaultAvatar(uID)
	err = s.Request(http.MethodGet, endpoint, nil, &binary)
	return
}

func (s *Session) SetUsername(data UsernameData) (user *User, err error) {
	err = s.Request(http.MethodPatch, URLUsersUsername, data, &user)
	return
}

func (s *Session) UserFlags(uID string) (flags int, err error) {
	endpoint := EndpointUsersFlags(uID)
	err = s.Request(http.MethodGet, endpoint, nil, &flags)
	return
}

func (s *Session) UserEdit(uID string, data UserEditData) (user *User, err error) {
	endpoint := EndpointUsers(uID)
	err = s.Request(http.MethodPatch, endpoint, data, &user)
	return
}

// Server fetches a server by its ID
func (s *Session) Server(id string) (server *Server, err error) {
	endpoint := EndpointServers(id)
	err = s.Request(http.MethodGet, endpoint, nil, &server)
	s.State.addServer(server)
	return
}

func (s *Session) ServerEdit(id string, data ServerEditData) (server *Server, err error) {
	endpoint := EndpointServers(id)
	err = s.Request(http.MethodPatch, endpoint, data, &server)
	return
}

// ServerCreate creates a server based on the data provided
func (s *Session) ServerCreate(data ServerCreateData) (server *Server, err error) {
	endpoint := EndpointServers("create")
	err = s.Request(http.MethodPost, endpoint, data, &server)
	return
}

// ChannelBeginTyping is a Websocket method to start typing in a channel
func (s *Session) ChannelBeginTyping(cID string) (err error) {
	data := WebsocketChannelTyping{Channel: cID, Type: WebsocketMessageTypeBeginTyping}
	return s.WriteSocket(data)
}

// ChannelEndTyping is a Websocket method to stop typing in a channel
func (s *Session) ChannelEndTyping(cID string) (err error) {
	data := WebsocketChannelTyping{Channel: cID, Type: WebsocketMessageTypeEndTyping}
	return s.WriteSocket(data)
}

func (s *Session) ChannelWebhooks(cID string) (webhooks []*Webhook, err error) {
	endpoint := EndpointChannelsWebhooks(cID)
	err = s.Request(http.MethodGet, endpoint, nil, &webhooks)
	s.State.addWebhooks(webhooks)
	return
}

func (s *Session) ChannelWebhooksCreate(cID string, data WebhookCreate) (webhook *Webhook, err error) {
	endpoint := EndpointChannelsWebhooks(cID)
	err = s.Request(http.MethodPost, endpoint, data, &webhook)
	return
}

// GroupCreate creates a group based on the data provided
// "Users" field is a list of user IDs that will be in the group
func (s *Session) GroupCreate(data GroupCreateData) (group *Group, err error) {
	endpoint := EndpointChannels("create")
	err = s.Request(http.MethodPost, endpoint, data, &group)
	return
}

func (s *Session) GroupMemberAdd(cID, mID string) (err error) {
	endpoint := EndpointChannelsRecipients(cID, mID)
	err = s.Request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) GroupMemberDelete(cID, mID string) (err error) {
	endpoint := EndpointChannelsRecipients(cID, mID)
	err = s.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) GroupMembers(cID string) (users []*User, err error) {
	endpoint := EndpointChannels(cID)
	err = s.Request(http.MethodGet, endpoint, nil, &users)
	return
}

func (s *Session) ChannelInviteCreate(cID string) (invite *InviteCreate, err error) {
	endpoint := EndpointChannelsInvites(cID)
	err = s.Request(http.MethodPost, endpoint, nil, &invite)
	return
}

func (s *Session) ChannelDelete(cID string) (err error) {
	endpoint := EndpointChannels(cID)
	err = s.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerAck(serverID string) (err error) {
	endpoint := EndpointServersAck(serverID)
	err = s.Request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) MessageAck(channelID, messageID string) (err error) {
	endpoint := EndpointChannelAckMessage(channelID, messageID)
	err = s.Request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) ServerBans(sID string) (bans []*ServerBans, err error) {
	endpoint := EndpointServersBans(sID)
	err = s.Request(http.MethodGet, endpoint, nil, &bans)
	return
}

func (s *Session) ServersRole(sID, rID string) (role *ServerRole, err error) {
	endpoint := EndpointServersRole(sID, rID)
	err = s.Request(http.MethodGet, endpoint, nil, &role)
	return
}

func (s *Session) Invite(iID string) (invite *Invite, err error) {
	endpoint := EndpointInvite(iID)
	err = s.Request(http.MethodGet, endpoint, nil, &invite)
	return
}

func (s *Session) InviteJoin(iID string) (invite *Invite, err error) {
	endpoint := EndpointInvite(iID)
	err = s.Request(http.MethodPost, endpoint, nil, &invite)
	return
}

func (s *Session) InviteDelete(iID string) (err error) {
	endpoint := EndpointInvite(iID)
	err = s.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerRoleDelete(sID, rID string) (err error) {
	endpoint := EndpointServersRole(sID, rID)
	err = s.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServersRoleEdit(sID, rID string, data ServerRoleEditData) (role *ServerRole, err error) {
	endpoint := EndpointServersRole(sID, rID)
	err = s.Request(http.MethodPatch, endpoint, data, &role)
	return
}

func (s *Session) ServersRoleCreate(sID string, data ServerRoleCreateData) (role *ServerRole, err error) {
	endpoint := EndpointServersRoles(sID)
	err = s.Request(http.MethodPost, endpoint, data, &role)
	return
}

func (s *Session) PermissionsSet(sID, rID string, data PermissionOverwrite) (err error) {
	endpoint := EndpointServerPermissions(sID, rID)
	err = s.Request(http.MethodPut, endpoint, data, nil)
	return
}

// ChannelPermissionsSet sets permissions for the specified role in this channel.
func (s *Session) ChannelPermissionsSet(cID, rID string, data PermissionOverwrite) (err error) {
	endpoint := EndpointChannelsPermissions(cID, rID)
	return s.Request(http.MethodPut, endpoint, data, nil)
}

// ChannelPermissionsSetDefault sets permissions for the default role in this channel.
func (s *Session) ChannelPermissionsSetDefault(cID string, data PermissionOverwrite) (err error) {
	return s.ChannelPermissionsSet(cID, "default", data)
}

// PermissionsSetDefault sets the permissions of a role in a server
func (s *Session) PermissionsSetDefault(sID string, data PermissionsSetDefaultData) (err error) {
	endpoint := EndpointServerPermissions(sID, "default")
	err = s.Request(http.MethodPut, endpoint, data, nil)
	return
}

func (s *Session) ChannelEdit(cID string, data ChannelEditData) (channel *Channel, err error) {
	endpoint := EndpointChannels(cID)
	err = s.Request(http.MethodPatch, endpoint, data, &channel)
	return
}

func (s *Session) ServerMemberUnban(sID, mID string) (err error) {
	endpoint := EndpointServersBan(sID, mID)
	err = s.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberBan(sID, mID string) (err error) {
	endpoint := EndpointServersBan(sID, mID)
	err = s.Request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberDelete(sID, mID string) (err error) {
	endpoint := EndpointServersMember(sID, mID)
	err = s.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberEdit(sID, mID string, data ServerMemberEditData) (member *ServerMember, err error) {
	endpoint := EndpointServersMember(sID, mID)
	err = s.Request(http.MethodPatch, endpoint, data, &member)
	return
}

func (s *Session) ServerMember(sID, mID string) (member *ServerMember, err error) {
	endpoint := EndpointServersMember(sID, mID)
	err = s.Request(http.MethodGet, endpoint, nil, &member)
	s.State.addServerMember(member)
	return
}

func (s *Session) ServerMembers(sID string) (members *ServerMembers, err error) {
	endpoint := EndpointServersMembers(sID)
	err = s.Request(http.MethodGet, endpoint, nil, &members)
	s.State.addServerMembersAndUsers(members)
	return
}

func (s *Session) ChannelMessage(cID, mID string) (message *Message, err error) {
	endpoint := EndpointChannelsMessage(cID, mID)
	err = s.Request(http.MethodGet, endpoint, nil, &message)
	return
}

// ChannelMessageReactionCreate adds a reaction (emoji ID) to a message
func (s *Session) ChannelMessageReactionCreate(cID, mID, eID string) (err error) {
	endpoint := EndpointChannelsMessageReaction(cID, mID, eID)
	err = s.Request(http.MethodPut, endpoint, nil, nil)
	return
}

// ChannelMessageReactionDelete deletes a singular reaction on a message
func (s *Session) ChannelMessageReactionDelete(cID, mID, eID string) (err error) {
	endpoint := EndpointChannelsMessageReaction(cID, mID, eID)
	err = s.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

// ChannelMessageReactionClear clears all reactions on a message
func (s *Session) ChannelMessageReactionClear(cID, mID string) (err error) {
	endpoint := EndpointChannelsMessageReactions(cID, mID)
	return s.Request(http.MethodDelete, endpoint, nil, nil)
}

// ChannelsJoinCall asks the voice server for a token to join the call.
func (s *Session) ChannelsJoinCall(cID string, data ChannelJoinCallData) (call ChannelJoinCall, err error) {
	endpoint := EndpointChannelsJoinCall(cID)
	err = s.Request(http.MethodPost, endpoint, data, &call)
	return
}

// ChannelsEndRing stops ringing a user in a DM if a call exists; returns NotConnected otherwise.
// Only works within DMs and groups; returns NoEffect in servers.
// Returns NotFound if the user is not in the DM/group channel.
func (s *Session) ChannelsEndRing(cID, uID string) error {
	endpoint := EndpointChannelsEndRing(cID, uID)
	return s.Request(http.MethodPut, endpoint, nil, nil)
}

func (s *Session) ServerChannelCreate(sID string, data ServerChannelCreateData) (channel *Channel, err error) {
	endpoint := EndpointServersChannels(sID)
	err = s.Request(http.MethodPost, endpoint, data, &channel)
	return
}

func (s *Session) ServerDelete(sID string) error {
	endpoint := EndpointServers(sID)
	return s.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) ChannelMessages(cID string, params ...ChannelMessagesParams) (messages []*Message, err error) {
	endpoint := EndpointChannelsMessages(cID)

	if len(params) > 0 {
		endpoint += params[0].Encode()
	}

	err = s.Request(http.MethodGet, endpoint, nil, &messages)
	return
}

func (s *Session) ChannelMessageEdit(cID, mID string, data MessageEditData) (message *Message, err error) {
	endpoint := EndpointChannelsMessage(cID, mID)
	err = s.Request(http.MethodPatch, endpoint, data, &message)
	return
}

func (s *Session) ChannelMessageSend(cID string, data MessageSend) (message *Message, err error) {
	endpoint := EndpointChannelsMessages(cID)
	err = s.Request(http.MethodPost, endpoint, data, &message)
	return
}

func (s *Session) ChannelMessageDelete(cID, mID string) error {
	endpoint := EndpointChannelsMessage(cID, mID)
	return s.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) ChannelMessageDeleteBulk(cID string, messages ChannelMessageBulkDeleteData) error {
	endpoint := EndpointChannelsMessage(cID, "bulk")
	return s.Request(http.MethodDelete, endpoint, messages, nil)
}

func (s *Session) AccountCreate(data AccountCreateData) error {
	endpoint := EndpointAuthAccount("create")
	return s.Request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) AccountReverify(data AccountReverifyData) error {
	endpoint := EndpointAuthAccount("reverify")
	return s.Request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) AccountDeleteConfirm(data AccountDeleteConfirmData) error {
	endpoint := EndpointAuthAccount("delete")
	return s.Request(http.MethodPut, endpoint, data, nil)
}

func (s *Session) AccountDelete() error {
	endpoint := EndpointAuthAccount("delete")
	return s.Request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) Account() (account *Account, err error) {
	endpoint := EndpointAuthAccount("")
	err = s.Request(http.MethodGet, endpoint, nil, &account)
	return
}

func (s *Session) AccountDisable() error {
	endpoint := EndpointAuthAccount("disable")
	return s.Request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) AccountChangePassword(data AccountChangePasswordData) error {
	endpoint := EndpointAuthAccountChange("password")
	return s.Request(http.MethodPatch, endpoint, data, nil)
}

func (s *Session) AccountChangeEmail(data AccountChangeEmailData) error {
	endpoint := EndpointAuthAccountChange("email")
	return s.Request(http.MethodPatch, endpoint, data, nil)
}

func (s *Session) VerifyEmail(code string) (ticket ChangeEmail, err error) {
	endpoint := EndpointAuthAccountVerify(code)
	err = s.Request(http.MethodPost, endpoint, nil, &ticket)
	return
}

// PasswordReset requests a password reset, which is sent to the email provided
func (s *Session) PasswordReset(data AccountReverifyData) error {
	endpoint := EndpointAuthAccount("reset_password")
	return s.Request(http.MethodPost, endpoint, data, nil)
}

// PasswordResetConfirm confirms a password reset
func (s *Session) PasswordResetConfirm(data PasswordResetConfirmData) error {
	endpoint := EndpointAuthAccount("reset_password")
	return s.Request(http.MethodPatch, endpoint, data, nil)
}

// Login as a regular user instead of bot. Friendly name is used to identify the session via MFA
func (s *Session) Login(data LoginData) (mfa LoginResponse, err error) {
	endpoint := EndpointAuthSession("login")
	err = s.Request(http.MethodPost, endpoint, data, &mfa)
	return
}

func (s *Session) Sessions() (sessions []*Sessions, err error) {
	endpoint := EndpointAuthSession("all")
	err = s.Request(http.MethodGet, endpoint, nil, &sessions)
	return
}

func (s *Session) SessionEdit(id string, data SessionEditData) (session SessionEditData, err error) {
	endpoint := EndpointAuthSession(id)
	err = s.Request(http.MethodPatch, endpoint, data, &session)
	return
}

// Onboarding returns whether the current account requires onboarding or whether you can continue to send requests as usual
func (s *Session) Onboarding() (onboarding Onboarding, err error) {
	endpoint := EndpointOnboard("hello")
	err = s.Request(http.MethodGet, endpoint, nil, &onboarding)
	return
}

// OnboardingComplete sets a new username, completes onboarding and allows a user to start using Revolt.
func (s *Session) OnboardingComplete(data OnboardingCompleteData) error {
	endpoint := EndpointOnboard("complete")
	return s.Request(http.MethodPost, endpoint, data, nil)
}

// SessionsDelete invalidates a session with the provided ID
func (s *Session) SessionsDelete(id string) error {
	endpoint := EndpointAuthSession(id)
	return s.Request(http.MethodDelete, endpoint, nil, nil)
}

// SessionsDeleteAll invalidates all sessions, including this one if revokeSelf is true
func (s *Session) SessionsDeleteAll(revokeSelf bool) error {
	endpoint := EndpointAuthSession("all")
	if revokeSelf {
		values := url.Values{}
		values.Set("revoke_self", "true")
		endpoint += fmt.Sprintf("?%s", values.Encode())
	}

	return s.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) Logout() error {
	endpoint := EndpointAuthSession("logout")
	return s.Request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) UserMutual(uID string) (mutual []*MutualFriendsAndServersResponse, err error) {
	endpoint := EndpointUsersMutual(uID)
	err = s.Request(http.MethodGet, endpoint, nil, &mutual)
	return
}

// DirectMessages returns a list of direct message channels.
func (s *Session) DirectMessages() (channels []*Channel, err error) {
	endpoint := EndpointUsers("dms")
	err = s.Request(http.MethodGet, endpoint, nil, &channels)
	return
}

// DirectMessageCreate opens a direct message channel with a user
// Will return an error "MissingPermission" "SendMessage" if you are not friends or blocked
func (s *Session) DirectMessageCreate(uID string) (channel *Channel, err error) {
	endpoint := EndpointUsersDM(uID)
	err = s.Request(http.MethodGet, endpoint, nil, &channel)
	return
}

// Relationships returns a list of relationships for the current user
func (s *Session) Relationships() (relationships []*UserRelations, err error) {
	endpoint := URLUsersRelationships
	err = s.Request(http.MethodGet, endpoint, nil, &relationships)
	return
}

// FriendAdd sends or accepts a friend Request.
func (s *Session) FriendAdd(username string) (user *User, err error) {
	endpoint := EndpointUsersFriend(username)
	err = s.Request(http.MethodPut, endpoint, nil, &user)
	return
}

// FriendDelete removes a friend or declines a friend Request.
func (s *Session) FriendDelete(username string) (user *User, err error) {
	endpoint := EndpointUsersFriend(username)
	err = s.Request(http.MethodDelete, endpoint, nil, &user)
	return
}

// Bot fetches details of a bot you own by its ID
func (s *Session) Bot(bID string) (bot *FetchedBot, err error) {
	endpoint := EndpointBots(bID)
	err = s.Request(http.MethodGet, endpoint, nil, &bot)
	return
}

// Bots returns a list of bots for the current user
func (s *Session) Bots() (bots *FetchedBots, err error) {
	endpoint := EndpointBots("@me")
	err = s.Request(http.MethodGet, endpoint, nil, &bots)
	return
}

// BotCreate creates a bot based on the data provided
func (s *Session) BotCreate(data BotCreateData) (bot *Bot, err error) {
	endpoint := EndpointBots("create")
	err = s.Request(http.MethodPost, endpoint, data, &bot)
	return
}

func (s *Session) BotEdit(id string, data BotEditData) (bot *Bot, err error) {
	endpoint := EndpointBots(id)
	err = s.Request(http.MethodPatch, endpoint, data, &bot)
	return
}

func (s *Session) BotDelete(bID string) error {
	endpoint := EndpointBots(bID)
	return s.Request(http.MethodDelete, endpoint, nil, nil)
}

// BotPublic fetches a public bot by its ID
func (s *Session) BotPublic(bID string) (bot *PublicBot, err error) {
	endpoint := EndpointBotsInvite(bID)
	err = s.Request(http.MethodGet, endpoint, nil, &bot)
	return
}

// BotInvite invites a bot by its ID to a server or group
func (s *Session) BotInvite(bID string, data BotInviteData) (err error) {
	endpoint := EndpointBotsInvite(bID)
	err = s.Request(http.MethodPost, endpoint, data, nil)
	return
}

func (s *Session) SyncUnreads() (data []SyncUnread, err error) {
	endpoint := EndpointSync("unreads")
	err = s.Request(http.MethodGet, endpoint, nil, &data)
	return
}

func (s *Session) SyncSettingsFetch(payload SyncSettingsFetchData) (data *SyncSettingsData, err error) {
	endpoint := EndpointSync("settings")
	err = s.Request(http.MethodPost, endpoint, payload, &data)
	return
}

func (s *Session) SyncSettingsSet(payload SyncSettingsData) error {
	endpoint := EndpointSync("settings")
	return s.Request(http.MethodPut, endpoint, payload, nil)
}

func (s *Session) PushSubscribe(data WebpushSubscription) error {
	endpoint := EndpointPush("subscribe")
	return s.Request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) PushUnsubscribe(data WebpushSubscription) error {
	endpoint := EndpointPush("unsubscribe")
	return s.Request(http.MethodPost, endpoint, data, nil)
}
