package revoltgo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false

const (
	httpHeaderSessionToken = "X-Session-Token"
	httpHeaderBotToken     = "X-Bot-Token"
)

type HTTPClient struct {
	Debug bool

	mu          sync.RWMutex
	client      *http.Client
	session     *Session
	ratelimiter *Ratelimiter
	headers     map[string]string
}

func newHTTPClient(session *Session) *HTTPClient {
	return &HTTPClient{
		session:     session,
		client:      &http.Client{Timeout: 10 * time.Second},
		ratelimiter: newRatelimiter(),
		headers: map[string]string{
			"User-Agent": fmt.Sprintf("RevoltGo/%s (github.com/sentinelb51/revoltgo)", VERSION),
		},
	}
}

// SetTimeout sets the HTTP client timeout between 1 and 300 seconds
func (c *HTTPClient) SetTimeout(timeout time.Duration) error {
	const (
		minTimeout = time.Second
		maxTimeout = 300 * time.Second
	)

	if timeout < minTimeout {
		return fmt.Errorf("timeout %s < %s", timeout, minTimeout)
	}

	if timeout > maxTimeout {
		return fmt.Errorf("timeout %s > %s", timeout, maxTimeout)
	}

	c.client.Timeout = timeout
	return nil
}

// AddHeader adds a header and checks if it already exists
func (c *HTTPClient) AddHeader(key, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.headers[key]; exists {
		return fmt.Errorf("header %q already exists", key)
	}

	c.headers[key] = value
	return nil
}

// SetHeader overwrites a header. Use AddHeader to avoid overwriting existing headers.
func (c *HTTPClient) SetHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers[key] = value
}

// RemoveHeader removes a header
func (c *HTTPClient) RemoveHeader(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.headers, key)
}

// Header retrieves a header value
func (c *HTTPClient) Header(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.headers[key]
}

// ResolveURL converts a relative URL to an absolute URL. Prefixes relative URLs with the API base URL.
// It also allows absolute URLs targeting the CDN. Otherwise, it rejects the URL.
func (c *HTTPClient) ResolveURL(destination string) (string, error) {
	destination = strings.TrimSpace(destination)
	if destination == "" {
		return "", fmt.Errorf("destination empty")
	}

	u, err := url.Parse(destination)
	if err != nil {
		return "", fmt.Errorf("parse(destination): %w", err)
	}

	// Reject scheme-less URLs (//host/path) and any provided scheme.
	if u.Scheme != "" || u.Host != "" {
		if sameHostname(u, parsedAPIBase) || sameHostname(u, parsedCDNBase) {
			return u.String(), nil
		}
		return "", fmt.Errorf("refusing external URL host %q", u.Host)
	}

	// Path-only (or query/fragment) reference.
	return parsedAPIBase.ResolveReference(u).String(), nil
}

func sameHostname(a, b *url.URL) bool {
	// Host may include port; compare case-insensitively.
	return strings.EqualFold(a.Host, b.Host)
}

// printDebugTX logs the outgoing request details if debugging is enabled.
func (c *HTTPClient) printDebugTX(method, destination string, data any) {
	if !c.Debug {
		return
	}

	var payload string
	if data != nil {
		if _, ok := data.(*File); ok {
			payload = "[Multipart File]"
		} else {
			if b, err := json.Marshal(data); err == nil {
				payload = string(b)
			}
		}
	}
	log.Printf("[HTTP/TX] %s %s -> %s", method, destination, payload)
}

// printDebugRX logs the incoming response details if debugging is enabled.
// It reads the body and restores it so it can be read again later.
func (c *HTTPClient) printDebugRX(response *http.Response) {
	if !c.Debug {
		return
	}

	// Read body for logging
	bodyBytes, _ := io.ReadAll(response.Body)
	response.Body.Close() // Close original network reader

	log.Printf("[HTTP/RX] %d %s", response.StatusCode, string(bodyBytes))

	// Restore body with NopCloser over a buffer for handleResponse
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
}

/*
Request sends a JSON Request with "method" to a destination URL
- "result" will be used to decode the response into, and
- "data" is the Request body which wil be encoded as JSON

- If the "data" is a *File, it will be uploaded as a multipart form
This function automatically handles rate-limiting and response status codes
*/
func (c *HTTPClient) Request(method, destination string, data, result any) error {

	destination, err := c.ResolveURL(destination)
	if err != nil {
		return err
	}

	rl := c.ratelimiter.get(method, destination)

	if wait := rl.delay(); wait > 0 {
		if c.Debug {
			log.Printf("[HTTP/RATELIMIT] %s %s, waiting %s", method, destination, wait)
		}

		time.Sleep(wait)
	}

	reader, contentType, err := c.prepareRequestBody(data)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(context.Background(), method, destination, reader)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", contentType)

	c.mu.RLock()
	for k, v := range c.headers {
		request.Header.Set(k, v)
	}
	c.mu.RUnlock()

	if c.session.Selfbot() {
		request.Header.Set(httpHeaderSessionToken, c.session.Token)
	} else {
		request.Header.Set(httpHeaderBotToken, c.session.Token)
	}

	if c.Debug {
		c.printDebugTX(method, destination, data)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}

	if c.Debug {
		c.printDebugRX(response)
	}

	defer response.Body.Close()

	if err = rl.update(response.Header); err != nil {
		return err
	}

	return c.handleResponse(response.StatusCode, response.Body, result)
}

// prepareRequestBody prepares an appropriate Request body and determines the content type
func (c *HTTPClient) prepareRequestBody(body any) (io.Reader, string, error) {
	if body == nil {
		return http.NoBody, "application/json", nil
	}

	if file, ok := body.(*File); ok {
		return c.prepareFileUpload(file)
	}

	return c.prepareJSONBody(body)
}

// prepareFileUpload prepares a multipart form for uploading a file
func (c *HTTPClient) prepareFileUpload(file *File) (io.Reader, string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", file.Name)
	if err != nil {
		return nil, "", fmt.Errorf("writer.CreateFormFile: %w", err)
	}

	if _, err = io.Copy(part, file.Reader); err != nil {
		return nil, "", fmt.Errorf("io.Copy: %w", err)
	}

	if err = writer.Close(); err != nil {
		return nil, "", fmt.Errorf("writer.Close: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

// prepareJSONBody encodes data as JSON
func (c *HTTPClient) prepareJSONBody(body any) (io.Reader, string, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("json.Marshal: %w", err)
	}

	return bytes.NewReader(data), "application/json", nil
}

// handleResponse processes the API response
func (c *HTTPClient) handleResponse(statusCode int, body io.Reader, result any) error {
	switch statusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusOK, http.StatusCreated:
		if result != nil {
			if err := json.NewDecoder(body).Decode(result); err != nil {
				return fmt.Errorf("handleResponse: %w", err)
			}
		}
	default:
		const limit = 1024
		message, _ := io.ReadAll(io.LimitReader(body, limit))
		return fmt.Errorf("bad status code %d: %s", statusCode, message)
	}

	return nil
}

/* HTTP data that can be sent to the REST API */

type LoginData struct {
	Email        string `msg:"email" json:"email,omitempty"`
	Password     string `msg:"password" json:"password,omitempty"`
	FriendlyName string `msg:"friendly_name" json:"friendly_name,omitempty"`
}

type BotEditData struct {
	Name            string   `msg:"name" json:"name,omitempty"`
	Public          bool     `msg:"public" json:"public,omitempty"`
	Analytics       bool     `msg:"analytics" json:"analytics,omitempty"`
	InteractionsURL string   `msg:"interactions_url" json:"interactions_url,omitempty"`
	Remove          []string `msg:"remove" json:"remove,omitempty"`
}

type BotInviteData struct {
	Server string `msg:"server" json:"server,omitempty"`
	Group  string `msg:"group" json:"group,omitempty"`
}

type BotCreateData struct {
	Name string `msg:"name" json:"name,omitempty"`
}

type AccountCreateData struct {
	Email    string `msg:"email" json:"email,omitempty"`
	Password string `msg:"password" json:"password,omitempty"`
	Invite   string `msg:"invite" json:"invite,omitempty"`
	Captcha  string `msg:"captcha" json:"captcha,omitempty"`
}

type AccountReverifyData struct {
	Email   string `msg:"email" json:"email,omitempty"`
	Captcha string `msg:"captcha" json:"captcha,omitempty"`
}

type OnboardingCompleteData struct {
	Username string `msg:"username" json:"username,omitempty"`
}

type SessionEditData struct {
	FriendlyName string `msg:"friendly_name" json:"friendly_name,omitempty"`
}

type PasswordResetConfirmData struct {
	Token          string `msg:"token" json:"token,omitempty"`
	Password       string `msg:"password" json:"password,omitempty"`
	RemoveSessions bool   `msg:"remove_sessions" json:"remove_sessions,omitempty"` // Whether to log out of all sessions
}

type AccountChangePasswordData struct {
	Password        string `msg:"password" json:"password,omitempty"`
	CurrentPassword string `msg:"current_password" json:"current_password,omitempty"`
}

type AccountChangeEmailData struct {
	Email           string `msg:"email" json:"email,omitempty"`
	CurrentPassword string `msg:"current_password" json:"current_password,omitempty"`
}

type AccountDeleteConfirmData struct {
	Token string `msg:"token" json:"token,omitempty"`
}

type UserEditData struct {
	DisplayName string       `msg:"display_name" json:"display_name,omitempty"`
	Avatar      string       `msg:"avatar" json:"avatar,omitempty"`
	Status      *UserStatus  `msg:"status" json:"status,omitempty"`
	Profile     *UserProfile `msg:"profile" json:"profile,omitempty"`
	Badges      *int         `msg:"badges" json:"badges,omitempty"`
	Flags       *int         `msg:"flags" json:"flags,omitempty"`
	Remove      []string     `msg:"remove" json:"remove,omitempty"`
}

type UsernameData struct {
	Username string `msg:"username" json:"username,omitempty"`
	Password string `msg:"password" json:"password,omitempty"`
}

// GroupCreateData describes how a group should be created
type GroupCreateData struct {
	Name        string   `msg:"name" json:"name,omitempty"`
	Description string   `msg:"description" json:"description,omitempty"`
	Users       []string `msg:"users" json:"users,omitempty"`
	NSFW        bool     `msg:"nsfw" json:"nsfw,omitempty"`
}

type ServerCreateData struct {
	Name        string `msg:"name" json:"name,omitempty"`
	Description string `msg:"description" json:"description,omitempty"`
}

type ServerEditDataRemove string

const (
	ServerEditDataRemoveIcon           ServerEditDataRemove = "Icon"
	ServerEditDataRemoveBanner         ServerEditDataRemove = "Banner"
	ServerEditDataRemoveCategories     ServerEditDataRemove = "Categories"
	ServerEditDataRemoveDescription    ServerEditDataRemove = "Description"
	ServerEditDataRemoveSystemMessages ServerEditDataRemove = "SystemMessages"
)

type ServerEditData struct {
	Name           string                 `msg:"name" json:"name,omitempty"`
	Description    string                 `msg:"description" json:"description,omitempty"`
	Icon           string                 `msg:"icon" json:"icon,omitempty"`
	Banner         string                 `msg:"banner" json:"banner,omitempty"`
	Categories     []*ServerCategory      `msg:"categories" json:"categories,omitempty"`
	SystemMessages *ServerSystemMessages  `msg:"system_messages" json:"system_messages,omitempty"`
	Flags          int                    `msg:"flags" json:"flags,omitempty"`
	Discoverable   *bool                  `msg:"discoverable" json:"discoverable,omitempty"`
	Analytics      *bool                  `msg:"analytics" json:"analytics,omitempty"`
	Remove         []ServerEditDataRemove `msg:"remove" json:"remove,omitempty"`
}

type ServerChannelCreateDataType string

const (
	ServerChannelCreateDataTypeText  ServerChannelCreateDataType = "Text"
	ServerChannelCreateDataTypeVoice ServerChannelCreateDataType = "Voice"
)

type ServerChannelCreateData struct {
	Type        ServerChannelCreateDataType `msg:"type" json:"type,omitempty"`
	Name        string                      `msg:"name" json:"name,omitempty"`
	Description string                      `msg:"description" json:"description,omitempty"`
	NSFW        bool                        `msg:"nsfw" json:"nsfw,omitempty"`
}

type ServerMemberEditData struct {
	Nickname string    `msg:"nickname" json:"nickname,omitempty"`
	Avatar   string    `msg:"avatar" json:"avatar,omitempty"`
	Roles    []string  `msg:"roles" json:"roles,omitempty"`
	Timeout  time.Time `msg:"timeout" json:"timeout,omitempty"`
	Remove   []string  `msg:"remove" json:"remove,omitempty"`
}

type MessageEditData struct {
	Content string          `msg:"content" json:"content,omitempty"`
	Embeds  []*MessageEmbed `msg:"embeds" json:"embeds,omitempty"`
}

type EmojiCreateData struct {
	Name   string       `msg:"name" json:"name,omitempty"`
	Parent *EmojiParent `msg:"parent" json:"parent,omitempty"`
	NSFW   bool         `msg:"nsfw" json:"nsfw,omitempty"`
}

type ChannelJoinCallData struct {
	Node string `msg:"node" json:"node,omitempty"` // Name of the node to join

	// Whether to force disconnect any other existing voice connections
	// Useful for disconnecting on another device and joining on a new one
	ForceDisconnect bool `msg:"force_disconnect" json:"force_disconnect,omitempty"`

	// Users which should be notified of the call starting
	// Only used when the user is the first one connected.
	Recipients []string `msg:"recipients" json:"recipients,omitempty"`
}

type ChannelMessagesParamsSortType string

const (
	ChannelMessagesParamsSortTypeRelevance ChannelMessagesParamsSortType = "Relevance"
	ChannelMessagesParamsSortTypeOldest    ChannelMessagesParamsSortType = "Oldest"
	ChannelMessagesParamsSortTypeLatest    ChannelMessagesParamsSortType = "Latest"
)

// ChannelMessagesParams is for /channels/{target}/messages
type ChannelMessagesParams struct {
	// Maximum number of messages to fetch. For nearby messages, this is (limit + 1)
	Limit int `msg:"limit" json:"limit,omitempty"`

	// Message ID before which messages should be fetched
	Before string `msg:"before" json:"before,omitempty"`

	// Message ID after which messages should be fetched
	After string `msg:"after" json:"after,omitempty"`

	// Message sort direction
	Sort ChannelMessagesParamsSortType `msg:"sort" json:"sort,omitempty"`

	// Message ID to search around. Specifying this ignores Before, After, and Sort
	Nearby string `msg:"nearby" json:"nearby,omitempty"`

	// Whether to include user (and member, if server channel) objects
	IncludeUsers bool `msg:"include_users" json:"include_users,omitempty"`
}

// ChannelSearchParams is for /channels/{target}/search
type ChannelSearchParams struct {
	ChannelMessagesParams `msg:",inline"`

	// Whether to only search for pinned messages; cannot be sent with query.
	Pinned bool `msg:"pinned" json:"pinned,omitempty"`

	// Full-text search query. See https://www.mongodb.com/docs/manual/text-search/#-text-operator
	Query string `msg:"query" json:"query,omitempty"`
}

func (p ChannelMessagesParams) Encode() string {
	values := url.Values{}

	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}

	if p.Before != "" {
		values.Set("before", p.Before)
	}

	if p.After != "" {
		values.Set("after", p.After)
	}

	if p.Sort != "" {
		values.Set("sort", string(p.Sort))
	}

	if p.Nearby != "" {
		values.Set("nearby", p.Nearby)
	}

	if p.IncludeUsers {
		values.Set("include_users", fmt.Sprint(p.IncludeUsers))
	}

	return values.Encode()
}

type ServerRoleEditData struct {
	Name   string   `msg:"name" json:"name,omitempty"`
	Colour string   `msg:"colour" json:"colour,omitempty"`
	Hoist  *bool    `msg:"hoist" json:"hoist,omitempty"`
	Rank   *int     `msg:"rank" json:"rank,omitempty"`
	Remove []string `msg:"remove" json:"remove,omitempty"`
}

type ServerRoleCreateData struct {
	Name string `msg:"name" json:"name,omitempty"`
	Rank int    `msg:"rank" json:"rank,omitempty"`
}

type PermissionsSetDefaultData struct {
	Permissions uint `msg:"permissions" json:"permissions,omitempty"`
}

type ChannelMessageBulkDeleteData struct {
	IDs []string `msg:"ids" json:"ids,omitempty"`
}

type ChannelEditData struct {
	Name        string   `msg:"name" json:"name,omitempty"`
	Description string   `msg:"description" json:"description,omitempty"`
	Owner       string   `msg:"owner" json:"owner,omitempty"`
	Icon        string   `msg:"icon" json:"icon,omitempty"`
	NSFW        bool     `msg:"nsfw" json:"nsfw,omitempty"`
	Archived    bool     `msg:"archived" json:"archived,omitempty"`
	Remove      []string `msg:"remove" json:"remove,omitempty"`
}

type SyncSettingsDataTuple struct {
	Timestamp time.Time `msg:"0" json:"0,omitempty"`
	Value     msgp.Raw  `msg:"1" json:"1,omitempty"` // Enjoy using this.
}

type SyncSettingsData map[string]SyncSettingsDataTuple

type SyncSettingsFetchData struct {
	Keys []string `msg:"keys" json:"keys,omitempty"`
}

type WebhookCreateData struct {
	Name   string `msg:"name" json:"name,omitempty"`
	Avatar string `msg:"avatar" json:"avatar,omitempty"`
}

type WebhookExecuteData Message

type WebhookEditData struct {
	Name        string               `msg:"name" json:"name,omitempty"`
	Avatar      string               `msg:"avatar" json:"avatar,omitempty"`
	Permissions string               `msg:"permissions" json:"permissions,omitempty"`
	Remove      []WebhookRemoveField `msg:"remove" json:"remove,omitempty"`
}

// AuthMFAData should only have one of its fields set, and is used for various MFA methods
type AuthMFAData struct {
	Password     string `msg:"password" json:"password,omitempty"`
	RecoveryCode string `msg:"recovery_code" json:"recovery_code,omitempty"`
	TOTPCode     string `msg:"totp_code" json:"totp_code,omitempty"`
}
