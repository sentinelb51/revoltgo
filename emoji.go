package revoltgo

//go:generate msgp -tests=false -io=false

type Emoji struct {
	ID        string       `msg:"_id" json:"_id,omitempty"`
	Parent    *EmojiParent `msg:"parent" json:"parent,omitempty"`
	CreatorID string       `msg:"creator_id" json:"creator_id,omitempty"`
	Name      string       `msg:"name" json:"name,omitempty"`
	Animated  bool         `msg:"animated" json:"animated,omitempty"`
	NSFW      bool         `msg:"nsfw" json:"nsfw,omitempty"`
}

type EmojiParent struct {
	Type string `msg:"type" json:"type,omitempty"`
	ID   string `msg:"id" json:"id,omitempty"`
}
