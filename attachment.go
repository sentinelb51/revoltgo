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
	ID string `msg:"_id" json:"_id,omitempty"`

	// Tag / bucket this file was uploaded to
	Tag string `msg:"tag" json:"tag,omitempty"`

	// Original filename
	Filename string `msg:"filename" json:"filename,omitempty"`

	// Metadata associated with file
	Metadata *AttachmentMetadata `msg:"metadata" json:"metadata,omitempty"`

	// Raw content type of this file
	ContentType string `msg:"content_type" json:"content_type,omitempty"`

	// Size of this file (in bytes)
	Size int `msg:"size" json:"size,omitempty"`

	// Whether this file was deleted
	Deleted bool `msg:"deleted" json:"deleted,omitempty"`

	// Whether this file was reported
	Reported bool `msg:"reported" json:"reported,omitempty"`

	MessageID string `msg:"message_id" json:"message_id,omitempty"`
	UserID    string `msg:"user_id" json:"user_id,omitempty"`
	ServerID  string `msg:"server_id" json:"server_id,omitempty"`
	ObjectID  string `msg:"object_id" json:"object_id,omitempty"`
}

func (a Attachment) URL(size string) string {
	return EndpointAutumnFile(a.Tag, a.ID, size)
}

type AttachmentMetadata struct {
	Type AttachmentMetadataType `msg:"type" json:"type,omitempty"`

	Width  int `msg:"width" json:"width,omitempty"`
	Height int `msg:"height" json:"height,omitempty"`
}
