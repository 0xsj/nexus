package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/identity/app/command"
	"github.com/0xsj/nexus/platform/internal/identity/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Identity context.
type Handler struct {
	commands *command.Handlers
	queries  *query.Handlers
}

// NewHandler creates a new Handler.
func NewHandler(
	commands *command.Handlers,
	queries *query.Handlers,
) *Handler {
	return &Handler{
		commands: commands,
		queries:  queries,
	}
}

// ============================================================================
// Auth - Magic Link
// ============================================================================

// SendMagicLink handles POST /auth/magic-link
func (h *Handler) SendMagicLink(w http.ResponseWriter, r *http.Request) {
	var req SendMagicLinkRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	email, err := types.NewEmail(req.Email)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid email format"))
		return
	}

	result, err := h.commands.HandleRequestMagicLink(r.Context(), command.RequestMagicLink{
		Email: email,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RequestMagicLinkResult)
	expiresIn := int(time.Until(data.ExpiresAt).Seconds())

	h.writeJSON(w, http.StatusOK, MagicLinkSentResponse{
		Message:   "Magic link sent successfully",
		Email:     data.Email,
		ExpiresIn: expiresIn,
	})
}

// VerifyMagicLink handles POST /auth/magic-link/verify
func (h *Handler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	var req VerifyMagicLinkRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	meta := h.extractMetadata(r)

	result, err := h.commands.HandleVerifyMagicLink(r.Context(), command.VerifyMagicLink{
		Token:     req.Token,
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.VerifyMagicLinkResult)
	h.writeJSON(w, http.StatusOK, h.toAuthResponseFromMagicLink(data))
}

// ============================================================================
// Auth - Wallet (SIWE)
// ============================================================================

// VerifyWallet handles POST /auth/wallet/verify
func (h *Handler) VerifyWallet(w http.ResponseWriter, r *http.Request) {
	var req RegisterWithWalletRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	meta := h.extractMetadata(r)

	result, err := h.commands.HandleAuthenticateWithWallet(r.Context(), command.AuthenticateWithWallet{
		Address:   req.Address,
		Message:   req.Message,
		Signature: req.Signature,
		ChainID:   req.ChainID,
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.AuthenticateWithWalletResult)
	h.writeJSON(w, http.StatusOK, h.toAuthResponseFromWallet(data))
}

// ============================================================================
// Auth - OAuth
// ============================================================================

// InitiateOAuth handles GET /auth/oauth/{provider}
func (h *Handler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("provider is required"))
		return
	}

	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("redirect_url is required"))
		return
	}

	result, err := h.commands.HandleInitiateOAuth(r.Context(), command.InitiateOAuth{
		Provider:    provider,
		RedirectURL: redirectURL,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.InitiateOAuthResult)
	h.writeJSON(w, http.StatusOK, OAuthInitiatedResponse{
		AuthURL:   data.AuthURL,
		State:     data.State,
		ExpiresIn: int(time.Until(data.ExpiresAt).Seconds()),
	})
}

// OAuthCallback handles POST /auth/oauth/{provider}/callback
func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("provider is required"))
		return
	}

	var req OAuthCallbackRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Code == "" || req.State == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("code and state are required"))
		return
	}

	meta := h.extractMetadata(r)

	result, err := h.commands.HandleAuthenticateWithOAuth(r.Context(), command.AuthenticateWithOAuth{
		Provider:  provider,
		Code:      req.Code,
		State:     req.State,
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.AuthenticateWithOAuthResult)
	h.writeJSON(w, http.StatusOK, h.toAuthResponseFromOAuth(data))
}

// ============================================================================
// Session Management
// ============================================================================

// RefreshSession handles POST /auth/refresh
func (h *Handler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	var req RefreshSessionRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	sessionIDStr := h.getSessionIDFromContext(r)
	if sessionIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	parsedSessionID, err := types.ParseID(sessionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid session id"))
		return
	}

	result, err := h.commands.HandleRefreshSession(r.Context(), command.RefreshSession{
		SessionID: parsedSessionID,
		Token:     req.RefreshToken,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RefreshSessionResult)
	h.writeJSON(w, http.StatusOK, AuthResponse{
		AccessToken: data.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Until(data.ExpiresAt).Seconds()),
		ExpiresAt:   data.ExpiresAt,
	})
}

// RevokeSession handles POST /auth/revoke
func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	var req RevokeSessionRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Get session ID from request or from auth context
	sessionIDStr := req.SessionID
	if sessionIDStr == "" {
		sessionIDStr = h.getSessionIDFromContext(r)
	}

	if sessionIDStr == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("session_id is required"))
		return
	}

	sessionID, err := types.ParseID(sessionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid session id"))
		return
	}

	var reason *string
	if req.Reason != "" {
		reason = &req.Reason
	}

	result, err := h.commands.HandleRevokeSession(r.Context(), command.RevokeSession{
		SessionID: sessionID,
		Reason:    reason,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RevokeSessionResult)
	h.writeJSON(w, http.StatusOK, SessionRevokedResponse{
		Message:   "Session revoked successfully",
		SessionID: data.SessionID,
	})
}

// Logout handles POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := h.getSessionIDFromContext(r)
	if sessionIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	sessionID, err := types.ParseID(sessionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid session id"))
		return
	}

	reason := "user logout"
	_, err = h.commands.HandleRevokeSession(r.Context(), command.RevokeSession{
		SessionID: sessionID,
		Reason:    &reason,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, MessageResponse{
		Message: "Logged out successfully",
	})
}

// ListSessions handles GET /sessions
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.queries.HandleListUserSessions(r.Context(), query.ListUserSessions{
		UserID:     parsedUserID,
		ActiveOnly: r.URL.Query().Get("active_only") == "true",
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	currentSessionID := h.getSessionIDFromContext(r)
	h.writeJSON(w, http.StatusOK, h.toSessionListResponse(result, currentSessionID))
}

// ============================================================================
// User Profile
// ============================================================================

// GetCurrentUser handles GET /me
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.queries.HandleGetUser(r.Context(), query.GetUser{
		UserID: parsedUserID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, h.toUserProfileResponse(result))
}

// GetUser handles GET /users/{userId}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	if userIDStr == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user id is required"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.queries.HandleGetUser(r.Context(), query.GetUser{
		UserID: parsedUserID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, h.toUserResponse(result))
}

// GetUserProfile handles GET /users/{userId}/profile
func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	if userIDStr == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user id is required"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.queries.HandleGetUserProfile(r.Context(), query.GetUserProfile{
		UserID: parsedUserID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, PublicProfileResponse{
		UserID:      result.UserID,
		DisplayName: result.DisplayName,
		PrimaryDID:  result.PrimaryDID,
		CreatedAt:   result.CreatedAt,
	})
}

// UpdateDisplayName handles PATCH /me/display-name
func (h *Handler) UpdateDisplayName(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	var req UpdateDisplayNameRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.commands.HandleChangeDisplayName(r.Context(), command.ChangeDisplayName{
		UserID:         parsedUserID,
		NewDisplayName: req.DisplayName,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ChangeDisplayNameResult)
	h.writeJSON(w, http.StatusOK, DisplayNameUpdatedResponse{
		Message:     "Display name updated successfully",
		DisplayName: data.DisplayName,
	})
}

// UpdateEmail handles PATCH /me/email
func (h *Handler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	var req UpdateEmailRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	email, err := types.NewEmail(req.Email)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid email format"))
		return
	}

	result, err := h.commands.HandleChangeEmail(r.Context(), command.ChangeEmail{
		UserID:   parsedUserID,
		NewEmail: email,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ChangeEmailResult)
	h.writeJSON(w, http.StatusOK, EmailUpdatedResponse{
		Message: "Email updated successfully",
		Email:   data.Email,
	})
}

// ============================================================================
// DID Management
// ============================================================================

// GetUserDIDs handles GET /me/dids
func (h *Handler) GetUserDIDs(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.queries.HandleGetUserDIDs(r.Context(), query.GetUserDIDs{
		UserID: parsedUserID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, h.toDIDListResponse(result))
}

// AddDID handles POST /me/dids
func (h *Handler) AddDID(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	var req AddDIDRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.commands.HandleAddDID(r.Context(), command.AddDID{
		UserID: parsedUserID,
		DID:    req.DID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.AddDIDResult)
	h.writeJSON(w, http.StatusCreated, DIDAddedResponse{
		Message: "DID added successfully",
		DID:     data.DID,
	})
}

// RemoveDID handles DELETE /me/dids/{did}
func (h *Handler) RemoveDID(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	did := chi.URLParam(r, "did")
	if did == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("did is required"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.commands.HandleRemoveDID(r.Context(), command.RemoveDID{
		UserID: parsedUserID,
		DID:    did,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RemoveDIDResult)
	h.writeJSON(w, http.StatusOK, DIDRemovedResponse{
		Message: "DID removed successfully",
		DID:     data.DID,
	})
}

// ResolveDID handles GET /dids/{did}
func (h *Handler) ResolveDID(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	if did == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("did is required"))
		return
	}

	result, err := h.queries.HandleResolveDID(r.Context(), query.ResolveDID{
		DID: did,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, DIDResolutionResponse{
		DID:         result.DID,
		UserID:      result.UserID,
		DisplayName: result.DisplayName,
		IsPrimary:   result.IsPrimary,
	})
}

// ============================================================================
// OAuth Link Management
// ============================================================================

// GetLinkedOAuthAccounts handles GET /me/oauth
func (h *Handler) GetLinkedOAuthAccounts(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.queries.HandleGetLinkedOAuthAccounts(r.Context(), query.GetLinkedOAuthAccounts{
		UserID: parsedUserID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, h.toOAuthAccountsResponse(result))
}

// InitiateLinkOAuth handles GET /me/oauth/{provider}
func (h *Handler) InitiateLinkOAuth(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	provider := chi.URLParam(r, "provider")
	if provider == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("provider is required"))
		return
	}

	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("redirect_url is required"))
		return
	}

	result, err := h.commands.HandleInitiateOAuth(r.Context(), command.InitiateOAuth{
		Provider:    provider,
		RedirectURL: redirectURL,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.InitiateOAuthResult)
	h.writeJSON(w, http.StatusOK, OAuthInitiatedResponse{
		AuthURL:   data.AuthURL,
		State:     data.State,
		ExpiresIn: int(time.Until(data.ExpiresAt).Seconds()),
	})
}

// LinkOAuthCallback handles POST /me/oauth/{provider}/callback
func (h *Handler) LinkOAuthCallback(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	provider := chi.URLParam(r, "provider")
	if provider == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("provider is required"))
		return
	}

	var req OAuthCallbackRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Code == "" || req.State == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("code and state are required"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.commands.HandleLinkOAuthAccount(r.Context(), command.LinkOAuthAccount{
		UserID:   parsedUserID,
		Provider: provider,
		Code:     req.Code,
		State:    req.State,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.LinkOAuthAccountResult)
	h.writeJSON(w, http.StatusOK, OAuthLinkedResponse{
		Message:  "OAuth account linked successfully",
		Provider: data.Provider,
	})
}

// UnlinkOAuth handles DELETE /me/oauth/{provider}
func (h *Handler) UnlinkOAuth(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	provider := chi.URLParam(r, "provider")
	if provider == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("provider is required"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	result, err := h.commands.HandleUnlinkOAuthAccount(r.Context(), command.UnlinkOAuthAccount{
		UserID:   parsedUserID,
		Provider: provider,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.UnlinkOAuthAccountResult)
	h.writeJSON(w, http.StatusOK, OAuthUnlinkedResponse{
		Message:  "OAuth account unlinked successfully",
		Provider: data.Provider,
	})
}

// ============================================================================
// Account Management
// ============================================================================

// DeleteAccount handles DELETE /me
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getUserIDFromContext(r)
	if userIDStr == "" {
		h.writeError(w, http.StatusUnauthorized, UnauthorizedResponse("not authenticated"))
		return
	}

	parsedUserID, err := types.ParseID(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid user id"))
		return
	}

	_, err = h.commands.HandleDeleteUser(r.Context(), command.DeleteUser{
		UserID: parsedUserID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, AccountDeletedResponse{
		Message: "Account deleted successfully",
	})
}

// ============================================================================
// Health
// ============================================================================

// Health handles GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
	})
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

func (h *Handler) toAuthResponseFromMagicLink(data command.VerifyMagicLinkResult) AuthResponse {
	return AuthResponse{
		AccessToken: data.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Until(data.ExpiresAt).Seconds()),
		ExpiresAt:   data.ExpiresAt,
		User: &UserResponse{
			ID: data.UserID,
		},
	}
}

func (h *Handler) toAuthResponseFromWallet(data command.AuthenticateWithWalletResult) AuthResponse {
	return AuthResponse{
		AccessToken: data.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Until(data.ExpiresAt).Seconds()),
		ExpiresAt:   data.ExpiresAt,
		User: &UserResponse{
			ID:         data.UserID,
			PrimaryDID: data.DID,
		},
	}
}

func (h *Handler) toAuthResponseFromOAuth(data command.AuthenticateWithOAuthResult) AuthResponse {
	return AuthResponse{
		AccessToken: data.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Until(data.ExpiresAt).Seconds()),
		ExpiresAt:   data.ExpiresAt,
		User: &UserResponse{
			ID:    data.UserID,
			Email: data.Email,
		},
	}
}

func (h *Handler) toUserResponse(view *query.UserView) UserResponse {
	return UserResponse{
		ID:          view.UserID,
		Email:       view.Email,
		DisplayName: view.DisplayName,
		Status:      view.Status,
		PrimaryDID:  view.PrimaryDID,
		CreatedAt:   view.CreatedAt,
		UpdatedAt:   view.UpdatedAt,
	}
}

func (h *Handler) toUserProfileResponse(view *query.UserView) UserProfileResponse {
	oauthAccounts := make([]OAuthAccountResponse, len(view.OAuthAccounts))
	for i, acc := range view.OAuthAccounts {
		oauthAccounts[i] = OAuthAccountResponse{
			Provider:   acc.Provider,
			ExternalID: acc.ExternalID,
			Email:      acc.Email,
		}
	}

	return UserProfileResponse{
		User:          h.toUserResponse(view),
		DIDs:          view.DIDs,
		AuthMethods:   view.AuthMethods,
		OAuthAccounts: oauthAccounts,
	}
}

func (h *Handler) toSessionListResponse(view *query.SessionListView, currentSessionID string) SessionListResponse {
	sessions := make([]SessionResponse, len(view.Sessions))
	for i, s := range view.Sessions {
		sessions[i] = SessionResponse{
			ID:         s.SessionID,
			AuthMethod: s.AuthMethod,
			Status:     s.Status,
			IPAddress:  s.IPAddress,
			UserAgent:  s.UserAgent,
			ExpiresAt:  s.ExpiresAt,
			CreatedAt:  s.CreatedAt,
			IsCurrent:  s.SessionID == currentSessionID,
		}
	}

	return SessionListResponse{
		Sessions:   sessions,
		TotalCount: view.TotalCount,
	}
}

func (h *Handler) toDIDListResponse(view *query.UserDIDsView) DIDListResponse {
	dids := make([]DIDResponse, len(view.DIDs))
	for i, d := range view.DIDs {
		dids[i] = DIDResponse{
			DID:       d.DID,
			Method:    d.Method,
			IsPrimary: d.IsPrimary,
			AddedAt:   d.AddedAt,
		}
	}

	return DIDListResponse{
		UserID:     view.UserID,
		PrimaryDID: view.PrimaryDID,
		DIDs:       dids,
	}
}

func (h *Handler) toOAuthAccountsResponse(view *query.LinkedOAuthAccountsView) OAuthAccountsResponse {
	accounts := make([]OAuthAccountResponse, len(view.Accounts))
	for i, acc := range view.Accounts {
		accounts[i] = OAuthAccountResponse{
			Provider:   acc.Provider,
			ExternalID: acc.ExternalID,
			Email:      acc.Email,
		}
	}

	return OAuthAccountsResponse{
		UserID:   view.UserID,
		Accounts: accounts,
	}
}

// ============================================================================
// Request Helpers
// ============================================================================

func (h *Handler) decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (h *Handler) extractMetadata(r *http.Request) RequestMetadata {
	return NewRequestMetadata(
		h.getClientIP(r),
		r.UserAgent(),
		r.Header.Get("X-Request-ID"),
	)
}

func (h *Handler) getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

func (h *Handler) getUserIDFromContext(r *http.Request) string {
	if userID := r.Context().Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

func (h *Handler) getSessionIDFromContext(r *http.Request) string {
	if sessionID := r.Context().Value("session_id"); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}
	return ""
}

// ============================================================================
// Response Helpers
// ============================================================================

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, errResp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResp)
}
