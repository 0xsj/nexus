package command

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandGiveVouch   = "trust.GiveVouch"
	CommandAcceptVouch = "trust.AcceptVouch"
	CommandRevokeVouch = "trust.RevokeVouch"
)

// ============================================================================
// GiveVouch
// ============================================================================

// GiveVouch creates a new vouch from one user to another.
type GiveVouch struct {
	VoucherID    string     `json:"voucher_id" validate:"required"`
	VoucheeID    string     `json:"vouchee_id" validate:"required"`
	CredentialID string     `json:"credential_id" validate:"omitempty"`
	ClaimKey     string     `json:"claim_key" validate:"omitempty"`
	Relationship string     `json:"relationship" validate:"required"`
	Strength     int        `json:"strength" validate:"required,min=1,max=10"`
	Statement    string     `json:"statement" validate:"omitempty"`
	Context      string     `json:"context" validate:"omitempty"`
	ExpiresAt    *time.Time `json:"expires_at" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c GiveVouch) CommandName() string {
	return CommandGiveVouch
}

// Validate implements cqrs.Validatable.
func (c GiveVouch) Validate() error {
	if c.VoucherID == "" {
		return cqrs.ErrCommandValidation("GiveVouch.Validate", "voucher_id is required")
	}
	if c.VoucheeID == "" {
		return cqrs.ErrCommandValidation("GiveVouch.Validate", "vouchee_id is required")
	}
	if c.VoucherID == c.VoucheeID {
		return cqrs.ErrCommandValidation("GiveVouch.Validate", "cannot vouch for yourself")
	}
	if c.Relationship == "" {
		return cqrs.ErrCommandValidation("GiveVouch.Validate", "relationship is required")
	}
	if c.Strength < 1 || c.Strength > 10 {
		return cqrs.ErrCommandValidation("GiveVouch.Validate", "strength must be between 1 and 10")
	}
	if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
		return cqrs.ErrCommandValidation("GiveVouch.Validate", "expires_at must be in the future")
	}
	return nil
}

// GiveVouchResult is the result data for GiveVouch.
type GiveVouchResult struct {
	VouchID string `json:"vouch_id"`
	Status  string `json:"status"`
}

// ============================================================================
// AcceptVouch
// ============================================================================

// AcceptVouch accepts a pending vouch. Only the vouchee can accept.
type AcceptVouch struct {
	VouchID   types.ID `json:"vouch_id" validate:"required"`
	VoucheeID string   `json:"vouchee_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c AcceptVouch) CommandName() string {
	return CommandAcceptVouch
}

// Validate implements cqrs.Validatable.
func (c AcceptVouch) Validate() error {
	if c.VouchID.IsZero() {
		return cqrs.ErrCommandValidation("AcceptVouch.Validate", "vouch_id is required")
	}
	if c.VoucheeID == "" {
		return cqrs.ErrCommandValidation("AcceptVouch.Validate", "vouchee_id is required")
	}
	return nil
}

// AcceptVouchResult is the result data for AcceptVouch.
type AcceptVouchResult struct {
	VouchID string `json:"vouch_id"`
	Status  string `json:"status"`
}

// ============================================================================
// RevokeVouch
// ============================================================================

// RevokeVouch revokes an existing vouch.
type RevokeVouch struct {
	VouchID types.ID `json:"vouch_id" validate:"required"`
	Reason  string   `json:"reason" validate:"required,max=500"`
}

// CommandName implements cqrs.Command.
func (c RevokeVouch) CommandName() string {
	return CommandRevokeVouch
}

// Validate implements cqrs.Validatable.
func (c RevokeVouch) Validate() error {
	if c.VouchID.IsZero() {
		return cqrs.ErrCommandValidation("RevokeVouch.Validate", "vouch_id is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation("RevokeVouch.Validate", "reason is required")
	}
	if len(c.Reason) > 500 {
		return cqrs.ErrCommandValidation("RevokeVouch.Validate", "reason must be 500 characters or less")
	}
	return nil
}

// RevokeVouchResult is the result data for RevokeVouch.
type RevokeVouchResult struct {
	VouchID string `json:"vouch_id"`
	Status  string `json:"status"`
}
