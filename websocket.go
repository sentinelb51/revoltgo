package revoltgo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net/http"
	"time"
)

var Events = make(chan []byte)

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

	// These fields may be completely redundant
	//ID     string `json:"_id,omitempty"`
	//UID    string `json:"uid,omitempty"`
	//Name   string `json:"name,omitempty"` // Always "revolt"?
	//Result string `json:"result,omitempty"`
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

	defer close(Events)

	for {
		message, op, err := wsutil.ReadServerData(s.Socket)
		if err != nil {
			s.Close()
			panic(err)
		}

		if op != ws.OpText {
			continue
		}

		Events <- message
	}
}

func (s *Session) eventHandler() {

	for {
		select {
		case raw := <-Events:
			// fmt.Println("websocket/message ->", string(raw))

			var data Event
			err := json.Unmarshal(raw, &data)
			if err != nil {
				panic(err)
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
				s.handleCache(event)

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

				for _, h := range s.HandlersChannelCreate {
					h(s, event)
				}
			case EventTypeChannelUpdate:
				event := data.Type.Unmarshal(raw).(*EventChannelUpdate)

				if value, exists := s.State.Channels[event.ID]; exists {
					value = merge(value, event.Data).(*Channel)
				}

				for _, h := range s.HandlersChannelUpdate {
					h(s, event)
				}
			case EventTypeChannelDelete:
				event := data.Type.Unmarshal(raw).(*EventChannelDelete)

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

				s.State.Servers[event.Server.ID] = event.Server

				for _, h := range s.HandlersServerCreate {
					h(s, event)
				}
			case EventTypeServerUpdate:
				event := data.Type.Unmarshal(raw).(*EventServerUpdate)

				if value, exists := s.State.Servers[event.ID]; exists {
					value = merge(value, event.Data).(*Server)
				}

				for _, h := range s.HandlersServerUpdate {
					h(s, event)
				}
			case EventTypeServerDelete:
				event := data.Type.Unmarshal(raw).(*EventServerDelete)

				delete(s.State.Servers, event.ID)

				for _, h := range s.HandlersServerDelete {
					h(s, event)
				}
			case EventTypeServerMemberUpdate:
				event := data.Type.Unmarshal(raw).(*EventServerMemberUpdate)

				if value, exists := s.State.Members[event.ID]; exists {
					value = merge(value, event.Data).(*ServerMember)
				}

				for _, h := range s.HandlersServerMemberUpdate {
					h(s, event)
				}
			case EventTypeServerMemberJoin:
				event := data.Type.Unmarshal(raw).(*EventServerMemberJoin)

				for _, h := range s.HandlersServerMemberJoin {
					h(s, event)
				}
			case EventTypeServerMemberLeave:
				event := data.Type.Unmarshal(raw).(*EventServerMemberLeave)

				for _, h := range s.HandlersServerMemberLeave {
					h(s, event)
				}
			case EventTypeUserUpdate:
				event := data.Type.Unmarshal(raw).(*EventUserUpdate)

				if value, exists := s.State.Users[event.ID]; exists {
					value = merge(value, event.Data).(*User)
				}

				for _, h := range s.HandlersUserUpdate {
					h(s, event)
				}
			case EventTypeChannelAck:
				event := data.Type.Unmarshal(raw).(*EventChannelAck)

				for _, h := range s.HandlersChannelAck {
					h(s, event)
				}
			default:
				for _, h := range s.HandlersUnknown {
					h(s, string(raw))
				}
			}
		}
	}
}

// Close the websocket.
func (s *Session) Close() error {
	return s.Socket.Close()
}

// ping pings the websocket every websocketHeartbeatInterval interval
func (s *Session) ping() {

	// Avoid duplicate ping goroutines
	if s.Connected {
		return
	}

	wsPing := WebsocketMessagePing{
		Type: WebsocketMessageTypeHeartbeat,
		Data: 1337,
	}

	// Encode and send the message
	// We can re-use the message, unless we want to randomise the data field for some reason?
	pingJSON, err := json.Marshal(wsPing)
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(s.HeartbeatInterval)
		err = wsutil.WriteClientText(s.Socket, pingJSON)
		if err != nil {
			panic(err)
		}
		s.LastHeartbeatSent = time.Now()
	}
}

func (s *Session) handleCache(ready *EventReady) {

	state := &State{
		Users:    make(map[string]*User, len(ready.Users)),
		Servers:  make(map[string]*Server, len(ready.Servers)),
		Channels: make(map[string]*Channel, len(ready.Channels)),
		Members:  make(map[string]*ServerMember, len(ready.Members)),
		Emojis:   make(map[string]*Emoji, len(ready.Emojis)),
	}

	// Populate the caches

	for _, user := range ready.Users {
		state.Users[user.ID] = user
	}

	for _, server := range ready.Servers {
		state.Servers[server.ID] = server
	}

	for _, channel := range ready.Channels {
		state.Channels[channel.ID] = channel
	}

	for _, member := range ready.Members {
		state.Members[member.ID.User] = member
	}

	for _, emoji := range ready.Emojis {
		state.Emojis[emoji.ID] = emoji
	}

	s.State = state
}
