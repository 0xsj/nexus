package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Get User Handler
// ============================================================================

// GetUserHandler handles GetUser queries.
type GetUserHandler struct {
	reader UserReader
}

// NewGetUserHandler creates a new GetUserHandler.
func NewGetUserHandler(reader UserReader) *GetUserHandler {
	return &GetUserHandler{reader: reader}
}

// Handle handles the GetUser query.
func (h *GetUserHandler) Handle(ctx context.Context, q *GetUser) (*UserView, error) {
	return h.reader.GetUser(ctx, q.UserID)
}

// ============================================================================
// Get User By DID Handler
// ============================================================================

// GetUserByDIDHandler handles GetUserByDID queries.
type GetUserByDIDHandler struct {
	reader UserReader
}

// NewGetUserByDIDHandler creates a new GetUserByDIDHandler.
func NewGetUserByDIDHandler(reader UserReader) *GetUserByDIDHandler {
	return &GetUserByDIDHandler{reader: reader}
}

// Handle handles the GetUserByDID query.
func (h *GetUserByDIDHandler) Handle(ctx context.Context, q *GetUserByDID) (*UserView, error) {
	return h.reader.GetUserByDID(ctx, q.DID)
}

// ============================================================================
// Get User By Wallet Handler
// ============================================================================

// GetUserByWalletHandler handles GetUserByWallet queries.
type GetUserByWalletHandler struct {
	reader UserReader
}

// NewGetUserByWalletHandler creates a new GetUserByWalletHandler.
func NewGetUserByWalletHandler(reader UserReader) *GetUserByWalletHandler {
	return &GetUserByWalletHandler{reader: reader}
}

// Handle handles the GetUserByWallet query.
func (h *GetUserByWalletHandler) Handle(ctx context.Context, q *GetUserByWallet) (*UserView, error) {
	return h.reader.GetUserByWallet(ctx, q.Address, q.Chain)
}

// ============================================================================
// Get User By Email Handler
// ============================================================================

// GetUserByEmailHandler handles GetUserByEmail queries.
type GetUserByEmailHandler struct {
	reader UserReader
}

// NewGetUserByEmailHandler creates a new GetUserByEmailHandler.
func NewGetUserByEmailHandler(reader UserReader) *GetUserByEmailHandler {
	return &GetUserByEmailHandler{reader: reader}
}

// Handle handles the GetUserByEmail query.
func (h *GetUserByEmailHandler) Handle(ctx context.Context, q *GetUserByEmail) (*UserView, error) {
	return h.reader.GetUserByEmail(ctx, q.Email)
}

// ============================================================================
// List Users Handler
// ============================================================================

// ListUsersHandler handles ListUsers queries.
type ListUsersHandler struct {
	reader UserReader
}

// NewListUsersHandler creates a new ListUsersHandler.
func NewListUsersHandler(reader UserReader) *ListUsersHandler {
	return &ListUsersHandler{reader: reader}
}

// Handle handles the ListUsers query.
func (h *ListUsersHandler) Handle(ctx context.Context, q *ListUsers) (*UserListView, error) {
	opts := ListUsersOptions{
		ListOptions: ListOptions{
			Limit:     q.Limit,
			Offset:    q.Offset,
			SortBy:    q.SortBy,
			SortOrder: q.SortOrder,
		},
		Status: q.Status,
	}
	return h.reader.ListUsers(ctx, opts)
}

// ============================================================================
// Get User Stats Handler
// ============================================================================

// GetUserStatsHandler handles GetUserStats queries.
type GetUserStatsHandler struct {
	reader UserReader
}

// NewGetUserStatsHandler creates a new GetUserStatsHandler.
func NewGetUserStatsHandler(reader UserReader) *GetUserStatsHandler {
	return &GetUserStatsHandler{reader: reader}
}

// Handle handles the GetUserStats query.
func (h *GetUserStatsHandler) Handle(ctx context.Context, q *GetUserStats) (*UserStatsView, error) {
	return h.reader.GetUserStats(ctx, q.UserID)
}

// ============================================================================
// Get Public Profile Handler
// ============================================================================

// GetPublicProfileHandler handles GetPublicProfile queries.
type GetPublicProfileHandler struct {
	reader UserReader
}

// NewGetPublicProfileHandler creates a new GetPublicProfileHandler.
func NewGetPublicProfileHandler(reader UserReader) *GetPublicProfileHandler {
	return &GetPublicProfileHandler{reader: reader}
}

// Handle handles the GetPublicProfile query.
func (h *GetPublicProfileHandler) Handle(ctx context.Context, q *GetPublicProfile) (*PublicProfileView, error) {
	return h.reader.GetPublicProfile(ctx, q.DID)
}

// ============================================================================
// Check User Exists Handler
// ============================================================================

// CheckUserExistsHandler handles CheckUserExists queries.
type CheckUserExistsHandler struct {
	reader UserReader
}

// NewCheckUserExistsHandler creates a new CheckUserExistsHandler.
func NewCheckUserExistsHandler(reader UserReader) *CheckUserExistsHandler {
	return &CheckUserExistsHandler{reader: reader}
}

// Handle handles the CheckUserExists query.
func (h *CheckUserExistsHandler) Handle(ctx context.Context, q *CheckUserExists) (bool, error) {
	if q.Email != nil {
		return h.reader.UserExistsByEmail(ctx, *q.Email)
	}
	if q.DID != nil {
		return h.reader.UserExistsByDID(ctx, *q.DID)
	}
	if q.Wallet != nil {
		return h.reader.UserExistsByWallet(ctx, q.Wallet.Address, q.Wallet.Chain)
	}
	return false, nil
}

// ============================================================================
// Get Session Handler
// ============================================================================

// GetSessionHandler handles GetSession queries.
type GetSessionHandler struct {
	reader SessionReader
}

// NewGetSessionHandler creates a new GetSessionHandler.
func NewGetSessionHandler(reader SessionReader) *GetSessionHandler {
	return &GetSessionHandler{reader: reader}
}

// Handle handles the GetSession query.
func (h *GetSessionHandler) Handle(ctx context.Context, q *GetSession) (*SessionView, error) {
	session, err := h.reader.GetSession(ctx, q.SessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.UserID != q.UserID {
		return nil, domain.ErrSessionNotFound("GetSessionHandler.Handle", q.SessionID)
	}

	return session, nil
}

// ============================================================================
// List User Sessions Handler
// ============================================================================

// ListUserSessionsHandler handles ListUserSessions queries.
type ListUserSessionsHandler struct {
	reader SessionReader
}

// NewListUserSessionsHandler creates a new ListUserSessionsHandler.
func NewListUserSessionsHandler(reader SessionReader) *ListUserSessionsHandler {
	return &ListUserSessionsHandler{reader: reader}
}

// Handle handles the ListUserSessions query.
func (h *ListUserSessionsHandler) Handle(ctx context.Context, q *ListUserSessions) (*SessionListView, error) {
	opts := ListSessionsOptions{
		ListOptions: ListOptions{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
		ActiveOnly:     q.ActiveOnly,
		CurrentSession: q.CurrentSession,
	}
	return h.reader.ListUserSessions(ctx, q.UserID, opts)
}

// ============================================================================
// Validate Session Handler
// ============================================================================

// ValidateSessionHandler handles ValidateSession queries.
type ValidateSessionHandler struct {
	reader SessionReader
}

// NewValidateSessionHandler creates a new ValidateSessionHandler.
func NewValidateSessionHandler(reader SessionReader) *ValidateSessionHandler {
	return &ValidateSessionHandler{reader: reader}
}

// Handle handles the ValidateSession query.
func (h *ValidateSessionHandler) Handle(ctx context.Context, q *ValidateSession) (*SessionView, error) {
	return h.reader.ValidateSession(ctx, q.SessionID, q.TokenHash)
}

// ============================================================================
// Get API Key Handler
// ============================================================================

// GetAPIKeyHandler handles GetAPIKey queries.
type GetAPIKeyHandler struct {
	reader APIKeyReader
}

// NewGetAPIKeyHandler creates a new GetAPIKeyHandler.
func NewGetAPIKeyHandler(reader APIKeyReader) *GetAPIKeyHandler {
	return &GetAPIKeyHandler{reader: reader}
}

// Handle handles the GetAPIKey query.
func (h *GetAPIKeyHandler) Handle(ctx context.Context, q *GetAPIKey) (*APIKeyView, error) {
	apiKey, err := h.reader.GetAPIKey(ctx, q.KeyID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if apiKey.UserID != q.UserID {
		return nil, domain.ErrAPIKeyNotFound("GetAPIKeyHandler.Handle", q.KeyID)
	}

	return apiKey, nil
}

// ============================================================================
// List User API Keys Handler
// ============================================================================

// ListUserAPIKeysHandler handles ListUserAPIKeys queries.
type ListUserAPIKeysHandler struct {
	reader APIKeyReader
}

// NewListUserAPIKeysHandler creates a new ListUserAPIKeysHandler.
func NewListUserAPIKeysHandler(reader APIKeyReader) *ListUserAPIKeysHandler {
	return &ListUserAPIKeysHandler{reader: reader}
}

// Handle handles the ListUserAPIKeys query.
func (h *ListUserAPIKeysHandler) Handle(ctx context.Context, q *ListUserAPIKeys) (*APIKeyListView, error) {
	opts := ListAPIKeysOptions{
		ListOptions: ListOptions{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
		ActiveOnly: q.ActiveOnly,
	}
	return h.reader.ListUserAPIKeys(ctx, q.UserID, opts)
}

// ============================================================================
// Validate API Key Handler
// ============================================================================

// ValidateAPIKeyHandler handles ValidateAPIKey queries.
type ValidateAPIKeyHandler struct {
	reader APIKeyReader
}

// NewValidateAPIKeyHandler creates a new ValidateAPIKeyHandler.
func NewValidateAPIKeyHandler(reader APIKeyReader) *ValidateAPIKeyHandler {
	return &ValidateAPIKeyHandler{reader: reader}
}

// Handle handles the ValidateAPIKey query.
func (h *ValidateAPIKeyHandler) Handle(ctx context.Context, q *ValidateAPIKey) (*APIKeyValidationView, error) {
	keyHash := domain.HashAPIKey(q.RawKey)
	return h.reader.ValidateAPIKey(ctx, keyHash)
}

// ============================================================================
// Validate Token Handler
// ============================================================================

// ValidateTokenHandler handles ValidateToken queries.
type ValidateTokenHandler struct {
	reader TokenReader
}

// NewValidateTokenHandler creates a new ValidateTokenHandler.
func NewValidateTokenHandler(reader TokenReader) *ValidateTokenHandler {
	return &ValidateTokenHandler{reader: reader}
}

// Handle handles the ValidateToken query.
func (h *ValidateTokenHandler) Handle(ctx context.Context, q *ValidateToken) (*TokenValidationView, error) {
	switch q.TokenType {
	case domain.TokenTypeAccess:
		return h.reader.ValidateAccessToken(ctx, q.Token)
	case domain.TokenTypeRefresh:
		return h.reader.ValidateRefreshToken(ctx, q.Token)
	default:
		return h.reader.ValidateAccessToken(ctx, q.Token)
	}
}

// ============================================================================
// Handler Dependencies
// ============================================================================

// HandlerDependencies contains all dependencies needed for query handlers.
type HandlerDependencies struct {
	UserReader    UserReader
	SessionReader SessionReader
	APIKeyReader  APIKeyReader
	TokenReader   TokenReader
}

// ============================================================================
// Handler Registration
// ============================================================================

// RegisterHandlers registers all identity query handlers with the query bus.
func RegisterHandlers(bus *cqrs.InMemoryQueryBus, deps HandlerDependencies) error {
	handlers := map[string]any{
		TypeGetUser:          NewGetUserHandler(deps.UserReader),
		TypeGetUserByDID:     NewGetUserByDIDHandler(deps.UserReader),
		TypeGetUserByWallet:  NewGetUserByWalletHandler(deps.UserReader),
		TypeGetUserByEmail:   NewGetUserByEmailHandler(deps.UserReader),
		TypeListUsers:        NewListUsersHandler(deps.UserReader),
		TypeGetUserStats:     NewGetUserStatsHandler(deps.UserReader),
		TypeGetPublicProfile: NewGetPublicProfileHandler(deps.UserReader),
		TypeCheckUserExists:  NewCheckUserExistsHandler(deps.UserReader),
		TypeGetSession:       NewGetSessionHandler(deps.SessionReader),
		TypeListUserSessions: NewListUserSessionsHandler(deps.SessionReader),
		TypeValidateSession:  NewValidateSessionHandler(deps.SessionReader),
		TypeGetAPIKey:        NewGetAPIKeyHandler(deps.APIKeyReader),
		TypeListUserAPIKeys:  NewListUserAPIKeysHandler(deps.APIKeyReader),
		TypeValidateAPIKey:   NewValidateAPIKeyHandler(deps.APIKeyReader),
		TypeValidateToken:    NewValidateTokenHandler(deps.TokenReader),
	}

	for queryType, handler := range handlers {
		if err := bus.Register(queryType, handler); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.QueryHandler[*GetUser, *UserView]                    = (*GetUserHandler)(nil)
	_ cqrs.QueryHandler[*GetUserByDID, *UserView]               = (*GetUserByDIDHandler)(nil)
	_ cqrs.QueryHandler[*GetUserByWallet, *UserView]            = (*GetUserByWalletHandler)(nil)
	_ cqrs.QueryHandler[*GetUserByEmail, *UserView]             = (*GetUserByEmailHandler)(nil)
	_ cqrs.QueryHandler[*ListUsers, *UserListView]              = (*ListUsersHandler)(nil)
	_ cqrs.QueryHandler[*GetUserStats, *UserStatsView]          = (*GetUserStatsHandler)(nil)
	_ cqrs.QueryHandler[*GetPublicProfile, *PublicProfileView]  = (*GetPublicProfileHandler)(nil)
	_ cqrs.QueryHandler[*CheckUserExists, bool]                 = (*CheckUserExistsHandler)(nil)
	_ cqrs.QueryHandler[*GetSession, *SessionView]              = (*GetSessionHandler)(nil)
	_ cqrs.QueryHandler[*ListUserSessions, *SessionListView]    = (*ListUserSessionsHandler)(nil)
	_ cqrs.QueryHandler[*ValidateSession, *SessionView]         = (*ValidateSessionHandler)(nil)
	_ cqrs.QueryHandler[*GetAPIKey, *APIKeyView]                = (*GetAPIKeyHandler)(nil)
	_ cqrs.QueryHandler[*ListUserAPIKeys, *APIKeyListView]      = (*ListUserAPIKeysHandler)(nil)
	_ cqrs.QueryHandler[*ValidateAPIKey, *APIKeyValidationView] = (*ValidateAPIKeyHandler)(nil)
	_ cqrs.QueryHandler[*ValidateToken, *TokenValidationView]   = (*ValidateTokenHandler)(nil)
)
