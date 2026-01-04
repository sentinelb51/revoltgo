package revoltgo_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sentinelb51/revoltgo"
)

// Example: Using HTTPRequester interface for testing
// This example shows how the interface enables testing without real HTTP calls

// MockUserService demonstrates testing business logic with a mock HTTP client
type MockUserService struct {
	http revoltgo.HTTPRequester
}

func NewMockUserService(http revoltgo.HTTPRequester) *MockUserService {
	return &MockUserService{http: http}
}

func (s *MockUserService) GetUserDisplayName(userID string) (string, error) {
	var user revoltgo.User
	endpoint := fmt.Sprintf("/users/%s", userID)

	err := s.http.Request("GET", endpoint, nil, &user)
	if err != nil {
		return "", err
	}

	if user.DisplayName != nil {
		return *user.DisplayName, nil
	}
	return user.Username, nil
}

// ConfigurableHTTPRequester allows customizing request behavior for testing
type ConfigurableHTTPRequester struct {
	Responses map[string]any   // Map endpoint to response data
	Errors    map[string]error // Map endpoint to errors
	Headers   map[string]string
}

func NewConfigurableHTTPRequester() *ConfigurableHTTPRequester {
	return &ConfigurableHTTPRequester{
		Responses: make(map[string]any),
		Errors:    make(map[string]error),
		Headers:   make(map[string]string),
	}
}

func (c *ConfigurableHTTPRequester) Request(method, destination string, data, result any) error {
	// Check if we should return an error
	if err, exists := c.Errors[destination]; exists {
		return err
	}

	// Check if we have a configured response
	if response, exists := c.Responses[destination]; exists {
		// Marshal and unmarshal to simulate JSON encoding/decoding
		jsonData, _ := json.Marshal(response)
		return json.Unmarshal(jsonData, result)
	}

	return nil
}

func (c *ConfigurableHTTPRequester) SetTimeout(timeout time.Duration) error { return nil }
func (c *ConfigurableHTTPRequester) AddHeader(key, value string) error {
	c.Headers[key] = value
	return nil
}
func (c *ConfigurableHTTPRequester) SetHeader(key, value string) { c.Headers[key] = value }
func (c *ConfigurableHTTPRequester) RemoveHeader(key string)     { delete(c.Headers, key) }
func (c *ConfigurableHTTPRequester) Header(key string) string    { return c.Headers[key] }
func (c *ConfigurableHTTPRequester) ResolveURL(destination string) (string, error) {
	return destination, nil
}

// Example_mockHTTPRequester demonstrates using the interface for testing
func Example_mockHTTPRequester() {
	// Create a configurable mock
	mock := NewConfigurableHTTPRequester()

	// Configure the mock to return a user
	displayName := "Test User"
	mock.Responses["/users/123"] = &revoltgo.User{
		ID:          "123",
		Username:    "testuser",
		DisplayName: &displayName,
	}

	// Use the mock in our service
	service := NewMockUserService(mock)
	name, err := service.GetUserDisplayName("123")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("User display name: %s\n", name)
	// Output: User display name: Test User
}

// TestUserServiceWithMock demonstrates a complete test using the interface
func TestUserServiceWithMock(t *testing.T) {
	mock := NewConfigurableHTTPRequester()

	// Test 1: User with display name
	displayName := "John Doe"
	mock.Responses["/users/user1"] = &revoltgo.User{
		ID:          "user1",
		Username:    "johndoe",
		DisplayName: &displayName,
	}

	service := NewMockUserService(mock)
	name, err := service.GetUserDisplayName("user1")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if name != "John Doe" {
		t.Errorf("Expected 'John Doe', got '%s'", name)
	}

	// Test 2: User without display name (should fall back to username)
	mock.Responses["/users/user2"] = &revoltgo.User{
		ID:       "user2",
		Username: "janedoe",
	}

	name, err = service.GetUserDisplayName("user2")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if name != "janedoe" {
		t.Errorf("Expected 'janedoe', got '%s'", name)
	}

	// Test 3: Error handling
	mock.Errors["/users/user3"] = fmt.Errorf("user not found")

	_, err = service.GetUserDisplayName("user3")

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// Example showing how the Updatable interface enables generic code

// GenericEntityUpdater can update any entity that implements Updatable
type GenericEntityUpdater[T any, U revoltgo.Updatable[T]] struct {
	entities map[string]U
}

func NewGenericEntityUpdater[T any, U revoltgo.Updatable[T]]() *GenericEntityUpdater[T, U] {
	return &GenericEntityUpdater[T, U]{
		entities: make(map[string]U),
	}
}

func (g *GenericEntityUpdater[T, U]) Add(id string, entity U) {
	g.entities[id] = entity
}

func (g *GenericEntityUpdater[T, U]) Update(id string, data T, clearFields []string) bool {
	entity, exists := g.entities[id]
	if !exists {
		return false
	}

	revoltgo.ApplyPartialUpdate(entity, data, clearFields)
	return true
}

// Example_genericEntityUpdater demonstrates generic entity management
func Example_genericEntityUpdater() {
	// Create a generic updater for users
	userUpdater := NewGenericEntityUpdater[revoltgo.PartialUser, *revoltgo.User]()

	// Add a user
	user := &revoltgo.User{
		ID:       "123",
		Username: "testuser",
	}
	userUpdater.Add("123", user)

	// Update the user
	newUsername := "updateduser"
	updated := userUpdater.Update("123", revoltgo.PartialUser{
		Username: &newUsername,
	}, nil)

	fmt.Printf("Updated: %v, New username: %s\n", updated, user.Username)
	// Output: Updated: true, New username: updateduser
}

// TestGenericEntityUpdater demonstrates the generic updater with different entity types
func TestGenericEntityUpdater(t *testing.T) {
	// Test with User entities
	userUpdater := NewGenericEntityUpdater[revoltgo.PartialUser, *revoltgo.User]()

	user := &revoltgo.User{ID: "u1", Username: "oldname"}
	userUpdater.Add("u1", user)

	newName := "newname"
	if !userUpdater.Update("u1", revoltgo.PartialUser{Username: &newName}, nil) {
		t.Error("Failed to update user")
	}

	if user.Username != "newname" {
		t.Errorf("Expected username 'newname', got '%s'", user.Username)
	}

	// Test with Server entities
	serverUpdater := NewGenericEntityUpdater[revoltgo.PartialServer, *revoltgo.Server]()

	server := &revoltgo.Server{ID: "s1", Name: "Old Server"}
	serverUpdater.Add("s1", server)

	newServerName := "New Server"
	if !serverUpdater.Update("s1", revoltgo.PartialServer{Name: &newServerName}, nil) {
		t.Error("Failed to update server")
	}

	if server.Name != "New Server" {
		t.Errorf("Expected server name 'New Server', got '%s'", server.Name)
	}
}
