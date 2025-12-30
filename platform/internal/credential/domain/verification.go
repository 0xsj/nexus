package domain

import (
	"context"
	"time"
)

// ============================================================================
// Verification Service Interface
// ============================================================================

// VerificationService defines the interface for verifying credentials.
type VerificationService interface {
	// VerifyCredential verifies a JWT-VC and returns the verification result.
	VerifyCredential(ctx context.Context, jwt string) (*VerificationResult, error)
}

// ============================================================================
// Verification Result
// ============================================================================

// VerificationResult contains the result of credential verification.
type VerificationResult struct {
	// Valid indicates if the credential passed all verification checks.
	Valid bool

	// Issuer is the DID of the credential issuer.
	Issuer string

	// Holder is the DID of the credential holder/subject.
	Holder string

	// CredentialID is the unique identifier of the credential.
	CredentialID string

	// CredentialTypes are the types of the credential.
	CredentialTypes []string

	// Claims are the credential claims.
	Claims map[string]any

	// IssuedAt is when the credential was issued.
	IssuedAt *time.Time

	// ExpiresAt is when the credential expires (nil if no expiration).
	ExpiresAt *time.Time

	// Checks contains the result of individual verification checks.
	Checks VerificationChecks

	// Error contains the error message if verification failed.
	Error string
}

// ============================================================================
// Verification Checks
// ============================================================================

// VerificationChecks contains the result of individual verification checks.
type VerificationChecks struct {
	// Signature indicates if the cryptographic signature is valid.
	Signature CheckResult

	// Expiration indicates if the credential is not expired.
	Expiration CheckResult

	// NotBefore indicates if the credential is currently valid (past nbf).
	NotBefore CheckResult

	// IssuerDID indicates if the issuer DID is valid and resolvable.
	IssuerDID CheckResult

	// HolderDID indicates if the holder DID is valid.
	HolderDID CheckResult
}

// AllPassed returns true if all checks passed.
func (c VerificationChecks) AllPassed() bool {
	return c.Signature == CheckPassed &&
		c.Expiration == CheckPassed &&
		c.NotBefore == CheckPassed &&
		c.IssuerDID == CheckPassed &&
		c.HolderDID == CheckPassed
}

// ============================================================================
// Check Result
// ============================================================================

// CheckResult represents the result of a single verification check.
type CheckResult string

const (
	// CheckPassed indicates the check passed.
	CheckPassed CheckResult = "passed"

	// CheckFailed indicates the check failed.
	CheckFailed CheckResult = "failed"

	// CheckSkipped indicates the check was skipped.
	CheckSkipped CheckResult = "skipped"

	// CheckError indicates an error occurred during the check.
	CheckError CheckResult = "error"
)

// String returns the string representation of the check result.
func (r CheckResult) String() string {
	return string(r)
}

// IsPassed returns true if the check passed.
func (r CheckResult) IsPassed() bool {
	return r == CheckPassed
}

// ============================================================================
// Verification Options
// ============================================================================

// VerificationOptions configures verification behavior.
type VerificationOptions struct {
	// SkipExpirationCheck skips the expiration check.
	SkipExpirationCheck bool

	// SkipNotBeforeCheck skips the not-before check.
	SkipNotBeforeCheck bool

	// CheckRevocation checks if the credential is revoked in our system.
	CheckRevocation bool

	// ExpectedIssuer requires the issuer to match this DID.
	ExpectedIssuer string

	// ExpectedHolder requires the holder to match this DID.
	ExpectedHolder string

	// ExpectedTypes requires the credential to have these types.
	ExpectedTypes []string
}

// DefaultVerificationOptions returns default verification options.
func DefaultVerificationOptions() VerificationOptions {
	return VerificationOptions{
		SkipExpirationCheck: false,
		SkipNotBeforeCheck:  false,
		CheckRevocation:     false,
	}
}
