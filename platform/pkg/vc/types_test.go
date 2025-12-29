package vc

import (
	"testing"
)

func TestFormat_String(t *testing.T) {
	tests := []struct {
		format Format
		want   string
	}{
		{FormatJWT, "jwt_vc"},
		{FormatJSONLD, "ldp_vc"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormat_Validate(t *testing.T) {
	tests := []struct {
		format  Format
		wantErr bool
	}{
		{FormatJWT, false},
		{FormatJSONLD, false},
		{Format("invalid"), true},
		{Format(""), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			err := tt.format.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestFormat_IsSupported(t *testing.T) {
	tests := []struct {
		format Format
		want   bool
	}{
		{FormatJWT, true},
		{FormatJSONLD, false}, // Phase 2
		{Format("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := tt.format.IsSupported(); got != tt.want {
				t.Errorf("IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofType_String(t *testing.T) {
	tests := []struct {
		proofType ProofType
		want      string
	}{
		{ProofTypeJWT, "JwtProof2020"},
		{ProofTypeEd25519Signature2020, "Ed25519Signature2020"},
		{ProofTypeEcdsaSecp256k1Signature2019, "EcdsaSecp256k1Signature2019"},
		{ProofTypeBbsBlsSignature2020, "BbsBlsSignature2020"},
		{ProofTypeBbsBlsSignatureProof2020, "BbsBlsSignatureProof2020"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.proofType.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofType_Validate(t *testing.T) {
	tests := []struct {
		proofType ProofType
		wantErr   bool
	}{
		{ProofTypeJWT, false},
		{ProofTypeEd25519Signature2020, false},
		{ProofTypeEcdsaSecp256k1Signature2019, false},
		{ProofTypeBbsBlsSignature2020, false},
		{ProofTypeBbsBlsSignatureProof2020, false},
		{ProofType("invalid"), true},
		{ProofType(""), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.proofType), func(t *testing.T) {
			err := tt.proofType.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestProofType_IsSupported(t *testing.T) {
	tests := []struct {
		proofType ProofType
		want      bool
	}{
		{ProofTypeJWT, true},
		{ProofTypeEd25519Signature2020, true},
		{ProofTypeEcdsaSecp256k1Signature2019, false}, // Phase 2
		{ProofTypeBbsBlsSignature2020, false},         // Phase 2
		{ProofTypeBbsBlsSignatureProof2020, false},    // Phase 2
		{ProofType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.proofType), func(t *testing.T) {
			if got := tt.proofType.IsSupported(); got != tt.want {
				t.Errorf("IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofType_SupportsSelectiveDisclosure(t *testing.T) {
	tests := []struct {
		proofType ProofType
		want      bool
	}{
		{ProofTypeJWT, false},
		{ProofTypeEd25519Signature2020, false},
		{ProofTypeEcdsaSecp256k1Signature2019, false},
		{ProofTypeBbsBlsSignature2020, true},
		{ProofTypeBbsBlsSignatureProof2020, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.proofType), func(t *testing.T) {
			if got := tt.proofType.SupportsSelectiveDisclosure(); got != tt.want {
				t.Errorf("SupportsSelectiveDisclosure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusActive, "active"},
		{StatusRevoked, "revoked"},
		{StatusExpired, "expired"},
		{StatusSuspended, "suspended"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_Validate(t *testing.T) {
	tests := []struct {
		status  Status
		wantErr bool
	}{
		{StatusActive, false},
		{StatusRevoked, false},
		{StatusExpired, false},
		{StatusSuspended, false},
		{Status("invalid"), true},
		{Status(""), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			err := tt.status.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusActive, true},
		{StatusRevoked, false},
		{StatusExpired, false},
		{StatusSuspended, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialType_String(t *testing.T) {
	tests := []struct {
		credType CredentialType
		want     string
	}{
		{TypeVerifiableCredential, "VerifiableCredential"},
		{TypeVerifiablePresentation, "VerifiablePresentation"},
		{TypeGitHubContributorCredential, "GitHubContributorCredential"},
		{TypeLinkedInEmploymentCredential, "LinkedInEmploymentCredential"},
		{TypeEducationCredential, "EducationCredential"},
		{TypeCertificationCredential, "CertificationCredential"},
		{TypeIdentityCredential, "IdentityCredential"},
		{TypeSkillCredential, "SkillCredential"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.credType.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultContext(t *testing.T) {
	ctx := DefaultContext()

	if len(ctx) != 2 {
		t.Errorf("DefaultContext() length = %v, want 2", len(ctx))
	}

	if ctx[0] != ContextCredentialsV1 {
		t.Errorf("DefaultContext()[0] = %v, want %v", ctx[0], ContextCredentialsV1)
	}

	if ctx[1] != ContextEd25519 {
		t.Errorf("DefaultContext()[1] = %v, want %v", ctx[1], ContextEd25519)
	}
}

func TestContextConstants(t *testing.T) {
	// Verify context URIs are correct
	if ContextCredentialsV1 != "https://www.w3.org/2018/credentials/v1" {
		t.Errorf("ContextCredentialsV1 = %v, want https://www.w3.org/2018/credentials/v1", ContextCredentialsV1)
	}

	if ContextCredentialsV2 != "https://www.w3.org/ns/credentials/v2" {
		t.Errorf("ContextCredentialsV2 = %v, want https://www.w3.org/ns/credentials/v2", ContextCredentialsV2)
	}

	if ContextEd25519 != "https://w3id.org/security/suites/ed25519-2020/v1" {
		t.Errorf("ContextEd25519 = %v, want https://w3id.org/security/suites/ed25519-2020/v1", ContextEd25519)
	}

	if ContextBBS != "https://w3id.org/security/bbs/v1" {
		t.Errorf("ContextBBS = %v, want https://w3id.org/security/bbs/v1", ContextBBS)
	}
}
