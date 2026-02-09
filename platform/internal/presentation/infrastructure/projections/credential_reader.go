package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/presentation/domain"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
)

// CredentialReader implements domain.CredentialReader by reading from
// the presentation_credential_projections table, which is kept up-to-date
// by the CredentialProjector consuming Credential.* events.
type CredentialReader struct {
	queries *generated.Queries
}

var _ domain.CredentialReader = (*CredentialReader)(nil)

// NewCredentialReader creates a new CredentialReader.
func NewCredentialReader(queries *generated.Queries) *CredentialReader {
	return &CredentialReader{queries: queries}
}

// GetCredential checks if a credential exists and is active in the local projection.
func (r *CredentialReader) GetCredential(ctx context.Context, credentialID string) error {
	exists, err := r.queries.CredentialProjectionExists(ctx, credentialID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("credential not found: %s", credentialID)
	}
	return nil
}

// GetUserCredentials returns active credential IDs for a user from the local projection.
func (r *CredentialReader) GetUserCredentials(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.queries.ListCredentialProjectionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
