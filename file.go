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
	ID string `msg:"_id" json:"_id,omitzero"`

	// Raw content type of this file
	ContentType string `msg:"content_type" json:"content_type,omitzero"`

	// Original filename
	Filename string `msg:"filename" json:"filename,omitzero"`

	// Metadata associated with file
	Metadata *AttachmentMetadata `msg:"metadata" json:"metadata,omitzero"`

	// FileParams size in bytes
	Size int `msg:"size" json:"size,omitzero"`

	// Tag (bucket) this file was uploaded to
	Tag string `msg:"tag" json:"tag,omitzero"`

	// Whether this file was deleted
	Deleted bool `msg:"deleted" json:"deleted,omitzero"`

	MessageID string `msg:"message_id" json:"message_id,omitzero"`
	ObjectID  string `msg:"object_id" json:"object_id,omitzero"`

	// Whether this file was reported
	Reported bool `msg:"reported" json:"reported,omitzero"`

	ServerID string `msg:"server_id" json:"server_id,omitzero"`
	UserID   string `msg:"user_id" json:"user_id,omitzero"`
}

func (f *File) URL(size string) string {
	return EndpointAutumnFile(f.Tag, f.ID, size)
}

type AttachmentMetadata struct {
	Type FileMetadataType `msg:"type" json:"type,omitzero"`

	Width  int `msg:"width" json:"width,omitzero"`
	Height int `msg:"height" json:"height,omitzero"`
}
