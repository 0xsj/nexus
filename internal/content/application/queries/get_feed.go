package queries

import (
	"context"

	"github.com/0xsj/nexus/internal/content/application"
	"github.com/0xsj/nexus/internal/content/domain"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/result"
)

// GetFeedQuery represents the intent to get a user's feed.
type GetFeedQuery struct {
	UserID       string
	FollowingIDs []string
	Limit        int
	Offset       int
	Algorithm    domain.FeedAlgorithm
}

// GetFeedHandler handles retrieving a user's feed.
type GetFeedHandler struct {
	repo   domain.PostRepository
	logger logger.Logger
}

// NewGetFeedHandler creates a new GetFeedHandler.
func NewGetFeedHandler(
	repo domain.PostRepository,
	log logger.Logger,
) *GetFeedHandler {
	return &GetFeedHandler{
		repo:   repo,
		logger: log,
	}
}

// Handle executes the GetFeed query.
func (h *GetFeedHandler) Handle(ctx context.Context, query GetFeedQuery) result.Result[application.PostListDTO] {
	const op = "queries.GetFeedHandler.Handle"

	h.logger.Debug("Handling GetFeed query",
		logger.String("user_id", query.UserID),
		logger.Int("following_count", len(query.FollowingIDs)),
	)

	params := domain.FeedParams{
		UserID:       &query.UserID,
		FollowingIDs: query.FollowingIDs,
		Limit:        query.Limit,
		Offset:       query.Offset,
		Algorithm:    query.Algorithm,
	}

	// Get feed from repository
	feedResult, err := result.Extract(h.repo.FindPublishedPosts(ctx, params), op+".find_feed")
	if err != nil {
		h.logger.Error("Failed to get feed",
			logger.Err(err),
			logger.String("user_id", query.UserID),
		)
		return result.Err[application.PostListDTO](err)
	}

	h.logger.Debug("Retrieved feed",
		logger.String("user_id", query.UserID),
		logger.Int("post_count", len(feedResult.Posts)),
	)

	return result.Ok(application.PostListFromDomain(feedResult))
}
