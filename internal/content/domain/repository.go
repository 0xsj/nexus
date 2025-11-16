package domain

import (
	"context"

	"github.com/0xsj/result"
)

// PostRepository defines the interface for post persistence operations.
type PostRepository interface {
	// Create persists a new post.
	Create(ctx context.Context, post *Post) result.Result[*Post]

	// FindByID retrieves a post by its ID.
	// Returns ErrPostNotFound if the post doesn't exist.
	FindByID(ctx context.Context, id string) result.Result[*Post]

	// FindByUserID retrieves all posts by a specific user.
	FindByUserID(ctx context.Context, userID string, params ListParams) result.Result[*PostListResult]

	// Update updates an existing post.
	// Returns ErrPostNotFound if the post doesn't exist.
	Update(ctx context.Context, post *Post) result.Result[*Post]

	// Delete soft-deletes a post by marking it as deleted.
	// Returns ErrPostNotFound if the post doesn't exist.
	Delete(ctx context.Context, id string) result.Result[struct{}]

	// List retrieves a paginated list of posts.
	List(ctx context.Context, params ListParams) result.Result[*PostListResult]

	// FindPublishedPosts retrieves published posts with filters.
	FindPublishedPosts(ctx context.Context, params FeedParams) result.Result[*PostListResult]

	// FindDrafts retrieves draft posts for a user.
	FindDrafts(ctx context.Context, userID string, params ListParams) result.Result[*PostListResult]

	// FindScheduledPosts retrieves posts that are scheduled to be published.
	FindScheduledPosts(ctx context.Context) result.Result[[]*Post]

	// FindByTag retrieves posts with a specific tag.
	FindByTag(ctx context.Context, tag string, params ListParams) result.Result[*PostListResult]

	// FindTrending retrieves trending posts.
	FindTrending(ctx context.Context, params TrendingParams) result.Result[*PostListResult]

	// ExistsByID checks if a post exists.
	ExistsByID(ctx context.Context, id string) result.Result[bool]
}

// ListParams defines parameters for listing posts.
type ListParams struct {
	Limit      int
	Offset     int
	SortBy     string // "created_at", "published_at", "updated_at"
	Order      string // "asc", "desc"
	Status     *PostStatus
	Visibility *PostVisibility
}

// FeedParams defines parameters for feed generation.
type FeedParams struct {
	Limit        int
	Offset       int
	UserID       *string  // For personalized feeds
	FollowingIDs []string // IDs of users being followed
	ExcludeIDs   []string // Post IDs to exclude
	Algorithm    FeedAlgorithm
}

// FeedAlgorithm represents the algorithm used for feed generation.
type FeedAlgorithm string

const (
	AlgorithmChronological FeedAlgorithm = "chronological"
	AlgorithmPopular       FeedAlgorithm = "popular"
	AlgorithmPersonalized  FeedAlgorithm = "personalized"
)

// TrendingParams defines parameters for trending posts.
type TrendingParams struct {
	Limit      int
	TimeWindow string // "1h", "24h", "7d", "30d"
	Category   *string
}

// PostListResult contains paginated post results.
type PostListResult struct {
	Posts   []*Post
	Total   int
	Limit   int
	Offset  int
	HasMore bool
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

// WithSortBy sets the sort field.
func (p ListParams) WithSortBy(sortBy string) ListParams {
	validSortFields := map[string]bool{
		"created_at":   true,
		"published_at": true,
		"updated_at":   true,
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

// WithStatus filters by post status.
func (p ListParams) WithStatus(status PostStatus) ListParams {
	p.Status = &status
	return p
}

// WithVisibility filters by visibility.
func (p ListParams) WithVisibility(visibility PostVisibility) ListParams {
	p.Visibility = &visibility
	return p
}

// NewFeedParams creates FeedParams with defaults.
func NewFeedParams(limit, offset int) FeedParams {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return FeedParams{
		Limit:     limit,
		Offset:    offset,
		Algorithm: AlgorithmChronological,
	}
}

// WithAlgorithm sets the feed algorithm.
func (p FeedParams) WithAlgorithm(algorithm FeedAlgorithm) FeedParams {
	p.Algorithm = algorithm
	return p
}

// WithUserID sets the user ID for personalized feeds.
func (p FeedParams) WithUserID(userID string) FeedParams {
	p.UserID = &userID
	return p
}

// WithFollowing sets the following user IDs.
func (p FeedParams) WithFollowing(followingIDs []string) FeedParams {
	p.FollowingIDs = followingIDs
	return p
}

// WithExclude sets post IDs to exclude.
func (p FeedParams) WithExclude(excludeIDs []string) FeedParams {
	p.ExcludeIDs = excludeIDs
	return p
}

// NewTrendingParams creates TrendingParams with defaults.
func NewTrendingParams(limit int, timeWindow string) TrendingParams {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Default to 24 hours if invalid
	validWindows := map[string]bool{
		"1h":  true,
		"24h": true,
		"7d":  true,
		"30d": true,
	}

	if !validWindows[timeWindow] {
		timeWindow = "24h"
	}

	return TrendingParams{
		Limit:      limit,
		TimeWindow: timeWindow,
	}
}

// WithCategory sets the category filter.
func (p TrendingParams) WithCategory(category string) TrendingParams {
	p.Category = &category
	return p
}
