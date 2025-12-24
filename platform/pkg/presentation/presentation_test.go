package presentation

import (
	"strings"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

func TestNew(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	p := New(holderDID)

	if p == nil {
		t.Fatal("New() returned nil")
	}

	if p.ID == "" {
		t.Error("ID should be generated")
	}

	if len(p.Context) == 0 {
		t.Error("Context should have default values")
	}

	if len(p.Type) != 1 || p.Type[0] != TypeVerifiablePresentation {
		t.Error("Type should include VerifiablePresentation")
	}

	if !p.Holder.Equals(holderDID) {
		t.Errorf("Holder = %v, want %v", p.Holder, holderDID)
	}
}

func TestPresentation_WithCredentials(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	issuerDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := vc.NewSubject(holderDID)
	cred1 := *vc.NewCredential("urn:uuid:cred-1", issuerDID, subject)
	cred2 := *vc.NewCredential("urn:uuid:cred-2", issuerDID, subject)

	p := New(holderDID).
		WithCredentials(cred1, cred2)

	if len(p.VerifiableCredential) != 2 {
		t.Errorf("VerifiableCredential count = %d, want 2", len(p.VerifiableCredential))
	}
}

func TestPresentation_WithCredentialJWTs(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	jwt1 := "eyJhbGciOiJFZERTQSJ9.eyJpc3MiOiJkaWQ6a2V5Onoifq.sig1"
	jwt2 := "eyJhbGciOiJFZERTQSJ9.eyJpc3MiOiJkaWQ6a2V5Onoifq.sig2"

	p := New(holderDID).
		WithCredentialJWTs(jwt1, jwt2)

	if len(p.VerifiableCredentialJWT) != 2 {
		t.Errorf("VerifiableCredentialJWT count = %d, want 2", len(p.VerifiableCredentialJWT))
	}
}

func TestPresentation_WithType(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	customType := PresentationType("CustomPresentation")

	p := New(holderDID).
		WithType(customType)

	if len(p.Type) != 2 {
		t.Errorf("Type count = %d, want 2", len(p.Type))
	}

	if !p.HasType(customType) {
		t.Error("Should have custom type")
	}
}

func TestPresentation_WithContext(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	customContext := "https://example.com/custom/v1"

	p := New(holderDID).
		WithContext(customContext)

	if len(p.Context) != 2 {
		t.Errorf("Context count = %d, want 2", len(p.Context))
	}

	found := false
	for _, ctx := range p.Context {
		if ctx == customContext {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have custom context")
	}
}

func TestPresentation_WithID(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	customID := "urn:uuid:custom-id-123"

	p := New(holderDID).
		WithID(customID)

	if p.ID != customID {
		t.Errorf("ID = %v, want %v", p.ID, customID)
	}
}

func TestPresentation_Validate(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	tests := []struct {
		name    string
		setup   func() *Presentation
		wantErr bool
	}{
		{
			name: "valid with credential JWTs",
			setup: func() *Presentation {
				return New(holderDID).
					WithCredentialJWTs("jwt1", "jwt2")
			},
			wantErr: false,
		},
		{
			name: "valid with credentials",
			setup: func() *Presentation {
				issuerDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
				subject := vc.NewSubject(holderDID)
				cred := *vc.NewCredential("urn:uuid:cred-1", issuerDID, subject)
				return New(holderDID).WithCredentials(cred)
			},
			wantErr: false,
		},
		{
			name: "missing context",
			setup: func() *Presentation {
				p := New(holderDID).WithCredentialJWTs("jwt1")
				p.Context = nil
				return p
			},
			wantErr: true,
		},
		{
			name: "missing type",
			setup: func() *Presentation {
				p := New(holderDID).WithCredentialJWTs("jwt1")
				p.Type = nil
				return p
			},
			wantErr: true,
		},
		{
			name: "missing VerifiablePresentation type",
			setup: func() *Presentation {
				p := New(holderDID).WithCredentialJWTs("jwt1")
				p.Type = []PresentationType{"OtherType"}
				return p
			},
			wantErr: true,
		},
		{
			name: "missing holder",
			setup: func() *Presentation {
				p := New(holderDID).WithCredentialJWTs("jwt1")
				p.Holder = did.DID{}
				return p
			},
			wantErr: true,
		},
		{
			name: "no credentials",
			setup: func() *Presentation {
				return New(holderDID)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			err := p.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestPresentation_CredentialCount(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	issuerDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := vc.NewSubject(holderDID)
	cred := *vc.NewCredential("urn:uuid:cred-1", issuerDID, subject)

	p := New(holderDID).
		WithCredentials(cred).
		WithCredentialJWTs("jwt1", "jwt2")

	if p.CredentialCount() != 3 {
		t.Errorf("CredentialCount() = %d, want 3", p.CredentialCount())
	}

	if !p.HasCredentials() {
		t.Error("HasCredentials() should return true")
	}
}

func TestPresentation_GetCredentialByID(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	issuerDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := vc.NewSubject(holderDID)
	cred1 := *vc.NewCredential("urn:uuid:cred-1", issuerDID, subject)
	cred2 := *vc.NewCredential("urn:uuid:cred-2", issuerDID, subject)

	p := New(holderDID).WithCredentials(cred1, cred2)

	// Found
	found := p.GetCredentialByID("urn:uuid:cred-1")
	if found == nil {
		t.Error("GetCredentialByID() should find credential")
	}
	if found.ID != "urn:uuid:cred-1" {
		t.Errorf("GetCredentialByID() returned wrong credential: %v", found.ID)
	}

	// Not found
	notFound := p.GetCredentialByID("urn:uuid:nonexistent")
	if notFound != nil {
		t.Error("GetCredentialByID() should return nil for nonexistent ID")
	}
}

func TestPresentation_GetCredentialsByType(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	issuerDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := vc.NewSubject(holderDID)

	cred1 := *vc.NewCredential("urn:uuid:cred-1", issuerDID, subject).
		WithType(vc.TypeGitHubContributorCredential)
	cred2 := *vc.NewCredential("urn:uuid:cred-2", issuerDID, subject).
		WithType(vc.TypeGitHubContributorCredential)
	cred3 := *vc.NewCredential("urn:uuid:cred-3", issuerDID, subject).
		WithType(vc.TypeEducationCredential)

	p := New(holderDID).WithCredentials(cred1, cred2, cred3)

	// Find GitHub credentials
	githubCreds := p.GetCredentialsByType(vc.TypeGitHubContributorCredential)
	if len(githubCreds) != 2 {
		t.Errorf("GetCredentialsByType() returned %d credentials, want 2", len(githubCreds))
	}

	// Find education credentials
	eduCreds := p.GetCredentialsByType(vc.TypeEducationCredential)
	if len(eduCreds) != 1 {
		t.Errorf("GetCredentialsByType() returned %d credentials, want 1", len(eduCreds))
	}

	// Find nonexistent type
	emptyCreds := p.GetCredentialsByType(vc.TypeCertificationCredential)
	if len(emptyCreds) != 0 {
		t.Errorf("GetCredentialsByType() should return empty for nonexistent type")
	}
}

func TestPresentation_HasType(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	p := New(holderDID)

	if !p.HasType(TypeVerifiablePresentation) {
		t.Error("Should have VerifiablePresentation type")
	}

	if p.HasType("NonexistentType") {
		t.Error("Should not have nonexistent type")
	}
}

func TestPresentation_HolderDID(t *testing.T) {
	holderDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	p := New(holderDID)

	if !p.HolderDID().Equals(holderDID) {
		t.Errorf("HolderDID() = %v, want %v", p.HolderDID(), holderDID)
	}
}

func TestNewProof(t *testing.T) {
	proof := NewProof(vc.ProofTypeEd25519Signature2020, "did:key:z6Mk...#key-1")

	if proof.Type != vc.ProofTypeEd25519Signature2020 {
		t.Errorf("Type = %v, want %v", proof.Type, vc.ProofTypeEd25519Signature2020)
	}

	if proof.VerificationMethod != "did:key:z6Mk...#key-1" {
		t.Errorf("VerificationMethod = %v, want did:key:z6Mk...#key-1", proof.VerificationMethod)
	}

	if proof.ProofPurpose != ProofPurposeAuthentication {
		t.Errorf("ProofPurpose = %v, want %v", proof.ProofPurpose, ProofPurposeAuthentication)
	}

	if proof.Created.IsZero() {
		t.Error("Created should be set")
	}
}

func TestProof_Builder(t *testing.T) {
	proof := NewProof(vc.ProofTypeEd25519Signature2020, "did:key:z6Mk...#key-1").
		WithChallenge("challenge-123").
		WithDomain("example.com").
		WithNonce("nonce-456").
		WithProofValue("proof-value-789")

	if proof.Challenge != "challenge-123" {
		t.Errorf("Challenge = %v, want challenge-123", proof.Challenge)
	}

	if proof.Domain != "example.com" {
		t.Errorf("Domain = %v, want example.com", proof.Domain)
	}

	if proof.Nonce != "nonce-456" {
		t.Errorf("Nonce = %v, want nonce-456", proof.Nonce)
	}

	if proof.ProofValue != "proof-value-789" {
		t.Errorf("ProofValue = %v, want proof-value-789", proof.ProofValue)
	}
}

func TestProof_WithJWS(t *testing.T) {
	proof := NewProof(vc.ProofTypeJWT, "did:key:z6Mk...#key-1").
		WithJWS("eyJhbGciOiJFZERTQSJ9..signature")

	if proof.JWS != "eyJhbGciOiJFZERTQSJ9..signature" {
		t.Errorf("JWS = %v, want eyJhbGciOiJFZERTQSJ9..signature", proof.JWS)
	}
}

func TestProof_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Proof
		wantErr bool
	}{
		{
			name: "valid with proof value",
			setup: func() *Proof {
				return NewProof(vc.ProofTypeEd25519Signature2020, "did:key:z6Mk...#key-1").
					WithProofValue("proof-value")
			},
			wantErr: false,
		},
		{
			name: "valid with JWS",
			setup: func() *Proof {
				return NewProof(vc.ProofTypeJWT, "did:key:z6Mk...#key-1").
					WithJWS("jws-signature")
			},
			wantErr: false,
		},
		{
			name: "missing type",
			setup: func() *Proof {
				p := NewProof(vc.ProofTypeEd25519Signature2020, "did:key:z6Mk...#key-1")
				p.Type = ""
				return p.WithProofValue("proof")
			},
			wantErr: true,
		},
		{
			name: "missing created",
			setup: func() *Proof {
				p := NewProof(vc.ProofTypeEd25519Signature2020, "did:key:z6Mk...#key-1")
				p.Created = time.Time{}
				return p.WithProofValue("proof")
			},
			wantErr: true,
		},
		{
			name: "missing verification method",
			setup: func() *Proof {
				p := NewProof(vc.ProofTypeEd25519Signature2020, "")
				return p.WithProofValue("proof")
			},
			wantErr: true,
		},
		{
			name: "missing proof purpose",
			setup: func() *Proof {
				p := NewProof(vc.ProofTypeEd25519Signature2020, "did:key:z6Mk...#key-1")
				p.ProofPurpose = ""
				return p.WithProofValue("proof")
			},
			wantErr: true,
		},
		{
			name: "missing proof value and JWS",
			setup: func() *Proof {
				return NewProof(vc.ProofTypeEd25519Signature2020, "did:key:z6Mk...#key-1")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof := tt.setup()
			err := proof.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestGeneratePresentationID(t *testing.T) {
	id1 := GeneratePresentationID()
	id2 := GeneratePresentationID()

	if id1 == "" {
		t.Error("GeneratePresentationID() returned empty string")
	}

	if !strings.HasPrefix(id1, "urn:uuid:") {
		t.Errorf("GeneratePresentationID() should start with 'urn:uuid:', got %v", id1)
	}

	if id1 == id2 {
		t.Error("GeneratePresentationID() should generate unique IDs")
	}
}

func TestDefaultContext(t *testing.T) {
	ctx := DefaultContext()

	if len(ctx) == 0 {
		t.Error("DefaultContext() should return non-empty slice")
	}

	if ctx[0] != ContextCredentialsV1 {
		t.Errorf("DefaultContext()[0] = %v, want %v", ctx[0], ContextCredentialsV1)
	}
}

func TestPresentationType_String(t *testing.T) {
	pt := TypeVerifiablePresentation

	if pt.String() != "VerifiablePresentation" {
		t.Errorf("String() = %v, want VerifiablePresentation", pt.String())
	}
}
