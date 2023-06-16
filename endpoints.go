package revoltgo

import (
	"fmt"
)

const (
	wsURL   = "wss://ws.revolt.chat/"
	baseURL = "https://api.revolt.chat"
)

// API Endpoints
const (
	URLUsers              = baseURL + "/users/%s"
	URLUsersFriend        = URLUsers + "/friend"
	URLUsersProfile       = URLUsers + "profile"
	URLUsersRelationships = URLUsers + "/relationships"
	URLUsersMutualServers = URLUsers + "/mutual"
	URLUsersAvatar        = URLUsers + "/avatar"
	URLUsersBanner        = URLUsers + "/banner"

	URLServers         = baseURL + "/servers/%s"
	URLServersChannels = baseURL + "/servers/%s/channels"
	URLServersMembers  = baseURL + "/servers/%s/members"
	URLServersMember   = baseURL + "/servers/%s/members/%s"
	URLServersBans     = baseURL + "/servers/%s/bans"
	URLServersBan      = baseURL + "/servers/%s/bans/%s"
	URLServersInvites  = baseURL + "/servers/%s/invites"
	URLServersRoles    = baseURL + "/servers/%s/roles"
	URLServersRole     = baseURL + "/servers/%s/roles/%s"
	URLServersAvatar   = baseURL + "/servers/%s/avatar"
	URLServersBanner   = baseURL + "/servers/%s/banner"

	URLChannels         = baseURL + "/channels/%s"
	URLChannelsMessages = baseURL + "/channels/%s/messages"
	URLChannelsMessage  = baseURL + "/channels/%s/messages/%s"
	URLChannelsTyping   = baseURL + "/channels/%s/typing"
	URLChannelsInvites  = baseURL + "/channels/%s/invites"
	URLChannelsInvite   = baseURL + "/channels/%s/invites/%s"

	URLInvite = baseURL + "/invites/%s"

	URLBots         = baseURL + "/bots/%s"
	URLBotsCommands = baseURL + "/bots/%s/commands"
	URLBotsCommand  = baseURL + "/bots/%s/commands/%s"

	URLAuth              = baseURL + "/auth/"
	URLAuthSessions      = URLAuth + "sessions"
	URLAuthSessionsLogin = URLAuthSessions + "/login"
)

func EndpointUsers(userID string) string {
	return fmt.Sprintf(URLUsers, userID)
}

func EndpointUsersFriend(userID string) string {
	return fmt.Sprintf(URLUsersFriend, userID)
}

func EndpointUserProfile(userID string) string {
	return fmt.Sprintf(URLUsersProfile, userID)
}

func EndpointUserMutualServers(userID string) string {
	return fmt.Sprintf(URLUsersMutualServers, userID)
}

func EndpointUserAvatar(userID string) string {
	return fmt.Sprintf(URLUsersAvatar, userID)
}

func EndpointUserBanner(userID string) string {
	return fmt.Sprintf(URLUsersBanner, userID)
}

func EndpointServers(serverID string) string {
	return fmt.Sprintf(URLServers, serverID)
}

func EndpointServersChannels(serverID string) string {
	return fmt.Sprintf(URLServersChannels, serverID)
}

func EndpointServersMembers(serverID string) string {
	return fmt.Sprintf(URLServersMembers, serverID)
}

func EndpointServersMember(serverID, userID string) string {
	return fmt.Sprintf(URLServersMember, serverID, userID)
}

func EndpointServersBans(serverID string) string {
	return fmt.Sprintf(URLServersBans, serverID)
}

func EndpointServersBan(serverID, userID string) string {
	return fmt.Sprintf(URLServersBan, serverID, userID)
}

func EndpointServersInvites(serverID string) string {
	return fmt.Sprintf(URLServersInvites, serverID)
}

func EndpointServersRoles(serverID string) string {
	return fmt.Sprintf(URLServersRoles, serverID)
}

func EndpointServersRole(serverID, roleID string) string {
	return fmt.Sprintf(URLServersRole, serverID, roleID)
}

func EndpointServersAvatar(serverID string) string {
	return fmt.Sprintf(URLServersAvatar, serverID)
}

func EndpointServersBanner(serverID string) string {
	return fmt.Sprintf(URLServersBanner, serverID)
}

func EndpointChannels(channelID string) string {
	return fmt.Sprintf(URLChannels, channelID)
}

func EndpointChannelMessages(channelID string) string {
	return fmt.Sprintf(URLChannelsMessages, channelID)
}

func EndpointChannelMessagesMessage(channelID, messageID string) string {
	return fmt.Sprintf(URLChannelsMessage, channelID, messageID)
}

func EndpointChannelTyping(channelID string) string {
	return fmt.Sprintf(URLChannelsTyping, channelID)
}

func EndpointChannelInvites(channelID string) string {
	return fmt.Sprintf(URLChannelsInvites, channelID)
}

func EndpointChannelInvite(channelID, inviteID string) string {
	return fmt.Sprintf(URLChannelsInvite, channelID, inviteID)
}

func EndpointInvite(inviteID string) string {
	return fmt.Sprintf(URLInvite, inviteID)
}

/* Bot endpoints */

func EndpointBots(botID string) string {
	return fmt.Sprintf(URLBots, botID)
}

func EndpointBotsCommands(botID string) string {
	return fmt.Sprintf(URLBotsCommands, botID)
}

func EndpointBotsCommand(botID, commandID string) string {
	return fmt.Sprintf(URLBotsCommand, botID, commandID)
}
