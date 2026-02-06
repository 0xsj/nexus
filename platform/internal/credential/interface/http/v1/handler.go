package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/credential/app/command"
	"github.com/0xsj/nexus/platform/internal/credential/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Credential context.
type Handler struct {
	commands *command.Handlers
	queries  *query.Handlers
}

// NewHandler creates a new Handler.
func NewHandler(commands *command.Handlers, queries *query.Handlers) *Handler {
	return &Handler{
		commands: commands,
		queries:  queries,
	}
}

// ============================================================================
// Credential Issuance
// ============================================================================

// IssueCredential handles POST /api/v1/credentials
func (h *Handler) IssueCredential(w http.ResponseWriter, r *http.Request) {
	var req IssueCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.CredentialType == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("credential_type", "credential_type is required"))
		return
	}
	if req.IssuerDID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("issuer_did", "issuer_did is required"))
		return
	}
	if req.SubjectDID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("subject_did", "subject_did is required"))
		return
	}
	if len(req.Claims) == 0 {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("claims", "claims are required"))
		return
	}

	// Build command
	cmd := command.IssueCredential{
		CredentialType: req.CredentialType,
		IssuerDID:      req.IssuerDID,
		SubjectDID:     req.SubjectDID,
		Claims:         req.Claims,
		VerificationID: req.VerificationID,
	}
	if req.ExpiresAt != nil {
		cmd.ExpiresAt = *req.ExpiresAt
	}

	// Execute command
	result, err := h.commands.HandleIssueCredential(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.IssueCredentialResult)
	h.writeJSON(w, http.StatusCreated, CredentialIssuedResponse{
		CredentialID:   data.CredentialID,
		CredentialType: data.CredentialType,
		Status:         data.Status,
		IssuedAt:       time.Now().UTC(),
	})
}

// ============================================================================
// Credential Retrieval
// ============================================================================

// GetCredential handles GET /api/v1/credentials/{credentialId}
func (h *Handler) GetCredential(w http.ResponseWriter, r *http.Request) {
	credentialID := chi.URLParam(r, "credentialId")
	if credentialID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("credential_id is required"))
		return
	}

	id, err := types.ParseID(credentialID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("credential_id", "invalid credential_id format"))
		return
	}

	result, err := h.queries.HandleGetCredential(r.Context(), query.GetCredential{
		CredentialID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toCredentialResponse(result))
}

// ============================================================================
// Credential Status Management
// ============================================================================

// RevokeCredential handles POST /api/v1/credentials/{credentialId}/revoke
func (h *Handler) RevokeCredential(w http.ResponseWriter, r *http.Request) {
	credentialID := chi.URLParam(r, "credentialId")
	if credentialID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("credential_id is required"))
		return
	}

	id, err := types.ParseID(credentialID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("credential_id", "invalid credential_id format"))
		return
	}

	var req RevokeCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Reason == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("reason", "reason is required"))
		return
	}
	if len(req.Reason) > 500 {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("reason", "reason must be 500 characters or less"))
		return
	}

	result, err := h.commands.HandleRevokeCredential(r.Context(), command.RevokeCredential{
		CredentialID: id,
		Reason:       req.Reason,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RevokeCredentialResult)
	h.writeJSON(w, http.StatusOK, CredentialRevokedResponse{
		CredentialID: data.CredentialID,
		Status:       data.Status,
		RevokedAt:    time.Now().UTC(),
	})
}

// ExpireCredential handles POST /api/v1/credentials/{credentialId}/expire
func (h *Handler) ExpireCredential(w http.ResponseWriter, r *http.Request) {
	credentialID := chi.URLParam(r, "credentialId")
	if credentialID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("credential_id is required"))
		return
	}

	id, err := types.ParseID(credentialID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("credential_id", "invalid credential_id format"))
		return
	}

	result, err := h.commands.HandleExpireCredential(r.Context(), command.ExpireCredential{
		CredentialID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ExpireCredentialResult)
	h.writeJSON(w, http.StatusOK, CredentialExpiredResponse{
		CredentialID: data.CredentialID,
		Status:       data.Status,
		ExpiredAt:    time.Now().UTC(),
	})
}

// ============================================================================
// Credential Lists
// ============================================================================

// ListCredentialsBySubject handles GET /api/v1/credentials/subject/{subjectDid}
func (h *Handler) ListCredentialsBySubject(w http.ResponseWriter, r *http.Request) {
	subjectDID := chi.URLParam(r, "subjectDid")
	if subjectDID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("subject_did is required"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListCredentialsBySubject(r.Context(), query.ListCredentialsBySubject{
		SubjectDID: subjectDID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toCredentialListResponse(result))
}

// ListCredentialsByIssuer handles GET /api/v1/credentials/issuer/{issuerDid}
func (h *Handler) ListCredentialsByIssuer(w http.ResponseWriter, r *http.Request) {
	issuerDID := chi.URLParam(r, "issuerDid")
	if issuerDID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("issuer_did is required"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListCredentialsByIssuer(r.Context(), query.ListCredentialsByIssuer{
		IssuerDID: issuerDID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toCredentialListResponse(result))
}

// ============================================================================
// Request Parsing Helpers
// ============================================================================

func (h *Handler) parsePagination(r *http.Request) (int, int) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

func toCredentialResponse(v *query.CredentialView) CredentialResponse {
	return CredentialResponse{
		CredentialID:     v.CredentialID,
		CredentialType:   v.CredentialType,
		IssuerDID:        v.IssuerDID,
		SubjectDID:       v.SubjectDID,
		Claims:           v.Claims,
		IssuedAt:         v.IssuedAt,
		ExpiresAt:        v.ExpiresAt,
		Status:           v.Status,
		RevokedAt:        v.RevokedAt,
		RevocationReason: v.RevocationReason,
		JWT:              v.JWT,
		VerificationID:   v.VerificationID,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
	}
}

func toCredentialListResponse(v *query.CredentialListView) CredentialListResponse {
	summaries := make([]CredentialSummaryResponse, len(v.Credentials))
	for i, c := range v.Credentials {
		summaries[i] = CredentialSummaryResponse{
			CredentialID:   c.CredentialID,
			CredentialType: c.CredentialType,
			IssuerDID:      c.IssuerDID,
			SubjectDID:     c.SubjectDID,
			Status:         c.Status,
			IssuedAt:       c.IssuedAt,
			ExpiresAt:      c.ExpiresAt,
		}
	}

	return CredentialListResponse{
		Credentials: summaries,
		TotalCount:  v.TotalCount,
		Limit:       v.Limit,
		Offset:      v.Offset,
		HasMore:     v.HasMore,
	}
}

// ============================================================================
// Response Writing Helpers
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
