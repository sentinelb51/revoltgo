package revoltgo

//go:generate msgp -tests=false -io=false

type AttachmentMetadataType string

const (
	AttachmentMetadataTypeFile  AttachmentMetadataType = "File"
	AttachmentMetadataTypeText  AttachmentMetadataType = "Text"
	AttachmentMetadataTypeImage AttachmentMetadataType = "Image"
	AttachmentMetadataTypeVideo AttachmentMetadataType = "Video"
	AttachmentMetadataTypeAudio AttachmentMetadataType = "Audio"
)

type Attachment struct {
	ID string `msg:"_id"`

	// Tag / bucket this file was uploaded to
	Tag string `msg:"tag"`

	// Original filename
	Filename string `msg:"filename"`

	// Metadata associated with file
	Metadata *AttachmentMetadata `msg:"metadata"`

	// Raw content type of this file
	ContentType string `msg:"content_type"`

	// Size of this file (in bytes)
	Size int `msg:"size"`

	// Whether this file was deleted
	Deleted bool `msg:"deleted"`

	// Whether this file was reported
	Reported bool `msg:"reported"`

	MessageID string `msg:"message_id"`
	UserID    string `msg:"user_id"`
	ServerID  string `msg:"server_id"`
	ObjectID  string `msg:"object_id"`
}

func (a Attachment) URL(size string) string {
	return EndpointAutumnFile(a.Tag, a.ID, size)
}

type AttachmentMetadata struct {
	Type AttachmentMetadataType `msg:"type"`

	Width  int `msg:"width"`
	Height int `msg:"height"`
}
