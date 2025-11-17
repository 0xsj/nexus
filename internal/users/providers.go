package users

import (
	"github.com/0xsj/nexus/internal/users/application/commands"
	"github.com/0xsj/nexus/internal/users/application/queries"
	"github.com/0xsj/nexus/internal/users/domain"
	usergrpc "github.com/0xsj/nexus/internal/users/presentation/grpc"
	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/google/wire"
)

// ProviderSet provides all users domain dependencies.
var ProviderSet = wire.NewSet(
	ProvideUserRepository,
	ProvideCreateUserHandler,
	ProvideGetUserHandler,
	ProvideUserGRPCHandler,
)

// ProvideUserRepository provides a stub user repository.
func ProvideUserRepository() domain.UserRepository {
	// TODO: Replace with real implementation later
	// For now, return nil - handlers will stub responses
	return nil
}

// ProvideCreateUserHandler provides the CreateUser command handler.
func ProvideCreateUserHandler(
	repo domain.UserRepository,
	eventBus events.EventBus,
	log logger.Logger,
) *commands.CreateUserHandler {
	return commands.NewCreateUserHandler(repo, eventBus, log)
}

// ProvideGetUserHandler provides the GetUser query handler.
func ProvideGetUserHandler(
	repo domain.UserRepository,
	log logger.Logger,
) *queries.GetUserHandler {
	return queries.NewGetUserHandler(repo, log)
}

// ProvideUserGRPCHandler provides the Users gRPC handler.
func ProvideUserGRPCHandler(
	createUserHandler *commands.CreateUserHandler,
	getUserHandler *queries.GetUserHandler,
	log logger.Logger,
) *usergrpc.UserHandler {
	return usergrpc.NewUserHandler(createUserHandler, getUserHandler, log)
}
