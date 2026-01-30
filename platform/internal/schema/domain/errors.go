// Package domain contains the core business logic for the Schema bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Schema-specific)
// ============================================================================

const (
	// Schema errors
	CodeSchemaNotFound      pkgerrors.Code = "SCHEMA_NOT_FOUND"
	CodeSchemaAlreadyExists pkgerrors.Code = "SCHEMA_ALREADY_EXISTS"
	CodeSchemaDeprecated    pkgerrors.Code = "SCHEMA_DEPRECATED"
	CodeSchemaInvalid       pkgerrors.Code = "SCHEMA_INVALID"

	// Version errors
	CodeVersionNotFound      pkgerrors.Code = "VERSION_NOT_FOUND"
	CodeVersionAlreadyExists pkgerrors.Code = "VERSION_ALREADY_EXISTS"
	CodeInvalidVersionFormat pkgerrors.Code = "INVALID_VERSION_FORMAT"

	// Claim errors
	CodeClaimNotFound          pkgerrors.Code = "CLAIM_NOT_FOUND"
	CodeClaimAlreadyExists     pkgerrors.Code = "CLAIM_ALREADY_EXISTS"
	CodeInvalidClaimDefinition pkgerrors.Code = "INVALID_CLAIM_DEFINITION"

	// Validation errors
	CodeInvalidDataType       pkgerrors.Code = "INVALID_DATA_TYPE"
	CodeClaimValidationFailed pkgerrors.Code = "CLAIM_VALIDATION_FAILED"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	// Schema errors
	ErrSchemaNotFound      = errors.New("schema not found")
	ErrSchemaAlreadyExists = errors.New("schema already exists")
	ErrSchemaDeprecated    = errors.New("schema is deprecated")
	ErrSchemaInvalid       = errors.New("schema is invalid")

	// Version errors
	ErrVersionNotFound      = errors.New("schema version not found")
	ErrVersionAlreadyExists = errors.New("schema version already exists")
	ErrInvalidVersionFormat = errors.New("invalid version format")

	// Claim errors
	ErrClaimNotFound          = errors.New("claim not found")
	ErrClaimAlreadyExists     = errors.New("claim already exists")
	ErrInvalidClaimDefinition = errors.New("invalid claim definition")

	// Validation errors
	ErrInvalidDataType       = errors.New("invalid data type")
	ErrClaimValidationFailed = errors.New("claim validation failed")
)

// ============================================================================
// Error Constructors
// ============================================================================

// SchemaNotFound creates a schema not found error.
func SchemaNotFound(operation string, schemaID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "schema").
		WithCode(CodeSchemaNotFound).
		WithMeta("schema_id", schemaID)
}

// SchemaNotFoundByType creates a schema not found error when looking up by type.
func SchemaNotFoundByType(operation string, schemaType string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "schema").
		WithCode(CodeSchemaNotFound).
		WithMeta("schema_type", schemaType)
}

// SchemaAlreadyExists creates a schema already exists error.
func SchemaAlreadyExists(operation string, schemaType string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "schema").
		WithCode(CodeSchemaAlreadyExists).
		WithMeta("schema_type", schemaType)
}

// SchemaDeprecated creates a schema deprecated error.
func SchemaDeprecated(operation string, schemaID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "schema is deprecated and cannot be used for new credentials").
		WithCode(CodeSchemaDeprecated).
		WithMeta("schema_id", schemaID)
}

// SchemaInvalid creates a schema invalid error.
func SchemaInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid schema: "+reason).
		WithCode(CodeSchemaInvalid)
}

// VersionNotFound creates a version not found error.
func VersionNotFound(operation string, schemaID string, version string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "schema version").
		WithCode(CodeVersionNotFound).
		WithMeta("schema_id", schemaID).
		WithMeta("version", version)
}

// VersionAlreadyExists creates a version already exists error.
func VersionAlreadyExists(operation string, schemaID string, version string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "schema version").
		WithCode(CodeVersionAlreadyExists).
		WithMeta("schema_id", schemaID).
		WithMeta("version", version)
}

// InvalidVersionFormat creates an invalid version format error.
func InvalidVersionFormat(operation string, version string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid version format: "+reason).
		WithCode(CodeInvalidVersionFormat).
		WithMeta("version", version)
}

// ClaimNotFound creates a claim not found error.
func ClaimNotFound(operation string, schemaID string, claimKey string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "claim").
		WithCode(CodeClaimNotFound).
		WithMeta("schema_id", schemaID).
		WithMeta("claim_key", claimKey)
}

// ClaimAlreadyExists creates a claim already exists error.
func ClaimAlreadyExists(operation string, schemaID string, claimKey string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "claim").
		WithCode(CodeClaimAlreadyExists).
		WithMeta("schema_id", schemaID).
		WithMeta("claim_key", claimKey)
}

// InvalidClaimDefinition creates an invalid claim definition error.
func InvalidClaimDefinition(operation string, claimKey string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid claim definition: "+reason).
		WithCode(CodeInvalidClaimDefinition).
		WithMeta("claim_key", claimKey)
}

// InvalidDataType creates an invalid data type error.
func InvalidDataType(operation string, dataType string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid data type: "+dataType).
		WithCode(CodeInvalidDataType).
		WithMeta("data_type", dataType)
}

// ClaimValidationFailed creates a claim validation failed error.
func ClaimValidationFailed(operation string, claimKey string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "claim validation failed: "+reason).
		WithCode(CodeClaimValidationFailed).
		WithMeta("claim_key", claimKey)
}
