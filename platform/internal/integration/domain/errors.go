// Package domain contains the core business logic for the Integration bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Integration-specific)
// ============================================================================

const (
	CodeIntegrationNotFound      pkgerrors.Code = "INTEGRATION_NOT_FOUND"
	CodeIntegrationDisconnected  pkgerrors.Code = "INTEGRATION_DISCONNECTED"
	CodeIntegrationSuspended     pkgerrors.Code = "INTEGRATION_SUSPENDED"
	CodeIntegrationInvalid       pkgerrors.Code = "INTEGRATION_INVALID"
	CodeProviderNotSupported     pkgerrors.Code = "PROVIDER_NOT_SUPPORTED"
	CodeIntegrationAlreadyExists pkgerrors.Code = "INTEGRATION_ALREADY_EXISTS"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrIntegrationNotFound      = errors.New("integration not found")
	ErrIntegrationDisconnected  = errors.New("integration is disconnected")
	ErrIntegrationSuspended     = errors.New("integration is suspended")
	ErrIntegrationInvalid       = errors.New("integration is invalid")
	ErrProviderNotSupported     = errors.New("provider type is not supported")
	ErrIntegrationAlreadyExists = errors.New("integration already exists")
)

// ============================================================================
// Error Constructors
// ============================================================================

// IntegrationNotFound creates an integration not found error.
func IntegrationNotFound(operation string, integrationID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "integration").
		WithCode(CodeIntegrationNotFound).
		WithMeta("integration_id", integrationID)
}

// IntegrationDisconnected creates an integration disconnected error.
func IntegrationDisconnected(operation string, integrationID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "integration is disconnected").
		WithCode(CodeIntegrationDisconnected).
		WithMeta("integration_id", integrationID)
}

// IntegrationSuspendedErr creates an integration suspended error.
func IntegrationSuspendedErr(operation string, integrationID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "integration is suspended").
		WithCode(CodeIntegrationSuspended).
		WithMeta("integration_id", integrationID)
}

// IntegrationInvalid creates an integration invalid error.
func IntegrationInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "integration is invalid: "+reason).
		WithCode(CodeIntegrationInvalid)
}

// ProviderNotSupported creates a provider not supported error.
func ProviderNotSupported(operation string, providerType string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "provider type not supported: "+providerType).
		WithCode(CodeProviderNotSupported).
		WithMeta("provider_type", providerType)
}

// IntegrationAlreadyExistsErr creates an integration already exists error.
func IntegrationAlreadyExistsErr(operation string, userID, providerType string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "integration").
		WithCode(CodeIntegrationAlreadyExists).
		WithMeta("user_id", userID).
		WithMeta("provider_type", providerType)
}
