# Interfaces in RevoltGo

This document explains the interfaces introduced in RevoltGo and when to use them.

## Overview

RevoltGo now provides several interfaces that improve code organization, testability, and extensibility:

1. **`HTTPRequester`** - Abstracts HTTP operations
2. **`StateStore`** - Abstracts state storage operations  
3. **`Updatable[T]`** - Generic interface for entity updates

## HTTPRequester Interface

### Purpose
The `HTTPRequester` interface abstracts HTTP request operations, making the code more testable and allowing alternative HTTP implementations.

### Definition
```go
type HTTPRequester interface {
    Request(method, destination string, data, result any) error
    SetTimeout(timeout time.Duration) error
    AddHeader(key, value string) error
    SetHeader(key, value string)
    RemoveHeader(key string)
    Header(key string) string
    ResolveURL(destination string) (string, error)
}
```

### Use Cases

#### Testing
Create mock HTTP clients for unit testing without making real network calls:

```go
type MockHTTPRequester struct {
    RequestFunc func(method, destination string, data, result any) error
}

func (m *MockHTTPRequester) Request(method, destination string, data, result any) error {
    if m.RequestFunc != nil {
        return m.RequestFunc(method, destination, data, result)
    }
    return nil
}

// Implement other methods...
```

#### Custom HTTP Behavior
Wrap the default client to add custom behavior like retries, circuit breakers, or additional logging:

```go
type RetryingHTTPRequester struct {
    underlying HTTPRequester
    maxRetries int
}

func (r *RetryingHTTPRequester) Request(method, destination string, data, result any) error {
    var err error
    for i := 0; i < r.maxRetries; i++ {
        err = r.underlying.Request(method, destination, data, result)
        if err == nil {
            return nil
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }
    return err
}
```

## StateStore Interface

### Purpose
The `StateStore` interface abstracts state storage operations, allowing alternative storage backends beyond the default in-memory implementation.

### Definition
```go
type StateStore interface {
    Self() *User
    User(id string) *User
    Server(id string) *Server
    Channel(id string) *Channel
    Member(uID, sID string) *ServerMember
    Members(sID string) []*ServerMember
    Role(sID, rID string) *ServerRole
    Emoji(id string) *Emoji
    Webhook(id string) *Webhook
    TrackUsers() bool
    TrackServers() bool
    TrackChannels() bool
    TrackMembers() bool
    TrackEmojis() bool
    TrackWebhooks() bool
    TrackAPICalls() bool
    TrackBulkAPICalls() bool
}
```

### Use Cases

#### Alternative Storage Backends
Implement persistent or distributed state storage:

```go
type RedisStateStore struct {
    client *redis.Client
}

func (r *RedisStateStore) User(id string) *User {
    // Fetch user from Redis
    return nil
}
```

#### Testing
Create mock state stores for testing without a full state object:

```go
type MockStateStore struct {
    users map[string]*User
}

func (m *MockStateStore) User(id string) *User {
    return m.users[id]
}
```

## Updatable[T] Interface

### Purpose
The `Updatable[T]` interface unifies the update pattern across different entity types, enabling generic code for entity updates.

### Definition
```go
type Updatable[T any] interface {
    update(data T)
    clear(fields []string)
}
```

### Implementing Types
The following types implement `Updatable`:
- `User` (with `PartialUser`)
- `Server` (with `PartialServer`)
- `Channel` (with `PartialChannel`)
- `ServerRole` (with `PartialServerRole`)
- `ServerMember` (with `PartialServerMember`)
- `Webhook` (with `PartialWebhook`)

### Use Cases

#### Generic Update Logic
Write functions that work with any updatable entity:

```go
// Apply both update and clear operations
func ApplyPartialUpdate[T any, U Updatable[T]](entity U, data T, clearFields []string) {
    entity.update(data)
    entity.clear(clearFields)
}

// Use with any entity type
user := &User{ID: "123"}
ApplyPartialUpdate(user, PartialUser{Username: &newName}, []string{"Avatar"})
```

#### Consistent Update Handling
Process updates uniformly across different entity types:

```go
func HandleUpdate[T any, U Updatable[T]](entity U, updateData T) {
    UpdateEntity(entity, updateData)
    // Additional processing...
}
```

## Design Principles

### When to Use Interfaces
These interfaces were added when they provide **genuine value**:

1. **Testing** - Enables dependency injection and mocking
2. **Flexibility** - Allows swapping implementations
3. **Code Reuse** - Enables generic algorithms across types

### When NOT to Use Interfaces
We avoided adding interfaces where they would:

1. Add unnecessary abstraction without clear benefit
2. Complicate the API without improving it
3. Just follow "best practices" without real-world value

## Examples

See `interfaces_test.go` for complete working examples of:
- Creating mock HTTP clients for testing
- Using generic update helpers
- Verifying interface implementations
