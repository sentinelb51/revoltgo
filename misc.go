package revoltgo

type CompositeChannelID struct {
	Channel string `json:"channel"`
	User    string `json:"user"`
}

type SyncUnread struct {
	ID       CompositeChannelID `json:"_id"`
	LastID   string             `json:"last_id"`
	Mentions []string           `json:"mentions"`
}

type SyncSettingsData map[string]UpdateTuple

type SyncSettingsFetchData struct {
	Keys []string `json:"keys"`
}
