package revoltgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

type ChannelType string

const (
	ChannelTypeSavedMessages ChannelType = "SavedMessages"
	ChannelTypeText          ChannelType = "TextChannel"
	ChannelTypeVoice         ChannelType = "VoiceChannel"
	ChannelTypeDM            ChannelType = "DirectMessage"
	ChannelTypeGroup         ChannelType = "Group"
)

// ServerChannel struct.
type ServerChannel struct {
	ChannelType        ChannelType  `json:"channel_type"`
	ID                 string       `json:"_id"`
	Server             string       `json:"server"`
	Name               string       `json:"name"`
	Description        string       `json:"description"`
	Icon               *Attachment  `json:"icon"`
	DefaultPermissions PermissionAD `json:"default_permissions"`

	// RolePermissions is a map of role IDs to PermissionAD structs.
	RolePermissions map[string]PermissionAD `json:"role_permissions"`

	NSFW bool `json:"nsfw"`
}

// Fetched messages struct.
type FetchedMessages struct {
	Messages []*Message `json:"messages"`
	Users    []*User    `json:"users"`
}

// Send a message to the channel.
func (c ServerChannel) SendMessage(session *Session, message *MessageSend) (*Message, error) {
	if message.Nonce == "" {
		message.CreateNonce()
	}

	respMessage := &Message{}
	msgData, err := json.Marshal(message)

	if err != nil {
		return respMessage, err
	}

	response, err := session.request(http.MethodPost, "/channels/"+c.ID+"/messages", msgData)

	if err != nil {
		return respMessage, err
	}

	err = json.Unmarshal(response, respMessage)
	return respMessage, err
}

// Fetch messages from channel.
// Check: https://developers.revolt.chat/api/#tag/Messaging/paths/~1channels~1:channel~1messages/get for map parameters.
func (c ServerChannel) FetchMessages(session *Session, options map[string]interface{}) (*FetchedMessages, error) {
	// Format url
	url := "/channels/" + c.ID + "/messages?"

	for key, value := range options {
		if !reflect.ValueOf(value).IsZero() {
			url += fmt.Sprintf("%s=%v&", key, value)
		}
	}

	url = url[:len(url)-1]

	fetchedMsgs := &FetchedMessages{}

	// Send request
	response, err := session.request(http.MethodGet, url, nil)

	if err != nil {
		return fetchedMsgs, err
	}

	err = json.Unmarshal(response, &fetchedMsgs)

	if err != nil {
		err = json.Unmarshal([]byte(fmt.Sprintf("{\"messages\": %s}", response)), &fetchedMsgs)

		if err != nil {
			return fetchedMsgs, err
		}
	}

	return fetchedMsgs, nil
}

// Fetch a message from channel by ID.
func (c ServerChannel) FetchMessage(session *Session, id string) (*Message, error) {
	msg := &Message{}

	response, err := session.request(http.MethodGet, "/channels/"+c.ID+"/messages/"+id, nil)

	if err != nil {
		return msg, err
	}

	err = json.Unmarshal(response, msg)
	return msg, err
}

// Edit channel.
func (c ServerChannel) Edit(session *Session, ec *EditChannel) error {
	data, err := json.Marshal(ec)

	if err != nil {
		return err
	}

	_, err = session.request(http.MethodPatch, "/channels/"+c.ID, data)
	return err
}

// Delete channel.
func (c ServerChannel) Delete(session *Session) error {
	_, err := session.request("DELETE", "/channels/"+c.ID, nil)
	return err
}

// Create a new invite.
// Returns a string (invite code) and error (nil if not exists).
func (c ServerChannel) CreateInvite(session *Session) (string, error) {
	data, err := session.request(http.MethodPost, "/channels/"+c.ID+"/invites", nil)

	if err != nil {
		return "", err
	}

	dataStruct := &struct {
		InviteCode string `json:"code"`
	}{}

	err = json.Unmarshal(data, dataStruct)
	return dataStruct.InviteCode, err
}

// Set channel permissions for a role.
// Leave role field empty if you want to edit default permissions
func (c ServerChannel) SetPermissions(session *Session, role_id string, permissions uint) error {
	if role_id == "" {
		role_id = "default"
	}

	_, err := session.request("PUT", "/channels/"+c.ID+"/permissions/"+role_id, []byte(fmt.Sprintf("{\"permissions\":%d}", permissions)))
	return err
}

// Send a typing start event to the channel.
func (c *ServerChannel) BeginTyping(session *Session) {
	session.Socket.SendText(fmt.Sprintf("{\"type\":\"BeginTyping\",\"channel\":\"%s\"}", c.ID))
}

// End the typing event in the channel.
func (c *ServerChannel) EndTyping(session *Session) {
	session.Socket.SendText(fmt.Sprintf("{\"type\":\"EndTyping\",\"channel\":\"%s\"}", c.ID))
}
