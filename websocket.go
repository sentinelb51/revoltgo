package revoltgo

import (
	"context"
	"encoding/binary"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"
	"github.com/lxzan/gws"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[R-GO] ")
}

type WebsocketMessageType string

const (
	WebsocketKeepAlivePeriod = 60 * time.Second

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
	Data int64                `json:"data"`
}

type WebsocketChannelTyping struct {
	Type    WebsocketMessageType `json:"type"`
	Channel string               `json:"channel"`
}

// todo: migrate fields like heartbeat from Session to websocket?

type Websocket struct {
	url     string
	session *Session

	mu   sync.RWMutex
	conn *gws.Conn

	ctx    context.Context
	cancel context.CancelFunc

	heartbeatCount    int64
	heartbeatLastSent time.Time
	heartbeatLastAck  time.Time

	/* Configurable options */

	// Interval between sending heartbeats. Lower values update the latency faster.
	// Values too high (>=100 seconds) may cause Cloudflare to drop the connection
	HeartbeatInterval time.Duration

	Debug             bool                   // Prints sending (not a typo) and received websocket messages
	ShouldReconnect   bool                   // Whether the websocket should attempt to reconnect on disconnection
	ReconnectInterval time.Duration          // Interval between reconnecting, if connection fails
	CustomCompression *gws.PermessageDeflate // Defines a custom compression algorithm for the Websocket.
}

// newWebsocket constructs a websocket wrapper.
func newWebsocket(session *Session, url string) *Websocket {
	ctx, cancel := context.WithCancel(context.Background())
	return &Websocket{
		url:     url,
		session: session,
		ctx:     ctx,
		cancel:  cancel,

		ShouldReconnect:   true,
		HeartbeatInterval: 30 * time.Second,
		ReconnectInterval: 5 * time.Second,
	}
}

func (ws *Websocket) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.conn != nil
}

// Latency returns the Websocket latency
func (ws *Websocket) Latency() time.Duration {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.heartbeatLastAck.Sub(ws.heartbeatLastSent)
}

// Uptime approximates the duration the Websocket has been connected for
func (ws *Websocket) Uptime() time.Duration {
	// Use atomic load because heartbeatCount is updated atomically elsewhere
	count := atomic.LoadInt64(&ws.heartbeatCount)

	ws.mu.RLock()
	lastSent := ws.heartbeatLastSent
	ws.mu.RUnlock()

	uptime := time.Duration(count) * ws.HeartbeatInterval
	if count != 0 {
		uptime += time.Since(lastSent)
	}

	return uptime
}

// connect dials a new gws connection.
func (ws *Websocket) connect() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// If we are already shutting down, do not reconnect
	if ws.ctx.Err() != nil {
		return
	}

	log.Printf("Connecting to %s...\n", StrTrimAfter(ws.url, "?"))

	options := &gws.ClientOption{
		Addr:             ws.url,
		ParallelEnabled:  true,
		ParallelGolimit:  runtime.NumCPU(),
		CheckUtf8Enabled: false,
	}

	if ws.CustomCompression != nil {
		options.PermessageDeflate = *ws.CustomCompression
	}

	socket, response, err := gws.NewClient(ws, options)
	if err != nil {
		log.Printf("Connection failed: %s\n", err)
		go ws.reconnectLoop()
		return
	}

	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}

	ws.conn = socket

	go socket.ReadLoop()
}

func (ws *Websocket) reconnectLoop() {
	select {
	case <-ws.ctx.Done():
		return
	case <-time.After(ws.ReconnectInterval):
		log.Printf("Re-connecting...")
		ws.connect()
	}
}

func (ws *Websocket) heartbeatLoop() {
	ticker := time.NewTicker(ws.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			ws.mu.RLock()
			conn := ws.conn
			ws.mu.RUnlock()

			if conn == nil {
				continue
			}

			count := atomic.LoadInt64(&ws.heartbeatCount)
			payload := make([]byte, 8)
			binary.LittleEndian.PutUint64(payload, uint64(count))

			if err := conn.WritePing(payload); err != nil {
				log.Printf("Heartbeat failed: %s\n", err)
				_ = conn.WriteClose(1000, nil) // Fires OnClose: handle reconnection logic
				return
			}

			ws.mu.Lock()
			ws.heartbeatLastSent = time.Now()
			ws.mu.Unlock()
		}
	}
}

func (ws *Websocket) OnOpen(socket *gws.Conn) {
	log.Printf("Resolved: %s\n", socket.RemoteAddr())
	atomic.StoreInt64(&ws.heartbeatCount, 0)

	if err := socket.SetDeadline(time.Now().Add(WebsocketKeepAlivePeriod * 2)); err != nil {
		log.Fatalf("Set deadline failed: %s\n", err)
	}

	go ws.heartbeatLoop()
}

func (ws *Websocket) OnClose(_ *gws.Conn, err error) {
	ws.mu.Lock()
	ws.conn = nil
	ws.mu.Unlock()

	if err != nil && err.Error() != "" {
		log.Printf("Connection closed unexpectedly: %v\n", err)
	} else {
		log.Println("Connection closed.")
	}

	// Trigger reconnect if the session is still active
	if ws.ShouldReconnect && ws.ctx.Err() == nil {
		go ws.reconnectLoop()
	}
}

func (ws *Websocket) OnPong(socket *gws.Conn, payload []byte) {
	if len(payload) < 8 {
		return
	}

	count := int64(binary.LittleEndian.Uint64(payload))
	current := atomic.LoadInt64(&ws.heartbeatCount)

	if count != current {
		log.Printf("Heartbeat mismatch: %d != %d\n", count, current)
		return
	}

	ws.mu.Lock()
	ws.heartbeatLastAck = time.Now()
	ws.mu.Unlock()

	atomic.AddInt64(&ws.heartbeatCount, 1)
	_ = socket.SetDeadline(time.Now().Add(ws.HeartbeatInterval * 2))
}

func (ws *Websocket) OnPing(_ *gws.Conn, payload []byte) {
	log.Printf("Received unexpected ping: %s\n", string(payload))
}

func (ws *Websocket) OnMessage(_ *gws.Conn, message *gws.Message) {
	// Extract to buffer
	data := message.Data.Bytes()
	buffer := make([]byte, len(data))
	copy(buffer, data)

	// Release resources
	_ = message.Close()

	if ws.Debug {
		log.Printf("[WS/RX]: %s\n", string(buffer))
	}

	// Dispatch in separate goroutine; don't block ReadLoop
	go ws.handle(buffer)
}

func (ws *Websocket) WriteMessage(opcode gws.Opcode, payload []byte) error {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if ws.conn == nil {
		return gws.ErrConnClosed
	}

	if ws.Debug {
		log.Printf("[WS/TX]: %s\n", string(payload))
	}

	return ws.conn.WriteMessage(opcode, payload)
}

func (ws *Websocket) WriteClose() error {

	// Stop heartbeat and reconnect logic
	ws.cancel()

	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.conn == nil {
		return nil
	}

	return ws.conn.WriteClose(1000, nil)
}

func (ws *Websocket) handle(raw []byte) {

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
		dispatch(ws.session, raw, ws.session.handlersError, eventConstructor)
	case "Bulk":
		dispatch(ws.session, raw, ws.session.handlersBulk, eventConstructor)
	case "Pong":
		dispatch(ws.session, raw, ws.session.handlersPong, eventConstructor)
	case
		"MessageUpdate",
		"ServerUpdate",
		"ChannelUpdate",
		"ServerRoleUpdate",
		"WebhookUpdate",
		"UserUpdate",
		"ServerMemberUpdate":
		dispatch(ws.session, raw, ws.session.handlersAbstractEventUpdate, eventConstructor)
	case "Authenticated":
		dispatch(ws.session, raw, ws.session.handlersAuthenticated, eventConstructor)
	case "Auth":
		dispatch(ws.session, raw, ws.session.handlersAuth, eventConstructor)
	case "Ready":
		dispatch(ws.session, raw, ws.session.handlersReady, eventConstructor)
	case "Message":
		dispatch(ws.session, raw, ws.session.handlersMessage, eventConstructor)
	case "MessageAppend":
		dispatch(ws.session, raw, ws.session.handlersMessageAppend, eventConstructor)
	case "MessageDelete":
		dispatch(ws.session, raw, ws.session.handlersMessageDelete, eventConstructor)
	case "MessageReact":
		dispatch(ws.session, raw, ws.session.handlersMessageReact, eventConstructor)
	case "MessageUnreact":
		dispatch(ws.session, raw, ws.session.handlersMessageUnreact, eventConstructor)
	case "ChannelCreate":
		dispatch(ws.session, raw, ws.session.handlersChannelCreate, eventConstructor)
	case "ChannelDelete":
		dispatch(ws.session, raw, ws.session.handlersChannelDelete, eventConstructor)
	case "ChannelGroupJoin":
		dispatch(ws.session, raw, ws.session.handlersGroupJoin, eventConstructor)
	case "ChannelGroupLeave":
		dispatch(ws.session, raw, ws.session.handlersGroupLeave, eventConstructor)
	case "ChannelStartTyping":
		dispatch(ws.session, raw, ws.session.handlersChannelStartTyping, eventConstructor)
	case "ChannelStopTyping":
		dispatch(ws.session, raw, ws.session.handlersChannelStopTyping, eventConstructor)
	case "ServerCreate":
		dispatch(ws.session, raw, ws.session.handlersServerCreate, eventConstructor)
	case "ServerDelete":
		dispatch(ws.session, raw, ws.session.handlersServerDelete, eventConstructor)
	case "ServerMemberJoin":
		dispatch(ws.session, raw, ws.session.handlersServerMemberJoin, eventConstructor)
	case "ServerMemberLeave":
		dispatch(ws.session, raw, ws.session.handlersServerMemberLeave, eventConstructor)
	case "ChannelAck":
		dispatch(ws.session, raw, ws.session.handlersChannelAck, eventConstructor)
	case "ServerRoleDelete":
		dispatch(ws.session, raw, ws.session.handlersServerRoleDelete, eventConstructor)
	case "EmojiCreate":
		dispatch(ws.session, raw, ws.session.handlersEmojiCreate, eventConstructor)
	case "EmojiDelete":
		dispatch(ws.session, raw, ws.session.handlersEmojiDelete, eventConstructor)
	case "UserSettingsUpdate":
		dispatch(ws.session, raw, ws.session.handlersUserSettingsUpdate, eventConstructor)
	case "UserRelationship":
		dispatch(ws.session, raw, ws.session.handlersUserRelationship, eventConstructor)
	case "UserPlatformWipe":
		dispatch(ws.session, raw, ws.session.handlersUserPlatformWipe, eventConstructor)
	case "WebhookCreate":
		dispatch(ws.session, raw, ws.session.handlersWebhookCreate, eventConstructor)
	case "WebhookDelete":
		dispatch(ws.session, raw, ws.session.handlersWebhookDelete, eventConstructor)
	}
}

// dispatch is a generic helper to unmarshal an event and invoke registered handlers.
func dispatch[T any](s *Session, raw []byte, handlers []func(*Session, T), constructor func() any) {

	// No registered handlers
	if len(handlers) == 0 {
		return
	}

	eventConstructor := constructor()
	if err := json.Unmarshal(raw, eventConstructor); err != nil {
		log.Printf("unmarshal event: %s: %s", string(raw), err)
		return
	}

	event, ok := eventConstructor.(T)
	if !ok {
		log.Printf("event type mismatch for %T", eventConstructor)
		return
	}

	for _, h := range handlers {
		h(s, event)
	}
}
