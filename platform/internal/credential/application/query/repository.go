package query

import (
	"context"
)

// ============================================================================
// Read Repository Interface
// ============================================================================

// CredentialReadRepository defines the interface for credential queries (read side).
type CredentialReadRepository interface {
	// GetByID retrieves a credential view by ID.
	GetByID(ctx context.Context, id string) (*CredentialView, error)

	// List retrieves a paginated list of credentials.
	List(ctx context.Context, opts ListOptions) (*CredentialListResult, error)

	// ListByHolder retrieves credentials for a specific holder.
	ListByHolder(ctx context.Context, holderDID string, opts ListOptions) (*CredentialListResult, error)

	// ListByIssuer retrieves credentials issued by a specific issuer.
	ListByIssuer(ctx context.Context, issuerDID string, opts ListOptions) (*CredentialListResult, error)

	// Count returns the total number of credentials.
	Count(ctx context.Context) (int, error)

	// CountByStatus returns the number of credentials with a specific status.
	CountByStatus(ctx context.Context, status string) (int, error)
}

// ============================================================================
// List Options
// ============================================================================

// ListOptions contains options for listing credentials.
type ListOptions struct {
	Status         string
	CredentialType string
	Limit          int
	Offset         int
	Cursor         string
	SortBy         string
	SortOrder      string
}

// DefaultListOptions returns sensible defaults.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}
