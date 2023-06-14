package revoltgo

// Emoji struct.
type Emoji struct {
	ID       string      `json:"_id"`
	Name     string      `json:"name"`
	Animated bool        `json:"animated"`
	NSFW     bool        `json:"nsfw,omitempty"`
	Parent   EmojiParent `json:"parent"`
}

type EmojiParent struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
