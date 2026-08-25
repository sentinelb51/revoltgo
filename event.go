package revoltgo

import (
	"bytes"
	"fmt"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false

// eventKeyType is the map key naming an event's variant.
var eventKeyType = []byte("type")

// eventTypeFromMSGP extracts the event type from a MessagePack map without
// decoding the frame. Neither the map's size nor its key order is fixed: a
// payload with 16 or more pairs takes a map16 header, and the encoder is free
// to place "type" anywhere, so the offset-based read is only a fast path and
// the walk behind it is what has to be correct. The returned slice aliases data.
func eventTypeFromMSGP(data []byte) ([]byte, error) {
	if value, ok := eventTypeFixmap(data); ok {
		return value, nil
	}

	pairs, bts, err := msgp.ReadMapHeaderBytes(data)
	if err != nil {
		return nil, err
	}

	for range pairs {
		var key []byte

		if key, bts, err = msgp.ReadMapKeyZC(bts); err != nil {
			return nil, err
		}

		if !bytes.Equal(key, eventKeyType) {
			if bts, err = msgp.Skip(bts); err != nil {
				return nil, err
			}
			continue
		}

		value, _, err := msgp.ReadStringZC(bts)
		if err != nil {
			return nil, err
		}

		return value, nil
	}

	return nil, fmt.Errorf("no type field in map of %d pairs", pairs)
}

// eventTypeFixmap reads the shape the gateway actually sends: a fixmap whose
// first key is "type" and whose value is a fixstr. Anything else reports false
// and is left to the walk.
func eventTypeFixmap(data []byte) ([]byte, bool) {
	const valueAt = 1 + 1 + len("type") // fixmap header, 0xa4, the key

	if len(data) <= valueAt {
		return nil, false
	}

	if data[0]&0xF0 != 0x80 || data[1] != 0xA4 || !bytes.Equal(data[2:valueAt], eventKeyType) {
		return nil, false
	}

	header := data[valueAt]
	if header < 0xA0 || header > 0xBF {
		return nil, false
	}

	start := valueAt + 1
	end := start + int(header&0x1F)
	if end > len(data) {
		return nil, false
	}

	return data[start:end], true
}

type Event struct {
	Type string `msg:"type" json:"type,omitzero"`
}

func (e *Event) String() string {
	return e.Type
}

var eventConstructors = map[string]func() msgp.Unmarshaler{
	"Error":         func() msgp.Unmarshaler { return new(EventError) },
	"Bulk":          func() msgp.Unmarshaler { return new(EventBulk) },
	"Pong":          func() msgp.Unmarshaler { return new(EventPong) },
	"Ready":         func() msgp.Unmarshaler { return new(EventReady) },
	"Auth":          func() msgp.Unmarshaler { return new(EventAuth) },
	"Authenticated": func() msgp.Unmarshaler { return new(EventAuthenticated) },
	"Logout":        func() msgp.Unmarshaler { return new(EventLogout) },

	"Message":               func() msgp.Unmarshaler { return new(EventMessage) },
	"MessageAppend":         func() msgp.Unmarshaler { return new(EventMessageAppend) },
	"MessageDelete":         func() msgp.Unmarshaler { return new(EventMessageDelete) },
	"BulkMessageDelete":     func() msgp.Unmarshaler { return new(EventBulkMessageDelete) },
	"MessageReact":          func() msgp.Unmarshaler { return new(EventMessageReact) },
	"MessageUnreact":        func() msgp.Unmarshaler { return new(EventMessageUnreact) },
	"MessageRemoveReaction": func() msgp.Unmarshaler { return new(EventMessageRemoveReaction) },
	"MessageUpdate":         func() msgp.Unmarshaler { return new(EventMessageUpdate) },

	"ChannelCreate":      func() msgp.Unmarshaler { return new(EventChannelCreate) },
	"ChannelDelete":      func() msgp.Unmarshaler { return new(EventChannelDelete) },
	"ChannelAck":         func() msgp.Unmarshaler { return new(EventChannelAck) },
	"ChannelStartTyping": func() msgp.Unmarshaler { return new(EventChannelStartTyping) },
	"ChannelStopTyping":  func() msgp.Unmarshaler { return new(EventChannelStopTyping) },
	"ChannelGroupJoin":   func() msgp.Unmarshaler { return new(EventChannelGroupJoin) },
	"ChannelUpdate":      func() msgp.Unmarshaler { return new(EventChannelUpdate) },
	"ChannelGroupLeave":  func() msgp.Unmarshaler { return new(EventChannelGroupLeave) },

	"ServerCreate":          func() msgp.Unmarshaler { return new(EventServerCreate) },
	"ServerDelete":          func() msgp.Unmarshaler { return new(EventServerDelete) },
	"ServerUpdate":          func() msgp.Unmarshaler { return new(EventServerUpdate) },
	"ServerRoleDelete":      func() msgp.Unmarshaler { return new(EventServerRoleDelete) },
	"ServerRoleUpdate":      func() msgp.Unmarshaler { return new(EventServerRoleUpdate) },
	"ServerRoleRanksUpdate": func() msgp.Unmarshaler { return new(EventServerRoleRanksUpdate) },
	"ServerMemberJoin":      func() msgp.Unmarshaler { return new(EventServerMemberJoin) },
	"ServerMemberLeave":     func() msgp.Unmarshaler { return new(EventServerMemberLeave) },
	"ServerMemberUpdate":    func() msgp.Unmarshaler { return new(EventServerMemberUpdate) },

	"EmojiCreate": func() msgp.Unmarshaler { return new(EventEmojiCreate) },
	"EmojiDelete": func() msgp.Unmarshaler { return new(EventEmojiDelete) },

	"UserSettingsUpdate": func() msgp.Unmarshaler { return new(EventUserSettingsUpdate) },
	"UserRelationship":   func() msgp.Unmarshaler { return new(EventUserRelationship) },
	"UserPlatformWipe":   func() msgp.Unmarshaler { return new(EventUserPlatformWipe) },
	"UserUpdate":         func() msgp.Unmarshaler { return new(EventUserUpdate) },

	"WebhookCreate": func() msgp.Unmarshaler { return new(EventWebhookCreate) },
	"WebhookDelete": func() msgp.Unmarshaler { return new(EventWebhookDelete) },
	"WebhookUpdate": func() msgp.Unmarshaler { return new(EventWebhookUpdate) },

	"VoiceChannelJoin":     func() msgp.Unmarshaler { return new(EventVoiceChannelJoin) },
	"VoiceChannelLeave":    func() msgp.Unmarshaler { return new(EventVoiceChannelLeave) },
	"VoiceChannelMove":     func() msgp.Unmarshaler { return new(EventVoiceChannelMove) },
	"UserVoiceStateUpdate": func() msgp.Unmarshaler { return new(EventUserVoiceStateUpdate) },
	"UserMoveVoiceChannel": func() msgp.Unmarshaler { return new(EventUserMoveVoiceChannel) },

	"ReportCreate": func() msgp.Unmarshaler { return new(EventReportCreate) },
}
