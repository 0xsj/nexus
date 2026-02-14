package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/trust/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
)

// organizationReaderQueries defines the query methods needed by OrganizationReader.
type organizationReaderQueries interface {
	GetOrganizationProjection(ctx context.Context, organizationID string) (generated.TrustOrganizationProjection, error)
}

// OrganizationReader implements domain.OrganizationReader using the local organization projection.
type OrganizationReader struct {
	queries organizationReaderQueries
}

// Compile-time check.
var _ domain.OrganizationReader = (*OrganizationReader)(nil)

// NewOrganizationReader creates a new OrganizationReader.
func NewOrganizationReader(queries *generated.Queries) *OrganizationReader {
	return &OrganizationReader{queries: queries}
}

// IsTrustAnchor checks if an organization is a trust anchor (verified organization).
func (r *OrganizationReader) IsTrustAnchor(ctx context.Context, orgID string) (bool, error) {
	proj, err := r.queries.GetOrganizationProjection(ctx, orgID)
	if err != nil {
		return false, fmt.Errorf("organization not found: %s", orgID)
	}
	return proj.VerificationStatus == "verified" && proj.Active, nil
}
