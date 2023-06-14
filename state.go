package revoltgo

type State struct {
	Users    map[string]*User
	Servers  map[string]*Server
	Channels map[string]*ServerChannel
	Members  map[string]*Member
	Emojis   map[string]*Emoji
}
