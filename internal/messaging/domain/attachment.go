package domain

// MediaAttachment represents a media attachment in a message.
type MediaAttachment struct {
	ID           string
	Type         MessageType
	URL          string
	ThumbnailURL *string
	FileName     *string
	FileSize     *int64
	MimeType     *string
	Width        *int
	Height       *int
	Duration     *int // For audio/video in seconds
}

// NewMediaAttachment creates a new media attachment.
func NewMediaAttachment(id string, mediaType MessageType, url string) *MediaAttachment {
	return &MediaAttachment{
		ID:   id,
		Type: mediaType,
		URL:  url,
	}
}

// WithThumbnail sets the thumbnail URL.
func (m *MediaAttachment) WithThumbnail(url string) *MediaAttachment {
	m.ThumbnailURL = &url
	return m
}

// WithFileName sets the file name.
func (m *MediaAttachment) WithFileName(name string) *MediaAttachment {
	m.FileName = &name
	return m
}

// WithFileSize sets the file size in bytes.
func (m *MediaAttachment) WithFileSize(size int64) *MediaAttachment {
	m.FileSize = &size
	return m
}

// WithMimeType sets the MIME type.
func (m *MediaAttachment) WithMimeType(mimeType string) *MediaAttachment {
	m.MimeType = &mimeType
	return m
}

// WithDimensions sets width and height for images/videos.
func (m *MediaAttachment) WithDimensions(width, height int) *MediaAttachment {
	m.Width = &width
	m.Height = &height
	return m
}

// WithDuration sets duration for audio/video.
func (m *MediaAttachment) WithDuration(seconds int) *MediaAttachment {
	m.Duration = &seconds
	return m
}
