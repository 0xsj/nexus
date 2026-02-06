package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/issuer/app/command"
	"github.com/0xsj/nexus/platform/internal/issuer/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Issuer context.
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
// Issuer Registration
// ============================================================================

// RegisterIssuer handles POST /api/v1/issuers
func (h *Handler) RegisterIssuer(w http.ResponseWriter, r *http.Request) {
	var req RegisterIssuerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.OrganizationID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("organization_id", "organization_id is required"))
		return
	}
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("name", "name is required"))
		return
	}

	cmd := command.RegisterIssuer{
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		WebhookURL:     req.WebhookURL,
	}

	result, err := h.commands.HandleRegisterIssuer(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RegisterIssuerResult)
	h.writeJSON(w, http.StatusCreated, IssuerRegisteredResponse{
		IssuerID:  data.IssuerID,
		Status:    data.Status,
		CreatedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Issuer Retrieval
// ============================================================================

// GetIssuer handles GET /api/v1/issuers/{issuerId}
func (h *Handler) GetIssuer(w http.ResponseWriter, r *http.Request) {
	issuerID := chi.URLParam(r, "issuerId")
	if issuerID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("issuer_id is required"))
		return
	}

	id, err := types.ParseID(issuerID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("issuer_id", "invalid issuer_id format"))
		return
	}

	result, err := h.queries.HandleGetIssuer(r.Context(), query.GetIssuer{
		IssuerID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toIssuerResponse(result))
}

// GetIssuerByOrganization handles GET /api/v1/issuers/organization/{orgId}
func (h *Handler) GetIssuerByOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("organization_id is required"))
		return
	}

	result, err := h.queries.HandleGetIssuerByOrganization(r.Context(), query.GetIssuerByOrganization{
		OrganizationID: orgID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toIssuerResponse(result))
}

// ListIssuers handles GET /api/v1/issuers
func (h *Handler) ListIssuers(w http.ResponseWriter, r *http.Request) {
	limit, offset := h.parsePagination(r)

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	result, err := h.queries.HandleListIssuers(r.Context(), query.ListIssuers{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		statusCode, errResp := MapError(err)
		h.writeError(w, statusCode, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toIssuerListResponse(result))
}

// ============================================================================
// Issuer Status Management
// ============================================================================

// ActivateIssuer handles POST /api/v1/issuers/{issuerId}/activate
func (h *Handler) ActivateIssuer(w http.ResponseWriter, r *http.Request) {
	issuerID := chi.URLParam(r, "issuerId")
	if issuerID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("issuer_id is required"))
		return
	}

	id, err := types.ParseID(issuerID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("issuer_id", "invalid issuer_id format"))
		return
	}

	var req ActivateIssuerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.DID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("did", "did is required"))
		return
	}
	if req.APIKey == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("api_key", "api_key is required"))
		return
	}

	result, err := h.commands.HandleActivateIssuer(r.Context(), command.ActivateIssuer{
		IssuerID:          id,
		DID:               req.DID,
		APIKey:            req.APIKey,
		LogoURL:           req.LogoURL,
		PrimaryColor:      req.PrimaryColor,
		SecondaryColor:    req.SecondaryColor,
		CertificateDesign: req.CertificateDesign,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ActivateIssuerResult)
	h.writeJSON(w, http.StatusOK, IssuerActivatedResponse{
		IssuerID:    data.IssuerID,
		Status:      data.Status,
		ActivatedAt: time.Now().UTC(),
	})
}

// SuspendIssuer handles POST /api/v1/issuers/{issuerId}/suspend
func (h *Handler) SuspendIssuer(w http.ResponseWriter, r *http.Request) {
	issuerID := chi.URLParam(r, "issuerId")
	if issuerID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("issuer_id is required"))
		return
	}

	id, err := types.ParseID(issuerID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("issuer_id", "invalid issuer_id format"))
		return
	}

	var req SuspendIssuerRequest
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

	result, err := h.commands.HandleSuspendIssuer(r.Context(), command.SuspendIssuer{
		IssuerID: id,
		Reason:   req.Reason,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.SuspendIssuerResult)
	h.writeJSON(w, http.StatusOK, IssuerSuspendedResponse{
		IssuerID:    data.IssuerID,
		Status:      data.Status,
		SuspendedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Template Management
// ============================================================================

// CreateTemplate handles POST /api/v1/issuers/{issuerId}/templates
func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	issuerID := chi.URLParam(r, "issuerId")
	if issuerID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("issuer_id is required"))
		return
	}

	id, err := types.ParseID(issuerID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("issuer_id", "invalid issuer_id format"))
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("name", "name is required"))
		return
	}
	if req.SchemaType == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_type", "schema_type is required"))
		return
	}

	result, err := h.commands.HandleCreateTemplate(r.Context(), command.CreateTemplate{
		IssuerID:       id,
		Name:           req.Name,
		Description:    req.Description,
		SchemaType:     req.SchemaType,
		ClaimMappings:  req.ClaimMappings,
		DefaultValues:  req.DefaultValues,
		ExpirationDays: req.ExpirationDays,
		AutoApprove:    req.AutoApprove,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.CreateTemplateResult)
	h.writeJSON(w, http.StatusCreated, TemplateCreatedResponse{
		TemplateID: data.TemplateID,
		IssuerID:   data.IssuerID,
		Status:     data.Status,
		CreatedAt:  time.Now().UTC(),
	})
}

// ListTemplatesByIssuer handles GET /api/v1/issuers/{issuerId}/templates
func (h *Handler) ListTemplatesByIssuer(w http.ResponseWriter, r *http.Request) {
	issuerID := chi.URLParam(r, "issuerId")
	if issuerID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("issuer_id is required"))
		return
	}

	id, err := types.ParseID(issuerID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("issuer_id", "invalid issuer_id format"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListTemplatesByIssuer(r.Context(), query.ListTemplatesByIssuer{
		IssuerID: id,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toTemplateListResponse(result))
}

// GetTemplate handles GET /api/v1/issuers/{issuerId}/templates/{templateId}
func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	if templateID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("template_id is required"))
		return
	}

	id, err := types.ParseID(templateID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("template_id", "invalid template_id format"))
		return
	}

	result, err := h.queries.HandleGetTemplate(r.Context(), query.GetTemplate{
		TemplateID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toTemplateResponse(result))
}

// UpdateTemplate handles PUT /api/v1/issuers/{issuerId}/templates/{templateId}
func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	if templateID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("template_id is required"))
		return
	}

	id, err := types.ParseID(templateID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("template_id", "invalid template_id format"))
		return
	}

	var req UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("name", "name is required"))
		return
	}

	result, err := h.commands.HandleUpdateTemplate(r.Context(), command.UpdateTemplate{
		TemplateID:     id,
		Name:           req.Name,
		Description:    req.Description,
		ClaimMappings:  req.ClaimMappings,
		DefaultValues:  req.DefaultValues,
		ExpirationDays: req.ExpirationDays,
		AutoApprove:    req.AutoApprove,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.UpdateTemplateResult)
	h.writeJSON(w, http.StatusOK, TemplateUpdatedResponse{
		TemplateID: data.TemplateID,
		Version:    data.Version,
		Status:     data.Status,
		UpdatedAt:  time.Now().UTC(),
	})
}

// ArchiveTemplate handles POST /api/v1/issuers/{issuerId}/templates/{templateId}/archive
func (h *Handler) ArchiveTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	if templateID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("template_id is required"))
		return
	}

	id, err := types.ParseID(templateID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("template_id", "invalid template_id format"))
		return
	}

	result, err := h.commands.HandleArchiveTemplate(r.Context(), command.ArchiveTemplate{
		TemplateID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ArchiveTemplateResult)
	h.writeJSON(w, http.StatusOK, TemplateArchivedResponse{
		TemplateID: data.TemplateID,
		Status:     data.Status,
		ArchivedAt: time.Now().UTC(),
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

func toIssuerResponse(v *query.IssuerView) IssuerResponse {
	return IssuerResponse{
		IssuerID:       v.IssuerID,
		OrganizationID: v.OrganizationID,
		Name:           v.Name,
		Description:    v.Description,
		DID:            v.DID,
		WebhookURL:     v.WebhookURL,
		Status:         v.Status,
		Branding:       v.Branding,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func toIssuerListResponse(v *query.IssuerListView) IssuerListResponse {
	summaries := make([]IssuerSummaryResponse, len(v.Issuers))
	for i, iss := range v.Issuers {
		summaries[i] = IssuerSummaryResponse{
			IssuerID:       iss.IssuerID,
			OrganizationID: iss.OrganizationID,
			Name:           iss.Name,
			Status:         iss.Status,
			CreatedAt:      iss.CreatedAt,
		}
	}

	return IssuerListResponse{
		Issuers:    summaries,
		TotalCount: v.TotalCount,
		Limit:      v.Limit,
		Offset:     v.Offset,
		HasMore:    v.HasMore,
	}
}

func toTemplateResponse(v *query.TemplateView) TemplateResponse {
	return TemplateResponse{
		TemplateID:     v.TemplateID,
		IssuerID:       v.IssuerID,
		Name:           v.Name,
		Description:    v.Description,
		SchemaType:     v.SchemaType,
		ClaimMappings:  v.ClaimMappings,
		DefaultValues:  v.DefaultValues,
		ExpirationDays: v.ExpirationDays,
		AutoApprove:    v.AutoApprove,
		Status:         v.Status,
		Version:        v.Version,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func toTemplateListResponse(v *query.TemplateListView) TemplateListResponse {
	summaries := make([]TemplateSummaryResponse, len(v.Templates))
	for i, t := range v.Templates {
		summaries[i] = TemplateSummaryResponse{
			TemplateID: t.TemplateID,
			IssuerID:   t.IssuerID,
			Name:       t.Name,
			SchemaType: t.SchemaType,
			Status:     t.Status,
			Version:    t.Version,
			CreatedAt:  t.CreatedAt,
		}
	}

	return TemplateListResponse{
		Templates:  summaries,
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
