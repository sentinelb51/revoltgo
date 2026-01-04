package revoltgo_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sentinelb51/revoltgo"
)

// MockHTTPRequester is a mock implementation of HTTPRequester for testing.
// This demonstrates the value of the interface - we can easily create test doubles.
type MockHTTPRequester struct {
	RequestFunc   func(method, destination string, data, result any) error
	RequestCalls  []MockRequestCall
	TimeoutValue  time.Duration
	Headers       map[string]string
}

type MockRequestCall struct {
	Method      string
	Destination string
	Data        any
	Result      any
}

func NewMockHTTPRequester() *MockHTTPRequester {
	return &MockHTTPRequester{
		Headers: make(map[string]string),
	}
}

func (m *MockHTTPRequester) Request(method, destination string, data, result any) error {
	m.RequestCalls = append(m.RequestCalls, MockRequestCall{
		Method:      method,
		Destination: destination,
		Data:        data,
		Result:      result,
	})

	if m.RequestFunc != nil {
		return m.RequestFunc(method, destination, data, result)
	}
	return nil
}

func (m *MockHTTPRequester) SetTimeout(timeout time.Duration) error {
	m.TimeoutValue = timeout
	return nil
}

func (m *MockHTTPRequester) AddHeader(key, value string) error {
	if _, exists := m.Headers[key]; exists {
		return fmt.Errorf("header %q already exists", key)
	}
	m.Headers[key] = value
	return nil
}

func (m *MockHTTPRequester) SetHeader(key, value string) {
	m.Headers[key] = value
}

func (m *MockHTTPRequester) RemoveHeader(key string) {
	delete(m.Headers, key)
}

func (m *MockHTTPRequester) Header(key string) string {
	return m.Headers[key]
}

func (m *MockHTTPRequester) ResolveURL(destination string) (string, error) {
	return destination, nil
}

// TestHTTPRequesterInterface demonstrates that HTTPClient implements the interface.
func TestHTTPRequesterInterface(t *testing.T) {
	session := revoltgo.New("test-token")
	
	// Verify that HTTPClient implements HTTPRequester
	var _ revoltgo.HTTPRequester = session.HTTP
	
	// Test that we can use the interface methods
	err := session.HTTP.SetTimeout(30 * time.Second)
	if err != nil {
		t.Errorf("SetTimeout failed: %v", err)
	}
}

// TestMockHTTPRequester demonstrates using a mock for testing.
func TestMockHTTPRequester(t *testing.T) {
	mock := NewMockHTTPRequester()
	
	// Verify the mock implements the interface
	var _ revoltgo.HTTPRequester = mock
	
	// Test basic operations
	mock.SetHeader("X-Custom", "value")
	if mock.Header("X-Custom") != "value" {
		t.Error("Header not set correctly")
	}
	
	// Test request tracking
	_ = mock.Request("GET", "/test", nil, nil)
	if len(mock.RequestCalls) != 1 {
		t.Errorf("Expected 1 request call, got %d", len(mock.RequestCalls))
	}
	
	if mock.RequestCalls[0].Method != "GET" {
		t.Errorf("Expected GET method, got %s", mock.RequestCalls[0].Method)
	}
}

// TestUpdatableInterface demonstrates the Updatable interface usage.
func TestUpdatableInterface(t *testing.T) {
	// Create a user to test with
	user := &revoltgo.User{
		ID:       "user123",
		Username: "testuser",
	}
	
	// Create partial update data
	newUsername := "updateduser"
	partialUser := revoltgo.PartialUser{
		Username: &newUsername,
	}
	
	// Use the generic update helper
	revoltgo.UpdateEntity(user, partialUser)
	
	if user.Username != "updateduser" {
		t.Errorf("Expected username to be 'updateduser', got '%s'", user.Username)
	}
	
	// Test clearing fields
	revoltgo.ClearEntityFields(user, []string{"DisplayName"})
	if user.DisplayName != nil {
		t.Error("Expected DisplayName to be cleared")
	}
}

// TestStateStoreInterface demonstrates that State implements StateStore.
func TestStateStoreInterface(t *testing.T) {
	session := revoltgo.New("test-token")
	
	// Verify that State implements StateStore
	var _ revoltgo.StateStore = session.State
	
	// Test that we can use interface methods
	if !session.State.TrackUsers() {
		t.Error("Expected TrackUsers to return true")
	}
}
