package query

import (
	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Query Types
// ============================================================================

const (
	TypeGetVerification            = "verification.get"
	TypeGetVerificationByState     = "verification.get_by_state"
	TypeListVerificationsByUser    = "verification.list_by_user"
	TypeGetProviderConnections     = "verification.get_provider_connections"
	TypeGetLatestByUserAndProvider = "verification.get_latest_by_user_provider"
	TypeCheckProviderConnected     = "verification.check_provider_connected"
)

// ============================================================================
// Get Verification Query
// ============================================================================

// GetVerification retrieves a verification by ID.
type GetVerification struct {
	VerificationID string `json:"verification_id"`
}

// QueryName returns the query name.
func (q *GetVerification) QueryName() string {
	return TypeGetVerification
}

// NewGetVerification creates a new GetVerification query.
func NewGetVerification(verificationID string) *GetVerification {
	return &GetVerification{
		VerificationID: verificationID,
	}
}

// ============================================================================
// Get Verification By State Query
// ============================================================================

// GetVerificationByState retrieves a verification by OAuth state.
type GetVerificationByState struct {
	State string `json:"state"`
}

// QueryName returns the query name.
func (q *GetVerificationByState) QueryName() string {
	return TypeGetVerificationByState
}

// NewGetVerificationByState creates a new GetVerificationByState query.
func NewGetVerificationByState(state string) *GetVerificationByState {
	return &GetVerificationByState{
		State: state,
	}
}

// ============================================================================
// List Verifications By User Query
// ============================================================================

// ListVerificationsByUser retrieves verifications for a user.
type ListVerificationsByUser struct {
	UserID   string                     `json:"user_id"`
	Provider *domain.Provider           `json:"provider,omitempty"`
	Status   *domain.VerificationStatus `json:"status,omitempty"`
	Limit    int                        `json:"limit"`
	Offset   int                        `json:"offset"`
	SortBy   string                     `json:"sort_by,omitempty"`
	SortDesc bool                       `json:"sort_desc,omitempty"`
}

// QueryName returns the query name.
func (q *ListVerificationsByUser) QueryName() string {
	return TypeListVerificationsByUser
}

// NewListVerificationsByUser creates a new ListVerificationsByUser query with defaults.
func NewListVerificationsByUser(userID string) *ListVerificationsByUser {
	return &ListVerificationsByUser{
		UserID:   userID,
		Limit:    20,
		Offset:   0,
		SortBy:   "initiated_at",
		SortDesc: true,
	}
}

// WithProvider filters by provider.
func (q *ListVerificationsByUser) WithProvider(provider domain.Provider) *ListVerificationsByUser {
	q.Provider = &provider
	return q
}

// WithStatus filters by status.
func (q *ListVerificationsByUser) WithStatus(status domain.VerificationStatus) *ListVerificationsByUser {
	q.Status = &status
	return q
}

// WithLimit sets the limit.
func (q *ListVerificationsByUser) WithLimit(limit int) *ListVerificationsByUser {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *ListVerificationsByUser) WithOffset(offset int) *ListVerificationsByUser {
	q.Offset = offset
	return q
}

// WithSort sets the sort field and direction.
func (q *ListVerificationsByUser) WithSort(sortBy string, desc bool) *ListVerificationsByUser {
	q.SortBy = sortBy
	q.SortDesc = desc
	return q
}

// ============================================================================
// Get Provider Connections Query
// ============================================================================

// GetProviderConnections retrieves all provider connection statuses for a user.
type GetProviderConnections struct {
	UserID string `json:"user_id"`
}

// QueryName returns the query name.
func (q *GetProviderConnections) QueryName() string {
	return TypeGetProviderConnections
}

// NewGetProviderConnections creates a new GetProviderConnections query.
func NewGetProviderConnections(userID string) *GetProviderConnections {
	return &GetProviderConnections{
		UserID: userID,
	}
}

// ============================================================================
// Get Latest By User And Provider Query
// ============================================================================

// GetLatestByUserAndProvider retrieves the most recent verification for a user and provider.
type GetLatestByUserAndProvider struct {
	UserID   string          `json:"user_id"`
	Provider domain.Provider `json:"provider"`
}

// QueryName returns the query name.
func (q *GetLatestByUserAndProvider) QueryName() string {
	return TypeGetLatestByUserAndProvider
}

// NewGetLatestByUserAndProvider creates a new GetLatestByUserAndProvider query.
func NewGetLatestByUserAndProvider(userID string, provider domain.Provider) *GetLatestByUserAndProvider {
	return &GetLatestByUserAndProvider{
		UserID:   userID,
		Provider: provider,
	}
}

// ============================================================================
// Check Provider Connected Query
// ============================================================================

// CheckProviderConnected checks if a user has a verified connection to a provider.
type CheckProviderConnected struct {
	UserID   string          `json:"user_id"`
	Provider domain.Provider `json:"provider"`
}

// QueryName returns the query name.
func (q *CheckProviderConnected) QueryName() string {
	return TypeCheckProviderConnected
}

// NewCheckProviderConnected creates a new CheckProviderConnected query.
func NewCheckProviderConnected(userID string, provider domain.Provider) *CheckProviderConnected {
	return &CheckProviderConnected{
		UserID:   userID,
		Provider: provider,
	}
}

// CheckProviderConnectedResult is the result of checking provider connection.
type CheckProviderConnectedResult struct {
	Connected    bool    `json:"connected"`
	CredentialID *string `json:"credential_id,omitempty"`
	Username     *string `json:"username,omitempty"`
}

// ============================================================================
// List Options
// ============================================================================

// ListOptions contains pagination and sorting options.
type ListOptions struct {
	Limit    int
	Offset   int
	SortBy   string
	SortDesc bool
}

// DefaultListOptions returns default list options.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Limit:    20,
		Offset:   0,
		SortBy:   "initiated_at",
		SortDesc: true,
	}
}

// ============================================================================
// List Result
// ============================================================================

// VerificationListResult contains a paginated list of verifications.
type VerificationListResult struct {
	Items   []*domain.VerificationSummary `json:"items"`
	Total   int                           `json:"total"`
	Limit   int                           `json:"limit"`
	Offset  int                           `json:"offset"`
	HasMore bool                          `json:"has_more"`
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Query = (*GetVerification)(nil)
	_ cqrs.Query = (*GetVerificationByState)(nil)
	_ cqrs.Query = (*ListVerificationsByUser)(nil)
	_ cqrs.Query = (*GetProviderConnections)(nil)
	_ cqrs.Query = (*GetLatestByUserAndProvider)(nil)
	_ cqrs.Query = (*CheckProviderConnected)(nil)
)
