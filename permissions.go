package revoltgo

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"time"
)

//go:generate msgp -tests=false -io=false

// PermissionOverwrite is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/permissions/src/models/server.rs#L52.
type PermissionOverwrite struct {
	Allow int64 `msg:"a" json:"a,omitzero"`
	Deny  int64 `msg:"d" json:"d,omitzero"`
}

// ErrNotCached is returned when a permission calculation needs a record the state does not hold.
// It means "unknown", not "denied"; a genuine lack of permissions is 0 with a nil error.
var ErrNotCached = errors.New("not cached")

const (
	UserPermissionAccess      = 1 << 0
	UserPermissionViewProfile = 1 << 1
	UserPermissionSendMessage = 1 << 2
	UserPermissionInvite      = 1 << 3

	// UserPermissionGrantAll stands in for the backend's u64::MAX, which does not fit an int64.
	// Only the four bits above are ever tested, so the two behave the same.
	UserPermissionGrantAll = UserPermissionAccess + UserPermissionViewProfile + UserPermissionSendMessage + UserPermissionInvite
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
	PermissionBypassSlowmode      = 1 << 39
	PermissionViewAuditLogs       = 1 << 40
	PermissionGrantAllSafe        = 0x000F_FFFF_FFFF_FFFF
)

const (
	PermissionPresetTimeout  = PermissionViewChannel + PermissionReadMessageHistory
	PermissionPresetViewOnly = PermissionViewChannel + PermissionReadMessageHistory
	PermissionPresetDefault  = PermissionPresetViewOnly + PermissionSendMessage + PermissionInviteOthers + PermissionSendEmbeds + PermissionUploadFiles + PermissionConnect + PermissionSpeak + PermissionVideo + PermissionListen
	PermissionPresetDM       = PermissionPresetDefault + PermissionReact + PermissionManageChannel + PermissionMasquerade
	PermissionPresetServer   = PermissionPresetDefault + PermissionReact + PermissionChangeNickname + PermissionChangeAvatar
)

// apply grants an overwrite's allows, then revokes its denies.
func (o PermissionOverwrite) apply(permissions int64) int64 {
	return (permissions | o.Allow) &^ o.Deny
}

// inTimeout reports whether the member is currently timed out.
func (m *ServerMember) inTimeout() bool {
	return m.Timeout != nil && time.Now().Before(*m.Timeout)
}

// roleIDs returns the member's roles that exist on the server, ordered by descending rank.
// A lower rank outranks a higher one, so applying overwrites in this order lets the
// member's highest role win last. Unknown role IDs are skipped, as the backend does.
func (m *ServerMember) roleIDs(server *Server) []string {
	ids := make([]string, 0, len(m.Roles))

	for _, rID := range m.Roles {
		if server.Roles[rID] != nil {
			ids = append(ids, rID)
		}
	}

	slices.SortStableFunc(ids, func(a, b string) int {
		return cmp.Compare(server.Roles[b].Rank, server.Roles[a].Rank)
	})

	return ids
}

// revokeVoice strips the voice permissions a server-wide mute or deafen takes away.
// Absent means allowed, matching the backend's default.
func (m *ServerMember) revokeVoice(permissions int64) int64 {
	if m.CanPublish != nil && !*m.CanPublish {
		permissions &^= PermissionSpeak | PermissionVideo
	}

	if m.CanReceive != nil && !*m.CanReceive {
		permissions &^= PermissionListen
	}

	return permissions
}

// relationship resolves user's relationship towards target.
// A bot counts as its own owner, matching the backend.
func (s *State) relationship(user, target *User) UserRelationshipType {
	if user.ID == target.ID {
		return UserRelationshipTypeUser
	}

	if target.Bot != nil && target.Bot.Owner == user.ID {
		return UserRelationshipTypeUser
	}

	for _, relation := range user.Relations {
		if relation.ID == target.ID {
			return relation.Status
		}
	}

	// Relations is only populated on ourselves; every other cached user instead
	// carries Relationship, which is already relative to us.
	if self := s.Self(); self != nil && self.ID == user.ID && target.Relationship != "" {
		return target.Relationship
	}

	return UserRelationshipTypeNone
}

// mutualConnection reports whether two users share a server or a group.
func (s *State) mutualConnection(a, b string) bool {
	mutual := false

	s.membersMu.RLock()
	for sID := range s.members {
		if s.members.get(sID, a) != nil && s.members.get(sID, b) != nil {
			mutual = true
			break
		}
	}
	s.membersMu.RUnlock()

	if mutual {
		return true
	}

	s.channelsMu.RLock()
	defer s.channelsMu.RUnlock()

	for _, channel := range s.channels {
		if channel.ChannelType != ChannelTypeGroup {
			continue
		}

		if slices.Contains(channel.Recipients, a) && slices.Contains(channel.Recipients, b) {
			return true
		}
	}

	return false
}

// UserPermissions calculates user's permissions towards target, in UserPermission* bits.
// Mutual connections are read from the state, so an unpopulated cache under-reports.
// Reference: calculate_user_permissions in core/permissions/src/impl.rs.
func (s *State) UserPermissions(user, target *User) int64 {
	if user.Privileged || user.ID == target.ID {
		return UserPermissionGrantAll
	}

	var permissions int64

	switch s.relationship(user, target) {
	case UserRelationshipTypeFriend:
		return UserPermissionGrantAll
	case UserRelationshipTypeBlocked, UserRelationshipTypeBlockedOther:
		return UserPermissionAccess
	case UserRelationshipTypeIncoming, UserRelationshipTypeOutgoing:
		permissions = UserPermissionAccess
	}

	if !s.mutualConnection(user.ID, target.ID) {
		return permissions
	}

	permissions = UserPermissionAccess | UserPermissionViewProfile

	// Only bots may open a DM off a mutual connection alone
	if target.Bot != nil || user.Bot != nil {
		permissions |= UserPermissionSendMessage
	}

	return permissions
}

// ServerPermissions calculates user's permissions in server.
// A non-member gets 0; the backend does not distinguish that from having no permissions.
// Reference: calculate_server_permissions in core/permissions/src/impl.rs.
func (s *State) ServerPermissions(user *User, server *Server) int64 {
	if user.Privileged || server.Owner == user.ID {
		return PermissionGrantAllSafe
	}

	member := s.Member(server.ID, user.ID)
	if member == nil {
		return 0
	}

	permissions := server.DefaultPermissions

	for _, rID := range member.roleIDs(server) {
		permissions = server.Roles[rID].Permissions.apply(permissions)
	}

	permissions = member.revokeVoice(permissions)

	if member.inTimeout() {
		permissions &= PermissionPresetTimeout
	}

	return permissions
}

// ChannelPermissions calculates user's permissions in channel.
// It returns 0 for a user with no permissions, and wraps ErrNotCached when the
// state is missing a record the calculation needs.
// Reference: calculate_channel_permissions in core/permissions/src/impl.rs.
func (s *State) ChannelPermissions(user *User, channel *Channel) (int64, error) {
	if user.Privileged {
		return PermissionGrantAllSafe, nil
	}

	switch channel.ChannelType {
	case ChannelTypeSavedMessages:
		if channel.User != user.ID {
			return 0, nil
		}

		return PermissionGrantAllSafe, nil
	case ChannelTypeDM:
		if !slices.Contains(channel.Recipients, user.ID) {
			return 0, nil
		}

		// A DM with oneself has no other recipient, so target stays user
		target := user

		for _, rID := range channel.Recipients {
			if rID == user.ID {
				continue
			}

			if target = s.User(rID); target == nil {
				return 0, fmt.Errorf("user %s: %w", rID, ErrNotCached)
			}

			break
		}

		if s.UserPermissions(user, target)&UserPermissionSendMessage == 0 {
			return PermissionPresetViewOnly, nil
		}

		return PermissionPresetDM, nil
	case ChannelTypeGroup:
		if channel.Owner == user.ID {
			return PermissionGrantAllSafe, nil
		}

		if !slices.Contains(channel.Recipients, user.ID) {
			return 0, nil
		}

		// Group permissions are an allow-only overwrite over the view-only preset
		allow := int64(PermissionPresetDM)
		if channel.Permissions != nil {
			allow = *channel.Permissions
		}

		return PermissionPresetViewOnly | allow, nil
	case ChannelTypeText:
		if channel.Server == nil {
			return 0, fmt.Errorf("channel %s has no server", channel.ID)
		}

		server := s.Server(*channel.Server)
		if server == nil {
			return 0, fmt.Errorf("server %s: %w", *channel.Server, ErrNotCached)
		}

		if server.Owner == user.ID {
			return PermissionGrantAllSafe, nil
		}

		member := s.Member(server.ID, user.ID)
		if member == nil {
			return 0, nil
		}

		permissions := server.DefaultPermissions

		if channel.DefaultPermissions != nil {
			permissions = channel.DefaultPermissions.apply(permissions)
		}

		// Server overwrites come first, then the channel's, both lowest rank last
		roleIDs := member.roleIDs(server)

		for _, rID := range roleIDs {
			permissions = server.Roles[rID].Permissions.apply(permissions)
		}

		for _, rID := range roleIDs {
			if overwrite, ok := channel.RolePermissions[rID]; ok {
				permissions = overwrite.apply(permissions)
			}
		}

		permissions = member.revokeVoice(permissions)

		if member.inTimeout() {
			permissions &= PermissionPresetTimeout
		}

		// Losing sight of the channel loses everything with it
		if permissions&PermissionViewChannel == 0 {
			return 0, nil
		}

		return permissions, nil
	default:
		return 0, nil
	}
}
