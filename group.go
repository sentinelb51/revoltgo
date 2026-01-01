package revoltgo

//go:generate msgp -tests=false -io=false

type Group struct {
	ID          string   `msg:"_id" json:"_id,omitempty"`
	OwnerID     string   `msg:"owner" json:"owner,omitempty"`
	Name        string   `msg:"name" json:"name,omitempty"`
	Description string   `msg:"description" json:"description,omitempty"`
	Users       []string `msg:"users" json:"users,omitempty"`
}

type FetchedGroupMembers struct {
	Messages []*Message `msg:"messages" json:"messages,omitempty"`
	Users    []*User    `msg:"users" json:"users,omitempty"`
}

type GroupSystemMessages struct {
	UserJoined string `msg:"user_joined" json:"user_joined,omitempty"`
	UserLeft   string `msg:"user_left" json:"user_left,omitempty"`
}
