package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/presentation/app/command"
	"github.com/0xsj/nexus/platform/internal/presentation/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Presentation context.
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
// Presentation Creation
// ============================================================================

// CreatePresentation handles POST /api/v1/presentations
func (h *Handler) CreatePresentation(w http.ResponseWriter, r *http.Request) {
	var req CreatePresentationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.HolderDID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("holder_did", "holder_did is required"))
		return
	}
	if len(req.CredentialIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("credential_ids", "credential_ids are required"))
		return
	}
	if req.DisclosurePolicyType == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("disclosure_policy_type", "disclosure_policy_type is required"))
		return
	}

	// Build command
	cmd := command.CreatePresentation{
		HolderDID:            req.HolderDID,
		CredentialIDs:        req.CredentialIDs,
		DisclosurePolicyType: req.DisclosurePolicyType,
		AllowedClaims:        req.AllowedClaims,
		BlockedClaims:        req.BlockedClaims,
		Purpose:              req.Purpose,
	}

	// Execute command
	result, err := h.commands.HandleCreatePresentation(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.CreatePresentationResult)
	h.writeJSON(w, http.StatusCreated, PresentationCreatedResponse{
		PresentationID: data.PresentationID,
		Status:         data.Status,
		CreatedAt:      time.Now().UTC(),
	})
}

// ============================================================================
// Presentation Retrieval
// ============================================================================

// GetPresentation handles GET /api/v1/presentations/{presentationId}
func (h *Handler) GetPresentation(w http.ResponseWriter, r *http.Request) {
	presentationID := chi.URLParam(r, "presentationId")
	if presentationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("presentation_id is required"))
		return
	}

	id, err := types.ParseID(presentationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("presentation_id", "invalid presentation_id format"))
		return
	}

	result, err := h.queries.HandleGetPresentation(r.Context(), query.GetPresentation{
		PresentationID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toPresentationResponse(result))
}

// ListPresentationsByHolder handles GET /api/v1/presentations/holder/{holderDid}
func (h *Handler) ListPresentationsByHolder(w http.ResponseWriter, r *http.Request) {
	holderDID := chi.URLParam(r, "holderDid")
	if holderDID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("holder_did is required"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListPresentationsByHolder(r.Context(), query.ListPresentationsByHolder{
		HolderDID: holderDID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toPresentationListResponse(result))
}

// ============================================================================
// Presentation Status Management
// ============================================================================

// RevokePresentation handles POST /api/v1/presentations/{presentationId}/revoke
func (h *Handler) RevokePresentation(w http.ResponseWriter, r *http.Request) {
	presentationID := chi.URLParam(r, "presentationId")
	if presentationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("presentation_id is required"))
		return
	}

	id, err := types.ParseID(presentationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("presentation_id", "invalid presentation_id format"))
		return
	}

	var req RevokePresentationRequest
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

	result, err := h.commands.HandleRevokePresentation(r.Context(), command.RevokePresentation{
		PresentationID: id,
		Reason:         req.Reason,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RevokePresentationResult)
	h.writeJSON(w, http.StatusOK, PresentationRevokedResponse{
		PresentationID: data.PresentationID,
		Status:         data.Status,
		RevokedAt:      time.Now().UTC(),
	})
}

// ============================================================================
// ShareLink Creation
// ============================================================================

// CreateShareLink handles POST /api/v1/presentations/{presentationId}/share-links
func (h *Handler) CreateShareLink(w http.ResponseWriter, r *http.Request) {
	presentationID := chi.URLParam(r, "presentationId")
	if presentationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("presentation_id is required"))
		return
	}

	presID, err := types.ParseID(presentationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("presentation_id", "invalid presentation_id format"))
		return
	}

	var req CreateShareLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	cmd := command.CreateShareLink{
		PresentationID: presID,
		ExpiresAt:      req.ExpiresAt,
		MaxViews:       req.MaxViews,
		Pin:            req.Pin,
		Audience:       req.Audience,
	}

	result, err := h.commands.HandleCreateShareLink(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.CreateShareLinkResult)
	h.writeJSON(w, http.StatusCreated, ShareLinkCreatedResponse{
		ShareLinkID:    data.ShareLinkID,
		PresentationID: data.PresentationID,
		Token:          data.Token,
		Status:         data.Status,
		CreatedAt:      time.Now().UTC(),
	})
}

// ============================================================================
// ShareLink Retrieval
// ============================================================================

// ListShareLinks handles GET /api/v1/presentations/{presentationId}/share-links
func (h *Handler) ListShareLinks(w http.ResponseWriter, r *http.Request) {
	presentationID := chi.URLParam(r, "presentationId")
	if presentationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("presentation_id is required"))
		return
	}

	presID, err := types.ParseID(presentationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("presentation_id", "invalid presentation_id format"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListShareLinks(r.Context(), query.ListShareLinks{
		PresentationID: presID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toShareLinkListResponse(result))
}

// GetShareLink handles GET /api/v1/share-links/{shareLinkId}
func (h *Handler) GetShareLink(w http.ResponseWriter, r *http.Request) {
	shareLinkID := chi.URLParam(r, "shareLinkId")
	if shareLinkID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("share_link_id is required"))
		return
	}

	id, err := types.ParseID(shareLinkID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("share_link_id", "invalid share_link_id format"))
		return
	}

	result, err := h.queries.HandleGetShareLink(r.Context(), query.GetShareLink{
		ShareLinkID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toShareLinkResponse(result))
}

// ============================================================================
// ShareLink Access & Revocation
// ============================================================================

// AccessShareLink handles POST /api/v1/share-links/{shareLinkId}/access
func (h *Handler) AccessShareLink(w http.ResponseWriter, r *http.Request) {
	shareLinkID := chi.URLParam(r, "shareLinkId")
	if shareLinkID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("share_link_id is required"))
		return
	}

	id, err := types.ParseID(shareLinkID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("share_link_id", "invalid share_link_id format"))
		return
	}

	var req AccessShareLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Okay if body is empty/invalid, verifier_did and ip_address are optional
		req = AccessShareLinkRequest{}
	}

	// Extract IP address from request
	ipAddress := r.RemoteAddr

	cmd := command.AccessShareLink{
		ShareLinkID: id,
		VerifierDID: req.VerifierDID,
		IPAddress:   ipAddress,
	}

	result, err := h.commands.HandleAccessShareLink(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.AccessShareLinkResult)
	h.writeJSON(w, http.StatusOK, ShareLinkAccessedResponse{
		ShareLinkID:  data.ShareLinkID,
		CurrentViews: data.CurrentViews,
		Status:       data.Status,
	})
}

// RevokeShareLink handles POST /api/v1/share-links/{shareLinkId}/revoke
func (h *Handler) RevokeShareLink(w http.ResponseWriter, r *http.Request) {
	shareLinkID := chi.URLParam(r, "shareLinkId")
	if shareLinkID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("share_link_id is required"))
		return
	}

	id, err := types.ParseID(shareLinkID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("share_link_id", "invalid share_link_id format"))
		return
	}

	result, err := h.commands.HandleRevokeShareLink(r.Context(), command.RevokeShareLink{
		ShareLinkID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RevokeShareLinkResult)
	h.writeJSON(w, http.StatusOK, ShareLinkRevokedResponse{
		ShareLinkID: data.ShareLinkID,
		Status:      data.Status,
		RevokedAt:   time.Now().UTC(),
	})
}

// ============================================================================
// Access Log
// ============================================================================

// GetAccessLog handles GET /api/v1/share-links/{shareLinkId}/access-log
func (h *Handler) GetAccessLog(w http.ResponseWriter, r *http.Request) {
	shareLinkID := chi.URLParam(r, "shareLinkId")
	if shareLinkID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("share_link_id is required"))
		return
	}

	id, err := types.ParseID(shareLinkID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("share_link_id", "invalid share_link_id format"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleGetAccessLog(r.Context(), query.GetAccessLog{
		ShareLinkID: id,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toAccessLogResponse(result))
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

func toPresentationResponse(v *query.PresentationView) PresentationResponse {
	return PresentationResponse{
		PresentationID:   v.PresentationID,
		HolderDID:        v.HolderDID,
		CredentialIDs:    v.CredentialIDs,
		DisclosurePolicy: v.DisclosurePolicy,
		VPJWT:            v.VPJWT,
		Purpose:          v.Purpose,
		Status:           v.Status,
		RevokedAt:        v.RevokedAt,
		RevocationReason: v.RevocationReason,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
	}
}

func toPresentationListResponse(v *query.PresentationListView) PresentationListResponse {
	summaries := make([]PresentationSummaryResponse, len(v.Presentations))
	for i, p := range v.Presentations {
		summaries[i] = PresentationSummaryResponse{
			PresentationID: p.PresentationID,
			HolderDID:      p.HolderDID,
			Purpose:        p.Purpose,
			Status:         p.Status,
			CreatedAt:      p.CreatedAt,
		}
	}

	return PresentationListResponse{
		Presentations: summaries,
		TotalCount:    v.TotalCount,
		Limit:         v.Limit,
		Offset:        v.Offset,
		HasMore:       v.HasMore,
	}
}

func toShareLinkResponse(v *query.ShareLinkView) ShareLinkResponse {
	return ShareLinkResponse{
		ShareLinkID:    v.ShareLinkID,
		PresentationID: v.PresentationID,
		Token:          v.Token,
		ExpiresAt:      v.ExpiresAt,
		MaxViews:       v.MaxViews,
		CurrentViews:   v.CurrentViews,
		PinProtected:   v.PinProtected,
		Audience:       v.Audience,
		Status:         v.Status,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func toShareLinkListResponse(v *query.ShareLinkListView) ShareLinkListResponse {
	links := make([]ShareLinkResponse, len(v.ShareLinks))
	for i, sl := range v.ShareLinks {
		links[i] = ShareLinkResponse{
			ShareLinkID:    sl.ShareLinkID,
			PresentationID: sl.PresentationID,
			Token:          sl.Token,
			ExpiresAt:      sl.ExpiresAt,
			MaxViews:       sl.MaxViews,
			CurrentViews:   sl.CurrentViews,
			PinProtected:   sl.PinProtected,
			Audience:       sl.Audience,
			Status:         sl.Status,
			CreatedAt:      sl.CreatedAt,
			UpdatedAt:      sl.UpdatedAt,
		}
	}

	return ShareLinkListResponse{
		ShareLinks: links,
		TotalCount: v.TotalCount,
		Limit:      v.Limit,
		Offset:     v.Offset,
		HasMore:    v.HasMore,
	}
}

func toAccessLogResponse(v *query.AccessLogView) AccessLogResponse {
	grants := make([]AccessGrantResponse, len(v.AccessGrants))
	for i, g := range v.AccessGrants {
		grants[i] = AccessGrantResponse{
			ID:              g.ID,
			ShareLinkID:     g.ShareLinkID,
			VerifierDID:     g.VerifierDID,
			AccessedAt:      g.AccessedAt,
			IPAddress:       g.IPAddress,
			DisclosedClaims: g.DisclosedClaims,
		}
	}

	return AccessLogResponse{
		AccessGrants: grants,
		TotalCount:   v.TotalCount,
		Limit:        v.Limit,
		Offset:       v.Offset,
		HasMore:      v.HasMore,
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
