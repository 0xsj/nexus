package query

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Query Types
// ============================================================================

const (
	TypeGetCredential          = "credential.get"
	TypeListCredentials        = "credential.list"
	TypeGetCredentialsByHolder = "credential.list_by_holder"
	TypeGetCredentialsByIssuer = "credential.list_by_issuer"
)

// ============================================================================
// GetCredential Query
// ============================================================================

// GetCredential retrieves a single credential by ID.
type GetCredential struct {
	cqrs.BaseQuery

	CredentialID string `json:"credential_id"`
}

// QueryType returns the query type.
func (q *GetCredential) QueryType() string {
	return TypeGetCredential
}

// NewGetCredential creates a new GetCredential query.
func NewGetCredential(credentialID string) *GetCredential {
	return &GetCredential{
		CredentialID: credentialID,
	}
}

// ============================================================================
// ListCredentials Query
// ============================================================================

// ListCredentials retrieves a paginated list of credentials.
type ListCredentials struct {
	cqrs.BaseQuery

	// Filters
	Status         string `json:"status,omitempty"`
	CredentialType string `json:"credential_type,omitempty"`

	// Pagination
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Cursor string `json:"cursor,omitempty"`

	// Sorting
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

// QueryType returns the query type.
func (q *ListCredentials) QueryType() string {
	return TypeListCredentials
}

// NewListCredentials creates a new ListCredentials query.
func NewListCredentials() *ListCredentials {
	return &ListCredentials{
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// WithStatus filters by status.
func (q *ListCredentials) WithStatus(status string) *ListCredentials {
	q.Status = status
	return q
}

// WithCredentialType filters by credential type.
func (q *ListCredentials) WithCredentialType(credentialType string) *ListCredentials {
	q.CredentialType = credentialType
	return q
}

// WithLimit sets the limit.
func (q *ListCredentials) WithLimit(limit int) *ListCredentials {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *ListCredentials) WithOffset(offset int) *ListCredentials {
	q.Offset = offset
	return q
}

// WithCursor sets the cursor for cursor-based pagination.
func (q *ListCredentials) WithCursor(cursor string) *ListCredentials {
	q.Cursor = cursor
	return q
}

// WithSort sets sorting options.
func (q *ListCredentials) WithSort(sortBy, sortOrder string) *ListCredentials {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// ============================================================================
// GetCredentialsByHolder Query
// ============================================================================

// GetCredentialsByHolder retrieves all credentials for a holder.
type GetCredentialsByHolder struct {
	cqrs.BaseQuery

	HolderDID string `json:"holder_did"`

	// Filters
	Status         string `json:"status,omitempty"`
	CredentialType string `json:"credential_type,omitempty"`

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// QueryType returns the query type.
func (q *GetCredentialsByHolder) QueryType() string {
	return TypeGetCredentialsByHolder
}

// NewGetCredentialsByHolder creates a new GetCredentialsByHolder query.
func NewGetCredentialsByHolder(holderDID string) *GetCredentialsByHolder {
	return &GetCredentialsByHolder{
		HolderDID: holderDID,
		Limit:     20,
	}
}

// WithStatus filters by status.
func (q *GetCredentialsByHolder) WithStatus(status string) *GetCredentialsByHolder {
	q.Status = status
	return q
}

// WithCredentialType filters by credential type.
func (q *GetCredentialsByHolder) WithCredentialType(credentialType string) *GetCredentialsByHolder {
	q.CredentialType = credentialType
	return q
}

// WithLimit sets the limit.
func (q *GetCredentialsByHolder) WithLimit(limit int) *GetCredentialsByHolder {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *GetCredentialsByHolder) WithOffset(offset int) *GetCredentialsByHolder {
	q.Offset = offset
	return q
}

// ============================================================================
// GetCredentialsByIssuer Query
// ============================================================================

// GetCredentialsByIssuer retrieves all credentials issued by an issuer.
type GetCredentialsByIssuer struct {
	cqrs.BaseQuery

	IssuerDID string `json:"issuer_did"`

	// Filters
	Status         string `json:"status,omitempty"`
	CredentialType string `json:"credential_type,omitempty"`

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// QueryType returns the query type.
func (q *GetCredentialsByIssuer) QueryType() string {
	return TypeGetCredentialsByIssuer
}

// NewGetCredentialsByIssuer creates a new GetCredentialsByIssuer query.
func NewGetCredentialsByIssuer(issuerDID string) *GetCredentialsByIssuer {
	return &GetCredentialsByIssuer{
		IssuerDID: issuerDID,
		Limit:     20,
	}
}

// WithStatus filters by status.
func (q *GetCredentialsByIssuer) WithStatus(status string) *GetCredentialsByIssuer {
	q.Status = status
	return q
}

// WithCredentialType filters by credential type.
func (q *GetCredentialsByIssuer) WithCredentialType(credentialType string) *GetCredentialsByIssuer {
	q.CredentialType = credentialType
	return q
}

// WithLimit sets the limit.
func (q *GetCredentialsByIssuer) WithLimit(limit int) *GetCredentialsByIssuer {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *GetCredentialsByIssuer) WithOffset(offset int) *GetCredentialsByIssuer {
	q.Offset = offset
	return q
}

// ============================================================================
// Query Results (Read Models)
// ============================================================================

// CredentialView is the read model for a credential.
type CredentialView struct {
	ID             string         `json:"id"`
	CredentialType string         `json:"credential_type"`
	SchemaID       string         `json:"schema_id,omitempty"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	Status         string         `json:"status"`
	Claims         map[string]any `json:"claims,omitempty"`
	IssuedAt       *time.Time     `json:"issued_at,omitempty"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	RevokedAt      *time.Time     `json:"revoked_at,omitempty"`
	SuspendedAt    *time.Time     `json:"suspended_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Version        int            `json:"version"`
}

// CredentialListResult is the result of a list query.
type CredentialListResult struct {
	Credentials []*CredentialView `json:"credentials"`
	Total       int               `json:"total"`
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
	HasMore     bool              `json:"has_more"`
	NextCursor  string            `json:"next_cursor,omitempty"`
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Query = (*GetCredential)(nil)
	_ cqrs.Query = (*ListCredentials)(nil)
	_ cqrs.Query = (*GetCredentialsByHolder)(nil)
	_ cqrs.Query = (*GetCredentialsByIssuer)(nil)
)
