package revoltgo

import (
	"encoding/json"
	"net/http"
)

// User struct.
type User struct {
	ID             string           `json:"_id"`
	Username       string           `json:"username"`
	Avatar         *Attachment      `json:"avatar"`
	Relations      []*UserRelations `json:"relations"`
	Badges         int              `json:"badges"`
	Status         *UserStatus      `json:"status"`
	Relationship   string           `json:"relationship"`
	IsOnline       bool             `json:"online"`
	Flags          int              `json:"flags"`
	BotInformation *BotInformation  `json:"bot"`
}

// User relations struct.
type UserRelations struct {
	ID     string `json:"_id"`
	Status string `json:"status"`
}

// User status struct.
type UserStatus struct {
	Text     string `json:"text"`
	Presence string `json:"presence"`
}

// Bot information struct.
type BotInformation struct {
	Owner string `json:"owner"`
}

// Create a mention format.
func (u User) FormatMention() string {
	return "<@" + u.ID + ">"
}

// Open a DM with the user.
func (u User) OpenDirectMessage(session *Session) (*ServerChannel, error) {
	dmChannel := &ServerChannel{}

	response, err := session.request(http.MethodGet, "/users/"+u.ID+"/dm", nil)

	if err != nil {
		return dmChannel, err
	}

	err = json.Unmarshal(response, dmChannel)
	return dmChannel, err
}

// Fetch default user avatar.
func (u User) FetchDefaultAvatar(session *Session) (*Binary, error) {
	avatarData := &Binary{}

	response, err := session.request(http.MethodGet, "/users/"+u.ID+"/default_avatar", nil)

	if err != nil {
		return avatarData, err
	}

	avatarData.Data = response
	return avatarData, nil
}

// Fetch user relationship.
func (u User) FetchRelationship(session *Session) (*UserRelations, error) {
	relationshipData := &UserRelations{}
	relationshipData.ID = u.ID

	response, err := session.request(http.MethodGet, "/users/"+u.ID+"/relationship", nil)

	if err != nil {
		return relationshipData, err
	}

	err = json.Unmarshal(response, relationshipData)
	return relationshipData, err
}

// Block user.
func (u User) Block(session *Session) (*UserRelations, error) {
	relationshipData := &UserRelations{}
	relationshipData.ID = u.ID

	response, err := session.request("PUT", "/users/"+u.ID+"/block", nil)

	if err != nil {
		return relationshipData, err
	}

	err = json.Unmarshal(response, relationshipData)
	return relationshipData, err
}

// Un-block user.
func (u User) Unblock(session *Session) (*UserRelations, error) {
	relationshipData := &UserRelations{}
	relationshipData.ID = u.ID

	response, err := session.request("DELETE", "/users/"+u.ID+"/block", nil)

	if err != nil {
		return relationshipData, err
	}

	err = json.Unmarshal(response, relationshipData)
	return relationshipData, err
}
