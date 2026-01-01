package revoltgo

//go:generate msgp -tests=false -io=false
// todo: maybe move CompositeChannelID to channel.go?

type CompositeChannelID struct {
	Channel string `msg:"channel"`
	User    string `msg:"user"`
}

type SyncUnread struct {
	ID       CompositeChannelID `msg:"_id"`
	LastID   string             `msg:"last_id"`
	Mentions []string           `msg:"mentions"`
}
