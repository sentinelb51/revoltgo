package revoltgo

import (
	"fmt"
	"time"
)

// PermissionAD describes the default allowed and denied permissions
type PermissionAD struct {
	Allow uint `json:"a"`
	Deny  uint `json:"d"`
}

const (
	UserPermissionAccess = 1 << iota
	UserPermissionViewProfile
	UserPermissionSendMessage
	UserPermissionInvite
)

const (
	PermissionManageChannel       = 2 << 0
	PermissionManageServer        = 2 << 1
	PermissionManagePermissions   = 2 << 2
	PermissionManageRole          = 2 << 3
	PermissionManageCustomisation = 2 << 4
	PermissionKickMembers         = 2 << 6
	PermissionBanMembers          = 2 << 7
	PermissionTimeoutMembers      = 2 << 8
	PermissionAssignRoles         = 2 << 9
	PermissionChangeNickname      = 2 << 10
	PermissionManageNicknames     = 2 << 11
	PermissionChangeAvatar        = 2 << 12
	PermissionRemoveAvatars       = 2 << 13
	PermissionViewChannel         = 2 << 20
	PermissionReadMessageHistory  = 2 << 21
	PermissionSendMessage         = 2 << 22
	PermissionManageMessages      = 2 << 23
	PermissionManageWebhooks      = 2 << 24
	PermissionInviteOthers        = 2 << 25
	PermissionSendEmbeds          = 2 << 26
	PermissionUploadFiles         = 2 << 27
	PermissionMasquerade          = 2 << 28
	PermissionReact               = 2 << 29
	PermissionConnect             = 2 << 30
	PermissionSpeak               = 2 << 31
	PermissionVideo               = 2 << 32
	PermissionMuteMembers         = 2 << 33
	PermissionDeafenMembers       = 2 << 34
	PermissionMoveMembers         = 2 << 35
	PermissionGrantAllSafe        = 0x000F_FFFF_FFFF_FFFF
)

const (
	PermissionPresetTimeout  = PermissionViewChannel + PermissionReadMessageHistory
	PermissionPresetViewOnly = PermissionViewChannel + PermissionReadMessageHistory
	PermissionPresetDefault  = PermissionPresetViewOnly + PermissionSendMessage + PermissionInviteOthers + PermissionSendEmbeds + PermissionUploadFiles + PermissionConnect + PermissionSpeak
	PermissionPresetDM       = PermissionPresetDefault + PermissionReact + PermissionManageChannel
)

// ServerPermissions is a utility function to calculate permissions for a user in a Server
func (s *State) ServerPermissions(user *User, server *Server) (uint, error) {
	if server.Owner == user.ID {
		return PermissionGrantAllSafe, nil
	}

	// Get member
	key := MemberCompoundID{User: user.ID, Server: server.ID}
	member, exists := s.Members[key.String()]
	if !exists {
		return 0, fmt.Errorf("member %s not found", key.String())
	}

	permissions := *server.DefaultPermissions

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
	if member.Timeout != nil && time.Now().Before(*member.Timeout) {
		permissions &= PermissionPresetTimeout
	}

	return permissions, nil
}

// ChannelPermissions is a utility function to calculate permissions for a user in a Channel
func (s *State) ChannelPermissions(user *User, channel *Channel) (uint, error) {
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
		server := s.Servers[channel.Server]
		if server == nil {
			return 0, fmt.Errorf("server %s not found", channel.Server)
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
