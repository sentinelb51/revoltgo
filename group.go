package revoltgo

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
)

// Group channel struct.
type Group struct {
	Client    *Session
	CreatedAt time.Time

	ID          string   `json:"_id"`
	Nonce       string   `json:"nonce"`
	OwnerID     string   `json:"owner"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Users       []string `json:"users"`
}

// Fetched group members struct.
type FetchedGroupMembers struct {
	Messages []*Message `json:"messages"`
	Users    []*User    `json:"users"`
}

// System messages struct.
type GroupSystemMessages struct {
	UserJoined string `json:"user_joined,omitempty"`
	UserLeft   string `json:"user_left,omitempty"`
}

// Calculate creation date and edit the struct.
func (c *Group) CalculateCreationDate() error {
	ulid, err := ulid.Parse(c.ID)

	if err != nil {
		return err
	}

	c.CreatedAt = time.UnixMilli(int64(ulid.Time()))
	return nil
}

// Fetch all of the members from group.
func (c ServerChannel) FetchMembers(session *Session) ([]*User, error) {
	var groupMembers []*User

	response, err := session.handleRequest(http.MethodGet, "/channels/"+c.ID+"/members", nil)

	if err != nil {
		return groupMembers, err
	}

	err = json.Unmarshal(response, &groupMembers)
	return groupMembers, err
}

// Add a new group recipient.
func (c ServerChannel) AddGroupRecipient(session *Session, uid string) error {
	_, err := session.handleRequest(http.MethodPut, "/channels/"+c.ID+"/recipients/"+uid, nil)
	return err
}

// Delete a group recipient.
func (c ServerChannel) DeleteGroupRecipient(session *Session, uid string) error {
	_, err := session.handleRequest(http.MethodDelete, "/channels/"+c.ID+"/recipients/"+uid, nil)
	return err
}
