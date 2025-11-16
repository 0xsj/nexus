package domain

// PostStatus represents the publication status of a post.
type PostStatus string

const (
	StatusDraft     PostStatus = "draft"
	StatusPublished PostStatus = "published"
	StatusArchived  PostStatus = "archived"
	StatusDeleted   PostStatus = "deleted"
)

// IsValid checks if the status is valid.
func (s PostStatus) IsValid() bool {
	switch s {
	case StatusDraft, StatusPublished, StatusArchived, StatusDeleted:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s PostStatus) String() string {
	return string(s)
}

// IsDraft checks if the post is a draft.
func (s PostStatus) IsDraft() bool {
	return s == StatusDraft
}

// IsPublished checks if the post is published.
func (s PostStatus) IsPublished() bool {
	return s == StatusPublished
}

// IsArchived checks if the post is archived.
func (s PostStatus) IsArchived() bool {
	return s == StatusArchived
}

// IsDeleted checks if the post is deleted.
func (s PostStatus) IsDeleted() bool {
	return s == StatusDeleted
}

// CanTransitionTo checks if the status can transition to a new status.
func (s PostStatus) CanTransitionTo(newStatus PostStatus) bool {
	switch s {
	case StatusDraft:
		// Draft can become published or deleted
		return newStatus == StatusPublished || newStatus == StatusDeleted
	case StatusPublished:
		// Published can become archived or deleted
		return newStatus == StatusArchived || newStatus == StatusDeleted
	case StatusArchived:
		// Archived can be republished or deleted
		return newStatus == StatusPublished || newStatus == StatusDeleted
	case StatusDeleted:
		// Deleted is final
		return false
	default:
		return false
	}
}
