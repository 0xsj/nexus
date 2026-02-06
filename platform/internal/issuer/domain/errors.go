// Package domain contains the core business logic for the Issuer bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Issuer-specific)
// ============================================================================

const (
	CodeIssuerNotFound      pkgerrors.Code = "ISSUER_NOT_FOUND"
	CodeIssuerNotActive     pkgerrors.Code = "ISSUER_NOT_ACTIVE"
	CodeIssuerSuspended     pkgerrors.Code = "ISSUER_SUSPENDED"
	CodeIssuerInvalid       pkgerrors.Code = "ISSUER_INVALID"
	CodeIssuerAlreadyActive pkgerrors.Code = "ISSUER_ALREADY_ACTIVE"
	CodeTemplateNotFound    pkgerrors.Code = "TEMPLATE_NOT_FOUND"
	CodeTemplateArchived    pkgerrors.Code = "TEMPLATE_ARCHIVED"
	CodeTemplateInvalid     pkgerrors.Code = "TEMPLATE_INVALID"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrIssuerNotFound      = errors.New("issuer not found")
	ErrIssuerNotActive     = errors.New("issuer is not active")
	ErrIssuerSuspended     = errors.New("issuer has been suspended")
	ErrIssuerInvalid       = errors.New("issuer is invalid")
	ErrIssuerAlreadyActive = errors.New("issuer is already active")
	ErrTemplateNotFound    = errors.New("template not found")
	ErrTemplateArchived    = errors.New("template has been archived")
	ErrTemplateInvalid     = errors.New("template is invalid")
)

// ============================================================================
// Error Constructors
// ============================================================================

// IssuerNotFound creates an issuer not found error.
func IssuerNotFound(operation string, issuerID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "issuer").
		WithCode(CodeIssuerNotFound).
		WithMeta("issuer_id", issuerID)
}

// IssuerNotActive creates an issuer not active error.
func IssuerNotActive(operation string, issuerID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "issuer is not active").
		WithCode(CodeIssuerNotActive).
		WithMeta("issuer_id", issuerID)
}

// IssuerSuspendedErr creates an issuer suspended error.
func IssuerSuspendedErr(operation string, issuerID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "issuer has been suspended").
		WithCode(CodeIssuerSuspended).
		WithMeta("issuer_id", issuerID)
}

// IssuerInvalid creates an issuer invalid error.
func IssuerInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "issuer is invalid: "+reason).
		WithCode(CodeIssuerInvalid)
}

// IssuerAlreadyActive creates an issuer already active error.
func IssuerAlreadyActive(operation string, issuerID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "issuer is already active").
		WithCode(CodeIssuerAlreadyActive).
		WithMeta("issuer_id", issuerID)
}

// TemplateNotFound creates a template not found error.
func TemplateNotFound(operation string, templateID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "template").
		WithCode(CodeTemplateNotFound).
		WithMeta("template_id", templateID)
}

// TemplateArchived creates a template archived error.
func TemplateArchived(operation string, templateID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "template has been archived").
		WithCode(CodeTemplateArchived).
		WithMeta("template_id", templateID)
}

// TemplateInvalid creates a template invalid error.
func TemplateInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "template is invalid: "+reason).
		WithCode(CodeTemplateInvalid)
}
