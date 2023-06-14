package revoltgo

import (
	"fmt"
)

/*
	This file contains all URLs and endpoints for the Revolt API
	- "URL" variables are either found on their own, in conjunction with other URLs
    - "Endpoint" functions help construct dynamic URLs by accepting parameters

	Depending on where slashes are, this tells you how the URLs can be used in conjunction
	For instance, a URL:
	- prefixed with a slash "/path" indicates that it's always found at the end of URLs
	- not prefixed or suffixed with slashes is commonly found in-between other URLs
	- suffixed with a slash "path/" indicates that it's always found before other URLs
*/

const (
	URLWebsocket = "wss://ws.revolt.chat/"
	URLAPI       = "https://api.revolt.chat/"

	URLCreate = "create"

	URLChannels = "channels/"
	URLUsers    = "users/"
	URLServers  = "servers/"
	URLAuth     = "auth/"
	URLSession  = "session/"
	URLBots     = "bots/"

	URLBotsCreate         = URLAPI + URLBots + URLCreate
	URLUsersRelationships = URLAPI + URLUsers + "relationships"
	URLChannelsCreate     = URLAPI + URLChannels + URLCreate
	URLAuthSessionLogin   = URLAPI + URLAuth + URLSession + "login"
	URLServersCreate      = URLAPI + URLServers + URLCreate
)

func EndpointChannels(id string) string {
	return fmt.Sprintf("%s%s%s", URLAPI, URLChannels, id)
}

// EndpointUsers constructs a URL to fetch users.
// "@me" will return the current user
// "dms" will return the current user's direct messages
func EndpointUsers(id string) string {
	return fmt.Sprintf("%s%s%s", URLAPI, URLUsers, id)
}

func EndpointServers(id string) string {
	return fmt.Sprintf("%s%s%s", URLAPI, URLServers, id)
}

func EndpointUsersFriend(username string) string {
	return fmt.Sprintf("%s%s%s/friend", URLAPI, URLUsers, username)
}

// EndpointBots constructs a URL to fetch bots based on an ID
// "@me" will return bots that this client owns
func EndpointBots(id string) string {
	return fmt.Sprintf("%s%s%s", URLAPI, URLBots, id)
}
