package revoltgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sacOO7/gowebsocket"
)

const (
	websocketHeartbeatInterval = 30 * time.Second
)

type WebsocketMessageType string

const (
	WebsocketMessageTypeAuthenticate WebsocketMessageType = "Authenticate"
	WebsocketMessageTypeHeartbeat    WebsocketMessageType = "Ping"
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

// Open initiates the websocket connection to the Revolt API
// todo: replace with a proper websocket library
func (s *Session) Open() {

	s.Socket = gowebsocket.New(wsURL)
	s.HTTP = &http.Client{}

	// Login the user if self-bot.
	// if s.SelfBot != nil {
	// 	s.Login()
	// }

	// Attempt authentication when the socket is connected
	s.Socket.OnConnected = func(_ gowebsocket.Socket) {

		fmt.Println("websocket/connect: connected")

		wsAuth := WebsocketMessageAuthenticate{
			Type: WebsocketMessageTypeAuthenticate,
		}

		// Populate fields for self-bots if specified, default to bots only
		switch {
		case s.SelfBot != nil:
			wsAuth.Token = s.SelfBot.SessionToken
			wsAuth.ID = s.SelfBot.ID
			wsAuth.UID = s.SelfBot.UserID
			wsAuth.Name = "revolt"
			wsAuth.Result = "Success"
			wsAuth.Token = s.Token
		default:
			wsAuth.Token = s.Token
		}

		// Encode and send the message
		authJSON, err := json.Marshal(wsAuth)
		if err != nil {
			panic(err)
		}

		s.Socket.SendText(string(authJSON))
	}

	// Using a map of event types to their respective handlers would be more concise, but slower.
	// Consider in the future.
	s.Socket.OnTextMessage = func(message string, _ gowebsocket.Socket) {
		fmt.Println("websocket/message ->", message)

		var event Event
		err := json.Unmarshal([]byte(message), &event)
		if err != nil {
			s.Close()
			panic(err)
		}

		// Asynchronously handle the event
		go s.handleEvent(event, []byte(message))
	}

	// Open connection.
	fmt.Println("websocket/connect: connecting...")
	s.Socket.Connect()
}

// Close the websocket.
func (s *Session) Close() {
	s.Socket.Close()
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
		time.Sleep(websocketHeartbeatInterval)
		s.Socket.SendText(string(pingJSON))
		s.LastHeartbeatSent = time.Now()
	}
}

func (s *Session) handleEvent(data Event, raw []byte) {

	// todo: add session handlers for pong and authenticated
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

func (s *Session) handleCache(ready *EventReady) {

	state := &State{
		Users:    make(map[string]*User, len(ready.Users)),
		Servers:  make(map[string]*Server, len(ready.Servers)),
		Channels: make(map[string]*ServerChannel, len(ready.Channels)),
		Members:  make(map[string]*Member, len(ready.Members)),
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
		state.Members[member.Informations.UserID] = member
	}

	for _, emoji := range ready.Emojis {
		state.Emojis[emoji.ID] = emoji
	}

	fmt.Println("state fulfilled: ", state)
	s.State = state
	fmt.Println("s.State ->", s.State)
}
