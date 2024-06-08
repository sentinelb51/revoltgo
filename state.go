package revoltgo

import (
	"fmt"
	"log"
	"sync"
)

type State struct {
	sync.RWMutex

	// The current user, also present in users
	Self *User

	/* Caches */
	users    map[string]*User
	servers  map[string]*Server
	channels map[string]*Channel
	members  map[string]*ServerMember
	emojis   map[string]*Emoji
	webhooks map[string]*Webhook

	/* Tracking options */
	TrackUsers    bool
	TrackServers  bool
	TrackChannels bool
	TrackMembers  bool
	TrackEmojis   bool
	TrackWebhooks bool

	// TrackAPICalls additionally updates the state from API calls
	// This concept may future-proof against any de-syncs, but may use more CPU time
	TrackAPICalls bool

	// TrackBulkAPICalls will update the state from bulk API calls
	// This option activates internal State.addServerMembersAndUsers and State.addWebhooks functions
	TrackBulkAPICalls bool
}

/*
	API call updates
	Used when TrackAPICalls/TrackBulkAPICalls is enabled
*/

func (s *State) addUser(user *User) {

	if !s.TrackAPICalls || user == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.users[user.ID] = user
}

func (s *State) addServer(server *Server) {

	if !s.TrackAPICalls || server == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.servers[server.ID] = server
}

func (s *State) addChannel(channel *Channel) {

	if !s.TrackAPICalls || channel == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.channels[channel.ID] = channel
}

func (s *State) addServerMember(member *ServerMember) {

	if !s.TrackAPICalls || member == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.members[member.ID.String()] = member
}

func (s *State) addServerMembersAndUsers(data *ServerMembers) {

	if !s.TrackBulkAPICalls || data == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	for _, member := range data.Members {
		s.members[member.ID.String()] = member
	}

	for _, user := range data.Users {
		s.users[user.ID] = user
	}
}

func (s *State) addEmoji(emoji *Emoji) {

	if !s.TrackAPICalls || emoji == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.emojis[emoji.ID] = emoji
}

func (s *State) addWebhooks(webhook []*Webhook) {

	if !s.TrackBulkAPICalls || webhook == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	for _, w := range webhook {
		s.webhooks[w.ID] = w
	}
}

/*
	Getter functions
*/

func (s *State) User(id string) *User {
	s.RLock()
	defer s.RUnlock()

	return s.users[id]
}

func (s *State) Server(id string) *Server {
	s.RLock()
	defer s.RUnlock()

	return s.servers[id]
}

func (s *State) Role(sID, rID string) *ServerRole {
	server := s.servers[sID]
	if server == nil {
		return nil
	}

	return server.Roles[rID]
}

func (s *State) Channel(id string) *Channel {
	s.RLock()
	defer s.RUnlock()

	return s.channels[id]
}

func (s *State) Member(uID, sID string) *ServerMember {
	s.RLock()
	defer s.RUnlock()

	return s.members[MemberCompositeID{User: uID, Server: sID}.String()]
}

func (s *State) Emoji(id string) *Emoji {
	s.RLock()
	defer s.RUnlock()

	return s.emojis[id]
}

func (s *State) Webhook(id string) *Webhook {
	s.RLock()
	defer s.RUnlock()

	return s.webhooks[id]
}

func newState() *State {
	return &State{
		TrackUsers:        true,
		TrackServers:      true,
		TrackChannels:     true,
		TrackMembers:      true,
		TrackEmojis:       true,
		TrackAPICalls:     true,
		TrackBulkAPICalls: true,
		TrackWebhooks:     false,
	}
}

/*
	Websocket state updates
*/

// populate populates the state with the data from the ready event.
// It will overwrite any existing data in the state.
func (s *State) populate(ready *EventReady) {

	s.Lock()
	defer s.Unlock()

	// The last user in the ready event is the current user
	s.Self = ready.Users[len(ready.Users)-1]

	/* Populate the caches */

	if s.TrackUsers {
		s.users = make(map[string]*User, len(ready.Users))
		for _, user := range ready.Users {
			s.users[user.ID] = user
		}
	}

	if s.TrackServers {
		s.servers = make(map[string]*Server, len(ready.Servers))
		for _, server := range ready.Servers {
			s.servers[server.ID] = server
		}
	}

	if s.TrackChannels {
		s.channels = make(map[string]*Channel, len(ready.Channels))
		for _, channel := range ready.Channels {
			s.channels[channel.ID] = channel
		}
	}

	if s.TrackMembers {
		s.members = make(map[string]*ServerMember, len(ready.Members))
		for _, member := range ready.Members {
			s.members[member.ID.String()] = member
		}
	}

	if s.TrackEmojis {
		s.emojis = make(map[string]*Emoji, len(ready.Emojis))
		for _, emoji := range ready.Emojis {
			s.emojis[emoji.ID] = emoji
		}
	}
}

func (s *State) platformWipe(event *EventUserPlatformWipe) {
	if !s.TrackUsers {
		return
	}

	s.Lock()
	defer s.Unlock()

	delete(s.users, event.UserID)
}

func (s *State) updateServerRole(event *AbstractEventUpdate) {

	if !s.TrackServers {
		return
	}

	server := s.servers[event.ID.StringID]
	if server == nil {
		log.Printf("update for role %s in unknown server %s\n", event.RoleID, event.ID)
		return
	}

	role := server.Roles[event.RoleID]
	if role == nil {
		log.Printf("update for unknown role %s in server %s\n", event.RoleID, event.ID)
		return
	}

	s.Lock()
	defer s.Unlock()

	mergeJSON[ServerRole](role, event.Data, event.Clear)
}

func (s *State) deleteServerRole(data *EventServerRoleDelete) {

	if !s.TrackServers {
		return
	}

	s.Lock()
	defer s.Unlock()

	server := s.servers[data.ID]
	if server != nil {
		delete(server.Roles, data.RoleID)
	}
}

func (s *State) createServerMember(data *EventServerMemberJoin) {

	if !s.TrackMembers {
		return
	}

	s.Lock()
	defer s.Unlock()

	id := MemberCompositeID{User: data.User, Server: data.ID}
	s.members[id.String()] = &ServerMember{ID: id}
}

func (s *State) deleteServerMember(data *EventServerMemberLeave) {

	// We left the server, remove it from the state
	if data.User == s.Self.ID {
		delete(s.servers, data.ID)
	}

	if !s.TrackMembers {
		return
	}

	s.Lock()
	defer s.Unlock()

	id := MemberCompositeID{User: data.User, Server: data.ID}
	delete(s.members, id.String())
}

func (s *State) updateServerMember(event *AbstractEventUpdate) {

	if !s.TrackMembers {
		return
	}

	mID := event.ID.MemberID
	member := s.members[mID.String()]
	if member == nil {
		log.Printf("unknown member update %s in server %s\n", mID.User, mID.Server)
		return
	}

	s.Lock()
	defer s.Unlock()

	mergeJSON[ServerMember](member, event.Data, event.Clear)
}

func (s *State) createChannel(event *EventChannelCreate) {

	if !s.TrackChannels {
		return
	}

	server := s.servers[event.Server]
	if server == nil {
		fmt.Printf("channel %s created in unknown server %s\n", event.ID, event.Server)
		return
	}

	s.Lock()
	defer s.Unlock()

	s.channels[event.ID] = &Channel{
		ID:     event.ID,
		Server: event.Server,
		Type:   event.Type,
		Name:   event.Name,
	}

	server.Channels = append(server.Channels, event.ID)
}

func (s *State) updateChannel(event *AbstractEventUpdate) {

	if !s.TrackChannels {
		return
	}

	channel := s.channels[event.ID.StringID]
	if channel == nil {
		fmt.Printf("unknown channel updated %s\n", event.ID)
		return
	}

	s.Lock()
	defer s.Unlock()

	mergeJSON[Channel](channel, event.Data, event.Clear)
}

func (s *State) deleteChannel(event *EventChannelDelete) {

	if !s.TrackChannels {
		return
	}

	s.Lock()
	defer s.Unlock()

	channel := s.channels[event.ID]
	if channel != nil {
		delete(s.channels, event.ID)
	}

	server, exists := s.servers[channel.Server]
	if !exists {
		return
	}

	for i, cID := range server.Channels {
		if cID == event.ID {
			server.Channels = sliceRemoveIndex(server.Channels, i)
			return
		}
	}
}

func (s *State) createServer(event *EventServerCreate) {

	if !s.TrackServers {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.servers[event.ID] = event.Server
}

func (s *State) updateServer(event *AbstractEventUpdate) {

	if !s.TrackServers {
		return
	}

	server := s.servers[event.ID.StringID]
	if server == nil {
		log.Printf("unknown server update %s\n", event.ID)
		return
	}

	s.Lock()
	defer s.Unlock()

	mergeJSON[Server](server, event.Data, event.Clear)
}

func (s *State) deleteServer(event *EventServerDelete) {

	if !s.TrackServers {
		return
	}

	s.Lock()
	defer s.Unlock()

	delete(s.servers, event.ID)
}

func (s *State) updateUser(event *AbstractEventUpdate) {

	if !s.TrackUsers {
		return
	}

	user := s.users[event.ID.StringID]
	if user == nil {
		// For self-bots, this will ignore a lot of events; maybe add more mechanisms for caching users?
		// log.Printf("unknown user update %s\n", event.ID)
		return
	}

	s.Lock()
	defer s.Unlock()

	mergeJSON[User](user, event.Data, event.Clear)
}

func (s *State) createEmoji(event *EventEmojiCreate) {

	if !s.TrackEmojis {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.emojis[event.ID] = event.Emoji
}

func (s *State) deleteEmoji(event *EventEmojiDelete) {

	if !s.TrackEmojis {
		return
	}

	s.Lock()
	defer s.Unlock()

	delete(s.emojis, event.ID)
}

func (s *State) createWebhook(event *EventWebhookCreate) {

	if !s.TrackWebhooks {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.webhooks[event.ID] = event.Webhook
}

func (s *State) updateWebhook(event *AbstractEventUpdate) {

	if !s.TrackWebhooks {
		return
	}

	webhook := s.webhooks[event.ID.StringID]
	if webhook == nil {
		log.Printf("unknown webhook update %s\n", event.ID)
		return
	}

	s.Lock()
	defer s.Unlock()

	mergeJSON[Webhook](webhook, event.Data, event.Clear)
}

func (s *State) deleteWebhook(event *EventWebhookDelete) {

	if !s.TrackWebhooks {
		return
	}

	s.Lock()
	defer s.Unlock()

	delete(s.webhooks, event.ID)
}
