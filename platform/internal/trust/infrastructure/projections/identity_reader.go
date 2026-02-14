package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/trust/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
)

// identityReaderQueries defines the query methods needed by IdentityReader.
type identityReaderQueries interface {
	UserProjectionExists(ctx context.Context, userID string) (bool, error)
	CountCredentialProjectionsByUserID(ctx context.Context, userID string) (int32, error)
}

// IdentityReader implements domain.IdentityReader using the local user projection.
type IdentityReader struct {
	queries identityReaderQueries
}

// Compile-time check.
var _ domain.IdentityReader = (*IdentityReader)(nil)

// NewIdentityReader creates a new IdentityReader.
func NewIdentityReader(queries *generated.Queries) *IdentityReader {
	return &IdentityReader{queries: queries}
}

// GetUser checks if a user exists. Returns an error if the user is not found.
func (r *IdentityReader) GetUser(ctx context.Context, userID string) error {
	exists, err := r.queries.UserProjectionExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("checking user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}
	return nil
}

// HasVerifiedCredential checks if a user has at least one verified credential.
func (r *IdentityReader) HasVerifiedCredential(ctx context.Context, userID string) (bool, error) {
	count, err := r.queries.CountCredentialProjectionsByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
