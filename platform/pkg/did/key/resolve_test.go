package key

import (
	"context"
	"testing"

	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
)

func TestResolver_Method(t *testing.T) {
	r := NewResolver()

	if r.Method() != did.MethodKey {
		t.Errorf("Method() = %v, want %v", r.Method(), did.MethodKey)
	}
}

func TestResolver_Resolve(t *testing.T) {
	r := NewResolver()
	ctx := context.Background()

	// Generate a DID
	d, kp, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Resolve it
	doc, err := r.Resolve(ctx, d)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}

	// Check document ID
	if !doc.ID.Equals(d) {
		t.Errorf("Document ID = %v, want %v", doc.ID, d)
	}

	// Check context
	if len(doc.Context) < 1 {
		t.Error("Document should have context")
	}

	// Check verification method
	if len(doc.VerificationMethod) != 1 {
		t.Fatalf("VerificationMethod count = %v, want 1", len(doc.VerificationMethod))
	}

	vm := doc.VerificationMethod[0]
	if vm.Type != did.Ed25519VerificationKey2020 {
		t.Errorf("VerificationMethod Type = %v, want %v", vm.Type, did.Ed25519VerificationKey2020)
	}

	if !vm.Controller.Equals(d) {
		t.Errorf("VerificationMethod Controller = %v, want %v", vm.Controller, d)
	}

	if vm.PublicKeyMultibase != kp.PublicKeyMultibase() {
		t.Errorf("VerificationMethod PublicKeyMultibase = %v, want %v",
			vm.PublicKeyMultibase, kp.PublicKeyMultibase())
	}

	// Check authentication
	if len(doc.Authentication) != 1 {
		t.Errorf("Authentication count = %v, want 1", len(doc.Authentication))
	}

	// Check assertion method
	if len(doc.AssertionMethod) != 1 {
		t.Errorf("AssertionMethod count = %v, want 1", len(doc.AssertionMethod))
	}

	// Check capability invocation
	if len(doc.CapabilityInvocation) != 1 {
		t.Errorf("CapabilityInvocation count = %v, want 1", len(doc.CapabilityInvocation))
	}

	// Check capability delegation
	if len(doc.CapabilityDelegation) != 1 {
		t.Errorf("CapabilityDelegation count = %v, want 1", len(doc.CapabilityDelegation))
	}
}

func TestResolver_Resolve_WrongMethod(t *testing.T) {
	r := NewResolver()
	ctx := context.Background()

	// Create a did:web DID
	d, err := did.New(did.MethodWeb, "example.com")
	if err != nil {
		t.Fatalf("did.New() error: %v", err)
	}

	// Resolve should fail
	_, err = r.Resolve(ctx, d)
	if err == nil {
		t.Error("Resolve() should fail for non did:key DID")
	}
}

func TestResolver_Resolve_InvalidMultibase(t *testing.T) {
	r := NewResolver()
	ctx := context.Background()

	tests := []struct {
		name             string
		methodSpecificID string
	}{
		{
			name:             "invalid prefix",
			methodSpecificID: "a1234567890", // 'a' is not a valid multibase prefix we support
		},
		{
			name:             "too short",
			methodSpecificID: "z",
		},
		{
			name:             "invalid base58",
			methodSpecificID: "z0OIl", // contains invalid base58 chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := did.New(did.MethodKey, tt.methodSpecificID)
			if err != nil {
				t.Fatalf("did.New() error: %v", err)
			}

			_, err = r.Resolve(ctx, d)
			if err == nil {
				t.Error("Resolve() should fail for invalid multibase")
			}
		})
	}
}

func TestExtractPublicKey(t *testing.T) {
	// Generate a DID
	d, kp, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Extract public key
	publicKey, algorithm, err := ExtractPublicKey(d)
	if err != nil {
		t.Fatalf("ExtractPublicKey() error: %v", err)
	}

	// Check algorithm
	if algorithm != crypto.AlgorithmEd25519 {
		t.Errorf("Algorithm = %v, want %v", algorithm, crypto.AlgorithmEd25519)
	}

	// Check public key matches
	if len(publicKey) != ed25519.PublicKeySize {
		t.Errorf("PublicKey size = %v, want %v", len(publicKey), ed25519.PublicKeySize)
	}

	// Compare with original
	for i, b := range publicKey {
		if b != kp.PublicKey()[i] {
			t.Errorf("PublicKey mismatch at byte %d", i)
			break
		}
	}
}

func TestExtractPublicKey_WrongMethod(t *testing.T) {
	d, err := did.New(did.MethodWeb, "example.com")
	if err != nil {
		t.Fatalf("did.New() error: %v", err)
	}

	_, _, err = ExtractPublicKey(d)
	if err == nil {
		t.Error("ExtractPublicKey() should fail for non did:key DID")
	}
}

func TestToDID(t *testing.T) {
	// Generate a key pair to get valid multibase
	kp, err := ed25519.Generate()
	if err != nil {
		t.Fatalf("ed25519.Generate() error: %v", err)
	}

	multibase := kp.PublicKeyMultibase()

	d, err := ToDID(multibase)
	if err != nil {
		t.Fatalf("ToDID() error: %v", err)
	}

	if d.Method() != did.MethodKey {
		t.Errorf("Method() = %v, want %v", d.Method(), did.MethodKey)
	}

	if d.MethodSpecificID() != multibase {
		t.Errorf("MethodSpecificID() = %v, want %v", d.MethodSpecificID(), multibase)
	}
}

func TestToDID_Invalid(t *testing.T) {
	_, err := ToDID("invalid")
	if err == nil {
		t.Error("ToDID() should fail for invalid multibase")
	}
}

func TestFingerprint(t *testing.T) {
	kp, err := ed25519.Generate()
	if err != nil {
		t.Fatalf("ed25519.Generate() error: %v", err)
	}

	fingerprint := Fingerprint(kp.PublicKey(), crypto.AlgorithmEd25519)

	// Should start with 'z' (base58btc multibase prefix)
	if fingerprint[0] != 'z' {
		t.Errorf("Fingerprint should start with 'z', got %c", fingerprint[0])
	}

	// Should match the key pair's multibase
	if fingerprint != kp.PublicKeyMultibase() {
		t.Errorf("Fingerprint() = %v, want %v", fingerprint, kp.PublicKeyMultibase())
	}
}

func TestFingerprint_UnsupportedAlgorithm(t *testing.T) {
	fingerprint := Fingerprint([]byte{1, 2, 3}, crypto.Algorithm("unsupported"))

	if fingerprint != "" {
		t.Errorf("Fingerprint() with unsupported algorithm should return empty string, got %v", fingerprint)
	}
}

func TestIsKeyDID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK", true},
		{"did:key:z6Mk...", true},
		{"did:web:example.com", false},
		{"did:pkh:eip155:1:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb", false},
		{"not-a-did", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsKeyDID(tt.input)
			if got != tt.want {
				t.Errorf("IsKeyDID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBase58RoundTrip(t *testing.T) {
	testCases := [][]byte{
		{},
		{0},
		{0, 0, 0},
		{1, 2, 3, 4, 5},
		{0, 1, 2, 3, 4, 5},
		make([]byte, 32),
		{255, 255, 255, 255},
	}

	for i, input := range testCases {
		encoded := base58Encode(input)
		decoded, err := base58Decode(encoded)
		if err != nil {
			t.Errorf("case %d: base58Decode() error: %v", i, err)
			continue
		}

		if len(decoded) != len(input) {
			t.Errorf("case %d: length mismatch: got %d, want %d", i, len(decoded), len(input))
			continue
		}

		for j := range input {
			if decoded[j] != input[j] {
				t.Errorf("case %d: byte %d mismatch: got %d, want %d", i, j, decoded[j], input[j])
				break
			}
		}
	}
}

func TestBase58Decode_InvalidChar(t *testing.T) {
	// '0', 'O', 'I', 'l' are not in base58 alphabet
	invalidStrings := []string{
		"0abc",
		"Oabc",
		"Iabc",
		"labc",
	}

	for _, s := range invalidStrings {
		_, err := base58Decode(s)
		if err == nil {
			t.Errorf("base58Decode(%q) should return error for invalid character", s)
		}
	}
}

func TestResolver_Resolve_Deterministic(t *testing.T) {
	r := NewResolver()
	ctx := context.Background()

	// Generate a DID
	d, _, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Resolve multiple times
	doc1, err := r.Resolve(ctx, d)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}

	doc2, err := r.Resolve(ctx, d)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}

	// Should produce identical documents
	if !doc1.ID.Equals(doc2.ID) {
		t.Error("Resolve() should be deterministic - IDs don't match")
	}

	if len(doc1.VerificationMethod) != len(doc2.VerificationMethod) {
		t.Error("Resolve() should be deterministic - verification method count doesn't match")
	}

	if doc1.VerificationMethod[0].PublicKeyMultibase != doc2.VerificationMethod[0].PublicKeyMultibase {
		t.Error("Resolve() should be deterministic - public keys don't match")
	}
}
