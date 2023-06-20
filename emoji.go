package revoltgo

type Emoji struct {
	ID        string       `json:"_id"`
	Parent    *EmojiParent `json:"parent"`
	CreatorID string       `json:"creator_id"`
	Name      string       `json:"name"`
	Animated  bool         `json:"animated"`
	NSFW      bool         `json:"nsfw"`
}

type EmojiParent struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
