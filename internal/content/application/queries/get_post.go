package queries

import (
	"context"

	"github.com/0xsj/nexus/internal/content/application"
	"github.com/0xsj/nexus/internal/content/domain"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/result"
)

// GetPostQuery represents the intent to get a post by ID.
type GetPostQuery struct {
	PostID string
}

// GetPostHandler handles retrieving a post by ID.
type GetPostHandler struct {
	repo   domain.PostRepository
	logger logger.Logger
}

// NewGetPostHandler creates a new GetPostHandler.
func NewGetPostHandler(
	repo domain.PostRepository,
	log logger.Logger,
) *GetPostHandler {
	return &GetPostHandler{
		repo:   repo,
		logger: log,
	}
}

// Handle executes the GetPost query.
func (h *GetPostHandler) Handle(ctx context.Context, query GetPostQuery) result.Result[application.PostDTO] {
	const op = "queries.GetPostHandler.Handle"

	h.logger.Debug("Handling GetPost query",
		logger.String("post_id", query.PostID),
	)

	// Get post from repository
	post, err := result.Extract(h.repo.FindByID(ctx, query.PostID), op+".find_post")
	if err != nil {
		h.logger.Error("Failed to find post",
			logger.Err(err),
			logger.String("post_id", query.PostID),
		)
		return result.Err[application.PostDTO](err)
	}

	h.logger.Debug("Found post",
		logger.String("post_id", post.ID),
		logger.String("user_id", post.UserID),
	)

	return result.Ok(application.PostFromDomain(post))
}
