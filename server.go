package revoltgo

import (
	"fmt"
	"log"
	"time"
)

//go:generate msgp -tests=false -io=false

// Server is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/servers.rs#L14
type Server struct {
	ID                 string                 `msg:"_id" json:"_id,omitzero"`
	Owner              string                 `msg:"owner" json:"owner,omitzero"`
	Name               string                 `msg:"name" json:"name,omitzero"`
	Description        string                 `msg:"description" json:"description,omitzero"`
	Channels           []string               `msg:"channels" json:"channels,omitzero"`
	Categories         []*ServerCategory      `msg:"categories" json:"categories,omitzero"`
	SystemMessages     ServerSystemMessages   `msg:"system_messages" json:"system_messages,omitzero"`
	Roles              map[string]*ServerRole `msg:"roles" json:"roles,omitzero"` // Roles is a map of role ID to ServerRole structs.
	DefaultPermissions int64                  `msg:"default_permissions" json:"default_permissions,omitzero"`
	Flags              uint32                 `msg:"flags" json:"flags,omitzero"`
	NSFW               bool                   `msg:"nsfw" json:"nsfw,omitzero"`
	Analytics          bool                   `msg:"analytics" json:"analytics,omitzero"`
	Discoverable       bool                   `msg:"discoverable" json:"discoverable,omitzero"`
	Icon               *File                  `msg:"icon" json:"icon,omitzero"`
	Banner             *File                  `msg:"banner" json:"banner,omitzero"`
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

func (s *Server) clear(fields []ServerEditParamsRemove) {
	for _, field := range fields {
		switch field {
		case ServerEditDataRemoveIcon:
			s.Icon = nil
		case ServerEditDataRemoveBanner:
			s.Banner = nil
		case ServerEditDataRemoveDescription:
			s.Description = ""
		case ServerEditDataRemoveCategories:
			s.Categories = nil
		case ServerEditDataRemoveSystemMessages:
			s.SystemMessages = ServerSystemMessages{}
		default:
			log.Printf("Server.clear(): unknown field %s\n", field)
		}
	}
}

// PartialServer is only found within EventServerUpdate and used to update the state.
type PartialServer struct {
	Owner              *string                `msg:"owner" json:"owner,omitzero"`
	Name               *string                `msg:"name" json:"name,omitzero"`
	Description        *string                `msg:"description" json:"description,omitzero"`
	Channels           *[]string              `msg:"channels" json:"channels,omitzero"`
	Categories         *[]*ServerCategory     `msg:"categories" json:"categories,omitzero"`
	SystemMessages     *ServerSystemMessages  `msg:"system_messages" json:"system_messages,omitzero"`
	Roles              map[string]*ServerRole `msg:"roles" json:"roles,omitzero"`
	DefaultPermissions *int64                 `msg:"default_permissions" json:"default_permissions,omitzero"`
	Icon               *File                  `msg:"icon" json:"icon,omitzero"`
	Banner             *File                  `msg:"banner" json:"banner,omitzero"`
	Flags              *uint32                `msg:"flags" json:"flags,omitzero"`
	NSFW               *bool                  `msg:"nsfw" json:"nsfw,omitzero"`
	Analytics          *bool                  `msg:"analytics" json:"analytics,omitzero"`
	Discoverable       *bool                  `msg:"discoverable" json:"discoverable,omitzero"`
}

// ServerRole is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/models/servers/model.rs#L70
type ServerRole struct {
	ID          string              `msg:"_id" json:"_id,omitzero"`
	Name        string              `msg:"name" json:"name,omitzero"`
	Permissions PermissionOverwrite `msg:"permissions" json:"permissions,omitzero"`
	Colour      *string             `msg:"colour" json:"colour,omitzero"`
	Hoist       bool                `msg:"hoist" json:"hoist,omitzero"`
	Icon        *File               `msg:"icon" json:"icon,omitzero"`
	Rank        int64               `msg:"rank" json:"rank,omitzero"`
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

type ServerRoleClearType string

const (
	ServerRoleClearColour ServerRoleClearType = "Colour"
	ServerRoleClearIcon   ServerRoleClearType = "Icon"
)

func (r *ServerRole) clear(fields []ServerRoleClearType) {
	for _, field := range fields {
		switch field {
		case ServerRoleClearColour:
			r.Colour = nil
		case ServerRoleClearIcon:
			r.Icon = nil
		default:
			log.Printf("ServerRole.clear(): unknown field %s\n", field)
		}
	}
}

type PartialServerRole struct {
	Name        *string              `msg:"name" json:"name,omitzero"`
	Permissions *PermissionOverwrite `msg:"permissions" json:"permissions,omitzero"`
	Colour      *string              `msg:"colour" json:"colour,omitzero"`
	Hoist       *bool                `msg:"hoist" json:"hoist,omitzero"`
	Rank        *int64               `msg:"rank" json:"rank,omitzero"`
}

// ServerCategory Server categories struct.
type ServerCategory struct {
	ID       string   `msg:"id" json:"id,omitzero"`
	Title    string   `msg:"title" json:"title,omitzero"`
	Channels []string `msg:"channels" json:"channels,omitzero"`
}

// ServerSystemMessages System messages struct.
type ServerSystemMessages struct {
	UserJoined string `msg:"user_joined" json:"user_joined,omitzero"`
	UserLeft   string `msg:"user_left" json:"user_left,omitzero"`
	UserKicked string `msg:"user_kicked" json:"user_kicked,omitzero"`
	UserBanned string `msg:"user_banned" json:"user_banned,omitzero"`
}

// ServerMember is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/models/src/v0/server_members.rs#L44
type ServerMember struct {
	ID       MemberCompositeID `msg:"_id" json:"_id,omitzero"`
	JoinedAt time.Time         `msg:"joined_at" json:"joined_at,omitzero"`

	Nickname *string  `msg:"nickname" json:"nickname,omitzero"`
	Pronouns *string  `msg:"pronouns" json:"pronouns,omitzero"`
	Avatar   *File    `msg:"avatar" json:"avatar,omitzero"`
	Roles    []string `msg:"roles" json:"roles,omitzero"`

	Timeout *time.Time `msg:"timeout" json:"timeout,omitzero"`

	// False means server-wide voice-muted; nil if unset
	CanPublish *bool `msg:"can_publish" json:"can_publish,omitzero"`

	// False means server-wide voice-deafened; nil if unset
	CanReceive *bool `msg:"can_receive" json:"can_receive,omitzero"`
}

func (m *ServerMember) update(data PartialServerMember) {

	if data.Nickname != nil {
		m.Nickname = data.Nickname
	}

	if data.Pronouns != nil {
		m.Pronouns = data.Pronouns
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
		m.CanPublish = data.CanPublish
	}

	if data.CanReceive != nil {
		m.CanReceive = data.CanReceive
	}
}

type ServerMemberClearType string

const (
	ServerMemberClearNickname     ServerMemberClearType = "Nickname"
	ServerMemberClearPronouns     ServerMemberClearType = "Pronouns"
	ServerMemberClearAvatar       ServerMemberClearType = "Avatar"
	ServerMemberClearRoles        ServerMemberClearType = "Roles"
	ServerMemberClearTimeout      ServerMemberClearType = "Timeout"
	ServerMemberClearCanReceive   ServerMemberClearType = "CanReceive"
	ServerMemberClearCanPublish   ServerMemberClearType = "CanPublish"
	ServerMemberClearJoinedAt     ServerMemberClearType = "JoinedAt"
	ServerMemberClearVoiceChannel ServerMemberClearType = "VoiceChannel"
)

// Clear resets nullable fields to nil based on the JSON key name.
func (m *ServerMember) clear(fields []ServerMemberClearType) {
	for _, field := range fields {
		switch field {
		case ServerMemberClearNickname:
			m.Nickname = nil
		case ServerMemberClearPronouns:
			m.Pronouns = nil
		case ServerMemberClearAvatar:
			m.Avatar = nil
		case ServerMemberClearRoles:
			m.Roles = nil
		case ServerMemberClearTimeout:
			m.Timeout = nil
		case ServerMemberClearCanReceive:
			// Not nullable server-side: clearing resets to true, un-deafened.
			allowed := true
			m.CanReceive = &allowed
		case ServerMemberClearCanPublish:
			allowed := true
			m.CanPublish = &allowed
		case ServerMemberClearJoinedAt, ServerMemberClearVoiceChannel:
			// Neither names a nullable field, and the backend's own
			// remove_field no-ops both. joined_at is required on a live member;
			// the variant exists to $unset it on the tombstone a timed-out
			// member leaves. VoiceChannel is not a member field at all: in an
			// edit it disconnects the member from voice, and the event echoes
			// the request's remove list back verbatim.
		default:
			log.Printf("ServerMember.clear(): unhandled field %s\n", field)
		}
	}
}

type PartialServerMember struct {
	Nickname   *string    `msg:"nickname" json:"nickname,omitzero"`
	Pronouns   *string    `msg:"pronouns" json:"pronouns,omitzero"`
	Avatar     *File      `msg:"avatar" json:"avatar,omitzero"`
	Roles      *[]string  `msg:"roles" json:"roles,omitzero"`
	Timeout    *time.Time `msg:"timeout" json:"timeout,omitzero"`
	CanPublish *bool      `msg:"can_publish" json:"can_publish,omitzero"`
	CanReceive *bool      `msg:"can_receive" json:"can_receive,omitzero"`
}

// Mention is a proxy function that calls ServerMember.ID.Mention().
func (m *ServerMember) Mention() string {
	return m.ID.Mention()
}

type MemberCompositeID struct {
	User   string `msg:"user" json:"user,omitzero"`
	Server string `msg:"server" json:"server,omitzero"`
}

func (m MemberCompositeID) Mention() string {
	return fmt.Sprintf("<@%s>", m.User)
}

type ServerCreateResponse struct {
	Server   *Server    `msg:"server" json:"server,omitzero"`
	Channels []*Channel `msg:"channels" json:"channels,omitzero"`
}

type ServerMembers struct {
	Members []*ServerMember `msg:"members" json:"members,omitzero"`
	Users   []*User         `msg:"users" json:"users,omitzero"`
}

type ServerBans struct {
	Users []*User      `msg:"users" json:"users,omitzero"`
	Bans  []*ServerBan `msg:"bans" json:"bans,omitzero"`
}

type ServerBan struct {
	ID     MemberCompositeID `msg:"_id" json:"_id,omitzero"`
	Reason string            `msg:"reason" json:"reason,omitzero"`
}
