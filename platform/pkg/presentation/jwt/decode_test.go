package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/presentation"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

func TestNewVerifier(t *testing.T) {
	resolver := key.NewResolver()
	verifier := NewVerifier(resolver, nil)

	if verifier == nil {
		t.Fatal("NewVerifier() returned nil")
	}

	formats := verifier.SupportedFormats()
	if len(formats) != 1 || formats[0] != vc.FormatJWT {
		t.Errorf("SupportedFormats() = %v, want [jwt_vc]", formats)
	}
}

func TestParse(t *testing.T) {
	holderDID, _ := key.MustGenerate()

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	encoded, err := EncodePresentation(p, "EdDSA", "did:key:z6Mk...#key-1")
	if err != nil {
		t.Fatalf("EncodePresentation() error: %v", err)
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

	if token.Claims.Issuer != holderDID.String() {
		t.Errorf("Claims.Issuer = %v, want %v", token.Claims.Issuer, holderDID.String())
	}

	if token.Claims.Subject != holderDID.String() {
		t.Errorf("Claims.Subject = %v, want %v", token.Claims.Subject, holderDID.String())
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
		{"invalid header", "!!!invalid!!!.eyJpc3MiOiJ0ZXN0In0.c2ln"},
		{"invalid claims", "eyJhbGciOiJFZERTQSJ9.!!!invalid!!!.c2ln"},
		{"invalid signature", "eyJhbGciOiJFZERTQSJ9.eyJpc3MiOiJ0ZXN0In0.!!!"},
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

func TestParseUnverified(t *testing.T) {
	holderDID, _ := key.MustGenerate()

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	encoded, _ := EncodePresentation(p, "EdDSA", "did:key:z6Mk...#key-1")
	fakeSignature := "dGVzdHNpZ25hdHVyZQ"
	jwtStr := encoded + "." + fakeSignature

	token, err := ParseUnverified(jwtStr)
	if err != nil {
		t.Fatalf("ParseUnverified() error: %v", err)
	}

	if token == nil {
		t.Fatal("ParseUnverified() returned nil")
	}
}

func TestVerifier_Verify(t *testing.T) {
	// Generate holder key pair and DID
	holderDID, holderKP := key.MustGenerate()

	// Create signer
	signer, err := ed25519.NewSigner(holderKP)
	if err != nil {
		t.Fatalf("ed25519.NewSigner() error: %v", err)
	}

	// Create JWT signer
	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   holderKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, err := NewSigner(config)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}

	// Create and sign presentation
	p := presentation.New(holderDID).
		WithCredentialJWTs("eyJhbGciOiJFZERTQSJ9.eyJpc3MiOiJ0ZXN0In0.dGVzdA")

	jwtBytes, err := jwtSigner.Sign(p, presentation.SigningOptions{})
	if err != nil {
		t.Fatalf("Sign() error: %v", err)
	}

	// Create verifier with did:key resolver
	resolver := key.NewResolver()
	verifier := NewVerifier(resolver, nil)

	// Verify
	ctx := context.Background()
	verified, err := verifier.Verify(ctx, jwtBytes)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}

	// Check verified presentation
	if verified.Presentation == nil {
		t.Fatal("Presentation should not be nil")
	}

	if !verified.Holder.Equals(holderDID) {
		t.Errorf("Holder = %v, want %v", verified.Holder, holderDID)
	}

	if verified.VerifiedAt.IsZero() {
		t.Error("VerifiedAt should be set")
	}
}

func TestVerifier_VerifyWithOptions_ExpectedHolder(t *testing.T) {
	holderDID, holderKP := key.MustGenerate()
	wrongHolderDID, _ := key.MustGenerate()

	signer, _ := ed25519.NewSigner(holderKP)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   holderKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	jwtBytes, _ := jwtSigner.Sign(p, presentation.SigningOptions{})

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver, nil)

	ctx := context.Background()

	// Correct holder
	opts := presentation.DefaultVerificationOptions().WithExpectedHolder(holderDID)
	_, err := verifier.VerifyWithOptions(ctx, jwtBytes, opts)
	if err != nil {
		t.Errorf("VerifyWithOptions() unexpected error for correct holder: %v", err)
	}

	// Wrong holder
	opts = presentation.DefaultVerificationOptions().WithExpectedHolder(wrongHolderDID)
	_, err = verifier.VerifyWithOptions(ctx, jwtBytes, opts)
	if err == nil {
		t.Error("VerifyWithOptions() expected error for wrong holder")
	}
}

func TestVerifier_VerifyWithOptions_Challenge(t *testing.T) {
	holderDID, holderKP := key.MustGenerate()

	signer, _ := ed25519.NewSigner(holderKP)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   holderKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	// Sign with nonce
	signingOpts := presentation.SigningOptions{
		Nonce: "challenge-123",
	}
	jwtBytes, _ := jwtSigner.Sign(p, signingOpts)

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver, nil)

	ctx := context.Background()

	// Correct challenge
	opts := presentation.DefaultVerificationOptions().WithChallenge("challenge-123")
	_, err := verifier.VerifyWithOptions(ctx, jwtBytes, opts)
	if err != nil {
		t.Errorf("VerifyWithOptions() unexpected error for correct challenge: %v", err)
	}

	// Wrong challenge
	opts = presentation.DefaultVerificationOptions().WithChallenge("wrong-challenge")
	_, err = verifier.VerifyWithOptions(ctx, jwtBytes, opts)
	if err == nil {
		t.Error("VerifyWithOptions() expected error for wrong challenge")
	}
}

func TestVerifier_VerifyWithOptions_MaxPresentationAge(t *testing.T) {
	holderDID, holderKP := key.MustGenerate()

	signer, _ := ed25519.NewSigner(holderKP)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   holderKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	jwtBytes, _ := jwtSigner.Sign(p, presentation.SigningOptions{})

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver, nil)

	ctx := context.Background()

	// Long max age - should pass
	opts := presentation.DefaultVerificationOptions().WithMaxPresentationAge(1 * time.Hour)
	_, err := verifier.VerifyWithOptions(ctx, jwtBytes, opts)
	if err != nil {
		t.Errorf("VerifyWithOptions() unexpected error for valid max age: %v", err)
	}

	// Very short max age - should fail
	opts = presentation.DefaultVerificationOptions().WithMaxPresentationAge(1 * time.Nanosecond)
	time.Sleep(2 * time.Millisecond)
	_, err = verifier.VerifyWithOptions(ctx, jwtBytes, opts)
	if err == nil {
		t.Error("VerifyWithOptions() expected error for expired presentation age")
	}
}

func TestVerifier_VerifyWithRequest(t *testing.T) {
	holderDID, holderKP := key.MustGenerate()

	signer, _ := ed25519.NewSigner(holderKP)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   holderKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	request := presentation.NewRequest("test-challenge").
		WithDomain("example.com")

	// Sign with request
	jwtBytes, _ := jwtSigner.SignWithRequest(p, request)

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver, nil)

	ctx := context.Background()

	// Verify with request
	verified, err := verifier.VerifyWithRequest(ctx, jwtBytes, request)
	if err != nil {
		t.Fatalf("VerifyWithRequest() error: %v", err)
	}

	if verified == nil {
		t.Fatal("VerifyWithRequest() returned nil")
	}
}

func TestVerifier_Verify_InvalidSignature(t *testing.T) {
	holderDID, holderKP := key.MustGenerate()

	signer, _ := ed25519.NewSigner(holderKP)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   holderKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	jwtBytes, _ := jwtSigner.Sign(p, presentation.SigningOptions{})

	// Tamper with signature
	jwtStr := string(jwtBytes)
	jwtStr = jwtStr[:len(jwtStr)-5] + "XXXXX"

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver, nil)

	ctx := context.Background()
	_, err := verifier.Verify(ctx, []byte(jwtStr))
	if err == nil {
		t.Error("Verify() expected error for invalid signature")
	}
}

func TestVerifier_Verify_TamperedClaims(t *testing.T) {
	holderDID, holderKP := key.MustGenerate()

	signer, _ := ed25519.NewSigner(holderKP)

	config := presentation.SignerConfig{
		HolderDID: holderDID,
		KeyPair:   holderKP,
		Signer:    signer,
		Format:    vc.FormatJWT,
	}

	jwtSigner, _ := NewSigner(config)

	p := presentation.New(holderDID).
		WithCredentialJWTs("jwt1")

	jwtBytes, _ := jwtSigner.Sign(p, presentation.SigningOptions{})

	// Create a different presentation and use its claims
	p2 := presentation.New(holderDID).
		WithCredentialJWTs("jwt1", "jwt2", "jwt3") // Different credentials

	encoded2, _ := EncodePresentation(p2, "EdDSA", holderDID.String()+"#"+holderDID.MethodSpecificID())

	// Use original signature with tampered claims
	originalParts := splitJWT(string(jwtBytes))
	tamperedParts := splitJWT(encoded2 + ".sig")

	tamperedJWT := originalParts[0] + "." + tamperedParts[1] + "." + originalParts[2]

	resolver := key.NewResolver()
	verifier := NewVerifier(resolver, nil)

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

func TestDecodeMultibase(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"too short", "z", true},
		{"valid base58btc", "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK", false},
		{"unsupported prefix", "m123456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeMultibase(tt.input)
			if tt.wantErr && err == nil {
				t.Error("decodeMultibase() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("decodeMultibase() unexpected error: %v", err)
			}
		})
	}
}

func TestDecodeBase58(t *testing.T) {
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
			_, err := decodeBase58(tt.input)
			if tt.wantErr && err == nil {
				t.Error("decodeBase58() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("decodeBase58() unexpected error: %v", err)
			}
		})
	}
}

func TestDecodeJWK(t *testing.T) {
	tests := []struct {
		name    string
		jwk     *did.JWK
		wantErr bool
	}{
		{
			name: "valid Ed25519",
			jwk: &did.JWK{
				Kty: "OKP",
				Crv: "Ed25519",
				X:   "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo", // Valid base64url
			},
			wantErr: false,
		},
		{
			name: "unsupported key type",
			jwk: &did.JWK{
				Kty: "RSA",
				X:   "abc",
			},
			wantErr: true,
		},
		{
			name: "unsupported curve",
			jwk: &did.JWK{
				Kty: "OKP",
				Crv: "X25519",
				X:   "abc",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeJWK(tt.jwk)
			if tt.wantErr && err == nil {
				t.Error("decodeJWK() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("decodeJWK() unexpected error: %v", err)
			}
		})
	}
}

func TestSigningInput(t *testing.T) {
	tests := []struct {
		name    string
		jwt     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid JWT",
			jwt:     "header.claims.signature",
			want:    "header.claims",
			wantErr: false,
		},
		{
			name:    "invalid JWT - two parts",
			jwt:     "header.claims",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid JWT - one part",
			jwt:     "header",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SigningInput(tt.jwt)
			if tt.wantErr && err == nil {
				t.Error("SigningInput() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("SigningInput() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("SigningInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
