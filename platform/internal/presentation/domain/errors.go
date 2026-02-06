// Package domain contains the core business logic for the Presentation bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Presentation-specific)
// ============================================================================

const (
	CodePresentationNotFound pkgerrors.Code = "PRESENTATION_NOT_FOUND"
	CodePresentationRevoked  pkgerrors.Code = "PRESENTATION_REVOKED"
	CodePresentationInvalid  pkgerrors.Code = "PRESENTATION_INVALID"

	CodeShareLinkNotFound        pkgerrors.Code = "SHARE_LINK_NOT_FOUND"
	CodeShareLinkRevoked         pkgerrors.Code = "SHARE_LINK_REVOKED"
	CodeShareLinkExpired         pkgerrors.Code = "SHARE_LINK_EXPIRED"
	CodeShareLinkMaxViewsReached pkgerrors.Code = "SHARE_LINK_MAX_VIEWS_REACHED"
	CodeShareLinkInvalid         pkgerrors.Code = "SHARE_LINK_INVALID"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrPresentationNotFound = errors.New("presentation not found")
	ErrPresentationRevoked  = errors.New("presentation has been revoked")
	ErrPresentationInvalid  = errors.New("presentation is invalid")

	ErrShareLinkNotFound        = errors.New("share link not found")
	ErrShareLinkRevoked         = errors.New("share link has been revoked")
	ErrShareLinkExpired         = errors.New("share link has expired")
	ErrShareLinkMaxViewsReached = errors.New("share link max views reached")
	ErrShareLinkInvalid         = errors.New("share link is invalid")
)

// ============================================================================
// Error Constructors
// ============================================================================

// PresentationNotFound creates a presentation not found error.
func PresentationNotFound(operation string, presentationID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "presentation").
		WithCode(CodePresentationNotFound).
		WithMeta("presentation_id", presentationID)
}

// PresentationRevoked creates a presentation revoked error.
func PresentationRevoked(operation string, presentationID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "presentation has been revoked").
		WithCode(CodePresentationRevoked).
		WithMeta("presentation_id", presentationID)
}

// PresentationInvalid creates a presentation invalid error.
func PresentationInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "presentation is invalid: "+reason).
		WithCode(CodePresentationInvalid)
}

// ShareLinkNotFound creates a share link not found error.
func ShareLinkNotFound(operation string, shareLinkID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "share link").
		WithCode(CodeShareLinkNotFound).
		WithMeta("share_link_id", shareLinkID)
}

// ShareLinkRevoked creates a share link revoked error.
func ShareLinkRevoked(operation string, shareLinkID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "share link has been revoked").
		WithCode(CodeShareLinkRevoked).
		WithMeta("share_link_id", shareLinkID)
}

// ShareLinkExpiredErr creates a share link expired error.
func ShareLinkExpiredErr(operation string, shareLinkID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "share link has expired").
		WithCode(CodeShareLinkExpired).
		WithMeta("share_link_id", shareLinkID)
}

// ShareLinkMaxViewsReached creates a share link max views reached error.
func ShareLinkMaxViewsReached(operation string, shareLinkID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "share link max views reached").
		WithCode(CodeShareLinkMaxViewsReached).
		WithMeta("share_link_id", shareLinkID)
}

// ShareLinkInvalid creates a share link invalid error.
func ShareLinkInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "share link is invalid: "+reason).
		WithCode(CodeShareLinkInvalid)
}
