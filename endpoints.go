package revoltgo

import (
	"fmt"
	"log"
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
	return fmt.Sprintf(URLOnboard, action)
}

func EndpointAuthSession(action string) string {
	return fmt.Sprintf(URLAuthSession, action)
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

func EndpointUser(uID string) string {
	return fmt.Sprintf(URLUser, uID)
}

func EndpointUserBlock(uID string) string {
	return fmt.Sprintf(URLUserBlock, uID)
}

func EndpointUserMutual(uID string) string {
	return fmt.Sprintf(URLUserMutual, uID)
}

func EndpointUserDM(uID string) string {
	return fmt.Sprintf(URLUserDM, uID)
}

func EndpointUserDefaultAvatar(uID string) string {
	return fmt.Sprintf(URLUserDefaultAvatar, uID)
}

func EndpointUserFlags(uID string) string {
	return fmt.Sprintf(URLUserFlags, uID)
}

func EndpointUserFriend(uID string) string {

	if uID == "" {
		return fmt.Sprintf(URLUser, "friend")
	}

	return fmt.Sprintf(URLUserFriend, uID)
}

func EndpointUserProfile(uID string) string {
	return fmt.Sprintf(URLUserProfile, uID)
}

func EndpointServer(sID string) string {
	return fmt.Sprintf(URLServer, sID)
}

func EndpointServerAck(sID string) string {
	return fmt.Sprintf(URLServerAck, sID)
}

func EndpointServerChannels(sID string) string {
	return fmt.Sprintf(URLServerChannels, sID)
}

func EndpointChannelPermission(cID, rID string) string {
	return fmt.Sprintf(URLChannelPermission, cID, rID)
}

func EndpointServerMembers(sID string) string {
	return fmt.Sprintf(URLServerMembers, sID)
}

func EndpointServerMember(sID, mID string) string {
	return fmt.Sprintf(URLServerMember, sID, mID)
}

func EndpointServerBans(sID string) string {
	return fmt.Sprintf(URLServerBans, sID)
}

func EndpointServerBan(sID, uID string) string {
	return fmt.Sprintf(URLServerBan, sID, uID)
}

func EndpointInvite(sID string) string {
	return fmt.Sprintf(URLInvites, sID)
}

func EndpointServerInvites(sID string) string {
	return fmt.Sprintf(URLServerInvites, sID)
}

func EndpointServerRoles(sID string) string {
	return fmt.Sprintf(URLServerRoles, sID)
}

func EndpointServerRolesRanks(sID string) string {
	return fmt.Sprintf(URLServerRolesRanks, sID)
}

func EndpointServerEmojis(sID string) string {
	return fmt.Sprintf(URLServerEmojis, sID)
}

func EndpointServerRole(sID, rID string) string {
	return fmt.Sprintf(URLServerRole, sID, rID)
}

func EndpointChannel(cID string) string {
	return fmt.Sprintf(URLChannel, cID)
}

func EndpointChannelMembers(cID string) string {
	return fmt.Sprintf(URLChannelMembers, cID)
}

func EndpointChannelJoinCall(cID string) string {
	return fmt.Sprintf(URLChannelJoinCall, cID)
}

func EndpointChannelEndRing(cID, uID string) string {
	return fmt.Sprintf(URLChannelEndRing, cID, uID)
}

func EndpointChannelAckMessage(cID, mID string) string {
	return fmt.Sprintf(URLChannelAckMessage, cID, mID)
}

func EndpointChannelRecipients(cID, mID string) string {
	return fmt.Sprintf(URLChannelRecipient, cID, mID)
}

func EndpointServerPermissions(sID, rID string) string {
	return fmt.Sprintf(URLServerPermissions, sID, rID)
}

func EndpointChannelMessages(cID string) string {
	return fmt.Sprintf(URLChannelMessages, cID)
}

func EndpointChannelMessageReaction(cID, mID, rID string) string {
	return fmt.Sprintf(URLChannelMessageReaction, cID, mID, rID)
}

func EndpointChannelMessageReactions(cID, mID string) string {
	return fmt.Sprintf(URLChannelMessageReactions, cID, mID)
}

func EndpointChannelMessage(cID, mID string) string {
	return fmt.Sprintf(URLChannelMessage, cID, mID)
}

func EndpointChannelMessagesPin(cID, mID string) string {
	return fmt.Sprintf(URLChannelMessagePin, cID, mID)
}

func EndpointChannelSearch(cID string) string {
	return fmt.Sprintf(URLChannelSearch, cID)
}

func EndpointChannelInvites(cID string) string {
	return fmt.Sprintf(URLChannelInvites, cID)
}

func EndpointChannelWebhooks(cID string) string {
	return fmt.Sprintf(URLChannelWebhooks, cID)
}

func EndpointWebhook(wID string) string {
	return fmt.Sprintf(URLWebhooks, wID)
}

func EndpointWebhookToken(wID, token string) string {
	return fmt.Sprintf(URLWebhookToken, wID, token)
}

func EndpointWebhookGitHub(wID, token string) string {
	return fmt.Sprintf(URLWebhookTokenGitHub, wID, token)
}

/* Bot endpoints */

func EndpointBot(bID string) string {
	return fmt.Sprintf(URLBots, bID)
}

func EndpointBotInvite(bID string) string {
	return fmt.Sprintf(URLBotInvite, bID)
}

/* Custom endpoints */

func EndpointCustomEmoji(eID string) string {
	return fmt.Sprintf(URLCustomEmoji, eID)
}

/* Miscellaneous endpoints */

func EndpointPolicy(action string) string {
	return fmt.Sprintf(URLPolicy, action)
}

func EndpointSync(id string) string {
	return fmt.Sprintf(URLSync, id)
}

// EndpointSyncSettings supports either "set" or "fetch" as action.
func EndpointSyncSettings(action string) string {
	return fmt.Sprintf(URLSync, fmt.Sprintf("settings/%s", action))
}

func EndpointPush(action string) string {
	return fmt.Sprintf(URLPush, action)
}

/* Authentication MFA endpoints */

func EndpointAuthMFA(action string) string {
	return fmt.Sprintf(URLAuthMFA, action)
}

/* CDN endpoints */

func EndpointAutumn(tag string) (url string) {
	url = fmt.Sprintf("%s/%s", cdnURL, tag)
	return
}

func EndpointAutumnFile(tag, id, size string) (url string) {
	url = fmt.Sprintf("%s/%s/%s", cdnURL, tag, id)
	if size != "" {
		url += "?max_side=" + size
	}
	return
}
