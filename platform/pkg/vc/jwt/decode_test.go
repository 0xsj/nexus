package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

func TestNewVerifier(t *testing.T) {
	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	if verifier == nil {
		t.Fatal("NewVerifier() returned nil")
	}

	formats := verifier.SupportedFormats()
	if len(formats) != 1 || formats[0] != vc.FormatJWT {
		t.Errorf("SupportedFormats() = %v, want [jwt_vc]", formats)
	}
}

func TestParse(t *testing.T) {
	// Create a valid JWT structure (unsigned for parsing test)
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := vc.NewSubject(subjectDID).
		WithClaim("username", "testuser")

	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	encoded, err := EncodeCredential(credential, "EdDSA", "did:key:z6Mk...#key-1")
	if err != nil {
		t.Fatalf("EncodeCredential() error: %v", err)
	}

	// Add a valid base64url-encoded fake signature
	fakeSignature := "dGVzdHNpZ25hdHVyZQ" // base64url of "testsignature"
	jwtStr := encoded + "." + fakeSignature

	token, err := Parse(jwtStr)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if token.Header.Alg != "EdDSA" {
		t.Errorf("Header.Alg = %v, want EdDSA", token.Header.Alg)
	}

	if token.Header.Typ != "JWT" {
		t.Errorf("Header.Typ = %v, want JWT", token.Header.Typ)
	}

	if token.Claims.Issuer != issuerDID.String() {
		t.Errorf("Claims.Issuer = %v, want %v", token.Claims.Issuer, issuerDID.String())
	}

	if token.Claims.Subject != subjectDID.String() {
		t.Errorf("Claims.Subject = %v, want %v", token.Claims.Subject, subjectDID.String())
	}

	if token.Claims.JWTID != "urn:uuid:test-123" {
		t.Errorf("Claims.JWTID = %v, want urn:uuid:test-123", token.Claims.JWTID)
	}
}

func TestParseUnverified(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	encoded, _ := EncodeCredential(credential, "EdDSA", "did:key:z6Mk...#key-1")

	// Add a valid base64url-encoded fake signature
	fakeSignature := "dGVzdHNpZ25hdHVyZQ" // base64url of "testsignature"
	jwtStr := encoded + "." + fakeSignature

	token, err := ParseUnverified(jwtStr)
	if err != nil {
		t.Fatalf("ParseUnverified() error: %v", err)
	}

	if token == nil {
		t.Fatal("ParseUnverified() returned nil")
	}
}

func TestParse_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"one part", "header"},
		{"two parts", "header.claims"},
		{"four parts", "a.b.c.d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Error("Parse() expected error for invalid format")
			}
		})
	}
}

func TestParse_InvalidBase64(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid header", "!!!invalid!!!.eyJpc3MiOiJ0ZXN0In0.sig"},
		{"invalid claims", "eyJhbGciOiJFZERTQSJ9.!!!invalid!!!.sig"},
		{"invalid signature", "eyJhbGciOiJFZERTQSJ9.eyJpc3MiOiJ0ZXN0In0.!!!invalid!!!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Error("Parse() expected error for invalid base64")
			}
		})
	}
}

func TestVerifier_Verify(t *testing.T) {
	// Generate issuer key pair and DID
	issuerDID, issuerKP := key.MustGenerate()

	// Generate subject DID
	subjectDID, _, err := key.Generate()
	if err != nil {
		t.Fatalf("key.Generate() error: %v", err)
	}

	// Create signer
	signer, err := ed25519.NewSigner(issuerKP)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	// Create JWT signer
	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, err := NewSigner(config)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}

	// Create and sign credential
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

	// Create verifier with did:key resolver
	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	// Verify
	ctx := context.Background()
	verified, err := verifier.Verify(ctx, jwtBytes)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}

	// Check verified credential
	if verified.ID != "urn:uuid:test-123" {
		t.Errorf("ID = %v, want urn:uuid:test-123", verified.ID)
	}

	if !verified.Issuer.ID.Equals(issuerDID) {
		t.Errorf("Issuer.ID = %v, want %v", verified.Issuer.ID, issuerDID)
	}

	if !verified.CredentialSubject.ID.Equals(subjectDID) {
		t.Errorf("Subject.ID = %v, want %v", verified.CredentialSubject.ID, subjectDID)
	}

	if !verified.HasType(vc.TypeGitHubContributorCredential) {
		t.Error("Should have GitHubContributorCredential type")
	}

	// Check claims
	username, ok := verified.CredentialSubject.GetStringClaim("username")
	if !ok || username != "testuser" {
		t.Errorf("username = %v, want testuser", username)
	}
}

func TestVerifier_Verify_ExpiredCredential(t *testing.T) {
	issuerDID, issuerKP := key.MustGenerate()
	subjectDID, _, _ := key.Generate()

	signer, _ := ed25519.NewSigner(issuerKP)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithExpirationDate(time.Now().Add(-1 * time.Hour)) // Expired

	jwtBytes, _ := jwtSigner.Sign(credential)

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	ctx := context.Background()
	_, err := verifier.Verify(ctx, jwtBytes)
	if err == nil {
		t.Error("Verify() expected error for expired credential")
	}
}

func TestVerifier_VerifyWithOptions_AllowExpired(t *testing.T) {
	issuerDID, issuerKP := key.MustGenerate()
	subjectDID, _, _ := key.Generate()

	signer, _ := ed25519.NewSigner(issuerKP)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithExpirationDate(time.Now().Add(-1 * time.Hour)) // Expired

	jwtBytes, _ := jwtSigner.Sign(credential)

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	ctx := context.Background()
	options := vc.DefaultVerificationOptions().WithAllowExpired(true)

	verified, err := verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err != nil {
		t.Fatalf("VerifyWithOptions() error: %v", err)
	}

	if verified == nil {
		t.Fatal("VerifyWithOptions() returned nil")
	}
}

func TestVerifier_VerifyWithOptions_ExpectedIssuer(t *testing.T) {
	issuerDID, issuerKP := key.MustGenerate()
	subjectDID, _, _ := key.Generate()
	wrongIssuerDID, _ := key.MustGenerate()

	signer, _ := ed25519.NewSigner(issuerKP)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	jwtBytes, _ := jwtSigner.Sign(credential)

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	ctx := context.Background()

	// Correct issuer
	options := vc.DefaultVerificationOptions().WithExpectedIssuer(issuerDID)
	_, err := verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err != nil {
		t.Errorf("VerifyWithOptions() unexpected error for correct issuer: %v", err)
	}

	// Wrong issuer
	options = vc.DefaultVerificationOptions().WithExpectedIssuer(wrongIssuerDID)
	_, err = verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err == nil {
		t.Error("VerifyWithOptions() expected error for wrong issuer")
	}
}

func TestVerifier_VerifyWithOptions_ExpectedSubject(t *testing.T) {
	issuerDID, issuerKP := key.MustGenerate()
	subjectDID, _, _ := key.Generate()
	wrongSubjectDID, _, _ := key.Generate()

	signer, _ := ed25519.NewSigner(issuerKP)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	jwtBytes, _ := jwtSigner.Sign(credential)

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	ctx := context.Background()

	// Correct subject
	options := vc.DefaultVerificationOptions().WithExpectedSubject(subjectDID)
	_, err := verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err != nil {
		t.Errorf("VerifyWithOptions() unexpected error for correct subject: %v", err)
	}

	// Wrong subject
	options = vc.DefaultVerificationOptions().WithExpectedSubject(wrongSubjectDID)
	_, err = verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err == nil {
		t.Error("VerifyWithOptions() expected error for wrong subject")
	}
}

func TestVerifier_VerifyWithOptions_ExpectedTypes(t *testing.T) {
	issuerDID, issuerKP := key.MustGenerate()
	subjectDID, _, _ := key.Generate()

	signer, _ := ed25519.NewSigner(issuerKP)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithType(vc.TypeGitHubContributorCredential)

	jwtBytes, _ := jwtSigner.Sign(credential)

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	ctx := context.Background()

	// Correct type
	options := vc.DefaultVerificationOptions().
		WithExpectedTypes(vc.TypeVerifiableCredential, vc.TypeGitHubContributorCredential)
	_, err := verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err != nil {
		t.Errorf("VerifyWithOptions() unexpected error for correct types: %v", err)
	}

	// Missing type
	options = vc.DefaultVerificationOptions().
		WithExpectedTypes(vc.TypeEducationCredential)
	_, err = verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err == nil {
		t.Error("VerifyWithOptions() expected error for missing type")
	}
}

func TestVerifier_VerifyWithOptions_MaxProofAge(t *testing.T) {
	issuerDID, issuerKP := key.MustGenerate()
	subjectDID, _, _ := key.Generate()

	signer, _ := ed25519.NewSigner(issuerKP)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	jwtBytes, _ := jwtSigner.Sign(credential)

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	ctx := context.Background()

	// Long max age - should pass
	options := vc.DefaultVerificationOptions().WithMaxProofAge(1 * time.Hour)
	_, err := verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err != nil {
		t.Errorf("VerifyWithOptions() unexpected error for valid max age: %v", err)
	}

	// Very short max age - should fail
	options = vc.DefaultVerificationOptions().WithMaxProofAge(1 * time.Nanosecond)
	// Need to wait a tiny bit for the proof to be "old"
	time.Sleep(2 * time.Nanosecond)
	_, err = verifier.VerifyWithOptions(ctx, jwtBytes, options)
	if err == nil {
		t.Error("VerifyWithOptions() expected error for expired proof age")
	}
}

func TestVerifier_Verify_InvalidSignature(t *testing.T) {
	issuerDID, issuerKP := key.MustGenerate()
	subjectDID, _, _ := key.Generate()

	signer, _ := ed25519.NewSigner(issuerKP)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	jwtBytes, _ := jwtSigner.Sign(credential)

	// Tamper with signature
	jwtStr := string(jwtBytes)
	jwtStr = jwtStr[:len(jwtStr)-5] + "XXXXX"

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	ctx := context.Background()
	_, err := verifier.Verify(ctx, []byte(jwtStr))
	if err == nil {
		t.Error("Verify() expected error for invalid signature")
	}
}

func TestVerifier_Verify_TamperedClaims(t *testing.T) {
	issuerDID, issuerKP := key.MustGenerate()
	subjectDID, _, _ := key.Generate()

	signer, _ := ed25519.NewSigner(issuerKP)

	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   issuerKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	subject := vc.NewSubject(subjectDID)
	credential := vc.NewCredential("urn:uuid:test-123", issuerDID, subject)

	jwtBytes, _ := jwtSigner.Sign(credential)

	// Parse and get parts
	token, _ := Parse(string(jwtBytes))

	// Create a different credential and use its claims
	subject2 := vc.NewSubject(subjectDID).WithClaim("commits", 9999)
	credential2 := vc.NewCredential("urn:uuid:tampered", issuerDID, subject2)
	encoded2, _ := EncodeCredential(credential2, "EdDSA", token.Header.Kid)

	// Use original signature with tampered claims
	parts := make([]string, 3)
	parts[0] = encoded2[:len(encoded2)-len(encoded2[len(encoded2)-1:])] // This won't work correctly
	// Actually, let's just replace the claims part
	originalParts := splitJWT(string(jwtBytes))
	tamperedParts := splitJWT(encoded2 + ".sig")

	tamperedJWT := originalParts[0] + "." + tamperedParts[1] + "." + originalParts[2]

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver)

	ctx := context.Background()
	_, err := verifier.Verify(ctx, []byte(tamperedJWT))
	if err == nil {
		t.Error("Verify() expected error for tampered claims")
	}
}

func splitJWT(jwt string) []string {
	parts := make([]string, 0, 3)
	start := 0
	for i := 0; i < len(jwt); i++ {
		if jwt[i] == '.' {
			parts = append(parts, jwt[start:i])
			start = i + 1
		}
	}
	parts = append(parts, jwt[start:])
	return parts
}

func TestBase58Decode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", false},
		{"valid", "3mJr7AoUXx2Wqd", false},
		{"leading ones", "111", false},
		{"invalid char 0", "0invalid", true},
		{"invalid char O", "Oinvalid", true},
		{"invalid char I", "Iinvalid", true},
		{"invalid char l", "linvalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := base58Decode(tt.input)
			if tt.wantErr && err == nil {
				t.Error("base58Decode() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("base58Decode() unexpected error: %v", err)
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want vc.Format
	}{
		{"empty", []byte{}, vc.FormatJWT},
		{"jwt prefix", []byte("eyJhbGciOiJFZERTQSJ9..."), vc.FormatJWT},
		{"json object", []byte(`{"@context":["https://..."]}`), vc.FormatJSONLD},
		{"random", []byte("random data"), vc.FormatJWT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vc.DetectFormat(tt.data)
			if got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
