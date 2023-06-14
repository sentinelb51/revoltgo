package revoltgo

import (
	"encoding/json"
	"net/http"
)

// Message struct
type Message struct {
	ID          string          `json:"_id"`
	Nonce       string          `json:"nonce"`
	Channel     string          `json:"channel"`
	Author      string          `json:"author"`
	Webhook     Webhook         `json:"webhook"`
	Content     string          `json:"content"`
	System      MessageSystem   `json:"system"`
	Attachments []*Attachment   `json:"attachments"`
	Edited      string          `json:"edited"` // can this be $date?
	Embeds      []*MessageEmbed `json:"embeds"`
	Mentions    []string        `json:"mentions"`
	Replies     []string        `json:"replies"`

	//todo: add
	Reactions    interface{} `json:"reactions"`
	Interactions interface{} `json:"interactions"`
	Masquerade   interface{} `json:"masquerade"`
}

type MessageSystemType string

const (
	MessageSystemTypeText                      MessageSystemType = "text"
	MessageSystemTypeUserAdded                 MessageSystemType = "user_added"
	MessageSystemTypeUserRemove                MessageSystemType = "user_remove"
	MessageSystemTypeUserJoined                MessageSystemType = "user_joined"
	MessageSystemTypeUserLeft                  MessageSystemType = "user_left"
	MessageSystemTypeUserKicked                MessageSystemType = "user_kicked"
	MessageSystemTypeUserBanned                MessageSystemType = "user_banned"
	MessageSystemTypeChannelRenamed            MessageSystemType = "channel_renamed"
	MessageSystemTypeChannelDescriptionChanged MessageSystemType = "channel_description_changed"
	MessageSystemTypeChannelIconChanged        MessageSystemType = "channel_icon_changed"
	MessageSystemTypeChannelOwnershipChanged   MessageSystemType = "channel_ownership_changed"
)

type MessageSystem struct {
	Type MessageSystemType `json:"type"`
	ID   string            `json:"id"`
}

// Attachment struct.
type Attachment struct {
	ID          string `json:"_id"`
	Tag         string `json:"tag"`
	FileName    string `json:"filename"`
	Metadata    *AttachmentMetadata
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
	Deleted     bool   `json:"deleted"`
	Reported    bool   `json:"reported"`
	MessageID   string `json:"message"`
	UserID      string `json:"user"`
	ServerID    string `json:"server"`
	ObjectID    string `json:"object_id"`
}

// Attachment metadata struct.
type AttachmentMetadata struct {
	Type   string `json:"type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// MessageEdited struct.
type MessageEdited struct {
	Date int `json:"$date"`
}

// Message embed struct.
type MessageEmbed struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	Special     *MessageSpecialEmbed
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Image       *MessageEmbeddedImage `json:"image"`
	Video       *MessageEmbeddedVideo `json:"video"`
	IconUrl     string                `json:"icon_url"`
	Color       string                `json:"color"`
}

// Message special embed struct.
type MessageSpecialEmbed struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	ContentType string `json:"content_type"`
}

// Message embedded image struct
type MessageEmbeddedImage struct {
	Size   string `json:"size"`
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Message embedded video struct
type MessageEmbeddedVideo struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Edit message content.
func (m *Message) Edit(session *Session, content string) error {
	_, err := session.request(http.MethodPatch, "/channels/"+m.Channel+"/messages/"+m.ID, []byte("{\"content\": \""+content+"\"}"))

	if err != nil {
		return err
	}

	m.Content = content
	return nil
}

// Delete the message.
func (m Message) Delete(session *Session) error {
	_, err := session.request("DELETE", "/channels/"+m.Channel+"/messages/"+m.ID, nil)
	return err
}

// Reply to the message.
func (m Message) Reply(session *Session, mention bool, sm *MessageSend) (*Message, error) {
	if sm.Nonce == "" {
		sm.CreateNonce()
	}

	sm.AddReply(m.ID, mention)

	respMessage := &Message{}
	msgData, err := json.Marshal(sm)

	if err != nil {
		return respMessage, err
	}

	response, err := session.request(http.MethodPost, "/channels/"+m.Channel+"/messages", msgData)

	if err != nil {
		return respMessage, err
	}

	err = json.Unmarshal(response, respMessage)

	return respMessage, err
}
