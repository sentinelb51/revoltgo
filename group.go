package revoltgo

//go:generate msgp -tests=false -io=false

type Group struct {
	ID          string   `msg:"_id" json:"_id,omitzero"`
	OwnerID     string   `msg:"owner" json:"owner,omitzero"`
	Name        string   `msg:"name" json:"name,omitzero"`
	Description string   `msg:"description" json:"description,omitzero"`
	Users       []string `msg:"users" json:"users,omitzero"`
}

type FetchedGroupMembers struct {
	Messages []*Message `msg:"messages" json:"messages,omitzero"`
	Users    []*User    `msg:"users" json:"users,omitzero"`
}

type GroupSystemMessages struct {
	UserJoined string `msg:"user_joined" json:"user_joined,omitzero"`
	UserLeft   string `msg:"user_left" json:"user_left,omitzero"`
}
