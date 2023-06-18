package revoltgo

type State struct {
	Users    map[string]*User
	Servers  map[string]*Server
	Channels map[string]*Channel
	Members  map[string]*ServerMember
	Emojis   map[string]*Emoji
}
