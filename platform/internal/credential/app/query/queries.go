package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetCredential            = "credential.GetCredential"
	QueryListCredentialsBySubject = "credential.ListCredentialsBySubject"
	QueryListCredentialsByIssuer  = "credential.ListCredentialsByIssuer"
)

// ============================================================================
// GetCredential
// ============================================================================

// GetCredential retrieves a single credential by ID.
type GetCredential struct {
	CredentialID types.ID `json:"credential_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetCredential) QueryName() string {
	return QueryGetCredential
}

// Validate implements cqrs.Validatable.
func (q GetCredential) Validate() error {
	if q.CredentialID.IsZero() {
		return cqrs.ErrQueryValidation("GetCredential.Validate", "credential_id is required")
	}
	return nil
}

// ============================================================================
// ListCredentialsBySubject
// ============================================================================

// ListCredentialsBySubject lists credentials for a subject DID.
type ListCredentialsBySubject struct {
	SubjectDID string `json:"subject_did" validate:"required"`
	Limit      int    `json:"limit" validate:"omitempty"`
	Offset     int    `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListCredentialsBySubject) QueryName() string {
	return QueryListCredentialsBySubject
}

// Validate implements cqrs.Validatable.
func (q ListCredentialsBySubject) Validate() error {
	if q.SubjectDID == "" {
		return cqrs.ErrQueryValidation("ListCredentialsBySubject.Validate", "subject_did is required")
	}
	return nil
}

// ============================================================================
// ListCredentialsByIssuer
// ============================================================================

// ListCredentialsByIssuer lists credentials for an issuer DID.
type ListCredentialsByIssuer struct {
	IssuerDID string `json:"issuer_did" validate:"required"`
	Limit     int    `json:"limit" validate:"omitempty"`
	Offset    int    `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListCredentialsByIssuer) QueryName() string {
	return QueryListCredentialsByIssuer
}

// Validate implements cqrs.Validatable.
func (q ListCredentialsByIssuer) Validate() error {
	if q.IssuerDID == "" {
		return cqrs.ErrQueryValidation("ListCredentialsByIssuer.Validate", "issuer_did is required")
	}
	return nil
}
