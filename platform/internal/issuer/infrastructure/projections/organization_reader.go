package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/issuer/domain"
	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
)

// organizationReaderQueries defines the query methods needed by OrganizationReader.
type organizationReaderQueries interface {
	OrganizationProjectionExists(ctx context.Context, organizationID string) (bool, error)
	GetOrganizationProjection(ctx context.Context, organizationID string) (generated.IssuerOrganizationProjection, error)
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

// GetOrganization verifies that an organization exists.
func (r *OrganizationReader) GetOrganization(ctx context.Context, orgID string) error {
	exists, err := r.queries.OrganizationProjectionExists(ctx, orgID)
	if err != nil {
		return fmt.Errorf("checking organization existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("organization not found: %s", orgID)
	}
	return nil
}

// IsVerified checks if an organization is verified.
func (r *OrganizationReader) IsVerified(ctx context.Context, orgID string) (bool, error) {
	proj, err := r.queries.GetOrganizationProjection(ctx, orgID)
	if err != nil {
		return false, fmt.Errorf("organization not found: %s", orgID)
	}
	return proj.VerificationStatus == "verified", nil
}
