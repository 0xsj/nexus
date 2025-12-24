package jwt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/presentation"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

// Verifier verifies JWT-encoded Verifiable Presentations.
type Verifier struct {
	resolver   did.Resolver
	vcVerifier vc.Verifier
}

// Ensure Verifier implements presentation.Verifier.
var _ presentation.Verifier = (*Verifier)(nil)

// NewVerifier creates a new JWT presentation verifier.
func NewVerifier(resolver did.Resolver, vcVerifier vc.Verifier) *Verifier {
	return &Verifier{
		resolver:   resolver,
		vcVerifier: vcVerifier,
	}
}

// Verify verifies a presentation and returns the verified presentation.
func (v *Verifier) Verify(ctx context.Context, data []byte) (*presentation.VerifiedPresentation, error) {
	return v.VerifyWithOptions(ctx, data, presentation.DefaultVerificationOptions())
}

// VerifyWithRequest verifies a presentation against a request.
func (v *Verifier) VerifyWithRequest(ctx context.Context, data []byte, request *presentation.Request) (*presentation.VerifiedPresentation, error) {
	opts := presentation.VerificationOptionsFromRequest(request)
	return v.VerifyWithOptions(ctx, data, opts)
}

// VerifyWithOptions verifies a presentation with custom options.
func (v *Verifier) VerifyWithOptions(ctx context.Context, data []byte, opts presentation.VerificationOptions) (*presentation.VerifiedPresentation, error) {
	const op = "jwt.Verifier.VerifyWithOptions"

	// Parse the JWT
	token, err := Parse(string(data))
	if err != nil {
		return nil, err
	}

	// Get holder DID from claims
	holderDID, err := did.Parse(token.Claims.Issuer)
	if err != nil {
		return nil, presentation.ErrInvalidHolder(op, "invalid holder DID: "+err.Error())
	}

	// Check expected holder
	if !opts.ExpectedHolder.IsZero() && !opts.ExpectedHolder.Equals(holderDID) {
		return nil, presentation.ErrInvalidHolder(op, "holder does not match expected")
	}

	// Check nonce/challenge
	if opts.Challenge != "" && token.Claims.Nonce != opts.Challenge {
		return nil, presentation.ErrInvalidChallenge(op)
	}

	// Check timestamps
	now := opts.CurrentTime()
	if token.Claims.NotBefore > 0 && now.Unix() < token.Claims.NotBefore {
		return nil, presentation.ErrInvalidPresentation(op, "presentation not yet valid")
	}

	if token.Claims.ExpiresAt != nil && now.Unix() > *token.Claims.ExpiresAt {
		if !opts.AllowExpiredPresentation {
			return nil, presentation.ErrPresentationExpired(op, token.Claims.JWTID)
		}
	}

	// Check presentation age
	if opts.MaxPresentationAge > 0 {
		issuedAt := time.Unix(token.Claims.IssuedAt, 0)
		age := now.Sub(issuedAt)
		if age > opts.MaxPresentationAge {
			return nil, presentation.ErrPresentationExpired(op, token.Claims.JWTID)
		}
	}

	// Resolve holder's DID document
	doc, err := v.resolver.Resolve(ctx, holderDID)
	if err != nil {
		return nil, presentation.ErrVerificationFailed(op, "failed to resolve holder DID: "+err.Error())
	}

	// Get verification method
	vm := doc.GetVerificationMethod(token.Header.Kid)
	if vm == nil {
		// Try default key
		vm = doc.GetVerificationMethod(holderDID.String() + "#" + holderDID.MethodSpecificID())
	}
	if vm == nil && len(doc.VerificationMethod) > 0 {
		vm = &doc.VerificationMethod[0]
	}
	if vm == nil {
		return nil, presentation.ErrVerificationFailed(op, "no verification method found")
	}

	// Verify signature
	if err := v.verifySignature(token, vm); err != nil {
		return nil, err
	}

	// Build presentation from claims
	pres := buildPresentationFromClaims(holderDID, token.Claims)

	// Verify embedded credentials
	var verifiedCredentials []presentation.VerifiedCredential
	if opts.ShouldVerifyCredentials() && v.vcVerifier != nil {
		verifiedCredentials, err = v.verifyCredentials(ctx, token.Claims.VP.VerifiableCredential, opts)
		if err != nil {
			return nil, err
		}
	}

	// Check expected credential types
	if len(opts.ExpectedCredentialTypes) > 0 {
		if err := v.checkExpectedTypes(verifiedCredentials, opts.ExpectedCredentialTypes); err != nil {
			return nil, err
		}
	}

	return &presentation.VerifiedPresentation{
		Presentation: pres,
		Holder:       holderDID,
		VerifiedAt:   now,
		Challenge:    token.Claims.Nonce,
		Credentials:  verifiedCredentials,
	}, nil
}

// SupportedFormats returns the formats this verifier supports.
func (v *Verifier) SupportedFormats() []vc.Format {
	return []vc.Format{vc.FormatJWT}
}

// ============================================================================
// Parsing
// ============================================================================

// Token represents a parsed JWT presentation.
type Token struct {
	Header    Header
	Claims    Claims
	Signature []byte
	Raw       string
}

// Parse parses a JWT string into a Token.
func Parse(jwt string) (*Token, error) {
	const op = "jwt.Parse"

	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, presentation.ErrInvalidPresentation(op, "invalid JWT format: expected 3 parts")
	}

	// Decode header
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, presentation.ErrInvalidPresentation(op, "invalid JWT header encoding")
	}

	var header Header
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, presentation.ErrInvalidPresentation(op, "invalid JWT header: "+err.Error())
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, presentation.ErrInvalidPresentation(op, "invalid JWT claims encoding")
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, presentation.ErrInvalidPresentation(op, "invalid JWT claims: "+err.Error())
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, presentation.ErrInvalidPresentation(op, "invalid JWT signature encoding")
	}

	return &Token{
		Header:    header,
		Claims:    claims,
		Signature: signature,
		Raw:       jwt,
	}, nil
}

// ParseUnverified parses a JWT without verifying the signature.
func ParseUnverified(jwt string) (*Token, error) {
	return Parse(jwt)
}

// ============================================================================
// Signature Verification
// ============================================================================

// verifySignature verifies the JWT signature.
func (v *Verifier) verifySignature(token *Token, vm *did.VerificationMethod) error {
	const op = "jwt.verifySignature"

	// Get signing input (header.claims)
	parts := strings.Split(token.Raw, ".")
	signingInput := parts[0] + "." + parts[1]

	// Get public key from verification method
	pubKeyBytes, err := extractPublicKey(vm)
	if err != nil {
		return presentation.ErrVerificationFailed(op, "failed to get public key: "+err.Error())
	}

	// Verify based on algorithm
	switch token.Header.Alg {
	case "EdDSA":
		verifier, err := ed25519.NewVerifier(pubKeyBytes)
		if err != nil {
			return presentation.ErrVerificationFailed(op, "failed to create verifier: "+err.Error())
		}
		if err := verifier.Verify([]byte(signingInput), token.Signature); err != nil {
			return presentation.ErrVerificationFailed(op, "EdDSA signature invalid")
		}
		return nil

	case "ES256K":
		// TODO: Implement secp256k1 verification
		return presentation.ErrUnsupportedFormat(op, "ES256K verification not yet implemented")

	default:
		return presentation.ErrUnsupportedFormat(op, "unsupported algorithm: "+token.Header.Alg)
	}
}

// extractPublicKey extracts the public key bytes from a verification method.
func extractPublicKey(vm *did.VerificationMethod) ([]byte, error) {
	// Try multibase first
	if vm.PublicKeyMultibase != "" {
		return decodeMultibase(vm.PublicKeyMultibase)
	}

	// Try base58
	if vm.PublicKeyBase58 != "" {
		return decodeBase58(vm.PublicKeyBase58)
	}

	// Try JWK
	if vm.PublicKeyJwk != nil {
		return decodeJWK(vm.PublicKeyJwk)
	}

	return nil, errNoPublicKey
}

var errNoPublicKey = &keyError{msg: "no public key found in verification method"}

type keyError struct {
	msg string
}

func (e *keyError) Error() string {
	return e.msg
}

// decodeMultibase decodes a multibase-encoded public key.
// Multibase format: <base-prefix><encoded-data>
// For Ed25519 keys: 'z' prefix (base58btc) followed by multicodec + key bytes.
func decodeMultibase(s string) ([]byte, error) {
	if len(s) < 2 {
		return nil, &keyError{msg: "invalid multibase string"}
	}

	prefix := s[0]
	data := s[1:]

	switch prefix {
	case 'z': // base58btc
		decoded, err := decodeBase58(data)
		if err != nil {
			return nil, err
		}
		// Skip multicodec prefix (0xed01 for Ed25519 public key)
		if len(decoded) > 2 && decoded[0] == 0xed && decoded[1] == 0x01 {
			return decoded[2:], nil
		}
		return decoded, nil

	default:
		return nil, &keyError{msg: "unsupported multibase prefix: " + string(prefix)}
	}
}

// decodeBase58 decodes a base58-encoded string.
func decodeBase58(s string) ([]byte, error) {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	result := make([]byte, 0, len(s))

	for i := 0; i < len(s); i++ {
		c := s[i]
		idx := strings.IndexByte(alphabet, c)
		if idx == -1 {
			return nil, &keyError{msg: "invalid base58 character"}
		}

		carry := idx
		for j := 0; j < len(result); j++ {
			carry += int(result[j]) * 58
			result[j] = byte(carry & 0xff)
			carry >>= 8
		}

		for carry > 0 {
			result = append(result, byte(carry&0xff))
			carry >>= 8
		}
	}

	// Handle leading zeros
	for i := 0; i < len(s) && s[i] == '1'; i++ {
		result = append(result, 0)
	}

	// Reverse
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// decodeJWK decodes a public key from JWK format.
func decodeJWK(jwk *did.JWK) ([]byte, error) {
	if jwk.Kty == "OKP" && jwk.Crv == "Ed25519" {
		// Decode x coordinate (base64url)
		return base64.RawURLEncoding.DecodeString(jwk.X)
	}

	return nil, &keyError{msg: "unsupported JWK type: " + jwk.Kty + "/" + jwk.Crv}
}

// ============================================================================
// Credential Verification
// ============================================================================

// verifyCredentials verifies the embedded credentials.
func (v *Verifier) verifyCredentials(ctx context.Context, credentialJWTs []string, opts presentation.VerificationOptions) ([]presentation.VerifiedCredential, error) {
	results := make([]presentation.VerifiedCredential, 0, len(credentialJWTs))

	vcOpts := vc.DefaultVerificationOptions().
		WithAllowExpired(opts.AllowExpiredCredentials)

	if opts.MaxCredentialAge > 0 {
		vcOpts = vcOpts.WithMaxProofAge(opts.MaxCredentialAge)
	}

	for _, jwt := range credentialJWTs {
		result := presentation.VerifiedCredential{}

		cred, err := v.vcVerifier.VerifyWithOptions(ctx, []byte(jwt), vcOpts)
		if err != nil {
			result.Valid = false
			result.Error = err.Error()
		} else {
			result.Credential = cred
			result.Issuer = cred.Issuer.ID
			result.Valid = true

			// Check trusted issuers
			if !opts.IsTrustedIssuer(cred.Issuer.ID) {
				result.Valid = false
				result.Error = "issuer not trusted"
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// checkExpectedTypes verifies all expected credential types are present.
func (v *Verifier) checkExpectedTypes(credentials []presentation.VerifiedCredential, expected []vc.CredentialType) error {
	const op = "jwt.checkExpectedTypes"

	for _, expectedType := range expected {
		found := false
		for _, vc := range credentials {
			if vc.Valid && vc.Credential != nil && vc.Credential.HasType(expectedType) {
				found = true
				break
			}
		}
		if !found {
			return presentation.ErrCredentialMissing(op, string(expectedType))
		}
	}

	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// buildPresentationFromClaims builds a Presentation from JWT claims.
func buildPresentationFromClaims(holder did.DID, claims Claims) *presentation.Presentation {
	types := make([]presentation.PresentationType, len(claims.VP.Type))
	for i, t := range claims.VP.Type {
		types[i] = presentation.PresentationType(t)
	}

	return &presentation.Presentation{
		Context:                 claims.VP.Context,
		ID:                      claims.JWTID,
		Type:                    types,
		Holder:                  holder,
		VerifiableCredentialJWT: claims.VP.VerifiableCredential,
	}
}

// SigningInput returns the signing input (header.claims) from a JWT.
func SigningInput(jwt string) (string, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return "", presentation.ErrInvalidPresentation("SigningInput", "invalid JWT format")
	}
	return parts[0] + "." + parts[1], nil
}
