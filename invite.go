package revoltgo

//go:generate msgp -tests=false -io=false

type InviteType string

const (
	InviteTypeServer InviteType = "Server"
	InviteTypeGroup  InviteType = "Group"
)

type Invite struct {
	Type               InviteType  `msg:"type"`
	ServerID           string      `msg:"server_id"`
	ServerName         string      `msg:"server_name"`
	ServerIcon         *Attachment `msg:"server_icon"`
	ServerBanner       *Attachment `msg:"server_banner"`
	ServerFlags        uint32      `msg:"server_flags"`
	ChannelID          string      `msg:"channel_id"`
	ChannelName        string      `msg:"channel_name"`
	ChannelDescription string      `msg:"channel_description"`
	UserName           string      `msg:"user_name"`
	UserAvatar         *Attachment `msg:"user_avatar"`
	MemberCount        uint64      `msg:"member_count"`
}

type InviteJoin struct {
	Type     InviteType `msg:"type"`
	Channels []*Channel
	Server   *Server `msg:"server"`
}

// InviteCreate seems deprecated/no longer documented
// todo: remove in the future
type InviteCreate struct {
	Type InviteType `msg:"type"`

	// ID is the code of the invite
	ID      string `msg:"_id"`
	Server  string `msg:"server"`
	Creator string `msg:"creator"`
	Channel string `msg:"channel"`
}
