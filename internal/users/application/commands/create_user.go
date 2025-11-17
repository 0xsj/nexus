package commands

import (
	"context"

	"github.com/0xsj/nexus/internal/users/application"
	"github.com/0xsj/nexus/internal/users/domain"
	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/result"
)

// CreateUserCommand represents the intent to create a new user.
type CreateUserCommand struct {
	Email    string
	Username string
	Password string
}

// CreateUserHandler handles user creation.
type CreateUserHandler struct {
	repo     domain.UserRepository
	eventBus events.EventBus
	logger   logger.Logger
}

// NewCreateUserHandler creates a new CreateUserHandler.
func NewCreateUserHandler(
	repo domain.UserRepository,
	eventBus events.EventBus,
	log logger.Logger,
) *CreateUserHandler {
	return &CreateUserHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   log,
	}
}

// Handle executes the CreateUser command.
func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) result.Result[application.UserDTO] {
	const op = "commands.CreateUserHandler.Handle"

	h.logger.Debug("Handling CreateUser command",
		logger.String("email", cmd.Email),
		logger.String("username", cmd.Username),
	)

	// Step 1: Validate and check email uniqueness
	email, err := result.Extract(
		result.ValidateAndCheck(
			domain.NewEmail(cmd.Email),
			func(e domain.Email) result.Result[bool] {
				return h.repo.ExistsByEmail(ctx, e)
			},
			func(e domain.Email) error {
				return domain.ErrEmailAlreadyExists(e.Value())
			},
			op+".email",
		),
		op,
	)
	if err != nil {
		return result.Err[application.UserDTO](err)
	}

	// Step 2: Validate and check username uniqueness
	username, err := result.Extract(
		result.ValidateAndCheck(
			domain.NewUsername(cmd.Username),
			func(u domain.Username) result.Result[bool] {
				return h.repo.ExistsByUsername(ctx, u)
			},
			func(u domain.Username) error {
				return domain.ErrUsernameAlreadyExists(u.Value())
			},
			op+".username",
		),
		op,
	)
	if err != nil {
		return result.Err[application.UserDTO](err)
	}

	// Step 3: Validate password
	password, err := result.Extract(domain.NewPassword(cmd.Password), op+".password")
	if err != nil {
		return result.Err[application.UserDTO](err)
	}

	// Step 4: Create user entity
	user := domain.NewUser(email, username, password)

	// Step 5: Save to repository
	savedUser, err := result.Extract(h.repo.Create(ctx, user), op+".create")
	if err != nil {
		h.logger.Error("Failed to create user",
			logger.Err(err),
			logger.String("email", cmd.Email),
		)
		return result.Err[application.UserDTO](err)
	}

	// Step 6: Publish domain event (async, don't fail on error)
	h.publishUserCreatedEvent(ctx, savedUser)

	// Step 7: Convert to DTO and return
	h.logger.Info("User created successfully",
		logger.String("user_id", savedUser.ID),
		logger.String("email", savedUser.Email),
	)

	return result.Ok(application.FromDomain(savedUser))
}

// publishUserCreatedEvent publishes a UserCreated event.
func (h *CreateUserHandler) publishUserCreatedEvent(ctx context.Context, user *domain.User) {
	event := domain.NewUserCreatedEvent(user.ID, user.Email, user.Username)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.logger.Error("Failed to publish UserCreated event",
			logger.Err(err),
			logger.String("user_id", user.ID),
		)
	}
}
