package revoltgo

import (
	"log"
	"sync"
)

type State struct {
	sync.RWMutex

	// The current user, also present in Users
	User *User

	Users    map[string]*User
	Servers  map[string]*Server
	Channels map[string]*Channel
	Members  map[string]*ServerMember
	Emojis   map[string]*Emoji
}

func newState(ready *EventReady) *State {

	state := &State{
		Users:    make(map[string]*User, len(ready.Users)),
		Servers:  make(map[string]*Server, len(ready.Servers)),
		Channels: make(map[string]*Channel, len(ready.Channels)),
		Members:  make(map[string]*ServerMember, len(ready.Members)),
		Emojis:   make(map[string]*Emoji, len(ready.Emojis)),
	}

	// The last user in the ready event is the current user
	state.User = ready.Users[len(ready.Users)-1]

	/* Populate the caches */

	for _, user := range ready.Users {
		state.Users[user.ID] = user
	}

	for _, server := range ready.Servers {
		state.Servers[server.ID] = server
	}

	for _, channel := range ready.Channels {
		state.Channels[channel.ID] = channel
	}

	for _, member := range ready.Members {
		state.Members[member.ID.String()] = member
	}

	for _, emoji := range ready.Emojis {
		state.Emojis[emoji.ID] = emoji
	}

	return state
}

func (s *State) platformWipe(event *EventUserPlatformWipe) {
	delete(s.Users, event.UserID)
}

func (s *State) updateRole(data *EventServerRoleUpdate) {

	s.Lock()
	defer s.Unlock()

	server, exists := s.Servers[data.ID]
	if !exists {
		log.Printf("update for role %s in unknown server %s\n", data.RoleID, data.ID)
		return
	}

	role, exists := server.Roles[data.RoleID]
	if !exists {
		server.Roles[data.RoleID] = data.Data
		return
	}

	role = merge[*ServerRole](role, data.Data)
	for _, field := range data.Clear {
		clearByJSON(role, field)
	}
}

func (s *State) deleteRole(data *EventServerRoleDelete) {

	s.Lock()
	defer s.Unlock()

	server, exists := s.Servers[data.ID]
	if exists {
		delete(server.Roles, data.RoleID)
	}
}

func (s *State) createServerMember(data *EventServerMemberJoin) {

	s.Lock()
	defer s.Unlock()

	id := MemberCompoundID{User: data.User, Server: data.ID}
	s.Members[id.String()] = &ServerMember{ID: id}
}

func (s *State) deleteServerMember(data *EventServerMemberLeave) {

	s.Lock()
	defer s.Unlock()

	id := MemberCompoundID{User: data.User, Server: data.ID}
	delete(s.Members, id.String())
}

func (s *State) updateServerMember(data *EventServerMemberUpdate) {

	s.Lock()
	defer s.Unlock()

	member, exists := s.Members[data.ID.String()]
	if !exists {
		data.Data.ID = data.ID
		s.Members[data.ID.String()] = data.Data
		return
	}

	member = merge[*ServerMember](member, data.Data)
	for _, field := range data.Clear {
		clearByJSON(member, field)
	}
}

func (s *State) createChannel(event *EventChannelCreate) {
	s.Lock()
	defer s.Unlock()

	server, exists := s.Servers[event.Server]
	if !exists {
		return
	}

	s.Channels[event.ID] = &Channel{
		ID:          event.ID,
		Server:      event.Server,
		ChannelType: event.ChannelType,
		Name:        event.Name,
	}

	server.Channels = append(server.Channels, event.ID)
}

func (s *State) updateChannel(event *EventChannelUpdate) {
	s.Lock()
	defer s.Unlock()

	server, exist := s.Servers[event.ID]
	if !exist {
		event.Data.ID = event.ID
		s.Channels[event.ID] = event.Data
		return
	}

	server = merge[*Server](server, event.Data)
	for _, field := range event.Clear {
		clearByJSON(server, field)
	}
}

func (s *State) deleteChannel(event *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()

	channel, exists := s.Channels[event.ID]
	if exists {
		delete(s.Channels, event.ID)
	}

	server, exists := s.Servers[channel.Server]
	if exists {
		for i, cID := range server.Channels {
			if cID == event.ID {
				server.Channels = sliceRemoveIndex(server.Channels, i)
				return
			}
		}
	}
}

func (s *State) createServer(event *EventServerCreate) {
	s.Lock()
	defer s.Unlock()

	s.Servers[event.ID] = event.Server
}

func (s *State) updateServer(event *EventServerUpdate) {
	s.Lock()
	defer s.Unlock()

	server, exists := s.Servers[event.ID]
	if !exists {
		event.Data.ID = event.ID
		s.Servers[event.ID] = event.Data
		return
	}

	server = merge[*Server](server, event.Data)
	for _, field := range event.Clear {
		clearByJSON(server, field)
	}
}

func (s *State) deleteServer(event *EventServerDelete) {
	s.Lock()
	defer s.Unlock()

	delete(s.Servers, event.ID)
}

func (s *State) updateUsers(event any) {

	s.Lock()
	defer s.Unlock()

	switch data := event.(type) {
	case *EventUserUpdate:
		if value, exists := s.Users[data.ID]; exists {
			value = merge[*User](value, data.Data)
		} else {
			data.Data.ID = data.ID
			s.Users[data.ID] = data.Data
		}
	default:
		panic("cannot process this event type")
	}
}

func (s *State) createEmoji(event *EventEmojiCreate) {
	s.Lock()
	defer s.Unlock()

	s.Emojis[event.ID] = event.Emoji
}

func (s *State) deleteEmoji(event *EventEmojiDelete) {
	s.Lock()
	defer s.Unlock()

	delete(s.Emojis, event.ID)
}
