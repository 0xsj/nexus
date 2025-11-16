package domain

// MediaItem represents a media attachment (image or video).
type MediaItem struct {
	ID              string
	URL             string
	ThumbnailURL    string
	Type            string // "image" or "video"
	Width           *int
	Height          *int
	DurationSeconds *int
	Description     *string // Alt text for accessibility
}

// NewMediaItem creates a new MediaItem.
func NewMediaItem(id, url, thumbnailURL, mediaType string) *MediaItem {
	return &MediaItem{
		ID:           id,
		URL:          url,
		ThumbnailURL: thumbnailURL,
		Type:         mediaType,
	}
}

// IsImage checks if the media is an image.
func (m *MediaItem) IsImage() bool {
	return m.Type == "image"
}

// IsVideo checks if the media is a video.
func (m *MediaItem) IsVideo() bool {
	return m.Type == "video"
}

// WithDimensions sets the width and height.
func (m *MediaItem) WithDimensions(width, height int) *MediaItem {
	m.Width = &width
	m.Height = &height
	return m
}

// WithDuration sets the duration for videos.
func (m *MediaItem) WithDuration(seconds int) *MediaItem {
	m.DurationSeconds = &seconds
	return m
}

// WithDescription sets the alt text.
func (m *MediaItem) WithDescription(description string) *MediaItem {
	m.Description = &description
	return m
}

// LinkPreview represents a preview of a linked URL.
type LinkPreview struct {
	URL         string
	Title       string
	Description string
	ImageURL    string
	SiteName    string
}

// NewLinkPreview creates a new LinkPreview.
func NewLinkPreview(url, title, description, imageURL, siteName string) *LinkPreview {
	return &LinkPreview{
		URL:         url,
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		SiteName:    siteName,
	}
}
