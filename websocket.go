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

	// Everything below is for self-bot authentication
	ID     string `json:"_id,omitempty"`
	UID    string `json:"uid,omitempty"`
	Name   string `json:"name,omitempty"` // Always "revolt"?
	Result string `json:"result,omitempty"`
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

	// Login the user if self-bot.
	// if s.SelfBot != nil {
	// s.Login()
	// }

	fmt.Println("websocket/connect: connected")

	wsAuth := WebsocketMessageAuthenticate{
		Type: WebsocketMessageTypeAuthenticate,
	}

	// Populate fields for self-bots if specified, default to bots only
	switch {
	case s.SelfBot != nil:
		wsAuth.Token = s.SelfBot.SessionToken
		wsAuth.ID = s.SelfBot.ID
		wsAuth.UID = s.SelfBot.UID
		wsAuth.Name = "revolt"
		wsAuth.Result = "Success"
		wsAuth.Token = s.Token
	default:
		wsAuth.Token = s.Token
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
			s.Close()
			panic(err)
		}

		if op != ws.OpText {
			continue
		}

		events <- message
	}
}

func (s *Session) eventHandler() {

	for {
		select {
		case raw := <-events:
			fmt.Println("websocket/message ->", string(raw))

			var data Event
			err := json.Unmarshal(raw, &data)
			if err != nil {
				s.Close()
				panic(err)
			}

			switch data.Type {
			case EventTypePong:

				pong := data.Type.Unmarshal(raw).(*EventPong)
				s.LastHeartbeatAck = time.Now()

				fmt.Println("Websocket latency:", s.LastHeartbeatAck.Sub(s.LastHeartbeatSent), "Data:", pong.Data)

			case EventTypeAuthenticated:
				// Ignore if already connected.
				if s.Connected {
					return
				}

				go s.ping()
			case EventTypeReady:
				event := data.Type.Unmarshal(raw).(*EventReady)
				s.handleCache(event)

				for _, h := range s.OnReadyHandlers {
					h(s, event)
				}
			case EventTypeMessage:
				event := data.Type.Unmarshal(raw).(*EventMessage)

				for _, h := range s.OnMessageHandlers {
					h(s, event)
				}
			case EventTypeMessageUpdate:
				event := data.Type.Unmarshal(raw).(*EventMessageUpdate)

				for _, h := range s.OnMessageUpdateHandlers {
					h(s, event)
				}
			case EventTypeMessageDelete:
				event := data.Type.Unmarshal(raw).(*EventMessageDelete)

				for _, h := range s.OnMessageDeleteHandlers {
					h(s, event)
				}
			case EventTypeMessageReact:
				event := data.Type.Unmarshal(raw).(*EventMessageReact)

				for _, h := range s.OnMessageReactHandlers {
					h(s, event)
				}
			case EventTypeMessageUnreact:
				event := data.Type.Unmarshal(raw).(*EventMessageUnreact)

				for _, h := range s.OnMessageUnreactHandlers {
					h(s, event)
				}
			case EventTypeChannelCreate:
				event := data.Type.Unmarshal(raw).(*EventChannelCreate)

				for _, h := range s.OnChannelCreateHandlers {
					h(s, event)
				}
			case EventTypeChannelUpdate:
				event := data.Type.Unmarshal(raw).(*EventChannelUpdate)

				for _, h := range s.OnChannelUpdateHandlers {
					h(s, event)
				}
			case EventTypeChannelDelete:
				event := data.Type.Unmarshal(raw).(*EventChannelDelete)

				for _, h := range s.OnChannelDeleteHandlers {
					h(s, event)
				}
			case EventTypeChannelGroupJoin:
				event := data.Type.Unmarshal(raw).(*EventChannelGroupJoin)

				for _, h := range s.OnChannelGroupJoinHandlers {
					h(s, event)
				}
			case EventTypeChannelGroupLeave:
				event := data.Type.Unmarshal(raw).(*EventChannelGroupLeave)

				for _, h := range s.OnChannelGroupLeaveHandlers {
					h(s, event)
				}
			case EventTypeChannelStartTyping:
				event := data.Type.Unmarshal(raw).(*EventChannelStartTyping)

				for _, h := range s.OnChannelStartTypingHandlers {
					h(s, event)
				}
			case EventTypeChannelStopTyping:
				event := data.Type.Unmarshal(raw).(*EventChannelStopTyping)

				for _, h := range s.OnChannelStopTypingHandlers {
					h(s, event)
				}
			case EventTypeServerCreate:
				event := data.Type.Unmarshal(raw).(*EventServerCreate)

				for _, h := range s.OnServerCreateHandlers {
					h(s, event)
				}
			case EventTypeServerUpdate:
				event := data.Type.Unmarshal(raw).(*EventServerUpdate)

				for _, h := range s.OnServerUpdateHandlers {
					h(s, event)
				}
			case EventTypeServerDelete:
				event := data.Type.Unmarshal(raw).(*EventServerDelete)

				for _, h := range s.OnServerDeleteHandlers {
					h(s, event)
				}
			case EventTypeServerMemberUpdate:
				event := data.Type.Unmarshal(raw).(*EventServerMemberUpdate)

				for _, h := range s.OnServerMemberUpdateHandlers {
					h(s, event)
				}
			case EventTypeServerMemberJoin:
				event := data.Type.Unmarshal(raw).(*EventServerMemberJoin)

				for _, h := range s.OnServerMemberJoinHandlers {
					h(s, event)
				}
			case EventTypeServerMemberLeave:
				event := data.Type.Unmarshal(raw).(*EventServerMemberLeave)

				for _, h := range s.OnServerMemberLeaveHandlers {
					h(s, event)
				}
			default:
				for _, h := range s.OnUnknownEventHandlers {
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
		fmt.Println(member.ID)
		state.Members[member.ID.User] = member
	}

	for _, emoji := range ready.Emojis {
		state.Emojis[emoji.ID] = emoji
	}

	s.State = state
}
