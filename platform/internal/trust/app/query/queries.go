package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetVouch             = "trust.GetVouch"
	QueryListVouchesByVouchee = "trust.ListVouchesByVouchee"
	QueryListVouchesByVoucher = "trust.ListVouchesByVoucher"
	QueryGetReputation        = "trust.GetReputation"
)

// ============================================================================
// GetVouch
// ============================================================================

// GetVouch retrieves a single vouch by ID.
type GetVouch struct {
	VouchID types.ID `json:"vouch_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetVouch) QueryName() string {
	return QueryGetVouch
}

// Validate implements cqrs.Validatable.
func (q GetVouch) Validate() error {
	if q.VouchID.IsZero() {
		return cqrs.ErrQueryValidation("GetVouch.Validate", "vouch_id is required")
	}
	return nil
}

// ============================================================================
// ListVouchesByVouchee
// ============================================================================

// ListVouchesByVouchee lists vouches received by a user.
type ListVouchesByVouchee struct {
	VoucheeID string `json:"vouchee_id" validate:"required"`
	Limit     int    `json:"limit" validate:"omitempty"`
	Offset    int    `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListVouchesByVouchee) QueryName() string {
	return QueryListVouchesByVouchee
}

// Validate implements cqrs.Validatable.
func (q ListVouchesByVouchee) Validate() error {
	if q.VoucheeID == "" {
		return cqrs.ErrQueryValidation("ListVouchesByVouchee.Validate", "vouchee_id is required")
	}
	return nil
}

// ============================================================================
// ListVouchesByVoucher
// ============================================================================

// ListVouchesByVoucher lists vouches given by a user.
type ListVouchesByVoucher struct {
	VoucherID string `json:"voucher_id" validate:"required"`
	Limit     int    `json:"limit" validate:"omitempty"`
	Offset    int    `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListVouchesByVoucher) QueryName() string {
	return QueryListVouchesByVoucher
}

// Validate implements cqrs.Validatable.
func (q ListVouchesByVoucher) Validate() error {
	if q.VoucherID == "" {
		return cqrs.ErrQueryValidation("ListVouchesByVoucher.Validate", "voucher_id is required")
	}
	return nil
}

// ============================================================================
// GetReputation
// ============================================================================

// GetReputation retrieves a user's reputation score.
type GetReputation struct {
	UserID string `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetReputation) QueryName() string {
	return QueryGetReputation
}

// Validate implements cqrs.Validatable.
func (q GetReputation) Validate() error {
	if q.UserID == "" {
		return cqrs.ErrQueryValidation("GetReputation.Validate", "user_id is required")
	}
	return nil
}
