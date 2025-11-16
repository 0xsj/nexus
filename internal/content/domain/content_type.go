package domain

// ContentType represents the type of content in a post.
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeVideo ContentType = "video"
	ContentTypeLink  ContentType = "link"
	ContentTypePoll  ContentType = "poll"
)

// IsValid checks if the content type is valid.
func (ct ContentType) IsValid() bool {
	switch ct {
	case ContentTypeText, ContentTypeImage, ContentTypeVideo, ContentTypeLink, ContentTypePoll:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (ct ContentType) String() string {
	return string(ct)
}

// RequiresMedia checks if this content type requires media attachments.
func (ct ContentType) RequiresMedia() bool {
	return ct == ContentTypeImage || ct == ContentTypeVideo
}

// AllowsMultipleMedia checks if this content type allows multiple media items.
func (ct ContentType) AllowsMultipleMedia() bool {
	return ct == ContentTypeImage || ct == ContentTypeVideo
}
