package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/organization/app/command"
	"github.com/0xsj/nexus/platform/internal/organization/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Organization context.
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
// Organization Creation
// ============================================================================

// CreateOrganization handles POST /api/v1/organizations
func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("name", "name is required"))
		return
	}
	if req.Slug == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("slug", "slug is required"))
		return
	}
	if req.OrgType == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("org_type", "org_type is required"))
		return
	}
	if req.OwnerUserID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("owner_user_id", "owner_user_id is required"))
		return
	}

	// Build command
	cmd := command.CreateOrganization{
		Name:        req.Name,
		Slug:        req.Slug,
		OrgType:     req.OrgType,
		OwnerUserID: req.OwnerUserID,
	}

	// Execute command
	result, err := h.commands.HandleCreateOrganization(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.CreateOrganizationResult)
	h.writeJSON(w, http.StatusCreated, OrganizationCreatedResponse{
		OrganizationID: data.OrganizationID,
		Slug:           data.Slug,
		OwnerMemberID:  data.OwnerMemberID,
		CreatedAt:      time.Now().UTC(),
	})
}

// ============================================================================
// Organization Retrieval
// ============================================================================

// GetOrganization handles GET /api/v1/organizations/{orgId}
func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	id, err := types.ParseID(orgID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "invalid organization_id format"))
		return
	}

	result, err := h.queries.HandleGetOrganization(r.Context(), query.GetOrganization{
		OrganizationID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toOrganizationResponse(result))
}

// GetOrganizationBySlug handles GET /api/v1/organizations/slug/{slug}
func (h *Handler) GetOrganizationBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("slug is required"))
		return
	}

	result, err := h.queries.HandleGetOrganizationBySlug(r.Context(), query.GetOrganizationBySlug{
		Slug: slug,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toOrganizationResponse(result))
}

// ListOrganizations handles GET /api/v1/organizations
func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListOrganizations(r.Context(), query.ListOrganizations{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toOrganizationListResponse(result))
}

// ============================================================================
// Organization Updates
// ============================================================================

// UpdateOrganization handles PATCH /api/v1/organizations/{orgId}
func (h *Handler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	id, err := types.ParseID(orgID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "invalid organization_id format"))
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Name == "" && req.Description == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("at least one of name or description is required"))
		return
	}

	result, err := h.commands.HandleUpdateOrganization(r.Context(), command.UpdateOrganization{
		OrganizationID: id,
		Name:           req.Name,
		Description:    req.Description,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.UpdateOrganizationResult)
	h.writeJSON(w, http.StatusOK, OrganizationUpdatedResponse{
		OrganizationID: data.OrganizationID,
		UpdatedAt:      time.Now().UTC(),
	})
}

// DeleteOrganization handles DELETE /api/v1/organizations/{orgId}
func (h *Handler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	id, err := types.ParseID(orgID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "invalid organization_id format"))
		return
	}

	result, err := h.commands.HandleDeleteOrganization(r.Context(), command.DeleteOrganization{
		OrganizationID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.DeleteOrganizationResult)
	h.writeJSON(w, http.StatusOK, OrganizationDeletedResponse{
		OrganizationID: data.OrganizationID,
		DeletedAt:      time.Now().UTC(),
	})
}

// ============================================================================
// Member Management
// ============================================================================

// AddMember handles POST /api/v1/organizations/{orgId}/members
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	id, err := types.ParseID(orgID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "invalid organization_id format"))
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.UserID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("user_id", "user_id is required"))
		return
	}
	if req.Role == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("role", "role is required"))
		return
	}

	result, err := h.commands.HandleAddMember(r.Context(), command.AddMember{
		OrganizationID: id,
		UserID:         req.UserID,
		Role:           req.Role,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.AddMemberResult)
	h.writeJSON(w, http.StatusCreated, MemberAddedResponse{
		OrganizationID: data.OrganizationID,
		MemberID:       data.MemberID,
		Role:           data.Role,
		JoinedAt:       time.Now().UTC(),
	})
}

// RemoveMember handles DELETE /api/v1/organizations/{orgId}/members/{memberId}
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	memberIDStr := chi.URLParam(r, "memberId")
	if memberIDStr == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("member_id is required"))
		return
	}

	orgTypesID, err := types.ParseID(orgID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "invalid organization_id format"))
		return
	}

	memberTypesID, err := types.ParseID(memberIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("member_id", "invalid member_id format"))
		return
	}

	// Parse optional reason from body
	var req RemoveMemberRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
			return
		}
	}

	result, err := h.commands.HandleRemoveMember(r.Context(), command.RemoveMember{
		OrganizationID: orgTypesID,
		MemberID:       memberTypesID,
		Reason:         req.Reason,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RemoveMemberResult)
	h.writeJSON(w, http.StatusOK, MemberRemovedResponse{
		OrganizationID: data.OrganizationID,
		MemberID:       data.MemberID,
		RemovedAt:      time.Now().UTC(),
	})
}

// ChangeMemberRole handles PATCH /api/v1/organizations/{orgId}/members/{memberId}/role
func (h *Handler) ChangeMemberRole(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	memberIDStr := chi.URLParam(r, "memberId")
	if memberIDStr == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("member_id is required"))
		return
	}

	orgTypesID, err := types.ParseID(orgID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "invalid organization_id format"))
		return
	}

	memberTypesID, err := types.ParseID(memberIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("member_id", "invalid member_id format"))
		return
	}

	var req ChangeMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.NewRole == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("new_role", "new_role is required"))
		return
	}

	result, err := h.commands.HandleChangeMemberRole(r.Context(), command.ChangeMemberRole{
		OrganizationID: orgTypesID,
		MemberID:       memberTypesID,
		NewRole:        req.NewRole,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ChangeMemberRoleResult)
	h.writeJSON(w, http.StatusOK, MemberRoleChangedResponse{
		OrganizationID: data.OrganizationID,
		MemberID:       data.MemberID,
		NewRole:        data.NewRole,
		ChangedAt:      time.Now().UTC(),
	})
}

// ============================================================================
// Ownership Transfer
// ============================================================================

// TransferOwnership handles POST /api/v1/organizations/{orgId}/transfer-ownership
func (h *Handler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	orgTypesID, err := types.ParseID(orgID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "invalid organization_id format"))
		return
	}

	var req TransferOwnershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.ToMemberID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("to_member_id", "to_member_id is required"))
		return
	}

	toMemberTypesID, err := types.ParseID(req.ToMemberID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("to_member_id", "invalid to_member_id format"))
		return
	}

	result, err := h.commands.HandleTransferOwnership(r.Context(), command.TransferOwnership{
		OrganizationID: orgTypesID,
		ToMemberID:     toMemberTypesID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.TransferOwnershipResult)
	h.writeJSON(w, http.StatusOK, OwnershipTransferredResponse{
		OrganizationID: data.OrganizationID,
		NewOwnerID:     data.NewOwnerID,
		TransferredAt:  time.Now().UTC(),
	})
}

// ============================================================================
// Verification
// ============================================================================

// CompleteVerification handles POST /api/v1/organizations/{orgId}/verify
func (h *Handler) CompleteVerification(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	orgTypesID, err := types.ParseID(orgID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "invalid organization_id format"))
		return
	}

	var req CompleteVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.DID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("did", "did is required"))
		return
	}

	result, err := h.commands.HandleCompleteVerification(r.Context(), command.CompleteVerification{
		OrganizationID: orgTypesID,
		DID:            req.DID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.CompleteVerificationResult)
	h.writeJSON(w, http.StatusOK, VerificationCompletedResponse{
		OrganizationID:     data.OrganizationID,
		DID:                data.DID,
		VerificationStatus: data.VerificationStatus,
		VerifiedAt:         time.Now().UTC(),
	})
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

func toOrganizationResponse(v *query.OrganizationView) OrganizationResponse {
	return OrganizationResponse{
		OrganizationID:     v.OrganizationID,
		Name:               v.Name,
		Slug:               v.Slug,
		OrgType:            v.OrgType,
		Description:        v.Description,
		VerificationStatus: v.VerificationStatus,
		DID:                v.DID,
		OwnerMemberID:      v.OwnerMemberID,
		MemberCount:        v.MemberCount,
		CreatedAt:          v.CreatedAt,
		UpdatedAt:          v.UpdatedAt,
	}
}

func toOrganizationListResponse(v *query.OrganizationListView) OrganizationListResponse {
	summaries := make([]OrganizationSummaryResponse, len(v.Organizations))
	for i, o := range v.Organizations {
		summaries[i] = OrganizationSummaryResponse{
			OrganizationID:     o.OrganizationID,
			Name:               o.Name,
			Slug:               o.Slug,
			OrgType:            o.OrgType,
			VerificationStatus: o.VerificationStatus,
			MemberCount:        o.MemberCount,
			CreatedAt:          o.CreatedAt,
		}
	}

	return OrganizationListResponse{
		Organizations: summaries,
		TotalCount:    v.TotalCount,
		Limit:         v.Limit,
		Offset:        v.Offset,
		HasMore:       v.HasMore,
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
