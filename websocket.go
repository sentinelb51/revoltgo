package revoltgo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net/http"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[R-GO] ")
}

type WebsocketMessageType string

const (
	WebsocketMessageTypeAuthenticate WebsocketMessageType = "Authenticate"
	WebsocketMessageTypeHeartbeat    WebsocketMessageType = "Ping"
	WebsocketMessageTypeBeginTyping  WebsocketMessageType = "BeginTyping"
	WebsocketMessageTypeEndTyping    WebsocketMessageType = "EndTyping"
)

type WebsocketMessageAuthenticate struct {
	Type  WebsocketMessageType `json:"type"`
	Token string               `json:"token"`
}

type WebsocketMessagePing struct {
	Type WebsocketMessageType `json:"type"`
	Data int                  `json:"data"`
}

type WebsocketChannelTyping struct {
	Type    WebsocketMessageType `json:"type"`
	Channel string               `json:"channel"`
}

// Open initiates the websocket connection to the Revolt API
func (s *Session) Open() (err error) {

	if s.Connected {
		return fmt.Errorf("already connected")
	}

	// Determine the websocket URL
	var query RevoltAPI
	err = s.request(http.MethodGet, baseURL, nil, &query)
	if err != nil {
		return
	}

	dialer := ws.Dialer{
		Timeout: s.ReconnectInterval,
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.ReconnectInterval)
	defer cancel()

	connection, _, _, err := dialer.Dial(ctx, query.Ws)
	if err != nil {
		return
	}

	s.Socket = connection

	wsAuth := WebsocketMessageAuthenticate{
		Type:  WebsocketMessageTypeAuthenticate,
		Token: s.Token,
	}

	// Send an initial authentication message
	err = s.WriteSocket(wsAuth)
	if err != nil {
		return err
	}

	// Assume we have a successful connection, until we don't
	s.Connected = true
	go s.listen()
	return
}

// listen reads messages from the websocket
func (s *Session) listen() {
	for s.Connected {
		message, op, err := wsutil.ReadServerData(s.Socket)
		if err != nil {
			log.Printf("listen error: %s\n", err)
			s.Connected = false
			break
		}

		if op != ws.OpText {
			continue
		}

		go s.handle(message)
	}
}

// Close the websocket.
func (s *Session) Close() error {
	s.Connected = false
	return s.Socket.Close()
}

// ping pings the websocket every HeartbeatInterval interval
// It keeps the websocket connection alive, and triggers a re-connect if a problem occurs
func (s *Session) ping() {

	for s.Connected {
		time.Sleep(s.HeartbeatInterval)

		ping := WebsocketMessagePing{
			Type: WebsocketMessageTypeHeartbeat,
			Data: s.heartbeatCount,
		}

		err := s.WriteSocket(ping)
		if err != nil {
			log.Printf("heartbeat failed: %s\n", err)
			break
		}

		s.LastHeartbeatSent = time.Now()
	}

	s.Connected = false
	log.Println("triggering reconnect...")

	for !s.Connected {
		err := s.Open()
		if err != nil {
			log.Printf("reconnect failed: %v\n", err)
			log.Printf("retrying in %.f seconds...\n", s.ReconnectInterval.Seconds())
			time.Sleep(s.ReconnectInterval)
		}
	}
}

func (s *Session) handle(raw []byte) {
	var data Event
	err := json.Unmarshal(raw, &data)
	if err != nil {
		log.Printf("handle: %v", err)
		return
	}

	eventConstructor, ok := eventToStruct[data.Type]
	if !ok {
		log.Printf("unknown event type: %s", data.Type)
		for _, h := range s.HandlersUnknown {
			h(s, string(raw))
		}
		return
	}

	event := eventConstructor()
	if err = json.Unmarshal(raw, &event); err != nil {
		log.Printf("unmarshal event: %s", err)
		return
	}

	switch e := event.(type) {
	case *EventPong:
		if e.Data != s.heartbeatCount {
			log.Printf("heartbeat fibrillation %d != %d\n", e.Data, s.heartbeatCount)
			break
		}

		s.heartbeatCount++
		s.LastHeartbeatAck = time.Now()

		for _, h := range s.HandlersPong {
			h(s, e)
		}
	case *EventAuthenticated:
		go s.ping()
		for _, h := range s.HandlersAuthenticated {
			h(s, e)
		}
	case *EventAuth:
		for _, h := range s.HandlersAuth {
			h(s, e)
		}
	case *EventReady:
		s.State = newState(e)
		for _, h := range s.HandlersReady {
			h(s, e)
		}
	case *EventMessage:
		for _, h := range s.HandlersMessage {
			h(s, e)
		}
	case *EventMessageAppend:
		for _, h := range s.HandlersMessageAppend {
			h(s, e)
		}
	case *EventMessageUpdate:
		for _, h := range s.HandlersMessageUpdate {
			h(s, e)
		}
	case *EventMessageDelete:
		for _, h := range s.HandlersMessageDelete {
			h(s, e)
		}
	case *EventMessageReact:
		for _, h := range s.HandlersMessageReact {
			h(s, e)
		}
	case *EventMessageUnreact:
		for _, h := range s.HandlersMessageUnreact {
			h(s, e)
		}
	case *EventChannelCreate:
		s.State.createChannel(e)
		for _, h := range s.HandlersChannelCreate {
			h(s, e)
		}
	case *EventChannelUpdate:
		s.State.updateChannel(e)
		for _, h := range s.HandlersChannelUpdate {
			h(s, e)
		}
	case *EventChannelDelete:
		s.State.deleteChannel(e)
		for _, h := range s.HandlersChannelDelete {
			h(s, e)
		}
	case *EventGroupJoin:
		for _, h := range s.HandlersGroupJoin {
			h(s, e)
		}
	case *EventGroupLeave:
		for _, h := range s.HandlersGroupLeave {
			h(s, e)
		}
	case *EventChannelStartTyping:
		for _, h := range s.HandlersChannelStartTyping {
			h(s, e)
		}
	case *EventChannelStopTyping:
		for _, h := range s.HandlersChannelStopTyping {
			h(s, e)
		}
	case *EventServerCreate:
		s.State.createServer(e)
		for _, h := range s.HandlersServerCreate {
			h(s, e)
		}
	case *EventServerUpdate:
		s.State.updateServer(e)
		for _, h := range s.HandlersServerUpdate {
			h(s, e)
		}
	case *EventServerDelete:
		s.State.deleteServer(e)
		for _, h := range s.HandlersServerDelete {
			h(s, e)
		}
	case *EventServerMemberUpdate:
		s.State.updateServerMember(e)
		for _, h := range s.HandlersServerMemberUpdate {
			h(s, e)
		}
	case *EventServerMemberJoin:
		s.State.createServerMember(e)
		for _, h := range s.HandlersServerMemberJoin {
			h(s, e)
		}
	case *EventServerMemberLeave:
		s.State.deleteServerMember(e)
		for _, h := range s.HandlersServerMemberLeave {
			h(s, e)
		}
	case *EventUserUpdate:
		for _, h := range s.HandlersUserUpdate {
			h(s, e)
		}
	case *EventChannelAck:
		for _, h := range s.HandlersChannelAck {
			h(s, e)
		}
	case *EventServerRoleUpdate:
		s.State.updateRole(e)
		for _, h := range s.HandlersServerRoleUpdate {
			h(s, e)
		}
	case *EventServerRoleDelete:
		s.State.deleteRole(e)
		for _, h := range s.HandlersServerRoleDelete {
			h(s, e)
		}
	case *EventEmojiCreate:
		s.State.createEmoji(e)
		for _, h := range s.HandlersEmojiCreate {
			h(s, e)
		}
	case *EventEmojiDelete:
		s.State.deleteEmoji(e)
		for _, h := range s.HandlersEmojiDelete {
			h(s, e)
		}
	case *EventUserSettingsUpdate:
		for _, h := range s.HandlersUserSettingsUpdate {
			h(s, e)
		}
	case *EventUserRelationship:
		for _, h := range s.HandlersUserRelationship {
			h(s, e)
		}
	case *EventUserPlatformWipe:
		s.State.platformWipe(e)
		for _, h := range s.HandlersUserPlatformWipe {
			h(s, e)
		}
	}
}
