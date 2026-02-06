package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/trust/app/command"
	"github.com/0xsj/nexus/platform/internal/trust/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Trust context.
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
// Vouch Creation
// ============================================================================

// GiveVouch handles POST /api/v1/trust/vouches
func (h *Handler) GiveVouch(w http.ResponseWriter, r *http.Request) {
	var req GiveVouchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.VoucherID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("voucher_id", "voucher_id is required"))
		return
	}
	if req.VoucheeID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("vouchee_id", "vouchee_id is required"))
		return
	}
	if req.Relationship == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("relationship", "relationship is required"))
		return
	}
	if req.Strength < 1 || req.Strength > 10 {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("strength", "strength must be between 1 and 10"))
		return
	}

	// Build command
	cmd := command.GiveVouch{
		VoucherID:    req.VoucherID,
		VoucheeID:    req.VoucheeID,
		CredentialID: req.CredentialID,
		ClaimKey:     req.ClaimKey,
		Relationship: req.Relationship,
		Strength:     req.Strength,
		Statement:    req.Statement,
		Context:      req.Context,
		ExpiresAt:    req.ExpiresAt,
	}

	// Execute command
	result, err := h.commands.HandleGiveVouch(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.GiveVouchResult)
	h.writeJSON(w, http.StatusCreated, VouchGivenResponse{
		VouchID:   data.VouchID,
		Status:    data.Status,
		CreatedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Vouch Retrieval
// ============================================================================

// GetVouch handles GET /api/v1/trust/vouches/{vouchId}
func (h *Handler) GetVouch(w http.ResponseWriter, r *http.Request) {
	vouchID := chi.URLParam(r, "vouchId")
	if vouchID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("vouch_id is required"))
		return
	}

	id, err := types.ParseID(vouchID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("vouch_id", "invalid vouch_id format"))
		return
	}

	result, err := h.queries.HandleGetVouch(r.Context(), query.GetVouch{
		VouchID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toVouchResponse(result))
}

// ============================================================================
// Vouch Status Management
// ============================================================================

// AcceptVouch handles POST /api/v1/trust/vouches/{vouchId}/accept
func (h *Handler) AcceptVouch(w http.ResponseWriter, r *http.Request) {
	vouchID := chi.URLParam(r, "vouchId")
	if vouchID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("vouch_id is required"))
		return
	}

	id, err := types.ParseID(vouchID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("vouch_id", "invalid vouch_id format"))
		return
	}

	var req AcceptVouchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.VoucheeID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("vouchee_id", "vouchee_id is required"))
		return
	}

	result, err := h.commands.HandleAcceptVouch(r.Context(), command.AcceptVouch{
		VouchID:   id,
		VoucheeID: req.VoucheeID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.AcceptVouchResult)
	h.writeJSON(w, http.StatusOK, VouchAcceptedResponse{
		VouchID:    data.VouchID,
		Status:     data.Status,
		AcceptedAt: time.Now().UTC(),
	})
}

// RevokeVouch handles POST /api/v1/trust/vouches/{vouchId}/revoke
func (h *Handler) RevokeVouch(w http.ResponseWriter, r *http.Request) {
	vouchID := chi.URLParam(r, "vouchId")
	if vouchID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("vouch_id is required"))
		return
	}

	id, err := types.ParseID(vouchID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("vouch_id", "invalid vouch_id format"))
		return
	}

	var req RevokeVouchRequest
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

	result, err := h.commands.HandleRevokeVouch(r.Context(), command.RevokeVouch{
		VouchID: id,
		Reason:  req.Reason,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RevokeVouchResult)
	h.writeJSON(w, http.StatusOK, VouchRevokedResponse{
		VouchID:   data.VouchID,
		Status:    data.Status,
		RevokedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Vouch Lists
// ============================================================================

// ListVouchesByVouchee handles GET /api/v1/trust/vouches/received/{userId}
func (h *Handler) ListVouchesByVouchee(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListVouchesByVouchee(r.Context(), query.ListVouchesByVouchee{
		VoucheeID: userID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toVouchListResponse(result))
}

// ListVouchesByVoucher handles GET /api/v1/trust/vouches/given/{userId}
func (h *Handler) ListVouchesByVoucher(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListVouchesByVoucher(r.Context(), query.ListVouchesByVoucher{
		VoucherID: userID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toVouchListResponse(result))
}

// ============================================================================
// Reputation
// ============================================================================

// GetReputation handles GET /api/v1/trust/reputation/{userId}
func (h *Handler) GetReputation(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	result, err := h.queries.HandleGetReputation(r.Context(), query.GetReputation{
		UserID: userID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toReputationResponse(result))
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

func toVouchResponse(v *query.VouchView) VouchResponse {
	return VouchResponse{
		VouchID:          v.VouchID,
		VoucherID:        v.VoucherID,
		VoucheeID:        v.VoucheeID,
		CredentialID:     v.CredentialID,
		ClaimKey:         v.ClaimKey,
		Relationship:     v.Relationship,
		Strength:         v.Strength,
		Statement:        v.Statement,
		Context:          v.Context,
		Status:           v.Status,
		ExpiresAt:        v.ExpiresAt,
		AcceptedAt:       v.AcceptedAt,
		RevokedAt:        v.RevokedAt,
		RevocationReason: v.RevocationReason,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
	}
}

func toVouchListResponse(v *query.VouchListView) VouchListResponse {
	summaries := make([]VouchSummaryResponse, len(v.Vouches))
	for i, vouch := range v.Vouches {
		summaries[i] = VouchSummaryResponse{
			VouchID:      vouch.VouchID,
			VoucherID:    vouch.VoucherID,
			VoucheeID:    vouch.VoucheeID,
			Relationship: vouch.Relationship,
			Strength:     vouch.Strength,
			Status:       vouch.Status,
			CreatedAt:    vouch.CreatedAt,
			ExpiresAt:    vouch.ExpiresAt,
		}
	}

	return VouchListResponse{
		Vouches:    summaries,
		TotalCount: v.TotalCount,
		Limit:      v.Limit,
		Offset:     v.Offset,
		HasMore:    v.HasMore,
	}
}

func toReputationResponse(v *query.ReputationView) ReputationResponse {
	return ReputationResponse{
		UserID:           v.UserID,
		OverallScore:     v.OverallScore,
		CredentialScore:  v.CredentialScore,
		VouchScore:       v.VouchScore,
		NetworkScore:     v.NetworkScore,
		VouchCount:       v.VouchCount,
		LastCalculatedAt: v.LastCalculatedAt,
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
