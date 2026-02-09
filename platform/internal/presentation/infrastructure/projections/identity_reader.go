package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/presentation/domain"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
)

// IdentityReader implements domain.IdentityReader by reading from
// the presentation_user_projections table, which is kept up-to-date
// by the IdentityProjector consuming User.* events.
type IdentityReader struct {
	queries *generated.Queries
}

var _ domain.IdentityReader = (*IdentityReader)(nil)

// NewIdentityReader creates a new IdentityReader.
func NewIdentityReader(queries *generated.Queries) *IdentityReader {
	return &IdentityReader{queries: queries}
}

// GetUserDID returns the primary DID for a user from the local projection.
func (r *IdentityReader) GetUserDID(ctx context.Context, userID string) (string, error) {
	did, err := r.queries.GetUserProjectionDID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user DID not found: %s", userID)
	}
	return did, nil
}
