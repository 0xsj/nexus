package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/verification/app/command"
	"github.com/0xsj/nexus/platform/internal/verification/app/query"
	"github.com/0xsj/nexus/platform/internal/verification/domain"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Verification context.
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
// Start Verification
// ============================================================================

// StartVerification handles POST /api/v1/verifications
func (h *Handler) StartVerification(w http.ResponseWriter, r *http.Request) {
	var req StartVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.ProviderType == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("provider_type", "provider_type is required"))
		return
	}

	// Parse provider type
	providerType, err := domain.ParseProviderType(req.ProviderType)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("provider_type", "invalid provider_type"))
		return
	}

	// TODO: Extract user_id from auth context in production
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("user_id", "user_id is required"))
		return
	}

	// Build command
	cmd := command.StartVerification{
		UserID:       userID,
		ProviderType: providerType,
		RedirectURI:  req.RedirectURI,
	}

	// Execute command
	result, err := h.commands.HandleStartVerification(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.StartVerificationResult)
	h.writeJSON(w, http.StatusCreated, VerificationStartedResponse{
		VerificationID: data.VerificationID,
		ProviderType:   data.ProviderType,
		Status:         data.Status,
		OAuthState:     data.OAuthState,
		AuthURL:        data.AuthURL,
		StartedAt:      time.Now().UTC(),
	})
}

// ============================================================================
// Get Verification
// ============================================================================

// GetVerification handles GET /api/v1/verifications/{verificationId}
func (h *Handler) GetVerification(w http.ResponseWriter, r *http.Request) {
	verificationID := chi.URLParam(r, "verificationId")
	if verificationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("verification_id is required"))
		return
	}

	result, err := h.queries.HandleGetVerification(r.Context(), query.GetVerification{
		VerificationID: verificationID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toVerificationResponse(result))
}

// ============================================================================
// List Verifications by User
// ============================================================================

// ListByUser handles GET /api/v1/verifications/user/{userId}
func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListVerificationsByUser(r.Context(), query.ListVerificationsByUser{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toVerificationListResponse(result))
}

// ============================================================================
// OAuth Callback
// ============================================================================

// OAuthCallback handles POST /api/v1/verifications/{verificationId}/callback
func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	verificationID := chi.URLParam(r, "verificationId")
	if verificationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("verification_id is required"))
		return
	}

	var req OAuthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.Code == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("code", "code is required"))
		return
	}
	if req.State == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("state", "state is required"))
		return
	}

	// Build command
	cmd := command.ReceiveOAuthCallback{
		VerificationID: verificationID,
		Code:           req.Code,
		State:          req.State,
		RedirectURI:    req.RedirectURI,
	}

	// Execute command
	result, err := h.commands.HandleReceiveOAuthCallback(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.ReceiveOAuthCallbackResult)
	h.writeJSON(w, http.StatusOK, OAuthCallbackResponse{
		VerificationID: data.VerificationID,
		Status:         data.Status,
		CredentialID:   data.CredentialID,
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

func toVerificationResponse(v *query.VerificationView) VerificationResponse {
	return VerificationResponse{
		VerificationID: v.VerificationID,
		UserID:         v.UserID,
		ProviderType:   v.ProviderType,
		Status:         v.Status,
		OAuthState:     v.OAuthState,
		ErrorMessage:   v.ErrorMessage,
		ErrorCode:      v.ErrorCode,
		CredentialID:   v.CredentialID,
		StartedAt:      v.StartedAt,
		CompletedAt:    v.CompletedAt,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func toVerificationListResponse(v *query.VerificationListView) VerificationListResponse {
	summaries := make([]VerificationSummaryResponse, len(v.Verifications))
	for i, ver := range v.Verifications {
		summaries[i] = VerificationSummaryResponse{
			VerificationID: ver.VerificationID,
			UserID:         ver.UserID,
			ProviderType:   ver.ProviderType,
			Status:         ver.Status,
			StartedAt:      ver.StartedAt,
			CompletedAt:    ver.CompletedAt,
		}
	}

	return VerificationListResponse{
		Verifications: summaries,
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
