package revoltgo

import (
	"fmt"
)

// List of URLs for the Revolt API
const (
	baseURL          = "https://api.revolt.chat"
	URLUsersUsername = baseURL + "/users/me/username"

	URLUsers              = baseURL + "/users/%s"
	URLUsersMutual        = URLUsers + "/mutual"
	URLUsersDM            = URLUsers + "/dm"
	URLUsersFlags         = URLUsers + "/flags"
	URLUsersFriend        = URLUsers + "/friend"
	URLUsersBlock         = URLUsers + "/block"
	URLUsersProfile       = URLUsers + "/profile"
	URLUsersRelationships = URLUsers + "/relationships"
	URLUsersMutualServers = URLUsers + "/mutual"
	URLUsersAvatar        = URLUsers + "/avatar"
	URLUsersBanner        = URLUsers + "/banner"
	URLUsersDefaultAvatar = URLUsers + "/default_avatar"

	URLServers            = baseURL + "/servers/%s"
	URLServersAck         = URLServers + "/ack"
	URLServersChannels    = URLServers + "/channels"
	URLServersMembers     = URLServers + "/members"
	URLServersMember      = URLServersMembers + "/%s"
	URLServersBans        = URLServers + "/bans"
	URLServersBan         = URLServersBans + "/%s"
	URLServersRoles       = URLServers + "/roles"
	URLServersRole        = URLServers + "/roles/%s"
	URLServersAvatar      = URLServers + "/avatar"
	URLServersBanner      = URLServers + "/banner"
	URLServersPermissions = URLServers + "/permissions/%s"

	URLChannels            = baseURL + "/channels/%s"
	URLChannelsMessages    = URLChannels + "/messages"
	URLChannelsMessage     = URLChannelsMessages + "/%s"
	URLChannelsTyping      = URLChannels + "/typing"
	URLChannelsInvites     = URLChannels + "/invites"
	URLChannelsInvite      = URLChannelsInvites + "/%s"
	URLChannelsPermissions = URLChannels + "/permissions/%s"
	URLChannelsRecipients  = URLChannels + "/recipients/%s"

	URLInvites = baseURL + "/invites/%s"

	URLBots         = baseURL + "/bots/%s"
	URLBotsInvite   = URLBots + "/invite"
	URLBotsCommands = URLBots + "/commands"
	URLBotsCommand  = URLBotsCommands + "/%s"

	URLAuth         = baseURL + "/auth"
	URLAuthAccount  = URLAuth + "/account/%s"
	URLAuthSessions = URLAuth + "/session/%s"

	URLCustom      = baseURL + "/custom"
	URLCustomEmoji = URLCustom + "/emoji/%s"

	URLOnboard = baseURL + "/onboard/%s"
)

func EndpointOnboard(action string) string {
	return fmt.Sprintf(URLOnboard, action)
}

func EndpointAuthSession(action string) string {
	return fmt.Sprintf(URLAuthSessions, action)
}

func EndpointAuthAccountVerify(code string) string {
	return fmt.Sprintf(URLAuthAccount, fmt.Sprintf("verify/%s", code))
}

func EndpointAuthAccount(action string) string {
	return fmt.Sprintf(URLAuthAccount, action)
}

func EndpointAuthAccountChange(detail string) string {
	return fmt.Sprintf(URLAuthAccount, fmt.Sprintf("change/%s", detail))
}

func EndpointEmoji(eID string) string {
	return fmt.Sprintf(URLCustomEmoji, eID)
}

func EndpointUsers(uID string) string {
	return fmt.Sprintf(URLUsers, uID)
}

func EndpointUsersBlock(uID string) string {
	return fmt.Sprintf(URLUsersBlock, uID)
}

func EndpointUsersMutual(uID string) string {
	return fmt.Sprintf(URLUsersMutual, uID)
}

func EndpointUsersDM(uID string) string {
	return fmt.Sprintf(URLUsersDM, uID)
}

func EndpointUsersDefaultAvatar(uID string) string {
	return fmt.Sprintf(URLUsersDefaultAvatar, uID)
}

func EndpointUsersFlags(uID string) string {
	return fmt.Sprintf(URLUsersFlags, uID)
}

func EndpointUsersFriend(uID string) string {
	return fmt.Sprintf(URLUsersFriend, uID)
}

func EndpointUsersProfile(uID string) string {
	return fmt.Sprintf(URLUsersProfile, uID)
}

func EndpointUserMutualServers(uID string) string {
	return fmt.Sprintf(URLUsersMutualServers, uID)
}

func EndpointUserAvatar(uID string) string {
	return fmt.Sprintf(URLUsersAvatar, uID)
}

func EndpointUserBanner(uID string) string {
	return fmt.Sprintf(URLUsersBanner, uID)
}

func EndpointServers(sID string) string {
	return fmt.Sprintf(URLServers, sID)
}

func EndpointServersAck(sID string) string {
	return fmt.Sprintf(URLServersAck, sID)
}

func EndpointServersChannels(sID string) string {
	return fmt.Sprintf(URLServersChannels, sID)
}

func EndpointChannelsPermissions(sID, cID string) string {
	return fmt.Sprintf(URLChannelsPermissions, sID, cID)
}

func EndpointServersMembers(sID string) string {
	return fmt.Sprintf(URLServersMembers, sID)
}

func EndpointServersMember(sID, mID string) string {
	return fmt.Sprintf(URLServersMember, sID, mID)
}

func EndpointServersBans(sID string) string {
	return fmt.Sprintf(URLServersBans, sID)
}

func EndpointServersBan(sID, uID string) string {
	return fmt.Sprintf(URLServersBan, sID, uID)
}

func EndpointInvites(sID string) string {
	return fmt.Sprintf(URLInvites, sID)
}

func EndpointServersRoles(sID string) string {
	return fmt.Sprintf(URLServersRoles, sID)
}

func EndpointServersRole(sID, rID string) string {
	return fmt.Sprintf(URLServersRole, sID, rID)
}

func EndpointServersAvatar(sID string) string {
	return fmt.Sprintf(URLServersAvatar, sID)
}

func EndpointServersBanner(sID string) string {
	return fmt.Sprintf(URLServersBanner, sID)
}

func EndpointChannels(cID string) string {
	return fmt.Sprintf(URLChannels, cID)
}

func EndpointChannelsRecipients(cID, mID string) string {
	return fmt.Sprintf(URLChannelsRecipients, cID, mID)
}

func EndpointPermissions(sID, rID string) string {
	return fmt.Sprintf(URLServersPermissions, sID, rID)
}

func EndpointChannelMessages(cID string) string {
	return fmt.Sprintf(URLChannelsMessages, cID)
}

func EndpointChannelMessagesMessage(cID, mID string) string {
	return fmt.Sprintf(URLChannelsMessage, cID, mID)
}

func EndpointChannelTyping(cID string) string {
	return fmt.Sprintf(URLChannelsTyping, cID)
}

func EndpointChannelInvites(cID string) string {
	return fmt.Sprintf(URLChannelsInvites, cID)
}

func EndpointChannelInvite(cID, iID string) string {
	return fmt.Sprintf(URLChannelsInvite, cID, iID)
}

func EndpointInvite(iID string) string {
	return fmt.Sprintf(URLInvites, iID)
}

/* Bot endpoints */

func EndpointBots(bID string) string {
	return fmt.Sprintf(URLBots, bID)
}

func EndpointBotsInvite(bID string) string {
	return fmt.Sprintf(URLBotsInvite, bID)
}

func EndpointBotsCommands(bID string) string {
	return fmt.Sprintf(URLBotsCommands, bID)
}

func EndpointBotsCommand(bID, cmdID string) string {
	return fmt.Sprintf(URLBotsCommand, bID, cmdID)
}
