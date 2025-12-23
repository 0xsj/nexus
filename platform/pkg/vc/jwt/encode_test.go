package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

func TestNewSigner(t *testing.T) {
	issuerDID, kp := key.MustGenerate()
	signer, err := ed25519.NewSigner(kp)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	config := vc.SignerConfig{
		IssuerDID:  issuerDID,
		KeyPair:    kp,
		Signer:     signer,
		Format:     vc.FormatJWT,
		IssuerName: "Test Issuer",
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

	if !jwtSigner.IssuerDID().Equals(issuerDID) {
		t.Errorf("IssuerDID() = %v, want %v", jwtSigner.IssuerDID(), issuerDID)
	}

	if jwtSigner.Algorithm() != crypto.AlgorithmEd25519 {
		t.Errorf("Algorithm() = %v, want %v", jwtSigner.Algorithm(), crypto.AlgorithmEd25519)
	}
}

func TestNewSigner_InvalidFormat(t *testing.T) {
	issuerDID, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJSONLD, // Not supported by JWT signer
	}

	_, err := NewSigner(config)
	if err == nil {
		t.Error("NewSigner() expected error for unsupported format")
	}
}

func TestNewSigner_MissingIssuer(t *testing.T) {
	_, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := vc.SignerConfig{
		KeyPair: kp,
		Signer:  signer,
		Format:  vc.FormatJWT,
	}

	_, err := NewSigner(config)
	if err == nil {
		t.Error("NewSigner() expected error for missing issuer")
	}
}

func TestNewSigner_MissingKeyPair(t *testing.T) {
	issuerDID, kp := key.MustGenerate()
	signer, _ := ed25519.NewSigner(kp)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	_, err := NewSigner(config)
	if err == nil {
		t.Error("NewSigner() expected error for missing key pair")
	}
}

func TestNewSigner_MissingSigner(t *testing.T) {
	issuerDID, kp := key.MustGenerate()

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   kp,
		Format:    vc.FormatJWT,
	}

	_, err := NewSigner(config)
	if err == nil {
		t.Error("NewSigner() expected error for missing signer")
	}
}

func TestSigner_Sign(t *testing.T) {
	issuerDID, kp := key.MustGenerate()
	subjectDID, _, err := key.Generate()
	if err != nil {
		t.Fatalf("key.Generate() error: %v", err)
	}

	signer, err := ed25519.NewSigner(kp)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	config := vc.SignerConfig{
		IssuerDID:  issuerDID,
		KeyPair:    kp,
		Signer:     signer,
		Format:     vc.FormatJWT,
		IssuerName: "Test Issuer",
	}

	jwtSigner, err := NewSigner(config)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}

	subject := vc.NewSubject(subjectDID).
		WithClaim("username", "testuser").
		WithClaim("commits", 500)

	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithType(vc.TypeGitHubContributorCredential).
		WithExpirationDate(time.Now().Add(24 * time.Hour))

	jwtBytes, err := jwtSigner.Sign(credential)
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

	if claims.Issuer != issuerDID.String() {
		t.Errorf("Claims.Issuer = %v, want %v", claims.Issuer, issuerDID.String())
	}

	if claims.Subject != subjectDID.String() {
		t.Errorf("Claims.Subject = %v, want %v", claims.Subject, subjectDID.String())
	}

	if claims.JWTID != "urn:uuid:test-123" {
		t.Errorf("Claims.JWTID = %v, want urn:uuid:test-123", claims.JWTID)
	}

	if claims.ExpiresAt == nil {
		t.Error("Claims.ExpiresAt should be set")
	}

	// Verify VC claim
	if len(claims.VC.Type) != 2 {
		t.Errorf("VC.Type count = %v, want 2", len(claims.VC.Type))
	}

	// Verify signature is present
	if len(parts[2]) == 0 {
		t.Error("Signature should not be empty")
	}
}

func TestSigner_Sign_InvalidCredential(t *testing.T) {
	issuerDID, kp := key.MustGenerate()

	signer, err := ed25519.NewSigner(kp)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, err := NewSigner(config)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}

	// Invalid credential - missing subject
	credential := &vc.Credential{
		Context:      vc.DefaultContext(),
		Type:         []vc.CredentialType{vc.TypeVerifiableCredential},
		Issuer:       vc.Issuer{ID: issuerDID},
		IssuanceDate: time.Now(),
	}

	_, err = jwtSigner.Sign(credential)
	if err == nil {
		t.Error("Sign() expected error for invalid credential")
	}
}

func TestSigner_Sign_SetsIssuer(t *testing.T) {
	issuerDID, kp := key.MustGenerate()
	subjectDID, _, err := key.Generate()
	if err != nil {
		t.Fatalf("key.Generate() error: %v", err)
	}

	signer, err := ed25519.NewSigner(kp)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	config := vc.SignerConfig{
		IssuerDID:  issuerDID,
		KeyPair:    kp,
		Signer:     signer,
		Format:     vc.FormatJWT,
		IssuerName: "Test Issuer",
		IssuerURL:  "https://issuer.example.com",
	}

	jwtSigner, err := NewSigner(config)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}

	subject := vc.NewSubject(subjectDID)

	// Credential with issuer ID but no name/URL
	credential := &vc.Credential{
		Context:           vc.DefaultContext(),
		ID:                "urn:uuid:test-123",
		Type:              []vc.CredentialType{vc.TypeVerifiableCredential},
		Issuer:            vc.Issuer{ID: issuerDID}, // ID is set, but no name/URL
		IssuanceDate:      time.Now(),
		CredentialSubject: subject,
	}

	jwtBytes, err := jwtSigner.Sign(credential)
	if err != nil {
		t.Fatalf("Sign() error: %v", err)
	}

	// Parse and verify issuer was set
	parts := strings.Split(string(jwtBytes), ".")
	claimsJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])

	var claims Claims
	json.Unmarshal(claimsJSON, &claims)

	if claims.Issuer != issuerDID.String() {
		t.Errorf("Issuer should be set from config, got %v", claims.Issuer)
	}

	// Verify issuer name and URL were set on credential
	if credential.Issuer.Name != "Test Issuer" {
		t.Errorf("Issuer.Name = %v, want Test Issuer", credential.Issuer.Name)
	}

	if credential.Issuer.URL != "https://issuer.example.com" {
		t.Errorf("Issuer.URL = %v, want https://issuer.example.com", credential.Issuer.URL)
	}
}

func TestEncodeCredential(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := vc.NewSubject(subjectDID).
		WithClaim("username", "testuser")

	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	encoded, err := EncodeCredential(credential, "EdDSA", "did:key:z6Mk...#key-1")
	if err != nil {
		t.Fatalf("EncodeCredential() error: %v", err)
	}

	// Should have 2 parts (no signature)
	parts := strings.Split(encoded, ".")
	if len(parts) != 2 {
		t.Errorf("Encoded credential should have 2 parts, got %d", len(parts))
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
}

func TestCreateJWT(t *testing.T) {
	header := Header{
		Alg: "EdDSA",
		Typ: "JWT",
		Kid: "did:key:z6Mk...#key-1",
	}

	claims := Claims{
		Issuer:    "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Subject:   "did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy",
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Unix(),
		JWTID:     "urn:uuid:test-123",
		VC: VCClaim{
			Context: vc.DefaultContext(),
			Type:    []string{"VerifiableCredential"},
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

func TestTimestampToTime(t *testing.T) {
	timestamp := int64(1704067200) // 2024-01-01 00:00:00 UTC
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	result := TimestampToTime(timestamp)

	if !result.Equal(expected) {
		t.Errorf("TimestampToTime() = %v, want %v", result, expected)
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
	exp := time.Now().Add(24 * time.Hour).Unix()

	original := Claims{
		Issuer:    "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Subject:   "did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy",
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Unix(),
		ExpiresAt: &exp,
		JWTID:     "urn:uuid:test-123",
		VC: VCClaim{
			Context: vc.DefaultContext(),
			Type:    []string{"VerifiableCredential", "GitHubContributorCredential"},
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

	if decoded.ExpiresAt == nil || *decoded.ExpiresAt != *original.ExpiresAt {
		t.Errorf("ExpiresAt mismatch: got %v, want %v", decoded.ExpiresAt, original.ExpiresAt)
	}

	if len(decoded.VC.Type) != len(original.VC.Type) {
		t.Errorf("VC.Type count mismatch: got %v, want %v", len(decoded.VC.Type), len(original.VC.Type))
	}
}

func TestSigner_SignAndVerifySignature(t *testing.T) {
	issuerDID, kp := key.MustGenerate()
	subjectDID, _, err := key.Generate()
	if err != nil {
		t.Fatalf("key.Generate() error: %v", err)
	}

	signer, err := ed25519.NewSigner(kp)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   kp,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, err := NewSigner(config)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}

	subject := vc.NewSubject(subjectDID).
		WithClaim("username", "testuser")

	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	jwtBytes, err := jwtSigner.Sign(credential)
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
