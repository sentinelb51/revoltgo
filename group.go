package revoltgo

// Group channel struct.
type Group struct {
	ID          string   `json:"_id"`
	Nonce       string   `json:"nonce"`
	OwnerID     string   `json:"owner"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Users       []string `json:"users"`
}

// Fetched group members struct.
type FetchedGroupMembers struct {
	Messages []*Message `json:"messages"`
	Users    []*User    `json:"users"`
}

// System messages struct.
type GroupSystemMessages struct {
	UserJoined string `json:"user_joined,omitempty"`
	UserLeft   string `json:"user_left,omitempty"`
}
