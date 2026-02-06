package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetPresentation           = "presentation.GetPresentation"
	QueryListPresentationsByHolder = "presentation.ListPresentationsByHolder"
	QueryGetShareLink              = "presentation.GetShareLink"
	QueryGetShareLinkByToken       = "presentation.GetShareLinkByToken"
	QueryListShareLinks            = "presentation.ListShareLinks"
	QueryGetAccessLog              = "presentation.GetAccessLog"
)

// ============================================================================
// GetPresentation
// ============================================================================

// GetPresentation retrieves a single presentation by ID.
type GetPresentation struct {
	PresentationID types.ID `json:"presentation_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetPresentation) QueryName() string {
	return QueryGetPresentation
}

// Validate implements cqrs.Validatable.
func (q GetPresentation) Validate() error {
	if q.PresentationID.IsZero() {
		return cqrs.ErrQueryValidation("GetPresentation.Validate", "presentation_id is required")
	}
	return nil
}

// ============================================================================
// ListPresentationsByHolder
// ============================================================================

// ListPresentationsByHolder lists presentations for a holder DID.
type ListPresentationsByHolder struct {
	HolderDID string `json:"holder_did" validate:"required"`
	Limit     int    `json:"limit" validate:"omitempty"`
	Offset    int    `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListPresentationsByHolder) QueryName() string {
	return QueryListPresentationsByHolder
}

// Validate implements cqrs.Validatable.
func (q ListPresentationsByHolder) Validate() error {
	if q.HolderDID == "" {
		return cqrs.ErrQueryValidation("ListPresentationsByHolder.Validate", "holder_did is required")
	}
	return nil
}

// ============================================================================
// GetShareLink
// ============================================================================

// GetShareLink retrieves a single share link by ID.
type GetShareLink struct {
	ShareLinkID types.ID `json:"share_link_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetShareLink) QueryName() string {
	return QueryGetShareLink
}

// Validate implements cqrs.Validatable.
func (q GetShareLink) Validate() error {
	if q.ShareLinkID.IsZero() {
		return cqrs.ErrQueryValidation("GetShareLink.Validate", "share_link_id is required")
	}
	return nil
}

// ============================================================================
// GetShareLinkByToken
// ============================================================================

// GetShareLinkByToken retrieves a share link by its unique token.
type GetShareLinkByToken struct {
	Token string `json:"token" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetShareLinkByToken) QueryName() string {
	return QueryGetShareLinkByToken
}

// Validate implements cqrs.Validatable.
func (q GetShareLinkByToken) Validate() error {
	if q.Token == "" {
		return cqrs.ErrQueryValidation("GetShareLinkByToken.Validate", "token is required")
	}
	return nil
}

// ============================================================================
// ListShareLinks
// ============================================================================

// ListShareLinks lists share links for a presentation.
type ListShareLinks struct {
	PresentationID types.ID `json:"presentation_id" validate:"required"`
	Limit          int      `json:"limit" validate:"omitempty"`
	Offset         int      `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListShareLinks) QueryName() string {
	return QueryListShareLinks
}

// Validate implements cqrs.Validatable.
func (q ListShareLinks) Validate() error {
	if q.PresentationID.IsZero() {
		return cqrs.ErrQueryValidation("ListShareLinks.Validate", "presentation_id is required")
	}
	return nil
}

// ============================================================================
// GetAccessLog
// ============================================================================

// GetAccessLog retrieves the access log for a share link.
type GetAccessLog struct {
	ShareLinkID types.ID `json:"share_link_id" validate:"required"`
	Limit       int      `json:"limit" validate:"omitempty"`
	Offset      int      `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q GetAccessLog) QueryName() string {
	return QueryGetAccessLog
}

// Validate implements cqrs.Validatable.
func (q GetAccessLog) Validate() error {
	if q.ShareLinkID.IsZero() {
		return cqrs.ErrQueryValidation("GetAccessLog.Validate", "share_link_id is required")
	}
	return nil
}
