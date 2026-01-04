package revoltgo_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sentinelb51/revoltgo"
)

// RetryHTTPRequester wraps an HTTPRequester and adds retry logic.
// This demonstrates how interfaces enable composition and decoration patterns.
type RetryHTTPRequester struct {
	underlying revoltgo.HTTPRequester
	maxRetries int
	retryDelay time.Duration
}

func NewRetryHTTPRequester(underlying revoltgo.HTTPRequester, maxRetries int, retryDelay time.Duration) *RetryHTTPRequester {
	return &RetryHTTPRequester{
		underlying: underlying,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (r *RetryHTTPRequester) Request(method, destination string, data, result any) error {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(r.retryDelay * time.Duration(attempt))
		}

		lastErr = r.underlying.Request(method, destination, data, result)
		if lastErr == nil {
			return nil // Success!
		}
	}

	return fmt.Errorf("failed after %d retries: %w", r.maxRetries, lastErr)
}

func (r *RetryHTTPRequester) SetTimeout(timeout time.Duration) error {
	return r.underlying.SetTimeout(timeout)
}

func (r *RetryHTTPRequester) AddHeader(key, value string) error {
	return r.underlying.AddHeader(key, value)
}

func (r *RetryHTTPRequester) SetHeader(key, value string) {
	r.underlying.SetHeader(key, value)
}

func (r *RetryHTTPRequester) RemoveHeader(key string) {
	r.underlying.RemoveHeader(key)
}

func (r *RetryHTTPRequester) Header(key string) string {
	return r.underlying.Header(key)
}

func (r *RetryHTTPRequester) ResolveURL(destination string) (string, error) {
	return r.underlying.ResolveURL(destination)
}

// TestRetryHTTPRequester demonstrates wrapping with retry logic
func TestRetryHTTPRequester(t *testing.T) {
	// Create a mock that fails twice then succeeds
	attempts := 0
	mock := NewMockHTTPRequester()
	mock.RequestFunc = func(method, destination string, data, result any) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("temporary error")
		}
		return nil
	}

	// Wrap with retry logic
	retryClient := NewRetryHTTPRequester(mock, 3, 10*time.Millisecond)

	// Should succeed after retries
	err := retryClient.Request("GET", "/test", nil, nil)
	if err != nil {
		t.Errorf("Expected success after retries, got: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if len(mock.RequestCalls) != 3 {
		t.Errorf("Expected 3 calls recorded, got %d", len(mock.RequestCalls))
	}
}

// LoggingHTTPRequester wraps an HTTPRequester and logs all requests.
// This demonstrates how interfaces enable cross-cutting concerns.
type LoggingHTTPRequester struct {
	underlying revoltgo.HTTPRequester
	logFunc    func(method, destination string)
}

func NewLoggingHTTPRequester(underlying revoltgo.HTTPRequester, logFunc func(method, destination string)) *LoggingHTTPRequester {
	return &LoggingHTTPRequester{
		underlying: underlying,
		logFunc:    logFunc,
	}
}

func (l *LoggingHTTPRequester) Request(method, destination string, data, result any) error {
	if l.logFunc != nil {
		l.logFunc(method, destination)
	}
	return l.underlying.Request(method, destination, data, result)
}

func (l *LoggingHTTPRequester) SetTimeout(timeout time.Duration) error {
	return l.underlying.SetTimeout(timeout)
}

func (l *LoggingHTTPRequester) AddHeader(key, value string) error {
	return l.underlying.AddHeader(key, value)
}

func (l *LoggingHTTPRequester) SetHeader(key, value string) {
	l.underlying.SetHeader(key, value)
}

func (l *LoggingHTTPRequester) RemoveHeader(key string) {
	l.underlying.RemoveHeader(key)
}

func (l *LoggingHTTPRequester) Header(key string) string {
	return l.underlying.Header(key)
}

func (l *LoggingHTTPRequester) ResolveURL(destination string) (string, error) {
	return l.underlying.ResolveURL(destination)
}

// TestLoggingHTTPRequester demonstrates request logging
func TestLoggingHTTPRequester(t *testing.T) {
	var loggedRequests []string

	mock := NewMockHTTPRequester()
	loggingClient := NewLoggingHTTPRequester(mock, func(method, destination string) {
		loggedRequests = append(loggedRequests, fmt.Sprintf("%s %s", method, destination))
	})

	// Make some requests
	_ = loggingClient.Request("GET", "/users/123", nil, nil)
	_ = loggingClient.Request("POST", "/channels/456/messages", nil, nil)

	if len(loggedRequests) != 2 {
		t.Errorf("Expected 2 logged requests, got %d", len(loggedRequests))
	}

	if loggedRequests[0] != "GET /users/123" {
		t.Errorf("Expected 'GET /users/123', got '%s'", loggedRequests[0])
	}

	if loggedRequests[1] != "POST /channels/456/messages" {
		t.Errorf("Expected 'POST /channels/456/messages', got '%s'", loggedRequests[1])
	}
}

// Example_httpRequesterComposition demonstrates composing multiple wrappers
func Example_httpRequesterComposition() {
	// Create base mock
	mock := NewMockHTTPRequester()

	// Add logging
	logged := NewLoggingHTTPRequester(mock, func(method, destination string) {
		fmt.Printf("Request: %s %s\n", method, destination)
	})

	// Add retry logic on top of logging
	retryLogged := NewRetryHTTPRequester(logged, 2, 10*time.Millisecond)

	// Now we have a client with both logging and retry capabilities
	_ = retryLogged.Request("GET", "/api/test", nil, nil)

	// Output:
	// Request: GET /api/test
}

// InMemoryStateStore is an alternative implementation of StateStore.
// This demonstrates how the interface allows different storage strategies.
type InMemoryStateStore struct {
	users    map[string]*revoltgo.User
	servers  map[string]*revoltgo.Server
	channels map[string]*revoltgo.Channel
	self     *revoltgo.User
}

func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		users:    make(map[string]*revoltgo.User),
		servers:  make(map[string]*revoltgo.Server),
		channels: make(map[string]*revoltgo.Channel),
	}
}

func (s *InMemoryStateStore) Self() *revoltgo.User                          { return s.self }
func (s *InMemoryStateStore) User(id string) *revoltgo.User                 { return s.users[id] }
func (s *InMemoryStateStore) Server(id string) *revoltgo.Server             { return s.servers[id] }
func (s *InMemoryStateStore) Channel(id string) *revoltgo.Channel           { return s.channels[id] }
func (s *InMemoryStateStore) Member(uID, sID string) *revoltgo.ServerMember { return nil }
func (s *InMemoryStateStore) Members(sID string) []*revoltgo.ServerMember   { return nil }
func (s *InMemoryStateStore) Role(sID, rID string) *revoltgo.ServerRole     { return nil }
func (s *InMemoryStateStore) Emoji(id string) *revoltgo.Emoji               { return nil }
func (s *InMemoryStateStore) Webhook(id string) *revoltgo.Webhook           { return nil }
func (s *InMemoryStateStore) TrackUsers() bool                              { return true }
func (s *InMemoryStateStore) TrackServers() bool                            { return true }
func (s *InMemoryStateStore) TrackChannels() bool                           { return true }
func (s *InMemoryStateStore) TrackMembers() bool                            { return false }
func (s *InMemoryStateStore) TrackEmojis() bool                             { return false }
func (s *InMemoryStateStore) TrackWebhooks() bool                           { return false }
func (s *InMemoryStateStore) TrackAPICalls() bool                           { return false }
func (s *InMemoryStateStore) TrackBulkAPICalls() bool                       { return false }

// TestCustomStateStore demonstrates using a custom state store
func TestCustomStateStore(t *testing.T) {
	store := NewInMemoryStateStore()

	// Verify it implements the interface
	var _ revoltgo.StateStore = store

	// Add a user
	user := &revoltgo.User{ID: "123", Username: "testuser"}
	store.users["123"] = user

	// Retrieve the user
	retrieved := store.User("123")
	if retrieved == nil {
		t.Error("Expected to retrieve user")
	}

	if retrieved.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrieved.Username)
	}
}
