package revoltgo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Server holds information about a server.
type Server struct {
	ID             string                `json:"_id"`
	Owner          string                `json:"owner"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Channels       []string              `json:"channels"`
	Categories     []*ServerCategory     `json:"categories"`
	SystemMessages *ServerSystemMessages `json:"system_messages"`

	// Roles is a map of role IDs to ServerRole structs.
	Roles map[string]*ServerRole `json:"roles"`

	DefaultPermissions uint        `json:"default_permissions"`
	Icon               *Attachment `json:"icon"`
	Banner             *Attachment `json:"banner"`
	Flags              uint        `json:"flags"`
	NSFW               bool        `json:"nsfw"`
	Analytics          bool        `json:"analytics"`
	Discoverable       bool        `json:"discoverable"`
}

type ServerRole struct {
	Name        string       `json:"name"`
	Permissions PermissionAD `json:"permissions"`
	Colour      string       `json:"colour"`
	Hoist       bool         `json:"hoist"`
	Rank        uint         `json:"rank"`
}

// ServerCategory Server categories struct.
type ServerCategory struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Channels []string `json:"channels"`
}

// ServerSystemMessages System messages struct.
type ServerSystemMessages struct {
	UserJoined string `json:"user_joined,omitempty"`
	UserLeft   string `json:"user_left,omitempty"`
	UserKicked string `json:"user_kicked,omitempty"`
	UserBanned string `json:"user_banned,omitempty"`
}

// Server member struct.
type Member struct {
	Informations struct {
		ServerID string `json:"server"`
		UserID   string `json:"user"`
	} `json:"_id"`
	Nickname string      `json:"nickname"`
	Avatar   *Attachment `json:"avatar"`
	Roles    []string    `json:"roles"`
}

// Fetched server members struct.
type FetchedMembers struct {
	Members []*Member `json:"members"`
	Users   []*User   `json:"users"`
}

// Fetched bans struct.
type FetchedBans struct {
	Users []*User `json:"users"`
	Bans  []struct {
		IDs struct {
			UserID   string `json:"user"`
			ServerUd string `json:"server"`
		} `json:"_id"`
		Reason string `json:"reason"`
	} `json:"bans"`
}

// Edit server.
func (s Server) Edit(session *Session, es *EditServer) error {
	data, err := json.Marshal(es)

	if err != nil {
		return err
	}

	_, err = session.handleRequest(http.MethodPatch, "/servers/"+s.ID, data)

	if err != nil {
		return err
	}

	return nil
}

// Delete / leave server.
// If the server not created by client, it will leave.
// Otherwise it will be deleted.
func (s Server) Delete(session *Session) error {
	_, err := session.handleRequest(http.MethodDelete, "/servers/"+s.ID, nil)

	if err != nil {
		return err
	}

	return nil
}

// Create a new text-channel.
func (s Server) CreateTextChannel(session *Session, name, description string) (*ServerChannel, error) {
	channel := &ServerChannel{}

	data, err := session.handleRequest(http.MethodPost, "/servers/"+s.ID+"/channels", []byte("{\"type\":\"Text\",\"name\":\""+name+"\",\"description\":\""+description+"\",\"nonce\":\""+ULID()+"\"}"))

	if err != nil {
		return channel, err
	}

	err = json.Unmarshal(data, channel)

	if err != nil {
		return channel, err
	}

	return channel, nil
}

// Create a new voice-channel.
func (s Server) CreateVoiceChannel(session *Session, name, description string) (*ServerChannel, error) {
	channel := &ServerChannel{}

	data, err := session.handleRequest(http.MethodPost, "/servers/"+s.ID+"/channels", []byte("{\"type\":\"Voice\",\"name\":\""+name+"\",\"description\":\""+description+"\",\"nonce\":\""+ULID()+"\"}"))

	if err != nil {
		return channel, err
	}

	err = json.Unmarshal(data, channel)

	if err != nil {
		return channel, err
	}

	return channel, nil
}

// Fetch a member from Server.
func (s Server) FetchMember(session *Session, id string) (*Member, error) {
	member := &Member{}

	data, err := session.handleRequest(http.MethodGet, "/servers/"+s.ID+"/members/"+id, nil)

	if err != nil {
		return member, err
	}

	err = json.Unmarshal(data, member)

	if err != nil {
		return member, err
	}

	return member, nil
}

// Fetch all of the members from Server.
func (s Server) FetchMembers(session *Session) (*FetchedMembers, error) {
	members := &FetchedMembers{}

	data, err := session.handleRequest(http.MethodGet, "/servers/"+s.ID+"/members", nil)

	if err != nil {
		return members, err
	}

	err = json.Unmarshal(data, members)
	return members, err
}

// Edit a member.
func (s Server) EditMember(session *Session, id string, em *EditMember) error {
	data, err := json.Marshal(em)

	if err != nil {
		return err
	}

	_, err = session.handleRequest(http.MethodPatch, "/servers/"+s.ID+"/members/"+id, data)

	if err != nil {
		return err
	}

	return nil
}

// Kick a member from server.
func (s Server) KickMember(session *Session, id string) error {
	_, err := session.handleRequest(http.MethodDelete, "/servers/"+s.ID+"/members/"+id, nil)

	if err != nil {
		return err
	}

	return nil
}

// Ban a member from server.
func (s Server) BanMember(session *Session, id, reason string) error {
	_, err := session.handleRequest(http.MethodPut, "/servers/"+s.ID+"/bans/"+id, []byte("{\"reason\":\""+reason+"\"}"))

	if err != nil {
		return err
	}

	return nil
}

// Unban a member from server.
func (s Server) UnbanMember(session *Session, id string) error {
	_, err := session.handleRequest(http.MethodDelete, "/servers/"+s.ID+"/bans/"+id, nil)

	if err != nil {
		return err
	}

	return nil
}

// Fetch server bans.
func (s Server) FetchBans(session *Session) (*FetchedBans, error) {
	bans := &FetchedBans{}

	data, err := session.handleRequest(http.MethodGet, "/servers/"+s.ID+"/bans", nil)

	if err != nil {
		return bans, err
	}

	err = json.Unmarshal(data, bans)

	if err != nil {
		return bans, err
	}

	return bans, nil
}

// Timeout a member from server.
func (s Server) TimeoutMember(id string) error {
	// Placeholder for timeout.

	return nil
}

// Set server permissions for a role.
// Leave role field empty if you want to edit default permissions
func (s Server) SetPermissions(session *Session, role_id string, channel_permissions, server_permissions uint) error {
	if role_id == "" {
		role_id = "default"
	}

	_, err := session.handleRequest(http.MethodPut, "/servers/"+s.ID+"/permissions/"+role_id, []byte(fmt.Sprintf("{\"permissions\":{\"server\":%d,\"channel\":%d}}", channel_permissions, server_permissions)))

	if err != nil {
		return err
	}

	return nil
}

// Create a new role for server.
// Returns string (role id), uint (server perms), uint (channel perms) and error.
func (s Server) CreateRole(session *Session, name string) (string, uint, uint, error) {
	role := &struct {
		ID          string `json:"id"`
		Permissions []uint `json:"permissions"`
	}{}

	data, err := session.handleRequest(http.MethodPost, "/servers/"+s.ID+"/roles", []byte("{\"name\":\""+name+"\"}"))

	if err != nil {
		return role.ID, 0, 0, err
	}

	err = json.Unmarshal(data, role)

	if err != nil {
		return role.ID, 0, 0, err
	}

	return role.ID, role.Permissions[0], role.Permissions[1], nil
}

// Edit a server role.
func (s Server) EditRole(session *Session, id string, er *EditRole) error {
	data, err := json.Marshal(er)

	if err != nil {
		return err
	}

	_, err = session.handleRequest(http.MethodPatch, "/servers/"+s.ID+"/roles/"+id, data)

	if err != nil {
		return err
	}

	return nil
}

// Delete a server role.
func (s Server) DeleteRole(session *Session, id string) error {
	_, err := session.handleRequest(http.MethodDelete, "/servers/"+s.ID+"/roles/"+id, nil)

	if err != nil {
		return err
	}

	return nil
}

// Fetch server invite.
func (s Server) FetchInvites(session *Session, id string) error {
	_, err := session.handleRequest(http.MethodGet, "/servers/"+id+"/invites", nil)

	if err != nil {
		return err
	}

	return nil
}

// Mark a server as read.
func (s Server) MarkServerAsRead(session *Session, id string) error {
	_, err := session.handleRequest(http.MethodPut, "/servers/"+id+"/ack", nil)

	if err != nil {
		return err
	}

	return nil
}
