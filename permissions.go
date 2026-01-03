package revoltgo

import (
	"fmt"
	"time"
)

//go:generate msgp -tests=false -io=false

// PermissionOverwrite is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/permissions/src/models/server.rs#L52.
type PermissionOverwrite struct {
	Allow int64 `msg:"a" json:"a,omitempty"`
	Deny  int64 `msg:"d" json:"d,omitempty"`
}

const (
	UserPermissionAccess      = 1 << 0
	UserPermissionViewProfile = 1 << 1
	UserPermissionSendMessage = 1 << 2
	UserPermissionInvite      = 1 << 3
)

const (
	PermissionManageChannel       = 1 << 0
	PermissionManageServer        = 1 << 1
	PermissionManagePermissions   = 1 << 2
	PermissionManageRole          = 1 << 3
	PermissionManageCustomisation = 1 << 4
	PermissionKickMembers         = 1 << 6
	PermissionBanMembers          = 1 << 7
	PermissionTimeoutMembers      = 1 << 8
	PermissionAssignRoles         = 1 << 9
	PermissionChangeNickname      = 1 << 10
	PermissionManageNicknames     = 1 << 11
	PermissionChangeAvatar        = 1 << 12
	PermissionRemoveAvatars       = 1 << 13
	PermissionViewChannel         = 1 << 20
	PermissionReadMessageHistory  = 1 << 21
	PermissionSendMessage         = 1 << 22
	PermissionManageMessages      = 1 << 23
	PermissionManageWebhooks      = 1 << 24
	PermissionInviteOthers        = 1 << 25
	PermissionSendEmbeds          = 1 << 26
	PermissionUploadFiles         = 1 << 27
	PermissionMasquerade          = 1 << 28
	PermissionReact               = 1 << 29
	PermissionConnect             = 1 << 30
	PermissionSpeak               = 1 << 31
	PermissionVideo               = 1 << 32
	PermissionMuteMembers         = 1 << 33
	PermissionDeafenMembers       = 1 << 34
	PermissionMoveMembers         = 1 << 35
	PermissionListen              = 1 << 36
	PermissionMentionEveryone     = 1 << 37
	PermissionMentionRoles        = 1 << 38
	PermissionGrantAllSafe        = 0x000F_FFFF_FFFF_FFFF
)

const (
	PermissionPresetTimeout  = PermissionViewChannel + PermissionReadMessageHistory
	PermissionPresetViewOnly = PermissionViewChannel + PermissionReadMessageHistory
	PermissionPresetDefault  = PermissionPresetViewOnly + PermissionSendMessage + PermissionInviteOthers + PermissionSendEmbeds + PermissionUploadFiles + PermissionConnect + PermissionSpeak + PermissionVideo + PermissionListen
	PermissionPresetDM       = PermissionPresetDefault + PermissionReact + PermissionManageChannel
	PermissionPresetServer   = PermissionPresetDefault + PermissionReact + PermissionChangeNickname + PermissionChangeAvatar
)

// ServerPermissions is a utility function to calculate permissions for a user in a Server
func (s *State) ServerPermissions(user *User, server *Server) (int64, error) {
	if server.Owner == user.ID {
		return PermissionGrantAllSafe, nil
	}

	// Get member
	member := s.Member(user.ID, server.ID)
	if member == nil {
		return 0, fmt.Errorf("member %s not found in %s", user.ID, server.ID)
	}

	permissions := server.DefaultPermissions

	// Apply role permissions
	for _, rID := range member.Roles {
		role := server.Roles[rID]
		if role == nil {
			return 0, fmt.Errorf("role %s not found", rID)
		}

		permissions |= role.Permissions.Allow
		permissions &= ^role.Permissions.Deny
	}

	// Apply timeout permissions if necessary
	if !member.Timeout.IsZero() && time.Now().Before(member.Timeout.Time) {
		permissions &= PermissionPresetTimeout
	}

	return permissions, nil
}

// ChannelPermissions is a utility function to calculate permissions for a user in a Channel
func (s *State) ChannelPermissions(user *User, channel *Channel) (int64, error) {
	switch channel.ChannelType {
	case ChannelTypeSavedMessages:
		return PermissionGrantAllSafe, nil
	case ChannelTypeDM:
		if *channel.Permissions&PermissionSendMessage == PermissionSendMessage {
			return PermissionPresetDM, nil
		}
		return PermissionPresetViewOnly, nil
	case ChannelTypeGroup:

		if channel.Owner == user.ID {
			return PermissionGrantAllSafe, nil
		}

		if channel.Permissions != nil {
			return *channel.Permissions, nil
		}

		return PermissionPresetDM, nil
	case ChannelTypeText, ChannelTypeVoice:
		server := s.Server(*channel.Server)
		if server == nil {
			return 0, fmt.Errorf("server %s not found", *channel.Server)
		}

		if server.Owner == user.ID {
			return PermissionGrantAllSafe, nil
		}

		permissions, err := s.ServerPermissions(user, server)
		if err != nil {
			return 0, err
		}

		// Apply default permissions for this channel
		if channel.DefaultPermissions != nil {
			permissions |= channel.DefaultPermissions.Allow
			permissions &= ^channel.DefaultPermissions.Deny
		}

		return permissions, nil
	default:
		return 0, fmt.Errorf("unknown channel type %v", channel.ChannelType)
	}
}
