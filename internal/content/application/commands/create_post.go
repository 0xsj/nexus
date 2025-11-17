package commands

import (
	"context"

	"github.com/0xsj/nexus/internal/content/application"
	"github.com/0xsj/nexus/internal/content/domain"
	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/result"
)

// CreatePostCommand represents the intent to create a new post.
type CreatePostCommand struct {
	UserID     string
	Type       domain.ContentType
	Content    string
	Visibility domain.PostVisibility
	MediaIDs   []string
	Tags       []string
}

// CreatePostHandler handles post creation.
type CreatePostHandler struct {
	repo     domain.PostRepository
	eventBus events.EventBus
	logger   logger.Logger
}

// NewCreatePostHandler creates a new CreatePostHandler.
func NewCreatePostHandler(
	repo domain.PostRepository,
	eventBus events.EventBus,
	log logger.Logger,
) *CreatePostHandler {
	return &CreatePostHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   log,
	}
}

// Handle executes the CreatePost command.
func (h *CreatePostHandler) Handle(ctx context.Context, cmd CreatePostCommand) result.Result[application.PostDTO] {
	const op = "commands.CreatePostHandler.Handle"

	h.logger.Debug("Handling CreatePost command",
		logger.String("user_id", cmd.UserID),
		logger.String("type", string(cmd.Type)),
	)

	// Create post entity
	postResult := domain.NewPost(cmd.UserID, cmd.Type, cmd.Content, cmd.Visibility)
	post, err := result.Extract(postResult, op+".new_post")
	if err != nil {
		return result.Err[application.PostDTO](err)
	}

	// Add tags if provided
	for _, tag := range cmd.Tags {
		post.AddTag(tag)
	}

	// Save to repository
	savedPost, err := result.Extract(h.repo.Create(ctx, post), op+".create")
	if err != nil {
		h.logger.Error("Failed to create post",
			logger.Err(err),
			logger.String("user_id", cmd.UserID),
		)
		return result.Err[application.PostDTO](err)
	}

	// Publish domain event using your existing event
	h.publishPostCreatedEvent(ctx, savedPost)

	h.logger.Info("Post created successfully",
		logger.String("post_id", savedPost.ID),
		logger.String("user_id", savedPost.UserID),
	)

	return result.Ok(application.PostFromDomain(savedPost))
}

// publishPostCreatedEvent publishes a PostCreated event.
func (h *CreatePostHandler) publishPostCreatedEvent(ctx context.Context, post *domain.Post) {
	event := domain.NewPostCreatedEvent(post)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.logger.Error("Failed to publish PostCreated event",
			logger.Err(err),
			logger.String("post_id", post.ID),
		)
	}
}
