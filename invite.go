package revoltgo

//go:generate msgp -tests=false -io=false

type InviteType string

const (
	InviteTypeServer InviteType = "Server"
	InviteTypeGroup  InviteType = "Group"
)

type Invite struct {
	Type               InviteType `msg:"type" json:"type,omitzero"`
	ServerID           string     `msg:"server_id" json:"server_id,omitzero"`
	ServerName         string     `msg:"server_name" json:"server_name,omitzero"`
	ServerIcon         *File      `msg:"server_icon" json:"server_icon,omitzero"`
	ServerBanner       *File      `msg:"server_banner" json:"server_banner,omitzero"`
	ServerFlags        uint32     `msg:"server_flags" json:"server_flags,omitzero"`
	ChannelID          string     `msg:"channel_id" json:"channel_id,omitzero"`
	ChannelName        string     `msg:"channel_name" json:"channel_name,omitzero"`
	ChannelDescription string     `msg:"channel_description" json:"channel_description,omitzero"`
	UserName           string     `msg:"user_name" json:"user_name,omitzero"`
	UserAvatar         *File      `msg:"user_avatar" json:"user_avatar,omitzero"`
	MemberCount        uint64     `msg:"member_count" json:"member_count,omitzero"`
}

type InviteJoin struct {
	Type InviteType `msg:"type" json:"type,omitzero"`

	Channels []*Channel `msg:"channels" json:"channels,omitzero"`
	Server   *Server    `msg:"server" json:"server,omitzero"`

	Channel *Channel `msg:"channel" json:"channel,omitzero"`
	Users   []*User  `msg:"users" json:"users,omitzero"`
}

// InviteCreate seems deprecated/no longer documented
// todo: remove in the future
type InviteCreate struct {
	Type InviteType `msg:"type" json:"type,omitzero"`

	// ID is the code of the invite
	ID      string `msg:"_id" json:"_id,omitzero"`
	Server  string `msg:"server" json:"server,omitzero"`
	Creator string `msg:"creator" json:"creator,omitzero"`
	Channel string `msg:"channel" json:"channel,omitzero"`
}
