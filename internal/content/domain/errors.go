package domain

import (
	"fmt"

	"github.com/0xsj/result"
)

// ============================================================================
// Validation Errors (Kind: KindValidation)
// ============================================================================

func ErrInvalidPostID() error {
	return result.Validation("content.post.validate", "invalid post ID", nil)
}

func ErrEmptyPostID() error {
	return result.Validation("content.post.validate", "post ID cannot be empty", nil)
}

func ErrEmptyContent() error {
	return result.Validation("content.post.validate", "post content cannot be empty", nil)
}

func ErrContentTooLong(length, max int) error {
	return result.Validation(
		"content.post.validate",
		fmt.Sprintf("content too long: %d characters (max: %d)", length, max),
		map[string]any{
			"length": length,
			"max":    max,
		},
	)
}

func ErrInvalidContentType() error {
	return result.Validation("content.post.validate", "invalid content type", nil)
}

func ErrInvalidVisibility() error {
	return result.Validation("content.post.validate", "invalid visibility setting", nil)
}

func ErrInvalidStatus() error {
	return result.Validation("content.post.validate", "invalid post status", nil)
}

func ErrTooManyMediaItems(count, max int) error {
	return result.Validation(
		"content.post.validate",
		fmt.Sprintf("too many media items: %d (max: %d)", count, max),
		map[string]any{
			"count": count,
			"max":   max,
		},
	)
}

func ErrTooManyTags(count, max int) error {
	return result.Validation(
		"content.post.validate",
		fmt.Sprintf("too many tags: %d (max: %d)", count, max),
		map[string]any{
			"count": count,
			"max":   max,
		},
	)
}

func ErrInvalidPollOptions() error {
	return result.Validation("content.poll.validate", "poll must have at least 2 options", nil)
}

func ErrTooManyPollOptions(count, max int) error {
	return result.Validation(
		"content.poll.validate",
		fmt.Sprintf("too many poll options: %d (max: %d)", count, max),
		map[string]any{
			"count": count,
			"max":   max,
		},
	)
}

func ErrInvalidLocation() error {
	return result.Validation("content.location.validate", "invalid location coordinates", nil)
}

// ============================================================================
// Not Found Errors (Kind: KindNotFound)
// ============================================================================

func ErrPostNotFound(postID string) error {
	return result.NotFound("content.repository.find", "post not found")
}

func ErrDraftNotFound(draftID string) error {
	return result.NotFound("content.repository.find", "draft not found")
}

// ============================================================================
// Conflict Errors (Kind: KindConflict)
// ============================================================================

func ErrPostAlreadyPublished(postID string) error {
	return result.Conflict("content.post.publish", "post already published")
}

func ErrPostAlreadyDeleted(postID string) error {
	return result.Conflict("content.post.delete", "post already deleted")
}

// ============================================================================
// Domain Errors (Kind: KindDomain)
// ============================================================================

func ErrCannotEditPublishedPost() error {
	return result.Domain("content.post.edit", "cannot edit published post")
}

func ErrCannotPublishDraft() error {
	return result.Domain("content.post.publish", "cannot publish draft in current state")
}

func ErrCannotDeletePublishedPost() error {
	return result.Domain("content.post.delete", "cannot delete published post without archiving first")
}

func ErrPollExpired() error {
	return result.Domain("content.poll.vote", "poll has expired")
}

func ErrAlreadyVoted() error {
	return result.Domain("content.poll.vote", "user has already voted in this poll")
}

func ErrScheduledTimeInPast() error {
	return result.Domain("content.post.schedule", "scheduled time must be in the future")
}

// ============================================================================
// Unauthorized Errors (Kind: KindUnauthorized)
// ============================================================================

func ErrNotPostAuthor() error {
	return result.Unauthorized("content.post.authorize", "user is not the post author")
}

func ErrCannotAccessPrivatePost() error {
	return result.Unauthorized("content.post.access", "cannot access private post")
}
