package revoltgo

//go:generate msgp -tests=false -io=false

type InviteType string

const (
	InviteTypeServer InviteType = "Server"
	InviteTypeGroup  InviteType = "Group"
)

type Invite struct {
	Type               InviteType  `msg:"type" json:"type,omitempty"`
	ServerID           string      `msg:"server_id" json:"server_id,omitempty"`
	ServerName         string      `msg:"server_name" json:"server_name,omitempty"`
	ServerIcon         *Attachment `msg:"server_icon" json:"server_icon,omitempty"`
	ServerBanner       *Attachment `msg:"server_banner" json:"server_banner,omitempty"`
	ServerFlags        uint32      `msg:"server_flags" json:"server_flags,omitempty"`
	ChannelID          string      `msg:"channel_id" json:"channel_id,omitempty"`
	ChannelName        string      `msg:"channel_name" json:"channel_name,omitempty"`
	ChannelDescription string      `msg:"channel_description" json:"channel_description,omitempty"`
	UserName           string      `msg:"user_name" json:"user_name,omitempty"`
	UserAvatar         *Attachment `msg:"user_avatar" json:"user_avatar,omitempty"`
	MemberCount        uint64      `msg:"member_count" json:"member_count,omitempty"`
}

type InviteJoin struct {
	Type     InviteType `msg:"type" json:"type,omitempty"`
	Channels []*Channel
	Server   *Server `msg:"server" json:"server,omitempty"`
}

// InviteCreate seems deprecated/no longer documented
// todo: remove in the future
type InviteCreate struct {
	Type InviteType `msg:"type" json:"type,omitempty"`

	// ID is the code of the invite
	ID      string `msg:"_id" json:"_id,omitempty"`
	Server  string `msg:"server" json:"server,omitempty"`
	Creator string `msg:"creator" json:"creator,omitempty"`
	Channel string `msg:"channel" json:"channel,omitempty"`
}
