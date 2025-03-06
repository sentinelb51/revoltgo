package revoltgo

import (
	"fmt"
)

/*
	This file contains all the endpoints used in this library for the Revolt API.
	The naming scheme of the constants and methods has rules:

	Constants:
	 - Prefix with "URL"
	 - Follow a hierarchical structure that build on top of each other
	 - Use plural/singular forms to somewhat reflect the relationship between resources;
		- For example, "EndpointChannelAckMessage" relates to a single Channel and a single Message object
		- Conversely, "URLChannelsMessages" relates to multiple messages, but inherits "Channels" due to the hierarchy

	Methods:
	 - Prefix with "Endpoint"
	 - Follow the same hierarchical structure as the constants
*/

const (
	apiURL = "https://api.revolt.chat"
	cdnURL = "https://autumn.revolt.chat/%s"

	URLUsersUsername = apiURL + "/users/me/username"

	URLUsers              = apiURL + "/users/%s"
	URLUsersMutual        = URLUsers + "/mutual"
	URLUsersDM            = URLUsers + "/dm"
	URLUsersFlags         = URLUsers + "/flags"
	URLUsersFriend        = URLUsers + "/friend"
	URLUsersBlock         = URLUsers + "/block"
	URLUsersProfile       = URLUsers + "/profile"
	URLUsersRelationships = URLUsers + "/relationships"

	URLUsersDefaultAvatar = URLUsers + "/default_avatar"

	URLServers         = apiURL + "/servers/%s"
	URLServersAck      = URLServers + "/ack"
	URLServersChannels = URLServers + "/channels"
	URLServersMembers  = URLServers + "/members"
	URLServersMember   = URLServersMembers + "/%s"
	URLServersBans     = URLServers + "/bans"
	URLServersBan      = URLServersBans + "/%s"
	URLServersRoles    = URLServers + "/roles"
	URLServersRole     = URLServers + "/roles/%s"

	URLServersPermissions = URLServers + "/permissions/%s"

	URLChannels                 = apiURL + "/channels/%s"
	URLChannelsAckMessage       = URLChannels + "/ack/%s"
	URLChannelsMessages         = URLChannels + "/messages"
	URLChannelsMessage          = URLChannelsMessages + "/%s"
	URLChannelsMessageReactions = URLChannelsMessage + "/reactions"
	URLChannelsMessageReaction  = URLChannelsMessageReactions + "/%s"
	URLChannelsTyping           = URLChannels + "/typing"
	URLChannelsInvites          = URLChannels + "/invites"
	URLChannelsInvite           = URLChannelsInvites + "/%s"
	URLChannelsPermissions      = URLChannels + "/permissions/%s"
	URLChannelsRecipients       = URLChannels + "/recipients/%s"
	URLChannelsWebhooks         = URLChannels + "/webhooks"

	URLInvites = apiURL + "/invites/%s"

	URLBots       = apiURL + "/bots/%s"
	URLBotsInvite = URLBots + "/invite"

	URLAuth         = apiURL + "/auth"
	URLAuthAccount  = URLAuth + "/account/%s"
	URLAuthSessions = URLAuth + "/session/%s"

	URLCustom      = apiURL + "/custom"
	URLCustomEmoji = URLCustom + "/emoji/%s"

	URLOnboard = apiURL + "/onboard/%s"

	URLSync = apiURL + "/sync/%s"

	URLPush = apiURL + "/push/%s"
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

func EndpointInvite(sID string) string {
	return fmt.Sprintf(URLInvites, sID)
}

func EndpointServersRoles(sID string) string {
	return fmt.Sprintf(URLServersRoles, sID)
}

func EndpointServersRole(sID, rID string) string {
	return fmt.Sprintf(URLServersRole, sID, rID)
}

func EndpointChannels(cID string) string {
	return fmt.Sprintf(URLChannels, cID)
}

func EndpointChannelAckMessage(cID, mID string) string {
	return fmt.Sprintf(URLChannelsAckMessage, cID, mID)
}

func EndpointChannelsRecipients(cID, mID string) string {
	return fmt.Sprintf(URLChannelsRecipients, cID, mID)
}

func EndpointServerPermissions(sID, rID string) string {
	return fmt.Sprintf(URLServersPermissions, sID, rID)
}

func EndpointChannelsMessages(cID string) string {
	return fmt.Sprintf(URLChannelsMessages, cID)
}

func EndpointChannelsMessageReaction(cID, mID, rID string) string {
	return fmt.Sprintf(URLChannelsMessageReaction, cID, mID, rID)
}

func EndpointChannelsMessageReactions(cID, mID string) string {
	return fmt.Sprintf(URLChannelsMessageReactions, cID, mID)
}

func EndpointChannelsMessage(cID, mID string) string {
	return fmt.Sprintf(URLChannelsMessage, cID, mID)
}

func EndpointChannelsTyping(cID string) string {
	return fmt.Sprintf(URLChannelsTyping, cID)
}

func EndpointChannelsInvites(cID string) string {
	return fmt.Sprintf(URLChannelsInvites, cID)
}

func EndpointChannelsInvite(cID, iID string) string {
	return fmt.Sprintf(URLChannelsInvite, cID, iID)
}

func EndpointChannelsWebhooks(cID string) string {
	return fmt.Sprintf(URLChannelsWebhooks, cID)
}

/* Bot endpoints */

func EndpointBots(bID string) string {
	return fmt.Sprintf(URLBots, bID)
}

func EndpointBotsInvite(bID string) string {
	return fmt.Sprintf(URLBotsInvite, bID)
}

/* Custom endpoints */

func EndpointCustomEmoji(eID string) string {
	return fmt.Sprintf(URLCustomEmoji, eID)
}

/* Miscellaneous endpoints */

func EndpointSync(setting string) string {
	return fmt.Sprintf(URLSync, setting)
}

func EndpointSyncSettings(sID string) string {
	return fmt.Sprintf(URLSync, fmt.Sprintf("settings/%s", sID))
}

func EndpointPush(action string) string {
	return fmt.Sprintf(URLPush, action)
}

/* CDN endpoints */

func EndpointAutumn(tag string) (url string) {
	url = fmt.Sprintf(cdnURL, tag)
	return
}

func EndpointAutumnFile(tag, id, size string) (url string) {
	url = fmt.Sprintf("%s/%s", cdnURL, tag, id)
	if size != "" {
		url += "?max_side=" + size
	}
	return
}
