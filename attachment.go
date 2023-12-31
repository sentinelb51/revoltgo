package revoltgo

type AttachmentMetadataType string

const (
	AttachmentMetadataTypeFile  AttachmentMetadataType = "File"
	AttachmentMetadataTypeText  AttachmentMetadataType = "Text"
	AttachmentMetadataTypeImage AttachmentMetadataType = "Image"
	AttachmentMetadataTypeVideo AttachmentMetadataType = "Video"
	AttachmentMetadataTypeAudio AttachmentMetadataType = "Audio"
)

type Attachment struct {
	ID string `json:"_id"`

	// Tag / bucket this file was uploaded to
	Tag string `json:"tag"`

	// Original filename
	FileName string `json:"filename"`

	// Metadata associated with file
	Metadata *AttachmentMetadata `json:"metadata"`

	// Raw content type of this file
	ContentType string `json:"content_type"`

	// Size of this file (in bytes)
	Size int `json:"size"`

	// Whether this file was deleted
	Deleted bool `json:"deleted"`

	// Whether this file was reported
	Reported bool `json:"reported"`

	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
	ServerID  string `json:"server_id"`
	ObjectID  string `json:"object_id"`
}

// URL is a helper function that returns the URL of the attachment.
// [!] Make sure attachment is not nil before calling this function.
func (a *Attachment) URL(size string) string {
	return EndpointIcon(a.ID, size)
}

type AttachmentMetadata struct {
	Type AttachmentMetadataType `json:"type"`

	Width  int `json:"width"`
	Height int `json:"height"`
}
