package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/schema/app/command"
	"github.com/0xsj/nexus/platform/internal/schema/app/query"
	"github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Schema context.
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
// Schema CRUD
// ============================================================================

// RegisterSchema handles POST /schemas
func (h *Handler) RegisterSchema(w http.ResponseWriter, r *http.Request) {
	var req RegisterSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.SchemaType == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_type", "schema_type is required"))
		return
	}
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("name", "name is required"))
		return
	}
	if req.Version == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("version", "version is required"))
		return
	}
	if len(req.Claims) == 0 {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("claims", "at least one claim is required"))
		return
	}

	// Parse version
	version, err := domain.ParseSchemaVersion(req.Version)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("version", "invalid version format"))
		return
	}

	// Convert claims
	claims, err := h.toCommandClaims(req.Claims)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse(err.Error()))
		return
	}

	// Parse issuer ID if provided
	var issuerID *types.ID
	if req.IssuerID != nil && *req.IssuerID != "" {
		id, err := types.ParseID(*req.IssuerID)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("issuer_id", "invalid issuer_id format"))
			return
		}
		issuerID = &id
	}

	// Generate schema ID
	schemaID := types.NewID()

	// Execute command
	result, err := h.commands.HandleRegisterSchema(r.Context(), command.RegisterSchema{
		SchemaID:    schemaID,
		SchemaType:  req.SchemaType,
		Name:        req.Name,
		Description: req.Description,
		Version:     version,
		Claims:      claims,
		IssuerID:    issuerID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.RegisterSchemaResult)
	h.writeJSON(w, http.StatusCreated, RegisterSchemaResponse{
		SchemaID:   data.SchemaID,
		SchemaType: data.SchemaType,
		Version:    data.Version,
		CreatedAt:  time.Now().UTC(),
	})
}

// GetSchema handles GET /schemas/{schemaId}
func (h *Handler) GetSchema(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	if schemaID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_id is required"))
		return
	}

	id, err := types.ParseID(schemaID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_id", "invalid schema_id format"))
		return
	}

	result, err := h.queries.HandleGetSchema(r.Context(), query.GetSchema{
		SchemaID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toSchemaResponse(result))
}

// GetSchemaByType handles GET /schemas/type/{schemaType}
func (h *Handler) GetSchemaByType(w http.ResponseWriter, r *http.Request) {
	schemaType := chi.URLParam(r, "schemaType")
	if schemaType == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_type is required"))
		return
	}

	result, err := h.queries.HandleGetSchemaByType(r.Context(), query.GetSchemaByType{
		SchemaType: schemaType,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toSchemaResponse(result))
}

// ListSchemas handles GET /schemas
func (h *Handler) ListSchemas(w http.ResponseWriter, r *http.Request) {
	req := h.parseListSchemasRequest(r)

	// Build query
	q := query.ListSchemas{
		PageSize: req.PageSize,
	}

	if req.Status != "" {
		q.Status = &req.Status
	}
	if req.IssuerID != "" {
		id, err := types.ParseID(req.IssuerID)
		if err == nil {
			q.IssuerID = &id
		}
	}
	if req.BuiltInOnly != nil {
		q.BuiltInOnly = req.BuiltInOnly
	}

	result, err := h.queries.HandleListSchemas(r.Context(), q)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toSchemaListResponse(result, req.Page, req.PageSize))
}

// ============================================================================
// Version Management
// ============================================================================

// AddSchemaVersion handles POST /schemas/{schemaId}/versions
func (h *Handler) AddSchemaVersion(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	if schemaID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_id is required"))
		return
	}

	id, err := types.ParseID(schemaID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_id", "invalid schema_id format"))
		return
	}

	var req AddSchemaVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate
	if req.Version == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("version", "version is required"))
		return
	}
	if len(req.Claims) == 0 {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("claims", "at least one claim is required"))
		return
	}

	// Parse version
	version, err := domain.ParseSchemaVersion(req.Version)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("version", "invalid version format"))
		return
	}

	// Convert claims
	claims, err := h.toCommandClaims(req.Claims)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse(err.Error()))
		return
	}

	result, err := h.commands.HandleAddSchemaVersion(r.Context(), command.AddSchemaVersion{
		SchemaID:      id,
		Version:       version,
		Claims:        claims,
		ChangeSummary: req.ChangeSummary,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.AddSchemaVersionResult)
	h.writeJSON(w, http.StatusCreated, AddVersionResponse{
		SchemaID:        data.SchemaID,
		Version:         data.Version,
		PreviousVersion: data.PreviousVersion,
	})
}

// GetSchemaVersion handles GET /schemas/{schemaId}/versions/{version}
func (h *Handler) GetSchemaVersion(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	version := chi.URLParam(r, "version")

	if schemaID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_id is required"))
		return
	}
	if version == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("version is required"))
		return
	}

	id, err := types.ParseID(schemaID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_id", "invalid schema_id format"))
		return
	}

	result, err := h.queries.HandleGetSchemaVersion(r.Context(), query.GetSchemaVersion{
		SchemaID: id,
		Version:  version,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toSchemaVersionResponse(result))
}

// ListSchemaVersions handles GET /schemas/{schemaId}/versions
func (h *Handler) ListSchemaVersions(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	if schemaID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_id is required"))
		return
	}

	id, err := types.ParseID(schemaID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_id", "invalid schema_id format"))
		return
	}

	result, err := h.queries.HandleListSchemaVersions(r.Context(), query.ListSchemaVersions{
		SchemaID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toVersionListResponse(result))
}

// ============================================================================
// Status Management
// ============================================================================

// DeprecateSchema handles POST /schemas/{schemaId}/deprecate
func (h *Handler) DeprecateSchema(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	if schemaID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_id is required"))
		return
	}

	id, err := types.ParseID(schemaID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_id", "invalid schema_id format"))
		return
	}

	var req DeprecateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is OK
		req = DeprecateSchemaRequest{}
	}

	var replacementID *types.ID
	if req.ReplacementSchemaID != nil && *req.ReplacementSchemaID != "" {
		rid, err := types.ParseID(*req.ReplacementSchemaID)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("replacement_schema_id", "invalid replacement_schema_id format"))
			return
		}
		replacementID = &rid
	}

	result, err := h.commands.HandleDeprecateSchema(r.Context(), command.DeprecateSchema{
		SchemaID:            id,
		Reason:              req.Reason,
		ReplacementSchemaID: replacementID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.DeprecateSchemaResult)
	h.writeJSON(w, http.StatusOK, DeprecateSchemaResponse{
		SchemaID:            data.SchemaID,
		ReplacementSchemaID: data.ReplacementSchemaID,
		DeprecatedAt:        time.Now().UTC(),
	})
}

// ActivateSchema handles POST /schemas/{schemaId}/activate
func (h *Handler) ActivateSchema(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	if schemaID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_id is required"))
		return
	}

	id, err := types.ParseID(schemaID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_id", "invalid schema_id format"))
		return
	}

	result, err := h.commands.HandleActivateSchema(r.Context(), command.ActivateSchema{
		SchemaID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.ActivateSchemaResult)
	h.writeJSON(w, http.StatusOK, ActivateSchemaResponse{
		SchemaID:    data.SchemaID,
		ActivatedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Metadata Management
// ============================================================================

// UpdateSchemaMetadata handles PATCH /schemas/{schemaId}
func (h *Handler) UpdateSchemaMetadata(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	if schemaID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_id is required"))
		return
	}

	id, err := types.ParseID(schemaID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_id", "invalid schema_id format"))
		return
	}

	var req UpdateSchemaMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Name == "" && req.Description == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("at least one field (name or description) is required"))
		return
	}

	result, err := h.commands.HandleUpdateSchemaMetadata(r.Context(), command.UpdateSchemaMetadata{
		SchemaID:    id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.UpdateSchemaMetadataResult)
	h.writeJSON(w, http.StatusOK, UpdateMetadataResponse{
		SchemaID:    data.SchemaID,
		Name:        data.Name,
		Description: data.Description,
		UpdatedAt:   time.Now().UTC(),
	})
}

// ============================================================================
// Lookup Endpoints
// ============================================================================

// CheckSchemaExists handles GET /schemas/exists
func (h *Handler) CheckSchemaExists(w http.ResponseWriter, r *http.Request) {
	schemaID := r.URL.Query().Get("schema_id")
	schemaType := r.URL.Query().Get("schema_type")

	if schemaID == "" && schemaType == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("either schema_id or schema_type is required"))
		return
	}

	q := query.SchemaExists{}

	if schemaID != "" {
		id, err := types.ParseID(schemaID)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("schema_id", "invalid schema_id format"))
			return
		}
		q.SchemaID = &id
	}

	if schemaType != "" {
		q.SchemaType = &schemaType
	}

	result, err := h.queries.HandleSchemaExists(r.Context(), q)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, SchemaExistsResponse{
		Exists:     result.Exists,
		SchemaID:   result.SchemaID,
		SchemaType: result.SchemaType,
	})
}

// ResolveSchemaType handles GET /schemas/resolve/{schemaType}
func (h *Handler) ResolveSchemaType(w http.ResponseWriter, r *http.Request) {
	schemaType := chi.URLParam(r, "schemaType")
	if schemaType == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("schema_type is required"))
		return
	}

	result, err := h.queries.HandleResolveSchemaType(r.Context(), query.ResolveSchemaType{
		SchemaType: schemaType,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, ResolveSchemaTypeResponse{
		SchemaType:     result.SchemaType,
		SchemaID:       result.SchemaID,
		CurrentVersion: result.CurrentVersion,
		Status:         result.Status,
	})
}

// ============================================================================
// Request Parsing Helpers
// ============================================================================

func (h *Handler) parseListSchemasRequest(r *http.Request) ListSchemasRequest {
	page, pageSize := h.parsePagination(r)

	var builtInOnly *bool
	if b := r.URL.Query().Get("built_in_only"); b != "" {
		val := b == "true"
		builtInOnly = &val
	}

	return ListSchemasRequest{
		Status:      r.URL.Query().Get("status"),
		IssuerID:    r.URL.Query().Get("issuer_id"),
		BuiltInOnly: builtInOnly,
		Search:      r.URL.Query().Get("search"),
		OrderBy:     r.URL.Query().Get("order_by"),
		OrderDir:    r.URL.Query().Get("order_dir"),
		Page:        page,
		PageSize:    pageSize,
	}
}

func (h *Handler) parsePagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed := parseInt(p); parsed > 0 {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed := parseInt(ps); parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	return page, pageSize
}

func parseInt(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return n
}

// ============================================================================
// Claim Conversion Helpers
// ============================================================================

func (h *Handler) toCommandClaims(reqs []ClaimDefinitionRequest) ([]domain.ClaimDefinition, error) {
	claims := make([]domain.ClaimDefinition, 0, len(reqs))

	for _, req := range reqs {
		claim, err := h.toCommandClaim(req)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	return claims, nil
}

func (h *Handler) toCommandClaim(req ClaimDefinitionRequest) (domain.ClaimDefinition, error) {
	// Parse data type
	dataType, err := domain.ParseDataType(req.DataType)
	if err != nil {
		return domain.ClaimDefinition{}, err
	}

	// Build claim type options
	var claimTypeOpts []domain.ClaimTypeOption
	if req.Constraints != nil {
		c := req.Constraints
		if len(c.AllowedValues) > 0 {
			claimTypeOpts = append(claimTypeOpts, domain.WithAllowedValues(c.AllowedValues...))
		}
		if c.Min != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithMin(*c.Min))
		}
		if c.Max != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithMax(*c.Max))
		}
		if c.MinLength != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithMinLength(*c.MinLength))
		}
		if c.MaxLength != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithMaxLength(*c.MaxLength))
		}
		if c.Pattern != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithPattern(*c.Pattern))
		}
		if c.Format != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithFormat(*c.Format))
		}
	}

	claimType, err := domain.NewClaimType(dataType, claimTypeOpts...)
	if err != nil {
		return domain.ClaimDefinition{}, err
	}

	// Build claim definition options
	var claimOpts []domain.ClaimDefinitionOption
	if req.Required {
		claimOpts = append(claimOpts, domain.Required())
	}
	if req.DisplayName != "" {
		claimOpts = append(claimOpts, domain.WithDisplayName(req.DisplayName))
	}
	if req.Description != "" {
		claimOpts = append(claimOpts, domain.WithDescription(req.Description))
	}
	if req.Disclosable != nil && !*req.Disclosable {
		claimOpts = append(claimOpts, domain.NonDisclosable())
	}
	if req.Order != 0 {
		claimOpts = append(claimOpts, domain.WithOrder(req.Order))
	}

	return domain.NewClaimDefinition(req.Key, claimType, claimOpts...)
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

func toSchemaResponse(v *query.SchemaView) SchemaResponse {
	return SchemaResponse{
		SchemaID:       v.SchemaID,
		SchemaType:     v.SchemaType,
		Name:           v.Name,
		Description:    v.Description,
		CurrentVersion: v.CurrentVersion,
		Status:         v.Status,
		IssuerID:       v.IssuerID,
		Claims:         toClaimResponses(v.Claims),
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func toClaimResponses(views []query.ClaimView) []ClaimResponse {
	responses := make([]ClaimResponse, len(views))
	for i, v := range views {
		responses[i] = toClaimResponse(v)
	}
	return responses
}

func toClaimResponse(v query.ClaimView) ClaimResponse {
	resp := ClaimResponse{
		Key:         v.Key,
		DataType:    v.DataType,
		Required:    v.Required,
		DisplayName: v.DisplayName,
		Description: v.Description,
		Disclosable: v.Disclosable,
		Order:       v.Order,
	}

	if v.Constraints != nil {
		resp.Constraints = &ConstraintsResponse{
			AllowedValues: v.Constraints.AllowedValues,
			Min:           v.Constraints.Min,
			Max:           v.Constraints.Max,
			MinLength:     v.Constraints.MinLength,
			MaxLength:     v.Constraints.MaxLength,
			Pattern:       v.Constraints.Pattern,
			Format:        v.Constraints.Format,
		}
	}

	return resp
}

func toSchemaListResponse(v *query.SchemaListView, page, pageSize int) SchemaListResponse {
	summaries := make([]SchemaSummaryResponse, len(v.Schemas))
	for i, s := range v.Schemas {
		summaries[i] = SchemaSummaryResponse{
			SchemaID:       s.SchemaID,
			SchemaType:     s.SchemaType,
			Name:           s.Name,
			CurrentVersion: s.CurrentVersion,
			Status:         s.Status,
			ClaimCount:     s.ClaimCount,
			CreatedAt:      s.CreatedAt,
		}
	}

	return SchemaListResponse{
		Schemas:    summaries,
		TotalCount: v.TotalCount,
		Page:       page,
		PageSize:   pageSize,
		HasMore:    v.NextCursor != nil,
	}
}

func toSchemaVersionResponse(v *query.SchemaVersionView) SchemaVersionResponse {
	return SchemaVersionResponse{
		SchemaID:   v.SchemaID,
		SchemaType: v.SchemaType,
		Name:       v.Name,
		Version:    v.Version,
		Status:     v.Status,
		Claims:     toClaimResponses(v.Claims),
		IsLatest:   v.IsLatest,
		CreatedAt:  v.CreatedAt,
	}
}

func toVersionListResponse(v *query.VersionListView) VersionListResponse {
	versions := make([]VersionInfoResponse, len(v.Versions))
	for i, ver := range v.Versions {
		versions[i] = VersionInfoResponse{
			Version:   ver.Version,
			IsLatest:  ver.IsLatest,
			CreatedAt: ver.CreatedAt,
		}
	}

	return VersionListResponse{
		SchemaID:       v.SchemaID,
		SchemaType:     v.SchemaType,
		CurrentVersion: v.CurrentVersion,
		Versions:       versions,
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
