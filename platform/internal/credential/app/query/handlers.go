package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Credential context.
type Handlers struct {
	lookup *postgres.CredentialLookup
	logger log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	lookup *postgres.CredentialLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		lookup: lookup,
		logger: logger,
	}
}

// ============================================================================
// GetCredential Handler
// ============================================================================

// HandleGetCredential handles the GetCredential query.
func (h *Handlers) HandleGetCredential(ctx context.Context, q GetCredential) (*CredentialView, error) {
	const op = "Handlers.HandleGetCredential"

	proj, err := h.lookup.GetByID(ctx, q.CredentialID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// ListCredentialsBySubject Handler
// ============================================================================

// HandleListCredentialsBySubject handles the ListCredentialsBySubject query.
func (h *Handlers) HandleListCredentialsBySubject(ctx context.Context, q ListCredentialsBySubject) (*CredentialListView, error) {
	const op = "Handlers.HandleListCredentialsBySubject"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	credentials, err := h.lookup.ListBySubjectDID(ctx, q.SubjectDID, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.lookup.CountBySubjectDID(ctx, q.SubjectDID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]CredentialSummaryView, len(credentials))
	for i, cred := range credentials {
		summaries[i] = mapProjectionToSummary(&cred)
	}

	return &CredentialListView{
		Credentials: summaries,
		TotalCount:  totalCount,
		Limit:       limit,
		Offset:      q.Offset,
		HasMore:     q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// ListCredentialsByIssuer Handler
// ============================================================================

// HandleListCredentialsByIssuer handles the ListCredentialsByIssuer query.
func (h *Handlers) HandleListCredentialsByIssuer(ctx context.Context, q ListCredentialsByIssuer) (*CredentialListView, error) {
	const op = "Handlers.HandleListCredentialsByIssuer"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	credentials, err := h.lookup.ListByIssuerDID(ctx, q.IssuerDID, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.lookup.CountByIssuerDID(ctx, q.IssuerDID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]CredentialSummaryView, len(credentials))
	for i, cred := range credentials {
		summaries[i] = mapProjectionToSummary(&cred)
	}

	return &CredentialListView{
		Credentials: summaries,
		TotalCount:  totalCount,
		Limit:       limit,
		Offset:      q.Offset,
		HasMore:     q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapProjectionToView maps a CredentialProjection to a CredentialView.
func mapProjectionToView(proj *postgres.CredentialProjection) *CredentialView {
	return &CredentialView{
		CredentialID:     proj.ID,
		CredentialType:   proj.CredentialType,
		IssuerDID:        proj.IssuerDID,
		SubjectDID:       proj.SubjectDID,
		Claims:           proj.Claims,
		IssuedAt:         proj.IssuedAt,
		ExpiresAt:        proj.ExpiresAt,
		Status:           proj.Status,
		RevokedAt:        proj.RevokedAt,
		RevocationReason: proj.RevocationReason,
		JWT:              proj.JWT,
		VerificationID:   proj.VerificationID,
		CreatedAt:        proj.CreatedAt,
		UpdatedAt:        proj.UpdatedAt,
	}
}

// mapProjectionToSummary maps a CredentialProjection to a CredentialSummaryView.
func mapProjectionToSummary(proj *postgres.CredentialProjection) CredentialSummaryView {
	return CredentialSummaryView{
		CredentialID:   proj.ID,
		CredentialType: proj.CredentialType,
		IssuerDID:      proj.IssuerDID,
		SubjectDID:     proj.SubjectDID,
		Status:         proj.Status,
		IssuedAt:       proj.IssuedAt,
		ExpiresAt:      proj.ExpiresAt,
	}
}
