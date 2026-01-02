package revoltgo

import (
	"fmt"
)

//go:generate msgp -tests=false -io=false

// Server is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/servers.rs#L14
type Server struct {
	ID                 string                 `msg:"_id" json:"_id,omitempty"`
	Owner              string                 `msg:"owner" json:"owner,omitempty"`
	Name               string                 `msg:"name" json:"name,omitempty"`
	Description        string                 `msg:"description" json:"description,omitempty"`
	Channels           []string               `msg:"channels" json:"channels,omitempty"`
	Categories         []*ServerCategory      `msg:"categories" json:"categories,omitempty"`
	SystemMessages     ServerSystemMessages   `msg:"system_messages" json:"system_messages,omitempty"`
	Roles              map[string]*ServerRole `msg:"roles" json:"roles,omitempty"` // Roles is a map of role ID to ServerRole structs.
	DefaultPermissions int64                  `msg:"default_permissions" json:"default_permissions,omitempty"`
	Flags              uint32                 `msg:"flags" json:"flags,omitempty"`
	NSFW               bool                   `msg:"nsfw" json:"nsfw,omitempty"`
	Analytics          bool                   `msg:"analytics" json:"analytics,omitempty"`
	Discoverable       bool                   `msg:"discoverable" json:"discoverable,omitempty"`
	Icon               *Attachment            `msg:"icon" json:"icon,omitempty"`
	Banner             *Attachment            `msg:"banner" json:"banner,omitempty"`
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
	Owner              *string                `msg:"owner" json:"owner,omitempty"`
	Name               *string                `msg:"name" json:"name,omitempty"`
	Description        *string                `msg:"description" json:"description,omitempty"`
	Channels           *[]string              `msg:"channels" json:"channels,omitempty"`
	Categories         *[]*ServerCategory     `msg:"categories" json:"categories,omitempty"`
	SystemMessages     *ServerSystemMessages  `msg:"system_messages" json:"system_messages,omitempty"`
	Roles              map[string]*ServerRole `msg:"roles" json:"roles,omitempty"`
	DefaultPermissions *int64                 `msg:"default_permissions" json:"default_permissions,omitempty"`
	Icon               *Attachment            `msg:"icon" json:"icon,omitempty"`
	Banner             *Attachment            `msg:"banner" json:"banner,omitempty"`
	Flags              *uint32                `msg:"flags" json:"flags,omitempty"`
	NSFW               *bool                  `msg:"nsfw" json:"nsfw,omitempty"`
	Analytics          *bool                  `msg:"analytics" json:"analytics,omitempty"`
	Discoverable       *bool                  `msg:"discoverable" json:"discoverable,omitempty"`
}

// ServerRole is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/models/servers/model.rs#L70
type ServerRole struct {
	Name        string              `msg:"name" json:"name,omitempty"`
	Permissions PermissionOverwrite `msg:"permissions" json:"permissions,omitempty"`
	Colour      *string             `msg:"colour" json:"colour,omitempty"`
	Hoist       bool                `msg:"hoist" json:"hoist,omitempty"`
	Rank        int64               `msg:"rank" json:"rank,omitempty"`
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
	Name        *string              `msg:"name" json:"name,omitempty"`
	Permissions *PermissionOverwrite `msg:"permissions" json:"permissions,omitempty"`
	Colour      *string              `msg:"colour" json:"colour,omitempty"`
	Hoist       *bool                `msg:"hoist" json:"hoist,omitempty"`
	Rank        *int64               `msg:"rank" json:"rank,omitempty"`
}

// ServerCategory Server categories struct.
type ServerCategory struct {
	ID       string   `msg:"id" json:"id,omitempty"`
	Title    string   `msg:"title" json:"title,omitempty"`
	Channels []string `msg:"channels" json:"channels,omitempty"`
}

// ServerSystemMessages System messages struct.
type ServerSystemMessages struct {
	UserJoined string `msg:"user_joined" json:"user_joined,omitempty"`
	UserLeft   string `msg:"user_left" json:"user_left,omitempty"`
	UserKicked string `msg:"user_kicked" json:"user_kicked,omitempty"`
	UserBanned string `msg:"user_banned" json:"user_banned,omitempty"`
}

// ServerMember is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/server_members.rs#L44
type ServerMember struct {
	ID       MemberCompositeID `msg:"_id" json:"_id,omitempty"`
	JoinedAt Timestamp         `msg:"joined_at" json:"joined_at,omitempty"`

	Nickname *string     `msg:"nickname" json:"nickname,omitempty"`
	Avatar   *Attachment `msg:"avatar" json:"avatar,omitempty"`
	Timeout  Timestamp   `msg:"timeout" json:"timeout,omitempty"`

	Roles      []string `msg:"roles" json:"roles,omitempty"`
	CanPublish bool     `msg:"can_publish" json:"can_publish,omitempty"`
	CanReceive bool     `msg:"can_receive" json:"can_receive,omitempty"`
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
		case string(WebhookRemoveNickname):
			m.Nickname = nil
		case string(WebhookRemoveAvatar):
			m.Avatar = nil
		default:
			fmt.Printf("ServerMember.clear(): unknown field %s\n", field)
		}
	}
}

type PartialServerMember struct {
	Nickname   *string     `msg:"nickname" json:"nickname,omitempty"`
	Avatar     *Attachment `msg:"avatar" json:"avatar,omitempty"`
	Roles      *[]string   `msg:"roles" json:"roles,omitempty"`
	Timeout    Timestamp   `msg:"timeout" json:"timeout,omitempty"`
	CanPublish *bool       `msg:"can_publish" json:"can_publish,omitempty"`
	CanReceive *bool       `msg:"can_receive" json:"can_receive,omitempty"`
}

// Mention is a proxy function that calls ServerMember.ID.Mention().
func (m *ServerMember) Mention() string {
	return m.ID.Mention()
}

type MemberCompositeID struct {
	User   string `msg:"user" json:"user,omitempty"`
	Server string `msg:"server" json:"server,omitempty"`
}

func (m MemberCompositeID) Mention() string {
	return fmt.Sprintf("<@%s>", m.User)
}

type ServerMembers struct {
	Members []*ServerMember `msg:"members" json:"members,omitempty"`
	Users   []*User         `msg:"users" json:"users,omitempty"`
}

type ServerBans struct {
	Users []*User      `msg:"users" json:"users,omitempty"`
	Bans  []*ServerBan `msg:"bans" json:"bans,omitempty"`
}

type ServerBan struct {
	ID     MemberCompositeID `msg:"_id" json:"_id,omitempty"`
	Reason string            `msg:"reason" json:"reason,omitempty"`
}
