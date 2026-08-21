package revoltgo

import (
	"bytes"
	json "encoding/json/v2"
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

	"github.com/klauspost/compress/zstd"
	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false

const (
	httpHeaderSessionToken = "X-Session-Token"
	httpHeaderBotToken     = "X-Bot-Token"
)

// zstdPool reuses streaming decoders across requests. The pool grows to one
// decoder per concurrent zstd response; concurrency 1 keeps each lean to pool.
var zstdPool = sync.Pool{
	New: func() any {
		d, _ := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
		return d
	},
}

type HTTPClient struct {
	Debug bool

	mu          sync.RWMutex
	client      *http.Client
	transport   *h3Transport
	session     *Session
	ratelimiter *Ratelimiter
	headers     map[string]string
}

func newHTTPClient(session *Session) *HTTPClient {
	transport := newH3Transport()

	return &HTTPClient{
		session:     session,
		transport:   transport,
		client:      &http.Client{Transport: transport, Timeout: 10 * time.Second},
		ratelimiter: newRatelimiter(),
		headers: map[string]string{
			"User-Agent":      fmt.Sprintf("RevoltGo/%s (github.com/sentinelb51/revoltgo)", VERSION),
			"Accept-Encoding": "zstd",
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

	// Fast path: our endpoints are usually absolute paths ("/endpoint") -> Skip url.Parse/ResolveReference.
	if strings.HasPrefix(destination, "/") && !strings.HasPrefix(destination, "//") {
		return apiURL + destination, nil
	}

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
	var payload string
	if data != nil {
		if _, ok := data.(*FileParams); ok {
			payload = "[Multipart File]"
		} else {
			if b, err := json.Marshal(data); err == nil {
				payload = string(b)
			}
		}
	}
	log.Printf("[HTTP/TX] %s %s -> %s", method, destination, payload)
}

// printDebugRX logs the incoming response body and returns a replayable reader
// so handleResponse can read it again.
func (c *HTTPClient) printDebugRX(statusCode int, body io.Reader) io.Reader {
	bodyBytes, _ := io.ReadAll(body)
	log.Printf("[HTTP/RX] %d %s", statusCode, string(bodyBytes))
	return bytes.NewReader(bodyBytes)
}

/*
Request sends a JSON Request with "method" to a destination URL
- "result" will be used to decode the response into, and
- "data" is the Request body which wil be encoded as JSON

- If the "data" is a *FileParams, it will be uploaded as a multipart form
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

	request, err := http.NewRequest(method, destination, reader)
	if err != nil {
		// Nothing owns the reader yet;
		// Close it so a streaming body's writer doesn't pull a "step-bro I'm stuck >.<"
		if closer, ok := reader.(io.Closer); ok {
			_ = closer.Close()
		}
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

	if err = rl.update(response.Header); err != nil {
		_ = response.Body.Close()
		return err
	}

	// Retry once on 429; rl.update already absorbed the reset window from the response.
	// Requests without GetBody (file uploads stream from a pipe) cannot be replayed.
	if response.StatusCode == http.StatusTooManyRequests && request.GetBody != nil {
		_ = response.Body.Close()

		if wait := rl.delay(); wait > 0 {
			if c.Debug {
				log.Printf("[HTTP/RATELIMIT] %s %s got 429, retrying in %s", method, destination, wait)
			}

			time.Sleep(wait)
		}

		if request.Body, err = request.GetBody(); err != nil {
			return err
		}

		if response, err = c.client.Do(request); err != nil {
			return err
		}

		if err = rl.update(response.Header); err != nil {
			_ = response.Body.Close()
			return err
		}
	}

	defer func() { _ = response.Body.Close() }()

	body := io.Reader(response.Body)

	// Handle ZStandard decompression
	if response.Header.Get("Content-Encoding") == "zstd" {
		decoder := zstdPool.Get().(*zstd.Decoder)
		_ = decoder.Reset(response.Body) // only errors on a closed decoder; pooled ones never are
		defer func() {
			_ = decoder.Reset(nil) // release body; Reset (not Close) keeps it poolable
			zstdPool.Put(decoder)
		}()
		body = decoder
	}

	if c.Debug {
		body = c.printDebugRX(response.StatusCode, body)
	}

	return c.handleResponse(response.StatusCode, body, result)
}

// prepareRequestBody prepares an appropriate Request body and determines the content type
func (c *HTTPClient) prepareRequestBody(body any) (io.Reader, string, error) {
	if body == nil {
		return http.NoBody, "application/json", nil
	}

	if file, ok := body.(*FileParams); ok {
		return c.prepareFileUpload(file)
	}

	return c.prepareJSONBody(body)
}

// prepareFileUpload prepares a multipart form for uploading a file
func (c *HTTPClient) prepareFileUpload(file *FileParams) (io.Reader, string, error) {
	reader, writer := io.Pipe()
	form := multipart.NewWriter(writer)

	// Stream the multipart body so large uploads aren't held in memory
	go func() {
		part, err := form.CreateFormFile("file", file.Name)
		if err != nil {
			writer.CloseWithError(fmt.Errorf("form.CreateFormFile: %w", err))
			return
		}

		if _, err = io.Copy(part, file.Reader); err != nil {
			writer.CloseWithError(fmt.Errorf("io.Copy: %w", err))
			return
		}

		writer.CloseWithError(form.Close())
	}()

	return reader, form.FormDataContentType(), nil
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
		switch result := result.(type) {
		case nil:
		case *[]byte:
			// Binary endpoints (e.g. default avatars) return raw bytes, not JSON
			data, err := io.ReadAll(body)
			if err != nil {
				return fmt.Errorf("handleResponse: %w", err)
			}
			*result = data
		default:
			if err := json.UnmarshalRead(body, result); err != nil {
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

type LoginParams struct {
	Email        string `msg:"email" json:"email,omitzero"`
	Password     string `msg:"password" json:"password,omitzero"`
	FriendlyName string `msg:"friendly_name" json:"friendly_name,omitzero"`
}

type BotEditParams struct {
	Name            string           `msg:"name" json:"name,omitzero"`
	Public          *bool            `msg:"public" json:"public,omitzero"`
	Analytics       *bool            `msg:"analytics" json:"analytics,omitzero"`
	InteractionsURL string           `msg:"interactions_url" json:"interactions_url,omitzero"`
	Remove          []BotRemoveField `msg:"remove" json:"remove,omitzero"`
}

type BotInviteParams struct {
	Server string `msg:"server" json:"server,omitzero"`
	Group  string `msg:"group" json:"group,omitzero"`
}

type BotCreateParams struct {
	Name string `msg:"name" json:"name,omitzero"`
}

type AccountCreateParams struct {
	Email    string `msg:"email" json:"email,omitzero"`
	Password string `msg:"password" json:"password,omitzero"`
	Invite   string `msg:"invite" json:"invite,omitzero"`
	Captcha  string `msg:"captcha" json:"captcha,omitzero"`
}

type AccountReverifyParams struct {
	Email   string `msg:"email" json:"email,omitzero"`
	Captcha string `msg:"captcha" json:"captcha,omitzero"`
}

type OnboardingCompleteParams struct {
	Username string `msg:"username" json:"username,omitzero"`
}

type SessionEditParams struct {
	FriendlyName string `msg:"friendly_name" json:"friendly_name,omitzero"`
}

type PasswordResetConfirmParams struct {
	Token          string `msg:"token" json:"token,omitzero"`
	Password       string `msg:"password" json:"password,omitzero"`
	RemoveSessions bool   `msg:"remove_sessions" json:"remove_sessions,omitzero"` // Whether to log out of all sessions
}

type AccountChangePasswordParams struct {
	Password        string `msg:"password" json:"password,omitzero"`
	CurrentPassword string `msg:"current_password" json:"current_password,omitzero"`
}

type AccountChangeEmailParams struct {
	Email           string `msg:"email" json:"email,omitzero"`
	CurrentPassword string `msg:"current_password" json:"current_password,omitzero"`
}

type AccountDeleteConfirmParams struct {
	Token string `msg:"token" json:"token,omitzero"`
}

type UserEditParams struct {
	DisplayName string             `msg:"display_name" json:"display_name,omitzero"`
	Pronouns    string             `msg:"pronouns" json:"pronouns,omitzero"`
	Avatar      string             `msg:"avatar" json:"avatar,omitzero"`
	Status      *UserStatus        `msg:"status" json:"status,omitzero"`
	Profile     *UserProfileParams `msg:"profile" json:"profile,omitzero"`
	Badges      *uint32            `msg:"badges" json:"badges,omitzero"`
	Flags       *uint32            `msg:"flags" json:"flags,omitzero"`

	Remove []UserRemoveField `msg:"remove" json:"remove,omitzero"`
}

type UserProfileParams struct {
	Content    *string `msg:"content" json:"content,omitzero"`
	Background string  `msg:"background" json:"background,omitzero"`
}

type UsernameParams struct {
	Username string `msg:"username" json:"username,omitzero"`
	Password string `msg:"password" json:"password,omitzero"`
}

// GroupCreateParams describes how a group should be created
type GroupCreateParams struct {
	Name        string   `msg:"name" json:"name,omitzero"`
	Description string   `msg:"description" json:"description,omitzero"`
	Users       []string `msg:"users" json:"users,omitzero"`
	NSFW        bool     `msg:"nsfw" json:"nsfw,omitzero"`
}

type ServerCreateParams struct {
	Name        string `msg:"name" json:"name,omitzero"`
	Description string `msg:"description" json:"description,omitzero"`
}

type ServerEditParamsRemove string

const (
	ServerEditDataRemoveIcon           ServerEditParamsRemove = "Icon"
	ServerEditDataRemoveBanner         ServerEditParamsRemove = "Banner"
	ServerEditDataRemoveCategories     ServerEditParamsRemove = "Categories"
	ServerEditDataRemoveDescription    ServerEditParamsRemove = "Description"
	ServerEditDataRemoveSystemMessages ServerEditParamsRemove = "SystemMessages"
)

type ServerEditParams struct {
	Name           string                   `msg:"name" json:"name,omitzero"`
	Description    string                   `msg:"description" json:"description,omitzero"`
	Icon           string                   `msg:"icon" json:"icon,omitzero"`
	Banner         string                   `msg:"banner" json:"banner,omitzero"`
	Categories     []*ServerCategory        `msg:"categories" json:"categories,omitzero"`
	SystemMessages *ServerSystemMessages    `msg:"system_messages" json:"system_messages,omitzero"`
	Flags          *uint32                  `msg:"flags" json:"flags,omitzero"`
	Discoverable   *bool                    `msg:"discoverable" json:"discoverable,omitzero"`
	Analytics      *bool                    `msg:"analytics" json:"analytics,omitzero"`
	Remove         []ServerEditParamsRemove `msg:"remove" json:"remove,omitzero"`
}

type ServerChannelCreateParamsType string

const (
	ServerChannelCreateDataTypeText  ServerChannelCreateParamsType = "Text"
	ServerChannelCreateDataTypeVoice ServerChannelCreateParamsType = "Voice"
)

type ServerChannelCreateParams struct {
	Type        ServerChannelCreateParamsType `msg:"type" json:"type,omitzero"`
	Name        string                        `msg:"name" json:"name,omitzero"`
	Description string                        `msg:"description" json:"description,omitzero"`
	NSFW        bool                          `msg:"nsfw" json:"nsfw,omitzero"`
}

type ServerMemberEditParams struct {
	Nickname string                  `msg:"nickname" json:"nickname,omitzero"`
	Avatar   string                  `msg:"avatar" json:"avatar,omitzero"`
	Roles    []string                `msg:"roles" json:"roles,omitzero"`
	Timeout  *time.Time              `msg:"timeout" json:"timeout,omitzero"`
	Remove   []ServerMemberClearType `msg:"remove" json:"remove,omitzero"`
}

// ServerMemberBanParams derived from:
// https://developers.stoat.chat/api-reference/#tag/server-members/PUT/servers/{server}/bans/{target}
type ServerMemberBanParams struct {
	DeleteMessageSeconds int64  `msg:"delete_message_seconds" json:"delete_message_seconds,omitzero"`
	Reason               string `msg:"reason" json:"reason,omitzero"`
}

type MessageEditParams struct {
	Content string          `msg:"content" json:"content,omitzero"`
	Embeds  []*MessageEmbed `msg:"embeds" json:"embeds,omitzero"`
}

type EmojiCreateParams struct {
	Name   string       `msg:"name" json:"name,omitzero"`
	Parent *EmojiParent `msg:"parent" json:"parent,omitzero"`
	NSFW   bool         `msg:"nsfw" json:"nsfw,omitzero"`
}

type ChannelJoinCallParams struct {
	Node string `msg:"node" json:"node,omitzero"` // Name of the node to join

	// Whether to force disconnect any other existing voice connections
	// Useful for disconnecting on another device and joining on a new one
	ForceDisconnect bool `msg:"force_disconnect" json:"force_disconnect,omitzero"`

	// Users which should be notified of the call starting
	// Only used when the user is the first one connected.
	Recipients []string `msg:"recipients" json:"recipients,omitzero"`
}

type ChannelMessagesParamsSortType string

const (
	ChannelMessagesParamsSortTypeRelevance ChannelMessagesParamsSortType = "Relevance"
	ChannelMessagesParamsSortTypeOldest    ChannelMessagesParamsSortType = "Oldest"
	ChannelMessagesParamsSortTypeLatest    ChannelMessagesParamsSortType = "Latest"
)

// ChannelMessagesParams is for /channels/{target}/messages
type ChannelMessagesParams struct {
	// Maximum number of messages to fetch. For nearby messages, this is (limit + 2)
	Limit int `msg:"limit" json:"limit,omitzero"`

	// Message ID before which messages should be fetched
	Before string `msg:"before" json:"before,omitzero"`

	// Message ID after which messages should be fetched
	After string `msg:"after" json:"after,omitzero"`

	// Message sort direction
	Sort ChannelMessagesParamsSortType `msg:"sort" json:"sort,omitzero"`

	// Message ID to search around. Specifying this ignores Before, After, and Sort
	Nearby string `msg:"nearby" json:"nearby,omitzero"`

	// Whether to include user (and member, if server channel) objects
	IncludeUsers bool `msg:"include_users" json:"include_users,omitzero"`
}

// ChannelSearchParams is for /channels/{target}/search
type ChannelSearchParams struct {
	ChannelMessagesParams `msg:",inline"`

	// Whether to only search for pinned messages; cannot be sent with query.
	Pinned bool `msg:"pinned" json:"pinned,omitzero"`

	// Full-text search query. See https://www.mongodb.com/docs/manual/text-search/#-text-operator
	Query string `msg:"query" json:"query,omitzero"`
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

type ServerRoleEditParams struct {
	Name   string                `msg:"name" json:"name,omitzero"`
	Colour string                `msg:"colour" json:"colour,omitzero"`
	Hoist  *bool                 `msg:"hoist" json:"hoist,omitzero"`
	Rank   *int                  `msg:"rank" json:"rank,omitzero"`
	Remove []ServerRoleClearType `msg:"remove" json:"remove,omitzero"`
}

type ServerRoleCreateParams struct {
	Name string `msg:"name" json:"name,omitzero"`
	Rank *int   `msg:"rank" json:"rank,omitzero"` // nil lets the API assign a rank
}

// ServerRoleCreateResponse is derived from
// https://developers.stoat.chat/api-reference/#tag/server-permissions/POST/servers/{target}/roles
type ServerRoleCreateResponse struct {
	ID   string     `msg:"id" json:"id,omitzero"`
	Role ServerRole `msg:"role" json:"role,omitzero"`
}

type PermissionsSetDefaultParams struct {
	// Always sent; 0 is a valid value that denies everything
	Permissions int64 `msg:"permissions" json:"permissions"`
}

// PermissionsSetParams is DataSetServerRolePermission and DataSetRolePermissions:
// https://developers.stoat.chat/api-reference/#tag/server-permissions/PUT/servers/{target}/permissions/{role_id}
type PermissionsSetParams struct {
	Permissions PermissionOverwriteParams `msg:"permissions" json:"permissions"`
}

// ServerRoleRanksParams is DataEditRoleRanks:
// https://developers.stoat.chat/api-reference/#tag/server-permissions/PATCH/servers/{target}/roles/ranks
type ServerRoleRanksParams struct {
	Ranks []string `msg:"ranks" json:"ranks"`
}

type ChannelMessageBulkDeleteParams struct {
	IDs []string `msg:"ids" json:"ids,omitzero"`
}

type ChannelEditParams struct {
	Name        string `msg:"name" json:"name,omitzero"`
	Description string `msg:"description" json:"description,omitzero"`
	Owner       string `msg:"owner" json:"owner,omitzero"`
	Icon        string `msg:"icon" json:"icon,omitzero"`
	NSFW        *bool  `msg:"nsfw" json:"nsfw,omitzero"`
	Archived    *bool  `msg:"archived" json:"archived,omitzero"`

	Voice    *ChannelVoiceInformation `msg:"voice" json:"voice,omitzero"`
	Slowmode *int                     `msg:"slowmode" json:"slowmode,omitzero"`

	Remove []ChannelClearType `msg:"remove" json:"remove,omitzero"`
}

type SyncSettingsParamsTuple struct {
	Timestamp time.Time `msg:"0" json:"0,omitzero"`
	Value     msgp.Raw  `msg:"1" json:"1,omitzero"` // Enjoy using this.
}

type SyncSettingsParams map[string]SyncSettingsParamsTuple

type SyncSettingsFetchParams struct {
	Keys []string `msg:"keys" json:"keys,omitzero"`
}

type WebhookCreateParams struct {
	Name   string `msg:"name" json:"name,omitzero"`
	Avatar string `msg:"avatar" json:"avatar,omitzero"`
}

type WebhookExecuteParams Message

type WebhookEditParams struct {
	Name        *string              `msg:"name" json:"name,omitzero"`
	Avatar      *string              `msg:"avatar" json:"avatar,omitzero"`
	Permissions *int64               `msg:"permissions" json:"permissions,omitzero"`
	Remove      []WebhookRemoveField `msg:"remove" json:"remove,omitzero"`
}

// AuthMFAParams should only have one of its fields set, and is used for various MFA methods
type AuthMFAParams struct {
	Password     string `msg:"password" json:"password,omitzero"`
	RecoveryCode string `msg:"recovery_code" json:"recovery_code,omitzero"`
	TOTPCode     string `msg:"totp_code" json:"totp_code,omitzero"`
}

// FileParams is used to upload files to the API. For dealing with files, see: File
// When FileParams is uploaded, the API responds with FileParamsData, which is just an ID of the file you uploaded
type FileParams struct {
	// The name of the file; this is completely arbitrary because the backend determines the file-type anyway
	// However, it should not be empty, otherwise the media will not load on the client
	Name string

	// The contents of the file to be read when uploading
	Reader io.Reader `msg:"-"`
}

// FileParamsData is the response from the API when uploading a file.
// To upload a file, you must reference this ID in MessageSend.Attachments.
type FileParamsData struct {
	ID string `msg:"id" json:"id,omitzero"`
}

// PermissionOverwriteParams is derived from
// https://developers.stoat.chat/api-reference/#tag/server-permissions/PUT/servers/{target}/permissions/{role_id}.
type PermissionOverwriteParams struct {
	Allow int64 `msg:"allow" json:"allow"`
	Deny  int64 `msg:"deny" json:"deny"`
}

/*
http.go: Lines 615-631: FileParams and FileParamsData; are these appropriate names for the structs? Especially FileParamsData, which is the response the API gives.
  [Context: I have recently changed "Attachment" struct name to "File" as it better fits the API spec conventions]
*/
