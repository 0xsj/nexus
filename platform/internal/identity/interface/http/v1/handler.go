package v1

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/identity/application/query"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles all identity HTTP requests.
type Handler struct {
	commandBus cqrs.CommandBus
	queryBus   cqrs.QueryBus
	logger     log.Logger
}

// NewHandler creates a new Handler instance.
func NewHandler(
	commandBus cqrs.CommandBus,
	queryBus cqrs.QueryBus,
	logger log.Logger,
) *Handler {
	return &Handler{
		commandBus: commandBus,
		queryBus:   queryBus,
		logger:     logger,
	}
}

// ============================================================================
// Command Helpers
// ============================================================================

// dispatchCommand dispatches a command and returns the result.
func (h *Handler) dispatchCommand(ctx context.Context, cmd cqrs.Command) (*cqrs.CommandResult, error) {
	return h.commandBus.Dispatch(ctx, cmd)
}

// ============================================================================
// Query Helpers
// ============================================================================

// dispatchQuery dispatches a query and returns the result.
func (h *Handler) dispatchQuery(ctx context.Context, q cqrs.Query) (any, error) {
	return h.queryBus.Dispatch(ctx, q)
}

// Helper functions for typed query results

func dispatchUserQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.UserView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.UserView), nil
}

func dispatchUserListQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.UserListView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.UserListView), nil
}

func dispatchUserStatsQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.UserStatsView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.UserStatsView), nil
}

func dispatchPublicProfileQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.PublicProfileView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.PublicProfileView), nil
}

func dispatchSessionQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.SessionView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.SessionView), nil
}

func dispatchSessionListQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.SessionListView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.SessionListView), nil
}

func dispatchAPIKeyQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.APIKeyView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.APIKeyView), nil
}

func dispatchAPIKeyListQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.APIKeyListView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.APIKeyListView), nil
}

func dispatchBoolQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (bool, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, nil
	}
	return result.(bool), nil
}
