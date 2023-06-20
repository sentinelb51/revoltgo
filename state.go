package revoltgo

type State struct {

	// The user we are logged in as
	User *User

	Users    map[string]*User
	Servers  map[string]*Server
	Channels map[string]*Channel
	Members  map[string]*ServerMember
	Emojis   map[string]*Emoji
}
