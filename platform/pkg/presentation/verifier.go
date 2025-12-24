package presentation

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

// Verifier verifies Verifiable Presentations.
type Verifier interface {
	// Verify verifies a presentation and returns the verified presentation.
	Verify(ctx context.Context, data []byte) (*VerifiedPresentation, error)

	// VerifyWithRequest verifies a presentation against a request.
	VerifyWithRequest(ctx context.Context, data []byte, request *Request) (*VerifiedPresentation, error)

	// VerifyWithOptions verifies a presentation with custom options.
	VerifyWithOptions(ctx context.Context, data []byte, opts VerificationOptions) (*VerifiedPresentation, error)

	// SupportedFormats returns the formats this verifier supports.
	SupportedFormats() []vc.Format
}

// VerifiedPresentation is the result of successful verification.
type VerifiedPresentation struct {
	// Presentation is the verified presentation.
	Presentation *Presentation

	// Holder is the verified holder DID.
	Holder did.DID

	// VerifiedAt is when the verification was performed.
	VerifiedAt time.Time

	// Challenge is the challenge that was verified.
	Challenge string

	// Domain is the domain that was verified.
	Domain string

	// Credentials contains verification results for each credential.
	Credentials []VerifiedCredential
}

// VerifiedCredential represents verification result for a single credential.
type VerifiedCredential struct {
	// Credential is the verified credential.
	Credential *vc.Credential

	// Issuer is the verified issuer DID.
	Issuer did.DID

	// Valid indicates if the credential passed verification.
	Valid bool

	// Error contains any verification error.
	Error string
}

// ============================================================================
// Verification Options
// ============================================================================

// VerificationOptions configures presentation verification.
type VerificationOptions struct {
	// Challenge is the expected challenge value.
	Challenge string

	// Domain is the expected domain value.
	Domain string

	// ExpectedHolder is the expected holder DID.
	ExpectedHolder did.DID

	// ExpectedCredentialTypes are the credential types that must be present.
	ExpectedCredentialTypes []vc.CredentialType

	// TrustedIssuers limits which issuers are accepted.
	// Empty means any issuer is accepted.
	TrustedIssuers []did.DID

	// AllowExpiredCredentials allows expired credentials.
	AllowExpiredCredentials bool

	// AllowExpiredPresentation allows expired presentations.
	AllowExpiredPresentation bool

	// MaxPresentationAge is the maximum age of the presentation.
	MaxPresentationAge time.Duration

	// MaxCredentialAge is the maximum age of credentials.
	MaxCredentialAge time.Duration

	// VerifyCredentials controls whether to verify each credential.
	// Default is true.
	VerifyCredentials *bool

	// CheckRevocation controls whether to check credential revocation.
	// Default is true.
	CheckRevocation *bool

	// Now overrides the current time for verification.
	// Useful for testing.
	Now time.Time
}

// DefaultVerificationOptions returns default verification options.
func DefaultVerificationOptions() VerificationOptions {
	verifyCredentials := true
	checkRevocation := true

	return VerificationOptions{
		VerifyCredentials: &verifyCredentials,
		CheckRevocation:   &checkRevocation,
	}
}

// WithChallenge sets the expected challenge.
func (o VerificationOptions) WithChallenge(challenge string) VerificationOptions {
	o.Challenge = challenge
	return o
}

// WithDomain sets the expected domain.
func (o VerificationOptions) WithDomain(domain string) VerificationOptions {
	o.Domain = domain
	return o
}

// WithExpectedHolder sets the expected holder.
func (o VerificationOptions) WithExpectedHolder(holder did.DID) VerificationOptions {
	o.ExpectedHolder = holder
	return o
}

// WithExpectedCredentialTypes sets expected credential types.
func (o VerificationOptions) WithExpectedCredentialTypes(types ...vc.CredentialType) VerificationOptions {
	o.ExpectedCredentialTypes = types
	return o
}

// WithTrustedIssuers sets trusted issuers.
func (o VerificationOptions) WithTrustedIssuers(issuers ...did.DID) VerificationOptions {
	o.TrustedIssuers = issuers
	return o
}

// WithAllowExpiredCredentials allows expired credentials.
func (o VerificationOptions) WithAllowExpiredCredentials(allow bool) VerificationOptions {
	o.AllowExpiredCredentials = allow
	return o
}

// WithAllowExpiredPresentation allows expired presentations.
func (o VerificationOptions) WithAllowExpiredPresentation(allow bool) VerificationOptions {
	o.AllowExpiredPresentation = allow
	return o
}

// WithMaxPresentationAge sets maximum presentation age.
func (o VerificationOptions) WithMaxPresentationAge(maxAge time.Duration) VerificationOptions {
	o.MaxPresentationAge = maxAge
	return o
}

// WithMaxCredentialAge sets maximum credential age.
func (o VerificationOptions) WithMaxCredentialAge(maxAge time.Duration) VerificationOptions {
	o.MaxCredentialAge = maxAge
	return o
}

// WithVerifyCredentials controls credential verification.
func (o VerificationOptions) WithVerifyCredentials(verify bool) VerificationOptions {
	o.VerifyCredentials = &verify
	return o
}

// WithCheckRevocation controls revocation checking.
func (o VerificationOptions) WithCheckRevocation(check bool) VerificationOptions {
	o.CheckRevocation = &check
	return o
}

// WithNow overrides the current time.
func (o VerificationOptions) WithNow(now time.Time) VerificationOptions {
	o.Now = now
	return o
}

// OptionsFromRequest creates verification options from a presentation request.
func VerificationOptionsFromRequest(r *Request) VerificationOptions {
	opts := DefaultVerificationOptions().
		WithChallenge(r.Challenge).
		WithDomain(r.Domain)

	// Add expected credential types
	for _, cr := range r.RequestedCredentials {
		if cr.Required {
			opts.ExpectedCredentialTypes = append(opts.ExpectedCredentialTypes, cr.Type)
		}
	}

	return opts
}

// ============================================================================
// Verification Helpers
// ============================================================================

// ShouldVerifyCredentials returns whether credentials should be verified.
func (o VerificationOptions) ShouldVerifyCredentials() bool {
	if o.VerifyCredentials == nil {
		return true
	}
	return *o.VerifyCredentials
}

// ShouldCheckRevocation returns whether revocation should be checked.
func (o VerificationOptions) ShouldCheckRevocation() bool {
	if o.CheckRevocation == nil {
		return true
	}
	return *o.CheckRevocation
}

// CurrentTime returns the time to use for verification.
func (o VerificationOptions) CurrentTime() time.Time {
	if o.Now.IsZero() {
		return time.Now()
	}
	return o.Now
}

// IsTrustedIssuer checks if an issuer is in the trusted list.
func (o VerificationOptions) IsTrustedIssuer(issuer did.DID) bool {
	// Empty list means all issuers are trusted
	if len(o.TrustedIssuers) == 0 {
		return true
	}

	for _, trusted := range o.TrustedIssuers {
		if trusted.Equals(issuer) {
			return true
		}
	}
	return false
}

// ============================================================================
// VerifiedPresentation Helpers
// ============================================================================

// AllCredentialsValid returns true if all credentials passed verification.
func (vp *VerifiedPresentation) AllCredentialsValid() bool {
	for _, vc := range vp.Credentials {
		if !vc.Valid {
			return false
		}
	}
	return true
}

// InvalidCredentials returns credentials that failed verification.
func (vp *VerifiedPresentation) InvalidCredentials() []VerifiedCredential {
	var invalid []VerifiedCredential
	for _, vc := range vp.Credentials {
		if !vc.Valid {
			invalid = append(invalid, vc)
		}
	}
	return invalid
}

// GetCredentialByType finds a verified credential by type.
func (vp *VerifiedPresentation) GetCredentialByType(credType vc.CredentialType) *VerifiedCredential {
	for i := range vp.Credentials {
		if vp.Credentials[i].Credential != nil && vp.Credentials[i].Credential.HasType(credType) {
			return &vp.Credentials[i]
		}
	}
	return nil
}

// HasCredentialType returns true if the presentation contains the credential type.
func (vp *VerifiedPresentation) HasCredentialType(credType vc.CredentialType) bool {
	return vp.GetCredentialByType(credType) != nil
}

// CredentialCount returns the number of credentials in the presentation.
func (vp *VerifiedPresentation) CredentialCount() int {
	return len(vp.Credentials)
}
