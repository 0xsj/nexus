package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/internal/schema/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// issuerReaderQueries defines the query methods needed by IssuerReader.
type issuerReaderQueries interface {
	GetIssuerProjection(ctx context.Context, issuerID string) (generated.SchemaIssuerProjection, error)
	IssuerProjectionExists(ctx context.Context, issuerID string) (bool, error)
}

// IssuerReader implements domain.IssuerReader using the local issuer projection.
type IssuerReader struct {
	queries issuerReaderQueries
}

// Compile-time check.
var _ domain.IssuerReader = (*IssuerReader)(nil)

// NewIssuerReader creates a new IssuerReader.
func NewIssuerReader(queries *generated.Queries) *IssuerReader {
	return &IssuerReader{queries: queries}
}

// GetIssuer retrieves issuer information by ID from the local projection.
func (r *IssuerReader) GetIssuer(ctx context.Context, issuerID types.ID) (*domain.IssuerInfo, error) {
	proj, err := r.queries.GetIssuerProjection(ctx, issuerID.String())
	if err != nil {
		return nil, fmt.Errorf("issuer not found: %s", issuerID.String())
	}
	return &domain.IssuerInfo{
		ID:     issuerID,
		Name:   proj.Name,
		Active: proj.Active,
	}, nil
}

// IssuerExists checks if an issuer exists and is active.
func (r *IssuerReader) IssuerExists(ctx context.Context, issuerID types.ID) (bool, error) {
	return r.queries.IssuerProjectionExists(ctx, issuerID.String())
}
