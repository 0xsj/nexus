package jwt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

// Verifier verifies JWT-encoded Verifiable Credentials.
type Verifier struct {
	resolver did.Resolver
}

// Ensure Verifier implements vc.Verifier.
var _ vc.Verifier = (*Verifier)(nil)

// NewVerifier creates a new JWT-VC verifier.
func NewVerifier(resolver did.Resolver) *Verifier {
	return &Verifier{
		resolver: resolver,
	}
}

// Verify verifies a JWT-VC and returns the parsed credential.
func (v *Verifier) Verify(ctx context.Context, data []byte) (*vc.Credential, error) {
	return v.VerifyWithOptions(ctx, data, vc.DefaultVerificationOptions())
}

// VerifyWithOptions verifies a JWT-VC with custom options.
func (v *Verifier) VerifyWithOptions(ctx context.Context, data []byte, options vc.VerificationOptions) (*vc.Credential, error) {
	const op = "jwt.Verifier.VerifyWithOptions"

	// Parse the JWT
	token, err := Parse(string(data))
	if err != nil {
		return nil, err
	}

	// Resolve issuer DID to get public key
	issuerDID, err := did.Parse(token.Claims.Issuer)
	if err != nil {
		return nil, vc.ErrInvalidIssuer(op, "invalid issuer DID: "+err.Error())
	}

	doc, err := v.resolver.Resolve(ctx, issuerDID)
	if err != nil {
		return nil, vc.ErrVerificationFailed(op, "failed to resolve issuer DID: "+err.Error())
	}

	// Find the verification method
	vm := findVerificationMethod(doc, token.Header.Kid)
	if vm == nil {
		return nil, vc.ErrVerificationFailed(op, "verification method not found: "+token.Header.Kid)
	}

	// Get the public key
	publicKey, algorithm, err := extractPublicKey(vm)
	if err != nil {
		return nil, vc.ErrVerificationFailed(op, "failed to extract public key: "+err.Error())
	}

	// Verify the signature
	if err := verifySignature(token, publicKey, algorithm); err != nil {
		return nil, err
	}

	// Convert to credential
	credential := tokenToCredential(token)

	// Validate temporal claims
	if err := validateTemporalClaims(token.Claims, options); err != nil {
		return nil, err
	}

	// Validate expected values
	if err := validateExpectedValues(credential, options); err != nil {
		return nil, err
	}

	return credential, nil
}

// SupportedFormats returns the formats this verifier supports.
func (v *Verifier) SupportedFormats() []vc.Format {
	return []vc.Format{vc.FormatJWT}
}

// ============================================================================
// JWT Parsing
// ============================================================================

// Token represents a parsed JWT.
type Token struct {
	// Raw is the original JWT string.
	Raw string

	// Header is the decoded header.
	Header Header

	// Claims is the decoded claims.
	Claims Claims

	// Signature is the raw signature bytes.
	Signature []byte

	// SigningInput is the header.claims portion.
	SigningInput string
}

// Parse parses a JWT string into a Token.
func Parse(jwtStr string) (*Token, error) {
	const op = "jwt.Parse"

	parts := strings.Split(jwtStr, ".")
	if len(parts) != 3 {
		return nil, vc.ErrInvalidCredential(op, "invalid JWT format: expected 3 parts")
	}

	// Decode header
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, vc.ErrInvalidCredential(op, "invalid JWT header encoding: "+err.Error())
	}

	var header Header
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, vc.ErrInvalidCredential(op, "invalid JWT header: "+err.Error())
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, vc.ErrInvalidCredential(op, "invalid JWT claims encoding: "+err.Error())
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, vc.ErrInvalidCredential(op, "invalid JWT claims: "+err.Error())
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, vc.ErrInvalidCredential(op, "invalid JWT signature encoding: "+err.Error())
	}

	return &Token{
		Raw:          jwtStr,
		Header:       header,
		Claims:       claims,
		Signature:    signature,
		SigningInput: parts[0] + "." + parts[1],
	}, nil
}

// ParseUnverified parses a JWT without verifying the signature.
// Useful for inspecting credentials before verification.
func ParseUnverified(jwtStr string) (*Token, error) {
	return Parse(jwtStr)
}

// ============================================================================
// Verification Helpers
// ============================================================================

// findVerificationMethod finds a verification method by ID in the DID document.
func findVerificationMethod(doc *did.Document, kid string) *did.VerificationMethod {
	// Try exact match first
	vm := doc.GetVerificationMethod(kid)
	if vm != nil {
		return vm
	}

	// If kid is just the fragment, try with full DID
	if !strings.HasPrefix(kid, "did:") {
		fullID := doc.ID.Fragment(kid)
		vm = doc.GetVerificationMethod(fullID)
		if vm != nil {
			return vm
		}
	}

	// Fall back to primary verification method
	return doc.PrimaryVerificationMethod()
}

// extractPublicKey extracts the public key from a verification method.
func extractPublicKey(vm *did.VerificationMethod) ([]byte, crypto.Algorithm, error) {
	const op = "jwt.extractPublicKey"

	switch vm.Type {
	case did.Ed25519VerificationKey2020:
		if vm.PublicKeyMultibase != "" {
			return decodeMultibaseKey(vm.PublicKeyMultibase)
		}
		if vm.PublicKeyBase58 != "" {
			return decodeBase58Key(vm.PublicKeyBase58, crypto.AlgorithmEd25519)
		}

	case did.EcdsaSecp256k1VerificationKey2019:
		if vm.PublicKeyMultibase != "" {
			return decodeMultibaseKey(vm.PublicKeyMultibase)
		}
		if vm.PublicKeyBase58 != "" {
			return decodeBase58Key(vm.PublicKeyBase58, crypto.AlgorithmSecp256k1)
		}

	case did.JsonWebKey2020:
		if vm.PublicKeyJwk != nil {
			return decodeJWK(vm.PublicKeyJwk)
		}
	}

	return nil, "", vc.ErrVerificationFailed(op, "unable to extract public key from verification method")
}

// decodeMultibaseKey decodes a multibase-encoded public key.
func decodeMultibaseKey(multibase string) ([]byte, crypto.Algorithm, error) {
	const op = "jwt.decodeMultibaseKey"

	if len(multibase) < 2 {
		return nil, "", vc.ErrVerificationFailed(op, "multibase string too short")
	}

	// Check multibase prefix (z = base58btc)
	if multibase[0] != 'z' {
		return nil, "", vc.ErrVerificationFailed(op, "unsupported multibase prefix")
	}

	// Decode base58
	decoded, err := base58Decode(multibase[1:])
	if err != nil {
		return nil, "", vc.ErrVerificationFailed(op, "base58 decode failed: "+err.Error())
	}

	// Check multicodec prefix
	if len(decoded) < 2 {
		return nil, "", vc.ErrVerificationFailed(op, "decoded data too short")
	}

	// Ed25519: 0xed01
	if decoded[0] == 0xed && decoded[1] == 0x01 {
		return decoded[2:], crypto.AlgorithmEd25519, nil
	}

	// secp256k1: 0xe701
	if decoded[0] == 0xe7 && decoded[1] == 0x01 {
		return decoded[2:], crypto.AlgorithmSecp256k1, nil
	}

	return nil, "", vc.ErrVerificationFailed(op, "unsupported multicodec prefix")
}

// decodeBase58Key decodes a base58-encoded public key.
func decodeBase58Key(encoded string, algorithm crypto.Algorithm) ([]byte, crypto.Algorithm, error) {
	const op = "jwt.decodeBase58Key"

	decoded, err := base58Decode(encoded)
	if err != nil {
		return nil, "", vc.ErrVerificationFailed(op, "base58 decode failed: "+err.Error())
	}

	return decoded, algorithm, nil
}

// decodeJWK decodes a JWK public key.
func decodeJWK(jwk *did.JWK) ([]byte, crypto.Algorithm, error) {
	const op = "jwt.decodeJWK"

	switch jwk.Kty {
	case "OKP":
		if jwk.Crv == "Ed25519" {
			x, err := base64.RawURLEncoding.DecodeString(jwk.X)
			if err != nil {
				return nil, "", vc.ErrVerificationFailed(op, "failed to decode JWK x: "+err.Error())
			}
			return x, crypto.AlgorithmEd25519, nil
		}

	case "EC":
		if jwk.Crv == "secp256k1" {
			x, err := base64.RawURLEncoding.DecodeString(jwk.X)
			if err != nil {
				return nil, "", vc.ErrVerificationFailed(op, "failed to decode JWK x: "+err.Error())
			}
			y, err := base64.RawURLEncoding.DecodeString(jwk.Y)
			if err != nil {
				return nil, "", vc.ErrVerificationFailed(op, "failed to decode JWK y: "+err.Error())
			}
			// Compressed or uncompressed format depends on use case
			// For now, return concatenated x||y
			return append(x, y...), crypto.AlgorithmSecp256k1, nil
		}
	}

	return nil, "", vc.ErrVerificationFailed(op, "unsupported JWK type")
}

// verifySignature verifies the JWT signature.
func verifySignature(token *Token, publicKey []byte, algorithm crypto.Algorithm) error {
	const op = "jwt.verifySignature"

	// Verify algorithm matches header
	expectedAlg := algorithm.JWAName()
	if token.Header.Alg != expectedAlg {
		return vc.ErrVerificationFailed(op, "algorithm mismatch: expected "+expectedAlg+", got "+token.Header.Alg)
	}

	// Verify based on algorithm
	switch algorithm {
	case crypto.AlgorithmEd25519:
		return verifyEd25519(token.SigningInput, token.Signature, publicKey)

	case crypto.AlgorithmSecp256k1:
		// Phase 2
		return vc.ErrVerificationFailed(op, "secp256k1 verification not yet implemented")

	default:
		return vc.ErrVerificationFailed(op, "unsupported algorithm: "+algorithm.String())
	}
}

// verifyEd25519 verifies an Ed25519 signature.
func verifyEd25519(signingInput string, signature, publicKey []byte) error {
	const op = "jwt.verifyEd25519"

	verifier, err := ed25519.NewVerifier(publicKey)
	if err != nil {
		return vc.ErrVerificationFailed(op, "failed to create verifier: "+err.Error())
	}

	if err := verifier.Verify([]byte(signingInput), signature); err != nil {
		return vc.ErrInvalidProof(op, "signature verification failed: "+err.Error())
	}

	return nil
}

// tokenToCredential converts a JWT token to a Credential.
func tokenToCredential(token *Token) *vc.Credential {
	issuerDID, _ := did.Parse(token.Claims.Issuer)
	subjectDID, _ := did.Parse(token.Claims.Subject)

	// Update subject with parsed DID
	subject := token.Claims.VC.CredentialSubject
	subject.ID = subjectDID

	credential := &vc.Credential{
		Context:           token.Claims.VC.Context,
		ID:                token.Claims.JWTID,
		Type:              stringsToCredentialTypes(token.Claims.VC.Type),
		Issuer:            vc.Issuer{ID: issuerDID},
		IssuanceDate:      TimestampToTime(token.Claims.IssuedAt),
		CredentialSubject: subject,
		CredentialStatus:  token.Claims.VC.CredentialStatus,
		CredentialSchema:  token.Claims.VC.CredentialSchema,
		Evidence:          token.Claims.VC.Evidence,
	}

	if token.Claims.ExpiresAt != nil {
		exp := TimestampToTime(*token.Claims.ExpiresAt)
		credential.ExpirationDate = &exp
	}

	return credential
}

// stringsToCredentialTypes converts strings to credential types.
func stringsToCredentialTypes(strs []string) []vc.CredentialType {
	types := make([]vc.CredentialType, len(strs))
	for i, s := range strs {
		types[i] = vc.CredentialType(s)
	}
	return types
}

// validateTemporalClaims validates the temporal JWT claims.
func validateTemporalClaims(claims Claims, options vc.VerificationOptions) error {
	const op = "jwt.validateTemporalClaims"

	now := options.Now()

	// Check expiration
	if !options.AllowExpired && claims.ExpiresAt != nil {
		exp := TimestampToTime(*claims.ExpiresAt)
		if now.After(exp) {
			return vc.ErrCredentialExpired(op, claims.JWTID)
		}
	}

	// Check not before
	if !options.AllowNotYetValid {
		nbf := TimestampToTime(claims.NotBefore)
		if now.Before(nbf) {
			return vc.ErrCredentialNotActive(op, claims.JWTID)
		}
	}

	// Check max proof age
	if options.MaxProofAge != nil {
		iat := TimestampToTime(claims.IssuedAt)
		if now.Sub(iat) > *options.MaxProofAge {
			return vc.ErrProofExpired(op)
		}
	}

	return nil
}

// validateExpectedValues validates expected credential values.
func validateExpectedValues(credential *vc.Credential, options vc.VerificationOptions) error {
	const op = "jwt.validateExpectedValues"

	// Check expected issuer
	if options.ExpectedIssuer != nil && !credential.Issuer.ID.Equals(*options.ExpectedIssuer) {
		return vc.ErrInvalidIssuer(op, "issuer mismatch")
	}

	// Check expected subject
	if options.ExpectedSubject != nil && !credential.CredentialSubject.ID.Equals(*options.ExpectedSubject) {
		return vc.ErrInvalidSubject(op, "subject mismatch")
	}

	// Check expected types
	if len(options.ExpectedTypes) > 0 {
		for _, expectedType := range options.ExpectedTypes {
			if !credential.HasType(expectedType) {
				return vc.ErrInvalidCredential(op, "missing expected type: "+expectedType.String())
			}
		}
	}

	return nil
}

// ============================================================================
// Base58 Decoding
// ============================================================================

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// base58Decode decodes a base58 string.
func base58Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}

	alphabetIdx := make(map[rune]int)
	for i, c := range base58Alphabet {
		alphabetIdx[c] = i
	}

	zeros := 0
	for _, c := range input {
		if c != '1' {
			break
		}
		zeros++
	}

	size := len(input)*733/1000 + 1
	buf := make([]byte, size)

	for _, c := range input {
		idx, ok := alphabetIdx[c]
		if !ok {
			return nil, vc.ErrInvalidCredential("base58Decode", "invalid base58 character")
		}

		carry := idx
		for i := size - 1; i >= 0; i-- {
			carry += 58 * int(buf[i])
			buf[i] = byte(carry % 256)
			carry /= 256
		}
	}

	i := 0
	for i < len(buf) && buf[i] == 0 {
		i++
	}

	result := make([]byte, zeros+len(buf)-i)
	copy(result[zeros:], buf[i:])

	return result, nil
}
