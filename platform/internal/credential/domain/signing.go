package domain

import (
	"context"
	"time"
)

// ============================================================================
// Signing Service Interface
// ============================================================================

// SigningService defines the interface for signing credentials.
type SigningService interface {
	// SignCredential signs a credential and returns the JWT-VC string.
	SignCredential(ctx context.Context, params SigningParams) (string, error)
}

// ============================================================================
// Signing Parameters
// ============================================================================

// SigningParams contains the parameters needed to sign a credential.
type SigningParams struct {
	// CredentialID is the unique identifier for the credential.
	CredentialID string

	// CredentialType is the type of credential (e.g., "VerifiedEmail").
	CredentialType string

	// IssuerDID is the DID of the issuer.
	IssuerDID string

	// HolderDID is the DID of the holder/subject.
	HolderDID string

	// Claims are the credential claims.
	Claims map[string]any

	// IssuedAt is when the credential was issued.
	IssuedAt time.Time

	// ExpiresAt is when the credential expires (optional).
	ExpiresAt *time.Time

	// SchemaID is the schema identifier (optional).
	SchemaID string
}

// ============================================================================
// Signing Result
// ============================================================================

// SigningResult contains the result of signing a credential.
type SigningResult struct {
	// JWT is the signed JWT-VC string.
	JWT string

	// SignedAt is when the credential was signed.
	SignedAt time.Time
}
