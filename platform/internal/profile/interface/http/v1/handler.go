package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/profile/app/command"
	"github.com/0xsj/nexus/platform/internal/profile/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Profile context.
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
// Profile Creation
// ============================================================================

// CreateProfile handles POST /api/v1/profiles
func (h *Handler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.UserID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("user_id", "user_id is required"))
		return
	}
	if req.DisplayName == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("display_name", "display_name is required"))
		return
	}
	if req.VanitySlug == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("vanity_slug", "vanity_slug is required"))
		return
	}

	// Build command
	cmd := command.CreateProfile{
		UserID:      req.UserID,
		DisplayName: req.DisplayName,
		Headline:    req.Headline,
		VanitySlug:  req.VanitySlug,
	}

	// Execute command
	result, err := h.commands.HandleCreateProfile(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.CreateProfileResult)
	h.writeJSON(w, http.StatusCreated, ProfileCreatedResponse{
		ProfileID:  data.ProfileID,
		VanitySlug: data.VanitySlug,
		CreatedAt:  time.Now().UTC(),
	})
}

// ============================================================================
// Profile Retrieval
// ============================================================================

// GetProfile handles GET /api/v1/profiles/{profileId}
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	if profileID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("profile_id is required"))
		return
	}

	id, err := types.ParseID(profileID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("profile_id", "invalid profile_id format"))
		return
	}

	result, err := h.queries.HandleGetProfile(r.Context(), query.GetProfile{
		ProfileID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toProfileResponse(result))
}

// GetProfileByUser handles GET /api/v1/profiles/user/{userId}
func (h *Handler) GetProfileByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	result, err := h.queries.HandleGetProfileByUser(r.Context(), query.GetProfileByUser{
		UserID: userID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toProfileResponse(result))
}

// GetProfileByVanitySlug handles GET /api/v1/profiles/@{vanitySlug}
func (h *Handler) GetProfileByVanitySlug(w http.ResponseWriter, r *http.Request) {
	vanitySlug := chi.URLParam(r, "vanitySlug")
	if vanitySlug == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("vanity_slug is required"))
		return
	}

	result, err := h.queries.HandleGetProfileByVanitySlug(r.Context(), query.GetProfileByVanitySlug{
		VanitySlug: vanitySlug,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toProfileResponse(result))
}

// ListProfiles handles GET /api/v1/profiles
func (h *Handler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListProfiles(r.Context(), query.ListProfiles{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toProfileListResponse(result))
}

// ============================================================================
// Profile Updates
// ============================================================================

// UpdateProfile handles PATCH /api/v1/profiles/{profileId}
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	if profileID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("profile_id is required"))
		return
	}

	id, err := types.ParseID(profileID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("profile_id", "invalid profile_id format"))
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.DisplayName == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("display_name", "display_name is required"))
		return
	}

	result, err := h.commands.HandleUpdateProfile(r.Context(), command.UpdateProfile{
		ProfileID:   id,
		DisplayName: req.DisplayName,
		Headline:    req.Headline,
		Bio:         req.Bio,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.UpdateProfileResult)
	h.writeJSON(w, http.StatusOK, ProfileUpdatedResponse{
		ProfileID: data.ProfileID,
		UpdatedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Badge Management
// ============================================================================

// AddBadge handles POST /api/v1/profiles/{profileId}/badges
func (h *Handler) AddBadge(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	if profileID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("profile_id is required"))
		return
	}

	id, err := types.ParseID(profileID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("profile_id", "invalid profile_id format"))
		return
	}

	var req AddBadgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.CredentialID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("credential_id", "credential_id is required"))
		return
	}
	if req.BadgeType == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("badge_type", "badge_type is required"))
		return
	}
	if req.DisplayName == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("display_name", "display_name is required"))
		return
	}
	if req.PrimaryValue == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("primary_value", "primary_value is required"))
		return
	}

	result, err := h.commands.HandleAddBadge(r.Context(), command.AddBadge{
		ProfileID:      id,
		CredentialID:   req.CredentialID,
		BadgeType:      req.BadgeType,
		DisplayName:    req.DisplayName,
		PrimaryValue:   req.PrimaryValue,
		SecondaryValue: req.SecondaryValue,
		Icon:           req.Icon,
		VerifiedAt:     req.VerifiedAt,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.AddBadgeResult)
	h.writeJSON(w, http.StatusCreated, BadgeAddedResponse{
		ProfileID: data.ProfileID,
		BadgeID:   data.BadgeID,
		AddedAt:   time.Now().UTC(),
	})
}

// RemoveBadge handles DELETE /api/v1/profiles/{profileId}/badges/{badgeId}
func (h *Handler) RemoveBadge(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	if profileID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("profile_id is required"))
		return
	}

	profID, err := types.ParseID(profileID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("profile_id", "invalid profile_id format"))
		return
	}

	badgeID := chi.URLParam(r, "badgeId")
	if badgeID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("badge_id is required"))
		return
	}

	bID, err := types.ParseID(badgeID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("badge_id", "invalid badge_id format"))
		return
	}

	result, err := h.commands.HandleRemoveBadge(r.Context(), command.RemoveBadge{
		ProfileID: profID,
		BadgeID:   bID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RemoveBadgeResult)
	h.writeJSON(w, http.StatusOK, BadgeRemovedResponse{
		ProfileID: data.ProfileID,
		BadgeID:   data.BadgeID,
		RemovedAt: time.Now().UTC(),
	})
}

// ChangeBadgeVisibility handles PATCH /api/v1/profiles/{profileId}/badges/{badgeId}/visibility
func (h *Handler) ChangeBadgeVisibility(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	if profileID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("profile_id is required"))
		return
	}

	profID, err := types.ParseID(profileID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("profile_id", "invalid profile_id format"))
		return
	}

	badgeID := chi.URLParam(r, "badgeId")
	if badgeID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("badge_id is required"))
		return
	}

	bID, err := types.ParseID(badgeID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("badge_id", "invalid badge_id format"))
		return
	}

	var req ChangeBadgeVisibilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Visibility == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("visibility", "visibility is required"))
		return
	}

	result, err := h.commands.HandleChangeBadgeVisibility(r.Context(), command.ChangeBadgeVisibility{
		ProfileID:  profID,
		BadgeID:    bID,
		Visibility: req.Visibility,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ChangeBadgeVisibilityResult)
	h.writeJSON(w, http.StatusOK, BadgeVisibilityChangedResponse{
		ProfileID:  data.ProfileID,
		BadgeID:    data.BadgeID,
		Visibility: data.Visibility,
		ChangedAt:  time.Now().UTC(),
	})
}

// ============================================================================
// Vanity URL
// ============================================================================

// ClaimVanityURL handles POST /api/v1/profiles/{profileId}/vanity-url
func (h *Handler) ClaimVanityURL(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	if profileID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("profile_id is required"))
		return
	}

	id, err := types.ParseID(profileID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("profile_id", "invalid profile_id format"))
		return
	}

	var req ClaimVanityURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Slug == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("slug", "slug is required"))
		return
	}

	result, err := h.commands.HandleClaimVanityURL(r.Context(), command.ClaimVanityURL{
		ProfileID: id,
		Slug:      req.Slug,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ClaimVanityURLResult)
	h.writeJSON(w, http.StatusOK, VanityURLClaimedResponse{
		ProfileID: data.ProfileID,
		Slug:      data.Slug,
		ClaimedAt: time.Now().UTC(),
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

func toProfileResponse(v *query.ProfileView) ProfileResponse {
	badges := make([]BadgeResponse, len(v.Badges))
	for i, b := range v.Badges {
		badges[i] = BadgeResponse{
			BadgeID:        b.BadgeID,
			CredentialID:   b.CredentialID,
			BadgeType:      b.BadgeType,
			DisplayName:    b.DisplayName,
			PrimaryValue:   b.PrimaryValue,
			SecondaryValue: b.SecondaryValue,
			Icon:           b.Icon,
			VerifiedAt:     b.VerifiedAt,
			Visibility:     b.Visibility,
		}
	}

	return ProfileResponse{
		ProfileID:   v.ProfileID,
		UserID:      v.UserID,
		DisplayName: v.DisplayName,
		Headline:    v.Headline,
		Bio:         v.Bio,
		VanitySlug:  v.VanitySlug,
		Badges:      badges,
		BadgeCount:  v.BadgeCount,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func toProfileListResponse(v *query.ProfileListView) ProfileListResponse {
	summaries := make([]ProfileSummaryResponse, len(v.Profiles))
	for i, p := range v.Profiles {
		summaries[i] = ProfileSummaryResponse{
			ProfileID:   p.ProfileID,
			UserID:      p.UserID,
			DisplayName: p.DisplayName,
			Headline:    p.Headline,
			VanitySlug:  p.VanitySlug,
			BadgeCount:  p.BadgeCount,
			CreatedAt:   p.CreatedAt,
		}
	}

	return ProfileListResponse{
		Profiles:   summaries,
		TotalCount: v.TotalCount,
		Limit:      v.Limit,
		Offset:     v.Offset,
		HasMore:    v.HasMore,
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
