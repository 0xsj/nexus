package vc

import (
	"testing"
	"time"
)

func TestNewProof(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewProof(ProofTypeEd25519Signature2020, vmID)

	if proof.Type != ProofTypeEd25519Signature2020 {
		t.Errorf("Type = %v, want %v", proof.Type, ProofTypeEd25519Signature2020)
	}

	if proof.VerificationMethod != vmID {
		t.Errorf("VerificationMethod = %v, want %v", proof.VerificationMethod, vmID)
	}

	if proof.ProofPurpose != ProofPurposeAssertionMethod {
		t.Errorf("ProofPurpose = %v, want %v", proof.ProofPurpose, ProofPurposeAssertionMethod)
	}

	if proof.Created.IsZero() {
		t.Error("Created should be set")
	}
}

func TestProof_WithCreated(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
	created := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	proof := NewProof(ProofTypeEd25519Signature2020, vmID).
		WithCreated(created)

	if !proof.Created.Equal(created) {
		t.Errorf("Created = %v, want %v", proof.Created, created)
	}
}

func TestProof_WithProofPurpose(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewProof(ProofTypeEd25519Signature2020, vmID).
		WithProofPurpose(ProofPurposeAuthentication)

	if proof.ProofPurpose != ProofPurposeAuthentication {
		t.Errorf("ProofPurpose = %v, want %v", proof.ProofPurpose, ProofPurposeAuthentication)
	}
}

func TestProof_WithProofValue(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewProof(ProofTypeEd25519Signature2020, vmID).
		WithProofValue("z3FXQfE...")

	if proof.ProofValue != "z3FXQfE..." {
		t.Errorf("ProofValue = %v, want z3FXQfE...", proof.ProofValue)
	}
}

func TestProof_WithJWS(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewProof(ProofTypeJWT, vmID).
		WithJWS("eyJhbGciOiJFZERTQSJ9...")

	if proof.JWS != "eyJhbGciOiJFZERTQSJ9..." {
		t.Errorf("JWS = %v, want eyJhbGciOiJFZERTQSJ9...", proof.JWS)
	}
}

func TestProof_WithNonce(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewProof(ProofTypeEd25519Signature2020, vmID).
		WithNonce("abc123")

	if proof.Nonce != "abc123" {
		t.Errorf("Nonce = %v, want abc123", proof.Nonce)
	}
}

func TestProof_WithDomain(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewProof(ProofTypeEd25519Signature2020, vmID).
		WithDomain("https://example.com")

	if proof.Domain != "https://example.com" {
		t.Errorf("Domain = %v, want https://example.com", proof.Domain)
	}
}

func TestProof_WithChallenge(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewProof(ProofTypeEd25519Signature2020, vmID).
		WithChallenge("challenge123")

	if proof.Challenge != "challenge123" {
		t.Errorf("Challenge = %v, want challenge123", proof.Challenge)
	}
}

func TestProof_Validate(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	// Valid proof with ProofValue
	proof := NewProof(ProofTypeEd25519Signature2020, vmID).
		WithProofValue("z3FXQfE...")

	if err := proof.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Valid proof with JWS
	proofJWT := NewProof(ProofTypeJWT, vmID).
		WithJWS("eyJhbGciOiJFZERTQSJ9...")

	if err := proofJWT.Validate(); err != nil {
		t.Errorf("Validate() unexpected error for JWT proof: %v", err)
	}
}

func TestProof_Validate_InvalidType(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := &Proof{
		Type:               ProofType("InvalidType"),
		Created:            time.Now(),
		VerificationMethod: vmID,
		ProofPurpose:       ProofPurposeAssertionMethod,
		ProofValue:         "z3FXQfE...",
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for invalid type")
	}
}

func TestProof_Validate_MissingCreated(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := &Proof{
		Type:               ProofTypeEd25519Signature2020,
		VerificationMethod: vmID,
		ProofPurpose:       ProofPurposeAssertionMethod,
		ProofValue:         "z3FXQfE...",
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for missing created")
	}
}

func TestProof_Validate_MissingVerificationMethod(t *testing.T) {
	proof := &Proof{
		Type:         ProofTypeEd25519Signature2020,
		Created:      time.Now(),
		ProofPurpose: ProofPurposeAssertionMethod,
		ProofValue:   "z3FXQfE...",
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for missing verificationMethod")
	}
}

func TestProof_Validate_MissingProofPurpose(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := &Proof{
		Type:               ProofTypeEd25519Signature2020,
		Created:            time.Now(),
		VerificationMethod: vmID,
		ProofValue:         "z3FXQfE...",
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for missing proofPurpose")
	}
}

func TestProof_Validate_MissingProofValueAndJWS(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := &Proof{
		Type:               ProofTypeEd25519Signature2020,
		Created:            time.Now(),
		VerificationMethod: vmID,
		ProofPurpose:       ProofPurposeAssertionMethod,
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for missing proofValue and jws")
	}
}

func TestProof_IsExpired(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	// Recent proof
	recentProof := NewProof(ProofTypeEd25519Signature2020, vmID).
		WithProofValue("z3FXQfE...")

	if recentProof.IsExpired(1 * time.Hour) {
		t.Error("IsExpired() should be false for recent proof")
	}

	// Old proof
	oldProof := &Proof{
		Type:               ProofTypeEd25519Signature2020,
		Created:            time.Now().Add(-2 * time.Hour),
		VerificationMethod: vmID,
		ProofPurpose:       ProofPurposeAssertionMethod,
		ProofValue:         "z3FXQfE...",
	}

	if !oldProof.IsExpired(1 * time.Hour) {
		t.Error("IsExpired() should be true for old proof")
	}
}

func TestProofPurposeConstants(t *testing.T) {
	// Verify proof purpose constants are correct
	if ProofPurposeAssertionMethod != "assertionMethod" {
		t.Errorf("ProofPurposeAssertionMethod = %v, want assertionMethod", ProofPurposeAssertionMethod)
	}

	if ProofPurposeAuthentication != "authentication" {
		t.Errorf("ProofPurposeAuthentication = %v, want authentication", ProofPurposeAuthentication)
	}

	if ProofPurposeKeyAgreement != "keyAgreement" {
		t.Errorf("ProofPurposeKeyAgreement = %v, want keyAgreement", ProofPurposeKeyAgreement)
	}

	if ProofPurposeCapabilityInvocation != "capabilityInvocation" {
		t.Errorf("ProofPurposeCapabilityInvocation = %v, want capabilityInvocation", ProofPurposeCapabilityInvocation)
	}

	if ProofPurposeCapabilityDelegation != "capabilityDelegation" {
		t.Errorf("ProofPurposeCapabilityDelegation = %v, want capabilityDelegation", ProofPurposeCapabilityDelegation)
	}
}

// ============================================================================
// Presentation Proof Tests
// ============================================================================

func TestNewPresentationProof(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
	holder := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewPresentationProof(ProofTypeEd25519Signature2020, vmID, holder)

	if proof.Type != ProofTypeEd25519Signature2020 {
		t.Errorf("Type = %v, want %v", proof.Type, ProofTypeEd25519Signature2020)
	}

	if proof.VerificationMethod != vmID {
		t.Errorf("VerificationMethod = %v, want %v", proof.VerificationMethod, vmID)
	}

	if proof.Holder != holder {
		t.Errorf("Holder = %v, want %v", proof.Holder, holder)
	}

	// Presentation proofs default to authentication purpose
	if proof.ProofPurpose != ProofPurposeAuthentication {
		t.Errorf("ProofPurpose = %v, want %v", proof.ProofPurpose, ProofPurposeAuthentication)
	}

	if proof.Created.IsZero() {
		t.Error("Created should be set")
	}
}

// ============================================================================
// Derived Proof Tests (BBS+)
// ============================================================================

func TestNewDerivedProof(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
	nonce := "verifier-nonce-123"

	proof := NewDerivedProof(vmID, nonce)

	if proof.Type != ProofTypeBbsBlsSignatureProof2020 {
		t.Errorf("Type = %v, want %v", proof.Type, ProofTypeBbsBlsSignatureProof2020)
	}

	if proof.VerificationMethod != vmID {
		t.Errorf("VerificationMethod = %v, want %v", proof.VerificationMethod, vmID)
	}

	if proof.Nonce != nonce {
		t.Errorf("Nonce = %v, want %v", proof.Nonce, nonce)
	}

	if proof.ProofPurpose != ProofPurposeAssertionMethod {
		t.Errorf("ProofPurpose = %v, want %v", proof.ProofPurpose, ProofPurposeAssertionMethod)
	}

	if proof.Created.IsZero() {
		t.Error("Created should be set")
	}
}

func TestDerivedProof_WithProofValue(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := NewDerivedProof(vmID, "nonce").
		WithProofValue("z5KpuB...")

	if proof.ProofValue != "z5KpuB..." {
		t.Errorf("ProofValue = %v, want z5KpuB...", proof.ProofValue)
	}
}

func TestDerivedProof_WithRevealedIndexes(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	indexes := []int{0, 2, 5}
	proof := NewDerivedProof(vmID, "nonce").
		WithRevealedIndexes(indexes)

	if len(proof.RevealedIndexes) != 3 {
		t.Errorf("RevealedIndexes count = %v, want 3", len(proof.RevealedIndexes))
	}

	for i, idx := range indexes {
		if proof.RevealedIndexes[i] != idx {
			t.Errorf("RevealedIndexes[%d] = %v, want %v", i, proof.RevealedIndexes[i], idx)
		}
	}
}

func TestDerivedProof_Validate(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	// Valid derived proof
	proof := NewDerivedProof(vmID, "nonce").
		WithProofValue("z5KpuB...")

	if err := proof.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestDerivedProof_Validate_WrongType(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := &DerivedProof{
		Type:               ProofTypeEd25519Signature2020, // Wrong type
		Created:            time.Now(),
		VerificationMethod: vmID,
		ProofPurpose:       ProofPurposeAssertionMethod,
		ProofValue:         "z5KpuB...",
		Nonce:              "nonce",
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for wrong type")
	}
}

func TestDerivedProof_Validate_MissingNonce(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := &DerivedProof{
		Type:               ProofTypeBbsBlsSignatureProof2020,
		Created:            time.Now(),
		VerificationMethod: vmID,
		ProofPurpose:       ProofPurposeAssertionMethod,
		ProofValue:         "z5KpuB...",
		// Missing nonce
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for missing nonce")
	}
}

func TestDerivedProof_Validate_MissingProofValue(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := &DerivedProof{
		Type:               ProofTypeBbsBlsSignatureProof2020,
		Created:            time.Now(),
		VerificationMethod: vmID,
		ProofPurpose:       ProofPurposeAssertionMethod,
		Nonce:              "nonce",
		// Missing proofValue
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for missing proofValue")
	}
}

func TestDerivedProof_Validate_MissingCreated(t *testing.T) {
	vmID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	proof := &DerivedProof{
		Type:               ProofTypeBbsBlsSignatureProof2020,
		VerificationMethod: vmID,
		ProofPurpose:       ProofPurposeAssertionMethod,
		ProofValue:         "z5KpuB...",
		Nonce:              "nonce",
		// Missing created
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for missing created")
	}
}

func TestDerivedProof_Validate_MissingVerificationMethod(t *testing.T) {
	proof := &DerivedProof{
		Type:         ProofTypeBbsBlsSignatureProof2020,
		Created:      time.Now(),
		ProofPurpose: ProofPurposeAssertionMethod,
		ProofValue:   "z5KpuB...",
		Nonce:        "nonce",
		// Missing verificationMethod
	}

	if err := proof.Validate(); err == nil {
		t.Error("Validate() expected error for missing verificationMethod")
	}
}
