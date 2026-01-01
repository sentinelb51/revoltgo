package revoltgo

import (
	"fmt"
)

//go:generate msgp -tests=false -io=false

// Server is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/servers.rs#L14
type Server struct {
	ID                 string                 `msg:"_id"`
	Owner              string                 `msg:"owner"`
	Name               string                 `msg:"name"`
	Description        string                 `msg:"description"`
	Channels           []string               `msg:"channels"`
	Categories         []*ServerCategory      `msg:"categories"`
	SystemMessages     ServerSystemMessages   `msg:"system_messages"`
	Roles              map[string]*ServerRole `msg:"roles"` // Roles is a map of role ID to ServerRole structs.
	DefaultPermissions int64                  `msg:"default_permissions"`
	Flags              uint32                 `msg:"flags"`
	NSFW               bool                   `msg:"nsfw"`
	Analytics          bool                   `msg:"analytics"`
	Discoverable       bool                   `msg:"discoverable"`
	Icon               *Attachment            `msg:"icon"`
	Banner             *Attachment            `msg:"banner"`
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
		s.Roles = data.Roles
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
	Owner              *string                `msg:"owner,omitempty"`
	Name               *string                `msg:"name,omitempty"`
	Description        *string                `msg:"description,omitempty"`
	Channels           *[]string              `msg:"channels,omitempty"`
	Categories         *[]*ServerCategory     `msg:"categories,omitempty"`
	SystemMessages     *ServerSystemMessages  `msg:"system_messages,omitempty"`
	Roles              map[string]*ServerRole `msg:"roles,omitempty"`
	DefaultPermissions *int64                 `msg:"default_permissions,omitempty"`
	Icon               *Attachment            `msg:"icon,omitempty"`
	Banner             *Attachment            `msg:"banner,omitempty"`
	Flags              *uint32                `msg:"flags,omitempty"`
	NSFW               *bool                  `msg:"nsfw,omitempty"`
	Analytics          *bool                  `msg:"analytics,omitempty"`
	Discoverable       *bool                  `msg:"discoverable,omitempty"`
}

// ServerRole is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/models/servers/model.rs#L70
type ServerRole struct {
	Name        string              `msg:"name"`
	Permissions PermissionOverwrite `msg:"permissions"`
	Colour      *string             `msg:"colour"`
	Hoist       bool                `msg:"hoist"`
	Rank        int64               `msg:"rank"`
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
	Name        *string              `msg:"name,omitempty"`
	Permissions *PermissionOverwrite `msg:"permissions,omitempty"`
	Colour      *string              `msg:"colour,omitempty"`
	Hoist       *bool                `msg:"hoist,omitempty"`
	Rank        *int64               `msg:"rank,omitempty"`
}

// ServerCategory Server categories struct.
type ServerCategory struct {
	ID       string   `msg:"id"`
	Title    string   `msg:"title"`
	Channels []string `msg:"channels"`
}

// ServerSystemMessages System messages struct.
type ServerSystemMessages struct {
	UserJoined string `msg:"user_joined,omitempty"`
	UserLeft   string `msg:"user_left,omitempty"`
	UserKicked string `msg:"user_kicked,omitempty"`
	UserBanned string `msg:"user_banned,omitempty"`
}

// ServerMember is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/server_members.rs#L44
type ServerMember struct {
	ID       MemberCompositeID `msg:"_id"`
	JoinedAt Timestamp         `msg:"joined_at"`

	Nickname *string     `msg:"nickname"`
	Avatar   *Attachment `msg:"avatar"`
	Timeout  Timestamp   `msg:"timeout"`

	Roles      []string `msg:"roles"`
	CanPublish bool     `msg:"can_publish"`
	CanReceive bool     `msg:"can_receive"`
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

	if !data.Timeout.IsZero() {
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
	Nickname   *string     `msg:"nickname,omitempty"`
	Avatar     *Attachment `msg:"avatar,omitempty"`
	Roles      *[]string   `msg:"roles,omitempty"`
	Timeout    Timestamp   `msg:"timeout,omitempty"`
	CanPublish *bool       `msg:"can_publish,omitempty"`
	CanReceive *bool       `msg:"can_receive,omitempty"`
}

// Mention is a proxy function that calls ServerMember.ID.Mention().
func (m *ServerMember) Mention() string {
	return m.ID.Mention()
}

type MemberCompositeID struct {
	User   string `msg:"user"`
	Server string `msg:"server"`
}

func (m MemberCompositeID) Mention() string {
	return fmt.Sprintf("<@%s>", m.User)
}

type ServerMembers struct {
	Members []*ServerMember `msg:"members"`
	Users   []*User         `msg:"users"`
}

type ServerBans struct {
	Users []*User      `msg:"users"`
	Bans  []*ServerBan `msg:"bans"`
}

type ServerBan struct {
	ID     MemberCompositeID `msg:"_id"`
	Reason string            `msg:"reason"`
}
