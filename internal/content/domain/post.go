package domain

import (
	"time"

	"github.com/0xsj/result"
	"github.com/google/uuid"
)

const (
	MaxContentLength = 5000
	MaxMediaItems    = 10
	MaxTags          = 30
	MaxMentions      = 50
)

// Post represents a content post (aggregate root).
type Post struct {
	ID          string
	UserID      string
	Type        ContentType
	Content     string
	Visibility  PostVisibility
	Status      PostStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
	ScheduledAt *time.Time

	// Media attachments
	MediaItems []*MediaItem

	// Link preview (for link posts)
	LinkPreview *LinkPreview

	// Poll (for poll posts)
	Poll *Poll

	// Tags and mentions
	Tags             []string
	MentionedUserIDs []string

	// Location
	Location *Location
}

// NewPost creates a new post as a draft.
func NewPost(
	userID string,
	contentType ContentType,
	content string,
	visibility PostVisibility,
) result.Result[*Post] {
	// Validate content type
	if !contentType.IsValid() {
		return result.Err[*Post](ErrInvalidContentType())
	}

	// Validate visibility
	if !visibility.IsValid() {
		return result.Err[*Post](ErrInvalidVisibility())
	}

	// Validate content length
	if len(content) == 0 {
		return result.Err[*Post](ErrEmptyContent())
	}

	if len(content) > MaxContentLength {
		return result.Err[*Post](ErrContentTooLong(len(content), MaxContentLength))
	}

	now := time.Now().UTC()

	post := &Post{
		ID:               uuid.New().String(),
		UserID:           userID,
		Type:             contentType,
		Content:          content,
		Visibility:       visibility,
		Status:           StatusDraft,
		CreatedAt:        now,
		UpdatedAt:        now,
		MediaItems:       []*MediaItem{},
		Tags:             []string{},
		MentionedUserIDs: []string{},
	}

	return result.Ok(post)
}

// NewPublishedPost creates a new post and immediately publishes it.
func NewPublishedPost(
	userID string,
	contentType ContentType,
	content string,
	visibility PostVisibility,
) result.Result[*Post] {
	postResult := NewPost(userID, contentType, content, visibility)
	if postResult.IsErr() {
		return postResult
	}

	post := postResult.Unwrap()
	if err := post.Publish(); err != nil {
		return result.Err[*Post](err)
	}

	return result.Ok(post)
}

// ============================================================================
// Media Management
// ============================================================================

// AddMedia adds a media item to the post.
func (p *Post) AddMedia(mediaItem *MediaItem) error {
	if len(p.MediaItems) >= MaxMediaItems {
		return ErrTooManyMediaItems(len(p.MediaItems)+1, MaxMediaItems)
	}

	p.MediaItems = append(p.MediaItems, mediaItem)
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// SetMedia sets the media items (replaces existing).
func (p *Post) SetMedia(mediaItems []*MediaItem) error {
	if len(mediaItems) > MaxMediaItems {
		return ErrTooManyMediaItems(len(mediaItems), MaxMediaItems)
	}

	p.MediaItems = mediaItems
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// ClearMedia removes all media items.
func (p *Post) ClearMedia() {
	p.MediaItems = []*MediaItem{}
	p.UpdatedAt = time.Now().UTC()
}

// ============================================================================
// Link Preview
// ============================================================================

// SetLinkPreview sets the link preview for link posts.
func (p *Post) SetLinkPreview(preview *LinkPreview) {
	p.LinkPreview = preview
	p.UpdatedAt = time.Now().UTC()
}

// ClearLinkPreview removes the link preview.
func (p *Post) ClearLinkPreview() {
	p.LinkPreview = nil
	p.UpdatedAt = time.Now().UTC()
}

// ============================================================================
// Poll Management
// ============================================================================

// SetPoll sets the poll for poll posts.
func (p *Post) SetPoll(poll *Poll) error {
	if p.Type != ContentTypePoll {
		return ErrInvalidContentType()
	}

	p.Poll = poll
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// ClearPoll removes the poll.
func (p *Post) ClearPoll() {
	p.Poll = nil
	p.UpdatedAt = time.Now().UTC()
}

// ============================================================================
// Tags and Mentions
// ============================================================================

// AddTag adds a tag to the post.
func (p *Post) AddTag(tag string) error {
	if len(p.Tags) >= MaxTags {
		return ErrTooManyTags(len(p.Tags)+1, MaxTags)
	}

	// Avoid duplicates
	for _, t := range p.Tags {
		if t == tag {
			return nil
		}
	}

	p.Tags = append(p.Tags, tag)
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// SetTags sets the tags (replaces existing).
func (p *Post) SetTags(tags []string) error {
	if len(tags) > MaxTags {
		return ErrTooManyTags(len(tags), MaxTags)
	}

	p.Tags = tags
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// AddMention adds a mentioned user ID.
func (p *Post) AddMention(userID string) error {
	if len(p.MentionedUserIDs) >= MaxMentions {
		return ErrTooManyTags(len(p.MentionedUserIDs)+1, MaxMentions) // Reuse error
	}

	// Avoid duplicates
	for _, id := range p.MentionedUserIDs {
		if id == userID {
			return nil
		}
	}

	p.MentionedUserIDs = append(p.MentionedUserIDs, userID)
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// SetMentions sets the mentioned user IDs (replaces existing).
func (p *Post) SetMentions(userIDs []string) error {
	if len(userIDs) > MaxMentions {
		return ErrTooManyTags(len(userIDs), MaxMentions)
	}

	p.MentionedUserIDs = userIDs
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// ============================================================================
// Location
// ============================================================================

// SetLocation sets the post location.
func (p *Post) SetLocation(location *Location) {
	p.Location = location
	p.UpdatedAt = time.Now().UTC()
}

// ClearLocation removes the location.
func (p *Post) ClearLocation() {
	p.Location = nil
	p.UpdatedAt = time.Now().UTC()
}

// ============================================================================
// Content Updates
// ============================================================================

// UpdateContent updates the post content.
func (p *Post) UpdateContent(content string) error {
	if p.Status == StatusPublished {
		return ErrCannotEditPublishedPost()
	}

	if len(content) == 0 {
		return ErrEmptyContent()
	}

	if len(content) > MaxContentLength {
		return ErrContentTooLong(len(content), MaxContentLength)
	}

	p.Content = content
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateVisibility updates the post visibility.
func (p *Post) UpdateVisibility(visibility PostVisibility) error {
	if !visibility.IsValid() {
		return ErrInvalidVisibility()
	}

	p.Visibility = visibility
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// ============================================================================
// Publishing and Status Management
// ============================================================================

// Publish publishes the post immediately.
func (p *Post) Publish() error {
	if !p.Status.CanTransitionTo(StatusPublished) {
		return ErrCannotPublishDraft()
	}

	now := time.Now().UTC()
	p.Status = StatusPublished
	p.PublishedAt = &now
	p.UpdatedAt = now
	return nil
}

// Schedule schedules the post for future publication.
func (p *Post) Schedule(scheduledAt time.Time) error {
	if scheduledAt.Before(time.Now().UTC()) {
		return ErrScheduledTimeInPast()
	}

	p.ScheduledAt = &scheduledAt
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// Archive archives the post.
func (p *Post) Archive() error {
	if !p.Status.CanTransitionTo(StatusArchived) {
		return ErrCannotDeletePublishedPost()
	}

	p.Status = StatusArchived
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// Delete soft-deletes the post.
func (p *Post) Delete() error {
	if !p.Status.CanTransitionTo(StatusDeleted) {
		return ErrCannotDeletePublishedPost()
	}

	p.Status = StatusDeleted
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// ============================================================================
// Query Methods
// ============================================================================

// IsDraft checks if the post is a draft.
func (p *Post) IsDraft() bool {
	return p.Status.IsDraft()
}

// IsPublished checks if the post is published.
func (p *Post) IsPublished() bool {
	return p.Status.IsPublished()
}

// IsArchived checks if the post is archived.
func (p *Post) IsArchived() bool {
	return p.Status.IsArchived()
}

// IsDeleted checks if the post is deleted.
func (p *Post) IsDeleted() bool {
	return p.Status.IsDeleted()
}

// IsScheduled checks if the post is scheduled for future publication.
func (p *Post) IsScheduled() bool {
	return p.ScheduledAt != nil && p.ScheduledAt.After(time.Now().UTC())
}

// ShouldPublish checks if a scheduled post should be published now.
func (p *Post) ShouldPublish() bool {
	if p.ScheduledAt == nil {
		return false
	}
	return time.Now().UTC().After(*p.ScheduledAt) && p.IsDraft()
}

// IsOwnedBy checks if the post belongs to a specific user.
func (p *Post) IsOwnedBy(userID string) bool {
	return p.UserID == userID
}

// CanBeEditedBy checks if a user can edit this post.
func (p *Post) CanBeEditedBy(userID string) bool {
	return p.IsOwnedBy(userID) && !p.IsDeleted()
}

// CanBeDeletedBy checks if a user can delete this post.
func (p *Post) CanBeDeletedBy(userID string) bool {
	return p.IsOwnedBy(userID) && !p.IsDeleted()
}

// HasMedia checks if the post has media attachments.
func (p *Post) HasMedia() bool {
	return len(p.MediaItems) > 0
}

// HasPoll checks if the post has a poll.
func (p *Post) HasPoll() bool {
	return p.Poll != nil
}

// HasLocation checks if the post has a location.
func (p *Post) HasLocation() bool {
	return p.Location != nil
}

// HasLinkPreview checks if the post has a link preview.
func (p *Post) HasLinkPreview() bool {
	return p.LinkPreview != nil
}
