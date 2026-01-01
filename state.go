package revoltgo

import (
	"log"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
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

	if !s.trackAPICalls || user == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
}

func (s *State) addServer(server *Server) {

	if !s.trackAPICalls || server == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.servers[server.ID] = server
}

func (s *State) addChannel(channel *Channel) {

	if !s.trackAPICalls || channel == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.channels[channel.ID] = channel
}

func (s *State) addServerMember(member *ServerMember) {

	if !s.trackAPICalls || member == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.members.add(member)
}

func (s *State) addServerMembersAndUsers(data *ServerMembers) {

	if !s.trackBulkAPICalls || data == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range data.Users {
		s.users[user.ID] = user
	}

	if len(data.Members) == 0 {
		return
	}

	s.members.addMany(data.Members)
}

func (s *State) addEmoji(emoji *Emoji) {

	if !s.trackAPICalls || emoji == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.emojis[emoji.ID] = emoji
}

func (s *State) addWebhooks(webhook []*Webhook) {

	if !s.trackBulkAPICalls || webhook == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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
	delete(s.users, event.UserID)
	s.mu.Unlock()
}

func (s *State) updateServerRole(event *EventServerRoleUpdate) {

	if !s.trackServers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	server := s.servers[event.ID]
	if server == nil {
		log.Printf("update for role %s in unknown server %s\n", event.RoleID, event.ID)
		return
	}

	role := server.Roles[event.RoleID]
	if role == nil {
		log.Printf("update for unknown role %s in server %s\n", event.RoleID, event.ID)
		return
	}

	role.update(event.Data)
	role.clear(event.Clear)
}

func (s *State) deleteServerRole(data *EventServerRoleDelete) {

	if !s.trackServers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	server := s.servers[data.ID]
	if server != nil {
		delete(server.Roles, data.RoleID)
	}
}

func (s *State) createServerMember(data *EventServerMemberJoin) {

	if !s.trackMembers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	member := &ServerMember{
		ID:       MemberCompositeID{User: data.User, Server: data.ID},
		JoinedAt: Timestamp{Time: time.Now()},
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

func (s *State) updateServerMember(event *EventServerMemberUpdate) {

	if !s.trackMembers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	members := s.members[event.ID.User]
	if members == nil {
		members = make(map[string]*ServerMember, initialMembersSize)
		s.members[event.ID.Server] = members
	}

	member := members[event.ID.User]
	if member == nil {
		member = &ServerMember{ID: event.ID}
		members[event.ID.User] = member
	}

	member.update(event.Data)
	member.clear(event.Clear)
}

func (s *State) createChannel(event *EventChannelCreate) {

	if !s.trackChannels {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.channels[event.ID] = &event.Channel

	if !s.trackServers {
		return
	}

	if event.Server == nil {
		return // Channel was not created in server
	}

	server := s.servers[*event.Server]
	if server == nil {
		log.Printf("channel %s created in unknown server %s\n", event.ID, *event.Server)
		return
	}

	server.Channels = append(server.Channels, event.ID)
}

func (s *State) addGroupParticipant(event *EventChannelGroupJoin) {

	if !s.trackChannels {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	channel := s.channels[event.ID]
	if channel == nil {
		log.Printf("%s joined unknown group: %s\n", event.User, event.ID)
		return
	}

	channel.Recipients = append(channel.Recipients, event.User)
}

func (s *State) removeGroupParticipant(event *EventChannelGroupLeave) {

	if !s.trackChannels {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *State) updateChannel(event *EventChannelUpdate) {

	if !s.trackChannels {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	channel := s.channels[event.ID]
	if channel == nil {
		log.Printf("unknown channel updated %s\n", event.ID)
		return
	}

	channel.update(event.Data)
	channel.clear(event.Clear)
}

func (s *State) deleteChannel(event *EventChannelDelete) {

	if !s.trackChannels {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	channel := s.channels[event.ID]
	if channel == nil {
		log.Printf("unknown channel deleted %s\n", event.ID)
		return
	}

	delete(s.channels, event.ID)

	if !s.trackServers {
		return
	}

	server := s.servers[*channel.Server]
	if server == nil {
		log.Printf("channel %s deleted from unknown server %s\n", event.ID, *channel.Server)
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

	if !s.trackServers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.servers[event.ID] = event.Server

	// If there's something you'll be first at in life, it's being a member in your own server.
	if s.trackMembers {
		member := &ServerMember{
			ID: MemberCompositeID{User: s.self.ID, Server: event.ID},
		}

		if id, err := ulid.Parse(event.Server.ID); err == nil {
			member.JoinedAt = Timestamp{Time: ulid.Time(id.Time())}
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

func (s *State) updateServer(event *EventServerUpdate) {

	if !s.trackServers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	server := s.servers[event.ID]
	if server == nil {
		log.Printf("unknown server update %s\n", event.ID)
		return
	}

	server.update(event.Data)
	server.clear(event.Clear)
}

func (s *State) deleteServer(event *EventServerDelete) {

	if !s.trackServers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.servers, event.ID)

	if !s.trackMembers {
		return
	}

	delete(s.members, event.ID)
}

func (s *State) updateUser(event *EventUserUpdate) {

	if !s.trackUsers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user := s.users[event.ID]
	if user == nil {
		// For self-bots, this will ignore a lot of events; maybe add more mechanisms for caching users?
		// log.Printf("unknown user update %s\n", event.ID)
		return
	}

	user.update(event.Data)
	user.clear(event.Clear)
}

func (s *State) createEmoji(event *EventEmojiCreate) {

	if !s.trackEmojis {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.emojis[event.ID] = &event.Emoji
}

func (s *State) deleteEmoji(event *EventEmojiDelete) {

	if !s.trackEmojis {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.emojis, event.ID)
}

func (s *State) createWebhook(event *EventWebhookCreate) {

	if !s.trackWebhooks {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.webhooks[event.ID] = &event.Webhook
}

func (s *State) updateWebhook(event *EventWebhookUpdate) {

	if !s.trackWebhooks {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	webhook := s.webhooks[event.ID]
	if webhook == nil {
		log.Printf("unknown webhook update %s\n", event.ID)
		return
	}

	webhook.update(event.Data)
	webhook.clear(event.Remove) // todo: does WebhookUpdate still call clear "Remove" in the API WS?
}

func (s *State) deleteWebhook(event *EventWebhookDelete) {

	if !s.trackWebhooks {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.webhooks, event.ID)
}
