package revoltgo

import "io"

type File struct {
	// The name of the file; this is completely arbitrary because the backend determines the file-type anyway
	// However, it should not be empty, otherwise the media will not load on the client
	Name string

	// The contents of the file to be read when uploading
	Reader io.Reader
}

// FileAttachment is the response from the API when uploading a file.
// To upload a file, you must reference this ID in MessageSend.Attachments.
type FileAttachment struct {
	ID string `msg:"id" json:"id,omitempty"`
}
