package query

import (
	"context"
	"strings"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Identity context.
type Handlers struct {
	userRepo       domain.UserRepository
	sessionRepo    domain.SessionRepository
	oauthStateRepo domain.OAuthStateRepository
	userLookup     domain.UserLookup
	sessionLookup  domain.SessionLookup
	logger         log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	oauthStateRepo domain.OAuthStateRepository,
	userLookup domain.UserLookup,
	sessionLookup domain.SessionLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		oauthStateRepo: oauthStateRepo,
		userLookup:     userLookup,
		sessionLookup:  sessionLookup,
		logger:         logger,
	}
}

// ============================================================================
// User Query Handlers
// ============================================================================

// HandleGetUser handles the GetUser query.
func (h *Handlers) HandleGetUser(ctx context.Context, q GetUser) (*GetUserResult, error) {
	const op = "Handlers.HandleGetUser"

	// Parse user ID
	userID := domain.UserIDFromTypesID(q.UserID)

	// Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapUserToResult(user), nil
}

// HandleGetUserByEmail handles the GetUserByEmail query.
func (h *Handlers) HandleGetUserByEmail(ctx context.Context, q GetUserByEmail) (*GetUserResult, error) {
	const op = "Handlers.HandleGetUserByEmail"

	// Find user ID by email
	userID, err := h.userLookup.GetUserIDByEmail(ctx, q.Email)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapUserToResult(user), nil
}

// HandleGetUserByDID handles the GetUserByDID query.
func (h *Handlers) HandleGetUserByDID(ctx context.Context, q GetUserByDID) (*GetUserResult, error) {
	const op = "Handlers.HandleGetUserByDID"

	// Find user ID by DID
	userID, err := h.userLookup.GetUserIDByDID(ctx, q.DID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapUserToResult(user), nil
}

// HandleListUsers handles the ListUsers query.
func (h *Handlers) HandleListUsers(ctx context.Context, q ListUsers) (*ListUsersResult, error) {
	const op = "Handlers.HandleListUsers"

	// Set default page size
	pageSize := q.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	// Build filter
	filter := domain.UserFilter{
		PageSize: pageSize,
		Cursor:   q.Cursor,
	}

	if q.Status != nil {
		status, err := domain.ParseUserStatus(*q.Status)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		filter.Status = &status
	}

	// Execute lookup
	users, nextCursor, totalCount, err := h.userLookup.ListUsers(ctx, filter)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Map to summaries
	summaries := make([]UserSummary, len(users))
	for i, user := range users {
		summaries[i] = UserSummary{
			UserID:      user.ID().String(),
			Email:       user.Email().String(),
			DisplayName: user.DisplayName().String(),
			Status:      user.Status().String(),
			CreatedAt:   user.CreatedAt(),
		}
	}

	return &ListUsersResult{
		Users:      summaries,
		NextCursor: nextCursor,
		TotalCount: totalCount,
	}, nil
}

// HandleSearchUsers handles the SearchUsers query.
func (h *Handlers) HandleSearchUsers(ctx context.Context, q SearchUsers) (*ListUsersResult, error) {
	const op = "Handlers.HandleSearchUsers"

	// Set default page size
	pageSize := q.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	// Execute search
	users, nextCursor, totalCount, err := h.userLookup.SearchUsers(ctx, q.Query, pageSize, q.Cursor)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Map to summaries
	summaries := make([]UserSummary, len(users))
	for i, user := range users {
		summaries[i] = UserSummary{
			UserID:      user.ID().String(),
			Email:       user.Email().String(),
			DisplayName: user.DisplayName().String(),
			Status:      user.Status().String(),
			CreatedAt:   user.CreatedAt(),
		}
	}

	return &ListUsersResult{
		Users:      summaries,
		NextCursor: nextCursor,
		TotalCount: totalCount,
	}, nil
}

// HandleGetUserProfile handles the GetUserProfile query.
func (h *Handlers) HandleGetUserProfile(ctx context.Context, q GetUserProfile) (*GetUserProfileResult, error) {
	const op = "Handlers.HandleGetUserProfile"

	// Parse user ID
	userID := domain.UserIDFromTypesID(q.UserID)

	// Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Check if user is active (only active users have public profiles)
	if !user.IsActive() {
		return nil, domain.UserNotFound(op, userID.String())
	}

	return &GetUserProfileResult{
		UserID:      user.ID().String(),
		DisplayName: user.DisplayName().String(),
		PrimaryDID:  user.PrimaryDID(),
		CreatedAt:   user.CreatedAt(),
	}, nil
}

// ============================================================================
// Session Query Handlers
// ============================================================================

// HandleGetSession handles the GetSession query.
func (h *Handlers) HandleGetSession(ctx context.Context, q GetSession) (*GetSessionResult, error) {
	const op = "Handlers.HandleGetSession"

	// Parse session ID
	sessionID := domain.SessionIDFromTypesID(q.SessionID)

	// Load session
	session, err := h.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapSessionToResult(session), nil
}

// HandleListUserSessions handles the ListUserSessions query.
func (h *Handlers) HandleListUserSessions(ctx context.Context, q ListUserSessions) (*ListUserSessionsResult, error) {
	const op = "Handlers.HandleListUserSessions"

	// Parse user ID
	userID := domain.UserIDFromTypesID(q.UserID)

	// Set default page size
	pageSize := q.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	// Build filter
	filter := domain.SessionFilter{
		UserID:     userID,
		ActiveOnly: q.ActiveOnly,
		PageSize:   pageSize,
		Cursor:     q.Cursor,
	}

	// Execute lookup
	sessions, nextCursor, totalCount, err := h.sessionLookup.ListSessions(ctx, filter)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Map to summaries
	summaries := make([]SessionSummary, len(sessions))
	for i, session := range sessions {
		summaries[i] = SessionSummary{
			SessionID:  session.ID().String(),
			AuthMethod: session.AuthMethod().String(),
			Status:     session.Status().String(),
			IPAddress:  session.IPAddress(),
			UserAgent:  session.UserAgent(),
			CreatedAt:  session.CreatedAt(),
			ExpiresAt:  session.ExpiresAt(),
			LastUsedAt: session.LastUsedAt(),
		}
	}

	return &ListUserSessionsResult{
		Sessions:   summaries,
		NextCursor: nextCursor,
		TotalCount: totalCount,
	}, nil
}

// HandleGetActiveSessionCount handles the GetActiveSessionCount query.
func (h *Handlers) HandleGetActiveSessionCount(ctx context.Context, q GetActiveSessionCount) (*GetActiveSessionCountResult, error) {
	const op = "Handlers.HandleGetActiveSessionCount"

	// Parse user ID
	userID := domain.UserIDFromTypesID(q.UserID)

	// Get count
	count, err := h.sessionLookup.CountActiveSessions(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &GetActiveSessionCountResult{
		UserID: userID.String(),
		Count:  count,
	}, nil
}

// HandleValidateSession handles the ValidateSession query.
func (h *Handlers) HandleValidateSession(ctx context.Context, q ValidateSession) (*ValidateSessionResult, error) {
	const op = "Handlers.HandleValidateSession"

	// Parse session ID
	sessionID := domain.SessionIDFromTypesID(q.SessionID)

	// Load session
	session, err := h.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return &ValidateSessionResult{
				Valid:  false,
				Reason: "session not found",
			}, nil
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	// Parse and verify token
	token, err := domain.ParseToken(q.Token)
	if err != nil {
		return &ValidateSessionResult{
			Valid:  false,
			Reason: "invalid token format",
		}, nil
	}

	if !session.VerifyToken(token) {
		return &ValidateSessionResult{
			Valid:  false,
			Reason: "token mismatch",
		}, nil
	}

	// Check session status
	if !session.IsActive() {
		return &ValidateSessionResult{
			Valid:  false,
			Reason: "session is " + session.Status().String(),
		}, nil
	}

	// Check expiration
	if session.IsExpired() {
		return &ValidateSessionResult{
			Valid:  false,
			Reason: "session expired",
		}, nil
	}

	// Load user to check status
	user, err := h.userRepo.Get(ctx, session.UserID())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return &ValidateSessionResult{
				Valid:  false,
				Reason: "user not found",
			}, nil
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	if !user.CanAuthenticate() {
		return &ValidateSessionResult{
			Valid:  false,
			Reason: "user is " + user.Status().String(),
		}, nil
	}

	return &ValidateSessionResult{
		Valid:      true,
		SessionID:  session.ID().String(),
		UserID:     session.UserID().String(),
		AuthMethod: session.AuthMethod().String(),
		ExpiresAt:  session.ExpiresAt(),
	}, nil
}

// ============================================================================
// DID Query Handlers
// ============================================================================

// HandleGetUserDIDs handles the GetUserDIDs query.
func (h *Handlers) HandleGetUserDIDs(ctx context.Context, q GetUserDIDs) (*GetUserDIDsResult, error) {
	const op = "Handlers.HandleGetUserDIDs"

	// Parse user ID
	userID := domain.UserIDFromTypesID(q.UserID)

	// Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Map DIDs to info
	dids := user.DIDs()
	primaryDID := user.PrimaryDID()
	didInfos := make([]DIDInfo, len(dids))

	for i, did := range dids {
		didInfos[i] = DIDInfo{
			DID:       did,
			Method:    extractDIDMethod(did),
			IsPrimary: did == primaryDID,
			AddedAt:   user.CreatedAt(), // Simplified; in real impl would track per-DID
		}
	}

	return &GetUserDIDsResult{
		UserID:     userID.String(),
		PrimaryDID: primaryDID,
		DIDs:       didInfos,
	}, nil
}

// HandleResolveDID handles the ResolveDID query.
func (h *Handlers) HandleResolveDID(ctx context.Context, q ResolveDID) (*ResolveDIDResult, error) {
	const op = "Handlers.HandleResolveDID"

	// Find user ID by DID
	userID, err := h.userLookup.GetUserIDByDID(ctx, q.DID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &ResolveDIDResult{
		DID:         q.DID,
		UserID:      userID.String(),
		DisplayName: user.DisplayName().String(),
		IsPrimary:   q.DID == user.PrimaryDID(),
	}, nil
}

// ============================================================================
// OAuth Query Handlers
// ============================================================================

// HandleGetLinkedOAuthAccounts handles the GetLinkedOAuthAccounts query.
func (h *Handlers) HandleGetLinkedOAuthAccounts(ctx context.Context, q GetLinkedOAuthAccounts) (*GetLinkedOAuthAccountsResult, error) {
	const op = "Handlers.HandleGetLinkedOAuthAccounts"

	// Parse user ID
	userID := domain.UserIDFromTypesID(q.UserID)

	// Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Map OAuth links to account info
	links := user.OAuthLinks()
	accounts := make([]OAuthAccountInfo, len(links))

	for i, link := range links {
		accounts[i] = OAuthAccountInfo{
			Provider:   link.Subject.Provider().String(),
			ExternalID: link.Subject.ExternalID(),
			Email:      link.Email,
		}
	}

	return &GetLinkedOAuthAccountsResult{
		UserID:   userID.String(),
		Accounts: accounts,
	}, nil
}

// HandleGetOAuthState handles the GetOAuthState query.
func (h *Handlers) HandleGetOAuthState(ctx context.Context, q GetOAuthState) (*GetOAuthStateResult, error) {
	const op = "Handlers.HandleGetOAuthState"

	// Get state record
	record, err := h.oauthStateRepo.GetByState(ctx, q.State)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return &GetOAuthStateResult{
				State: q.State,
				Valid: false,
			}, nil
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	// Check expiration
	expiresAt := time.Unix(record.ExpiresAt, 0)
	valid := time.Now().Before(expiresAt)

	return &GetOAuthStateResult{
		State:       record.State,
		Provider:    record.Provider,
		RedirectURL: record.RedirectURL,
		ExpiresAt:   expiresAt,
		Valid:       valid,
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// mapUserToResult maps a User aggregate to GetUserResult.
func (h *Handlers) mapUserToResult(user *domain.User) *GetUserResult {
	// Map auth methods
	authMethods := user.AuthMethods()
	authMethodStrs := make([]string, len(authMethods))
	for i, m := range authMethods {
		authMethodStrs[i] = m.String()
	}

	// Map OAuth accounts
	links := user.OAuthLinks()
	oauthAccounts := make([]OAuthAccountInfo, len(links))
	for i, link := range links {
		oauthAccounts[i] = OAuthAccountInfo{
			Provider:   link.Subject.Provider().String(),
			ExternalID: link.Subject.ExternalID(),
			Email:      link.Email,
		}
	}

	return &GetUserResult{
		UserID:        user.ID().String(),
		Email:         user.Email().String(),
		DisplayName:   user.DisplayName().String(),
		Status:        user.Status().String(),
		PrimaryDID:    user.PrimaryDID(),
		DIDs:          user.DIDs(),
		AuthMethods:   authMethodStrs,
		OAuthAccounts: oauthAccounts,
		CreatedAt:     user.CreatedAt(),
		UpdatedAt:     user.UpdatedAt(),
	}
}

// mapSessionToResult maps a Session aggregate to GetSessionResult.
func (h *Handlers) mapSessionToResult(session *domain.Session) *GetSessionResult {
	return &GetSessionResult{
		SessionID:  session.ID().String(),
		UserID:     session.UserID().String(),
		AuthMethod: session.AuthMethod().String(),
		Status:     session.Status().String(),
		IPAddress:  session.IPAddress(),
		UserAgent:  session.UserAgent(),
		CreatedAt:  session.CreatedAt(),
		ExpiresAt:  session.ExpiresAt(),
		LastUsedAt: session.LastUsedAt(),
	}
}

// extractDIDMethod extracts the method from a DID string.
// e.g., "did:key:z6Mk..." -> "key"
func extractDIDMethod(did string) string {
	parts := strings.SplitN(did, ":", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

// ============================================================================
// Query Registration
// ============================================================================

// RegisterQueries registers all Identity query handlers with the query bus.
func RegisterQueries(bus *cqrs.InMemoryQueryBus, handlers *Handlers) error {
	registrations := []struct {
		name    string
		handler any
	}{
		{QueryGetUser, handlers},
		{QueryGetUserByEmail, handlers},
		{QueryGetUserByDID, handlers},
		{QueryListUsers, handlers},
		{QuerySearchUsers, handlers},
		{QueryGetUserProfile, handlers},
		{QueryGetSession, handlers},
		{QueryListUserSessions, handlers},
		{QueryGetActiveSessionCount, handlers},
		{QueryValidateSession, handlers},
		{QueryGetUserDIDs, handlers},
		{QueryResolveDID, handlers},
		{QueryGetLinkedOAuthAccounts, handlers},
		{QueryGetOAuthState, handlers},
	}

	for _, r := range registrations {
		if err := bus.Register(r.name, r.handler); err != nil {
			return err
		}
	}

	return nil
}
