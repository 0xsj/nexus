package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/notification/domain"
	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/persistence/postgres/generated"
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

// GetUserEmail returns the email address for the given user ID.
func (r *IdentityReader) GetUserEmail(ctx context.Context, userID string) (string, error) {
	proj, err := r.queries.GetUserProjection(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %s", userID)
	}
	return proj.Email, nil
}

// UserExists checks if a user with the given ID exists and is active.
func (r *IdentityReader) UserExists(ctx context.Context, userID string) (bool, error) {
	return r.queries.UserProjectionExists(ctx, userID)
}
