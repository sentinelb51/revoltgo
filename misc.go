package revoltgo

// todo: maybe move CompositeChannelID to channel.go?

type CompositeChannelID struct {
	Channel string `json:"channel"`
	User    string `json:"user"`
}

type SyncUnread struct {
	ID       CompositeChannelID `json:"_id"`
	LastID   string             `json:"last_id"`
	Mentions []string           `json:"mentions"`
}
