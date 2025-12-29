package revoltgo

import (
	"fmt"
	"time"
)

// Server is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/servers.rs#L14
type Server struct {
	ID                 string                 `json:"_id"`
	Owner              string                 `json:"owner"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Channels           []string               `json:"channels"`
	Categories         []*ServerCategory      `json:"categories"`
	SystemMessages     ServerSystemMessages   `json:"system_messages"`
	Roles              map[string]*ServerRole `json:"roles"` // Roles is a map of role ID to ServerRole structs.
	DefaultPermissions int64                  `json:"default_permissions"`
	Flags              uint32                 `json:"flags"`
	NSFW               bool                   `json:"nsfw"`
	Analytics          bool                   `json:"analytics"`
	Discoverable       bool                   `json:"discoverable"`
	Icon               *Attachment            `json:"icon"`
	Banner             *Attachment            `json:"banner"`
}

func (s *Server) update(data PartialServer) {

	if data.Owner != nil {
		s.Owner = *data.Owner
	}

	if data.Name != nil {
		s.Name = *data.Name
	}

	if data.Description != nil {
		s.Description = *data.Description
	}

	if data.Channels != nil {
		s.Channels = *data.Channels
	}

	if data.Categories != nil {
		s.Categories = *data.Categories
	}

	if data.SystemMessages != nil {
		s.SystemMessages = *data.SystemMessages
	}

	if data.Roles != nil {
		s.Roles = *data.Roles
	}

	if data.DefaultPermissions != nil {
		s.DefaultPermissions = *data.DefaultPermissions
	}

	if data.Icon != nil {
		s.Icon = data.Icon
	}

	if data.Banner != nil {
		s.Banner = data.Banner
	}

	if data.Flags != nil {
		s.Flags = *data.Flags
	}

	if data.NSFW != nil {
		s.NSFW = *data.NSFW
	}

	if data.Analytics != nil {
		s.Analytics = *data.Analytics
	}

	if data.Discoverable != nil {
		s.Discoverable = *data.Discoverable
	}
}

func (s *Server) clear(fields []string) {
	for _, field := range fields {
		switch field {
		case "Icon":
			s.Icon = nil
		case "Banner":
			s.Banner = nil
		case "Description":
			s.Description = ""
		default:
			fmt.Printf("Server.clear(): unknown field %s\n", field)
		}
	}
}

// PartialServer is only found within EventServerUpdate and used to update the state.
type PartialServer struct {
	Owner              *string                 `json:"owner,omitempty"`
	Name               *string                 `json:"name,omitempty"`
	Description        *string                 `json:"description,omitempty"`
	Channels           *[]string               `json:"channels,omitempty"`
	Categories         *[]*ServerCategory      `json:"categories,omitempty"`
	SystemMessages     *ServerSystemMessages   `json:"system_messages,omitempty"`
	Roles              *map[string]*ServerRole `json:"roles,omitempty"`
	DefaultPermissions *int64                  `json:"default_permissions,omitempty"`
	Icon               *Attachment             `json:"icon,omitempty"`
	Banner             *Attachment             `json:"banner,omitempty"`
	Flags              *uint32                 `json:"flags,omitempty"`
	NSFW               *bool                   `json:"nsfw,omitempty"`
	Analytics          *bool                   `json:"analytics,omitempty"`
	Discoverable       *bool                   `json:"discoverable,omitempty"`
}

// ServerRole is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/models/servers/model.rs#L70
type ServerRole struct {
	Name        string              `json:"name"`
	Permissions PermissionOverwrite `json:"permissions"`
	Colour      *string             `json:"colour"`
	Hoist       bool                `json:"hoist"`
	Rank        int64               `json:"rank"`
}

func (r *ServerRole) update(data PartialServerRole) {
	if data.Name != nil {
		r.Name = *data.Name
	}

	if data.Permissions != nil {
		r.Permissions = *data.Permissions
	}

	if data.Colour != nil {
		r.Colour = data.Colour
	}

	if data.Hoist != nil {
		r.Hoist = *data.Hoist
	}

	if data.Rank != nil {
		r.Rank = *data.Rank
	}
}

func (r *ServerRole) clear(fields []string) {
	for _, field := range fields {
		switch field {
		case "Colour":
			r.Colour = nil
		default:
			fmt.Printf("ServerRole.clear(): unknown field %s\n", field)
		}
	}
}

type PartialServerRole struct {
	Name        *string              `json:"name,omitempty"`
	Permissions *PermissionOverwrite `json:"permissions,omitempty"`
	Colour      *string              `json:"colour,omitempty"`
	Hoist       *bool                `json:"hoist,omitempty"`
	Rank        *int64               `json:"rank,omitempty"`
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

// ServerMember is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/server_members.rs#L44
type ServerMember struct {
	ID       MemberCompositeID `json:"_id"`
	JoinedAt time.Time         `json:"joined_at"`

	Nickname *string     `json:"nickname"`
	Avatar   *Attachment `json:"avatar"`
	Timeout  *time.Time  `json:"timeout"`

	Roles      []string `json:"roles"`
	CanPublish bool     `json:"can_publish"`
	CanReceive bool     `json:"can_receive"`
}

func (m *ServerMember) update(data PartialServerMember) {

	if data.Nickname != nil {
		m.Nickname = data.Nickname
	}

	if data.Avatar != nil {
		m.Avatar = data.Avatar
	}

	if data.Roles != nil {
		m.Roles = *data.Roles
	}

	if data.Timeout != nil {
		m.Timeout = data.Timeout
	}

	if data.CanPublish != nil {
		m.CanPublish = *data.CanPublish
	}

	if data.CanReceive != nil {
		m.CanReceive = *data.CanReceive
	}
}

// Clear resets nullable fields to nil based on the JSON key name.
func (m *ServerMember) clear(fields []string) {
	for _, field := range fields {
		switch field {
		case "Nickname":
			m.Nickname = nil
		case "Avatar":
			m.Avatar = nil
		default:
			fmt.Printf("ServerMember.clear(): unknown field %s\n", field)
		}
	}
}

type PartialServerMember struct {
	Nickname   *string     `json:"nickname,omitempty"`
	Avatar     *Attachment `json:"avatar,omitempty"`
	Roles      *[]string   `json:"roles,omitempty"`
	Timeout    *time.Time  `json:"timeout,omitempty"`
	CanPublish *bool       `json:"can_publish,omitempty"`
	CanReceive *bool       `json:"can_receive,omitempty"`
}

// Mention is a proxy function that calls ServerMember.ID.Mention().
func (m *ServerMember) Mention() string {
	return m.ID.Mention()
}

type MemberCompositeID struct {
	User   string `json:"user"`
	Server string `json:"server"`
}

func (m MemberCompositeID) Mention() string {
	return fmt.Sprintf("<@%s>", m.User)
}

type ServerMembers struct {
	Members []*ServerMember `json:"members"`
	Users   []*User         `json:"users"`
}

type ServerBans struct {
	Users []*User      `json:"users"`
	Bans  []*ServerBan `json:"bans"`
}

type ServerBan struct {
	ID     MemberCompositeID `json:"_id"`
	Reason string            `json:"reason"`
}
