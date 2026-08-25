package revoltgo

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
	Ratelimiter uses a RWMutex since buckets are created once per endpoint, then read on every request (low write, high read)
	Ratelimit buckets use a simple mutex as they are updated on every response
*/

const (
	ratelimitHeaderLimit      = "X-RateLimit-Limit"
	ratelimitHeaderRemaining  = "X-RateLimit-Remaining"
	ratelimitHeaderResetAfter = "X-RateLimit-Reset-After"
)

// defaultIdleTimeout is how long an unused bucket is kept. A bucket that never
// saw a ratelimit header has no window to expire, so idleness is the only thing
// that can retire it.
const defaultIdleTimeout = 10 * time.Minute

type ratelimitBucket struct {
	sync.Mutex

	/* What the server has told us about this route */

	limit     int           // requests one window allows
	remaining int           // requests left in the current window
	window    time.Duration // how long a window lasts

	/* When this bucket may next be used */

	resetAfter time.Time // end of the current window; zero until a header reports one
	next       time.Time // reservation cursor: the earliest slot not yet claimed
	lastUsed   time.Time // for eviction of a bucket no header ever described
}

// spacing is the least time between two sends once a window's allowance is
// spent. It is zero until a response has reported both a limit and a window, so
// a route nothing is known about is never paced.
func (b *ratelimitBucket) spacing() time.Duration {

	if b.limit <= 0 || b.window <= 0 {
		return 0
	}

	return b.window / time.Duration(b.limit)
}

//msgp:ignore Ratelimiter

// Ratelimiter manages ratelimit buckets for different API endpoints.
type Ratelimiter struct {
	mu        sync.RWMutex
	endpoints map[string]*ratelimitBucket

	// Interval to clean-up stale ratelimit buckets.
	CleanInterval time.Duration
	// IdleTimeout is how long a bucket with no traffic is kept. Zero means
	// defaultIdleTimeout.
	IdleTimeout time.Duration
	// stop signals the background cleaner; nil when no cleaner is running
	stop chan struct{}
}

func newRatelimiter() *Ratelimiter {
	r := &Ratelimiter{
		endpoints:     make(map[string]*ratelimitBucket),
		CleanInterval: time.Minute,
		IdleTimeout:   defaultIdleTimeout,
	}

	r.start()
	return r
}

// start launches the background cleaner unless one is already running.
// Session.Open calls this so a session can be reopened after Close.
func (r *Ratelimiter) start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stop == nil {
		r.stop = make(chan struct{})
		go r.cleaner(r.stop)
	}
}

// Close stops the background cleaner goroutine. Safe to call more than once.
func (r *Ratelimiter) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stop != nil {
		close(r.stop)
		r.stop = nil
	}
}

// https://developers.stoat.chat/developers/api/ratelimits
// 		 /channels	15
// POST	/channels/:id/messages	10
//
// Buckets are keyed per route, not per object: if a limit turns out to be per
// channel id, one shared bucket is merely slower than it needs to be, whereas
// a bucket per id is a map that grows with everything the client has touched.
// todo: check if ratelimits for /channels/:id/messages is per channel ID or global?

// isULID reports whether s is a Crockford base32 ULID, which is the shape of
// every object id Stoat puts in a path. I, L, O and U are not in the alphabet.
func isULID(s string) bool {

	if len(s) != 26 {
		return false
	}

	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c >= '0' && c <= '9':
		case c >= 'A' && c <= 'Z' && c != 'I' && c != 'L' && c != 'O' && c != 'U':
		default:
			return false
		}
	}

	return true
}

// routeKey names the route an absolute URL addresses, not the object it names:
// scheme and host dropped, query dropped, and every id segment collapsed to
// ":id". Keyed on the URL instead, the bucket map would grow with the number of
// objects the client has ever touched rather than with the number of routes.
//
// Segments that are not ULIDs — invite codes, webhook tokens, emoji names —
// stay distinct. That over-counts buckets for a handful of routes, which is the
// safe direction to be wrong in.
func routeKey(method, endpoint string) string {

	// Strip query params without allocating memory (no string split)
	if index := strings.IndexByte(endpoint, '?'); index >= 0 {
		endpoint = endpoint[:index]
	}

	// Drop scheme and host; only the path names a route
	if index := strings.Index(endpoint, "://"); index >= 0 {
		if slash := strings.IndexByte(endpoint[index+3:], '/'); slash >= 0 {
			endpoint = endpoint[index+3+slash:]
		} else {
			endpoint = "/"
		}
	}

	segments := strings.Split(endpoint, "/")
	for i, segment := range segments {
		if isULID(segment) {
			segments[i] = ":id"
		}
	}

	return method + ":" + strings.Join(segments, "/")
}

func (r *Ratelimiter) get(method, endpoint string) *ratelimitBucket {

	key := routeKey(method, endpoint)

	// Optimistic read-lock (cheap)
	r.mu.RLock()
	bucket, exists := r.endpoints[key]
	r.mu.RUnlock()

	// If bucket exists, return it
	if exists {
		return bucket
	}

	// Bucket doesn't exist, upgrade to expensive lock (double-check locking pattern)
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check again in case another goroutine created a bucket during the lock upgrade
	if bucket, exists = r.endpoints[key]; exists {
		return bucket
	}

	bucket = &ratelimitBucket{lastUsed: time.Now()}
	r.endpoints[key] = bucket
	return bucket
}

// update updates the ratelimit handler by populating the remaining and resetAfter fields
func (b *ratelimitBucket) update(headers http.Header) error {
	headerRemaining := headers.Get(ratelimitHeaderRemaining)
	if headerRemaining == "" {
		// If the header is missing, we can assume the rest of the ratelimit headers are missing too
		return nil
	}

	headerResetAfter := headers.Get(ratelimitHeaderResetAfter)
	if headerResetAfter == "" {
		return fmt.Errorf("missing %s header (remaining was present?)", ratelimitHeaderResetAfter)
	}

	remaining, err := strconv.Atoi(headerRemaining)
	if err != nil {
		return err
	}

	resetAfter, err := strconv.Atoi(headerResetAfter)
	if err != nil {
		return err
	}

	// The limit header is not always sent; the highest allowance ever reported,
	// plus the request that reported it, converges on the same number.
	limit, _ := strconv.Atoi(headers.Get(ratelimitHeaderLimit))

	now := time.Now()
	window := time.Duration(resetAfter) * time.Millisecond

	b.Lock()
	defer b.Unlock()

	b.lastUsed = now

	if limit > 0 {
		b.limit = limit
	} else if remaining+1 > b.limit {
		b.limit = remaining + 1
	}

	// The first response of a window reports the whole of it, so the longest
	// reset ever seen is the window's length.
	if window > b.window {
		b.window = window
	}

	// A response only accounts for the requests the server has already seen.
	// Adopting its count outright would hand back the slots that requests still
	// in flight have claimed, so within one window it may only lower the count.
	if b.resetAfter.IsZero() || !now.Before(b.resetAfter) || remaining < b.remaining {
		b.remaining = remaining
	}

	b.resetAfter = now.Add(window)

	return nil
}

// delay claims this caller's send slot and returns how long to wait for it.
//
// The slot is claimed under the lock and the cursor moved past it, so callers
// that arrive while the window is spent are handed distinct wake times instead
// of all waking at the reset and arriving together. Slots the window still
// allows cost nothing: the server permits the burst, and its own count is what
// stops it.
func (b *ratelimitBucket) delay() time.Duration {
	b.Lock()
	defer b.Unlock()

	now := time.Now()
	b.lastUsed = now

	// The window rolled over with nobody looking. Nothing else refills the
	// allowance, and without this the bucket stops limiting after one window.
	if !b.resetAfter.IsZero() && !now.Before(b.resetAfter) {
		if b.window > 0 && b.limit > 0 {
			b.remaining = b.limit
			b.resetAfter = now.Add(b.window)
		} else {
			b.resetAfter = time.Time{}
		}
	}

	if b.next.Before(now) {
		b.next = now
	}

	// Allowance spent: the earliest slot is in the next window, not now.
	if b.remaining <= 0 && !b.resetAfter.IsZero() && b.next.Before(b.resetAfter) {
		b.next = b.resetAfter
	}

	at := b.next

	if b.remaining > 0 {
		b.remaining--
	} else {
		b.next = at.Add(b.spacing())
	}

	return at.Sub(now)
}

// cleaner takes stop as an argument so a restart cannot race with the channel it was started on
func (r *Ratelimiter) cleaner(stop chan struct{}) {
	ticker := time.NewTicker(r.CleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			r.clean()
		}
	}
}

func (r *Ratelimiter) clean() {
	r.mu.Lock()
	defer r.mu.Unlock()

	idle := r.IdleTimeout
	if idle <= 0 {
		idle = defaultIdleTimeout
	}

	now := time.Now()
	for key, bucket := range r.endpoints {

		// Idleness is the only thing that retires a bucket no header ever
		// described: it has no window to expire. A bucket is kept while its
		// window is still running or while a caller is still waiting on a slot
		// in it — dropping it there would lose the reservation.
		bucket.Lock()
		isExpired := now.Sub(bucket.lastUsed) > idle &&
			now.After(bucket.resetAfter) &&
			now.After(bucket.next)
		bucket.Unlock()

		if isExpired {
			delete(r.endpoints, key)
		}
	}
}
