package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/trust/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
)

// credentialReaderQueries defines the query methods needed by CredentialReader.
type credentialReaderQueries interface {
	CredentialProjectionExists(ctx context.Context, credentialID string) (bool, error)
	CountCredentialProjectionsByUserID(ctx context.Context, userID string) (int32, error)
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

// GetCredential checks if a credential exists. Returns an error if not found.
func (r *CredentialReader) GetCredential(ctx context.Context, credentialID string) error {
	exists, err := r.queries.CredentialProjectionExists(ctx, credentialID)
	if err != nil {
		return fmt.Errorf("checking credential existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("credential not found: %s", credentialID)
	}
	return nil
}

// CountUserCredentials returns the number of credentials for a user.
func (r *CredentialReader) CountUserCredentials(ctx context.Context, userID string) (int, error) {
	count, err := r.queries.CountCredentialProjectionsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
