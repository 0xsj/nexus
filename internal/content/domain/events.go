package domain

import (
	"fmt"
	"time"

	"github.com/0xsj/nexus/pkg/events"
)

const (
	// Event types
	EventTypePostCreated    = "content.post.created"
	EventTypePostUpdated    = "content.post.updated"
	EventTypePostPublished  = "content.post.published"
	EventTypePostScheduled  = "content.post.scheduled"
	EventTypePostArchived   = "content.post.archived"
	EventTypePostDeleted    = "content.post.deleted"
	EventTypeDraftCreated   = "content.draft.created"
	EventTypeDraftPublished = "content.draft.published"
	EventTypeMediaAttached  = "content.media.attached"
	EventTypePollCreated    = "content.poll.created"
	EventTypePollVoted      = "content.poll.voted"

	// Aggregate type
	AggregateTypeContent = "content"
)

// ============================================================================
// Post Events
// ============================================================================

// PostCreatedPayload is the payload for post created event.
type PostCreatedPayload struct {
	PostID     string         `json:"post_id"`
	UserID     string         `json:"user_id"`
	Type       ContentType    `json:"type"`
	Visibility PostVisibility `json:"visibility"`
	Status     PostStatus     `json:"status"`
	HasMedia   bool           `json:"has_media"`
	HasPoll    bool           `json:"has_poll"`
	TagsCount  int            `json:"tags_count"`
}

// NewPostCreatedEvent creates a new post created event.
func NewPostCreatedEvent(post *Post) events.Event {
	payload := PostCreatedPayload{
		PostID:     post.ID,
		UserID:     post.UserID,
		Type:       post.Type,
		Visibility: post.Visibility,
		Status:     post.Status,
		HasMedia:   post.HasMedia(),
		HasPoll:    post.HasPoll(),
		TagsCount:  len(post.Tags),
	}

	event := events.NewBaseEvent(
		EventTypePostCreated,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)
	event.WithMetadata("post_type", string(post.Type))
	event.WithMetadata("visibility", string(post.Visibility))

	return event
}

// PostUpdatedPayload is the payload for post updated event.
type PostUpdatedPayload struct {
	PostID        string     `json:"post_id"`
	UserID        string     `json:"user_id"`
	UpdatedFields []string   `json:"updated_fields"`
	Status        PostStatus `json:"status"`
}

// NewPostUpdatedEvent creates a new post updated event.
func NewPostUpdatedEvent(post *Post, updatedFields []string) events.Event {
	payload := PostUpdatedPayload{
		PostID:        post.ID,
		UserID:        post.UserID,
		UpdatedFields: updatedFields,
		Status:        post.Status,
	}

	event := events.NewBaseEvent(
		EventTypePostUpdated,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)

	return event
}

// PostPublishedPayload is the payload for post published event.
type PostPublishedPayload struct {
	PostID      string         `json:"post_id"`
	UserID      string         `json:"user_id"`
	Type        ContentType    `json:"type"`
	Visibility  PostVisibility `json:"visibility"`
	PublishedAt time.Time      `json:"published_at"`
	Tags        []string       `json:"tags,omitempty"`
	HasMedia    bool           `json:"has_media"`
	HasPoll     bool           `json:"has_poll"`
}

// NewPostPublishedEvent creates a new post published event.
func NewPostPublishedEvent(post *Post) events.Event {
	publishedAt := time.Now().UTC()
	if post.PublishedAt != nil {
		publishedAt = *post.PublishedAt
	}

	payload := PostPublishedPayload{
		PostID:      post.ID,
		UserID:      post.UserID,
		Type:        post.Type,
		Visibility:  post.Visibility,
		PublishedAt: publishedAt,
		Tags:        post.Tags,
		HasMedia:    post.HasMedia(),
		HasPoll:     post.HasPoll(),
	}

	event := events.NewBaseEvent(
		EventTypePostPublished,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)
	event.WithMetadata("post_type", string(post.Type))
	event.WithMetadata("visibility", string(post.Visibility))

	return event
}

// PostScheduledPayload is the payload for post scheduled event.
type PostScheduledPayload struct {
	PostID      string    `json:"post_id"`
	UserID      string    `json:"user_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

// NewPostScheduledEvent creates a new post scheduled event.
func NewPostScheduledEvent(post *Post) events.Event {
	scheduledAt := time.Now().UTC()
	if post.ScheduledAt != nil {
		scheduledAt = *post.ScheduledAt
	}

	payload := PostScheduledPayload{
		PostID:      post.ID,
		UserID:      post.UserID,
		ScheduledAt: scheduledAt,
	}

	event := events.NewBaseEvent(
		EventTypePostScheduled,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)
	event.WithMetadata("scheduled_at", scheduledAt.Format(time.RFC3339))

	return event
}

// PostArchivedPayload is the payload for post archived event.
type PostArchivedPayload struct {
	PostID string `json:"post_id"`
	UserID string `json:"user_id"`
}

// NewPostArchivedEvent creates a new post archived event.
func NewPostArchivedEvent(post *Post) events.Event {
	payload := PostArchivedPayload{
		PostID: post.ID,
		UserID: post.UserID,
	}

	event := events.NewBaseEvent(
		EventTypePostArchived,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)

	return event
}

// PostDeletedPayload is the payload for post deleted event.
type PostDeletedPayload struct {
	PostID string `json:"post_id"`
	UserID string `json:"user_id"`
}

// NewPostDeletedEvent creates a new post deleted event.
func NewPostDeletedEvent(post *Post) events.Event {
	payload := PostDeletedPayload{
		PostID: post.ID,
		UserID: post.UserID,
	}

	event := events.NewBaseEvent(
		EventTypePostDeleted,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)

	return event
}

// ============================================================================
// Draft Events
// ============================================================================

// DraftCreatedPayload is the payload for draft created event.
type DraftCreatedPayload struct {
	DraftID string      `json:"draft_id"`
	UserID  string      `json:"user_id"`
	Type    ContentType `json:"type"`
}

// NewDraftCreatedEvent creates a new draft created event.
func NewDraftCreatedEvent(post *Post) events.Event {
	payload := DraftCreatedPayload{
		DraftID: post.ID,
		UserID:  post.UserID,
		Type:    post.Type,
	}

	event := events.NewBaseEvent(
		EventTypeDraftCreated,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)

	return event
}

// DraftPublishedPayload is the payload for draft published event.
type DraftPublishedPayload struct {
	DraftID     string    `json:"draft_id"`
	PostID      string    `json:"post_id"`
	UserID      string    `json:"user_id"`
	PublishedAt time.Time `json:"published_at"`
}

// NewDraftPublishedEvent creates a new draft published event.
func NewDraftPublishedEvent(post *Post) events.Event {
	publishedAt := time.Now().UTC()
	if post.PublishedAt != nil {
		publishedAt = *post.PublishedAt
	}

	payload := DraftPublishedPayload{
		DraftID:     post.ID,
		PostID:      post.ID,
		UserID:      post.UserID,
		PublishedAt: publishedAt,
	}

	event := events.NewBaseEvent(
		EventTypeDraftPublished,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)

	return event
}

// ============================================================================
// Media Events
// ============================================================================

// MediaAttachedPayload is the payload for media attached event.
type MediaAttachedPayload struct {
	PostID     string   `json:"post_id"`
	UserID     string   `json:"user_id"`
	MediaIDs   []string `json:"media_ids"`
	MediaCount int      `json:"media_count"`
}

// NewMediaAttachedEvent creates a new media attached event.
func NewMediaAttachedEvent(post *Post, mediaIDs []string) events.Event {
	payload := MediaAttachedPayload{
		PostID:     post.ID,
		UserID:     post.UserID,
		MediaIDs:   mediaIDs,
		MediaCount: len(mediaIDs),
	}

	event := events.NewBaseEvent(
		EventTypeMediaAttached,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)
	event.WithMetadata("media_count", fmt.Sprintf("%d", len(mediaIDs)))

	return event
}

// ============================================================================
// Poll Events
// ============================================================================

// PollCreatedPayload is the payload for poll created event.
type PollCreatedPayload struct {
	PostID        string    `json:"post_id"`
	UserID        string    `json:"user_id"`
	Question      string    `json:"question"`
	OptionsCount  int       `json:"options_count"`
	ExpiresAt     time.Time `json:"expires_at"`
	AllowMultiple bool      `json:"allow_multiple"`
}

// NewPollCreatedEvent creates a new poll created event.
func NewPollCreatedEvent(post *Post, poll *Poll) events.Event {
	payload := PollCreatedPayload{
		PostID:        post.ID,
		UserID:        post.UserID,
		Question:      poll.Question,
		OptionsCount:  len(poll.Options),
		ExpiresAt:     poll.ExpiresAt,
		AllowMultiple: poll.AllowMultipleChoices,
	}

	event := events.NewBaseEvent(
		EventTypePollCreated,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)
	event.WithMetadata("poll_expires_at", poll.ExpiresAt.Format(time.RFC3339))

	return event
}

// PollVotedPayload is the payload for poll voted event.
type PollVotedPayload struct {
	PostID      string `json:"post_id"`
	UserID      string `json:"user_id"`
	VoterID     string `json:"voter_id"`
	OptionIndex int    `json:"option_index"`
	TotalVotes  int    `json:"total_votes"`
}

// NewPollVotedEvent creates a new poll voted event.
func NewPollVotedEvent(post *Post, voterID string, optionIndex int) events.Event {
	totalVotes := 0
	if post.Poll != nil {
		totalVotes = post.Poll.TotalVotes
	}

	payload := PollVotedPayload{
		PostID:      post.ID,
		UserID:      post.UserID,
		VoterID:     voterID,
		OptionIndex: optionIndex,
		TotalVotes:  totalVotes,
	}

	event := events.NewBaseEvent(
		EventTypePollVoted,
		AggregateTypeContent,
		post.ID,
		payload,
	)

	event.WithMetadata("user_id", post.UserID)
	event.WithMetadata("voter_id", voterID)

	return event
}
