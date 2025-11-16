package domain

import (
	"github.com/0xsj/result"
)

// ============================================================================
// Validation Errors (Kind: KindValidation)
// ============================================================================

func ErrInvalidUserID() error {
	return result.Validation("social.user.validate", "invalid user ID", nil)
}

func ErrCannotFollowSelf() error {
	return result.Validation("social.follow.validate", "cannot follow yourself", nil)
}

func ErrCannotBlockSelf() error {
	return result.Validation("social.block.validate", "cannot block yourself", nil)
}

func ErrCannotMuteSelf() error {
	return result.Validation("social.mute.validate", "cannot mute yourself", nil)
}

func ErrInvalidMuteDuration() error {
	return result.Validation("social.mute.validate", "mute duration must be in the future", nil)
}

// ============================================================================
// Not Found Errors (Kind: KindNotFound)
// ============================================================================

func ErrRelationshipNotFound(followerID, followingID string) error {
	return result.NotFound("social.relationship.find", "relationship not found")
}

func ErrFollowNotFound(followerID, followingID string) error {
	return result.NotFound("social.follow.find", "follow relationship not found")
}

func ErrBlockNotFound(blockerID, blockedID string) error {
	return result.NotFound("social.block.find", "block relationship not found")
}

func ErrMuteNotFound(muterID, mutedID string) error {
	return result.NotFound("social.mute.find", "mute relationship not found")
}

// ============================================================================
// Conflict Errors (Kind: KindConflict)
// ============================================================================

func ErrAlreadyFollowing(followerID, followingID string) error {
	return result.Conflict("social.follow.create", "already following this user")
}

func ErrAlreadyBlocked(blockerID, blockedID string) error {
	return result.Conflict("social.block.create", "user already blocked")
}

func ErrAlreadyMuted(muterID, mutedID string) error {
	return result.Conflict("social.mute.create", "user already muted")
}

// ============================================================================
// Domain Errors (Kind: KindDomain)
// ============================================================================

func ErrCannotFollowBlockedUser() error {
	return result.Domain("social.follow.validate", "cannot follow a user you've blocked")
}

func ErrCannotFollowBlocker() error {
	return result.Domain("social.follow.validate", "cannot follow a user who has blocked you")
}

func ErrUserIsBlocked() error {
	return result.Domain("social.relationship.validate", "user is blocked")
}

func ErrBlockedByUser() error {
	return result.Domain("social.relationship.validate", "you are blocked by this user")
}

// ============================================================================
// Unauthorized Errors (Kind: KindUnauthorized)
// ============================================================================

func ErrUnauthorizedAction() error {
	return result.Unauthorized("social.authorize", "unauthorized to perform this action")
}
