package revoltgo

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

/*
	This file contains structs and functions related to interacting with Revolt's REST API
*/

// request is a helper function to send HTTP requests to the API
func (s *Session) request(method, path string, data []byte) ([]byte, error) {

	// todo: maybe add URLAPI before path to make sure it's not sent to external sites?
	payload := bytes.NewBuffer(data)
	request, err := http.NewRequest(method, path, payload)
	if err != nil {
		return nil, err
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

	// Send request
	response, err := s.HTTP.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if !(response.StatusCode >= 200 && response.StatusCode < 300) {
		err = fmt.Errorf("%s: %s", response.Status, body)
	}

	return body, err
}

// GroupCreateData describes how a group should be created
type GroupCreateData struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Users       []string `json:"users"`
	Nonce       string   `json:"nonce"`
}

type ServerCreateData struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Nonce       string `json:"nonce"`
}
