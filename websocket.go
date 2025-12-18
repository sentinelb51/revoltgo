package revoltgo

import (
	"bytes"
	"encoding/binary"
	"log"
	"time"

	"github.com/goccy/go-json"
	"github.com/lxzan/gws"
)

type WebsocketMessageType string

const (
	WebsocketKeepAlivePeriod = 60 * time.Second

	WebsocketMessageTypeAuthenticate WebsocketMessageType = "Authenticate"
	WebsocketMessageTypeHeartbeat    WebsocketMessageType = "Ping"
	WebsocketMessageTypeBeginTyping  WebsocketMessageType = "BeginTyping"
	WebsocketMessageTypeEndTyping    WebsocketMessageType = "EndTyping"
)

type websocket struct {
	URL     string
	Session *Session
	Close   chan struct{}
}

func (ws *websocket) heartbeat(session *Session, socket *gws.Conn) {

	var (
		ticker = time.NewTicker(session.HeartbeatInterval)
		err    error
	)

	defer ticker.Stop()
	for session.Connected {
		select {
		case <-ticker.C:
			payload := bytes.NewBuffer(make([]byte, 0, 64))
			err = binary.Write(payload, binary.LittleEndian, session.HeartbeatCount)

			if err != nil {
				log.Printf("Heartbeat stopped: %s\n", err)
				session.Connected = false
				break
			}

			if err = socket.WritePing(payload.Bytes()); err != nil {
				log.Printf("Heartbeat stopped: %s\n", err)
				session.Connected = false
				break
			}

			session.LastHeartbeatSent = time.Now()
		case <-ws.Close:
			session.Connected = false
		}
	}

	for !session.Connected && session.ShouldReconnect {
		log.Printf("Re-connecting in %s...", session.ReconnectInterval.String())
		time.Sleep(session.ReconnectInterval)
		connect(session, ws.URL)
	}
}

// connect creates a new websocket connection to the given URL.
func connect(session *Session, url string) *gws.Conn {

	log.Println("Connecting...")

	handler := &websocket{
		URL:     url,
		Session: session,
	}

	options := &gws.ClientOption{
		Addr:             url,
		ParallelEnabled:  true,
		CheckUtf8Enabled: false,
	}

	if session.CustomCompression != nil {
		options.PermessageDeflate = *session.CustomCompression
	} else {
		options.PermessageDeflate = gws.PermessageDeflate{
			Enabled:               true,
			ServerContextTakeover: true,
			ClientContextTakeover: true,
		}
	}

	socket, _, err := gws.NewClient(handler, options)
	if err != nil {
		log.Panicf("Connection refused: %s\n", err)
	}

	session.Connected = true
	go socket.ReadLoop()
	return socket
}

func (ws *websocket) OnClose(_ *gws.Conn, err error) {

	ws.Close <- struct{}{}
	ws.Session.Connected = false

	if reason := err.Error(); reason == "" {
		log.Printf("Connection closed unexpectedly: %v (%d)\n", reason, len(reason))
		return
	}

	log.Println("Connection closed.")
}

// OnPong ensures the pong is valid, updates heartbeat times, and extends the connection deadline.
func (ws *websocket) OnPong(socket *gws.Conn, payload []byte) {

	var (
		count int64
		err   error
	)

	if err = binary.Read(bytes.NewReader(payload), binary.LittleEndian, &count); err != nil {
		log.Printf("Pong: read count: %s\n", err)
		return
	}

	if count != ws.Session.HeartbeatCount {
		log.Printf("Heartbeat fibrillation: %d != %d\n", count, ws.Session.HeartbeatCount)
		return
	}

	now := time.Now()
	ws.Session.LastHeartbeatAck = now
	ws.Session.HeartbeatCount++

	deadline := now.Add(ws.Session.HeartbeatInterval * 2)
	if err = socket.SetDeadline(deadline); err != nil {
		log.Printf("Pong: set deadline: %s\n", err)
		return
	}
}

func (ws *websocket) OnOpen(socket *gws.Conn) {

	ws.Session.HeartbeatCount = 0

	if err := socket.SetDeadline(time.Now().Add(WebsocketKeepAlivePeriod * 2)); err != nil {
		log.Printf("Open: set deadline: %s\n", err)
	}

	if ws.Session.HeartbeatInterval.Seconds() >= 99 {
		log.Printf("Heartbeat interval (%s) too high, and may cause disconnects\n",
			ws.Session.HeartbeatInterval.String())
	}

	log.Printf("Connected (%s)\n", socket.RemoteAddr())
	go ws.heartbeat(ws.Session, socket)
}

func (ws *websocket) OnPing(_ *gws.Conn, payload []byte) {
	// The websocket should not be pinging us; we're the client.
	log.Printf("Received unexpected ping: %s\n", string(payload))
}

func (ws *websocket) OnMessage(_ *gws.Conn, message *gws.Message) {
	handle(ws.Session, message.Data.Bytes())
	if err := message.Close(); err != nil {
		log.Printf("Message close: %s\n", err)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[R-GO] ")
}

type WebsocketMessageAuthenticate struct {
	Type  WebsocketMessageType `json:"type"`
	Token string               `json:"token"`
}

type WebsocketMessagePing struct {
	Type WebsocketMessageType `json:"type"`
	Data int64                `json:"data"`
}

type WebsocketChannelTyping struct {
	Type    WebsocketMessageType `json:"type"`
	Channel string               `json:"channel"`
}

// Close the websocket.
func (s *Session) Close() error {
	s.Connected = false
	return s.Socket.WriteClose(1000, nil)
}

func handle(s *Session, raw []byte) {

	eventType, err := eventTypeFromJSON(raw)
	if err != nil {
		log.Printf("event type detection failed: %s\n", err)
		return
	}

	eventConstructor, ok := eventToStruct[eventType]
	if !ok {
		log.Printf("unknown event type: %s", eventType)
		return
	}

	switch eventType {
	case "Error":

		if len(s.handlersError) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersError {
			h(s, event.(*EventError))
		}
	case "Bulk":

		if len(s.handlersBulk) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersBulk {
			h(s, event.(*EventBulk))
		}
	case "Pong":

		if len(s.handlersPong) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersPong {
			h(s, event.(*EventPong))
		}

	case
		"MessageUpdate",
		"ServerUpdate",
		"ChannelUpdate",
		"ServerRoleUpdate",
		"WebhookUpdate",
		"UserUpdate",
		"ServerMemberUpdate":

		// note: if this is empty, none of the above listed events will dispatch.
		// bad design.
		if len(s.handlersAbstractEventUpdate) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		aeu := event.(*AbstractEventUpdate)
		for _, h := range s.handlersAbstractEventUpdate {
			h(s, aeu)
		}
	case "Authenticated":

		if len(s.handlersAuthenticated) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersAuthenticated {
			h(s, event.(*EventAuthenticated))
		}
	case "Auth":

		if len(s.handlersAuth) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersAuth {
			h(s, event.(*EventAuth))
		}
	case "Ready":

		if len(s.handlersReady) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersReady {
			h(s, event.(*EventReady))
		}
	case "Message":

		if len(s.handlersMessage) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersMessage {
			h(s, event.(*EventMessage))
		}
	case "MessageAppend":

		if len(s.handlersMessageAppend) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersMessageAppend {
			h(s, event.(*EventMessageAppend))
		}
	case "MessageDelete":

		if len(s.handlersMessageDelete) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersMessageDelete {
			h(s, event.(*EventMessageDelete))
		}
	case "MessageReact":

		if len(s.handlersMessageReact) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersMessageReact {
			h(s, event.(*EventMessageReact))
		}
	case "MessageUnreact":

		if len(s.handlersMessageUnreact) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersMessageUnreact {
			h(s, event.(*EventMessageUnreact))
		}
	case "ChannelCreate":

		if len(s.handlersChannelCreate) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersChannelCreate {
			h(s, event.(*EventChannelCreate))
		}
	case "ChannelDelete":

		if len(s.handlersChannelDelete) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersChannelDelete {
			h(s, event.(*EventChannelDelete))
		}
	case "ChannelGroupJoin":

		if len(s.handlersGroupJoin) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersGroupJoin {
			h(s, event.(*EventChannelGroupJoin))
		}
	case "ChannelGroupLeave":

		if len(s.handlersGroupLeave) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersGroupLeave {
			h(s, event.(*EventChannelGroupLeave))
		}
	case "ChannelStartTyping":

		if len(s.handlersChannelStartTyping) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersChannelStartTyping {
			h(s, event.(*EventChannelStartTyping))
		}
	case "ChannelStopTyping":

		if len(s.handlersChannelStopTyping) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersChannelStopTyping {
			h(s, event.(*EventChannelStopTyping))
		}
	case "ServerCreate":

		if len(s.handlersServerCreate) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersServerCreate {
			h(s, event.(*EventServerCreate))
		}
	case "ServerDelete":

		if len(s.handlersServerDelete) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersServerDelete {
			h(s, event.(*EventServerDelete))
		}
	case "ServerMemberJoin":

		if len(s.handlersServerMemberJoin) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersServerMemberJoin {
			h(s, event.(*EventServerMemberJoin))
		}
	case "ServerMemberLeave":

		if len(s.handlersServerMemberLeave) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersServerMemberLeave {
			h(s, event.(*EventServerMemberLeave))
		}
	case "ChannelAck":

		if len(s.handlersChannelAck) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersChannelAck {
			h(s, event.(*EventChannelAck))
		}
	case "ServerRoleDelete":

		if len(s.handlersServerRoleDelete) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersServerRoleDelete {
			h(s, event.(*EventServerRoleDelete))
		}
	case "EmojiCreate":

		if len(s.handlersEmojiCreate) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersEmojiCreate {
			h(s, event.(*EventEmojiCreate))
		}
	case "EmojiDelete":

		if len(s.handlersEmojiDelete) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersEmojiDelete {
			h(s, event.(*EventEmojiDelete))
		}
	case "UserSettingsUpdate":

		if len(s.handlersUserSettingsUpdate) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersUserSettingsUpdate {
			h(s, event.(*EventUserSettingsUpdate))
		}
	case "UserRelationship":

		if len(s.handlersUserRelationship) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersUserRelationship {
			h(s, event.(*EventUserRelationship))
		}
	case "UserPlatformWipe":

		if len(s.handlersUserPlatformWipe) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersUserPlatformWipe {
			h(s, event.(*EventUserPlatformWipe))
		}
	case "WebhookCreate":

		if len(s.handlersWebhookCreate) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersWebhookCreate {
			h(s, event.(*EventWebhookCreate))
		}
	case "WebhookDelete":

		if len(s.handlersWebhookDelete) == 0 {
			return
		}

		event := eventConstructor()
		if err = json.Unmarshal(raw, &event); err != nil {
			log.Printf("unmarshal event: %s: %s", string(raw), err)
			return
		}

		for _, h := range s.handlersWebhookDelete {
			h(s, event.(*EventWebhookDelete))
		}
	}
}
