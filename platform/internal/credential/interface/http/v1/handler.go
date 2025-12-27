package v1

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/http/request"
	"github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability/log"

	"github.com/0xsj/nexus/platform/internal/credential/command"
	"github.com/0xsj/nexus/platform/internal/credential/query"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for credentials.
type Handler struct {
	commandBus cqrs.CommandBus
	queryBus   cqrs.QueryBus
	logger     log.Logger
}

// NewHandler creates a new credential HTTP handler.
func NewHandler(commandBus cqrs.CommandBus, queryBus cqrs.QueryBus, logger log.Logger) *Handler {
	return &Handler{
		commandBus: commandBus,
		queryBus:   queryBus,
		logger:     logger,
	}
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetCredential handles GET /credentials/{id}
func (h *Handler) GetCredential(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get credential ID from path
	credentialID, err := request.PathParamRequired(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Create and dispatch query
	q := query.NewGetCredential(credentialID)
	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert result
	view, ok := result.(*query.CredentialView)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected query result type"))
		return
	}

	response.OK(w, FromCredentialView(view))
}

// ListCredentials handles GET /credentials
func (h *Handler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	params := DefaultListCredentialsParams()
	params.Status = request.QueryParam(r, "status")
	params.CredentialType = request.QueryParam(r, "credential_type")

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
	params.SortOrder = request.QueryParamDefault(r, "sort_order", params.SortOrder)

	// Validate parameters
	if err := params.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Create and dispatch query
	q := query.NewListCredentials().
		WithStatus(params.Status).
		WithCredentialType(params.CredentialType).
		WithLimit(params.Limit).
		WithOffset(params.Offset).
		WithSort(params.SortBy, params.SortOrder)

	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert result
	listResult, ok := result.(*query.CredentialListResult)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected query result type"))
		return
	}

	response.OK(w, FromCredentialListResult(listResult))
}

// ListCredentialsByHolder handles GET /holders/{holder_did}/credentials
func (h *Handler) ListCredentialsByHolder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get holder DID from path
	holderDID, err := request.PathParamRequired(r, "holder_did")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Parse query parameters
	params := DefaultListCredentialsParams()
	params.Status = request.QueryParam(r, "status")
	params.CredentialType = request.QueryParam(r, "credential_type")

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

	// Create and dispatch query
	q := query.NewGetCredentialsByHolder(holderDID).
		WithStatus(params.Status).
		WithCredentialType(params.CredentialType).
		WithLimit(params.Limit).
		WithOffset(params.Offset)

	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert result
	listResult, ok := result.(*query.CredentialListResult)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected query result type"))
		return
	}

	response.OK(w, FromCredentialListResult(listResult))
}

// ListCredentialsByIssuer handles GET /issuers/{issuer_did}/credentials
func (h *Handler) ListCredentialsByIssuer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get issuer DID from path
	issuerDID, err := request.PathParamRequired(r, "issuer_did")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Parse query parameters
	params := DefaultListCredentialsParams()
	params.Status = request.QueryParam(r, "status")
	params.CredentialType = request.QueryParam(r, "credential_type")

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

	// Create and dispatch query
	q := query.NewGetCredentialsByIssuer(issuerDID).
		WithStatus(params.Status).
		WithCredentialType(params.CredentialType).
		WithLimit(params.Limit).
		WithOffset(params.Offset)

	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert result
	listResult, ok := result.(*query.CredentialListResult)
	if !ok {
		response.InternalError(w, response.ErrInternal("unexpected query result type"))
		return
	}

	response.OK(w, FromCredentialListResult(listResult))
}

// ============================================================================
// Command Handlers
// ============================================================================

// IssueCredential handles POST /credentials
func (h *Handler) IssueCredential(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Decode request body
	var req IssueCredentialRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Generate credential ID if not provided
	credentialID := req.CredentialID
	if credentialID == "" {
		credentialID = uuid.New().String()
	}

	// Create command
	cmd := command.NewIssueCredential(
		credentialID,
		req.HolderDID,
		req.IssuerDID,
		req.CredentialType,
		req.Claims,
	)

	if req.ExpiresAt != nil {
		cmd.WithExpiration(*req.ExpiresAt)
	}
	if req.SchemaID != "" {
		cmd.WithSchemaID(req.SchemaID)
	}

	// Dispatch command
	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	response.Created(w, NewCredentialIssuedResponse(result.ID, result.Version))
}

// RequestCredential handles POST /credentials/request
func (h *Handler) RequestCredential(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Decode request body
	var req RequestCredentialRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Generate credential ID if not provided
	credentialID := req.CredentialID
	if credentialID == "" {
		credentialID = uuid.New().String()
	}

	// Create command
	cmd := command.NewRequestCredential(
		credentialID,
		req.HolderDID,
		req.IssuerDID,
		req.CredentialType,
		req.Claims,
	)

	// Dispatch command
	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	response.Created(w, NewCredentialCreatedResponse(result.ID, result.Version))
}

// RevokeCredential handles POST /credentials/{id}/revoke
func (h *Handler) RevokeCredential(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get credential ID from path
	credentialID, err := request.PathParamRequired(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Decode request body
	var req RevokeCredentialRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Create command
	cmd := command.NewRevokeCredential(credentialID, req.RevokedBy, req.Reason)

	// Dispatch command
	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	response.OK(w, NewCredentialRevokedResponse(result.ID, result.Version))
}

// SuspendCredential handles POST /credentials/{id}/suspend
func (h *Handler) SuspendCredential(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get credential ID from path
	credentialID, err := request.PathParamRequired(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Decode request body
	var req SuspendCredentialRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Create command
	cmd := command.NewSuspendCredential(credentialID, req.SuspendedBy, req.Reason)
	if req.Until != nil {
		cmd.WithUntil(*req.Until)
	}

	// Dispatch command
	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	response.OK(w, NewCredentialSuspendedResponse(result.ID, result.Version))
}

// ReinstateCredential handles POST /credentials/{id}/reinstate
func (h *Handler) ReinstateCredential(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get credential ID from path
	credentialID, err := request.PathParamRequired(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}

	// Decode request body
	var req ReinstateCredentialRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	// Create command
	cmd := command.NewReinstateCredential(credentialID, req.ReinstatedBy, req.Reason)

	// Dispatch command
	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		WriteError(w, err)
		return
	}

	response.OK(w, NewCredentialReinstatedResponse(result.ID, result.Version))
}
