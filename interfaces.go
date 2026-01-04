package revoltgo

import (
	"time"
)

// Updatable represents entities that can be partially updated and have fields cleared.
// This interface unifies the update pattern across User, Server, Channel, ServerRole, ServerMember, and Webhook.
type Updatable[T any] interface {
	update(data T)
	clear(fields []string)
}

// HTTPRequester defines the interface for making HTTP requests.
// This abstraction allows for easier testing and potential alternative implementations
// (e.g., mock clients, rate-limited clients, retry logic wrappers).
type HTTPRequester interface {
	Request(method, destination string, data, result any) error
	SetTimeout(timeout time.Duration) error
	AddHeader(key, value string) error
	SetHeader(key, value string)
	RemoveHeader(key string)
	Header(key string) string
	ResolveURL(destination string) (string, error)
}

// StateStore defines the interface for state storage operations.
// This allows for alternative storage backends beyond the default in-memory implementation
// (e.g., Redis, database, distributed cache).
type StateStore interface {
	// Self returns the current user
	Self() *User

	// Getter methods for entities
	User(id string) *User
	Server(id string) *Server
	Channel(id string) *Channel
	Member(uID, sID string) *ServerMember
	Members(sID string) []*ServerMember
	Role(sID, rID string) *ServerRole
	Emoji(id string) *Emoji
	Webhook(id string) *Webhook

	// Tracking configuration
	TrackUsers() bool
	TrackServers() bool
	TrackChannels() bool
	TrackMembers() bool
	TrackEmojis() bool
	TrackWebhooks() bool
	TrackAPICalls() bool
	TrackBulkAPICalls() bool
}

// Ensure HTTPClient implements HTTPRequester
var _ HTTPRequester = (*HTTPClient)(nil)

// Ensure State implements StateStore
var _ StateStore = (*State)(nil)
