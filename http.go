package revoltgo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

// resolveDestination converts a relative URL to an absolute URL. Prefixes relative URLs with the API base URL.
// It also allows absolute URLs targeting the CDN. Otherwise, it rejects the URL.
func resolveDestination(destination string) (string, error) {
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

/*
Request sends a JSON Request with "method" to a destination URL
- "result" will be used to decode the response into, and
- "data" is the Request body which wil be encoded as JSON

- If the "data" is a *File, it will be uploaded as a multipart form
This function automatically handles rate-limiting and response status codes
*/
func (s *Session) Request(method, destination string, data, result any) error {

	// Handle ratelimits before appending the API URL to reduce key size
	rl := s.Ratelimiter.get(method, destination)

	destination, err := resolveDestination(destination)
	if err != nil {
		return err
	}

	if !rl.resetAfter.IsZero() {
		if wait := rl.delay(); wait > 0 {
			time.Sleep(wait)
		}
	}

	reader, contentType, err := prepareRequestBody(data)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(context.Background(), method, destination, reader)
	if err != nil {
		return err
	}

	request.Header.Set("User-Agent", s.UserAgent)
	request.Header.Set("Content-Type", contentType)

	if s.Selfbot() {
		request.Header.Set("X-Session-Token", s.Token)
	} else {
		request.Header.Set("X-Bot-Token", s.Token)
	}

	response, err := s.HTTP.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err = rl.update(response.Header); err != nil {
		return err
	}

	// Process response based on status code
	return handleResponse(response.StatusCode, response.Body, result)
}

// prepareRequestBody prepares an appropriate Request body and determines the content type
func prepareRequestBody(body any) (io.Reader, string, error) {
	if body == nil {
		return http.NoBody, "application/json", nil
	}

	if file, ok := body.(*File); ok {
		return prepareFileUpload(file)
	}

	return prepareJSONBody(body)
}

// prepareFileUpload prepares a multipart form for uploading a file
func prepareFileUpload(file *File) (io.Reader, string, error) {
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
func prepareJSONBody(body any) (io.Reader, string, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("json.Marshal: %w", err)
	}

	return bytes.NewReader(data), "application/json", nil
}

// handleResponse processes the API response
func handleResponse(statusCode int, body io.Reader, result any) error {
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

type RootData struct {
	Revolt   string `json:"revolt"`
	Features struct {
		Captcha struct {
			Enabled bool   `json:"enabled"`
			Key     string `json:"key"`
		} `json:"captcha"`
		Email      bool `json:"email"`
		InviteOnly bool `json:"invite_only"`
		Autumn     struct {
			Enabled bool   `json:"enabled"`
			URL     string `json:"url"`
		} `json:"autumn"`
		January struct {
			Enabled bool   `json:"enabled"`
			URL     string `json:"url"`
		} `json:"january"`
		Voso struct {
			Enabled bool   `json:"enabled"`
			URL     string `json:"url"`
			WS      string `json:"ws"`
		} `json:"voso"`
	} `json:"features"`
	WS    string `json:"ws"`
	App   string `json:"app"`
	VapID string `json:"vapid"`
	Build struct {
		CommitSha       string `json:"commit_sha"`
		CommitTimestamp string `json:"commit_timestamp"`
		SemVer          string `json:"semver"`
		OriginURL       string `json:"origin_url"`
		Timestamp       string `json:"timestamp"`
	} `json:"build"`
}

type LoginData struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	FriendlyName string `json:"friendly_name"`
}

type BotEditData struct {
	Name            string   `json:"name"`
	Public          bool     `json:"public"`
	Analytics       bool     `json:"analytics"`
	InteractionsURL string   `json:"interactions_url"`
	Remove          []string `json:"remove"`
}

type BotInviteData struct {
	Server string `json:"server"`
	Group  string `json:"group"`
}

type BotCreateData struct {
	Name string `json:"name"`
}

type AccountCreateData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Invite   string `json:"invite"`
	Captcha  string `json:"captcha"`
}

type AccountReverifyData struct {
	Email   string `json:"email"`
	Captcha string `json:"captcha"`
}

type OnboardingCompleteData struct {
	Username string `json:"username"`
}

type SessionEditData struct {
	FriendlyName string `json:"friendly_name"`
}

type PasswordResetConfirmData struct {
	Token    string `json:"token"`
	Password string `json:"password"`

	// Whether to log out of all sessions
	RemoveSessions bool `json:"remove_sessions"`
}

type AccountChangePasswordData struct {
	Password        string `json:"password"`
	CurrentPassword string `json:"current_password"`
}

type AccountChangeEmailData struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
}

type AccountDeleteConfirmData struct {
	Token string `json:"token"`
}

type UserEditData struct {
	DisplayName string       `json:"display_name,omitempty"`
	Avatar      string       `json:"avatar,omitempty"`
	Status      *UserStatus  `json:"status,omitempty"`
	Profile     *UserProfile `json:"profile,omitempty"`
	Badges      *int         `json:"badges,omitempty"`
	Flags       *int         `json:"flags,omitempty"`
	Remove      []string     `json:"remove,omitempty"`
}

type UsernameData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GroupCreateData describes how a group should be created
type GroupCreateData struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Users       []string `json:"users"`
	NSFW        bool     `json:"nsfw"`
}

type ServerCreateData struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
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
	Name           string                 `json:"name,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Icon           string                 `json:"icon,omitempty"`
	Banner         string                 `json:"banner,omitempty"`
	Categories     []*ServerCategory      `json:"categories,omitempty"`
	SystemMessages *ServerSystemMessages  `json:"system_messages,omitempty"`
	Flags          int                    `json:"flags,omitempty"`
	Discoverable   *bool                  `json:"discoverable,omitempty"`
	Analytics      *bool                  `json:"analytics,omitempty"`
	Remove         []ServerEditDataRemove `json:"remove"`
}

type ServerChannelCreateDataType string

const (
	ServerChannelCreateDataTypeText  ServerChannelCreateDataType = "Text"
	ServerChannelCreateDataTypeVoice ServerChannelCreateDataType = "Voice"
)

type ServerChannelCreateData struct {
	Type        ServerChannelCreateDataType `json:"type"`
	Name        string                      `json:"name"`
	Description string                      `json:"description,omitempty"`
	NSFW        bool                        `json:"nsfw,omitempty"`
}

type ServerMemberEditData struct {
	Nickname string    `json:"nickname,omitempty"`
	Avatar   string    `json:"avatar,omitempty"`
	Roles    []string  `json:"roles,omitempty"`
	Timeout  time.Time `json:"timeout,omitempty"`
	Remove   []string  `json:"remove,omitempty"`
}

type MessageEditData struct {
	Content string          `json:"content,omitempty"`
	Embeds  []*MessageEmbed `json:"embeds,omitempty"`
}

type EmojiCreateData struct {
	Name   string       `json:"name"`
	Parent *EmojiParent `json:"parent"`
	Nsfw   bool         `json:"nsfw"`
}

type ChannelMessagesParamsSortType string

const (
	ChannelMessagesParamsSortTypeRelevance ChannelMessagesParamsSortType = "Relevance"
	ChannelMessagesParamsSortTypeOldest    ChannelMessagesParamsSortType = "Oldest"
	ChannelMessagesParamsSortTypeLatest    ChannelMessagesParamsSortType = "Latest"
)

type ChannelMessagesParams struct {
	// Maximum number of messages to fetch. For nearby messages, this is (limit + 1)
	Limit int `json:"limit,omitempty"`

	// Message ID before which messages should be fetched
	Before string `json:"before,omitempty"`

	// Message ID after which messages should be fetched
	After string `json:"after,omitempty"`

	// Message sort direction
	Sort ChannelMessagesParamsSortType `json:"sort,omitempty"`

	// Message ID to search around. Specifying this ignores Before, After, and Sort
	Nearby string `json:"nearby,omitempty"`

	// Whether to include user (and member, if server channel) objects
	IncludeUsers bool `json:"include_users,omitempty"`
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
	Name   string   `json:"name,omitempty"`
	Colour string   `json:"colour,omitempty"`
	Hoist  *bool    `json:"hoist"`
	Rank   *int     `json:"rank"`
	Remove []string `json:"remove,omitempty"`
}

type ServerRoleCreateData struct {
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

type PermissionsSetDefaultData struct {
	Permissions uint `json:"permissions"`
}

type ChannelMessageBulkDeleteData struct {
	IDs []string `json:"ids"`
}

type ChannelEditData struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	NSFW        bool     `json:"nsfw,omitempty"`
	Archived    bool     `json:"archived,omitempty"`
	Remove      []string `json:"remove"`
}

type SyncSettingsData map[string]UpdateTuple

type SyncSettingsFetchData struct {
	Keys []string `json:"keys"`
}
