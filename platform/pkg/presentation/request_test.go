package presentation

import (
	"strings"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

func TestNewRequest(t *testing.T) {
	challenge := "test-challenge-123"

	r := NewRequest(challenge)

	if r == nil {
		t.Fatal("NewRequest() returned nil")
	}

	if r.ID == "" {
		t.Error("ID should be generated")
	}

	if r.Challenge != challenge {
		t.Errorf("Challenge = %v, want %v", r.Challenge, challenge)
	}

	if r.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewRequestWithChallenge(t *testing.T) {
	r := NewRequestWithChallenge()

	if r == nil {
		t.Fatal("NewRequestWithChallenge() returned nil")
	}

	if r.Challenge == "" {
		t.Error("Challenge should be generated")
	}

	// Ensure unique challenges
	r2 := NewRequestWithChallenge()
	if r.Challenge == r2.Challenge {
		t.Error("NewRequestWithChallenge() should generate unique challenges")
	}
}

func TestRequest_WithID(t *testing.T) {
	r := NewRequest("challenge").
		WithID("custom-request-id")

	if r.ID != "custom-request-id" {
		t.Errorf("ID = %v, want custom-request-id", r.ID)
	}
}

func TestRequest_WithVerifier(t *testing.T) {
	verifierDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	r := NewRequest("challenge").
		WithVerifier(verifierDID)

	if !r.Verifier.Equals(verifierDID) {
		t.Errorf("Verifier = %v, want %v", r.Verifier, verifierDID)
	}
}

func TestRequest_WithDomain(t *testing.T) {
	r := NewRequest("challenge").
		WithDomain("example.com")

	if r.Domain != "example.com" {
		t.Errorf("Domain = %v, want example.com", r.Domain)
	}
}

func TestRequest_WithPurpose(t *testing.T) {
	r := NewRequest("challenge").
		WithPurpose("Verify employment history")

	if r.Purpose != "Verify employment history" {
		t.Errorf("Purpose = %v, want 'Verify employment history'", r.Purpose)
	}
}

func TestRequest_WithExpiration(t *testing.T) {
	expiry := time.Now().Add(1 * time.Hour)

	r := NewRequest("challenge").
		WithExpiration(expiry)

	if !r.ExpiresAt.Equal(expiry) {
		t.Errorf("ExpiresAt = %v, want %v", r.ExpiresAt, expiry)
	}
}

func TestRequest_WithExpiresIn(t *testing.T) {
	before := time.Now()
	r := NewRequest("challenge").
		WithExpiresIn(1 * time.Hour)
	after := time.Now()

	expectedMin := before.Add(1 * time.Hour)
	expectedMax := after.Add(1 * time.Hour)

	if r.ExpiresAt.Before(expectedMin) || r.ExpiresAt.After(expectedMax) {
		t.Errorf("ExpiresAt = %v, expected between %v and %v", r.ExpiresAt, expectedMin, expectedMax)
	}
}

func TestRequest_WithCallbackURL(t *testing.T) {
	r := NewRequest("challenge").
		WithCallbackURL("https://example.com/callback")

	if r.CallbackURL != "https://example.com/callback" {
		t.Errorf("CallbackURL = %v, want https://example.com/callback", r.CallbackURL)
	}
}

func TestRequest_WithState(t *testing.T) {
	r := NewRequest("challenge").
		WithState("state-abc-123")

	if r.State != "state-abc-123" {
		t.Errorf("State = %v, want state-abc-123", r.State)
	}
}

func TestRequest_RequestCredential(t *testing.T) {
	r := NewRequest("challenge").
		RequestCredential(vc.TypeGitHubContributorCredential)

	if len(r.RequestedCredentials) != 1 {
		t.Fatalf("RequestedCredentials count = %d, want 1", len(r.RequestedCredentials))
	}

	cr := r.RequestedCredentials[0]
	if cr.Type != vc.TypeGitHubContributorCredential {
		t.Errorf("Type = %v, want %v", cr.Type, vc.TypeGitHubContributorCredential)
	}

	if !cr.Required {
		t.Error("Required should be true")
	}
}

func TestRequest_RequestOptionalCredential(t *testing.T) {
	r := NewRequest("challenge").
		RequestOptionalCredential(vc.TypeEducationCredential)

	if len(r.RequestedCredentials) != 1 {
		t.Fatalf("RequestedCredentials count = %d, want 1", len(r.RequestedCredentials))
	}

	cr := r.RequestedCredentials[0]
	if cr.Type != vc.TypeEducationCredential {
		t.Errorf("Type = %v, want %v", cr.Type, vc.TypeEducationCredential)
	}

	if cr.Required {
		t.Error("Required should be false")
	}
}

func TestRequest_RequestCredentialFromIssuers(t *testing.T) {
	issuer1 := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	issuer2 := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	r := NewRequest("challenge").
		RequestCredentialFromIssuers(vc.TypeGitHubContributorCredential, issuer1, issuer2)

	if len(r.RequestedCredentials) != 1 {
		t.Fatalf("RequestedCredentials count = %d, want 1", len(r.RequestedCredentials))
	}

	cr := r.RequestedCredentials[0]
	if len(cr.TrustedIssuers) != 2 {
		t.Errorf("TrustedIssuers count = %d, want 2", len(cr.TrustedIssuers))
	}

	if !cr.Required {
		t.Error("Required should be true")
	}
}

func TestRequest_WithCredentialRequest(t *testing.T) {
	cr := CredentialRequest{
		Type:            vc.TypeCertificationCredential,
		Required:        true,
		RequestedFields: []string{"name", "issueDate"},
	}

	r := NewRequest("challenge").
		WithCredentialRequest(cr)

	if len(r.RequestedCredentials) != 1 {
		t.Fatalf("RequestedCredentials count = %d, want 1", len(r.RequestedCredentials))
	}

	if len(r.RequestedCredentials[0].RequestedFields) != 2 {
		t.Errorf("RequestedFields count = %d, want 2", len(r.RequestedCredentials[0].RequestedFields))
	}
}

func TestRequest_MultipleCredentials(t *testing.T) {
	r := NewRequest("challenge").
		RequestCredential(vc.TypeGitHubContributorCredential).
		RequestCredential(vc.TypeLinkedInEmploymentCredential).
		RequestOptionalCredential(vc.TypeEducationCredential)

	if len(r.RequestedCredentials) != 3 {
		t.Errorf("RequestedCredentials count = %d, want 3", len(r.RequestedCredentials))
	}
}

func TestRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Request
		wantErr bool
	}{
		{
			name: "valid request",
			setup: func() *Request {
				return NewRequest("challenge-123")
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			setup: func() *Request {
				r := NewRequest("challenge-123")
				r.ID = ""
				return r
			},
			wantErr: true,
		},
		{
			name: "missing challenge",
			setup: func() *Request {
				r := NewRequest("")
				return r
			},
			wantErr: true,
		},
		{
			name: "expired request",
			setup: func() *Request {
				r := NewRequest("challenge-123")
				r.ExpiresAt = time.Now().Add(-1 * time.Hour)
				return r
			},
			wantErr: true,
		},
		{
			name: "valid with future expiration",
			setup: func() *Request {
				return NewRequest("challenge-123").
					WithExpiresIn(1 * time.Hour)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setup()
			err := r.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestRequest_IsExpired(t *testing.T) {
	// Not expired - no expiration set
	r1 := NewRequest("challenge")
	if r1.IsExpired() {
		t.Error("Request without expiration should not be expired")
	}

	// Not expired - future expiration
	r2 := NewRequest("challenge").WithExpiresIn(1 * time.Hour)
	if r2.IsExpired() {
		t.Error("Request with future expiration should not be expired")
	}

	// Expired
	r3 := NewRequest("challenge")
	r3.ExpiresAt = time.Now().Add(-1 * time.Hour)
	if !r3.IsExpired() {
		t.Error("Request with past expiration should be expired")
	}
}

func TestRequest_RequiredCredentialTypes(t *testing.T) {
	r := NewRequest("challenge").
		RequestCredential(vc.TypeGitHubContributorCredential).
		RequestCredential(vc.TypeLinkedInEmploymentCredential).
		RequestOptionalCredential(vc.TypeEducationCredential)

	required := r.RequiredCredentialTypes()

	if len(required) != 2 {
		t.Errorf("RequiredCredentialTypes() count = %d, want 2", len(required))
	}

	// Check types are present
	hasGitHub := false
	hasLinkedIn := false
	for _, ct := range required {
		if ct == vc.TypeGitHubContributorCredential {
			hasGitHub = true
		}
		if ct == vc.TypeLinkedInEmploymentCredential {
			hasLinkedIn = true
		}
	}

	if !hasGitHub {
		t.Error("RequiredCredentialTypes() should include GitHubContributorCredential")
	}
	if !hasLinkedIn {
		t.Error("RequiredCredentialTypes() should include LinkedInEmploymentCredential")
	}
}

func TestRequest_OptionalCredentialTypes(t *testing.T) {
	r := NewRequest("challenge").
		RequestCredential(vc.TypeGitHubContributorCredential).
		RequestOptionalCredential(vc.TypeEducationCredential).
		RequestOptionalCredential(vc.TypeCertificationCredential)

	optional := r.OptionalCredentialTypes()

	if len(optional) != 2 {
		t.Errorf("OptionalCredentialTypes() count = %d, want 2", len(optional))
	}
}

func TestRequest_AllCredentialTypes(t *testing.T) {
	r := NewRequest("challenge").
		RequestCredential(vc.TypeGitHubContributorCredential).
		RequestOptionalCredential(vc.TypeEducationCredential)

	all := r.AllCredentialTypes()

	if len(all) != 2 {
		t.Errorf("AllCredentialTypes() count = %d, want 2", len(all))
	}
}

func TestRequest_GetCredentialRequest(t *testing.T) {
	r := NewRequest("challenge").
		RequestCredential(vc.TypeGitHubContributorCredential).
		RequestOptionalCredential(vc.TypeEducationCredential)

	// Found
	cr := r.GetCredentialRequest(vc.TypeGitHubContributorCredential)
	if cr == nil {
		t.Error("GetCredentialRequest() should find request")
	}
	if cr.Type != vc.TypeGitHubContributorCredential {
		t.Errorf("Type = %v, want %v", cr.Type, vc.TypeGitHubContributorCredential)
	}

	// Not found
	notFound := r.GetCredentialRequest(vc.TypeCertificationCredential)
	if notFound != nil {
		t.Error("GetCredentialRequest() should return nil for nonexistent type")
	}
}

func TestRequest_IsTrustedIssuer(t *testing.T) {
	trustedIssuer := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	untrustedIssuer := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	r := NewRequest("challenge").
		RequestCredentialFromIssuers(vc.TypeGitHubContributorCredential, trustedIssuer).
		RequestCredential(vc.TypeEducationCredential) // Any issuer

	// Trusted issuer for GitHub
	if !r.IsTrustedIssuer(vc.TypeGitHubContributorCredential, trustedIssuer) {
		t.Error("IsTrustedIssuer() should return true for trusted issuer")
	}

	// Untrusted issuer for GitHub
	if r.IsTrustedIssuer(vc.TypeGitHubContributorCredential, untrustedIssuer) {
		t.Error("IsTrustedIssuer() should return false for untrusted issuer")
	}

	// Any issuer for Education (empty trusted list)
	if !r.IsTrustedIssuer(vc.TypeEducationCredential, untrustedIssuer) {
		t.Error("IsTrustedIssuer() should return true when trusted list is empty")
	}

	// Nonexistent credential type
	if r.IsTrustedIssuer(vc.TypeCertificationCredential, trustedIssuer) {
		t.Error("IsTrustedIssuer() should return false for nonexistent credential type")
	}
}

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("GenerateRequestID() returned empty string")
	}

	if !strings.HasPrefix(id1, "urn:uuid:") {
		t.Errorf("GenerateRequestID() should start with 'urn:uuid:', got %v", id1)
	}

	if id1 == id2 {
		t.Error("GenerateRequestID() should generate unique IDs")
	}
}

func TestGenerateChallenge(t *testing.T) {
	c1 := GenerateChallenge()
	c2 := GenerateChallenge()

	if c1 == "" {
		t.Error("GenerateChallenge() returned empty string")
	}

	if c1 == c2 {
		t.Error("GenerateChallenge() should generate unique challenges")
	}
}

func TestCredentialConstraints(t *testing.T) {
	minDate := time.Now().Add(-30 * 24 * time.Hour)
	maxAge := 24 * time.Hour

	constraints := &CredentialConstraints{
		MinIssuanceDate:   &minDate,
		MaxAge:            &maxAge,
		RequireNotExpired: true,
		RequireNotRevoked: true,
		FieldConstraints: []FieldConstraint{
			{
				Field:    "commits",
				Operator: OpGreaterEqual,
				Value:    100,
			},
		},
	}

	cr := CredentialRequest{
		Type:        vc.TypeGitHubContributorCredential,
		Required:    true,
		Constraints: constraints,
	}

	r := NewRequest("challenge").
		WithCredentialRequest(cr)

	if r.RequestedCredentials[0].Constraints == nil {
		t.Fatal("Constraints should not be nil")
	}

	if !r.RequestedCredentials[0].Constraints.RequireNotExpired {
		t.Error("RequireNotExpired should be true")
	}

	if len(r.RequestedCredentials[0].Constraints.FieldConstraints) != 1 {
		t.Error("FieldConstraints count should be 1")
	}
}

func TestConstraintOperator(t *testing.T) {
	tests := []struct {
		op   ConstraintOperator
		want string
	}{
		{OpEquals, "eq"},
		{OpNotEquals, "neq"},
		{OpGreaterThan, "gt"},
		{OpGreaterEqual, "gte"},
		{OpLessThan, "lt"},
		{OpLessEqual, "lte"},
		{OpContains, "contains"},
		{OpExists, "exists"},
	}

	for _, tt := range tests {
		t.Run(string(tt.op), func(t *testing.T) {
			if string(tt.op) != tt.want {
				t.Errorf("ConstraintOperator = %v, want %v", string(tt.op), tt.want)
			}
		})
	}
}
