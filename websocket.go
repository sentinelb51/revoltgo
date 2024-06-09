package revoltgo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/lxzan/gws"
	"log"
	"time"
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
	fmt.Println(string(message.Bytes()))
	handle(ws.Session, message.Data.Bytes())
	message.Close()
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[R-GO] > ")
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
	s.Socket.WriteClose(1000, nil)
	return nil
	// return s.Socket.Close()
}

// ping pings the websocket every HeartbeatInterval interval
// It keeps the websocket connection alive, and triggers a re-connect if a problem occurs
//func (s *Session) ping() {
//
//	for s.Connected {
//		time.Sleep(s.HeartbeatInterval)
//
//		ping := WebsocketMessagePing{
//			Type: WebsocketMessageTypeHeartbeat,
//			Data: s.heartbeatCount,
//		}
//
//		err := s.WriteSocket(ping)
//		if err != nil {
//			log.Printf("heartbeat failed: %s\n", err)
//			break
//		}
//
//		s.LastHeartbeatSent = time.Now()
//	}
//
//	s.Connected = false
//	log.Println("triggering reconnect...")
//
//	for !s.Connected {
//		err := s.Open()
//		if err != nil {
//			log.Printf("reconnect failed: %v\n", err)
//			log.Printf("retrying in %.f seconds...\n", s.ReconnectInterval.Seconds())
//			time.Sleep(s.ReconnectInterval)
//		}
//	}
//}

func handle(s *Session, raw []byte) {
	var data Event
	err := json.Unmarshal(raw, &data)
	if err != nil {
		log.Printf("handle: %v", err)
		return
	}

	eventConstructor, ok := eventToStruct[data.Type]
	if !ok {
		log.Printf("unknown event type: %s", data.Type)
		return
	}

	event := eventConstructor()
	if err = json.Unmarshal(raw, &event); err != nil {
		log.Printf("unmarshal event: %s: %s", string(raw), err)
		return
	}

	switch e := event.(type) {
	case *EventError:
		log.Panicf("authentication error: %s\n", e.Error)
	case *EventBulk:
		for _, event := range e.V {
			handle(s, event)
		}
	case *EventPong:
		if e.Data != s.HeartbeatCount {
			log.Printf("heartbeat fibrillation %d != %d\n", e.Data, s.HeartbeatCount)
			break
		}

		s.HeartbeatCount++
		s.LastHeartbeatAck = time.Now()

		for _, h := range s.handlersPong {
			h(s, e)
		}
	case *AbstractEventUpdate:

		switch e.Type {
		case "ServerUpdate":
			s.State.updateServer(e)
			for _, h := range s.handlersServerUpdate {
				h(s, e.EventServerUpdate())
			}
		case "ServerMemberUpdate":
			s.State.updateServerMember(e)
			for _, h := range s.handlersServerMemberUpdate {
				h(s, e.EventServerMemberUpdate())
			}
		case "ChannelUpdate":
			s.State.updateChannel(e)
			for _, h := range s.handlersChannelUpdate {
				h(s, e.EventChannelUpdate())
			}
		case "UserUpdate":
			s.State.updateUser(e)
			for _, h := range s.handlersUserUpdate {
				h(s, e.EventUserUpdate())
			}
		case "ServerRoleUpdate":
			s.State.updateServerRole(e)
			for _, h := range s.handlersServerRoleUpdate {
				h(s, e.EventServerRoleUpdate())
			}
		case "WebhookUpdate":
			s.State.updateWebhook(e)
			for _, h := range s.handlersWebhookUpdate {
				h(s, e.EventWebhookUpdate())
			}
		}
	case *EventAuthenticated:
		for _, h := range s.handlersAuthenticated {
			h(s, e)
		}
	case *EventAuth:
		for _, h := range s.handlersAuth {
			h(s, e)
		}
	case *EventReady:
		s.State.populate(e)
		s.Selfbot = s.State.Self != nil && s.State.Self.Bot == nil
		for _, h := range s.handlersReady {
			h(s, e)
		}
	case *EventMessage:
		for _, h := range s.handlersMessage {
			h(s, e)
		}
	case *EventMessageAppend:
		for _, h := range s.handlersMessageAppend {
			h(s, e)
		}
	case *EventMessageUpdate:
		for _, h := range s.handlersMessageUpdate {
			h(s, e)
		}
	case *EventMessageDelete:
		for _, h := range s.handlersMessageDelete {
			h(s, e)
		}
	case *EventMessageReact:
		for _, h := range s.handlersMessageReact {
			h(s, e)
		}
	case *EventMessageUnreact:
		for _, h := range s.handlersMessageUnreact {
			h(s, e)
		}
	case *EventChannelCreate:
		s.State.createChannel(e)
		for _, h := range s.handlersChannelCreate {
			h(s, e)
		}
	case *EventChannelDelete:
		s.State.deleteChannel(e)
		for _, h := range s.handlersChannelDelete {
			h(s, e)
		}
	case *EventChannelGroupJoin:
		for _, h := range s.handlersGroupJoin {
			h(s, e)
		}
	case *EventChannelGroupLeave:
		for _, h := range s.handlersGroupLeave {
			h(s, e)
		}
	case *EventChannelStartTyping:
		for _, h := range s.handlersChannelStartTyping {
			h(s, e)
		}
	case *EventChannelStopTyping:
		for _, h := range s.handlersChannelStopTyping {
			h(s, e)
		}
	case *EventServerCreate:
		s.State.createServer(e)
		for _, h := range s.handlersServerCreate {
			h(s, e)
		}
	case *EventServerDelete:
		s.State.deleteServer(e)
		for _, h := range s.handlersServerDelete {
			h(s, e)
		}
	case *EventServerMemberJoin:
		s.State.createServerMember(e)
		for _, h := range s.handlersServerMemberJoin {
			h(s, e)
		}
	case *EventServerMemberLeave:
		s.State.deleteServerMember(e)
		for _, h := range s.handlersServerMemberLeave {
			h(s, e)
		}
	case *EventChannelAck:
		for _, h := range s.handlersChannelAck {
			h(s, e)
		}
	case *EventServerRoleDelete:
		s.State.deleteServerRole(e)
		for _, h := range s.handlersServerRoleDelete {
			h(s, e)
		}
	case *EventEmojiCreate:
		s.State.createEmoji(e)
		for _, h := range s.handlersEmojiCreate {
			h(s, e)
		}
	case *EventEmojiDelete:
		s.State.deleteEmoji(e)
		for _, h := range s.handlersEmojiDelete {
			h(s, e)
		}
	case *EventUserSettingsUpdate:
		for _, h := range s.handlersUserSettingsUpdate {
			h(s, e)
		}
	case *EventUserRelationship:
		for _, h := range s.handlersUserRelationship {
			h(s, e)
		}
	case *EventUserPlatformWipe:
		s.State.platformWipe(e)
		for _, h := range s.handlersUserPlatformWipe {
			h(s, e)
		}
	case *EventWebhookCreate:
		s.State.createWebhook(e)
		for _, h := range s.handlersWebhookCreate {
			h(s, e)
		}
	case *EventWebhookDelete:
		s.State.deleteWebhook(e)
		for _, h := range s.handlersWebhookDelete {
			h(s, e)
		}
	}
}
