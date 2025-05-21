package revoltgo

import (
	"github.com/oklog/ulid/v2"
	"log"
	"sync"
	"time"
)

const initialMembersSize = 50

// stateMembers maps a Server.ID to "members"; "members" maps a User.ID to ServerMember
type stateMembers map[string]map[string]*ServerMember

// add adds a singular member to a server's members
func (sm stateMembers) add(member *ServerMember) {
	// Get the members for a particular server
	members := sm[member.ID.Server]

	// If the server's members are not allocated, allocate them
	if members == nil {
		members = make(map[string]*ServerMember, initialMembersSize)
		sm[member.ID.Server] = members
	}

	members[member.ID.User] = member
}

// addMany adds multiple members to multiple servers
func (sm stateMembers) addMany(members []*ServerMember) {
	// Group members based on their server ID
	groups := make(map[string][]*ServerMember)
	for _, member := range members {
		groups[member.ID.Server] = append(groups[member.ID.Server], member)
	}

	// For each server, fetch or allocate members, and add them in bulk
	for serverID, serverMembers := range groups {
		// Get the members for a particular server
		members := sm[serverID]

		// If the server's members are not allocated, allocate them
		if members == nil {
			members = make(map[string]*ServerMember, len(serverMembers))
			sm[serverID] = members
		}

		// Add the members to the server
		for _, member := range serverMembers {
			members[member.ID.User] = member
		}
	}
}

type State struct {
	mu   sync.RWMutex // Global mutex for race safety
	self *User        // The current user, also present in users

	/* Caches */
	users    map[string]*User
	servers  map[string]*Server
	channels map[string]*Channel
	emojis   map[string]*Emoji
	webhooks map[string]*Webhook

	// members maps Server.ID to members
	members stateMembers

	/* tracking options */

	trackUsers    bool
	trackServers  bool
	trackChannels bool
	trackMembers  bool
	trackEmojis   bool
	trackWebhooks bool

	// trackAPICalls additionally updates the state from API calls
	// This concept may future-proof against any de-syncs, but may use more CPU time
	trackAPICalls bool

	// trackBulkAPICalls will update the state from bulk API calls
	// This option activates internal State.addServerMembersAndUsers and State.addWebhooks functions
	trackBulkAPICalls bool
}

/*
	Getter functions
*/

func (s *State) Self() *User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.self
}

func (s *State) TrackUsers() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.trackUsers
}

func (s *State) TrackServers() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.trackServers
}

func (s *State) TrackChannels() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.trackChannels
}

func (s *State) TrackMembers() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.trackMembers
}

func (s *State) TrackEmojis() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.trackEmojis
}

func (s *State) TrackWebhooks() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.trackWebhooks
}

func (s *State) TrackAPICalls() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.trackAPICalls
}

func (s *State) TrackBulkAPICalls() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.trackBulkAPICalls
}

func (s *State) User(id string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.users[id]
}

func (s *State) Server(id string) *Server {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.servers[id]
}

func (s *State) Role(sID, rID string) *ServerRole {
	s.mu.RLock()
	defer s.mu.RUnlock()

	server := s.servers[sID]
	if server == nil {
		return nil
	}

	return server.Roles[rID]
}

func (s *State) Channel(id string) *Channel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.channels[id]
}

func (s *State) Members(sID string) []*ServerMember {
	s.mu.RLock()
	defer s.mu.RUnlock()

	serverMembers := s.members[sID]

	members := make([]*ServerMember, 0, len(serverMembers))
	for _, member := range serverMembers {
		members = append(members, member)
	}

	return members
}

func (s *State) Member(uID, sID string) *ServerMember {
	s.mu.RLock()
	defer s.mu.RUnlock()

	members := s.members[sID]
	if members == nil {
		return nil
	}

	return members[uID]
}

func (s *State) Emoji(id string) *Emoji {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.emojis[id]
}

func (s *State) Webhook(id string) *Webhook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.webhooks[id]
}

/*
	API call updates
	Used when (State.trackAPICalls or State.trackBulkAPICalls) is enabled
*/

func (s *State) addUser(user *User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackAPICalls || user == nil {
		return
	}

	s.users[user.ID] = user
}

func (s *State) addServer(server *Server) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackAPICalls || server == nil {
		return
	}

	s.servers[server.ID] = server
}

func (s *State) addChannel(channel *Channel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackAPICalls || channel == nil {
		return
	}

	s.channels[channel.ID] = channel
}

func (s *State) addServerMember(member *ServerMember) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackAPICalls || member == nil {
		return
	}

	s.members.add(member)
}

func (s *State) addServerMembersAndUsers(data *ServerMembers) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackBulkAPICalls || data == nil {
		return
	}

	for _, user := range data.Users {
		s.users[user.ID] = user
	}

	if len(data.Members) == 0 {
		return
	}

	s.members.addMany(data.Members)
}

func (s *State) addEmoji(emoji *Emoji) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackAPICalls || emoji == nil {
		return
	}

	s.emojis[emoji.ID] = emoji
}

func (s *State) addWebhooks(webhook []*Webhook) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackBulkAPICalls || webhook == nil {
		return
	}

	for _, w := range webhook {
		s.webhooks[w.ID] = w
	}
}

// todo: merge this with populate
func newState() *State {
	return &State{
		trackUsers:        true,
		trackServers:      true,
		trackChannels:     true,
		trackMembers:      true,
		trackEmojis:       true,
		trackAPICalls:     true,
		trackBulkAPICalls: true,
		trackWebhooks:     false,
	}
}

// populate populates the state with the data from the ready event.
// It will overwrite any existing data in the state.
func (s *State) populate(ready *EventReady) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// The last user in the ready event is the current user
	s.self = ready.Users[len(ready.Users)-1]

	/* Populate the caches */
	if s.trackUsers {
		s.users = make(map[string]*User, len(ready.Users))
		for _, user := range ready.Users {
			s.users[user.ID] = user
		}
	}

	if s.trackServers {
		s.servers = make(map[string]*Server, len(ready.Servers))
		for _, server := range ready.Servers {
			s.servers[server.ID] = server
		}
	}

	if s.trackChannels {
		s.channels = make(map[string]*Channel, len(ready.Channels))
		for _, channel := range ready.Channels {
			s.channels[channel.ID] = channel
		}
	}

	if s.trackMembers {
		s.members = make(stateMembers, len(ready.Members))
		for _, member := range ready.Members {
			s.members.add(member)
		}
	}

	if s.trackEmojis {
		s.emojis = make(map[string]*Emoji, len(ready.Emojis))
		for _, emoji := range ready.Emojis {
			s.emojis[emoji.ID] = emoji
		}
	}
}

func (s *State) platformWipe(event *EventUserPlatformWipe) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.users, event.UserID)
}

func (s *State) updateServerRole(event *AbstractEventUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackServers {
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

	mergeJSON[ServerRole](role, event.Data, event.Clear)
}

func (s *State) deleteServerRole(data *EventServerRoleDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackServers {
		return
	}

	server := s.servers[data.ID]
	if server != nil {
		delete(server.Roles, data.RoleID)
	}
}

func (s *State) createServerMember(data *EventServerMemberJoin) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackMembers {
		return
	}

	member := &ServerMember{
		ID:       MemberCompositeID{User: data.User, Server: data.ID},
		JoinedAt: time.Now(),
	}

	s.members.add(member)
}

func (s *State) deleteServerMember(data *EventServerMemberLeave) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Client has left the server; remove the server from the state
	// deleteServer() already handles removing server and members from state
	// Note: This is not the same as the server being deleted; the server still exists, you're just not in it.
	if data.User == s.Self().ID {
		event := &EventServerDelete{ID: data.ID}

		// We need to release the lock before calling deleteServer which will acquire it again
		s.mu.Unlock()
		s.deleteServer(event)
		s.mu.Lock() // Reacquire the lock to maintain the defer s.mu.Unlock() contract
		return
	}

	if !s.trackMembers {
		return
	}

	delete(s.members, data.User)
}

func (s *State) updateServerMember(event *AbstractEventUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackMembers {
		return
	}

	mID := event.ID.MemberID
	members := s.members[mID.Server]

	if members == nil {
		members = make(map[string]*ServerMember, initialMembersSize)
		s.members[mID.Server] = members
	}

	member := members[mID.User]
	if member == nil {
		member = &ServerMember{ID: mID}
		members[mID.User] = member
	}

	mergeJSON[ServerMember](member, event.Data, event.Clear)
}

func (s *State) createChannel(event *EventChannelCreate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackChannels {
		return
	}

	server := s.servers[event.Server]
	if server == nil {
		log.Printf("channel %s created in unknown server %s\n", event.ID, event.Server)
		return
	}

	s.channels[event.ID] = event.Channel

	if !s.trackServers {
		return
	}

	if event.Server != "" {
		server.Channels = append(server.Channels, event.ID)
	}
}

func (s *State) addGroupParticipant(event *EventChannelGroupJoin) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackChannels {
		return
	}

	channel := s.channels[event.ID]
	if channel == nil {
		log.Printf("%s joined unknown group: %s\n", event.User, event.ID)
		return
	}

	channel.Recipients = append(channel.Recipients, event.User)
}

func (s *State) removeGroupParticipant(event *EventChannelGroupLeave) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackChannels {
		return
	}

	channel := s.channels[event.ID]
	if channel == nil {
		log.Printf("%s left unknown group %s\n", event.User, event.ID)
		return
	}

	for i, uID := range channel.Recipients {
		if uID == event.User {
			channel.Recipients = sliceRemoveIndex(channel.Recipients, i)
			return
		}
	}
}

func (s *State) updateChannel(event *AbstractEventUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackChannels {
		return
	}

	channel := s.channels[event.ID.StringID]
	if channel == nil {
		log.Printf("unknown channel updated %s\n", event.ID)
		return
	}

	mergeJSON[Channel](channel, event.Data, event.Clear)
}

func (s *State) deleteChannel(event *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackChannels {
		return
	}

	channel := s.channels[event.ID]
	if channel == nil {
		log.Printf("unknown channel deleted %s\n", event.ID)
		return
	}

	delete(s.channels, event.ID)

	server := s.servers[channel.Server]
	if server == nil {
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackServers {
		return
	}

	s.servers[event.ID] = event.Server

	// If there's something you'll be first at in life, it's being a member in your own server.
	if s.trackMembers {
		member := &ServerMember{
			ID: MemberCompositeID{User: s.self.ID, Server: event.ID},
		}

		if id, err := ulid.Parse(event.Server.ID); err == nil {
			member.JoinedAt = ulid.Time(id.Time())
		}

		s.members.add(member)
	}

	if s.trackChannels {
		for _, channel := range event.Channels {
			// Need to add directly here to avoid nested lock acquisition
			if channel != nil {
				s.channels[channel.ID] = channel
			}
		}
	}

	if s.trackEmojis {
		for _, emoji := range event.Emojis {
			// Need to add directly here to avoid nested lock acquisition
			if emoji != nil {
				s.emojis[emoji.ID] = emoji
			}
		}
	}
}

func (s *State) updateServer(event *AbstractEventUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackServers {
		return
	}

	server := s.servers[event.ID.StringID]
	if server == nil {
		log.Printf("unknown server update %s\n", event.ID)
		return
	}

	mergeJSON[Server](server, event.Data, event.Clear)
}

func (s *State) deleteServer(event *EventServerDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackServers {
		return
	}

	delete(s.servers, event.ID)

	if !s.trackMembers {
		return
	}

	delete(s.members, event.ID)
}

func (s *State) updateUser(event *AbstractEventUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackUsers {
		return
	}

	user := s.users[event.ID.StringID]
	if user == nil {
		// For self-bots, this will ignore a lot of events; maybe add more mechanisms for caching users?
		// log.Printf("unknown user update %s\n", event.ID)
		return
	}

	mergeJSON[User](user, event.Data, event.Clear)
}

func (s *State) createEmoji(event *EventEmojiCreate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackEmojis {
		return
	}

	s.emojis[event.ID] = event.Emoji
}

func (s *State) deleteEmoji(event *EventEmojiDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackEmojis {
		return
	}

	delete(s.emojis, event.ID)
}

func (s *State) createWebhook(event *EventWebhookCreate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackWebhooks {
		return
	}

	s.webhooks[event.ID] = event.Webhook
}

func (s *State) updateWebhook(event *AbstractEventUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackWebhooks {
		return
	}

	webhook := s.webhooks[event.ID.StringID]
	if webhook == nil {
		log.Printf("unknown webhook update %s\n", event.ID)
		return
	}

	mergeJSON[Webhook](webhook, event.Data, event.Clear)
}

func (s *State) deleteWebhook(event *EventWebhookDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.trackWebhooks {
		return
	}

	delete(s.webhooks, event.ID)
}
