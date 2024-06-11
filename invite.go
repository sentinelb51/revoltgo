package revoltgo

type InviteType string

const (
	InviteTypeServer InviteType = "Server"
	InviteTypeGroup  InviteType = "Group"
)

type Invite struct {
	Type               InviteType  `json:"type"`
	ServerID           string      `json:"server_id"`
	ServerName         string      `json:"server_name"`
	ServerIcon         *Attachment `json:"server_icon"`
	ServerBanner       *Attachment `json:"server_banner"`
	ServerFlags        uint32      `json:"server_flags"`
	ChannelID          string      `json:"channel_id"`
	ChannelName        string      `json:"channel_name"`
	ChannelDescription string      `json:"channel_description"`
	UserName           string      `json:"user_name"`
	UserAvatar         *Attachment `json:"user_avatar"`
	MemberCount        uint64      `json:"member_count"`
}

type InviteJoin struct {
	Type     InviteType `json:"type"`
	Channels []*Channel
	Server   *Server `json:"server"`
}

// InviteCreate seems deprecated/no longer documented
// todo: remove in the future
type InviteCreate struct {
	Type InviteType `json:"type"`

	// ID is the code of the invite
	ID      string `json:"_id"`
	Server  string `json:"server"`
	Creator string `json:"creator"`
	Channel string `json:"channel"`
}
