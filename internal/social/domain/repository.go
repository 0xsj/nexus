package domain

import (
	"context"

	"github.com/0xsj/result"
)

// FollowRepository defines operations for follow relationship persistence.
type FollowRepository interface {
	// Create persists a new follow relationship.
	// Returns ErrAlreadyFollowing if already following.
	Create(ctx context.Context, follow *Follow) result.Result[*Follow]

	// FindByID retrieves a follow relationship by its ID.
	FindByID(ctx context.Context, id string) result.Result[*Follow]

	// FindByUsers retrieves a follow relationship between two users.
	// Returns ErrFollowNotFound if not following.
	FindByUsers(ctx context.Context, followerID, followingID string) result.Result[*Follow]

	// Delete removes a follow relationship (unfollow).
	// Returns ErrFollowNotFound if not following.
	Delete(ctx context.Context, followerID, followingID string) result.Result[struct{}]

	// GetFollowers retrieves users following a specific user.
	GetFollowers(ctx context.Context, userID string, params ListParams) result.Result[*FollowListResult]

	// GetFollowing retrieves users that a specific user is following.
	GetFollowing(ctx context.Context, userID string, params ListParams) result.Result[*FollowListResult]

	// GetMutualFollowers retrieves mutual followers (friends).
	GetMutualFollowers(ctx context.Context, userID string, params ListParams) result.Result[*FollowListResult]

	// IsFollowing checks if followerID follows followingID.
	IsFollowing(ctx context.Context, followerID, followingID string) result.Result[bool]

	// GetFollowerCount returns the number of followers for a user.
	GetFollowerCount(ctx context.Context, userID string) result.Result[int]

	// GetFollowingCount returns the number of users a user is following.
	GetFollowingCount(ctx context.Context, userID string) result.Result[int]

	// UpdateNotificationPreference updates the notification setting for a follow.
	UpdateNotificationPreference(ctx context.Context, followerID, followingID string, enable bool) result.Result[struct{}]

	// GetFollowingIDs returns just the IDs of users being followed (for feed generation).
	GetFollowingIDs(ctx context.Context, userID string) result.Result[[]string]
}

// BlockRepository defines operations for block relationship persistence.
type BlockRepository interface {
	// Create persists a new block relationship.
	// Returns ErrAlreadyBlocked if already blocked.
	Create(ctx context.Context, block *Block) result.Result[*Block]

	// FindByID retrieves a block relationship by its ID.
	FindByID(ctx context.Context, id string) result.Result[*Block]

	// FindByUsers retrieves a block relationship between two users.
	// Returns ErrBlockNotFound if not blocked.
	FindByUsers(ctx context.Context, blockerID, blockedID string) result.Result[*Block]

	// Delete removes a block relationship (unblock).
	// Returns ErrBlockNotFound if not blocked.
	Delete(ctx context.Context, blockerID, blockedID string) result.Result[struct{}]

	// GetBlockedUsers retrieves users blocked by a specific user.
	GetBlockedUsers(ctx context.Context, blockerID string, params ListParams) result.Result[*BlockListResult]

	// IsBlocked checks if blockerID has blocked blockedID.
	IsBlocked(ctx context.Context, blockerID, blockedID string) result.Result[bool]

	// IsBlockedBy checks if userID is blocked by any of the given user IDs.
	IsBlockedBy(ctx context.Context, userID string, potentialBlockers []string) result.Result[bool]
}

// MuteRepository defines operations for mute relationship persistence.
type MuteRepository interface {
	// Create persists a new mute relationship.
	// Returns ErrAlreadyMuted if already muted.
	Create(ctx context.Context, mute *Mute) result.Result[*Mute]

	// FindByID retrieves a mute relationship by its ID.
	FindByID(ctx context.Context, id string) result.Result[*Mute]

	// FindByUsers retrieves a mute relationship between two users.
	// Returns ErrMuteNotFound if not muted.
	FindByUsers(ctx context.Context, muterID, mutedID string) result.Result[*Mute]

	// Delete removes a mute relationship (unmute).
	// Returns ErrMuteNotFound if not muted.
	Delete(ctx context.Context, muterID, mutedID string) result.Result[struct{}]

	// GetMutedUsers retrieves users muted by a specific user.
	GetMutedUsers(ctx context.Context, muterID string, params ListParams) result.Result[*MuteListResult]

	// IsMuted checks if muterID has muted mutedID.
	IsMuted(ctx context.Context, muterID, mutedID string) result.Result[bool]

	// DeleteExpired removes all expired temporary mutes.
	DeleteExpired(ctx context.Context) result.Result[int64]

	// Update updates a mute relationship (e.g., extend duration).
	Update(ctx context.Context, mute *Mute) result.Result[*Mute]
}

// RelationshipRepository provides combined relationship queries.
type RelationshipRepository interface {
	// GetRelationshipStatus gets the bidirectional relationship status between two users.
	GetRelationshipStatus(ctx context.Context, userID, targetUserID string) result.Result[*RelationshipStatus]

	// GetRelationshipCounts gets follower/following counts for a user.
	GetRelationshipCounts(ctx context.Context, userID string) result.Result[*RelationshipCounts]
}

// ListParams defines parameters for listing relationships.
type ListParams struct {
	Limit  int
	Offset int
	Query  *string // Search query for filtering by username
	SortBy string  // "created_at", "username"
	Order  string  // "asc", "desc"
}

// FollowListResult contains paginated follow results.
type FollowListResult struct {
	UserIDs []string // Just user IDs (for simple cases)
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

// BlockListResult contains paginated block results.
type BlockListResult struct {
	UserIDs []string
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

// MuteListResult contains paginated mute results.
type MuteListResult struct {
	UserIDs []string
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

// RelationshipCounts contains follower/following counts.
type RelationshipCounts struct {
	FollowersCount int
	FollowingCount int
	MutualCount    int
}

// NewListParams creates ListParams with defaults.
func NewListParams(limit, offset int) ListParams {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return ListParams{
		Limit:  limit,
		Offset: offset,
		SortBy: "created_at",
		Order:  "desc",
	}
}

// WithQuery sets the search query.
func (p ListParams) WithQuery(query string) ListParams {
	p.Query = &query
	return p
}

// WithSortBy sets the sort field.
func (p ListParams) WithSortBy(sortBy string) ListParams {
	validSortFields := map[string]bool{
		"created_at": true,
		"username":   true,
	}

	if validSortFields[sortBy] {
		p.SortBy = sortBy
	}

	return p
}

// WithOrder sets the sort order.
func (p ListParams) WithOrder(order string) ListParams {
	if order == "asc" || order == "desc" {
		p.Order = order
	}

	return p
}
