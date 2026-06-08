package revoltgo

import (
	"log"
	"strconv"
)

/*
	This file contains all the endpoints used in this library for the Revolt API.
	The naming scheme of the constants and methods has rules:

 	Use plural/singular forms to somewhat reflect the relationship between resources:
	 - For example, "EndpointChannelAckMessage" relates to a single Channel and a single Message object
	 - Conversely, "URLChannelsMessages" relates to multiple messages, but inherits "Channels" due to the hierarchy

	Constants:
	 - Prefix with "URL"
	 - Follow a hierarchical structure that build on top of each other

	Methods:
	 - Prefix with "Endpoint"
	 - Used to generate a URL for a specific resource
	 - Follow the same hierarchical structure as the constants
*/

/* These base URLs are used by the Session.Request method */
var (
	apiURL = "https://api.stoat.chat"
	cdnURL = "https://cdn.stoatusercontent.com"

	parsedAPIBase = mustParseURL(apiURL)
	parsedCDNBase = mustParseURL(cdnURL)
)

func BaseURL() string {
	return apiURL
}

func CDNURL() string {
	return cdnURL
}

// SetBaseURL sets the base URL for the API.
// Call before opening any sessions; this is not mutex-protected.
func SetBaseURL(newURL string) error {
	u, err := validateBaseURL(newURL)
	if err != nil {
		return err
	}

	apiURL = u.String()
	parsedAPIBase = u

	log.Printf("Base URL set to %s", apiURL)
	return nil
}

// SetCDNURL sets the base URL for the CDN (Autumn).
// Call before opening any sessions; this is not mutex-protected.
func SetCDNURL(newURL string) error {
	u, err := validateBaseURL(newURL)
	if err != nil {
		return err
	}

	cdnURL = u.String()
	parsedCDNBase = u

	log.Printf("CDN URL set to %s", cdnURL)
	return nil
}

const (
	URLUser              = "/users/%s"
	URLUserMeUsername    = "/users/@me/username"
	URLUserMutual        = URLUser + "/mutual"
	URLUserDM            = URLUser + "/dm"
	URLUserFlags         = URLUser + "/flags"
	URLUserFriend        = URLUser + "/friend"
	URLUserBlock         = URLUser + "/block"
	URLUserProfile       = URLUser + "/profile"
	URLUserRelationships = URLUser + "/relationships"
	URLUserDefaultAvatar = URLUser + "/default_avatar"

	URLServer            = "/servers/%s"
	URLServerAck         = URLServer + "/ack"
	URLServerChannels    = URLServer + "/channels"
	URLServerMembers     = URLServer + "/members"
	URLServerRoles       = URLServer + "/roles"
	URLServerRole        = URLServer + "/roles/%s"
	URLServerPermissions = URLServer + "/permissions/%s"
	URLServerInvites     = URLServer + "/invites"
	URLServerEmojis      = URLServer + "/emojis"
	URLServerRolesRanks  = URLServer + "/roles/ranks"
	URLServerBans        = URLServer + "/bans"
	URLServerBan         = URLServerBans + "/%s"
	URLServerMember      = URLServerMembers + "/%s"

	URLChannel                 = "/channels/%s"
	URLChannelAckMessage       = URLChannel + "/ack/%s"
	URLChannelJoinCall         = URLChannel + "/join_call"
	URLChannelEndRing          = URLChannel + "/end_ring/%s"
	URLChannelInvites          = URLChannel + "/invites"
	URLChannelPermission       = URLChannel + "/permissions/%s"
	URLChannelRecipient        = URLChannel + "/recipients/%s"
	URLChannelSearch           = URLChannel + "/search"
	URLChannelWebhooks         = URLChannel + "/webhooks"
	URLChannelMessages         = URLChannel + "/messages"
	URLChannelMessage          = URLChannelMessages + "/%s"
	URLChannelMembers          = URLChannel + "/members"
	URLChannelMessageReactions = URLChannelMessage + "/reactions"
	URLChannelMessageReaction  = URLChannelMessageReactions + "/%s"
	URLChannelMessagePin       = URLChannelMessages + "/%s/pin"

	URLWebhooks           = "/webhooks/%s"
	URLWebhookToken       = URLWebhooks + "/%s"
	URLWebhookTokenGitHub = URLWebhookToken + "/github"

	URLInvites = "/invites/%s"

	URLBots      = "/bots/%s"
	URLBotInvite = URLBots + "/invite"

	URLAuth        = "/auth"
	URLAuthMFA     = URLAuth + "/mfa/%s"
	URLAuthAccount = URLAuth + "/account/%s"
	URLAuthSession = URLAuth + "/session/%s"

	URLCustom      = "/custom"
	URLCustomEmoji = URLCustom + "/emoji/%s"

	URLOnboard = "/onboard/%s"

	URLSync = "/sync/%s"

	URLPush = "/push/%s"

	URLSafetyReport = "/safety/report"

	URLPolicy = "/policy/%s"
)

func EndpointOnboard(action string) string {
	return "/onboard/" + action
}

func EndpointAuthSession(action string) string {
	return "/auth/session/" + action
}

func EndpointAuthAccount(action string) string {
	return "/auth/account/" + action
}

func EndpointAuthAccountVerify(code string) string {
	return EndpointAuthAccount("verify/" + code)
}

func EndpointAuthAccountChange(detail string) string {
	return EndpointAuthAccount("change/" + detail)
}

func EndpointUser(uID string) string {
	return "/users/" + uID
}

func EndpointUserBlock(uID string) string {
	return EndpointUser(uID) + "/block"
}

func EndpointUserMutual(uID string) string {
	return EndpointUser(uID) + "/mutual"
}

func EndpointUserDM(uID string) string {
	return EndpointUser(uID) + "/dm"
}

func EndpointUserDefaultAvatar(uID string) string {
	return EndpointUser(uID) + "/default_avatar"
}

func EndpointUserFlags(uID string) string {
	return EndpointUser(uID) + "/flags"
}

// todo: check FriendAdd method

func EndpointUserFriend(uID string) string {

	if uID == "" {
		return EndpointUser("friend")
	}

	return EndpointUser(uID) + "/friend"
}

func EndpointUserProfile(uID string) string {
	return EndpointUser(uID) + "/profile"
}

func EndpointServer(sID string) string {
	return "/servers/" + sID
}

func EndpointServerAck(sID string) string {
	return EndpointServer(sID) + "/ack"
}

func EndpointServerChannels(sID string) string {
	return EndpointServer(sID) + "/channels"
}

func EndpointChannelPermission(cID, rID string) string {
	return EndpointChannel(cID) + "/permissions/" + rID
}

func EndpointServerMembers(sID string, excludeOffline bool) string {
	return EndpointServer(sID) + "/members?exclude_offline=" + strconv.FormatBool(excludeOffline)
}

func EndpointServerMember(sID, mID string) string {
	return EndpointServer(sID) + "/members/" + mID
}

func EndpointServerBans(sID string) string {
	return EndpointServer(sID) + "/bans"
}

func EndpointServerBan(sID, uID string) string {
	return EndpointServerBans(sID) + "/" + uID
}

func EndpointInvite(sID string) string {
	return "/invites/" + sID
}

func EndpointServerInvites(sID string) string {
	return EndpointServer(sID) + "/invites"
}

func EndpointServerRoles(sID string) string {
	return EndpointServer(sID) + "/roles"
}

func EndpointServerRolesRanks(sID string) string {
	return EndpointServer(sID) + "/roles/ranks"
}

func EndpointServerEmojis(sID string) string {
	return EndpointServer(sID) + "/emojis"
}

func EndpointServerRole(sID, rID string) string {
	return EndpointServer(sID) + "/roles/" + rID
}

func EndpointChannel(cID string) string {
	return "/channels/" + cID
}

func EndpointChannelMembers(cID string) string {
	return EndpointChannel(cID) + "/members"
}

func EndpointChannelJoinCall(cID string) string {
	return EndpointChannel(cID) + "/join_call"
}

func EndpointChannelEndRing(cID, uID string) string {
	return EndpointChannel(cID) + "/end_ring/" + uID
}

func EndpointChannelAckMessage(cID, mID string) string {
	return EndpointChannel(cID) + "/ack/" + mID
}

func EndpointChannelRecipients(cID, uID string) string {
	return EndpointChannel(cID) + "/recipients/" + uID
}

func EndpointServerPermissions(sID, rID string) string {
	return EndpointServer(sID) + "/permissions/" + rID
}

func EndpointChannelMessages(cID string) string {
	return EndpointChannel(cID) + "/messages"
}

func EndpointChannelMessage(cID, mID string) string {
	return EndpointChannelMessages(cID) + "/" + mID
}

func EndpointChannelMessageReactions(cID, mID string) string {
	return EndpointChannelMessage(cID, mID) + "/reactions"
}

func EndpointChannelMessageReaction(cID, mID, rID string) string {
	return EndpointChannelMessageReactions(cID, mID) + "/" + rID
}

func EndpointChannelMessagePin(cID, mID string) string {
	return EndpointChannelMessages(cID) + "/" + mID + "/pin"
}

func EndpointChannelSearch(cID string) string {
	return EndpointChannel(cID) + "/search"
}

func EndpointChannelInvites(cID string) string {
	return EndpointChannel(cID) + "/invites"
}

func EndpointChannelWebhooks(cID string) string {
	return EndpointChannel(cID) + "/webhooks"
}

func EndpointWebhook(wID string) string {
	return "/webhooks/" + wID
}

func EndpointWebhookToken(wID, token string) string {
	return EndpointWebhook(wID) + "/" + token
}

func EndpointWebhookGitHub(wID, token string) string {
	return EndpointWebhookToken(wID, token) + "/github"
}

/* Bot endpoints */

func EndpointBot(bID string) string {
	return "/bots/" + bID
}

func EndpointBotInvite(bID string) string {
	return EndpointBot(bID) + "/invite"
}

/* Custom endpoints */

func EndpointCustomEmoji(eID string) string {
	return "/custom/emoji/" + eID
}

/* Miscellaneous endpoints */

func EndpointPolicy(action string) string {
	return "/policy/" + action
}

func EndpointSync(id string) string {
	return "/sync/" + id
}

// EndpointSyncSettings supports either "set" or "fetch" as action.
func EndpointSyncSettings(action string) string {
	return EndpointSync("settings/" + action)
}

func EndpointPush(action string) string {
	return "/push/" + action
}

/* Authentication MFA endpoints */

func EndpointAuthMFA(action string) string {
	return "/auth/mfa/" + action
}

/* CDN endpoints */

func EndpointAutumn(tag string) string {
	return cdnURL + "/" + tag
}

func EndpointAutumnFile(tag, id, size string) (url string) {
	url = cdnURL + "/" + tag + "/" + id
	if size != "" {
		url += "?max_side=" + size
	}
	return
}
