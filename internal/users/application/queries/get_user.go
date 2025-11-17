package queries

import (
	"context"

	"github.com/0xsj/nexus/internal/users/application"
	"github.com/0xsj/nexus/internal/users/domain"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/result"
)

// GetUserQuery represents the intent to get a user by ID.
type GetUserQuery struct {
	UserID string
}

// GetUserHandler handles retrieving a user by ID.
type GetUserHandler struct {
	repo   domain.UserRepository
	logger logger.Logger
}

// NewGetUserHandler creates a new GetUserHandler.
func NewGetUserHandler(
	repo domain.UserRepository,
	log logger.Logger,
) *GetUserHandler {
	return &GetUserHandler{
		repo:   repo,
		logger: log,
	}
}

// Handle executes the GetUser query.
func (h *GetUserHandler) Handle(ctx context.Context, query GetUserQuery) result.Result[application.UserDTO] {
	const op = "queries.GetUserHandler.Handle"

	h.logger.Debug("Handling GetUser query",
		logger.String("user_id", query.UserID),
	)

	// Get user from repository
	user, err := result.Extract(h.repo.FindByID(ctx, query.UserID), op+".find_user")
	if err != nil {
		h.logger.Error("Failed to find user",
			logger.Err(err),
			logger.String("user_id", query.UserID),
		)
		return result.Err[application.UserDTO](err)
	}

	h.logger.Debug("Found user",
		logger.String("user_id", user.ID),
		logger.String("email", user.Email),
	)

	return result.Ok(application.FromDomain(user))
}
