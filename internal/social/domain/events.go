package domain

import (
	"fmt"
	"time"

	"github.com/0xsj/nexus/pkg/events"
)

const (
	// Event types
	EventTypeUserFollowed        = "social.user.followed"
	EventTypeUserUnfollowed      = "social.user.unfollowed"
	EventTypeUserBlocked         = "social.user.blocked"
	EventTypeUserUnblocked       = "social.user.unblocked"
	EventTypeUserMuted           = "social.user.muted"
	EventTypeUserUnmuted         = "social.user.unmuted"
	EventTypeFollowerMilestone   = "social.follower.milestone"
	EventTypeMutualFollowCreated = "social.mutual_follow.created"

	// Aggregate type
	AggregateTypeSocial = "social"
)

// ============================================================================
// Follow Events
// ============================================================================

// UserFollowedPayload is the payload for user followed event.
type UserFollowedPayload struct {
	FollowID            string `json:"follow_id"`
	FollowerID          string `json:"follower_id"`
	FollowingID         string `json:"following_id"`
	EnableNotifications bool   `json:"enable_notifications"`
	IsMutual            bool   `json:"is_mutual"`
}

// NewUserFollowedEvent creates a new user followed event.
func NewUserFollowedEvent(follow *Follow, isMutual bool) events.Event {
	payload := UserFollowedPayload{
		FollowID:            follow.ID,
		FollowerID:          follow.FollowerID,
		FollowingID:         follow.FollowingID,
		EnableNotifications: follow.EnableNotifications,
		IsMutual:            isMutual,
	}

	event := events.NewBaseEvent(
		EventTypeUserFollowed,
		AggregateTypeSocial,
		follow.FollowingID, // Aggregate on the user being followed
		payload,
	)

	event.WithMetadata("follower_id", follow.FollowerID)
	event.WithMetadata("following_id", follow.FollowingID)
	if isMutual {
		event.WithMetadata("is_mutual", "true")
	}

	return event
}

// UserUnfollowedPayload is the payload for user unfollowed event.
type UserUnfollowedPayload struct {
	FollowerID  string `json:"follower_id"`
	FollowingID string `json:"following_id"`
}

// NewUserUnfollowedEvent creates a new user unfollowed event.
func NewUserUnfollowedEvent(followerID, followingID string) events.Event {
	payload := UserUnfollowedPayload{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	event := events.NewBaseEvent(
		EventTypeUserUnfollowed,
		AggregateTypeSocial,
		followingID,
		payload,
	)

	event.WithMetadata("follower_id", followerID)
	event.WithMetadata("following_id", followingID)

	return event
}

// ============================================================================
// Block Events
// ============================================================================

// UserBlockedPayload is the payload for user blocked event.
type UserBlockedPayload struct {
	BlockID   string  `json:"block_id"`
	BlockerID string  `json:"blocker_id"`
	BlockedID string  `json:"blocked_id"`
	Reason    *string `json:"reason,omitempty"`
}

// NewUserBlockedEvent creates a new user blocked event.
func NewUserBlockedEvent(block *Block) events.Event {
	payload := UserBlockedPayload{
		BlockID:   block.ID,
		BlockerID: block.BlockerID,
		BlockedID: block.BlockedID,
		Reason:    block.Reason,
	}

	event := events.NewBaseEvent(
		EventTypeUserBlocked,
		AggregateTypeSocial,
		block.BlockerID,
		payload,
	)

	event.WithMetadata("blocker_id", block.BlockerID)
	event.WithMetadata("blocked_id", block.BlockedID)

	return event
}

// UserUnblockedPayload is the payload for user unblocked event.
type UserUnblockedPayload struct {
	BlockerID string `json:"blocker_id"`
	BlockedID string `json:"blocked_id"`
}

// NewUserUnblockedEvent creates a new user unblocked event.
func NewUserUnblockedEvent(blockerID, blockedID string) events.Event {
	payload := UserUnblockedPayload{
		BlockerID: blockerID,
		BlockedID: blockedID,
	}

	event := events.NewBaseEvent(
		EventTypeUserUnblocked,
		AggregateTypeSocial,
		blockerID,
		payload,
	)

	event.WithMetadata("blocker_id", blockerID)
	event.WithMetadata("blocked_id", blockedID)

	return event
}

// ============================================================================
// Mute Events
// ============================================================================

// UserMutedPayload is the payload for user muted event.
type UserMutedPayload struct {
	MuteID      string     `json:"mute_id"`
	MuterID     string     `json:"muter_id"`
	MutedID     string     `json:"muted_id"`
	IsPermanent bool       `json:"is_permanent"`
	MutedUntil  *time.Time `json:"muted_until,omitempty"`
}

// NewUserMutedEvent creates a new user muted event.
func NewUserMutedEvent(mute *Mute) events.Event {
	payload := UserMutedPayload{
		MuteID:      mute.ID,
		MuterID:     mute.MuterID,
		MutedID:     mute.MutedID,
		IsPermanent: mute.IsPermanent(),
		MutedUntil:  mute.MutedUntil,
	}

	event := events.NewBaseEvent(
		EventTypeUserMuted,
		AggregateTypeSocial,
		mute.MuterID,
		payload,
	)

	event.WithMetadata("muter_id", mute.MuterID)
	event.WithMetadata("muted_id", mute.MutedID)

	return event
}

// UserUnmutedPayload is the payload for user unmuted event.
type UserUnmutedPayload struct {
	MuterID string `json:"muter_id"`
	MutedID string `json:"muted_id"`
}

// NewUserUnmutedEvent creates a new user unmuted event.
func NewUserUnmutedEvent(muterID, mutedID string) events.Event {
	payload := UserUnmutedPayload{
		MuterID: muterID,
		MutedID: mutedID,
	}

	event := events.NewBaseEvent(
		EventTypeUserUnmuted,
		AggregateTypeSocial,
		muterID,
		payload,
	)

	event.WithMetadata("muter_id", muterID)
	event.WithMetadata("muted_id", mutedID)

	return event
}

// ============================================================================
// Special Events
// ============================================================================

// FollowerMilestonePayload is the payload for follower milestone event.
type FollowerMilestonePayload struct {
	UserID        string `json:"user_id"`
	FollowerCount int    `json:"follower_count"`
	Milestone     int    `json:"milestone"` // e.g., 100, 1000, 10000
}

// NewFollowerMilestoneEvent creates a follower milestone event.
func NewFollowerMilestoneEvent(userID string, followerCount, milestone int) events.Event {
	payload := FollowerMilestonePayload{
		UserID:        userID,
		FollowerCount: followerCount,
		Milestone:     milestone,
	}

	event := events.NewBaseEvent(
		EventTypeFollowerMilestone,
		AggregateTypeSocial,
		userID,
		payload,
	)

	event.WithMetadata("milestone", fmt.Sprintf("%d", milestone))

	return event
}

// MutualFollowCreatedPayload is the payload when two users follow each other.
type MutualFollowCreatedPayload struct {
	UserID1 string `json:"user_id_1"`
	UserID2 string `json:"user_id_2"`
}

// NewMutualFollowCreatedEvent creates a mutual follow event.
func NewMutualFollowCreatedEvent(userID1, userID2 string) events.Event {
	payload := MutualFollowCreatedPayload{
		UserID1: userID1,
		UserID2: userID2,
	}

	event := events.NewBaseEvent(
		EventTypeMutualFollowCreated,
		AggregateTypeSocial,
		userID1,
		payload,
	)

	event.WithMetadata("user_id_1", userID1)
	event.WithMetadata("user_id_2", userID2)

	return event
}
