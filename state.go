package revoltgo

//msgp:ignore State

import (
	"iter"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oklog/ulid/v2"
)

/*
	Cache invariant

	A cached object is immutable once published. A mutator clones the object it
	is about, updates the clone, and swaps it into the map under the write lock;
	it never writes through a pointer a getter has already handed out. A getter
	may therefore return a pointer that goes stale, but never one that changes
	under the reader — which is what makes returning the raw pointer safe, and
	what keeps the Seq iterators allocation-free.

	The same rule bans filing an event's or an API result's own object: the
	caller keeps that one and may write to it. Everything entering a cache is
	copied by the clone helpers below, which are deep enough to cover every map
	and slice a mutator splices.
*/

// cloneUser returns a private copy of user, or nil.
func cloneUser(user *User) *User {

	if user == nil {
		return nil
	}

	next := *user
	next.Relations = slices.Clone(user.Relations)

	return &next
}

// cloneRole returns a private copy of role, or nil.
func cloneRole(role *ServerRole) *ServerRole {

	if role == nil {
		return nil
	}

	next := *role

	return &next
}

// cloneServer returns a private copy of server, or nil. Roles holds pointers,
// so each role is copied too; a struct copy alone would still share them.
func cloneServer(server *Server) *Server {

	if server == nil {
		return nil
	}

	next := *server
	next.Channels = slices.Clone(server.Channels)
	next.Categories = slices.Clone(server.Categories)

	if server.Roles != nil {
		next.Roles = make(map[string]*ServerRole, len(server.Roles))
		for id, role := range server.Roles {
			next.Roles[id] = cloneRole(role)
		}
	}

	return &next
}

// cloneChannel returns a private copy of channel, or nil.
func cloneChannel(channel *Channel) *Channel {

	if channel == nil {
		return nil
	}

	next := *channel
	next.Recipients = slices.Clone(channel.Recipients)
	next.RolePermissions = maps.Clone(channel.RolePermissions)

	return &next
}

// cloneMember returns a private copy of member, or nil.
func cloneMember(member *ServerMember) *ServerMember {

	if member == nil {
		return nil
	}

	next := *member
	next.Roles = slices.Clone(member.Roles)

	return &next
}

// cloneEmoji returns a private copy of emoji, or nil.
func cloneEmoji(emoji *Emoji) *Emoji {

	if emoji == nil {
		return nil
	}

	next := *emoji

	return &next
}

// cloneVoiceState returns a private copy of state, or nil.
func cloneVoiceState(state *UserVoiceState) *UserVoiceState {

	if state == nil {
		return nil
	}

	next := *state

	return &next
}

type uIDtoMember map[string]*ServerMember

// stateMembers maps a Server.ID -> [ User.ID -> ServerMember.ID ]
type stateMembers map[string]uIDtoMember

// add files a member under its server. The cache takes ownership: the member
// must be freshly built or cloned, never one the caller keeps a pointer to.
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

// addMany adds multiple members to multiple servers, copying each one in.
// Note that this does not lock the state; the caller must handle this
func (sm stateMembers) addMany(members []*ServerMember) {
	// Group members based on their server ID
	groups := make(map[string][]*ServerMember)
	for _, member := range members {
		if member == nil {
			continue
		}

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
			members[member.ID.User] = cloneMember(member)
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

// remove drops a single membership from a server, and the server once it is empty.
func (sm stateMembers) remove(sID, uID string) {
	members := sm[sID]
	delete(members, uID)

	if len(members) == 0 {
		delete(sm, sID)
	}
}

// removeServer drops a server's entire member cache.
func (sm stateMembers) removeServer(sID string) {
	delete(sm, sID)
}

// removeUser drops a user from every server they are cached in.
func (sm stateMembers) removeUser(uID string) {
	for sID, members := range sm {
		delete(members, uID)

		if len(members) == 0 {
			delete(sm, sID)
		}
	}
}

type uIDtoVoiceState map[string]*UserVoiceState

// stateVoice maps a Channel.ID -> [ User.ID -> UserVoiceState ]
type stateVoice map[string]uIDtoVoiceState

// add files a participant under a channel. The cache takes ownership: the state
// must be freshly built or cloned, never one the caller keeps a pointer to.
func (sv stateVoice) add(cID string, state *UserVoiceState) {
	// Get the participants for a particular channel
	states := sv[cID]

	// If the channel's participants are not allocated, allocate them
	if states == nil {
		states = make(uIDtoVoiceState)
		sv[cID] = states
	}

	states[state.ID] = state
}

// addMany adds the participants of multiple channels, copying each one in.
// Note that this does not lock the state; the caller must handle this
func (sv stateVoice) addMany(voiceStates []*ChannelVoiceState) {
	// The wire already groups participants by channel, so no grouping pass is needed
	for _, voiceState := range voiceStates {
		if voiceState == nil || len(voiceState.Participants) == 0 {
			continue
		}

		// Get the participants for a particular channel
		states := sv[voiceState.ID]

		// If the channel's participants are not allocated, allocate them
		if states == nil {
			states = make(uIDtoVoiceState, len(voiceState.Participants))
			sv[voiceState.ID] = states
		}

		for _, participant := range voiceState.Participants {
			if participant == nil {
				continue
			}

			states[participant.ID] = cloneVoiceState(participant)
		}
	}
}

// get returns a channel participant's voice state, or nil if the channel or user is not cached.
// Indexing a nil map is safe, so no nil check is needed.
func (sv stateVoice) get(cID, uID string) *UserVoiceState {
	return sv[cID][uID]
}

// countInChannel returns how many participants are cached for a channel.
func (sv stateVoice) countInChannel(cID string) int {
	return len(sv[cID])
}

// remove drops a participant from a channel, and the channel once it is empty.
func (sv stateVoice) remove(cID, uID string) {
	states := sv[cID]
	delete(states, uID)

	if len(states) == 0 {
		delete(sv, cID)
	}
}

// removeChannel drops a channel's entire participant cache.
func (sv stateVoice) removeChannel(cID string) {
	delete(sv, cID)
}

// removeUser drops a user from every channel they are cached in.
func (sv stateVoice) removeUser(uID string) {
	for cID, states := range sv {
		delete(states, uID)

		if len(states) == 0 {
			delete(sv, cID)
		}
	}
}

type State struct {
	self atomic.Pointer[User] // The current user, also present in users

	/* Caches */
	users    map[string]*User    // User.ID    -> User
	servers  map[string]*Server  // Server.ID  -> Server
	channels map[string]*Channel // Channel.ID -> Channel
	emojis   map[string]*Emoji   // Emoji.ID   -> Emoji.
	members  stateMembers        // Server.ID  -> [ User.ID -> Member.ID ]
	voice    stateVoice          // Channel.ID -> [ User.ID -> UserVoiceState ]

	/* Mutexes for caches */
	usersMu    sync.RWMutex
	serversMu  sync.RWMutex
	channelsMu sync.RWMutex
	membersMu  sync.RWMutex
	emojisMu   sync.RWMutex
	voiceMu    sync.RWMutex

	/* tracking options */

	trackUsers    bool
	trackServers  bool
	trackChannels bool
	trackMembers  bool
	trackEmojis   bool
	trackVoice    bool

	// trackAPICalls additionally updates the state from API calls
	// This concept may future-proof against any de-syncs, but may use more CPU time
	trackAPICalls bool

	// trackBulkAPICalls will update the state from bulk API calls
	// This option activates internal State.addServerMembersAndUsers methods
	trackBulkAPICalls bool
}

/*
	Getter functions

	Everything returned here is the cached object itself. Per the cache
	invariant it will never change, so it is safe to read from any goroutine —
	but it can go stale, and writing to it corrupts the cache.
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

func (s *State) TrackVoice() bool {
	return s.trackVoice
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

// VoiceState returns a participant's voice state in a channel, or nil if they are not in the call
func (s *State) VoiceState(cID, uID string) *UserVoiceState {
	s.voiceMu.RLock()
	defer s.voiceMu.RUnlock()

	return s.voice.get(cID, uID)
}

// VoiceStates returns a snapshot slice of a channel's participants. The same
// allocation trade-off as Members applies; prefer VoiceStatesSeq for a quick,
// read-only pass.
func (s *State) VoiceStates(cID string) []*UserVoiceState {
	s.voiceMu.RLock()
	defer s.voiceMu.RUnlock()

	channelStates := s.voice[cID]

	states := make([]*UserVoiceState, 0, len(channelStates))
	for _, state := range channelStates {
		states = append(states, state)
	}

	return states
}

// VoiceStateCount is a helper function to avoid costly len(s.VoiceStates(cID)) calls; avoid allocating a whole slice just to count
func (s *State) VoiceStateCount(cID string) int {
	s.voiceMu.RLock()
	defer s.voiceMu.RUnlock()
	return s.voice.countInChannel(cID)
}

// VoiceStatesSeq iterates a channel's participants without allocating a slice. The same
// loop-body rules as MembersSeq apply: keep it quick and don't call other State methods
// from inside it. Use VoiceStates if you need a snapshot.
func (s *State) VoiceStatesSeq(cID string) iter.Seq[*UserVoiceState] {
	return func(yield func(*UserVoiceState) bool) {
		s.voiceMu.RLock()
		defer s.voiceMu.RUnlock()

		for _, state := range s.voice[cID] {
			if !yield(state) {
				return
			}
		}
	}
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

	s.users[user.ID] = cloneUser(user)
}

func (s *State) addServer(server *Server) {

	if !s.trackAPICalls || server == nil {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	s.servers[server.ID] = cloneServer(server)
}

func (s *State) addChannels(channels []*Channel) {

	if !s.trackBulkAPICalls || len(channels) == 0 {
		return
	}

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

	for _, channel := range channels {
		if channel == nil {
			continue
		}

		s.channels[channel.ID] = cloneChannel(channel)
	}
}

func (s *State) addChannel(channel *Channel) {

	if !s.trackAPICalls || channel == nil {
		return
	}

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

	s.channels[channel.ID] = cloneChannel(channel)
}

func (s *State) addServerMember(member *ServerMember) {

	if !s.trackAPICalls || member == nil {
		return
	}

	s.membersMu.Lock()
	defer s.membersMu.Unlock()

	s.members.add(cloneMember(member))
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
			if user == nil {
				continue
			}

			s.users[user.ID] = cloneUser(user)
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

	s.emojis[emoji.ID] = cloneEmoji(emoji)
}

/*
	Escape hatches

	Nothing evicts a user automatically: one can be reachable from a member row,
	a group recipient list, a message author or a DM, so the cache has no safe
	rule of its own. A consumer that knows its own retention rules needs a way
	to act, and the maps are unexported.
*/

// AddUser files a copy of user in the cache, replacing any record of the same
// ID. It ignores TrackUsers and TrackAPICalls: the caller asked explicitly.
func (s *State) AddUser(user *User) {

	if user == nil {
		return
	}

	next := cloneUser(user)

	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	s.users[next.ID] = next

	if self := s.self.Load(); self != nil && self.ID == next.ID {
		s.self.Store(next)
	}
}

// EvictUser drops a user from the cache. Nothing else does, short of a platform
// wipe, so an account that sees many users grows without this.
func (s *State) EvictUser(id string) {

	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	delete(s.users, id)
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

	// TrackVoice caches who is connected to each voice channel.
	// It also asks for the voice states in the ready event; see Session.buildOpenQueryParams
	TrackVoice bool

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
		TrackVoice:        true,
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
	s.trackVoice = c.TrackVoice
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
		voice:    make(stateVoice),
	}

	s.applyConfig(DefaultStateConfig())
	return s
}

// populate populates the state with the data from the ready event.
// It will overwrite any existing data in the state.
func (s *State) populate(ready *EventReady) {

	/* Populate the caches. The event stays the caller's; the caches take copies */
	if s.trackUsers {

		s.usersMu.Lock()
		s.users = make(map[string]*User, len(ready.Users))
		for _, user := range ready.Users {
			if user == nil {
				continue
			}

			s.users[user.ID] = cloneUser(user)
		}
		s.usersMu.Unlock()
	}

	if len(ready.Users) > 0 {
		// The last user in the ready event should be the current user
		self := ready.Users[len(ready.Users)-1]

		// Sanity check: if last user is indeed self, it should be relationship type User
		if self != nil && self.Relationship == UserRelationshipTypeUser {
			// Prefer the cached copy, so Self and User(id) are the one object
			if cached := s.User(self.ID); cached != nil {
				s.self.Store(cached)
			} else {
				s.self.Store(cloneUser(self))
			}
		}
	}

	if s.trackServers {
		s.serversMu.Lock()
		s.servers = make(map[string]*Server, len(ready.Servers))
		for _, server := range ready.Servers {
			if server == nil {
				continue
			}

			s.servers[server.ID] = cloneServer(server)
		}
		s.serversMu.Unlock()
	}

	if s.trackChannels {
		s.channelsMu.Lock()
		s.channels = make(map[string]*Channel, len(ready.Channels))
		for _, channel := range ready.Channels {
			if channel == nil {
				continue
			}

			s.channels[channel.ID] = cloneChannel(channel)
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
			if emoji == nil {
				continue
			}

			s.emojis[emoji.ID] = cloneEmoji(emoji)
		}
		s.emojisMu.Unlock()
	}

	if s.trackVoice {
		s.voiceMu.Lock()
		s.voice = make(stateVoice, len(ready.VoiceStates))
		s.voice.addMany(ready.VoiceStates)
		s.voiceMu.Unlock()
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

	// Remove them from any call they were in
	s.voiceMu.Lock()
	s.voice.removeUser(event.UserID)
	s.voiceMu.Unlock()
}

func (s *State) updateServerRoleRanks(event *EventServerRoleRanksUpdate) {

	if !s.trackServers {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	server := s.servers[event.ID]
	if server == nil {
		logf("role ranks update for unknown server %s", event.ID)
		return
	}

	next := cloneServer(server)

	for index, rID := range event.Ranks {
		role, exists := next.Roles[rID]
		if !exists {
			logf("role ranks update for unknown role %s in server %s", rID, event.ID)
			continue
		}

		role.Rank = int64(index)
	}

	s.servers[event.ID] = next
}

func (s *State) updateServerRole(event *EventServerRoleUpdate) {

	if !s.trackServers {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	server := s.servers[event.ID]
	if server == nil {
		logf("update for role %s in unknown server %s", event.RoleID, event.ID)
		return
	}

	next := cloneServer(server)

	role := next.Roles[event.RoleID]
	if role == nil {
		// Role was created
		role = &ServerRole{ID: event.RoleID}

		if next.Roles == nil {
			next.Roles = make(map[string]*ServerRole, 1)
		}

		next.Roles[event.RoleID] = role
	}

	role.update(event.Data)
	role.clear(event.Clear)

	s.servers[event.ID] = next
}

func (s *State) deleteServerRole(data *EventServerRoleDelete) {

	if !s.trackServers {
		return
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	server := s.servers[data.ID]
	if server == nil {
		return
	}

	next := cloneServer(server)
	delete(next.Roles, data.RoleID)

	s.servers[data.ID] = next
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

	// Upserting keeps us in sync if we somehow missed the join
	next := ServerMember{ID: event.ID}
	if member := s.members.get(event.ID.Server, event.ID.User); member != nil {
		next = *member
	}

	// update installs the event's own slices, so the copy is taken after it
	next.update(event.Data)
	next.clear(event.Clear)

	s.members.add(cloneMember(&next))
}

func (s *State) createChannel(event *EventChannelCreate) {

	if !s.trackChannels {
		return
	}

	s.channelsMu.Lock()
	s.channels[event.ID] = cloneChannel(&event.Channel)
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
		logf("channel %s created in unknown server %s", event.ID, *event.Server)
		return
	}

	next := cloneServer(server)
	next.Channels = sliceAppendUnique(next.Channels, event.ID)

	s.servers[*event.Server] = next
}

func (s *State) addGroupParticipant(event *EventChannelGroupJoin) {

	if !s.trackChannels {
		return
	}

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

	channel := s.channels[event.ID]
	if channel == nil {
		logf("%s joined unknown group: %s", event.User, event.ID)
		return
	}

	next := cloneChannel(channel)
	next.Recipients = sliceAppendUnique(next.Recipients, event.User)

	s.channels[event.ID] = next
}

func (s *State) removeGroupParticipant(event *EventChannelGroupLeave) {

	if !s.trackChannels {
		return
	}

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

	channel := s.channels[event.ID]
	if channel == nil {
		logf("%s left unknown group %s", event.User, event.ID)
		return
	}

	index := slices.Index(channel.Recipients, event.User)
	if index < 0 {
		return
	}

	next := cloneChannel(channel)
	next.Recipients = sliceRemoveIndex(next.Recipients, index)

	s.channels[event.ID] = next
}

func (s *State) updateChannel(event *EventChannelUpdate) {

	if !s.trackChannels {
		return
	}

	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()

	channel := s.channels[event.ID]
	if channel == nil {
		logf("unknown channel updated %s", event.ID)
		return
	}

	// update installs the event's own maps and slices, so the deep copy is taken after it
	next := *channel
	next.update(event.Data)
	next.clear(event.Clear)

	s.channels[event.ID] = cloneChannel(&next)
}

func (s *State) deleteChannel(event *EventChannelDelete) {

	if s.trackVoice {
		s.voiceMu.Lock()
		s.voice.removeChannel(event.ID)
		s.voiceMu.Unlock()
	}

	if !s.trackChannels {
		return
	}

	s.channelsMu.Lock()
	channel := s.channels[event.ID]
	if channel == nil {
		s.channelsMu.Unlock()
		logf("unknown channel deleted %s", event.ID)
		return
	}

	delete(s.channels, event.ID)
	s.channelsMu.Unlock()

	if !s.trackServers {
		return
	}

	// Reading the dropped channel unlocked is safe: it is already published, so it cannot change
	if channel.Server == nil {
		return // Channel doesn't belong to a server
	}

	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	server := s.servers[*channel.Server]
	if server == nil {
		logf("channel %s deleted from unknown server %s", event.ID, *channel.Server)
		return
	}

	index := slices.Index(server.Channels, event.ID)
	if index < 0 {
		return
	}

	next := cloneServer(server)
	next.Channels = sliceRemoveIndex(next.Channels, index)

	s.servers[*channel.Server] = next
}

func (s *State) createServer(event *EventServerCreate) {

	if event.Server == nil {
		logf("State.createServer: server %s has no server data", event.ID)
		return
	}

	if s.trackServers {
		s.serversMu.Lock()
		s.servers[event.ID] = cloneServer(event.Server)
		s.serversMu.Unlock()
	}

	// If there's something you'll be first at in life, it's being a member in your own server.
	// Only this row needs self, so an unresolved one costs the row and not the rest of the event.
	if s.trackMembers {
		if self := s.Self(); self == nil {
			logf("State.createServer: self is nil; skipping self-membership for server %s", event.ID)
		} else {
			s.membersMu.Lock()
			member := &ServerMember{
				ID: MemberCompositeID{User: self.ID, Server: event.ID},
			}

			if id, err := ulid.Parse(event.Server.ID); err == nil {
				member.JoinedAt = ulid.Time(id.Time())
			}

			s.members.add(member)
			s.membersMu.Unlock()
		}
	}

	if s.trackChannels {
		s.channelsMu.Lock()
		for _, channel := range event.Channels {
			// Need to add directly here to avoid nested lock acquisition
			if channel != nil {
				s.channels[channel.ID] = cloneChannel(channel)
			}
		}
		s.channelsMu.Unlock()
	}

	if s.trackEmojis {
		s.emojisMu.Lock()
		for _, emoji := range event.Emojis {
			// Need to add directly here to avoid nested lock acquisition
			if emoji != nil {
				s.emojis[emoji.ID] = cloneEmoji(emoji)
			}
		}
		s.emojisMu.Unlock()
	}

	if s.trackVoice && len(event.VoiceStates) > 0 {
		s.voiceMu.Lock()
		s.voice.addMany(event.VoiceStates)
		s.voiceMu.Unlock()
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
		logf("unknown server update %s", event.ID)
		return
	}

	// update installs the event's own maps and slices, so the deep copy is taken after it
	next := *server
	next.update(event.Data)
	next.clear(event.Clear)

	s.servers[event.ID] = cloneServer(&next)
}

func (s *State) deleteServer(event *EventServerDelete) {

	if s.trackServers {
		s.serversMu.Lock()
		server := s.servers[event.ID]
		delete(s.servers, event.ID)
		s.serversMu.Unlock()

		// The server we just dropped is the only way to find its channels
		if s.trackVoice && server != nil {
			s.voiceMu.Lock()
			for _, cID := range server.Channels {
				s.voice.removeChannel(cID)
			}
			s.voiceMu.Unlock()
		}
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

	next := User{ID: event.ID}

	if user := s.users[event.ID]; user != nil {
		next = *user
	} else if event.Data.Username == nil {
		// An uncached user is only worth inserting if the partial names them: a
		// presence-only update would file a record nothing can render, and the
		// gateway sends those for everyone the account can see. EvictUser is the
		// only way back out, so the bar to entry is what bounds the map.
		logf("dropping update for uncached, unnamed user %s", event.ID)
		return
	}

	// update installs the event's own slices, so the copy is taken after it
	next.update(event.Data)
	next.clear(event.Clear)

	stored := cloneUser(&next)
	s.users[event.ID] = stored

	// Self is the same record under a second pointer; both move or neither does
	if self := s.self.Load(); self != nil && self.ID == event.ID {
		s.self.Store(stored)
	}
}

// updateUserRelationship applies a relationship change. The event carries the
// whole user, so one we have never seen is inserted rather than dropped.
func (s *State) updateUserRelationship(event *EventUserRelationship) {

	if !s.trackUsers || event.User == nil {
		return
	}

	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	var stored *User

	if user := s.users[event.User.ID]; user != nil {
		next := *user
		next.Relationship = event.User.Relationship
		stored = cloneUser(&next)
	} else {
		stored = cloneUser(event.User)
	}

	s.users[stored.ID] = stored

	if self := s.self.Load(); self != nil && self.ID == stored.ID {
		s.self.Store(stored)
	}
}

func (s *State) createEmoji(event *EventEmojiCreate) {

	if !s.trackEmojis {
		return
	}

	s.emojisMu.Lock()
	defer s.emojisMu.Unlock()

	s.emojis[event.ID] = cloneEmoji(&event.Emoji)
}

func (s *State) deleteEmoji(event *EventEmojiDelete) {

	if !s.trackEmojis {
		return
	}

	s.emojisMu.Lock()
	defer s.emojisMu.Unlock()

	delete(s.emojis, event.ID)
}

func (s *State) joinVoiceChannel(event *EventVoiceChannelJoin) {

	if !s.trackVoice {
		return
	}

	s.voiceMu.Lock()
	defer s.voiceMu.Unlock()

	s.voice.add(event.ID, cloneVoiceState(&event.State))
}

func (s *State) leaveVoiceChannel(event *EventVoiceChannelLeave) {

	if !s.trackVoice {
		return
	}

	s.voiceMu.Lock()
	defer s.voiceMu.Unlock()

	s.voice.remove(event.ID, event.User)
}

// moveVoiceChannel handles a participant being moved between two channels;
// the server sends this instead of a leave/join pair.
func (s *State) moveVoiceChannel(event *EventVoiceChannelMove) {

	if !s.trackVoice {
		return
	}

	s.voiceMu.Lock()
	defer s.voiceMu.Unlock()

	s.voice.remove(event.From, event.User)
	s.voice.add(event.To, cloneVoiceState(&event.State))
}

func (s *State) updateVoiceState(event *EventUserVoiceStateUpdate) {

	if !s.trackVoice {
		return
	}

	s.voiceMu.Lock()
	defer s.voiceMu.Unlock()

	// Upserting keeps us in sync if we somehow missed the join
	next := UserVoiceState{ID: event.ID}
	if state := s.voice.get(event.ChannelID, event.ID); state != nil {
		next = *state
	}

	next.update(event.Data)

	s.voice.add(event.ChannelID, &next)
}
