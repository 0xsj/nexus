package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// Query name constants
const (
	QueryGetVerification         = "verification.GetVerification"
	QueryListVerificationsByUser = "verification.ListVerificationsByUser"
)

// ============================================================================
// GetVerification
// ============================================================================

// GetVerification retrieves a single verification by ID.
type GetVerification struct {
	VerificationID string `json:"verification_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetVerification) QueryName() string {
	return QueryGetVerification
}

// Validate implements cqrs.Validatable.
func (q GetVerification) Validate() error {
	if q.VerificationID == "" {
		return cqrs.ErrQueryValidation("GetVerification.Validate", "verification_id is required")
	}
	return nil
}

// ============================================================================
// ListVerificationsByUser
// ============================================================================

// ListVerificationsByUser lists verifications for a user.
type ListVerificationsByUser struct {
	UserID string `json:"user_id" validate:"required"`
	Limit  int    `json:"limit" validate:"omitempty"`
	Offset int    `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListVerificationsByUser) QueryName() string {
	return QueryListVerificationsByUser
}

// Validate implements cqrs.Validatable.
func (q ListVerificationsByUser) Validate() error {
	if q.UserID == "" {
		return cqrs.ErrQueryValidation("ListVerificationsByUser.Validate", "user_id is required")
	}
	return nil
}
