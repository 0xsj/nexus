package query

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Read Repository Interface
// ============================================================================

// CredentialReadRepository defines the interface for credential queries.
type CredentialReadRepository interface {
	// GetByID retrieves a credential by ID.
	GetByID(ctx context.Context, id string) (*CredentialView, error)

	// List retrieves a paginated list of credentials.
	List(ctx context.Context, opts ListOptions) (*CredentialListResult, error)

	// ListByHolder retrieves credentials for a holder.
	ListByHolder(ctx context.Context, holderDID string, opts ListOptions) (*CredentialListResult, error)

	// ListByIssuer retrieves credentials issued by an issuer.
	ListByIssuer(ctx context.Context, issuerDID string, opts ListOptions) (*CredentialListResult, error)
}

// ListOptions contains common list/filter options.
type ListOptions struct {
	Status         string
	CredentialType string
	Limit          int
	Offset         int
	Cursor         string
	SortBy         string
	SortOrder      string
}

// ============================================================================
// GetCredential Handler
// ============================================================================

// GetCredentialHandler handles GetCredential queries.
type GetCredentialHandler struct {
	repo CredentialReadRepository
}

// NewGetCredentialHandler creates a new GetCredentialHandler.
func NewGetCredentialHandler(repo CredentialReadRepository) *GetCredentialHandler {
	return &GetCredentialHandler{repo: repo}
}

// Handle handles the GetCredential query.
func (h *GetCredentialHandler) Handle(ctx context.Context, q *GetCredential) (*CredentialView, error) {
	return h.repo.GetByID(ctx, q.CredentialID)
}

// ============================================================================
// ListCredentials Handler
// ============================================================================

// ListCredentialsHandler handles ListCredentials queries.
type ListCredentialsHandler struct {
	repo CredentialReadRepository
}

// NewListCredentialsHandler creates a new ListCredentialsHandler.
func NewListCredentialsHandler(repo CredentialReadRepository) *ListCredentialsHandler {
	return &ListCredentialsHandler{repo: repo}
}

// Handle handles the ListCredentials query.
func (h *ListCredentialsHandler) Handle(ctx context.Context, q *ListCredentials) (*CredentialListResult, error) {
	opts := ListOptions{
		Status:         q.Status,
		CredentialType: q.CredentialType,
		Limit:          q.Limit,
		Offset:         q.Offset,
		Cursor:         q.Cursor,
		SortBy:         q.SortBy,
		SortOrder:      q.SortOrder,
	}

	return h.repo.List(ctx, opts)
}

// ============================================================================
// GetCredentialsByHolder Handler
// ============================================================================

// GetCredentialsByHolderHandler handles GetCredentialsByHolder queries.
type GetCredentialsByHolderHandler struct {
	repo CredentialReadRepository
}

// NewGetCredentialsByHolderHandler creates a new GetCredentialsByHolderHandler.
func NewGetCredentialsByHolderHandler(repo CredentialReadRepository) *GetCredentialsByHolderHandler {
	return &GetCredentialsByHolderHandler{repo: repo}
}

// Handle handles the GetCredentialsByHolder query.
func (h *GetCredentialsByHolderHandler) Handle(ctx context.Context, q *GetCredentialsByHolder) (*CredentialListResult, error) {
	opts := ListOptions{
		Status:         q.Status,
		CredentialType: q.CredentialType,
		Limit:          q.Limit,
		Offset:         q.Offset,
	}

	return h.repo.ListByHolder(ctx, q.HolderDID, opts)
}

// ============================================================================
// GetCredentialsByIssuer Handler
// ============================================================================

// GetCredentialsByIssuerHandler handles GetCredentialsByIssuer queries.
type GetCredentialsByIssuerHandler struct {
	repo CredentialReadRepository
}

// NewGetCredentialsByIssuerHandler creates a new GetCredentialsByIssuerHandler.
func NewGetCredentialsByIssuerHandler(repo CredentialReadRepository) *GetCredentialsByIssuerHandler {
	return &GetCredentialsByIssuerHandler{repo: repo}
}

// Handle handles the GetCredentialsByIssuer query.
func (h *GetCredentialsByIssuerHandler) Handle(ctx context.Context, q *GetCredentialsByIssuer) (*CredentialListResult, error) {
	opts := ListOptions{
		Status:         q.Status,
		CredentialType: q.CredentialType,
		Limit:          q.Limit,
		Offset:         q.Offset,
	}

	return h.repo.ListByIssuer(ctx, q.IssuerDID, opts)
}

// ============================================================================
// Handler Registration
// ============================================================================

// RegisterHandlers registers all credential query handlers with the query bus.
func RegisterHandlers(bus *cqrs.InMemoryQueryBus, repo CredentialReadRepository) error {
	handlers := map[string]any{
		TypeGetCredential:          NewGetCredentialHandler(repo),
		TypeListCredentials:        NewListCredentialsHandler(repo),
		TypeGetCredentialsByHolder: NewGetCredentialsByHolderHandler(repo),
		TypeGetCredentialsByIssuer: NewGetCredentialsByIssuerHandler(repo),
	}

	for queryType, handler := range handlers {
		if err := bus.Register(queryType, handler); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.QueryHandler[*GetCredential, *CredentialView]                = (*GetCredentialHandler)(nil)
	_ cqrs.QueryHandler[*ListCredentials, *CredentialListResult]        = (*ListCredentialsHandler)(nil)
	_ cqrs.QueryHandler[*GetCredentialsByHolder, *CredentialListResult] = (*GetCredentialsByHolderHandler)(nil)
	_ cqrs.QueryHandler[*GetCredentialsByIssuer, *CredentialListResult] = (*GetCredentialsByIssuerHandler)(nil)
)
