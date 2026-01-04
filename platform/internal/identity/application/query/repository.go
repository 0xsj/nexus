package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
)

// ============================================================================
// List Options
// ============================================================================

// ListOptions contains common pagination and sorting options.
type ListOptions struct {
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

// DefaultListOptions returns sensible defaults.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// ============================================================================
// Read Repository
// ============================================================================

// ReadRepository defines the interface for read-optimized queries.
// This is separate from domain repositories to allow for optimized read models.
type ReadRepository interface {
	UserReader
	SessionReader
	ConnectionReader
	APIKeyReader
}

// ============================================================================
// User Reader
// ============================================================================

// UserReader provides read operations for users.
type UserReader interface {
	// GetUser retrieves a user by ID.
	GetUser(ctx context.Context, userID string) (*UserView, error)

	// GetUserByDID retrieves a user by DID.
	GetUserByDID(ctx context.Context, did string) (*UserView, error)

	// GetUserByWallet retrieves a user by wallet address.
	GetUserByWallet(ctx context.Context, address string, chain domain.Chain) (*UserView, error)

	// GetUserByEmail retrieves a user by email.
	GetUserByEmail(ctx context.Context, email string) (*UserView, error)

	// ListUsers retrieves a paginated list of users.
	ListUsers(ctx context.Context, opts ListUsersOptions) (*UserListView, error)

	// GetUserStats retrieves statistics for a user.
	GetUserStats(ctx context.Context, userID string) (*UserStatsView, error)

	// GetPublicProfile retrieves the public profile for a DID.
	GetPublicProfile(ctx context.Context, did string) (*PublicProfileView, error)

	// SearchUsers searches for users.
	SearchUsers(ctx context.Context, query string, opts ListOptions) (*UserListView, error)

	// UserExists checks if a user exists.
	UserExists(ctx context.Context, userID string) (bool, error)

	// UserExistsByDID checks if a user exists by DID.
	UserExistsByDID(ctx context.Context, did string) (bool, error)

	// UserExistsByWallet checks if a user exists by wallet.
	UserExistsByWallet(ctx context.Context, address string, chain domain.Chain) (bool, error)

	// UserExistsByEmail checks if a user exists by email.
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
}

// ListUsersOptions contains options for listing users.
type ListUsersOptions struct {
	ListOptions
	Status *domain.UserStatus
}

// ============================================================================
// Session Reader
// ============================================================================

// SessionReader provides read operations for sessions.
type SessionReader interface {
	// GetSession retrieves a session by ID.
	GetSession(ctx context.Context, sessionID string) (*SessionView, error)

	// GetSessionByTokenHash retrieves a session by token hash.
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*SessionView, error)

	// ListUserSessions retrieves sessions for a user.
	ListUserSessions(ctx context.Context, userID string, opts ListSessionsOptions) (*SessionListView, error)

	// CountUserSessions counts sessions for a user.
	CountUserSessions(ctx context.Context, userID string, activeOnly bool) (int, error)

	// SessionExists checks if a session exists and is valid.
	SessionExists(ctx context.Context, sessionID string) (bool, error)

	// ValidateSession validates a session.
	ValidateSession(ctx context.Context, sessionID string, tokenHash string) (*SessionView, error)
}

// ListSessionsOptions contains options for listing sessions.
type ListSessionsOptions struct {
	ListOptions
	ActiveOnly     bool
	CurrentSession string // To mark which is current
}

// ============================================================================
// Connection Reader
// ============================================================================

// ConnectionReader provides read operations for connections.
type ConnectionReader interface {
	// GetConnection retrieves a connection by ID.
	GetConnection(ctx context.Context, connectionID string) (*ConnectionView, error)

	// GetConnectionByProvider retrieves a connection by user and provider.
	GetConnectionByProvider(ctx context.Context, userID string, provider domain.OAuthProvider) (*ConnectionView, error)

	// ListUserConnections retrieves connections for a user.
	ListUserConnections(ctx context.Context, userID string, opts ListConnectionsOptions) (*ConnectionListView, error)

	// CountUserConnections counts connections for a user.
	CountUserConnections(ctx context.Context, userID string, activeOnly bool) (int, error)

	// ConnectionExists checks if a connection exists.
	ConnectionExists(ctx context.Context, userID string, provider domain.OAuthProvider) (bool, error)
}

// ListConnectionsOptions contains options for listing connections.
type ListConnectionsOptions struct {
	ListOptions
	ActiveOnly bool
}

// ============================================================================
// API Key Reader
// ============================================================================

// APIKeyReader provides read operations for API keys.
type APIKeyReader interface {
	// GetAPIKey retrieves an API key by ID.
	GetAPIKey(ctx context.Context, keyID string) (*APIKeyView, error)

	// GetAPIKeyByHash retrieves an API key by key hash.
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKeyView, error)

	// ListUserAPIKeys retrieves API keys for a user.
	ListUserAPIKeys(ctx context.Context, userID string, opts ListAPIKeysOptions) (*APIKeyListView, error)

	// CountUserAPIKeys counts API keys for a user.
	CountUserAPIKeys(ctx context.Context, userID string, activeOnly bool) (int, error)

	// ValidateAPIKey validates an API key and returns details.
	ValidateAPIKey(ctx context.Context, keyHash string) (*APIKeyValidationView, error)
}

// ListAPIKeysOptions contains options for listing API keys.
type ListAPIKeysOptions struct {
	ListOptions
	ActiveOnly bool
}

// ============================================================================
// Token Reader
// ============================================================================

// TokenReader provides read operations for tokens.
type TokenReader interface {
	// ValidateAccessToken validates an access token.
	ValidateAccessToken(ctx context.Context, token string) (*TokenValidationView, error)

	// ValidateRefreshToken validates a refresh token.
	ValidateRefreshToken(ctx context.Context, token string) (*TokenValidationView, error)

	// IsTokenRevoked checks if a token is revoked.
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)
}

// ============================================================================
// Composite Read Repository
// ============================================================================

// CompositeReadRepository combines all readers into a single repository.
type CompositeReadRepository struct {
	users       UserReader
	sessions    SessionReader
	connections ConnectionReader
	apiKeys     APIKeyReader
	tokens      TokenReader
}

// NewCompositeReadRepository creates a new composite read repository.
func NewCompositeReadRepository(
	users UserReader,
	sessions SessionReader,
	connections ConnectionReader,
	apiKeys APIKeyReader,
	tokens TokenReader,
) *CompositeReadRepository {
	return &CompositeReadRepository{
		users:       users,
		sessions:    sessions,
		connections: connections,
		apiKeys:     apiKeys,
		tokens:      tokens,
	}
}

// Users returns the user reader.
func (r *CompositeReadRepository) Users() UserReader {
	return r.users
}

// Sessions returns the session reader.
func (r *CompositeReadRepository) Sessions() SessionReader {
	return r.sessions
}

// Connections returns the connection reader.
func (r *CompositeReadRepository) Connections() ConnectionReader {
	return r.connections
}

// APIKeys returns the API key reader.
func (r *CompositeReadRepository) APIKeys() APIKeyReader {
	return r.apiKeys
}

// Tokens returns the token reader.
func (r *CompositeReadRepository) Tokens() TokenReader {
	return r.tokens
}
