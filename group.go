package revoltgo

type Group struct {
	ID          string   `json:"_id"`
	OwnerID     string   `json:"owner"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Users       []string `json:"users"`
}

type FetchedGroupMembers struct {
	Messages []*Message `json:"messages"`
	Users    []*User    `json:"users"`
}

type GroupSystemMessages struct {
	UserJoined string `json:"user_joined,omitempty"`
	UserLeft   string `json:"user_left,omitempty"`
}
