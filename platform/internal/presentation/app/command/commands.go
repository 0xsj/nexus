package command

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandCreatePresentation = "presentation.CreatePresentation"
	CommandRevokePresentation = "presentation.RevokePresentation"
	CommandCreateShareLink    = "presentation.CreateShareLink"
	CommandAccessShareLink    = "presentation.AccessShareLink"
	CommandRevokeShareLink    = "presentation.RevokeShareLink"
)

// ============================================================================
// CreatePresentation
// ============================================================================

// CreatePresentation creates a new verifiable presentation.
type CreatePresentation struct {
	HolderDID            string   `json:"holder_did" validate:"required"`
	CredentialIDs        []string `json:"credential_ids" validate:"required"`
	DisclosurePolicyType string   `json:"disclosure_policy_type" validate:"required"`
	AllowedClaims        []string `json:"allowed_claims,omitempty"`
	BlockedClaims        []string `json:"blocked_claims,omitempty"`
	Purpose              string   `json:"purpose,omitempty"`
}

// CommandName implements cqrs.Command.
func (c CreatePresentation) CommandName() string {
	return CommandCreatePresentation
}

// Validate implements cqrs.Validatable.
func (c CreatePresentation) Validate() error {
	if c.HolderDID == "" {
		return cqrs.ErrCommandValidation("CreatePresentation.Validate", "holder_did is required")
	}
	if len(c.CredentialIDs) == 0 {
		return cqrs.ErrCommandValidation("CreatePresentation.Validate", "credential_ids are required")
	}
	if c.DisclosurePolicyType == "" {
		return cqrs.ErrCommandValidation("CreatePresentation.Validate", "disclosure_policy_type is required")
	}
	return nil
}

// CreatePresentationResult is the result data for CreatePresentation.
type CreatePresentationResult struct {
	PresentationID string `json:"presentation_id"`
	Status         string `json:"status"`
}

// ============================================================================
// RevokePresentation
// ============================================================================

// RevokePresentation revokes an active presentation.
type RevokePresentation struct {
	PresentationID types.ID `json:"presentation_id" validate:"required"`
	Reason         string   `json:"reason" validate:"required,max=500"`
}

// CommandName implements cqrs.Command.
func (c RevokePresentation) CommandName() string {
	return CommandRevokePresentation
}

// Validate implements cqrs.Validatable.
func (c RevokePresentation) Validate() error {
	if c.PresentationID.IsZero() {
		return cqrs.ErrCommandValidation("RevokePresentation.Validate", "presentation_id is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation("RevokePresentation.Validate", "reason is required")
	}
	if len(c.Reason) > 500 {
		return cqrs.ErrCommandValidation("RevokePresentation.Validate", "reason must be 500 characters or less")
	}
	return nil
}

// RevokePresentationResult is the result data for RevokePresentation.
type RevokePresentationResult struct {
	PresentationID string `json:"presentation_id"`
	Status         string `json:"status"`
}

// ============================================================================
// CreateShareLink
// ============================================================================

// CreateShareLink creates a new share link for a presentation.
type CreateShareLink struct {
	PresentationID types.ID   `json:"presentation_id" validate:"required"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	MaxViews       int        `json:"max_views,omitempty"`
	Pin            string     `json:"pin,omitempty"`
	Audience       string     `json:"audience,omitempty"`
}

// CommandName implements cqrs.Command.
func (c CreateShareLink) CommandName() string {
	return CommandCreateShareLink
}

// Validate implements cqrs.Validatable.
func (c CreateShareLink) Validate() error {
	if c.PresentationID.IsZero() {
		return cqrs.ErrCommandValidation("CreateShareLink.Validate", "presentation_id is required")
	}
	if c.MaxViews < 0 {
		return cqrs.ErrCommandValidation("CreateShareLink.Validate", "max_views cannot be negative")
	}
	if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
		return cqrs.ErrCommandValidation("CreateShareLink.Validate", "expires_at must be in the future")
	}
	return nil
}

// CreateShareLinkResult is the result data for CreateShareLink.
type CreateShareLinkResult struct {
	ShareLinkID    string `json:"share_link_id"`
	PresentationID string `json:"presentation_id"`
	Token          string `json:"token"`
	Status         string `json:"status"`
}

// ============================================================================
// AccessShareLink
// ============================================================================

// AccessShareLink records an access to a share link.
type AccessShareLink struct {
	ShareLinkID types.ID `json:"share_link_id" validate:"required"`
	VerifierDID string   `json:"verifier_did,omitempty"`
	IPAddress   string   `json:"ip_address,omitempty"`
}

// CommandName implements cqrs.Command.
func (c AccessShareLink) CommandName() string {
	return CommandAccessShareLink
}

// Validate implements cqrs.Validatable.
func (c AccessShareLink) Validate() error {
	if c.ShareLinkID.IsZero() {
		return cqrs.ErrCommandValidation("AccessShareLink.Validate", "share_link_id is required")
	}
	return nil
}

// AccessShareLinkResult is the result data for AccessShareLink.
type AccessShareLinkResult struct {
	ShareLinkID  string `json:"share_link_id"`
	CurrentViews int    `json:"current_views"`
	Status       string `json:"status"`
}

// ============================================================================
// RevokeShareLink
// ============================================================================

// RevokeShareLink revokes an active share link.
type RevokeShareLink struct {
	ShareLinkID types.ID `json:"share_link_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c RevokeShareLink) CommandName() string {
	return CommandRevokeShareLink
}

// Validate implements cqrs.Validatable.
func (c RevokeShareLink) Validate() error {
	if c.ShareLinkID.IsZero() {
		return cqrs.ErrCommandValidation("RevokeShareLink.Validate", "share_link_id is required")
	}
	return nil
}

// RevokeShareLinkResult is the result data for RevokeShareLink.
type RevokeShareLinkResult struct {
	ShareLinkID string `json:"share_link_id"`
	Status      string `json:"status"`
}
