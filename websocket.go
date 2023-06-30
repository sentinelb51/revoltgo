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
	log.SetPrefix("[rgo] ")
}

var events = make(chan []byte)

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
func (s *Session) Open() error {

	if s.Connected {
		return fmt.Errorf("already connected")
	}

	// Determine the websocket URL
	var query QueryNode
	err := s.request(http.MethodGet, baseURL, nil, &query)
	if err != nil {
		return err
	}

	dialer := ws.Dialer{
		Timeout: 5 * time.Second,
	}

	connection, _, _, err := dialer.Dial(context.Background(), query.Ws)
	if err != nil {
		return err
	}
	s.Socket = connection

	wsAuth := WebsocketMessageAuthenticate{
		Type:  WebsocketMessageTypeAuthenticate,
		Token: s.Token,
	}

	// Send initial authentication message
	err = s.WriteSocket(wsAuth)

	// If successfully sent, start listening and processing websocket events
	if err == nil {
		go s.eventHandler()
		go s.listen()
	}

	return err
}

// listen reads messages from the websocket
func (s *Session) listen() {

	defer close(events)

	for {
		message, op, err := wsutil.ReadServerData(s.Socket)
		if err != nil {
			if !s.Connected {
				return
			}
			panic(err)
		}

		if op != ws.OpText {
			continue
		}

		events <- message
	}
}

func (s *Session) eventHandler() {

	for raw := range events {

		var data Event
		err := json.Unmarshal(raw, &data)
		if err != nil {
			log.Println("eventHandler: unmarshal:", err)
			continue
		}

		switch data.Type {
		case EventTypePong:
			event := data.Type.Unmarshal(raw).(*EventPong)
			s.LastHeartbeatAck = time.Now()

			for _, h := range s.HandlersPong {
				h(s, event)
			}
		case EventTypeAuthenticated:
			event := data.Type.Unmarshal(raw).(*EventAuthenticated)

			for _, h := range s.HandlersAuthenticated {
				h(s, event)
			}

			go s.ping()
		case EventTypeAuth:
			event := data.Type.Unmarshal(raw).(*EventAuth)

			for _, h := range s.HandlersAuth {
				h(s, event)
			}
		case EventTypeReady:
			event := data.Type.Unmarshal(raw).(*EventReady)
			s.State = newState(event)

			for _, h := range s.HandlersReady {
				h(s, event)
			}
		case EventTypeMessage:
			event := data.Type.Unmarshal(raw).(*EventMessage)

			for _, h := range s.HandlersMessage {
				h(s, event)
			}
		case EventTypeMessageAppend:
			event := data.Type.Unmarshal(raw).(*EventMessageAppend)

			for _, h := range s.HandlersMessageAppend {
				h(s, event)
			}
		case EventTypeMessageUpdate:
			event := data.Type.Unmarshal(raw).(*EventMessageUpdate)

			for _, h := range s.HandlersMessageUpdate {
				h(s, event)
			}
		case EventTypeMessageDelete:
			event := data.Type.Unmarshal(raw).(*EventMessageDelete)

			for _, h := range s.HandlersMessageDelete {
				h(s, event)
			}
		case EventTypeMessageReact:
			event := data.Type.Unmarshal(raw).(*EventMessageReact)

			for _, h := range s.HandlersMessageReact {
				h(s, event)
			}
		case EventTypeMessageUnreact:
			event := data.Type.Unmarshal(raw).(*EventMessageUnreact)

			for _, h := range s.HandlersMessageUnreact {
				h(s, event)
			}
		case EventTypeChannelCreate:
			event := data.Type.Unmarshal(raw).(*EventChannelCreate)
			s.State.updateChannels(event)

			for _, h := range s.HandlersChannelCreate {
				h(s, event)
			}
		case EventTypeChannelUpdate:
			event := data.Type.Unmarshal(raw).(*EventChannelUpdate)
			s.State.updateChannels(event)

			for _, h := range s.HandlersChannelUpdate {
				h(s, event)
			}
		case EventTypeChannelDelete:
			event := data.Type.Unmarshal(raw).(*EventChannelDelete)
			s.State.updateChannels(event)

			for _, h := range s.HandlersChannelDelete {
				h(s, event)
			}
		case EventTypeGroupJoin:
			event := data.Type.Unmarshal(raw).(*EventGroupJoin)

			for _, h := range s.HandlersGroupJoin {
				h(s, event)
			}
		case EventTypeGroupLeave:
			event := data.Type.Unmarshal(raw).(*EventGroupLeave)

			for _, h := range s.HandlersGroupLeave {
				h(s, event)
			}
		case EventTypeChannelStartTyping:
			event := data.Type.Unmarshal(raw).(*EventChannelStartTyping)

			for _, h := range s.HandlersChannelStartTyping {
				h(s, event)
			}
		case EventTypeChannelStopTyping:
			event := data.Type.Unmarshal(raw).(*EventChannelStopTyping)

			for _, h := range s.HandlersChannelStopTyping {
				h(s, event)
			}
		case EventTypeServerCreate:
			event := data.Type.Unmarshal(raw).(*EventServerCreate)
			s.State.updateServers(event)

			s.State.Servers[event.Server.ID] = event.Server

			for _, h := range s.HandlersServerCreate {
				h(s, event)
			}
		case EventTypeServerUpdate:
			event := data.Type.Unmarshal(raw).(*EventServerUpdate)
			s.State.updateServers(event)

			for _, h := range s.HandlersServerUpdate {
				h(s, event)
			}
		case EventTypeServerDelete:
			event := data.Type.Unmarshal(raw).(*EventServerDelete)
			s.State.updateServers(event)

			delete(s.State.Servers, event.ID)

			for _, h := range s.HandlersServerDelete {
				h(s, event)
			}
		case EventTypeServerMemberUpdate:
			event := data.Type.Unmarshal(raw).(*EventServerMemberUpdate)
			s.State.updateMembers(event)

			for _, h := range s.HandlersServerMemberUpdate {
				h(s, event)
			}
		case EventTypeServerMemberJoin:
			event := data.Type.Unmarshal(raw).(*EventServerMemberJoin)
			s.State.updateMembers(event)

			for _, h := range s.HandlersServerMemberJoin {
				h(s, event)
			}
		case EventTypeServerMemberLeave:
			event := data.Type.Unmarshal(raw).(*EventServerMemberLeave)
			s.State.updateMembers(event)

			for _, h := range s.HandlersServerMemberLeave {
				h(s, event)
			}
		case EventTypeUserUpdate:
			event := data.Type.Unmarshal(raw).(*EventUserUpdate)
			s.State.updateUsers(event)

			for _, h := range s.HandlersUserUpdate {
				h(s, event)
			}
		case EventTypeChannelAck:
			event := data.Type.Unmarshal(raw).(*EventChannelAck)

			for _, h := range s.HandlersChannelAck {
				h(s, event)
			}
		case EventTypeServerRoleUpdate:
			event := data.Type.Unmarshal(raw).(*EventServerRoleUpdate)
			s.State.updateRoles(event)

			for _, h := range s.HandlersServerRoleUpdate {
				h(s, event)
			}
		case EventTypeServerRoleDelete:
			event := data.Type.Unmarshal(raw).(*EventServerRoleDelete)
			s.State.updateRoles(event)

			for _, h := range s.HandlersServerRoleDelete {
				h(s, event)
			}
		case EventTypeEmojiCreate:
			event := data.Type.Unmarshal(raw).(*EventEmojiCreate)
			s.State.updateEmojis(event)

			for _, h := range s.HandlersEmojiCreate {
				h(s, event)
			}
		case EventTypeEmojiDelete:
			event := data.Type.Unmarshal(raw).(*EventEmojiDelete)
			s.State.updateEmojis(event)

			for _, h := range s.HandlersEmojiDelete {
				h(s, event)
			}
		default:
			for _, h := range s.HandlersUnknown {
				h(s, string(raw))
			}
		}
	}
}

// Close the websocket.
func (s *Session) Close() error {
	s.Connected = false
	return s.Socket.Close()
}

// ping pings the websocket every HeartbeatInterval interval
// It makes sure that the connection is still alive, and triggers a reconnect if a problem occurs
func (s *Session) ping() {

	// Avoid duplicate ping goroutines
	if s.Connected {
		return
	}

	wsPing := WebsocketMessagePing{
		Type: WebsocketMessageTypeHeartbeat,
		Data: 1337,
	}

	// Look into making WriteSocketRaw; avoid marshalling

	for {
		time.Sleep(s.HeartbeatInterval)
		err := s.WriteSocket(wsPing)
		if err != nil {
			log.Printf("heartbeat failed: %s", err)
			break
		}
		s.LastHeartbeatSent = time.Now()
	}

	log.Printf("heartbeat failed, triggering reconnect...")
	s.Connected = false
	_ = s.Open()
}
