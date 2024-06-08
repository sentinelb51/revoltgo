package revoltgo

func aeuConstructor() any {
	return new(AbstractEventUpdate)
}

var eventToStruct = map[string]func() any{
	"Error": func() any { return new(EventError) },
	"Bulk":  func() any { return new(EventBulk) },

	"Authenticated": func() any { return new(EventAuthenticated) },
	"Ready":         func() any { return new(EventReady) },
	"Pong":          func() any { return new(EventPong) },
	"Auth":          func() any { return new(EventAuth) },

	/* All update events are abstracted away. */
	"MessageUpdate":      aeuConstructor,
	"ServerUpdate":       aeuConstructor,
	"ChannelUpdate":      aeuConstructor,
	"ServerRoleUpdate":   aeuConstructor,
	"WebhookUpdate":      aeuConstructor,
	"UserUpdate":         aeuConstructor,
	"ServerMemberUpdate": aeuConstructor,

	"Message":        func() any { return new(EventMessage) },
	"MessageAppend":  func() any { return new(EventMessageAppend) },
	"MessageDelete":  func() any { return new(EventMessageDelete) },
	"MessageReact":   func() any { return new(EventMessageReact) },
	"MessageUnreact": func() any { return new(EventMessageUnreact) },

	"ChannelCreate":      func() any { return new(EventChannelCreate) },
	"ChannelDelete":      func() any { return new(EventChannelDelete) },
	"ChannelAck":         func() any { return new(EventChannelAck) },
	"ChannelStartTyping": func() any { return new(EventChannelStartTyping) },
	"ChannelStopTyping":  func() any { return new(EventChannelStopTyping) },

	"GroupJoin":  func() any { return new(EventGroupJoin) },
	"GroupLeave": func() any { return new(EventGroupLeave) },

	"ServerCreate":      func() any { return new(EventServerCreate) },
	"ServerDelete":      func() any { return new(EventServerDelete) },
	"ServerRoleDelete":  func() any { return new(EventServerRoleDelete) },
	"ServerMemberJoin":  func() any { return new(EventServerMemberJoin) },
	"ServerMemberLeave": func() any { return new(EventServerMemberLeave) },

	"EmojiCreate": func() any { return new(EventEmojiCreate) },
	"EmojiDelete": func() any { return new(EventEmojiDelete) },

	"UserSettingsUpdate": func() any { return new(EventUserSettingsUpdate) },
	"UserRelationship":   func() any { return new(EventUserRelationship) },
	"UserPlatformWipe":   func() any { return new(EventUserPlatformWipe) },

	"WebhookCreate": func() any { return new(EventWebhookCreate) },
	"WebhookDelete": func() any { return new(EventWebhookDelete) },
}
