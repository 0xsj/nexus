package query

import (
	"context"
	"strings"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
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
func (h *Handlers) HandleGetUser(ctx context.Context, q GetUser) (*UserView, error) {
	const op = "Handlers.HandleGetUser"

	userID := domain.UserIDFromTypesID(q.UserID)

	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapUserToView(user), nil
}

// HandleGetUserByEmail handles the GetUserByEmail query.
func (h *Handlers) HandleGetUserByEmail(ctx context.Context, q GetUserByEmail) (*UserView, error) {
	const op = "Handlers.HandleGetUserByEmail"

	userID, err := h.userLookup.GetUserIDByEmail(ctx, q.Email)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapUserToView(user), nil
}

// HandleGetUserByDID handles the GetUserByDID query.
func (h *Handlers) HandleGetUserByDID(ctx context.Context, q GetUserByDID) (*UserView, error) {
	const op = "Handlers.HandleGetUserByDID"

	userID, err := h.userLookup.GetUserIDByDID(ctx, q.DID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapUserToView(user), nil
}

// HandleGetUserProfile handles the GetUserProfile query.
func (h *Handlers) HandleGetUserProfile(ctx context.Context, q GetUserProfile) (*UserProfileView, error) {
	const op = "Handlers.HandleGetUserProfile"

	userID := domain.UserIDFromTypesID(q.UserID)

	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	if !user.IsActive() {
		return nil, domain.UserNotFound(op, userID.String())
	}

	return &UserProfileView{
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
func (h *Handlers) HandleGetSession(ctx context.Context, q GetSession) (*SessionView, error) {
	const op = "Handlers.HandleGetSession"

	sessionID := domain.SessionIDFromTypesID(q.SessionID)

	session, err := h.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapSessionToView(session), nil
}

// HandleListUserSessions handles the ListUserSessions query.
func (h *Handlers) HandleListUserSessions(ctx context.Context, q ListUserSessions) (*SessionListView, error) {
	const op = "Handlers.HandleListUserSessions"

	userID := domain.UserIDFromTypesID(q.UserID)

	sessions, err := h.sessionLookup.GetActiveSessionsForUser(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]SessionSummaryView, 0, len(sessions))
	for _, session := range sessions {
		if q.ActiveOnly && !session.IsActive() {
			continue
		}
		summaries = append(summaries, SessionSummaryView{
			SessionID:  session.ID().String(),
			AuthMethod: session.AuthMethod().String(),
			Status:     session.Status().String(),
			IPAddress:  session.IPAddress(),
			UserAgent:  session.UserAgent(),
			CreatedAt:  session.CreatedAt(),
			ExpiresAt:  session.ExpiresAt(),
			LastUsedAt: session.UpdatedAt(),
		})
	}

	return &SessionListView{
		Sessions:   summaries,
		NextCursor: nil,
		TotalCount: len(summaries),
	}, nil
}

// HandleGetActiveSessionCount handles the GetActiveSessionCount query.
func (h *Handlers) HandleGetActiveSessionCount(ctx context.Context, q GetActiveSessionCount) (*SessionCountView, error) {
	const op = "Handlers.HandleGetActiveSessionCount"

	userID := domain.UserIDFromTypesID(q.UserID)

	count, err := h.sessionLookup.CountActiveSessionsForUser(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &SessionCountView{
		UserID: userID.String(),
		Count:  count,
	}, nil
}

// HandleValidateSession handles the ValidateSession query.
func (h *Handlers) HandleValidateSession(ctx context.Context, q ValidateSession) (*SessionValidationView, error) {
	const op = "Handlers.HandleValidateSession"

	sessionID := domain.SessionIDFromTypesID(q.SessionID)

	session, err := h.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return &SessionValidationView{
				Valid:  false,
				Reason: "session not found",
			}, nil
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	token, err := domain.ParseToken(q.Token)
	if err != nil {
		return &SessionValidationView{
			Valid:  false,
			Reason: "invalid token format",
		}, nil
	}

	if !session.VerifyToken(token) {
		return &SessionValidationView{
			Valid:  false,
			Reason: "token mismatch",
		}, nil
	}

	if !session.IsActive() {
		return &SessionValidationView{
			Valid:  false,
			Reason: "session is " + session.Status().String(),
		}, nil
	}

	if session.IsExpired() {
		return &SessionValidationView{
			Valid:  false,
			Reason: "session expired",
		}, nil
	}

	user, err := h.userRepo.Get(ctx, session.UserID())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return &SessionValidationView{
				Valid:  false,
				Reason: "user not found",
			}, nil
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	if !user.CanAuthenticate() {
		return &SessionValidationView{
			Valid:  false,
			Reason: "user is " + user.Status().String(),
		}, nil
	}

	return &SessionValidationView{
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
func (h *Handlers) HandleGetUserDIDs(ctx context.Context, q GetUserDIDs) (*UserDIDsView, error) {
	const op = "Handlers.HandleGetUserDIDs"

	userID := domain.UserIDFromTypesID(q.UserID)

	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	dids := user.DIDs()
	primaryDID := user.PrimaryDID()
	didViews := make([]DIDView, len(dids))

	for i, did := range dids {
		didViews[i] = DIDView{
			DID:       did,
			Method:    extractDIDMethod(did),
			IsPrimary: did == primaryDID,
			AddedAt:   user.CreatedAt(),
		}
	}

	return &UserDIDsView{
		UserID:     userID.String(),
		PrimaryDID: primaryDID,
		DIDs:       didViews,
	}, nil
}

// HandleResolveDID handles the ResolveDID query.
func (h *Handlers) HandleResolveDID(ctx context.Context, q ResolveDID) (*DIDResolutionView, error) {
	const op = "Handlers.HandleResolveDID"

	userID, err := h.userLookup.GetUserIDByDID(ctx, q.DID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &DIDResolutionView{
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
func (h *Handlers) HandleGetLinkedOAuthAccounts(ctx context.Context, q GetLinkedOAuthAccounts) (*LinkedOAuthAccountsView, error) {
	const op = "Handlers.HandleGetLinkedOAuthAccounts"

	userID := domain.UserIDFromTypesID(q.UserID)

	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	links := user.OAuthLinks()
	accounts := make([]OAuthAccountView, len(links))

	for i, link := range links {
		accounts[i] = OAuthAccountView{
			Provider:   link.Provider().String(),
			ExternalID: link.ExternalID(),
			Email:      "",
		}
	}

	return &LinkedOAuthAccountsView{
		UserID:   userID.String(),
		Accounts: accounts,
	}, nil
}

// HandleGetOAuthState handles the GetOAuthState query.
func (h *Handlers) HandleGetOAuthState(ctx context.Context, q GetOAuthState) (*OAuthStateView, error) {
	const op = "Handlers.HandleGetOAuthState"

	record, err := h.oauthStateRepo.GetByState(ctx, q.State)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return &OAuthStateView{
				State: q.State,
				Valid: false,
			}, nil
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	expiresAt := time.Unix(record.ExpiresAt, 0)
	valid := time.Now().Before(expiresAt)

	return &OAuthStateView{
		State:       record.State,
		Provider:    record.Provider,
		RedirectURL: record.RedirectURL,
		ExpiresAt:   expiresAt,
		Valid:       valid,
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapUserToView maps a User aggregate to UserView.
func (h *Handlers) mapUserToView(user *domain.User) *UserView {
	links := user.OAuthLinks()
	oauthAccounts := make([]OAuthAccountView, len(links))
	for i, link := range links {
		oauthAccounts[i] = OAuthAccountView{
			Provider:   link.Provider().String(),
			ExternalID: link.ExternalID(),
			Email:      "",
		}
	}

	authMethods := h.deriveAuthMethods(user)

	return &UserView{
		UserID:        user.ID().String(),
		Email:         user.Email().String(),
		DisplayName:   user.DisplayName().String(),
		Status:        user.Status().String(),
		PrimaryDID:    user.PrimaryDID(),
		DIDs:          user.DIDs(),
		AuthMethods:   authMethods,
		OAuthAccounts: oauthAccounts,
		CreatedAt:     user.CreatedAt(),
		UpdatedAt:     user.UpdatedAt(),
	}
}

// deriveAuthMethods determines available auth methods based on user data.
func (h *Handlers) deriveAuthMethods(user *domain.User) []string {
	methods := make([]string, 0, 3)

	if !user.Email().IsEmpty() {
		methods = append(methods, domain.AuthMethodMagicLink.String())
	}

	for _, did := range user.DIDs() {
		if strings.HasPrefix(did, "did:pkh:") {
			methods = append(methods, domain.AuthMethodWallet.String())
			break
		}
	}

	if len(user.OAuthLinks()) > 0 {
		methods = append(methods, domain.AuthMethodOAuth.String())
	}

	return methods
}

// mapSessionToView maps a Session aggregate to SessionView.
func (h *Handlers) mapSessionToView(session *domain.Session) *SessionView {
	return &SessionView{
		SessionID:  session.ID().String(),
		UserID:     session.UserID().String(),
		AuthMethod: session.AuthMethod().String(),
		Status:     session.Status().String(),
		IPAddress:  session.IPAddress(),
		UserAgent:  session.UserAgent(),
		CreatedAt:  session.CreatedAt(),
		ExpiresAt:  session.ExpiresAt(),
		LastUsedAt: session.UpdatedAt(),
	}
}

// extractDIDMethod extracts the method from a DID string.
func extractDIDMethod(did string) string {
	parts := strings.SplitN(did, ":", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}
