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
		ShouldReconnect:   true,
		Token:             token,
		State:             newState(),
		Ratelimiter:       newRatelimiter(),
		HeartbeatInterval: 30 * time.Second,
		ReconnectInterval: 5 * time.Second,
		UserAgent:         fmt.Sprintf("RevoltGo/%s (github.com/sentinelb51/revoltgo)", VERSION),
		HTTP:              &http.Client{Timeout: 10 * time.Second},
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

// Session represents a connection to the Revolt API.
type Session struct {

	// Authorisation token
	Token string

	// The websocket connection
	Socket *gws.Conn

	// HTTP client used for the REST API
	HTTP *http.Client

	// Ratelimiter for the REST API
	Ratelimiter *Ratelimiter

	// State is a central store for all data received from the API
	State *State

	// The user agent used for REST APIs
	UserAgent string

	// Indicates whether the websocket is connected
	Connected bool

	// Whether the websocket should reconnect when the connection drops
	ShouldReconnect bool

	// Defines a custom compression algorithm for the websocket
	// By default, compression is enabled at the fastest level (1) for >=512 byte payloads
	// To enable, set gws.PermessageDeflate.Enabled true
	CustomCompression *gws.PermessageDeflate

	// Interval between sending heartbeats. Lower values update the latency faster
	// Values too high (~100 seconds) may cause Cloudflare to drop the connection
	HeartbeatInterval time.Duration

	// Heartbeat counter
	HeartbeatCount int64

	// Interval between reconnecting, if connection fails
	ReconnectInterval time.Duration

	// Last time a ping was sent
	LastHeartbeatSent time.Time

	// Last time a ping was received
	LastHeartbeatAck time.Time

	/* Private fields */

	// Whether the session is a user or bot
	selfbot bool

	/* Event handlers */

	// Authentication-related handlers
	handlersReady         []func(*Session, *EventReady)
	handlersAuth          []func(*Session, *EventAuth)
	handlersPong          []func(*Session, *EventPong)
	handlersAuthenticated []func(*Session, *EventAuthenticated)

	// User-related handlers
	handlersUserUpdate         []func(*Session, *EventUserUpdate)
	handlersUserSettingsUpdate []func(*Session, *EventUserSettingsUpdate)
	handlersUserRelationship   []func(*Session, *EventUserRelationship)
	handlersUserPlatformWipe   []func(*Session, *EventUserPlatformWipe)

	// Message-related handlers
	handlersMessage        []func(*Session, *EventMessage)
	handlersMessageAppend  []func(*Session, *EventMessageAppend)
	handlersMessageUpdate  []func(*Session, *EventMessageUpdate)
	handlersMessageDelete  []func(*Session, *EventMessageDelete)
	handlersMessageReact   []func(*Session, *EventMessageReact)
	handlersMessageUnreact []func(*Session, *EventMessageUnreact)

	// Channel-related handlers
	handlersChannelCreate      []func(*Session, *EventChannelCreate)
	handlersChannelUpdate      []func(*Session, *EventChannelUpdate)
	handlersChannelDelete      []func(*Session, *EventChannelDelete)
	handlersChannelStartTyping []func(*Session, *EventChannelStartTyping)
	handlersChannelStopTyping  []func(*Session, *EventChannelStopTyping)
	handlersChannelAck         []func(*Session, *EventChannelAck)

	// Group-related handlers
	handlersGroupJoin  []func(*Session, *EventChannelGroupJoin)
	handlersGroupLeave []func(*Session, *EventChannelGroupLeave)

	// Server-related handlers
	handlersServerCreate []func(*Session, *EventServerCreate)
	handlersServerUpdate []func(*Session, *EventServerUpdate)
	handlersServerDelete []func(*Session, *EventServerDelete)

	// ServerRole-related handlers
	handlersServerRoleUpdate []func(*Session, *EventServerRoleUpdate)
	handlersServerRoleDelete []func(*Session, *EventServerRoleDelete)

	// ServerMember-related handlers
	handlersServerMemberUpdate []func(*Session, *EventServerMemberUpdate)
	handlersServerMemberJoin   []func(*Session, *EventServerMemberJoin)
	handlersServerMemberLeave  []func(*Session, *EventServerMemberLeave)

	// Emoji-related handlers
	handlersEmojiCreate []func(*Session, *EventEmojiCreate)
	handlersEmojiDelete []func(*Session, *EventEmojiDelete)

	// Webhook-related handlers
	handlersWebhookCreate []func(*Session, *EventWebhookCreate)
	handlersWebhookUpdate []func(*Session, *EventWebhookUpdate)
	handlersWebhookDelete []func(*Session, *EventWebhookDelete)

	// System event handlers
	handlersAbstractEventUpdate []func(*Session, *AbstractEventUpdate)
	handlersError               []func(*Session, *EventError)
	handlersBulk                []func(*Session, *EventBulk)
}

// Selfbot returns whether the session is a selfbot
func (s *Session) Selfbot() bool {
	return s.selfbot
}

// addDefaultHandlers adds mission-critical handlers to keep the library working
func (s *Session) addDefaultHandlers() {

	// The websocket's first response if authentication was unsuccessful
	s.AddHandler(func(s *Session, e *EventError) {
		log.Printf("Authentication error: %s\n", e.Error)
	})

	s.AddHandler(func(s *Session, e *EventBulk) {
		for _, event := range e.V {
			go handle(s, event)
		}
	})

	s.AddHandler(func(s *Session, e *EventReady) {
		s.State.populate(e)
		s.selfbot = s.State.self != nil && s.State.self.Bot == nil
	})

	// If state is disabled, none of these handlers are required
	if s.State == nil {
		return
	}

	if s.State.TrackUsers() {
		s.AddHandler(func(s *Session, e *EventUserPlatformWipe) {
			s.State.platformWipe(e)
		})
	}

	if s.State.TrackChannels() {
		s.AddHandler(func(s *Session, e *EventChannelCreate) {
			s.State.createChannel(e)
		})

		s.AddHandler(func(s *Session, e *EventChannelDelete) {
			s.State.deleteChannel(e)
		})

		s.AddHandler(func(s *Session, e *EventChannelGroupJoin) {
			s.State.addGroupParticipant(e)
		})

		s.AddHandler(func(s *Session, e *EventChannelGroupLeave) {
			s.State.removeGroupParticipant(e)
		})
	}

	if s.State.TrackServers() {
		s.AddHandler(func(s *Session, e *EventServerCreate) {
			s.State.createServer(e)
		})

		s.AddHandler(func(s *Session, e *EventServerDelete) {
			s.State.deleteServer(e)
		})

		s.AddHandler(func(s *Session, e *EventServerRoleDelete) {
			s.State.deleteServerRole(e)
		})
	}

	if s.State.TrackMembers() {
		s.AddHandler(func(s *Session, e *EventServerMemberJoin) {
			s.State.createServerMember(e)
		})

		s.AddHandler(func(s *Session, e *EventServerMemberLeave) {
			s.State.deleteServerMember(e)
		})
	}

	if s.State.TrackEmojis() {
		s.AddHandler(func(s *Session, e *EventEmojiCreate) {
			s.State.createEmoji(e)
		})

		s.AddHandler(func(s *Session, e *EventEmojiDelete) {
			s.State.deleteEmoji(e)
		})
	}

	if s.State.TrackWebhooks() {
		s.AddHandler(func(s *Session, e *EventWebhookCreate) {
			s.State.createWebhook(e)
		})

		s.AddHandler(func(s *Session, e *EventWebhookDelete) {
			s.State.deleteWebhook(e)
		})
	}

	s.AddHandler(func(s *Session, e *AbstractEventUpdate) {
		e.standardise()

		switch e.Type {
		case "ServerUpdate":
			s.State.updateServer(e)

			if len(s.handlersServerUpdate) == 0 {
				return
			}

			event := e.EventServerUpdate()

			for _, h := range s.handlersServerUpdate {
				h(s, event)
			}
		case "ServerMemberUpdate":
			s.State.updateServerMember(e)

			if len(s.handlersServerMemberUpdate) == 0 {
				return
			}

			event := e.EventServerMemberUpdate()

			for _, h := range s.handlersServerMemberUpdate {
				h(s, event)
			}
		case "ChannelUpdate":
			s.State.updateChannel(e)

			if len(s.handlersChannelUpdate) == 0 {
				return
			}

			event := e.EventChannelUpdate()

			for _, h := range s.handlersChannelUpdate {
				h(s, event)
			}
		case "UserUpdate":
			s.State.updateUser(e)

			if len(s.handlersUserUpdate) == 0 {
				return
			}

			event := e.EventUserUpdate()

			for _, h := range s.handlersUserUpdate {
				h(s, event)
			}
		case "ServerRoleUpdate":
			s.State.updateServerRole(e)

			if len(s.handlersServerRoleUpdate) == 0 {
				return
			}

			event := e.EventServerRoleUpdate()

			for _, h := range s.handlersServerRoleUpdate {
				h(s, event)
			}
		case "WebhookUpdate":
			s.State.updateWebhook(e)

			if len(s.handlersWebhookUpdate) == 0 {
				return
			}

			event := e.EventWebhookUpdate()

			for _, h := range s.handlersWebhookUpdate {
				h(s, event)
			}
		}
	})
}

// AddHandler registers an event handler based on function signature
func (s *Session) AddHandler(handler any) {
	switch h := handler.(type) {
	case func(*Session, *AbstractEventUpdate):
		s.handlersAbstractEventUpdate = append(s.handlersAbstractEventUpdate, h)
	case func(*Session, *EventError):
		s.handlersError = append(s.handlersError, h)
	case func(*Session, *EventBulk):
		s.handlersBulk = append(s.handlersBulk, h)
	case func(*Session, *EventReady):
		s.handlersReady = append(s.handlersReady, h)
	case func(*Session, *EventAuth):
		s.handlersAuth = append(s.handlersAuth, h)
	case func(*Session, *EventPong):
		s.handlersPong = append(s.handlersPong, h)
	case func(*Session, *EventAuthenticated):
		s.handlersAuthenticated = append(s.handlersAuthenticated, h)
	case func(*Session, *EventUserUpdate):
		s.handlersUserUpdate = append(s.handlersUserUpdate, h)
	case func(*Session, *EventServerUpdate):
		s.handlersServerUpdate = append(s.handlersServerUpdate, h)
	case func(*Session, *EventChannelUpdate):
		s.handlersChannelUpdate = append(s.handlersChannelUpdate, h)
	case func(*Session, *EventServerRoleUpdate):
		s.handlersServerRoleUpdate = append(s.handlersServerRoleUpdate, h)
	case func(*Session, *EventWebhookUpdate):
		s.handlersWebhookUpdate = append(s.handlersWebhookUpdate, h)
	case func(*Session, *EventServerMemberUpdate):
		s.handlersServerMemberUpdate = append(s.handlersServerMemberUpdate, h)
	case func(*Session, *EventMessage):
		s.handlersMessage = append(s.handlersMessage, h)
	case func(*Session, *EventMessageAppend):
		s.handlersMessageAppend = append(s.handlersMessageAppend, h)
	case func(*Session, *EventMessageUpdate):
		s.handlersMessageUpdate = append(s.handlersMessageUpdate, h)
	case func(*Session, *EventMessageDelete):
		s.handlersMessageDelete = append(s.handlersMessageDelete, h)
	case func(*Session, *EventMessageReact):
		s.handlersMessageReact = append(s.handlersMessageReact, h)
	case func(*Session, *EventMessageUnreact):
		s.handlersMessageUnreact = append(s.handlersMessageUnreact, h)
	case func(*Session, *EventChannelCreate):
		s.handlersChannelCreate = append(s.handlersChannelCreate, h)
	case func(*Session, *EventChannelDelete):
		s.handlersChannelDelete = append(s.handlersChannelDelete, h)
	case func(*Session, *EventChannelStartTyping):
		s.handlersChannelStartTyping = append(s.handlersChannelStartTyping, h)
	case func(*Session, *EventChannelStopTyping):
		s.handlersChannelStopTyping = append(s.handlersChannelStopTyping, h)
	case func(*Session, *EventChannelAck):
		s.handlersChannelAck = append(s.handlersChannelAck, h)
	case func(*Session, *EventChannelGroupJoin):
		s.handlersGroupJoin = append(s.handlersGroupJoin, h)
	case func(*Session, *EventChannelGroupLeave):
		s.handlersGroupLeave = append(s.handlersGroupLeave, h)
	case func(*Session, *EventServerCreate):
		s.handlersServerCreate = append(s.handlersServerCreate, h)
	case func(*Session, *EventServerDelete):
		s.handlersServerDelete = append(s.handlersServerDelete, h)
	case func(*Session, *EventServerMemberJoin):
		s.handlersServerMemberJoin = append(s.handlersServerMemberJoin, h)
	case func(*Session, *EventServerMemberLeave):
		s.handlersServerMemberLeave = append(s.handlersServerMemberLeave, h)
	case func(*Session, *EventServerRoleDelete):
		s.handlersServerRoleDelete = append(s.handlersServerRoleDelete, h)
	case func(*Session, *EventEmojiCreate):
		s.handlersEmojiCreate = append(s.handlersEmojiCreate, h)
	case func(*Session, *EventEmojiDelete):
		s.handlersEmojiDelete = append(s.handlersEmojiDelete, h)
	case func(*Session, *EventUserSettingsUpdate):
		s.handlersUserSettingsUpdate = append(s.handlersUserSettingsUpdate, h)
	case func(*Session, *EventUserRelationship):
		s.handlersUserRelationship = append(s.handlersUserRelationship, h)
	case func(*Session, *EventUserPlatformWipe):
		s.handlersUserPlatformWipe = append(s.handlersUserPlatformWipe, h)
	case func(*Session, *EventWebhookCreate):
		s.handlersWebhookCreate = append(s.handlersWebhookCreate, h)
	case func(*Session, *EventWebhookDelete):
		s.handlersWebhookDelete = append(s.handlersWebhookDelete, h)
	default:

		handlerType := reflect.TypeOf(handler)

		if handlerType.Kind() != reflect.Func {
			log.Printf("Handler %s not registered: expected a function, got: %v", handlerType, handlerType)
			return
		}

		// Get the amount of arguments
		inputSize := handlerType.NumIn()
		inputSizeExpected := 2
		if inputSize != inputSizeExpected {
			log.Printf("Handler %s not registered: expected %d arguments, got: %d", handlerType, inputSizeExpected, inputSize)
			return
		}

		secondArgument := handlerType.In(1)
		secondArgumentName := secondArgument.String()
		secondArgumentExpected := strings.ReplaceAll(secondArgumentName, "revoltgo.", "revoltgo.Event")

		if secondArgumentName != secondArgumentExpected {
			log.Printf(
				"Handler %s not registered: %s is not an event. Did you mean: %s\n",
				handlerType, secondArgumentName, secondArgumentExpected,
			)
		}
	}
}

// Latency returns the websocket latency
func (s *Session) Latency() time.Duration {
	return s.LastHeartbeatAck.Sub(s.LastHeartbeatSent)
}

// Uptime returns the approximate duration the websocket has been connected for
func (s *Session) Uptime() time.Duration {
	// todo: add time difference between last heartbeat and next heartbeat
	return time.Duration(s.HeartbeatCount) * s.HeartbeatInterval
}

// Open determines the websocket URL and establishes a connection.
// It also detects if you are logged in as a user or a bot.
func (s *Session) Open() (err error) {

	if s.Connected {
		return fmt.Errorf("already connected")
	}

	if s.Token == "" {
		return fmt.Errorf("no token provided")
	}

	// Determine the websocket URL
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

	s.Socket = connect(s, wsURL.String())

	// Assume we have a successful connection, until we don't
	s.Connected = true
	return
}

// WriteSocket writes data to the websocket in JSON
func (s *Session) WriteSocket(data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Should we use WriteAsync?
	return s.Socket.WriteMessage(gws.OpcodeText, payload)
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

func (s *Session) PermissionsSet(sID, rID string, data PermissionAD) (err error) {
	endpoint := EndpointServerPermissions(sID, rID)
	err = s.Request(http.MethodPut, endpoint, data, nil)
	return
}

func (s *Session) ChannelPermissionsSet(sID, cID string, data PermissionAD) (err error) {
	endpoint := EndpointChannelsPermissions(sID, cID)
	err = s.Request(http.MethodPut, endpoint, data, nil)
	return
}

func (s *Session) ChannelPermissionsSetDefault(sID string, data PermissionAD) (err error) {
	endpoint := EndpointChannelsPermissions(sID, "default")
	err = s.Request(http.MethodPut, endpoint, data, nil)
	return
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
	err = s.Request(http.MethodDelete, endpoint, nil, nil)
	return
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
