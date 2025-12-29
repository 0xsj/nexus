package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Query Types
// ============================================================================

const (
	TypeGetCredential          = "credential.get"
	TypeListCredentials        = "credential.list"
	TypeGetCredentialsByHolder = "credential.get_by_holder"
	TypeGetCredentialsByIssuer = "credential.get_by_issuer"
)

// ============================================================================
// Get Credential Query
// ============================================================================

// GetCredential retrieves a single credential by ID.
type GetCredential struct {
	CredentialID string `json:"credential_id"`
}

// QueryName returns the query name.
func (q *GetCredential) QueryName() string {
	return TypeGetCredential
}

// NewGetCredential creates a new GetCredential query.
func NewGetCredential(credentialID string) *GetCredential {
	return &GetCredential{
		CredentialID: credentialID,
	}
}

// ============================================================================
// List Credentials Query
// ============================================================================

// ListCredentials retrieves a paginated list of credentials.
type ListCredentials struct {
	Status         string `json:"status,omitempty"`
	CredentialType string `json:"credential_type,omitempty"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"`
	Cursor         string `json:"cursor,omitempty"`
	SortBy         string `json:"sort_by,omitempty"`
	SortOrder      string `json:"sort_order,omitempty"`
}

// QueryName returns the query name.
func (q *ListCredentials) QueryName() string {
	return TypeListCredentials
}

// NewListCredentials creates a new ListCredentials query with defaults.
func NewListCredentials() *ListCredentials {
	defaults := DefaultListOptions()
	return &ListCredentials{
		Limit:     defaults.Limit,
		Offset:    defaults.Offset,
		SortBy:    defaults.SortBy,
		SortOrder: defaults.SortOrder,
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

// WithSort sets the sort field and order.
func (q *ListCredentials) WithSort(sortBy, sortOrder string) *ListCredentials {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// ============================================================================
// Get Credentials By Holder Query
// ============================================================================

// GetCredentialsByHolder retrieves credentials for a specific holder.
type GetCredentialsByHolder struct {
	HolderDID      string `json:"holder_did"`
	Status         string `json:"status,omitempty"`
	CredentialType string `json:"credential_type,omitempty"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"`
}

// QueryName returns the query name.
func (q *GetCredentialsByHolder) QueryName() string {
	return TypeGetCredentialsByHolder
}

// NewGetCredentialsByHolder creates a new GetCredentialsByHolder query.
func NewGetCredentialsByHolder(holderDID string) *GetCredentialsByHolder {
	defaults := DefaultListOptions()
	return &GetCredentialsByHolder{
		HolderDID: holderDID,
		Limit:     defaults.Limit,
		Offset:    defaults.Offset,
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
// Get Credentials By Issuer Query
// ============================================================================

// GetCredentialsByIssuer retrieves credentials issued by a specific issuer.
type GetCredentialsByIssuer struct {
	IssuerDID      string `json:"issuer_did"`
	Status         string `json:"status,omitempty"`
	CredentialType string `json:"credential_type,omitempty"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"`
}

// QueryName returns the query name.
func (q *GetCredentialsByIssuer) QueryName() string {
	return TypeGetCredentialsByIssuer
}

// NewGetCredentialsByIssuer creates a new GetCredentialsByIssuer query.
func NewGetCredentialsByIssuer(issuerDID string) *GetCredentialsByIssuer {
	defaults := DefaultListOptions()
	return &GetCredentialsByIssuer{
		IssuerDID: issuerDID,
		Limit:     defaults.Limit,
		Offset:    defaults.Offset,
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
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Query = (*GetCredential)(nil)
	_ cqrs.Query = (*ListCredentials)(nil)
	_ cqrs.Query = (*GetCredentialsByHolder)(nil)
	_ cqrs.Query = (*GetCredentialsByIssuer)(nil)
)
