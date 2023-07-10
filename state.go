package revoltgo

import (
	"log"
	"sync"
)

type State struct {
	sync.RWMutex

	// The current user's ID
	User string

	Users    map[string]*User
	Servers  map[string]*Server
	Channels map[string]*Channel
	Members  map[string]*ServerMember
	Emojis   map[string]*Emoji
}

func newState(ready *EventReady) *State {

	state := &State{
		User:     ready.Users[0].ID, // The first user is always us
		Users:    make(map[string]*User, len(ready.Users)),
		Servers:  make(map[string]*Server, len(ready.Servers)),
		Channels: make(map[string]*Channel, len(ready.Channels)),
		Members:  make(map[string]*ServerMember, len(ready.Members)),
		Emojis:   make(map[string]*Emoji, len(ready.Emojis)),
	}

	// Populate the caches

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

// updateRoles updates the state's roles cache
func (s *State) updateRoles(event any) {

	s.Lock()
	defer s.Unlock()

	switch data := event.(type) {
	case *EventServerRoleUpdate:
		server, exists := s.Servers[data.ID]
		if !exists {
			log.Printf("received update for role %s in unknown server %s\n", data.RoleID, data.ID)
			return
		}

		if value, exists := server.Roles[data.RoleID]; exists {
			value = merge(value, data.Data).(*ServerRole)
			for _, field := range data.Clear {
				clear(value, field)
			}
		} else {
			server.Roles[data.RoleID] = data.Data
		}
	case *EventServerRoleDelete:
		server, exists := s.Servers[data.ID]
		if exists {
			delete(server.Roles, data.RoleID)
		}
	default:
		panic("cannot process this event type")
	}
}

// updateMembers updates the state's members cache
func (s *State) updateMembers(event any) {

	s.Lock()
	defer s.Unlock()

	switch data := event.(type) {
	case *EventServerMemberJoin:
		id := MemberCompoundID{User: data.User, Server: data.ID}
		s.Members[id.String()] = &ServerMember{ID: id}
	case *EventServerMemberUpdate:
		if value, exists := s.Members[data.ID.String()]; exists {
			value = merge(value, data.Data).(*ServerMember)
			for _, field := range data.Clear {
				clear(value, field)
			}
		} else {
			data.Data.ID = data.ID
			s.Members[data.ID.String()] = data.Data
		}
	case *EventServerMemberLeave:
		id := MemberCompoundID{User: data.User, Server: data.ID}
		delete(s.Members, id.String())
	default:
		panic("cannot process this event type")
	}
}

func (s *State) updateChannels(event any) {

	s.Lock()
	defer s.Unlock()

	switch data := event.(type) {
	case *EventChannelCreate:
		s.Channels[data.ID] = &Channel{
			ID:          data.ID,
			Server:      data.Server,
			ChannelType: data.ChannelType,
			Name:        data.Name,
		}

		server, exists := s.Servers[data.Server]
		if exists {
			server.Channels = append(server.Channels, data.ID)
		}
	case *EventChannelUpdate:
		if value, exists := s.Channels[data.ID]; exists {
			value = merge(value, data.Data).(*Channel)
			for _, field := range data.Clear {
				clear(value, field)
			}
		} else {
			data.Data.ID = data.ID
			s.Channels[data.ID] = data.Data
		}
	case *EventChannelDelete:
		channel, exists := s.Channels[data.ID]
		if !exists {
			return
		}

		server, exists := s.Servers[channel.Server]
		delete(s.Channels, data.ID)

		if !exists {
			return
		}

		for i, cID := range server.Channels {
			if cID == data.ID {
				server.Channels = sliceRemoveIndex(server.Channels, i)
				return
			}
		}
	default:
		panic("cannot process this event type")
	}
}

func (s *State) updateServers(event any) {

	s.Lock()
	defer s.Unlock()

	switch data := event.(type) {
	case *EventServerCreate:
		s.Servers[data.ID] = data.Server
	case *EventServerUpdate:
		if value, exists := s.Servers[data.ID]; exists {
			value = merge(value, data.Data).(*Server)
			for _, field := range data.Clear {
				clear(value, field)
			}
		} else {
			data.Data.ID = data.ID
			s.Servers[data.ID] = data.Data
		}
	case *EventServerDelete:
		delete(s.Servers, data.ID)
	default:
		panic("cannot process this event type")
	}
}

func (s *State) updateUsers(event any) {

	s.Lock()
	defer s.Unlock()

	switch data := event.(type) {
	case *EventUserUpdate:
		if value, exists := s.Users[data.ID]; exists {
			value = merge(value, data.Data).(*User)
		} else {
			data.Data.ID = data.ID
			s.Users[data.ID] = data.Data
		}
	default:
		panic("cannot process this event type")
	}
}

func (s *State) updateEmojis(event any) {

	s.Lock()
	defer s.Unlock()

	switch data := event.(type) {
	case *EventEmojiCreate:
		s.Emojis[data.ID] = data.Emoji
	case *EventEmojiDelete:
		delete(s.Emojis, data.ID)
	default:
		panic("cannot process this event type")
	}
}
