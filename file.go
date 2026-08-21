package revoltgo

//go:generate msgp -tests=false -io=false

import (
	"log"
	"sync"
)

// FileTag is the Autumn bucket a file lives in. A file is looked up by its ID
// *and* its tag at the moment it is used, so the tag is half of what identifies
// one: an ID uploaded to attachments and then offered as an avatar names nothing
// the backend can find.
//
// The set below is the whole of it. Autumn defines the buckets, with a size
// limit and preview dimensions for each, in
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/config/Revolt.toml
// ([files.preview] and the per-tier upload limits). Which *field* takes which tag
// is delta's rather than Autumn's, in
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/models/files/model.rs
// — use_user_avatar wants avatars, use_background backgrounds, use_server_icon
// icons, use_server_banner banners, use_emoji emojis — and a route handed a file
// under the wrong one refuses it as a file that does not exist.
type FileTag string

const (
	FileTagAttachments FileTag = "attachments"
	FileTagAvatars     FileTag = "avatars"

	// FileTagBackgrounds is a UserProfile banner
	FileTagBackgrounds FileTag = "backgrounds"
	// FileTagBanners is a Server banner
	FileTagBanners FileTag = "banners"

	// FileTagIcons covers a server's icon and a group's alike.
	FileTagIcons  FileTag = "icons"
	FileTagEmojis FileTag = "emojis"
)

// Known reports whether the tag is one this library has a constant for.
//
// This is not validation and nothing refuses a tag that fails it: the backend
// may serve a bucket added after the list above was written, and a URL built
// from the tag the backend itself sent is right either way.
func (t FileTag) Known() bool {
	switch t {
	case FileTagAttachments, FileTagAvatars, FileTagBackgrounds,
		FileTagBanners, FileTagIcons, FileTagEmojis:
		return true
	}

	return false
}

// unknownFileTags remembers which unknown tags have already been reported: a
// channel of attachments carrying one would otherwise log per file, per render.
var unknownFileTags sync.Map

// check hands the tag straight back, having logged it once per distinct value if
// this library does not know it. A note for whoever maintains the list rather
// than a failure — an unknown tag is still the tag the file is served under, so
// every caller carries on with it.
func (t FileTag) check() FileTag {
	if t.Known() {
		return t
	}

	if _, seen := unknownFileTags.LoadOrStore(t, struct{}{}); !seen {
		log.Printf("revoltgo: file tag %q is not one of the buckets named in file.go; the list may be out of date", string(t))
	}

	return t
}

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

	// FileTag this file was uploaded to
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
	return EndpointAutumnFile(FileTag(f.Tag), f.ID, size)
}

type AttachmentMetadata struct {
	Type FileMetadataType `msg:"type" json:"type,omitzero"`

	Width  int `msg:"width" json:"width,omitzero"`
	Height int `msg:"height" json:"height,omitzero"`
}
