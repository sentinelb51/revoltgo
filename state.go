package revoltgo

import (
	"iter"
	"log"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oklog/ulid/v2"
)

type uIDtoMember map[string]*ServerMember

// stateMembers maps a Server.ID -> [ User.ID -> ServerMember.ID ]
type stateMembers map[string]uIDtoMember

// add adds a singular member to a server's members
func (sm stateMembers) add(member *ServerMember) {
	// Get the members for a particular server
	members := sm[member.ID.Server]

	// If the server's members are not allocated, allocate them
	if members == nil {
		members = make(uIDtoMember)
		sm[member.ID.Server] = members
	}

	members[member.ID.User] = member
}

// addMany adds multiple members to multiple servers
// Note that this does not lock the state; the caller must handle this
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
			members = make(uIDtoMember, len(serverMembers))
			sm[serverID] = members
		}

		// Add the members to the server
		for _, member := range serverMembers {
			members[member.ID.User] = member
		}
	}
}

// get returns a server's member, or nil if the server or user is not cached.
// Indexing a nil map is safe, so no nil check is needed.
func (sm stateMembers) get(sID, uID string) *ServerMember {
	return sm[sID][uID]
}

// countInServer returns how many members are cached for a server.
func (sm stateMembers) countInServer(sID string) int {
	return len(sm[sID])
}

// remove drops a single membership from a server.
func (sm stateMembers) remove(sID, uID string) {
	delete(sm[sID], uID)
}

// removeServer drops a server's entire member cache.
func (sm stateMembers) removeServer(sID string) {
	delete(sm, sID)
}

// removeUser drops a user from every server they are cached in.
func (sm stateMembers) removeUser(uID string) {
	for _, members := range sm {
		delete(members, uID)
	}
}

// upsert returns the cached member for id, creating an empty one if absent.
func (sm stateMembers) upsert(id MemberCompositeID) *ServerMember {
	members := sm[id.Server]
	if members == nil {
		members = make(uIDtoMember)
		sm[id.Server] = members
	}

	member := members[id.User]
	if member == nil {
		member = &ServerMember{ID: id}
		members[id.User] = member
	}

	return member
}

type State struct {
	self atomic.Pointer[User] // The current user, also present in users

	/* Caches */
	users    map[string]*User    // User.ID    -> User
	servers  map[string]*Server  // Server.ID  -> Server
	channels map[string]*Channel // Channel.ID -> Channel
	emojis   map[string]*Emoji   // Emoji.ID   -> Emoji.
	members  stateMembers        // Server.ID  -> [ User.ID -> Member.ID ]

	/* Mutexes for caches */
	usersMu    sync.RWMutex
	serversMu  sync.RWMutex
	channelsMu sync.RWMutex
	membersMu  sync.RWMutex
	emojisMu   sync.RWMutex

	/* tracking options */

	trackUsers    bool
	trackServers  bool
	trackChannels bool
	trackMembers  bool
	trackEmojis   bool

	// trackAPICalls additionally updates the state from API calls
	// This concept may future-proof against any de-syncs, but may use more CPU time
	trackAPICalls bool

	// trackBulkAPICalls will update the state from bulk API calls
	// This option activates internal State.addServerMembersAndUsers methods
	trackBulkAPICalls bool
}

/*
	Getter functions
*/

func (s *State) Self() *User {
	return s.self.Load()
}

// setSelf is meant to be used when we can't rely on the READY event and must fetch ourselves from the API
func (s *State) setSelf(user *User) {
	s.self.Store(user)
}

func (s *State) TrackUsers() bool {
	return s.trackUsers
}

func (s *State) TrackServers() bool {
	return s.trackServers
}

func (s *State) TrackChannels() bool {
	return s.trackChannels
}

func (s *State) TrackMembers() bool {
	return s.trackMembers
}

func (s *State) TrackEmojis() bool {
	return s.trackEmojis
}

func (s *State) TrackAPICalls() bool {
	return s.trackAPICalls
}

func (s *State) TrackBulkAPICalls() bool {
	return s.trackBulkAPICalls
}

func (s *State) User(id string) *User {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	return s.users[id]
}

func (s *State) UserCount() int {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	return len(s.users)
}

// UserSeq iterates all users without allocating a slice. The same loop-body
// rules as MembersSeq apply: keep it quick and don't call other State methods
// from inside it. Use Users if you need a snapshot.
func (s *State) UserSeq() iter.Seq[*User] {
	return func(yield func(*User) bool) {
		s.usersMu.RLock()
		defer s.usersMu.RUnlock()

		for _, user := range s.users {
			if !yield(user) {
				return
			}
		}
	}
}

// Users returns a slice of all users in state. For general use, User(id) is more common
func (s *State) Users() []*User {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users
}

func (s *State) Server(id string) *Server {
	s.serversMu.RLock()
	defer s.serversMu.RUnlock()

	return s.servers[id]
}

func (s *State) ServerCount() int {
	s.serversMu.RLock()
	defer s.serversMu.RUnlock()

	return len(s.servers)
}

// ServerSeq iterates all servers without allocating a slice. The same loop-body
// rules as MembersSeq apply: keep it quick and don't call other State methods
// from inside it. Use Servers if you need a snapshot.
func (s *State) ServerSeq() iter.Seq[*Server] {
	return func(yield func(*Server) bool) {
		s.serversMu.RLock()
		defer s.serversMu.RUnlock()

		for _, server := range s.servers {
			if !yield(server) {
				return
			}
		}
	}
}

// Servers returns a slice of all servers in state. For general use, Server(id) is more common
func (s *State) Servers() []*Server {
	s.serversMu.RLock()
	defer s.serversMu.RUnlock()

	servers := make([]*Server, 0, len(s.servers))
	for _, server := range s.servers {
		servers = append(servers, server)
	}

	return servers
}

func (s *State) Role(sID, rID string) *ServerRole {
	s.serversMu.RLock()
	defer s.serversMu.RUnlock()

	server := s.servers[sID]
	if server == nil {
		return nil
	}

	return server.Roles[rID]
}

func (s *State) Channel(id string) *Channel {
	s.channelsMu.RLock()
	defer s.channelsMu.RUnlock()

	return s.channels[id]
}

func (s *State) ChannelCount() int {
	s.channelsMu.RLock()
	defer s.channelsMu.RUnlock()

	return len(s.channels)
}

// ChannelSeq iterates all channels without allocating a slice. The same loop-body
// rules as MembersSeq apply: keep it quick and don't call other State methods
// from inside it. Use Channels if you need a snapshot.
func (s *State) ChannelSeq() iter.Seq[*Channel] {
	return func(yield func(*Channel) bool) {
		s.channelsMu.RLock()
		defer s.channelsMu.RUnlock()

		for _, channel := range s.channels {
			if !yield(channel) {
				return
			}
		}
	}
}

// Channels returns a slice of all channels in state. For general use, Channel(id) is more common
func (s *State) Channels() []*Channel {
	s.channelsMu.RLock()
	defer s.channelsMu.RUnlock()

	channels := make([]*Channel, 0, len(s.channels))
	for _, channel := range s.channels {
		channels = append(channels, channel)
	}

	return channels
}

// Members returns a snapshot slice of a server's members. The members are
// copied into a fresh slice while locked, then the lock is released, so you are
// free to do anything inside your loop afterwards, including calling other State
// methods. The trade-off is one slice allocation per call. If your loop only
// needs a quick, read-only pass, prefer MembersSeq to skip that allocation.
func (s *State) Members(sID string) []*ServerMember {
	s.membersMu.RLock()
	defer s.membersMu.RUnlock()

	serverMembers := s.members[sID]

	members := make([]*ServerMember, 0, len(serverMembers))
	for _, member := range serverMembers {
		members = append(members, member)
	}

	return members
}

func (s *State) Member(sID, uID string) *ServerMember {
	s.membersMu.RLock()
	defer s.membersMu.RUnlock()

	return s.members.get(sID, uID)
}

// MemberCount is a helper function to avoid costly len(s.Members(sID)) calls; avoid allocating a whole slice just to count
func (s *State) MemberCount(sID string) int {
	s.membersMu.RLock()
	defer s.membersMu.RUnlock()
	return s.members.countInServer(sID)
}

// MembersSeq iterates a server's members without allocating a slice:
//
//	for member := range session.State.MembersSeq(serverID) {
//		// ...
//	}
//
// The read lock is held for the whole loop, so keep the body quick and don't
// call other State methods from inside it — the lock is already held, so doing
// so can deadlock. Break/return is fine. If you need either, use Members.
func (s *State) MembersSeq(sID string) iter.Seq[*ServerMember] {
	return func(yield func(*ServerMember) bool) {
		s.membersMu.RLock()
		defer s.membersMu.RUnlock()

		for _, member := range s.members[sID] {
			if !yield(member) {
				return
			}
		}
	}
}

func (s *State) Emoji(id string) *Emoji {
	s.emojisMu.RLock()
	defer s.emojisMu.RUnlock()

	return s.emojis[id]
}

func (s *State) EmojiCount() int {
	s.emojisMu.RLock()
	defer s.emojisMu.RUnlock()

	return len(s.emojis)
}

// EmojiSeq iterates all emojis without allocating a slice. The same loop-body
// rules as MembersSeq apply: keep it quick and don't call other State methods
// from inside it. Use Emojis if you need a snapshot.
func (s *State) EmojiSeq() iter.Seq[*Emoji] {
	return func(yield func(*Emoji) bool) {
		s.emojisMu.RLock()
		defer s.emojisMu.RUnlock()

		for _, emoji := range s.emojis {
			if !yield(emoji) {
				return
			}
		}
	}
}

// Emojis returns a slice of all emojis in state. For general use, Emoji(id) is more common
func (s *State) Emojis() []*Emoji {
	s.emojisMu.RLock()
	defer s.emojisMu.RUnlock()

	emojis := make([]*Emoji, 0, len(s.emojis))
	for _, emoji := range s.emojis {
		emojis = append(emojis, emoji)
	}

	return emojis
}

/*
	API call updates
	Used when (State.trackAPICalls or State.trackBulkAPICalls) is enabled
*/

func (s *State) addUser(user *User) {

	if !s.trackAPICalls || user == nil {
		return
	}

	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	s.users[user.ID] = user
}

func (s *State) addServer(server *Server) {

	if !s.trackAPICalls || server == nil {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	s.servers[server.ID] = server
}

func (s *State) addChannel(channel *Channel) {

	if !s.trackAPICalls || channel == nil {
		return
	}

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

	s.channels[channel.ID] = channel
}

func (s *State) addServerMember(member *ServerMember) {

	if !s.trackAPICalls || member == nil {
		return
	}

	s.membersMu.Lock()
	defer s.membersMu.Unlock()

	s.members.add(member)
}

func (s *State) addServerMembersAndUsers(users []*User, members []*ServerMember) {

	if !s.trackBulkAPICalls {
		return
	}

	var (
		shouldProcessUsers   = len(users) != 0 && s.trackUsers
		shouldProcessMembers = len(members) != 0 && s.trackMembers
	)

	if !shouldProcessUsers && !shouldProcessMembers {
		return
	}

	if shouldProcessUsers {
		s.usersMu.Lock()
		for _, user := range users {
			s.users[user.ID] = user
		}
		s.usersMu.Unlock()
	}

	if shouldProcessMembers {
		s.membersMu.Lock()
		s.members.addMany(members)
		s.membersMu.Unlock()
	}
}

func (s *State) addEmoji(emoji *Emoji) {

	if !s.trackAPICalls || emoji == nil {
		return
	}

	s.emojisMu.Lock()
	defer s.emojisMu.Unlock()

	s.emojis[emoji.ID] = emoji
}

// StateConfig controls which entity caches the State maintains. Pass it to
// Session.Open. The zero value tracks nothing.
// Tracking is immutable once Session.Open() has connected.
type StateConfig struct {
	TrackUsers    bool
	TrackServers  bool
	TrackChannels bool
	TrackMembers  bool
	TrackEmojis   bool

	// TrackAPICalls additionally updates the state from single API calls
	TrackAPICalls bool

	// TrackBulkAPICalls additionally updates the state from bulk API calls
	TrackBulkAPICalls bool
}

// DefaultStateConfig returns a StateConfig that tracks everything.
func DefaultStateConfig() StateConfig {
	return StateConfig{
		TrackUsers:        true,
		TrackServers:      true,
		TrackChannels:     true,
		TrackMembers:      true,
		TrackEmojis:       true,
		TrackAPICalls:     true,
		TrackBulkAPICalls: true,
	}
}

func (s *State) applyConfig(c StateConfig) {
	s.trackUsers = c.TrackUsers
	s.trackServers = c.TrackServers
	s.trackChannels = c.TrackChannels
	s.trackMembers = c.TrackMembers
	s.trackEmojis = c.TrackEmojis
	s.trackAPICalls = c.TrackAPICalls
	s.trackBulkAPICalls = c.TrackBulkAPICalls
}

func newState() *State {
	s := &State{
		// We pre-alloc incase someone uses the library in an HTTP-before-READY way
		users:    make(map[string]*User),
		servers:  make(map[string]*Server),
		channels: make(map[string]*Channel),
		members:  make(stateMembers),
		emojis:   make(map[string]*Emoji),
	}

	s.applyConfig(DefaultStateConfig())
	return s
}

// populate populates the state with the data from the ready event.
// It will overwrite any existing data in the state.
func (s *State) populate(ready *EventReady) {

	if len(ready.Users) > 0 {
		// The last user in the ready event is the current user
		self := ready.Users[len(ready.Users)-1]
		s.self.Store(self)
	}

	/* Populate the caches */
	if s.trackUsers {

		s.usersMu.Lock()
		s.users = make(map[string]*User, len(ready.Users))
		for _, user := range ready.Users {
			s.users[user.ID] = user
		}
		s.usersMu.Unlock()
	}

	if s.trackServers {
		s.serversMu.Lock()
		s.servers = make(map[string]*Server, len(ready.Servers))
		for _, server := range ready.Servers {
			s.servers[server.ID] = server
		}
		s.serversMu.Unlock()
	}

	if s.trackChannels {
		s.channelsMu.Lock()
		s.channels = make(map[string]*Channel, len(ready.Channels))
		for _, channel := range ready.Channels {
			s.channels[channel.ID] = channel
		}
		s.channelsMu.Unlock()
	}

	if s.trackMembers {
		s.membersMu.Lock()
		s.members = make(stateMembers, len(ready.Servers))
		s.members.addMany(ready.Members)
		s.membersMu.Unlock()
	}

	if s.trackEmojis {
		s.emojisMu.Lock()
		s.emojis = make(map[string]*Emoji, len(ready.Emojis))
		for _, emoji := range ready.Emojis {
			s.emojis[emoji.ID] = emoji
		}
		s.emojisMu.Unlock()
	}
}

// platformWipe removes a user from users, channels (dms and groups), and servers member lists.
// It ignores all State.TrackX fields; the user is banned off the platform.
func (s *State) platformWipe(event *EventUserPlatformWipe) {
	// Remove from users
	s.usersMu.Lock()
	delete(s.users, event.UserID)
	s.usersMu.Unlock()

	// Remove direct messages or participant information
	s.channelsMu.Lock()
	for _, channel := range s.channels {
		switch channel.ChannelType {
		case ChannelTypeDM:
			delete(s.channels, channel.ID)
		case ChannelTypeGroup:
			if slices.Contains(channel.Recipients, event.UserID) {
				delete(s.channels, channel.ID)
			}
		}
	}
	s.channelsMu.Unlock()

	// Remove server memberships
	s.membersMu.Lock()
	s.members.removeUser(event.UserID)
	s.membersMu.Unlock()
}

func (s *State) updateServerRoleRanks(event *EventServerRoleRanksUpdate) {

	if !s.trackServers {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	server := s.servers[event.ID]
	if server == nil {
		log.Printf("role ranks update for unknown server %s\n", event.ID)
		return
	}

	for index, rID := range event.Ranks {
		role, exists := server.Roles[rID]
		if !exists {
			log.Printf("role ranks update for unknown role %s in server %s\n", rID, event.ID)
			continue
		}

		role.Rank = int64(index)
	}
}

func (s *State) updateServerRole(event *EventServerRoleUpdate) {

	if !s.trackServers {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	server := s.servers[event.ID]
	if server == nil {
		log.Printf("update for role %s in unknown server %s\n", event.RoleID, event.ID)
		return
	}

	role := server.Roles[event.RoleID]
	if role == nil {
		// Role was created
		role = new(ServerRole)
		server.Roles[event.RoleID] = role
	}

	role.update(event.Data)
	role.clear(event.Clear)
}

func (s *State) deleteServerRole(data *EventServerRoleDelete) {

	if !s.trackServers {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	server := s.servers[data.ID]
	if server != nil {
		delete(server.Roles, data.RoleID)
	}
}

func (s *State) createServerMember(data *EventServerMemberJoin) {

	if !s.trackMembers {
		return
	}

	s.membersMu.Lock()
	defer s.membersMu.Unlock()

	member := &ServerMember{
		ID:       MemberCompositeID{User: data.User, Server: data.ID},
		JoinedAt: time.Now(),
	}

	s.members.add(member)
}

func (s *State) deleteServerMember(data *EventServerMemberLeave) {

	/*
		If the user that left is us, we left the server, thus:
			We need to remove the server and its members from the state;
				deleteServer() already handles server deletion, so we construct an artificial event
					to handle it.

		Note: this is not the same as the server being deleted; server still exists, but YOU are not in it.
	*/

	self := s.Self()
	if self != nil && data.User == self.ID {
		s.deleteServer(&EventServerDelete{ID: data.ID})
		return
	}

	if !s.trackMembers {
		return
	}

	s.membersMu.Lock()
	defer s.membersMu.Unlock()

	s.members.remove(data.ID, data.User)
}

func (s *State) updateServerMember(event *EventServerMemberUpdate) {

	if !s.trackMembers {
		return
	}

	s.membersMu.Lock()
	defer s.membersMu.Unlock()

	member := s.members.upsert(event.ID)

	member.update(event.Data)
	member.clear(event.Clear)
}

func (s *State) createChannel(event *EventChannelCreate) {

	if !s.trackChannels {
		return
	}

	s.channelsMu.Lock()
	s.channels[event.ID] = &event.Channel
	s.channelsMu.Unlock()

	if !s.trackServers {
		return
	}

	if event.Server == nil {
		return // Channel was not created in server
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

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

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

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

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

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

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

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

	s.channelsMu.Lock()
	channel := s.channels[event.ID]
	if channel == nil {
		s.channelsMu.Unlock()
		log.Printf("unknown channel deleted %s\n", event.ID)
		return
	}

	delete(s.channels, event.ID)
	s.channelsMu.Unlock()

	if !s.trackServers {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

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

	s.serversMu.Lock()
	s.servers[event.ID] = event.Server
	s.serversMu.Unlock()

	// If there's something you'll be first at in life, it's being a member in your own server.
	if s.trackMembers {
		s.membersMu.Lock()
		member := &ServerMember{
			ID: MemberCompositeID{User: s.self.Load().ID, Server: event.ID},
		}

		if id, err := ulid.Parse(event.Server.ID); err == nil {
			member.JoinedAt = ulid.Time(id.Time())
		}

		s.members.add(member)
		s.membersMu.Unlock()
	}

	if s.trackChannels {
		s.channelsMu.Lock()
		for _, channel := range event.Channels {
			// Need to add directly here to avoid nested lock acquisition
			if channel != nil {
				s.channels[channel.ID] = channel
			}
		}
		s.channelsMu.Unlock()
	}

	if s.trackEmojis {
		s.emojisMu.Lock()
		for _, emoji := range event.Emojis {
			// Need to add directly here to avoid nested lock acquisition
			if emoji != nil {
				s.emojis[emoji.ID] = emoji
			}
		}
		s.emojisMu.Unlock()
	}
}

func (s *State) updateServer(event *EventServerUpdate) {

	if !s.trackServers {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	server := s.servers[event.ID]
	if server == nil {
		log.Printf("unknown server update %s\n", event.ID)
		return
	}

	server.update(event.Data)
	server.clear(event.Clear)
}

func (s *State) deleteServer(event *EventServerDelete) {

	if s.trackServers {
		s.serversMu.Lock()
		delete(s.servers, event.ID)
		s.serversMu.Unlock()
	}

	if s.trackMembers {
		s.membersMu.Lock()
		s.members.removeServer(event.ID)
		s.membersMu.Unlock()
	}
}

func (s *State) updateUser(event *EventUserUpdate) {

	if !s.trackUsers {
		return
	}

	s.usersMu.Lock()
	defer s.usersMu.Unlock()

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

	s.emojisMu.Lock()
	defer s.emojisMu.Unlock()

	s.emojis[event.ID] = &event.Emoji
}

func (s *State) deleteEmoji(event *EventEmojiDelete) {

	if !s.trackEmojis {
		return
	}

	s.emojisMu.Lock()
	defer s.emojisMu.Unlock()

	delete(s.emojis, event.ID)
}
