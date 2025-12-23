package key

import (
	"testing"

	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
)

func TestGenerate(t *testing.T) {
	d, kp, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Check DID is valid
	if d.IsZero() {
		t.Error("Generate() returned zero DID")
	}

	if d.Method() != did.MethodKey {
		t.Errorf("Method() = %v, want %v", d.Method(), did.MethodKey)
	}

	// Check key pair is valid
	if kp == nil {
		t.Fatal("Generate() returned nil key pair")
	}

	if !kp.HasPrivateKey() {
		t.Error("Generated key pair should have private key")
	}

	if len(kp.PublicKey()) != ed25519.PublicKeySize {
		t.Errorf("PublicKey size = %v, want %v", len(kp.PublicKey()), ed25519.PublicKeySize)
	}

	// Check DID starts with correct prefix
	didStr := d.String()
	if didStr[:8] != "did:key:" {
		t.Errorf("DID should start with 'did:key:', got %v", didStr[:8])
	}

	// Method-specific ID should start with 'z' (multibase base58btc)
	methodSpecificID := d.MethodSpecificID()
	if methodSpecificID[0] != 'z' {
		t.Errorf("MethodSpecificID should start with 'z', got %v", methodSpecificID[0])
	}
}

func TestGenerate_Unique(t *testing.T) {
	d1, _, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	d2, _, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if d1.Equals(d2) {
		t.Error("Generate() should produce unique DIDs")
	}
}

func TestFromKeyPair(t *testing.T) {
	// Generate a key pair
	kp, err := ed25519.Generate()
	if err != nil {
		t.Fatalf("ed25519.Generate() error: %v", err)
	}

	// Create DID from key pair
	d, err := FromKeyPair(kp)
	if err != nil {
		t.Fatalf("FromKeyPair() error: %v", err)
	}

	if d.IsZero() {
		t.Error("FromKeyPair() returned zero DID")
	}

	if d.Method() != did.MethodKey {
		t.Errorf("Method() = %v, want %v", d.Method(), did.MethodKey)
	}

	// Create another DID from same key pair - should be identical
	d2, err := FromKeyPair(kp)
	if err != nil {
		t.Fatalf("FromKeyPair() error: %v", err)
	}

	if !d.Equals(d2) {
		t.Error("FromKeyPair() with same key pair should produce same DID")
	}
}

func TestFromKeyPair_Nil(t *testing.T) {
	_, err := FromKeyPair(nil)
	if err == nil {
		t.Error("FromKeyPair(nil) should return error")
	}
}

func TestFromPublicKey(t *testing.T) {
	// Generate a key pair
	kp, err := ed25519.Generate()
	if err != nil {
		t.Fatalf("ed25519.Generate() error: %v", err)
	}

	// Create DID from public key only
	d, err := FromPublicKey(kp.PublicKey())
	if err != nil {
		t.Fatalf("FromPublicKey() error: %v", err)
	}

	if d.IsZero() {
		t.Error("FromPublicKey() returned zero DID")
	}

	// Should match DID from full key pair
	d2, err := FromKeyPair(kp)
	if err != nil {
		t.Fatalf("FromKeyPair() error: %v", err)
	}

	if !d.Equals(d2) {
		t.Error("FromPublicKey() should produce same DID as FromKeyPair()")
	}
}

func TestFromPublicKey_InvalidSize(t *testing.T) {
	// Too short
	_, err := FromPublicKey([]byte{1, 2, 3})
	if err == nil {
		t.Error("FromPublicKey() with invalid size should return error")
	}

	// Too long
	_, err = FromPublicKey(make([]byte, 64))
	if err == nil {
		t.Error("FromPublicKey() with invalid size should return error")
	}

	// Empty
	_, err = FromPublicKey([]byte{})
	if err == nil {
		t.Error("FromPublicKey() with empty key should return error")
	}
}

func TestFromMultibase(t *testing.T) {
	// Generate a key pair to get a valid multibase string
	kp, err := ed25519.Generate()
	if err != nil {
		t.Fatalf("ed25519.Generate() error: %v", err)
	}

	multibase := kp.PublicKeyMultibase()

	d, err := FromMultibase(multibase)
	if err != nil {
		t.Fatalf("FromMultibase() error: %v", err)
	}

	if d.IsZero() {
		t.Error("FromMultibase() returned zero DID")
	}

	if d.MethodSpecificID() != multibase {
		t.Errorf("MethodSpecificID() = %v, want %v", d.MethodSpecificID(), multibase)
	}
}

func TestFromMultibase_Empty(t *testing.T) {
	_, err := FromMultibase("")
	if err == nil {
		t.Error("FromMultibase(\"\") should return error")
	}
}

func TestMustGenerate(t *testing.T) {
	// Should not panic
	d, kp := MustGenerate()

	if d.IsZero() {
		t.Error("MustGenerate() returned zero DID")
	}

	if kp == nil {
		t.Error("MustGenerate() returned nil key pair")
	}
}

func TestGenerate_DIDResolvable(t *testing.T) {
	// Generate a DID
	d, _, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// The DID should be parseable back
	parsed, err := did.Parse(d.String())
	if err != nil {
		t.Fatalf("did.Parse() error: %v", err)
	}

	if !d.Equals(parsed) {
		t.Error("Generated DID should be parseable")
	}
}
