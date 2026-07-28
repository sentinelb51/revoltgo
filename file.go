package revoltgo

//go:generate msgp -tests=false -io=false

type FileMetadataType string

const (
	FileMetadataTypeFile  FileMetadataType = "File"
	FileMetadataTypeText  FileMetadataType = "Text"
	FileMetadataTypeImage FileMetadataType = "Image"
	FileMetadataTypeVideo FileMetadataType = "Video"
	FileMetadataTypeAudio FileMetadataType = "Audio"
)

type File struct {
	ID string `msg:"_id" json:"_id,omitempty"`

	// Raw content type of this file
	ContentType string `msg:"content_type" json:"content_type,omitempty"`

	// Original filename
	Filename string `msg:"filename" json:"filename,omitempty"`

	// Metadata associated with file
	Metadata *AttachmentMetadata `msg:"metadata" json:"metadata,omitempty"`

	// FileParams size in bytes
	Size int `msg:"size" json:"size,omitempty"`

	// Tag (bucket) this file was uploaded to
	Tag string `msg:"tag" json:"tag,omitempty"`

	// Whether this file was deleted
	Deleted bool `msg:"deleted" json:"deleted,omitempty"`

	MessageID string `msg:"message_id" json:"message_id,omitempty"`
	ObjectID  string `msg:"object_id" json:"object_id,omitempty"`

	// Whether this file was reported
	Reported bool `msg:"reported" json:"reported,omitempty"`

	ServerID string `msg:"server_id" json:"server_id,omitempty"`
	UserID   string `msg:"user_id" json:"user_id,omitempty"`
}

func (a File) URL(size string) string {
	return EndpointAutumnFile(a.Tag, a.ID, size)
}

type AttachmentMetadata struct {
	Type FileMetadataType `msg:"type" json:"type,omitempty"`

	Width  int `msg:"width" json:"width,omitempty"`
	Height int `msg:"height" json:"height,omitempty"`
}
