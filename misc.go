package revoltgo

//go:generate msgp -tests=false -io=false
// todo: maybe move CompositeChannelID to channel.go?

type CompositeChannelID struct {
	Channel string `msg:"channel" json:"channel,omitempty"`
	User    string `msg:"user" json:"user,omitempty"`
}

type SyncUnread struct {
	ID       CompositeChannelID `msg:"_id" json:"_id,omitempty"`
	LastID   string             `msg:"last_id" json:"last_id,omitempty"`
	Mentions []string           `msg:"mentions" json:"mentions,omitempty"`
}
