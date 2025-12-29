package vc

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
)

// Verifier verifies Verifiable Credentials.
type Verifier interface {
	// Verify verifies a credential and returns the parsed credential if valid.
	// The input can be a JWT string or JSON-LD bytes.
	Verify(ctx context.Context, data []byte) (*Credential, error)

	// VerifyWithOptions verifies with custom options.
	VerifyWithOptions(ctx context.Context, data []byte, options VerificationOptions) (*Credential, error)

	// SupportedFormats returns the formats this verifier supports.
	SupportedFormats() []Format
}

// VerificationOptions configures credential verification.
type VerificationOptions struct {
	// ExpectedIssuer requires the credential to be from this issuer.
	ExpectedIssuer *did.DID

	// ExpectedSubject requires the credential subject to be this DID.
	ExpectedSubject *did.DID

	// ExpectedTypes requires the credential to have these types.
	ExpectedTypes []CredentialType

	// ExpectedNonce requires the proof to have this nonce.
	ExpectedNonce string

	// ExpectedDomain requires the proof to have this domain.
	ExpectedDomain string

	// ExpectedChallenge requires the proof to have this challenge.
	ExpectedChallenge string

	// AllowExpired skips expiration checking.
	AllowExpired bool

	// AllowNotYetValid skips issuance date checking.
	AllowNotYetValid bool

	// MaxProofAge rejects proofs older than this duration.
	MaxProofAge *time.Duration

	// CheckRevocation enables revocation checking.
	CheckRevocation bool

	// CurrentTime overrides the current time for validation.
	// Useful for testing.
	CurrentTime *time.Time
}

// DefaultVerificationOptions returns default verification options.
func DefaultVerificationOptions() VerificationOptions {
	return VerificationOptions{
		AllowExpired:     false,
		AllowNotYetValid: false,
		CheckRevocation:  true,
	}
}

// WithExpectedIssuer sets the expected issuer.
func (o VerificationOptions) WithExpectedIssuer(issuer did.DID) VerificationOptions {
	o.ExpectedIssuer = &issuer
	return o
}

// WithExpectedSubject sets the expected subject.
func (o VerificationOptions) WithExpectedSubject(subject did.DID) VerificationOptions {
	o.ExpectedSubject = &subject
	return o
}

// WithExpectedTypes sets the expected types.
func (o VerificationOptions) WithExpectedTypes(types ...CredentialType) VerificationOptions {
	o.ExpectedTypes = types
	return o
}

// WithExpectedNonce sets the expected nonce.
func (o VerificationOptions) WithExpectedNonce(nonce string) VerificationOptions {
	o.ExpectedNonce = nonce
	return o
}

// WithExpectedDomain sets the expected domain.
func (o VerificationOptions) WithExpectedDomain(domain string) VerificationOptions {
	o.ExpectedDomain = domain
	return o
}

// WithExpectedChallenge sets the expected challenge.
func (o VerificationOptions) WithExpectedChallenge(challenge string) VerificationOptions {
	o.ExpectedChallenge = challenge
	return o
}

// WithAllowExpired allows expired credentials.
func (o VerificationOptions) WithAllowExpired(allow bool) VerificationOptions {
	o.AllowExpired = allow
	return o
}

// WithAllowNotYetValid allows not-yet-valid credentials.
func (o VerificationOptions) WithAllowNotYetValid(allow bool) VerificationOptions {
	o.AllowNotYetValid = allow
	return o
}

// WithMaxProofAge sets the maximum proof age.
func (o VerificationOptions) WithMaxProofAge(maxAge time.Duration) VerificationOptions {
	o.MaxProofAge = &maxAge
	return o
}

// WithCheckRevocation enables/disables revocation checking.
func (o VerificationOptions) WithCheckRevocation(check bool) VerificationOptions {
	o.CheckRevocation = check
	return o
}

// WithCurrentTime sets the current time for validation.
func (o VerificationOptions) WithCurrentTime(t time.Time) VerificationOptions {
	o.CurrentTime = &t
	return o
}

// Now returns the current time to use for validation.
func (o VerificationOptions) Now() time.Time {
	if o.CurrentTime != nil {
		return *o.CurrentTime
	}
	return time.Now()
}

// VerificationResult contains the result of credential verification.
type VerificationResult struct {
	// Valid indicates if the credential is valid.
	Valid bool

	// Credential is the parsed credential (if valid).
	Credential *Credential

	// Errors contains any validation errors.
	Errors []error

	// Warnings contains non-fatal issues.
	Warnings []string

	// Checks contains the results of individual checks.
	Checks []VerificationCheck
}

// VerificationCheck represents the result of a single verification check.
type VerificationCheck struct {
	// Name is the check name.
	Name string

	// Passed indicates if the check passed.
	Passed bool

	// Message provides details about the check.
	Message string
}

// NewVerificationResult creates a new verification result.
func NewVerificationResult() *VerificationResult {
	return &VerificationResult{
		Valid:    true,
		Errors:   make([]error, 0),
		Warnings: make([]string, 0),
		Checks:   make([]VerificationCheck, 0),
	}
}

// AddError adds an error and marks the result as invalid.
func (r *VerificationResult) AddError(err error) {
	r.Valid = false
	r.Errors = append(r.Errors, err)
}

// AddWarning adds a warning.
func (r *VerificationResult) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

// AddCheck adds a verification check result.
func (r *VerificationResult) AddCheck(name string, passed bool, message string) {
	r.Checks = append(r.Checks, VerificationCheck{
		Name:    name,
		Passed:  passed,
		Message: message,
	})
	if !passed {
		r.Valid = false
	}
}

// ============================================================================
// Revocation Checker
// ============================================================================

// RevocationChecker checks credential revocation status.
type RevocationChecker interface {
	// IsRevoked checks if a credential is revoked.
	IsRevoked(ctx context.Context, credential *Credential) (bool, error)
}

// ============================================================================
// Multi-Verifier
// ============================================================================

// MultiVerifier can verify credentials in multiple formats.
type MultiVerifier struct {
	verifiers map[Format]Verifier
	resolver  did.Resolver
}

// NewMultiVerifier creates a new multi-format verifier.
func NewMultiVerifier(resolver did.Resolver) *MultiVerifier {
	return &MultiVerifier{
		verifiers: make(map[Format]Verifier),
		resolver:  resolver,
	}
}

// Register registers a verifier for a format.
func (m *MultiVerifier) Register(format Format, verifier Verifier) {
	m.verifiers[format] = verifier
}

// Verify verifies a credential, auto-detecting the format.
func (m *MultiVerifier) Verify(ctx context.Context, data []byte) (*Credential, error) {
	format := DetectFormat(data)
	return m.VerifyWithFormat(ctx, data, format)
}

// VerifyWithFormat verifies a credential in the specified format.
func (m *MultiVerifier) VerifyWithFormat(ctx context.Context, data []byte, format Format) (*Credential, error) {
	const op = "MultiVerifier.VerifyWithFormat"

	verifier, ok := m.verifiers[format]
	if !ok {
		return nil, ErrUnsupportedFormat(op, format.String())
	}

	return verifier.Verify(ctx, data)
}

// SupportedFormats returns all registered formats.
func (m *MultiVerifier) SupportedFormats() []Format {
	formats := make([]Format, 0, len(m.verifiers))
	for f := range m.verifiers {
		formats = append(formats, f)
	}
	return formats
}

// ============================================================================
// Format Detection
// ============================================================================

// DetectFormat detects the credential format from the data.
func DetectFormat(data []byte) Format {
	if len(data) == 0 {
		return FormatJWT
	}

	// JWT starts with "ey" (base64 of {"alg"...)
	if len(data) > 2 && data[0] == 'e' && data[1] == 'y' {
		return FormatJWT
	}

	// JSON-LD starts with "{"
	if data[0] == '{' {
		return FormatJSONLD
	}

	// Default to JWT
	return FormatJWT
}
