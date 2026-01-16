package revoltgo

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/goccy/go-json"
	"github.com/lxzan/gws"
	"github.com/tinylib/msgp/msgp"
)

const ExpressLoginFile string = ".auth_token"

func New(token string) *Session {
	session := &Session{
		Token: token,
		State: newState(),
	}

	session.HTTP = newHTTPClient(session)
	session.addDefaultHandlers()

	// There may be some validation/boundary checks in the future.

	return session
}

// NewWithLogin exchanges an email and password for a session token, and then creates a new session.
// You are expected to store and re-use the token for future sessions.
func NewWithLogin(data LoginData) (*Session, LoginResponse, error) {
	session := New("")
	session.selfbot = true

	if data.FriendlyName == "" {
		data.FriendlyName = fmt.Sprintf("RevoltGo/%s (%d)", VERSION, os.Getpid())
	}

	mfa, err := session.Login(data)
	if err == nil {
		session.Token = mfa.Token
	}

	return session, mfa, err
}

// NewWithExpressLogin exchanges an email and password for a session token, and then creates a new session.
// Unlike NewWithLogin, this automatically reads and writes the token to a file for you.
// Make sure you trust the environment where this is run, as the token is stored in plaintext.
func NewWithExpressLogin(data LoginData) (*Session, error) {

	var token string

	// Check if file exists
	if _, err := os.Stat(ExpressLoginFile); err == nil {
		// File exists, read the token
		bytes, err := os.ReadFile(ExpressLoginFile)
		if err != nil {
			return nil, err
		}

		token = string(bytes)
	}

	// If token exists, use it
	if token != "" {
		log.Println("Attempting to re-use existing token")
		session := New(token)
		session.selfbot = true
		return session, nil
	}

	// Otherwise, perform login
	log.Println("Performing authentication...")
	session, mfa, err := NewWithLogin(data)
	if err != nil {
		return nil, err
	}

	// Save token to file
	log.Println("Saving authentication token for re-use")
	err = os.WriteFile(ExpressLoginFile, []byte(mfa.Token), 0o600)
	return session, err
}

// Session represents a connection to the Revolt API.
type Session struct {
	Token string      // Authorisation token
	WS    *Websocket  // Websocket handler for bidirectional events
	HTTP  *HTTPClient // HTTP handler for the REST API
	State *State      // State is a central store for all data received from the API

	/* Private fields */

	selfbot bool // Whether the session is a user or bot

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

		AddHandler(s, func(s *Session, e *EventServerRoleRanksUpdate) {
			s.State.updateServerRoleRanks(e)
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

	if t.Kind() != reflect.Struct {
		log.Fatalf("expected struct type, got %s", t.Kind())
	}

	if field, ok := t.FieldByName("Type"); ok {
		if field.Type.Kind() != reflect.String {
			log.Fatalf("event struct %s 'Type' field must be a string", t.Name())
		}
	} else {
		log.Fatalf("struct %s must have a 'Type' field", t.Name())
	}

	if !strings.HasPrefix(t.Name(), "Event") {
		log.Fatalf("struct %s must be prefixed with 'Event'", t.Name())
	}

	// Convert "EventMessage" struct name to "Message" event name
	name := strings.TrimPrefix(t.Name(), "Event")

	// Safety check (assuming eventConstructors is defined elsewhere)
	if _, found := eventConstructors[name]; !found {
		log.Fatalf("attempting to bind handler for unsupported event type: %s", t.Name())
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
	err = s.HTTP.Request(http.MethodGet, apiURL, nil, &query)
	if err != nil {
		return
	}

	log.Printf("API version detected: %s\n", query.Revolt)

	parameters := url.Values{}
	parameters.Set("token", s.Token)
	parameters.Set("format", "msgpack")
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

// WriteSocketJSON writes data to the websocket in JSON
func (s *Session) WriteSocketJSON(data any) error {
	payload, err := json.Marshal(data)
	if err == nil {
		err = s.WS.WriteMessage(gws.OpcodeText, payload)
	}

	return err
}

// WriteSocketMSGP writes data to the Websocket in MessagePack
func (s *Session) WriteSocketMSGP(data any) error {
	marshaler, ok := data.(msgp.Marshaler)
	if !ok {
		// todo: maybe err should just be "%T doesn't implement msgp.Marshaler"
		// todo: test if sending websocket events even works
		err := fmt.Errorf("%T doesn't implement msgp.Marshaler. Did you mean to use WriteSocketJSON, or is revoltgo_msgp_gen outdated", data)
		log.Println(err)
		return err
	}

	payload, err := marshaler.MarshalMsg(nil)
	if err == nil {
		err = s.WS.WriteMessage(gws.OpcodeBinary, payload)
	}

	return err
}

func (s *Session) AttachmentUpload(file *File) (attachment *FileAttachment, err error) {

	if file.Name == "" {
		log.Printf("Warning: uploading files without names may cause the media to not load on the client")
	}

	endpoint := EndpointAutumn("attachments")
	err = s.HTTP.Request(http.MethodPost, endpoint, file, &attachment)
	return
}

func (s *Session) Emoji(eID string) (emoji *Emoji, err error) {
	endpoint := EndpointCustomEmoji(eID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &emoji)
	s.State.addEmoji(emoji)
	return
}

func (s *Session) EmojiCreate(eID string, data EmojiCreateData) (emoji *Emoji, err error) {
	endpoint := EndpointCustomEmoji(eID)
	err = s.HTTP.Request(http.MethodPut, endpoint, data, &emoji)
	return
}

func (s *Session) EmojiDelete(eID string) error {
	endpoint := EndpointCustomEmoji(eID)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

// Channel fetches a channel using an API call
func (s *Session) Channel(cID string) (channel *Channel, err error) {
	endpoint := EndpointChannel(cID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &channel)
	s.State.addChannel(channel)
	return
}

// User fetches a user by their ID
// To fetch self, supply "@me" as the ID
func (s *Session) User(uID string) (user *User, err error) {
	endpoint := EndpointUser(uID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &user)
	s.State.addUser(user)
	return
}

func (s *Session) UserBlock(uID string) (user *User, err error) {
	endpoint := EndpointUserBlock(uID)
	err = s.HTTP.Request(http.MethodPut, endpoint, nil, user)
	return
}

func (s *Session) UserUnblock(uID string) (user *User, err error) {
	endpoint := EndpointUserBlock(uID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, user)
	return
}

func (s *Session) UserProfile(uID string) (profile *UserProfile, err error) {
	endpoint := EndpointUserProfile(uID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &profile)
	return
}

func (s *Session) UserDefaultAvatar(uID string) (binary []byte, err error) {
	endpoint := EndpointUserDefaultAvatar(uID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &binary)
	return
}

func (s *Session) SetUsername(data UsernameData) (user *User, err error) {
	err = s.HTTP.Request(http.MethodPatch, URLUserMeUsername, data, &user)
	return
}

func (s *Session) UserFlags(uID string) (flags int, err error) {
	endpoint := EndpointUserFlags(uID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &flags)
	return
}

func (s *Session) UserEdit(uID string, data UserEditData) (user *User, err error) {
	endpoint := EndpointUser(uID)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &user)
	return
}

// Server fetches a server by its ID
func (s *Session) Server(id string) (server *Server, err error) {
	endpoint := EndpointServer(id)
	// todo: yep... this exists. Turns channels into object array of channels
	// endpoint = fmt.Sprintf("%s?include_channels=true", endpoint)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &server)
	s.State.addServer(server)
	return
}

func (s *Session) ServerEdit(id string, data ServerEditData) (server *Server, err error) {
	endpoint := EndpointServer(id)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &server)
	return
}

// ServerCreate creates a server based on the data provided
func (s *Session) ServerCreate(data ServerCreateData) (server *Server, err error) {
	endpoint := EndpointServer("create")
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &server)
	return
}

// ChannelBeginTyping is a Websocket method to start typing in a channel
func (s *Session) ChannelBeginTyping(cID string) (err error) {
	data := WebsocketChannelTyping{Channel: cID, Type: WebsocketMessageTypeBeginTyping}
	return s.WriteSocketMSGP(data)
}

func (s *Session) ChannelSearch(cID string, query ChannelSearchParams) (messages []*Message, err error) {
	endpoint := EndpointChannelSearch(cID)
	err = s.HTTP.Request(http.MethodPost, endpoint, query, &messages)
	return
}

func (s *Session) ChannelMessagePin(cID, mID string) (err error) {
	endpoint := EndpointChannelMessagesPin(cID, mID)
	return s.HTTP.Request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) ChannelMessageUnpin(cID, mID string) (err error) {
	endpoint := EndpointChannelMessagesPin(cID, mID)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

// ChannelEndTyping is a Websocket method to stop typing in a channel
func (s *Session) ChannelEndTyping(cID string) (err error) {
	data := WebsocketChannelTyping{Channel: cID, Type: WebsocketMessageTypeEndTyping}
	return s.WriteSocketMSGP(data)
}

func (s *Session) ChannelWebhooks(cID string) (webhooks []*Webhook, err error) {
	endpoint := EndpointChannelWebhooks(cID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &webhooks)
	s.State.addWebhooks(webhooks)
	return
}

func (s *Session) ChannelWebhookCreate(cID string, data WebhookCreateData) (webhook *Webhook, err error) {
	endpoint := EndpointChannelWebhooks(cID)
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &webhook)
	return
}

// Webhook fetches a webhook using its ID
func (s *Session) Webhook(wID string) (webhook *Webhook, err error) {
	endpoint := EndpointWebhook(wID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &webhook)
	return
}

// WebhookToken fetches a webhook using its ID and token
func (s *Session) WebhookToken(wID, wToken string) (webhook *Webhook, err error) {
	endpoint := EndpointWebhookToken(wID, wToken)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &webhook)
	return
}

func (s *Session) WebhookTokenExecute(wID, wToken string, data WebhookExecuteData) (message *Message, err error) {
	endpoint := EndpointWebhookToken(wID, wToken)
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &message)
	return
}

func (s *Session) WebhookDelete(wID string) (err error) {
	endpoint := EndpointWebhook(wID)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) WebhookTokenDelete(wID, wToken string) (err error) {
	endpoint := EndpointWebhookToken(wID, wToken)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) WebhookEdit(wID string, data WebhookEditData) (webhook *Webhook, err error) {
	endpoint := EndpointWebhook(wID)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &webhook)
	return
}

func (s *Session) WebhookTokenEdit(wID, wToken string, data WebhookEditData) (webhook *Webhook, err error) {
	endpoint := EndpointWebhookToken(wID, wToken)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &webhook)
	return
}

// WebhookTokenExecuteGitHub is not implemented yet.
func (s *Session) WebhookTokenExecuteGitHub(wID, wToken, githubEventName string, data []byte) (err error) {
	// Header: X-Github-EventCopy
	// Body: application/octet-stream
	return fmt.Errorf("not implemented")
}

// GroupCreate creates a group based on the data provided
// "Users" field is a list of user IDs that will be in the group
func (s *Session) GroupCreate(data GroupCreateData) (group *Group, err error) {
	endpoint := EndpointChannel("create")
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &group)
	return
}

func (s *Session) GroupMemberAdd(cID, mID string) (err error) {
	endpoint := EndpointChannelRecipients(cID, mID)
	err = s.HTTP.Request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) GroupMemberDelete(cID, mID string) (err error) {
	endpoint := EndpointChannelRecipients(cID, mID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) GroupMembers(cID string) (users []*User, err error) {
	endpoint := EndpointChannelMembers(cID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &users)
	return
}

func (s *Session) ChannelInviteCreate(cID string) (invite *InviteCreate, err error) {
	endpoint := EndpointChannelInvites(cID)
	err = s.HTTP.Request(http.MethodPost, endpoint, nil, &invite)
	return
}

func (s *Session) ChannelDelete(cID string) (err error) {
	endpoint := EndpointChannel(cID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) MessageAck(channelID, messageID string) (err error) {
	endpoint := EndpointChannelAckMessage(channelID, messageID)
	err = s.HTTP.Request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) ServerBans(sID string) (bans []*ServerBans, err error) {
	endpoint := EndpointServerBans(sID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &bans)
	return
}

func (s *Session) ServerAck(serverID string) (err error) {
	endpoint := EndpointServerAck(serverID)
	err = s.HTTP.Request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) ServerInvites(sID string) (invites []*Invite, err error) {
	endpoint := EndpointServerInvites(sID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &invites)
	return
}

func (s *Session) ServersRole(sID, rID string) (role *ServerRole, err error) {
	endpoint := EndpointServerRole(sID, rID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &role)
	return
}

func (s *Session) Invite(iID string) (invite *Invite, err error) {
	endpoint := EndpointInvite(iID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &invite)
	return
}

func (s *Session) InviteJoin(iID string) (invite *Invite, err error) {
	endpoint := EndpointInvite(iID)
	err = s.HTTP.Request(http.MethodPost, endpoint, nil, &invite)
	return
}

func (s *Session) InviteDelete(iID string) (err error) {
	endpoint := EndpointInvite(iID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerRoleDelete(sID, rID string) (err error) {
	endpoint := EndpointServerRole(sID, rID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerRoleEdit(sID, rID string, data ServerRoleEditData) (role *ServerRole, err error) {
	endpoint := EndpointServerRole(sID, rID)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &role)
	return
}

func (s *Session) ServerEmojis(sID string) (emojis []*Emoji, err error) {
	endpoint := EndpointServerEmojis(sID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &emojis)
	return
}

func (s *Session) ServersRoleRanksEdit(sID string, ranks []string) (err error) {
	endpoint := EndpointServerRolesRanks(sID)
	err = s.HTTP.Request(http.MethodPatch, endpoint, ranks, nil)
	return
}

func (s *Session) ServersRoleCreate(sID string, data ServerRoleCreateData) (role *ServerRole, err error) {
	endpoint := EndpointServerRoles(sID)
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &role)
	return
}

func (s *Session) PermissionsSet(sID, rID string, data PermissionOverwrite) (err error) {
	endpoint := EndpointServerPermissions(sID, rID)
	err = s.HTTP.Request(http.MethodPut, endpoint, data, nil)
	return
}

// ChannelPermissionsSet sets permissions for the specified role in this channel.
func (s *Session) ChannelPermissionsSet(cID, rID string, data PermissionOverwrite) (err error) {
	endpoint := EndpointChannelPermission(cID, rID)
	return s.HTTP.Request(http.MethodPut, endpoint, data, nil)
}

// ChannelPermissionsSetDefault sets permissions for the default role in this channel.
func (s *Session) ChannelPermissionsSetDefault(cID string, data PermissionOverwrite) (err error) {
	return s.ChannelPermissionsSet(cID, "default", data)
}

// PermissionsSetDefault sets the permissions of a role in a server
func (s *Session) PermissionsSetDefault(sID string, data PermissionsSetDefaultData) (err error) {
	endpoint := EndpointServerPermissions(sID, "default")
	err = s.HTTP.Request(http.MethodPut, endpoint, data, nil)
	return
}

func (s *Session) ChannelEdit(cID string, data ChannelEditData) (channel *Channel, err error) {
	endpoint := EndpointChannel(cID)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &channel)
	return
}

func (s *Session) ServerMemberUnban(sID, mID string) (err error) {
	endpoint := EndpointServerBan(sID, mID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberBan(sID, mID string) (err error) {
	endpoint := EndpointServerBan(sID, mID)
	err = s.HTTP.Request(http.MethodPut, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberDelete(sID, mID string) (err error) {
	endpoint := EndpointServerMember(sID, mID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

func (s *Session) ServerMemberEdit(sID, mID string, data ServerMemberEditData) (member *ServerMember, err error) {
	endpoint := EndpointServerMember(sID, mID)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &member)
	return
}

func (s *Session) ServerMember(sID, mID string) (member *ServerMember, err error) {
	endpoint := EndpointServerMember(sID, mID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &member)
	s.State.addServerMember(member)
	return
}

func (s *Session) ServerMembers(sID string) (members *ServerMembers, err error) {
	endpoint := EndpointServerMembers(sID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &members)
	s.State.addServerMembersAndUsers(members)
	return
}

func (s *Session) ChannelMessage(cID, mID string) (message *Message, err error) {
	endpoint := EndpointChannelMessage(cID, mID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &message)
	return
}

// ChannelMessageReactionCreate adds a reaction (emoji ID) to a message
func (s *Session) ChannelMessageReactionCreate(cID, mID, eID string) (err error) {
	endpoint := EndpointChannelMessageReaction(cID, mID, eID)
	err = s.HTTP.Request(http.MethodPut, endpoint, nil, nil)
	return
}

// ChannelMessageReactionDelete deletes a singular reaction on a message
func (s *Session) ChannelMessageReactionDelete(cID, mID, eID string) (err error) {
	endpoint := EndpointChannelMessageReaction(cID, mID, eID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
	return
}

// ChannelMessageReactionClear clears all reactions on a message
func (s *Session) ChannelMessageReactionClear(cID, mID string) (err error) {
	endpoint := EndpointChannelMessageReactions(cID, mID)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

// ChannelsJoinCall asks the voice server for a token to join the call.
func (s *Session) ChannelsJoinCall(cID string, data ChannelJoinCallData) (call ChannelJoinCall, err error) {
	endpoint := EndpointChannelJoinCall(cID)
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &call)
	return
}

// ChannelsEndRing stops ringing a user in a DM if a call exists; returns NotConnected otherwise.
// Only works within DMs and groups; returns NoEffect in servers.
// Returns NotFound if the user is not in the DM/group channel.
func (s *Session) ChannelsEndRing(cID, uID string) error {
	endpoint := EndpointChannelEndRing(cID, uID)
	return s.HTTP.Request(http.MethodPut, endpoint, nil, nil)
}

func (s *Session) ServerChannelCreate(sID string, data ServerChannelCreateData) (channel *Channel, err error) {
	endpoint := EndpointServerChannels(sID)
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &channel)
	return
}

func (s *Session) ServerDelete(sID string) error {
	endpoint := EndpointServer(sID)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) ChannelMessages(cID string, params ...ChannelMessagesParams) (messages ChannelMessages, err error) {

	/*
		This method is special. It has to deal with the following bullshit:
		  - If NOT ChannelMessagesParams (thus IncludeUsers is false), API returns -> []*Message
		  - If ChannelMessagesParams with IncludeUsers=true, API returns -> { []*Message, []*User []*Member }

		This method will try to normalise it into ChannelMessages struct for simplicity
	*/

	// todo: maybe move this to EndpointChannelMessages to process?
	// for now it's an easy monkeypatch to just add "?" before the params

	endpoint := EndpointChannelMessages(cID)
	hasParams := len(params) > 0

	if hasParams {
		endpoint = fmt.Sprintf("%s?%s", endpoint, params[0].Encode())
		if params[0].IncludeUsers {
			err = s.HTTP.Request(http.MethodGet, endpoint, nil, &messages)
			return
		}
	}

	var intermediary []*Message
	if err = s.HTTP.Request(http.MethodGet, endpoint, nil, &intermediary); err == nil {
		messages.Messages = intermediary
	}

	return
}

func (s *Session) ChannelMessageEdit(cID, mID string, data MessageEditData) (message *Message, err error) {
	endpoint := EndpointChannelMessage(cID, mID)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &message)
	return
}

func (s *Session) ChannelMessageSend(cID string, data MessageSend) (message *Message, err error) {
	endpoint := EndpointChannelMessages(cID)
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &message)
	return
}

func (s *Session) ChannelMessageDelete(cID, mID string) error {
	endpoint := EndpointChannelMessage(cID, mID)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) ChannelMessageDeleteBulk(cID string, messages ChannelMessageBulkDeleteData) error {
	endpoint := EndpointChannelMessage(cID, "bulk")
	return s.HTTP.Request(http.MethodDelete, endpoint, messages, nil)
}

func (s *Session) AccountCreate(data AccountCreateData) error {
	endpoint := EndpointAuthAccount("create")
	return s.HTTP.Request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) AccountReverify(data AccountReverifyData) error {
	endpoint := EndpointAuthAccount("reverify")
	return s.HTTP.Request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) AccountDeleteConfirm(data AccountDeleteConfirmData) error {
	endpoint := EndpointAuthAccount("delete")
	return s.HTTP.Request(http.MethodPut, endpoint, data, nil)
}

func (s *Session) AccountDelete() error {
	endpoint := EndpointAuthAccount("delete")
	return s.HTTP.Request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) Account() (account *Account, err error) {
	endpoint := EndpointAuthAccount("")
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &account)
	return
}

func (s *Session) AccountDisable() error {
	endpoint := EndpointAuthAccount("disable")
	return s.HTTP.Request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) AccountChangePassword(data AccountChangePasswordData) error {
	endpoint := EndpointAuthAccountChange("password")
	return s.HTTP.Request(http.MethodPatch, endpoint, data, nil)
}

func (s *Session) AccountChangeEmail(data AccountChangeEmailData) error {
	endpoint := EndpointAuthAccountChange("email")
	return s.HTTP.Request(http.MethodPatch, endpoint, data, nil)
}

func (s *Session) VerifyEmail(code string) (ticket ChangeEmail, err error) {
	endpoint := EndpointAuthAccountVerify(code)
	err = s.HTTP.Request(http.MethodPost, endpoint, nil, &ticket)
	return
}

// PasswordReset requests a password reset, which is sent to the email provided
func (s *Session) PasswordReset(data AccountReverifyData) error {
	endpoint := EndpointAuthAccount("reset_password")
	return s.HTTP.Request(http.MethodPost, endpoint, data, nil)
}

// PasswordResetConfirm confirms a password reset
func (s *Session) PasswordResetConfirm(data PasswordResetConfirmData) error {
	endpoint := EndpointAuthAccount("reset_password")
	return s.HTTP.Request(http.MethodPatch, endpoint, data, nil)
}

// Login as a regular user instead of bot. Friendly name is used to identify the session via MFA
func (s *Session) Login(data LoginData) (mfa LoginResponse, err error) {
	endpoint := EndpointAuthSession("login")
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &mfa)
	return
}

func (s *Session) Sessions() (sessions []*Sessions, err error) {
	endpoint := EndpointAuthSession("all")
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &sessions)
	return
}

func (s *Session) SessionEdit(id string, data SessionEditData) (session SessionEditData, err error) {
	endpoint := EndpointAuthSession(id)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &session)
	return
}

// Onboarding returns whether the current account requires onboarding or whether you can continue to send requests as usual
func (s *Session) Onboarding() (onboarding Onboarding, err error) {
	endpoint := EndpointOnboard("hello")
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &onboarding)
	return
}

// OnboardingComplete sets a new username, completes onboarding and allows a user to start using Revolt.
func (s *Session) OnboardingComplete(data OnboardingCompleteData) error {
	endpoint := EndpointOnboard("complete")
	return s.HTTP.Request(http.MethodPost, endpoint, data, nil)
}

// SessionsDelete invalidates a session with the provided ID
func (s *Session) SessionsDelete(id string) error {
	endpoint := EndpointAuthSession(id)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

// SessionsDeleteAll invalidates all sessions, including this one if revokeSelf is true
func (s *Session) SessionsDeleteAll(revokeSelf bool) error {
	endpoint := EndpointAuthSession("all")
	if revokeSelf {
		values := url.Values{}
		values.Set("revoke_self", "true")
		endpoint += fmt.Sprintf("?%s", values.Encode())
	}

	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) Logout() error {
	endpoint := EndpointAuthSession("logout")
	return s.HTTP.Request(http.MethodPost, endpoint, nil, nil)
}

func (s *Session) UserMutual(uID string) (mutual []*MutualFriendsAndServersResponse, err error) {
	endpoint := EndpointUserMutual(uID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &mutual)
	return
}

// DirectMessages returns a list of direct message channels.
func (s *Session) DirectMessages() (channels []*Channel, err error) {
	endpoint := EndpointUser("dms")
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &channels)
	return
}

// DirectMessageCreate opens a direct message channel with a user
// Will return an error "MissingPermission" "SendMessage" if you are not friends or blocked
func (s *Session) DirectMessageCreate(uID string) (channel *Channel, err error) {
	endpoint := EndpointUserDM(uID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &channel)
	return
}

// Relationships returns a list of relationships for the current user
func (s *Session) Relationships() (relationships []*UserRelationship, err error) {
	endpoint := URLUserRelationships
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &relationships)
	return
}

// FriendAdd sends or accepts a friend Request.
func (s *Session) FriendAdd(uID string) (user *User, err error) {
	endpoint := EndpointUserFriend(uID)
	err = s.HTTP.Request(http.MethodPut, endpoint, nil, &user)
	return
}

// FriendDelete removes a friend or declines a friend Request.
func (s *Session) FriendDelete(uID string) (user *User, err error) {
	endpoint := EndpointUserFriend(uID)
	err = s.HTTP.Request(http.MethodDelete, endpoint, nil, &user)
	return
}

// Bot fetches details of a bot you own by its ID
func (s *Session) Bot(bID string) (bot *FetchedBot, err error) {
	endpoint := EndpointBot(bID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &bot)
	return
}

// Bots returns a list of bots for the current user
func (s *Session) Bots() (bots *FetchedBots, err error) {
	endpoint := EndpointBot("@me")
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &bots)
	return
}

// BotCreate creates a bot based on the data provided
func (s *Session) BotCreate(data BotCreateData) (bot *Bot, err error) {
	endpoint := EndpointBot("create")
	err = s.HTTP.Request(http.MethodPost, endpoint, data, &bot)
	return
}

func (s *Session) BotEdit(id string, data BotEditData) (bot *Bot, err error) {
	endpoint := EndpointBot(id)
	err = s.HTTP.Request(http.MethodPatch, endpoint, data, &bot)
	return
}

func (s *Session) BotDelete(bID string) error {
	endpoint := EndpointBot(bID)
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

// BotPublic fetches a public bot by its ID
func (s *Session) BotPublic(bID string) (bot *PublicBot, err error) {
	endpoint := EndpointBotInvite(bID)
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &bot)
	return
}

// BotInvite invites a bot by its ID to a server or group
func (s *Session) BotInvite(bID string, data BotInviteData) (err error) {
	endpoint := EndpointBotInvite(bID)
	err = s.HTTP.Request(http.MethodPost, endpoint, data, nil)
	return
}

func (s *Session) SyncUnreads() (data []SyncUnread, err error) {
	endpoint := EndpointSync("unreads")
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &data)
	return
}

func (s *Session) SyncSettingsFetch(payload SyncSettingsFetchData) (data *SyncSettingsData, err error) {
	endpoint := EndpointSync("settings")
	err = s.HTTP.Request(http.MethodPost, endpoint, payload, &data)
	return
}

func (s *Session) SyncSettingsSet(payload SyncSettingsData) error {
	endpoint := EndpointSync("settings")
	return s.HTTP.Request(http.MethodPost, endpoint, payload, nil)
}

func (s *Session) PushSubscribe(data WebpushSubscription) error {
	endpoint := EndpointPush("subscribe")
	return s.HTTP.Request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) PushUnsubscribe(data WebpushSubscription) error {
	endpoint := EndpointPush("unsubscribe")
	return s.HTTP.Request(http.MethodPost, endpoint, data, nil)
}

func (s *Session) AuthMFA() (mfa AuthMFAResponse, err error) {
	endpoint := EndpointAuthMFA("")
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &mfa)
	return
}

func (s *Session) AuthMFACreateTicket(data AuthMFAData) (ticket AuthMFATicketResponse, err error) {
	endpoint := EndpointAuthMFA("ticket")
	err = s.HTTP.Request(http.MethodPut, endpoint, data, &ticket)
	return
}

func (s *Session) AuthMFARecoveryCodes() (codes []string, err error) {
	endpoint := EndpointAuthMFA("recovery")
	err = s.HTTP.Request(http.MethodPost, endpoint, nil, &codes)
	return
}

func (s *Session) AuthMFAGenerateRecoveryCodes() (codes []string, err error) {
	endpoint := EndpointAuthMFA("recovery")
	err = s.HTTP.Request(http.MethodPatch, endpoint, nil, &codes)
	return
}

func (s *Session) AuthMFAMethods() (methods []AuthMFAMethod, err error) {
	endpoint := EndpointAuthMFA("methods")
	err = s.HTTP.Request(http.MethodGet, endpoint, nil, &methods)
	return
}

func (s *Session) AuthMFAEnable2FATOTP(data AuthMFAData) (err error) {
	endpoint := EndpointAuthMFA("totp")
	return s.HTTP.Request(http.MethodPut, endpoint, data, nil)
}

func (s *Session) AuthMFADisable2FATOTP() (err error) {
	endpoint := EndpointAuthMFA("totp")
	return s.HTTP.Request(http.MethodDelete, endpoint, nil, nil)
}

func (s *Session) AuthMFAGenerateTOTPSecret() (secret AuthMFATOTPSecretResponse, err error) {
	endpoint := EndpointAuthMFA("totp")
	err = s.HTTP.Request(http.MethodPost, endpoint, nil, &secret)
	return
}

func (s *Session) PolicyAck() (err error) {
	endpoint := EndpointPolicy("acknowledge")
	return s.HTTP.Request(http.MethodPost, endpoint, nil, nil)
}
