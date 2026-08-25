package revoltgo

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lxzan/gws"
	"github.com/tinylib/msgp/msgp"
)

//msgp:ignore Websocket

// reconnectIntervalMax caps the backoff between reconnection attempts.
const reconnectIntervalMax = 60 * time.Second

type Websocket struct {
	url     string
	session *Session

	mu   sync.RWMutex
	conn *gws.Conn

	ctx    context.Context
	cancel context.CancelFunc

	heartbeatCount    atomic.Int64
	heartbeatLastSent time.Time
	heartbeatLastAck  time.Time

	// latency is the last round trip in nanoseconds, computed where both ends
	// of it are known rather than from two timestamps read separately.
	latency atomic.Int64

	/* Configurable options */

	// Interval between sending heartbeats. Lower values update the latency faster.
	// Values too high (>=100 seconds) may cause Cloudflare to drop the connection
	HeartbeatInterval time.Duration

	Debug             bool                   // Prints sending (not a typo) and received websocket messages
	ShouldReconnect   bool                   // Whether the websocket should attempt to reconnect on disconnection
	ReconnectInterval time.Duration          // Base wait before reconnecting; doubles per failed attempt up to a minute
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
		ReconnectInterval: time.Second,
		// CustomCompression; not defined as the websocket doesn't support it yet
	}
}

func (ws *Websocket) printDebugData(data []byte) {
	var buf bytes.Buffer
	if _, err := msgp.CopyToJSON(&buf, bytes.NewReader(data)); err != nil {
		logf("Failed to convert msgpack to JSON for debug: %s", err)
		return
	}
	logf("[WS/RX]: %s", buf.String())
}

func (ws *Websocket) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.conn != nil
}

// Latency returns the round trip of the last acknowledged heartbeat, and zero
// until one has been acknowledged.
func (ws *Websocket) Latency() time.Duration {
	return time.Duration(ws.latency.Load())
}

// Uptime approximates the duration the Websocket has been connected for
func (ws *Websocket) Uptime() time.Duration {
	count := ws.heartbeatCount.Load()

	ws.mu.RLock()
	lastSent := ws.heartbeatLastSent
	ws.mu.RUnlock()

	uptime := time.Duration(count) * ws.HeartbeatInterval
	if count != 0 {
		uptime += time.Since(lastSent)
	}

	return uptime
}

// connect dials the gateway. The caller decides what a failure means:
// Open reports it, reconnectLoop retries it.
func (ws *Websocket) connect() error {

	// ctx is fixed at construction, so the shutdown check needs no lock. The
	// dial is deliberately outside the lock: it can take as long as the TCP and
	// TLS handshakes do, and WriteClose must not wait on it.
	if err := ws.ctx.Err(); err != nil {
		return err
	}

	url, _, _ := strings.Cut(ws.url, "?")
	logf("Connecting to %s...", url)

	options := &gws.ClientOption{
		Addr: ws.url,

		// Serial dispatch. Gateway order is the contract every handler is
		// written against, and gws's parallel pool does not preserve it.
		ParallelEnabled:  false,
		CheckUtf8Enabled: false,
	}

	if ws.CustomCompression != nil {
		options.PermessageDeflate = *ws.CustomCompression
	}

	socket, response, err := gws.NewClient(ws, options)
	if err != nil {
		return err
	}

	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// The session closed, or another dial won, while this one was in flight.
	// Either way nothing holds this socket, so it is this call's to drop.
	if err = ws.ctx.Err(); err != nil {
		_ = socket.WriteClose(1000, nil)
		return err
	}

	if ws.conn != nil {
		_ = socket.WriteClose(1000, nil)
		return nil
	}

	// Installed before the read loop starts, so OnClose can always recognise
	// its own socket.
	ws.conn = socket

	go socket.ReadLoop()
	return nil
}

// backoff is how long to wait before attempt n (counted from zero): the base
// doubled per failure and capped, then spread by up to a fifth either way so
// that clients dropped together do not return together.
func (ws *Websocket) backoff(attempt int) time.Duration {
	delay := ws.ReconnectInterval
	if delay <= 0 {
		delay = time.Second
	}

	for i := 0; i < attempt && delay < reconnectIntervalMax; i++ {
		delay *= 2
	}

	if delay > reconnectIntervalMax {
		delay = reconnectIntervalMax
	}

	spread := int64(delay) / 5

	return delay + time.Duration(rand.Int64N(2*spread+1)-spread)
}

// reconnectLoop retries until the connection succeeds, the session is closed,
// or reconnects are disabled.
func (ws *Websocket) reconnectLoop() {
	for attempt := 0; ws.ShouldReconnect; attempt++ {
		select {
		case <-ws.ctx.Done():
			return
		case <-time.After(ws.backoff(attempt)):
		}

		logf("Re-connecting...")

		err := ws.connect()
		if err == nil {
			return
		}

		logf("Connection failed: %s", err)
	}
}

func (ws *Websocket) heartbeatLoop(original *gws.Conn) {
	ticker := time.NewTicker(ws.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			ws.mu.RLock()
			current := ws.conn
			ws.mu.RUnlock()

			// Original connection died, kill this goroutine.
			if current != original {
				return
			}

			count := ws.heartbeatCount.Load()
			payload := make([]byte, 8)
			binary.LittleEndian.PutUint64(payload, uint64(count))

			// Stamped before the write: the pong can arrive before WritePing
			// returns, and would otherwise be timed against the previous send.
			ws.mu.Lock()
			ws.heartbeatLastSent = time.Now()
			ws.mu.Unlock()

			if err := current.WritePing(payload); err != nil {
				logf("Heartbeat failed: %s", err)
				_ = current.WriteClose(1000, nil) // Fires OnClose: handle reconnection logic
				return
			}
		}
	}
}

func (ws *Websocket) OnOpen(socket *gws.Conn) {
	logf("Resolved: %s", socket.RemoteAddr())
	ws.heartbeatCount.Store(0)

	if err := socket.SetDeadline(time.Now().Add(WebsocketKeepAlivePeriod * 2)); err != nil {
		logf("Set deadline failed: %s", err)
		_ = socket.WriteClose(1000, nil) // Fires OnClose: handle reconnection logic
		return
	}

	go ws.heartbeatLoop(socket)
}

func (ws *Websocket) OnClose(socket *gws.Conn, err error) {
	ws.mu.Lock()

	// A socket that has already been replaced must not clear the live
	// connection, nor start a second reconnect loop for it.
	if ws.conn != socket {
		ws.mu.Unlock()
		return
	}

	ws.conn = nil
	ws.mu.Unlock()

	if err == nil || ws.ctx.Err() != nil {
		logf("Connection closed gracefully")
		return
	}

	/*
		C:\Users\User\go\pkg\mod\github.com\lxzan\gws@v1.8.9\internal\error.go
		gws.CloseNormalClosure is 1000
	*/

	if closeErr, ok := errors.AsType[*gws.CloseError](err); ok {
		logf("Connection closed with code %d: %s", closeErr.Code, err)
	} else {
		logf("Connection closed with error: %s", err)
	}

	if ws.ShouldReconnect && ws.ctx.Err() == nil {
		go ws.reconnectLoop()
	}
}

func (ws *Websocket) OnPong(socket *gws.Conn, payload []byte) {

	// Any pong is proof the peer is alive, whatever counter it echoes back.
	_ = socket.SetDeadline(time.Now().Add(ws.HeartbeatInterval * 2))

	if len(payload) < 8 {
		return
	}

	count := int64(binary.LittleEndian.Uint64(payload))
	current := ws.heartbeatCount.Load()

	if count != current {
		logf("Heartbeat mismatch: %d != %d", count, current)
		return
	}

	acknowledged := time.Now()

	ws.mu.Lock()
	ws.heartbeatLastAck = acknowledged

	// An unsolicited pong carrying a zeroed payload matches count 0 before any
	// ping has been sent; timing it against the zero value is not a round trip.
	if !ws.heartbeatLastSent.IsZero() {
		ws.latency.Store(int64(acknowledged.Sub(ws.heartbeatLastSent)))
	}

	ws.mu.Unlock()

	ws.heartbeatCount.Add(1)
}

func (ws *Websocket) OnPing(_ *gws.Conn, payload []byte) {
	logf("Received unexpected ping: %s", string(payload))
}

// OnMessage dispatches one frame. It runs on the read loop, so a handler that
// blocks blocks the connection: past HeartbeatInterval*2 the read deadline
// kills the socket. Handlers that cannot answer promptly must hand off.
func (ws *Websocket) OnMessage(_ *gws.Conn, message *gws.Message) {

	data := message.Data.Bytes()
	ws.handle(data)

	if ws.Debug {
		ws.printDebugData(data)
	}

	if err := message.Close(); err != nil {
		logf("OnMessage Close() error: %s", err)
	}
}

func (ws *Websocket) WriteMessage(opcode gws.Opcode, payload []byte) error {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if ws.conn == nil {
		return gws.ErrConnClosed
	}

	if ws.Debug {
		logf("[WS/TX]: %s", string(payload))
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

	eventType, err := eventTypeFromMSGP(raw)
	if err != nil {
		logf("event type detection failed: %s", err)
		return
	}

	// Load the immutable handler snapshot lock-free. Writers swap in a fresh
	// snapshot via copy-on-write, so this read never blocks or tears.
	handlers := ws.session.handlers.Load()
	defaultHandler := handlers.defaults[string(eventType)]
	userHandlers := handlers.user[string(eventType)]

	// No one is listening for this event; drop it before paying the decode cost.
	if defaultHandler == nil && len(userHandlers) == 0 {
		return
	}

	constructor, found := eventConstructors[string(eventType)]
	if !found {
		logf("Unknown event type: %s", eventType)
		return
	}

	// constructor returns msgp.Unmarshaler, so the missing-method case is now a
	// compile-time error in eventConstructors rather than a runtime check here.
	event := constructor()
	if _, err = event.UnmarshalMsg(raw); err != nil {
		logf("Failed to unmarshal event %T: %s", event, err)
		return
	}

	// Library handler first, so user handlers observe up-to-date state.
	if defaultHandler != nil {
		defaultHandler(ws.session, event)
	}

	for _, h := range userHandlers {
		h(ws.session, event)
	}
}
