package revoltgo

//go:generate msgp -tests=false -io=false

type Emoji struct {
	ID        string       `msg:"_id" json:"_id,omitzero"`
	Parent    *EmojiParent `msg:"parent" json:"parent,omitzero"`
	CreatorID string       `msg:"creator_id" json:"creator_id,omitzero"`
	Name      string       `msg:"name" json:"name,omitzero"`
	Animated  bool         `msg:"animated" json:"animated,omitzero"`
	NSFW      bool         `msg:"nsfw" json:"nsfw,omitzero"`
}

type EmojiParent struct {
	Type string `msg:"type" json:"type,omitzero"`
	ID   string `msg:"id" json:"id,omitzero"`
}
