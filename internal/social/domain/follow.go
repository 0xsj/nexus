package domain

import (
	"time"

	"github.com/0xsj/result"
	"github.com/google/uuid"
)

// Follow represents a follow relationship between two users.
type Follow struct {
	ID                  string
	FollowerID          string
	FollowingID         string
	CreatedAt           time.Time
	EnableNotifications bool
}

// NewFollow creates a new follow relationship.
func NewFollow(followerID, followingID string, enableNotifications bool) result.Result[*Follow] {
	// Validate: cannot follow yourself
	if followerID == followingID {
		return result.Err[*Follow](ErrCannotFollowSelf())
	}

	// Validate: valid user IDs
	if followerID == "" || followingID == "" {
		return result.Err[*Follow](ErrInvalidUserID())
	}

	follow := &Follow{
		ID:                  uuid.New().String(),
		FollowerID:          followerID,
		FollowingID:         followingID,
		CreatedAt:           time.Now().UTC(),
		EnableNotifications: enableNotifications,
	}

	return result.Ok(follow)
}

// ToggleNotifications toggles notification preferences.
func (f *Follow) ToggleNotifications() {
	f.EnableNotifications = !f.EnableNotifications
}

// SetNotifications sets notification preference.
func (f *Follow) SetNotifications(enable bool) {
	f.EnableNotifications = enable
}
