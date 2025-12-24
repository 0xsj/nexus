package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/presentation"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

func TestNewSigner(t *testing.T) {
	holderDID, kp := key.MustGenerate()
	signer, err := ed25519.NewSigner(kp)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, err := NewSigner(config)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}

	if jwtSigner == nil {
		t.Fatal("NewSigner() returned nil")
	}

	if jwtSigner.Format() != vc.FormatJWT {
		t.Errorf("Format() = %v, want %v", jwtSigner.Format(), vc.FormatJWT)
	}

	if !jwtSigner.HolderDID().Equals(holderDID) {
		t.Errorf("HolderDID() = %v, want %v", jwtSigner.HolderDID(), holderDID)
	}

	if jwtSigner.Algorithm() != crypto.AlgorithmEd25519 {
		t.Errorf("Algorithm() = %v, want %v", jwtSigner.Algorithm(), crypto.AlgorithmEd25519)
	}
}

func TestNewSigner_InvalidFormat(t *testing.T) {
	holderDID, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJSONLD, // Not supported
	}

	_, err := NewSigner(config)
	if err == nil {
		t.Error("NewSigner() expected error for unsupported format")
	}
}

func TestNewSigner_MissingHolder(t *testing.T) {
	_, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := presentation.SignerConfig{
		KeyPair: kp,
		Signer:  signer,
		Format:  vc.FormatJWT,
	}

	_, err := NewSigner(config)
	if err == nil {
		t.Error("NewSigner() expected error for missing holder")
	}
}

func TestNewSigner_MissingKeyPair(t *testing.T) {
	holderDID, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	_, err := NewSigner(config)
	if err == nil {
		t.Error("NewSigner() expected error for missing key pair")
	}
}

func TestNewSigner_MissingSigner(t *testing.T) {
	holderDID, kp := key.MustGenerate()

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   kp,
		Format:    vc.FormatJWT,
	}

	_, err := NewSigner(config)
	if err == nil {
		t.Error("NewSigner() expected error for missing signer")
	}
}

func TestSigner_Sign(t *testing.T) {
	holderDID, kp := key.MustGenerate()
	signer, err := ed25519.NewSigner(kp)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, err := NewSigner(config)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}

	// Create presentation with JWT credentials
	p := presentation.New(holderDID).
		WithCredentialJWTs("eyJhbGciOiJFZERTQSJ9.eyJpc3MiOiJ0ZXN0In0.dGVzdHNpZw")

	opts := presentation.SigningOptions{
		Challenge: "test-challenge-123",
		Domain:    "example.com",
	}

	jwtBytes, err := jwtSigner.Sign(p, opts)
	if err != nil {
		t.Fatalf("Sign() error: %v", err)
	}

	jwtStr := string(jwtBytes)

	// JWT should have 3 parts
	parts := strings.Split(jwtStr, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT should have 3 parts, got %d", len(parts))
	}

	// Verify header
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("Failed to decode header: %v", err)
	}

	var header Header
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatalf("Failed to unmarshal header: %v", err)
	}

	if header.Alg != "EdDSA" {
		t.Errorf("Header.Alg = %v, want EdDSA", header.Alg)
	}

	if header.Typ != "JWT" {
		t.Errorf("Header.Typ = %v, want JWT", header.Typ)
	}

	// Verify claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("Failed to decode claims: %v", err)
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		t.Fatalf("Failed to unmarshal claims: %v", err)
	}

	if claims.Issuer != holderDID.String() {
		t.Errorf("Claims.Issuer = %v, want %v", claims.Issuer, holderDID.String())
	}

	if claims.Subject != holderDID.String() {
		t.Errorf("Claims.Subject = %v, want %v", claims.Subject, holderDID.String())
	}

	// Verify VP claim
	if len(claims.VP.Type) == 0 {
		t.Error("VP.Type should not be empty")
	}

	if len(claims.VP.VerifiableCredential) != 1 {
		t.Errorf("VP.VerifiableCredential count = %d, want 1", len(claims.VP.VerifiableCredential))
	}
}

func TestSigner_Sign_InvalidPresentation(t *testing.T) {
	holderDID, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	// Invalid presentation - no credentials
	p := presentation.New(holderDID)

	_, err := jwtSigner.Sign(p, presentation.SigningOptions{})
	if err == nil {
		t.Error("Sign() expected error for invalid presentation")
	}
}

func TestSigner_Sign_SetsHolder(t *testing.T) {
	holderDID, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	// Presentation without holder set
	p := &presentation.Presentation{
		Context:                 presentation.DefaultContext(),
		ID:                      "urn:uuid:test-123",
		Type:                    []presentation.PresentationType{presentation.TypeVerifiablePresentation},
		VerifiableCredentialJWT: []string{"jwt1"},
	}

	jwtBytes, err := jwtSigner.Sign(p, presentation.SigningOptions{})
	if err != nil {
		t.Fatalf("Sign() error: %v", err)
	}

	// Verify holder was set
	if !p.Holder.Equals(holderDID) {
		t.Errorf("Holder should be set from config, got %v", p.Holder)
	}

	// Parse and verify claims
	parts := strings.Split(string(jwtBytes), ".")
	claimsJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])

	var claims Claims
	json.Unmarshal(claimsJSON, &claims)

	if claims.Issuer != holderDID.String() {
		t.Errorf("Issuer should be holder DID, got %v", claims.Issuer)
	}
}

func TestSigner_SignWithRequest(t *testing.T) {
	holderDID, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	request := presentation.NewRequest("challenge-from-request").
		WithDomain("verifier.example.com")

	jwtBytes, err := jwtSigner.SignWithRequest(p, request)
	if err != nil {
		t.Fatalf("SignWithRequest() error: %v", err)
	}

	if len(jwtBytes) == 0 {
		t.Error("SignWithRequest() returned empty bytes")
	}
}

func TestSigner_SignAndVerifySignature(t *testing.T) {
	holderDID, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	jwtBytes, err := jwtSigner.Sign(p, presentation.SigningOptions{})
	if err != nil {
		t.Fatalf("Sign() error: %v", err)
	}

	// Parse the JWT
	parts := strings.Split(string(jwtBytes), ".")
	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("Failed to decode signature: %v", err)
	}

	// Verify with Ed25519 verifier
	verifier, err := ed25519.NewVerifier(kp.PublicKey())
	if err != nil {
		t.Fatalf("ed25519.NewVerifier() error: %v", err)
	}

	if err := verifier.Verify([]byte(signingInput), signature); err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}
}

func TestEncodePresentation(t *testing.T) {
	holderDID, _ := key.MustGenerate()

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1", "jwt2")

	encoded, err := EncodePresentation(p, "EdDSA", "did:key:z6Mk...#key-1")
	if err != nil {
		t.Fatalf("EncodePresentation() error: %v", err)
	}

	// Should have 2 parts (no signature)
	parts := strings.Split(encoded, ".")
	if len(parts) != 2 {
		t.Errorf("Encoded presentation should have 2 parts, got %d", len(parts))
	}

	// Verify header
	headerJSON, _ := base64.RawURLEncoding.DecodeString(parts[0])
	var header Header
	json.Unmarshal(headerJSON, &header)

	if header.Alg != "EdDSA" {
		t.Errorf("Header.Alg = %v, want EdDSA", header.Alg)
	}

	if header.Kid != "did:key:z6Mk...#key-1" {
		t.Errorf("Header.Kid = %v, want did:key:z6Mk...#key-1", header.Kid)
	}

	// Verify claims
	claimsJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims Claims
	json.Unmarshal(claimsJSON, &claims)

	if len(claims.VP.VerifiableCredential) != 2 {
		t.Errorf("VP.VerifiableCredential count = %d, want 2", len(claims.VP.VerifiableCredential))
	}
}

func TestCreateJWT(t *testing.T) {
	header := Header{
		Alg: "EdDSA",
		Typ: "JWT",
		Kid: "did:key:z6Mk...#key-1",
	}

	claims := Claims{
		Issuer:   "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Subject:  "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		IssuedAt: 1704067200,
		JWTID:    "urn:uuid:test-123",
		VP: VPClaim{
			Context: presentation.DefaultContext(),
			Type:    []string{"VerifiablePresentation"},
		},
	}

	signature := []byte("test-signature")

	jwt := CreateJWT(header, claims, signature)

	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT should have 3 parts, got %d", len(parts))
	}

	// Verify signature part
	sigDecoded, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("Failed to decode signature: %v", err)
	}

	if string(sigDecoded) != "test-signature" {
		t.Errorf("Signature mismatch: got %v", string(sigDecoded))
	}
}

func TestHeader_JSONRoundTrip(t *testing.T) {
	original := Header{
		Alg: "EdDSA",
		Typ: "JWT",
		Kid: "did:key:z6Mk...#key-1",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded Header
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if decoded.Alg != original.Alg {
		t.Errorf("Alg mismatch: got %v, want %v", decoded.Alg, original.Alg)
	}

	if decoded.Typ != original.Typ {
		t.Errorf("Typ mismatch: got %v, want %v", decoded.Typ, original.Typ)
	}

	if decoded.Kid != original.Kid {
		t.Errorf("Kid mismatch: got %v, want %v", decoded.Kid, original.Kid)
	}
}

func TestClaims_JSONRoundTrip(t *testing.T) {
	exp := int64(1704153600)

	original := Claims{
		Issuer:    "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Subject:   "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		IssuedAt:  1704067200,
		NotBefore: 1704067200,
		ExpiresAt: &exp,
		JWTID:     "urn:uuid:test-123",
		Nonce:     "nonce-456",
		VP: VPClaim{
			Context:              presentation.DefaultContext(),
			Type:                 []string{"VerifiablePresentation"},
			VerifiableCredential: []string{"jwt1", "jwt2"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded Claims
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if decoded.Issuer != original.Issuer {
		t.Errorf("Issuer mismatch: got %v, want %v", decoded.Issuer, original.Issuer)
	}

	if decoded.Subject != original.Subject {
		t.Errorf("Subject mismatch: got %v, want %v", decoded.Subject, original.Subject)
	}

	if decoded.JWTID != original.JWTID {
		t.Errorf("JWTID mismatch: got %v, want %v", decoded.JWTID, original.JWTID)
	}

	if decoded.Nonce != original.Nonce {
		t.Errorf("Nonce mismatch: got %v, want %v", decoded.Nonce, original.Nonce)
	}

	if decoded.ExpiresAt == nil || *decoded.ExpiresAt != *original.ExpiresAt {
		t.Errorf("ExpiresAt mismatch: got %v, want %v", decoded.ExpiresAt, original.ExpiresAt)
	}

	if len(decoded.VP.VerifiableCredential) != len(original.VP.VerifiableCredential) {
		t.Errorf("VP.VerifiableCredential count mismatch: got %v, want %v",
			len(decoded.VP.VerifiableCredential), len(original.VP.VerifiableCredential))
	}
}

func TestAlgorithmToJWA(t *testing.T) {
	tests := []struct {
		alg  crypto.Algorithm
		want string
	}{
		{crypto.AlgorithmEd25519, "EdDSA"},
		{crypto.AlgorithmSecp256k1, "ES256K"},
		{crypto.Algorithm("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.alg), func(t *testing.T) {
			got := algorithmToJWA(tt.alg)
			if got != tt.want {
				t.Errorf("algorithmToJWA(%v) = %v, want %v", tt.alg, got, tt.want)
			}
		})
	}
}
