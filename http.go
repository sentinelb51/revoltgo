package revoltgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

/*
	This file contains structs and functions related to interacting with Revolt's REST API
*/

// handleRequest sends HTTP requests and return the response body as a byte slice
func (s *Session) handleRequest(method, path string, data any) ([]byte, error) {

	// todo: maybe add URLAPI before path to make sure it's not sent to external sites?

	var (
		payload []byte
		err     error
	)

	if data != nil {
		payload, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("handleRequest: json.Marshal: %s", err)
		}
	}

	request, err := http.NewRequest(method, path, bytes.NewBuffer(payload))
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

	// Send handleRequest
	response, err := s.HTTP.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusBadGateway:
		// TODO: Implement re-tries with sequences
		fallthrough
	case http.StatusTooManyRequests:
		// TODO: Implement rate-limit handling
		fallthrough
	case http.StatusUnauthorized:
		fallthrough
	default: // Error condition
		err = fmt.Errorf("bad status code %d: %s", response.StatusCode, body)
	}

	return body, err
}

// request is a helper function to send HTTP requests using handleRequest and unmarshal the response into a struct
// * data will always be encoded in JSON
// * result will always be decoded from JSON, and must be a pointer
func (s *Session) request(method, url string, data, result any) (err error) {
	response, err := s.handleRequest(method, url, data)
	if err == nil {
		err = json.Unmarshal(response, result)
	}
	return
}

type LoginData struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	FriendlyName string `json:"friendly_name"`
}

type BotCreateData struct {
	Name string `json:"name"`
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
