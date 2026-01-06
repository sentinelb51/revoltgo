package revoltgo

import (
	"bytes"
	"fmt"
)

//go:generate msgp -tests=false -io=false

const jsonSkipAheadKeyType = len(`{"type":"`)

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
func eventTypeFromMSGP(data []byte) (string, error) {

	start := 6 // skip map header and "type" key
	header := data[start]

	if header < 0xA0 || header > 0xBF {
		return "", fmt.Errorf("expected fixstr, got byte 0x%X", header)
	}

	start++
	size := int(header & 0x1F)
	end := start + size

	if end > len(data) {
		return "", fmt.Errorf("given size %d exceeds data length %d", size, len(data))
	}

	return string(data[start:end]), nil
}

type Event struct {
	Type string `msg:"type" json:"type,omitempty"`
}

func (e *Event) String() string {
	return e.Type
}

var eventConstructors = map[string]func() any{
	"Error":         func() any { return new(EventError) },
	"Bulk":          func() any { return new(EventBulk) },
	"Pong":          func() any { return new(EventPong) },
	"Ready":         func() any { return new(EventReady) },
	"Auth":          func() any { return new(EventAuth) },
	"Authenticated": func() any { return new(EventAuthenticated) },

	"Message":               func() any { return new(EventMessage) },
	"MessageAppend":         func() any { return new(EventMessageAppend) },
	"MessageDelete":         func() any { return new(EventMessageDelete) },
	"BulkMessageDelete":     func() any { return new(EventBulkMessageDelete) },
	"MessageReact":          func() any { return new(EventMessageReact) },
	"MessageUnreact":        func() any { return new(EventMessageUnreact) },
	"MessageRemoveReaction": func() any { return new(EventMessageRemoveReaction) },
	"MessageUpdate":         func() any { return new(EventMessageUpdate) },

	"ChannelCreate":      func() any { return new(EventChannelCreate) },
	"ChannelDelete":      func() any { return new(EventChannelDelete) },
	"ChannelAck":         func() any { return new(EventChannelAck) },
	"ChannelStartTyping": func() any { return new(EventChannelStartTyping) },
	"ChannelStopTyping":  func() any { return new(EventChannelStopTyping) },
	"ChannelGroupJoin":   func() any { return new(EventChannelGroupJoin) },
	"ChannelUpdate":      func() any { return new(EventChannelUpdate) },
	"ChannelGroupLeave":  func() any { return new(EventChannelGroupLeave) },

	"ServerCreate":          func() any { return new(EventServerCreate) },
	"ServerDelete":          func() any { return new(EventServerDelete) },
	"ServerUpdate":          func() any { return new(EventServerUpdate) },
	"ServerRoleDelete":      func() any { return new(EventServerRoleDelete) },
	"ServerRoleUpdate":      func() any { return new(EventServerRoleUpdate) },
	"ServerRoleRanksUpdate": func() any { return new(EventServerRoleRanksUpdate) },
	"ServerMemberJoin":      func() any { return new(EventServerMemberJoin) },
	"ServerMemberLeave":     func() any { return new(EventServerMemberLeave) },
	"ServerMemberUpdate":    func() any { return new(EventServerMemberUpdate) },

	"EmojiCreate": func() any { return new(EventEmojiCreate) },
	"EmojiDelete": func() any { return new(EventEmojiDelete) },

	"UserSettingsUpdate": func() any { return new(EventUserSettingsUpdate) },
	"UserRelationship":   func() any { return new(EventUserRelationship) },
	"UserPlatformWipe":   func() any { return new(EventUserPlatformWipe) },
	"UserUpdate":         func() any { return new(EventUserUpdate) },

	"WebhookCreate": func() any { return new(EventWebhookCreate) },
	"WebhookDelete": func() any { return new(EventWebhookDelete) },
	"WebhookUpdate": func() any { return new(EventWebhookUpdate) },

	"VoiceChannelJoin":     func() any { return new(EventVoiceChannelJoin) },
	"VoiceChannelLeave":    func() any { return new(EventVoiceChannelLeave) },
	"VoiceChannelMove":     func() any { return new(EventVoiceChannelMove) },
	"UserVoiceStateUpdate": func() any { return new(EventUserVoiceStateUpdate) },
	"UserMoveVoiceChannel": func() any { return new(EventUserMoveVoiceChannel) },

	"ReportCreate": func() any { return new(EventReportCreate) },
}
