package revoltgo

//go:generate msgp -tests=false -io=false

type Group struct {
	ID          string   `msg:"_id"`
	OwnerID     string   `msg:"owner"`
	Name        string   `msg:"name"`
	Description string   `msg:"description,omitempty"`
	Users       []string `msg:"users"`
}

type FetchedGroupMembers struct {
	Messages []*Message `msg:"messages"`
	Users    []*User    `msg:"users"`
}

type GroupSystemMessages struct {
	UserJoined string `msg:"user_joined,omitempty"`
	UserLeft   string `msg:"user_left,omitempty"`
}
