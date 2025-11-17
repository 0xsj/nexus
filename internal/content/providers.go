package content

import (
	"github.com/0xsj/nexus/internal/content/application/commands"
	"github.com/0xsj/nexus/internal/content/application/queries"
	"github.com/0xsj/nexus/internal/content/domain"
	contentgrpc "github.com/0xsj/nexus/internal/content/presentation/grpc"
	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/google/wire"
)

// ProviderSet provides all content domain dependencies.
var ProviderSet = wire.NewSet(
	ProvidePostRepository,
	ProvideCreatePostHandler,
	ProvideGetPostHandler,
	ProvideGetFeedHandler,
	ProvideContentGRPCHandler,
)

// ProvidePostRepository provides a stub post repository.
func ProvidePostRepository() domain.PostRepository {
	// TODO: Replace with real implementation later
	// For now, return nil - handlers will stub responses
	return nil
}

// ProvideCreatePostHandler provides the CreatePost command handler.
func ProvideCreatePostHandler(
	repo domain.PostRepository,
	eventBus events.EventBus,
	log logger.Logger,
) *commands.CreatePostHandler {
	return commands.NewCreatePostHandler(repo, eventBus, log)
}

// ProvideGetPostHandler provides the GetPost query handler.
func ProvideGetPostHandler(
	repo domain.PostRepository,
	log logger.Logger,
) *queries.GetPostHandler {
	return queries.NewGetPostHandler(repo, log)
}

// ProvideGetFeedHandler provides the GetFeed query handler.
func ProvideGetFeedHandler(
	repo domain.PostRepository,
	log logger.Logger,
) *queries.GetFeedHandler {
	return queries.NewGetFeedHandler(repo, log)
}

// ProvideContentGRPCHandler provides the Content gRPC handler.
func ProvideContentGRPCHandler(
	createPostHandler *commands.CreatePostHandler,
	getPostHandler *queries.GetPostHandler,
	getFeedHandler *queries.GetFeedHandler,
	log logger.Logger,
) *contentgrpc.ContentHandler {
	return contentgrpc.NewContentHandler(
		createPostHandler,
		getPostHandler,
		getFeedHandler,
		log,
	)
}
