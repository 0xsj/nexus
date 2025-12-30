package verification

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/vc"
	vcjwt "github.com/0xsj/nexus/platform/pkg/vc/jwt"
)

// ============================================================================
// Verifier
// ============================================================================

// Verifier implements domain.VerificationService using pkg/vc.
type Verifier struct {
	resolver did.Resolver
	verifier *vcjwt.Verifier
}

// NewVerifier creates a new Verifier.
func NewVerifier() *Verifier {
	resolver := key.NewResolver()
	return &Verifier{
		resolver: resolver,
		verifier: vcjwt.NewVerifier(resolver),
	}
}

// ============================================================================
// VerificationService Implementation
// ============================================================================

// VerifyCredential verifies a JWT-VC and returns the verification result.
func (v *Verifier) VerifyCredential(ctx context.Context, jwt string) (*domain.VerificationResult, error) {
	result := &domain.VerificationResult{
		Valid: false,
		Checks: domain.VerificationChecks{
			Signature:  domain.CheckSkipped,
			Expiration: domain.CheckSkipped,
			NotBefore:  domain.CheckSkipped,
			IssuerDID:  domain.CheckSkipped,
			HolderDID:  domain.CheckSkipped,
		},
	}

	// Step 1: Parse the JWT without verification to extract claims
	token, err := vcjwt.ParseUnverified(jwt)
	if err != nil {
		result.Error = fmt.Sprintf("failed to parse JWT: %v", err)
		return result, nil
	}

	// Extract credential info from token
	result.CredentialID = token.Claims.JWTID
	result.Issuer = token.Claims.Issuer
	result.Holder = token.Claims.Subject
	result.CredentialTypes = token.Claims.VC.Type
	result.Claims = extractClaims(token.Claims.VC.CredentialSubject)

	issuedAt := vcjwt.TimestampToTime(token.Claims.IssuedAt)
	result.IssuedAt = &issuedAt

	if token.Claims.ExpiresAt != nil {
		expiresAt := vcjwt.TimestampToTime(*token.Claims.ExpiresAt)
		result.ExpiresAt = &expiresAt
	}

	// Step 2: Validate issuer DID format
	issuerDID, err := did.Parse(result.Issuer)
	if err != nil {
		result.Checks.IssuerDID = domain.CheckFailed
		result.Error = fmt.Sprintf("invalid issuer DID: %v", err)
		return result, nil
	}
	result.Checks.IssuerDID = domain.CheckPassed

	// Step 3: Validate holder DID format
	_, err = did.Parse(result.Holder)
	if err != nil {
		result.Checks.HolderDID = domain.CheckFailed
		result.Error = fmt.Sprintf("invalid holder DID: %v", err)
		return result, nil
	}
	result.Checks.HolderDID = domain.CheckPassed

	// Step 4: Check not-before (nbf)
	now := time.Now()
	nbf := vcjwt.TimestampToTime(token.Claims.NotBefore)
	if now.Before(nbf) {
		result.Checks.NotBefore = domain.CheckFailed
		result.Error = "credential is not yet valid"
		return result, nil
	}
	result.Checks.NotBefore = domain.CheckPassed

	// Step 5: Check expiration
	if token.Claims.ExpiresAt != nil {
		exp := vcjwt.TimestampToTime(*token.Claims.ExpiresAt)
		if now.After(exp) {
			result.Checks.Expiration = domain.CheckFailed
			result.Error = "credential has expired"
			return result, nil
		}
	}
	result.Checks.Expiration = domain.CheckPassed

	// Step 6: Verify signature (requires DID resolution)
	// Only did:key is supported for now
	if issuerDID.Method() != did.MethodKey {
		result.Checks.Signature = domain.CheckFailed
		result.Error = fmt.Sprintf("unsupported DID method for verification: %s", issuerDID.Method())
		return result, nil
	}

	// Use the full verifier which handles signature verification
	_, err = v.verifier.Verify(ctx, []byte(jwt))
	if err != nil {
		result.Checks.Signature = domain.CheckFailed
		result.Error = fmt.Sprintf("signature verification failed: %v", err)
		return result, nil
	}
	result.Checks.Signature = domain.CheckPassed

	// All checks passed
	result.Valid = true
	return result, nil
}

// ============================================================================
// Helpers
// ============================================================================

// extractClaims extracts claims from the credential subject.
func extractClaims(subject vc.Subject) map[string]any {
	if subject.Claims == nil {
		return make(map[string]any)
	}
	return subject.Claims
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.VerificationService = (*Verifier)(nil)
