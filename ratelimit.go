package revoltgo

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ratelimitHeaderRemaining  = "X-RateLimit-Remaining"
	ratelimitHeaderResetAfter = "X-RateLimit-Reset-After"
)

type ratelimitBucket struct {
	sync.RWMutex
	remaining  int
	resetAfter time.Time
}

// Ratelimiter silently ensures that requests do not exceed the ratelimit set by Revolt
type Ratelimiter struct {
	sync.Mutex
	endpoints map[string]*ratelimitBucket

	// Interval to clean-up stale ratelimit buckets.
	// Higher values will result in higher memory usage, but lower CPU, and vice versa.
	CleanInterval time.Duration
}

func newRatelimiter() *Ratelimiter {
	r := &Ratelimiter{
		endpoints:     make(map[string]*ratelimitBucket, 10),
		CleanInterval: time.Minute,
	}

	go r.cleaner()
	return r
}

func (r *Ratelimiter) get(method string, endpoint string) *ratelimitBucket {

	// Split to remove query parameters
	endpoint = strings.SplitN(endpoint, "?", 2)[0]

	// To reduce key size, we truncate the base URL from the endpoint
	// The HTTP method is prepended to the endpoint as ratelimits may differ between methods
	key := method + endpoint[len(apiURL):]

	r.Lock()
	defer r.Unlock()

	bucket, exists := r.endpoints[key]
	if !exists {
		bucket = &ratelimitBucket{}
		r.endpoints[key] = bucket
	}

	return bucket
}

// update updates the ratelimit handler by populating the remaining and resetAfter fields
func (b *ratelimitBucket) update(headers http.Header) error {

	var (
		value int
		err   error
	)

	headerRemaining := headers.Get(ratelimitHeaderRemaining)
	if headerRemaining == "" {
		// If the header is missing, we can assume the rest of the ratelimit headers are missing too
		return nil
	}

	headerResetAfter := headers.Get(ratelimitHeaderResetAfter)
	if headerResetAfter == "" {
		return fmt.Errorf("missing %s header (remaining was present?)", ratelimitHeaderResetAfter)
	}

	b.Lock()
	defer b.Unlock()

	value, err = strconv.Atoi(headerRemaining)
	if err != nil {
		return err
	}
	b.remaining = value

	value, err = strconv.Atoi(headerResetAfter)
	if err != nil {
		return err
	}
	b.resetAfter = time.Now().Add(time.Duration(value) * time.Millisecond)

	return err
}

// delay returns the time to wait before sending the request
func (b *ratelimitBucket) delay() time.Duration {
	b.RLock()
	defer b.RUnlock()

	if b.remaining > 0 {
		return 0
	}

	return b.resetAfter.Sub(time.Now())
}

func (r *Ratelimiter) cleaner() {
	for {
		time.Sleep(r.CleanInterval)

		// Avoid unnecessary locking
		if len(r.endpoints) == 0 {
			continue
		}

		r.Lock()
		for key, bucket := range r.endpoints {
			if !time.Now().After(bucket.resetAfter) {
				continue
			}

			delete(r.endpoints, key)
		}
		r.Unlock()
	}
}
