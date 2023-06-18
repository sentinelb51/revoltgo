package revoltgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

/*
	This file contains structs and functions related to interacting with Revolt's REST API
*/

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// request is a helper function to send HTTP requests, handle responses, and unmarshal JSON data
// * data is a struct that will be encoded to JSON and sent as the request body
// * result must be a pointer to a struct that the JSON response will be decoded into
func (s *Session) request(method, url string, data, result any) error {
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	// This may be problematic for Cloudflare if blank user agents are blocked
	request.Header.Set("User-Agent", s.UserAgent)
	request.Header.Set("Content-Type", "application/json")

	// Set auth headers
	if s.SelfBot == nil {
		request.Header.Set("X-Bot-Token", s.Token)
	} else if s.SelfBot.SessionToken != "" {
		request.Header.Set("X-Session-Token", s.SelfBot.SessionToken)
	}

	if data != nil {
		bodyBuffer := bufferPool.Get().(*bytes.Buffer)
		bodyBuffer.Reset()               // Reset the buffer to clear any previous data
		defer bufferPool.Put(bodyBuffer) // Return the buffer to the pool for reuse

		encoder := json.NewEncoder(bodyBuffer)
		if err = encoder.Encode(data); err != nil {
			return fmt.Errorf("request: json.Encode: %s", err)
		}

		request.Body = io.NopCloser(bodyBuffer)
	}

	response, err := s.HTTP.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
		return nil
	case http.StatusBadGateway:
		// TODO: Implement re-tries with sequences
		fallthrough
	case http.StatusTooManyRequests:
		// TODO: Implement rate-limit handling
		fallthrough
	case http.StatusUnauthorized:
		fallthrough
	default: // Error condition
		return fmt.Errorf("bad status code %d: %s", response.StatusCode, body)
	}

	if result != nil {
		if err = json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("request: json.Unmarshal: %s", err)
		}
	}

	return nil
}

type QueryNode struct {
	Revolt   string `json:"revolt"`
	Features struct {
		Captcha struct {
		} `json:"captcha"`
		Email      bool `json:"email"`
		InviteOnly bool `json:"invite_only"`
		Autumn     struct {
		} `json:"autumn"`
		January struct {
		} `json:"january"`
		Voso struct {
		} `json:"voso"`
	} `json:"features"`
	Ws    string `json:"ws"`
	App   string `json:"app"`
	Vapid string `json:"vapid"`
	Build struct {
		CommitSha       string `json:"commit_sha"`
		CommitTimestamp string `json:"commit_timestamp"`
		Semver          string `json:"semver"`
		OriginUrl       string `json:"origin_url"`
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

type UserEditData struct {
	DisplayName string       `json:"display_name"`
	Avatar      string       `json:"avatar"`
	Status      *UserStatus  `json:"status"`
	Profile     *UserProfile `json:"profile"`
	Badges      int          `json:"badges"`
	Flags       int          `json:"flags"`
	Remove      []string     `json:"remove"`
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

type ServerEditData struct {
	Name           string               `json:"name,omitempty"`
	Description    string               `json:"description,omitempty"`
	Icon           string               `json:"icon,omitempty"`
	Banner         string               `json:"banner,omitempty"`
	Categories     []*ServerCategory    `json:"categories,omitempty"`
	SystemMessages ServerSystemMessages `json:"system_messages,omitempty"`
	Flags          int                  `json:"flags"`
	Discoverable   bool                 `json:"discoverable"`
	Analytics      bool                 `json:"analytics"`
	Remove         []string             `json:"remove"`
}

type ChannelCreateData struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	NSFW        bool   `json:"nsfw"`
}

type ServerMemberEditData struct {
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"`
	Roles    []string  `json:"roles"`
	Timeout  time.Time `json:"timeout"`
	Remove   []string  `json:"remove"`
}

type MessageEditData struct {
	Content string          `json:"content,omitempty"`
	Embeds  []*MessageEmbed `json:"embeds,omitempty"`
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
	Name   string   `json:"name"`
	Colour string   `json:"colour"`
	Hoist  bool     `json:"hoist"`
	Rank   int      `json:"rank"`
	Remove []string `json:"remove"`
}

type ServerRoleCreateData struct {
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

type PermissionsSetDefaultData struct {
	Permissions uint `json:"permissions"`
}

type ChannelEditData struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Nsfw        bool     `json:"nsfw,omitempty"`
	Archived    bool     `json:"archived,omitempty"`
	Remove      []string `json:"remove"`
}
