package v1

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/wallet/application/query"
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles all wallet HTTP requests.
type Handler struct {
	commandBus   cqrs.CommandBus
	queryBus     cqrs.QueryBus
	challengeSvc domain.ChallengeService
	logger       log.Logger
}

// NewHandler creates a new Handler instance.
func NewHandler(
	commandBus cqrs.CommandBus,
	queryBus cqrs.QueryBus,
	challengeSvc domain.ChallengeService,
	logger log.Logger,
) *Handler {
	return &Handler{
		commandBus:   commandBus,
		queryBus:     queryBus,
		challengeSvc: challengeSvc,
		logger:       logger,
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

func dispatchWalletQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.WalletView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.WalletView), nil
}

func dispatchWalletListQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.WalletListView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.WalletListView), nil
}

func dispatchUserWalletStatsQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.UserWalletStatsView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.UserWalletStatsView), nil
}

func dispatchChainStatsQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.ChainStatsView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.ChainStatsView), nil
}

func dispatchChallengeQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (*query.ChallengeView, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*query.ChallengeView), nil
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

func dispatchIntQuery(ctx context.Context, bus cqrs.QueryBus, q cqrs.Query) (int, error) {
	result, err := bus.Dispatch(ctx, q)
	if err != nil {
		return 0, err
	}
	if result == nil {
		return 0, nil
	}
	return result.(int), nil
}

// ============================================================================
// Context Helpers
// ============================================================================

type contextKey string

const (
	contextKeyUserID    contextKey = "user_id"
	contextKeySessionID contextKey = "session_id"
)

// UserIDFromContext extracts the user ID from context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKeyUserID).(string)
	return userID, ok && userID != ""
}

// SessionIDFromContext extracts the session ID from context.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(contextKeySessionID).(string)
	return sessionID, ok && sessionID != ""
}

// ============================================================================
// Request Helpers
// ============================================================================

// GetClientIP extracts the client IP address from the request.
func GetClientIP(r interface {
	Header(string) string
	RemoteAddr() string
}) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		for i, c := range forwarded {
			if c == ',' {
				return forwarded[:i]
			}
		}
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	addr := r.RemoteAddr()
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}

	return addr
}

// GetUserAgent extracts the User-Agent from the request.
func GetUserAgent(r interface{ Header(string) string }) string {
	return r.Header("User-Agent")
}
