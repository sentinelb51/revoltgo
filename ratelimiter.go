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
	ratelimitHeaderRemaining  = "X-RateLimit-Remaining"
	ratelimitHeaderResetAfter = "X-RateLimit-Reset-After"
)

type ratelimitBucket struct {
	sync.Mutex
	remaining  int
	resetAfter time.Time
}

// Ratelimiter manages ratelimit buckets for different API endpoints.
type Ratelimiter struct {
	mu        sync.RWMutex
	endpoints map[string]*ratelimitBucket

	// Interval to clean-up stale ratelimit buckets.
	CleanInterval time.Duration
	// stop the background cleaner
	stop chan struct{}
}

func newRatelimiter() *Ratelimiter {
	r := &Ratelimiter{
		endpoints:     make(map[string]*ratelimitBucket),
		CleanInterval: time.Minute,
		stop:          make(chan struct{}),
	}

	go r.cleaner()
	return r
}

// Close stops the background cleaner goroutine
func (r *Ratelimiter) Close() {
	close(r.stop)
}

func (r *Ratelimiter) get(method, endpoint string) *ratelimitBucket {

	// Strip query params without allocating memory (no string split)
	if index := strings.IndexByte(endpoint, '?'); index >= 0 {
		endpoint = endpoint[:index]
	}

	key := fmt.Sprintf("%s:%s", method, endpoint)

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

	bucket = &ratelimitBucket{}
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

	b.Lock()
	b.remaining = remaining
	b.resetAfter = time.Now().Add(time.Duration(resetAfter) * time.Millisecond)
	b.Unlock()

	return nil
}

// delay returns the time to wait before sending the request
func (b *ratelimitBucket) delay() time.Duration {
	b.Lock()
	defer b.Unlock()

	if b.remaining > 0 {
		return 0
	}

	wait := time.Until(b.resetAfter)
	if wait < 0 {
		return 0
	}

	return wait
}

func (r *Ratelimiter) cleaner() {
	ticker := time.NewTicker(r.CleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.stop:
			return
		case <-ticker.C:
			r.clean()
		}
	}
}

func (r *Ratelimiter) clean() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, bucket := range r.endpoints {

		// Expired if bucket has received headers (not zero) or now is after resetAfter
		bucket.Lock()
		isExpired := !bucket.resetAfter.IsZero() && now.After(bucket.resetAfter)
		bucket.Unlock()

		if isExpired {
			delete(r.endpoints, key)
		}
	}
}
