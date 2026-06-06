package revoltgo

import (
	"bytes"
	"fmt"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false

const (
	jsonSkipAheadKeyType = len(`{"type":"`)
	msgpTypeValueOffset  = len(`"type"`)
)

// eventTypeFromJSON uses heuristics to quickly extract the event type from JSON data
func eventTypeFromJSON(data []byte) (string, error) {
	closingTagIndex := bytes.IndexByte(data[jsonSkipAheadKeyType:], '"')
	if closingTagIndex < 0 {
		return "", fmt.Errorf("closing quote of type field not found")
	}

	result := data[jsonSkipAheadKeyType : jsonSkipAheadKeyType+closingTagIndex]
	return string(result), nil
}

// eventTypeFromMSGP uses heuristics to quickly extract the event type from MessagePack data.
func eventTypeFromMSGP(data []byte) ([]byte, error) {
	if len(data) <= msgpTypeValueOffset {
		return nil, fmt.Errorf("data too short: %d bytes", len(data))
	}

	header := data[msgpTypeValueOffset]
	if header < 0xA0 || header > 0xBF {
		return nil, fmt.Errorf("expected fixstr, got byte 0x%X", header)
	}

	start := msgpTypeValueOffset + 1
	end := start + int(header&0x1F)
	if end > len(data) {
		return nil, fmt.Errorf("size %d exceeds data length %d", end-start, len(data))
	}

	return data[start:end], nil
}

type Event struct {
	Type string `msg:"type" json:"type,omitempty"`
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
