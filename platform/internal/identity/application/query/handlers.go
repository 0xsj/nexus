package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Query Handlers
// ============================================================================

// Handlers handles all identity queries.
type Handlers struct {
	users       UserReader
	sessions    SessionReader
	connections ConnectionReader
	apiKeys     APIKeyReader
	tokens      TokenReader
	logger      log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	users UserReader,
	sessions SessionReader,
	connections ConnectionReader,
	apiKeys APIKeyReader,
	tokens TokenReader,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		users:       users,
		sessions:    sessions,
		connections: connections,
		apiKeys:     apiKeys,
		tokens:      tokens,
		logger:      logger,
	}
}

// NewHandlersFromRepository creates handlers from a composite repository.
func NewHandlersFromRepository(repo *CompositeReadRepository, logger log.Logger) *Handlers {
	return &Handlers{
		users:       repo.Users(),
		sessions:    repo.Sessions(),
		connections: repo.Connections(),
		apiKeys:     repo.APIKeys(),
		tokens:      repo.Tokens(),
		logger:      logger,
	}
}

// ============================================================================
// User Queries
// ============================================================================

// GetUser retrieves a user by ID.
func (h *Handlers) GetUser(ctx context.Context, q GetUser) (*UserView, error) {
	const op = "query.Handlers.GetUser"

	user, err := h.users.GetUser(ctx, q.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return user, nil
}

// GetUserByDID retrieves a user by DID.
func (h *Handlers) GetUserByDID(ctx context.Context, q GetUserByDID) (*UserView, error) {
	const op = "query.Handlers.GetUserByDID"

	user, err := h.users.GetUserByDID(ctx, q.DID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return user, nil
}

// GetUserByWallet retrieves a user by wallet address.
func (h *Handlers) GetUserByWallet(ctx context.Context, q GetUserByWallet) (*UserView, error) {
	const op = "query.Handlers.GetUserByWallet"

	user, err := h.users.GetUserByWallet(ctx, q.Address, q.Chain)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (h *Handlers) GetUserByEmail(ctx context.Context, q GetUserByEmail) (*UserView, error) {
	const op = "query.Handlers.GetUserByEmail"

	user, err := h.users.GetUserByEmail(ctx, q.Email)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return user, nil
}

// ListUsers retrieves a paginated list of users.
func (h *Handlers) ListUsers(ctx context.Context, q ListUsers) (*UserListView, error) {
	const op = "query.Handlers.ListUsers"

	opts := ListUsersOptions{
		ListOptions: ListOptions{
			Limit:     q.Limit,
			Offset:    q.Offset,
			SortBy:    q.SortBy,
			SortOrder: q.SortOrder,
		},
		Status: q.Status,
	}
	opts.ListOptions.Validate()

	users, err := h.users.ListUsers(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return users, nil
}

// GetUserStats retrieves statistics for a user.
func (h *Handlers) GetUserStats(ctx context.Context, q GetUserStats) (*UserStatsView, error) {
	const op = "query.Handlers.GetUserStats"

	stats, err := h.users.GetUserStats(ctx, q.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return stats, nil
}

// GetPublicProfile retrieves the public profile for a DID.
func (h *Handlers) GetPublicProfile(ctx context.Context, q GetPublicProfile) (*PublicProfileView, error) {
	const op = "query.Handlers.GetPublicProfile"

	profile, err := h.users.GetPublicProfile(ctx, q.DID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return profile, nil
}

// SearchUsers searches for users.
func (h *Handlers) SearchUsers(ctx context.Context, q SearchUsers) (*UserListView, error) {
	const op = "query.Handlers.SearchUsers"

	opts := ListOptions{
		Limit:     q.Limit,
		Offset:    q.Offset,
		SortBy:    q.SortBy,
		SortOrder: q.SortOrder,
	}
	opts.Validate()

	users, err := h.users.SearchUsers(ctx, q.Query, opts)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return users, nil
}

// CheckUserExists checks if a user exists.
func (h *Handlers) CheckUserExists(ctx context.Context, q CheckUserExists) (bool, error) {
	const op = "query.Handlers.CheckUserExists"

	if q.UserID != nil {
		return h.users.UserExists(ctx, *q.UserID)
	}

	if q.DID != nil {
		return h.users.UserExistsByDID(ctx, *q.DID)
	}

	if q.Email != nil {
		return h.users.UserExistsByEmail(ctx, *q.Email)
	}

	if q.Wallet != nil {
		return h.users.UserExistsByWallet(ctx, q.Wallet.Address, q.Wallet.Chain)
	}

	return false, errors.Validation(op, "at least one identifier required")
}

// ============================================================================
// Session Queries
// ============================================================================

// GetSession retrieves a session by ID.
func (h *Handlers) GetSession(ctx context.Context, q GetSession) (*SessionView, error) {
	const op = "query.Handlers.GetSession"

	session, err := h.sessions.GetSession(ctx, q.SessionID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Verify ownership
	if session.UserID != q.UserID {
		return nil, domain.ErrSessionNotFound(op, q.SessionID)
	}

	return session, nil
}

// ListUserSessions retrieves sessions for a user.
func (h *Handlers) ListUserSessions(ctx context.Context, q ListUserSessions) (*SessionListView, error) {
	const op = "query.Handlers.ListUserSessions"

	opts := ListSessionsOptions{
		ListOptions: ListOptions{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
		ActiveOnly:     q.ActiveOnly,
		CurrentSession: q.CurrentSession,
	}
	opts.ListOptions.Validate()

	sessions, err := h.sessions.ListUserSessions(ctx, q.UserID, opts)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return sessions, nil
}

// ValidateSession validates a session.
func (h *Handlers) ValidateSession(ctx context.Context, q ValidateSession) (*SessionView, error) {
	const op = "query.Handlers.ValidateSession"

	session, err := h.sessions.ValidateSession(ctx, q.SessionID, q.TokenHash)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return session, nil
}

// CountUserSessions counts sessions for a user.
func (h *Handlers) CountUserSessions(ctx context.Context, q CountUserSessions) (int, error) {
	const op = "query.Handlers.CountUserSessions"

	count, err := h.sessions.CountUserSessions(ctx, q.UserID, q.ActiveOnly)
	if err != nil {
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

// ============================================================================
// Connection Queries
// ============================================================================

// GetConnection retrieves a connection by ID.
func (h *Handlers) GetConnection(ctx context.Context, q GetConnection) (*ConnectionView, error) {
	const op = "query.Handlers.GetConnection"

	connection, err := h.connections.GetConnection(ctx, q.ConnectionID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Verify ownership
	if connection.UserID != q.UserID {
		return nil, domain.ErrConnectionNotFound(op, connection.Provider, q.UserID)
	}

	return connection, nil
}

// GetConnectionByProvider retrieves a connection by provider.
func (h *Handlers) GetConnectionByProvider(ctx context.Context, q GetConnectionByProvider) (*ConnectionView, error) {
	const op = "query.Handlers.GetConnectionByProvider"

	connection, err := h.connections.GetConnectionByProvider(ctx, q.UserID, q.Provider)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return connection, nil
}

// ListUserConnections retrieves connections for a user.
func (h *Handlers) ListUserConnections(ctx context.Context, q ListUserConnections) (*ConnectionListView, error) {
	const op = "query.Handlers.ListUserConnections"

	opts := ListConnectionsOptions{
		ListOptions: ListOptions{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
		ActiveOnly: q.ActiveOnly,
	}
	opts.ListOptions.Validate()

	connections, err := h.connections.ListUserConnections(ctx, q.UserID, opts)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return connections, nil
}

// CheckConnectionExists checks if a connection exists.
func (h *Handlers) CheckConnectionExists(ctx context.Context, q CheckConnectionExists) (bool, error) {
	const op = "query.Handlers.CheckConnectionExists"

	exists, err := h.connections.ConnectionExists(ctx, q.UserID, q.Provider)
	if err != nil {
		return false, errors.Wrap(err, op)
	}

	return exists, nil
}

// ============================================================================
// API Key Queries
// ============================================================================

// GetAPIKey retrieves an API key by ID.
func (h *Handlers) GetAPIKey(ctx context.Context, q GetAPIKey) (*APIKeyView, error) {
	const op = "query.Handlers.GetAPIKey"

	apiKey, err := h.apiKeys.GetAPIKey(ctx, q.KeyID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Verify ownership
	if apiKey.UserID != q.UserID {
		return nil, domain.ErrAPIKeyNotFound(op, q.KeyID)
	}

	return apiKey, nil
}

// ListUserAPIKeys retrieves API keys for a user.
func (h *Handlers) ListUserAPIKeys(ctx context.Context, q ListUserAPIKeys) (*APIKeyListView, error) {
	const op = "query.Handlers.ListUserAPIKeys"

	opts := ListAPIKeysOptions{
		ListOptions: ListOptions{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
		ActiveOnly: q.ActiveOnly,
	}
	opts.ListOptions.Validate()

	apiKeys, err := h.apiKeys.ListUserAPIKeys(ctx, q.UserID, opts)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return apiKeys, nil
}

// ValidateAPIKey validates an API key.
func (h *Handlers) ValidateAPIKey(ctx context.Context, q ValidateAPIKey) (*APIKeyValidationView, error) {
	const op = "query.Handlers.ValidateAPIKey"

	keyHash := domain.HashAPIKey(q.RawKey)

	validation, err := h.apiKeys.ValidateAPIKey(ctx, keyHash)
	if err != nil {
		return &APIKeyValidationView{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return validation, nil
}

// CountUserAPIKeys counts API keys for a user.
func (h *Handlers) CountUserAPIKeys(ctx context.Context, q CountUserAPIKeys) (int, error) {
	const op = "query.Handlers.CountUserAPIKeys"

	count, err := h.apiKeys.CountUserAPIKeys(ctx, q.UserID, q.ActiveOnly)
	if err != nil {
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

// ============================================================================
// Token Queries
// ============================================================================

// ValidateAccessToken validates an access token.
func (h *Handlers) ValidateAccessToken(ctx context.Context, q ValidateAccessToken) (*TokenValidationView, error) {
	const op = "query.Handlers.ValidateAccessToken"

	if h.tokens == nil {
		return nil, errors.Internal(op, nil).WithMessage("token reader not configured")
	}

	validation, err := h.tokens.ValidateAccessToken(ctx, q.Token)
	if err != nil {
		return &TokenValidationView{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return validation, nil
}

// ValidateRefreshToken validates a refresh token.
func (h *Handlers) ValidateRefreshToken(ctx context.Context, q ValidateRefreshToken) (*TokenValidationView, error) {
	const op = "query.Handlers.ValidateRefreshToken"

	if h.tokens == nil {
		return nil, errors.Internal(op, nil).WithMessage("token reader not configured")
	}

	validation, err := h.tokens.ValidateRefreshToken(ctx, q.Token)
	if err != nil {
		return &TokenValidationView{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return validation, nil
}
