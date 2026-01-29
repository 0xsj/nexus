package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/verification/application/command"
	"github.com/0xsj/nexus/platform/internal/verification/application/query"
	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/http/request"
	"github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/id"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for verifications.
type Handler struct {
	commandBus  cqrs.CommandBus
	queryBus    cqrs.QueryBus
	idGenerator id.Generator
	logger      log.Logger
}

// NewHandler creates a new verification HTTP handler.
func NewHandler(
	commandBus cqrs.CommandBus,
	queryBus cqrs.QueryBus,
	idGenerator id.Generator,
	logger log.Logger,
) *Handler {
	return &Handler{
		commandBus:  commandBus,
		queryBus:    queryBus,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetVerification handles GET /verifications/{id}
func (h *Handler) GetVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get verification ID from path
	verificationID, err := request.PathParamRequired(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Create and dispatch query
	q := query.NewGetVerification(verificationID)
	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert result
	view, ok := result.(*domain.VerificationView)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected query result type"))
		return
	}

	response.OK(w, FromVerificationView(view))
}

// ListVerifications handles GET /verifications
func (h *Handler) ListVerifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID := log.UserIDFromContext(ctx)
	if userID == "" {
		WriteError(w, domain.ErrVerificationNotFound("ListVerifications", "user not authenticated"))
		return
	}

	// Parse query parameters
	params := DefaultListVerificationsParams()
	params.Provider = request.QueryParam(r, "provider")
	params.Status = request.QueryParam(r, "status")

	limit, err := request.QueryParamInt(r, "limit", params.Limit)
	if err != nil {
		WriteError(w, err)
		return
	}
	params.Limit = limit

	offset, err := request.QueryParamInt(r, "offset", params.Offset)
	if err != nil {
		WriteError(w, err)
		return
	}
	params.Offset = offset

	params.SortBy = request.QueryParamDefault(r, "sort_by", params.SortBy)

	sortDesc, err := request.QueryParamBool(r, "sort_desc", params.SortDesc)
	if err != nil {
		WriteError(w, err)
		return
	}
	params.SortDesc = sortDesc

	// Validate parameters
	if err := params.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Create and dispatch query
	q := query.NewListVerificationsByUser(userID)
	if params.Provider != "" {
		q.WithProvider(domain.Provider(params.Provider))
	}
	if params.Status != "" {
		q.WithStatus(domain.VerificationStatus(params.Status))
	}
	q.WithLimit(params.Limit).
		WithOffset(params.Offset).
		WithSort(params.SortBy, params.SortDesc)

	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert result
	listResult, ok := result.(*query.VerificationListResult)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected query result type"))
		return
	}

	response.OK(w, FromVerificationSummaries(listResult.Items, listResult.Total, listResult.Limit, listResult.Offset))
}

// GetProviderConnections handles GET /verifications/connections
func (h *Handler) GetProviderConnections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID := log.UserIDFromContext(ctx)
	if userID == "" {
		WriteError(w, domain.ErrVerificationNotFound("GetProviderConnections", "user not authenticated"))
		return
	}

	// Create and dispatch query
	q := query.NewGetProviderConnections(userID)
	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert result
	connections, ok := result.([]*domain.ProviderConnectionView)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected query result type"))
		return
	}

	response.OK(w, FromProviderConnectionViews(connections))
}

// CheckProviderConnected handles GET /verifications/connections/{provider}
func (h *Handler) CheckProviderConnected(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID := log.UserIDFromContext(ctx)
	if userID == "" {
		WriteError(w, domain.ErrVerificationNotFound("CheckProviderConnected", "user not authenticated"))
		return
	}

	// Get provider from path
	providerStr, err := request.PathParamRequired(r, "provider")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Validate provider
	params := CheckProviderConnectionParams{Provider: providerStr}
	if err := params.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Create and dispatch query
	q := query.NewCheckProviderConnected(userID, domain.Provider(providerStr))
	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert result
	checkResult, ok := result.(*query.CheckProviderConnectedResult)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected query result type"))
		return
	}

	response.OK(w, NewCheckProviderConnectedResponse(
		providerStr,
		checkResult.Connected,
		checkResult.CredentialID,
		checkResult.Username,
	))
}

// ============================================================================
// Command Handlers
// ============================================================================

// InitiateVerification handles POST /verifications
func (h *Handler) InitiateVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID := log.UserIDFromContext(ctx)
	if userID == "" {
		WriteError(w, domain.ErrVerificationNotFound("InitiateVerification", "user not authenticated"))
		return
	}

	// Decode request body
	var req InitiateVerificationRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Generate verification ID
	verificationID := h.idGenerator.Generate()

	// Create command
	cmd := command.NewInitiateVerification(
		verificationID.String(),
		userID,
		req.ToProvider(),
		req.ToCredentialType(),
	)
	if req.RedirectURL != "" {
		cmd.WithRedirectURL(req.RedirectURL)
	}

	// Dispatch command
	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Extract result data
	initiateResult, ok := result.Data.(*command.InitiateVerificationResult)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected command result type"))
		return
	}

	response.Created(w, NewInitiateVerificationResponse(
		initiateResult.VerificationID,
		initiateResult.AuthorizationURL,
		initiateResult.ExpiresIn,
	))
}

// HandleOAuthCallback handles GET /verifications/callback
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract OAuth callback parameters
	state := request.QueryParam(r, "state")
	code := request.QueryParam(r, "code")
	oauthError := request.QueryParam(r, "error")

	// Validate required parameters
	if state == "" {
		WriteError(w, domain.ErrOAuthStateMismatch("HandleOAuthCallback"))
		return
	}

	// Create command
	cmd := command.NewHandleOAuthCallback(state, code)
	if oauthError != "" {
		cmd.WithError(oauthError)
	}

	// Dispatch command
	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Extract result data
	callbackResult, ok := result.Data.(*command.HandleOAuthCallbackResult)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected command result type"))
		return
	}

	// If there's a redirect URL, redirect the user
	if callbackResult.RedirectURL != "" {
		http.Redirect(w, r, callbackResult.RedirectURL, http.StatusFound)
		return
	}

	// Otherwise return JSON response
	if callbackResult.Success {
		response.OK(w, NewOAuthCallbackSuccessResponse(callbackResult.VerificationID, callbackResult.RedirectURL))
	} else {
		response.OK(w, NewOAuthCallbackErrorResponse(callbackResult.VerificationID, callbackResult.Error))
	}
}

// CancelVerification handles POST /verifications/{id}/cancel
func (h *Handler) CancelVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID := log.UserIDFromContext(ctx)
	if userID == "" {
		WriteError(w, domain.ErrVerificationNotFound("CancelVerification", "user not authenticated"))
		return
	}

	// Get verification ID from path
	verificationID, err := request.PathParamRequired(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Decode request body (optional)
	var req CancelVerificationRequest
	if r.ContentLength > 0 {
		if err := request.DecodeJSON(r, &req); err != nil {
			WriteError(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			WriteError(w, err)
			return
		}
	}

	// Create command
	cmd := command.NewCancelVerification(verificationID, userID)
	if req.Reason != "" {
		cmd.WithReason(req.Reason)
	}

	// Dispatch command
	_, err = h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	response.OK(w, NewVerificationCancelledResponse(verificationID))
}

// ============================================================================
// Provider-specific Initiation (Convenience Endpoints)
// ============================================================================

// InitiateGitHubVerification handles POST /verifications/github
func (h *Handler) InitiateGitHubVerification(w http.ResponseWriter, r *http.Request) {
	h.initiateProviderVerification(w, r, domain.ProviderGitHub, domain.CredentialTypeGitHubContributor)
}

// InitiateLinkedInVerification handles POST /verifications/linkedin
func (h *Handler) InitiateLinkedInVerification(w http.ResponseWriter, r *http.Request) {
	h.initiateProviderVerification(w, r, domain.ProviderLinkedIn, domain.CredentialTypeLinkedInEmployment)
}

// initiateProviderVerification is a helper for provider-specific endpoints.
func (h *Handler) initiateProviderVerification(
	w http.ResponseWriter,
	r *http.Request,
	provider domain.Provider,
	credentialType domain.CredentialType,
) {
	ctx := r.Context()

	// Get user ID from context
	userID := log.UserIDFromContext(ctx)
	if userID == "" {
		WriteError(w, domain.ErrVerificationNotFound("initiateProviderVerification", "user not authenticated"))
		return
	}

	// Parse optional redirect URL from query or body
	redirectURL := request.QueryParam(r, "redirect_url")

	// Generate verification ID
	verificationID := h.idGenerator.Generate()

	// Create command
	cmd := command.NewInitiateVerification(
		verificationID.String(),
		userID,
		provider,
		credentialType,
	)
	if redirectURL != "" {
		cmd.WithRedirectURL(redirectURL)
	}

	// Dispatch command
	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Extract result data
	initiateResult, ok := result.Data.(*command.InitiateVerificationResult)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected command result type"))
		return
	}

	response.Created(w, NewInitiateVerificationResponse(
		initiateResult.VerificationID,
		initiateResult.AuthorizationURL,
		initiateResult.ExpiresIn,
	))
}
