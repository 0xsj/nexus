package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/organization/domain"
	"github.com/0xsj/nexus/platform/internal/organization/infrastructure/persistence/postgres/generated"
)

// IdentityReader implements domain.IdentityReader using the local projection table.
type IdentityReader struct {
	queries *generated.Queries
}

// Compile-time check.
var _ domain.IdentityReader = (*IdentityReader)(nil)

// NewIdentityReader creates a new IdentityReader.
func NewIdentityReader(queries *generated.Queries) *IdentityReader {
	return &IdentityReader{queries: queries}
}

// UserExists checks if a user with the given ID exists and is active.
func (r *IdentityReader) UserExists(ctx context.Context, userID string) (bool, error) {
	return r.queries.UserProjectionExists(ctx, userID)
}

// GetUserEmail retrieves a user's email by their ID.
func (r *IdentityReader) GetUserEmail(ctx context.Context, userID string) (string, error) {
	proj, err := r.queries.GetUserProjection(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %s", userID)
	}
	return proj.Email, nil
}
