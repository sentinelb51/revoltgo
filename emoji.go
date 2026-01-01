package revoltgo

//go:generate msgp -tests=false -io=false

type Emoji struct {
	ID        string       `msg:"_id"`
	Parent    *EmojiParent `msg:"parent"`
	CreatorID string       `msg:"creator_id"`
	Name      string       `msg:"name"`
	Animated  bool         `msg:"animated"`
	NSFW      bool         `msg:"nsfw"`
}

type EmojiParent struct {
	Type string `msg:"type"`
	ID   string `msg:"id"`
}
