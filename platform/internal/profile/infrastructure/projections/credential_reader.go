package projections

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/profile/domain"
	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres/generated"
)

// credentialReaderQueries defines the query methods needed by CredentialReader.
type credentialReaderQueries interface {
	GetCredentialProjection(ctx context.Context, credentialID string) (generated.ProfileCredentialProjection, error)
}

// CredentialReader implements domain.CredentialReader using the local credential projection.
type CredentialReader struct {
	queries credentialReaderQueries
}

// Compile-time check.
var _ domain.CredentialReader = (*CredentialReader)(nil)

// NewCredentialReader creates a new CredentialReader.
func NewCredentialReader(queries *generated.Queries) *CredentialReader {
	return &CredentialReader{queries: queries}
}

// GetCredentialByID checks if a credential exists and returns its type.
func (r *CredentialReader) GetCredentialByID(ctx context.Context, credentialID string) (bool, string, error) {
	proj, err := r.queries.GetCredentialProjection(ctx, credentialID)
	if err != nil {
		return false, "", nil
	}
	return true, proj.CredentialType, nil
}
