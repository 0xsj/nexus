package vc

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
)

func TestNewCredential(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject)

	if cred.ID != "urn:uuid:test-123" {
		t.Errorf("ID = %v, want urn:uuid:test-123", cred.ID)
	}

	if !cred.Issuer.ID.Equals(issuerDID) {
		t.Errorf("Issuer.ID = %v, want %v", cred.Issuer.ID, issuerDID)
	}

	if !cred.CredentialSubject.ID.Equals(subjectDID) {
		t.Errorf("CredentialSubject.ID = %v, want %v", cred.CredentialSubject.ID, subjectDID)
	}

	if len(cred.Context) == 0 {
		t.Error("Context should not be empty")
	}

	if len(cred.Type) != 1 || cred.Type[0] != TypeVerifiableCredential {
		t.Error("Type should contain VerifiableCredential")
	}

	if cred.IssuanceDate.IsZero() {
		t.Error("IssuanceDate should be set")
	}
}

func TestCredential_WithType(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithType(TypeGitHubContributorCredential)

	if len(cred.Type) != 2 {
		t.Fatalf("Type count = %v, want 2", len(cred.Type))
	}

	if cred.Type[0] != TypeVerifiableCredential {
		t.Errorf("Type[0] = %v, want %v", cred.Type[0], TypeVerifiableCredential)
	}

	if cred.Type[1] != TypeGitHubContributorCredential {
		t.Errorf("Type[1] = %v, want %v", cred.Type[1], TypeGitHubContributorCredential)
	}
}

func TestCredential_WithTypes(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithTypes(TypeGitHubContributorCredential, TypeSkillCredential)

	if len(cred.Type) != 3 {
		t.Fatalf("Type count = %v, want 3", len(cred.Type))
	}

	if cred.Type[0] != TypeVerifiableCredential {
		t.Errorf("Type[0] = %v, want %v", cred.Type[0], TypeVerifiableCredential)
	}
}

func TestCredential_WithExpirationDate(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	expiration := time.Now().Add(24 * time.Hour)
	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithExpirationDate(expiration)

	if cred.ExpirationDate == nil {
		t.Fatal("ExpirationDate should be set")
	}

	if cred.ExpirationDate.Unix() != expiration.UTC().Unix() {
		t.Errorf("ExpirationDate = %v, want %v", cred.ExpirationDate.Unix(), expiration.UTC().Unix())
	}
}

func TestCredential_WithValidityPeriod(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	from := time.Now()
	until := from.Add(30 * 24 * time.Hour)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithValidityPeriod(from, until)

	if cred.IssuanceDate.Unix() != from.UTC().Unix() {
		t.Errorf("IssuanceDate = %v, want %v", cred.IssuanceDate.Unix(), from.UTC().Unix())
	}

	if cred.ExpirationDate == nil {
		t.Fatal("ExpirationDate should be set")
	}

	if cred.ExpirationDate.Unix() != until.UTC().Unix() {
		t.Errorf("ExpirationDate = %v, want %v", cred.ExpirationDate.Unix(), until.UTC().Unix())
	}
}

func TestCredential_WithIssuerName(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithIssuerName("Proof Platform")

	if cred.Issuer.Name != "Proof Platform" {
		t.Errorf("Issuer.Name = %v, want Proof Platform", cred.Issuer.Name)
	}
}

func TestCredential_WithIssuerURL(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithIssuerURL("https://proof.example.com")

	if cred.Issuer.URL != "https://proof.example.com" {
		t.Errorf("Issuer.URL = %v, want https://proof.example.com", cred.Issuer.URL)
	}
}

func TestCredential_WithStatus(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	status := &CredentialStatus{
		ID:                   "https://example.com/status/1",
		Type:                 "StatusList2021Entry",
		StatusListIndex:      "0",
		StatusListCredential: "https://example.com/status-list",
		StatusPurpose:        "revocation",
	}

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithStatus(status)

	if cred.CredentialStatus == nil {
		t.Fatal("CredentialStatus should be set")
	}

	if cred.CredentialStatus.ID != status.ID {
		t.Errorf("CredentialStatus.ID = %v, want %v", cred.CredentialStatus.ID, status.ID)
	}
}

func TestCredential_WithSchema(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	schema := &CredentialSchema{
		ID:   "https://example.com/schemas/github-contributor",
		Type: "JsonSchema",
	}

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithSchema(schema)

	if cred.CredentialSchema == nil {
		t.Fatal("CredentialSchema should be set")
	}

	if cred.CredentialSchema.ID != schema.ID {
		t.Errorf("CredentialSchema.ID = %v, want %v", cred.CredentialSchema.ID, schema.ID)
	}
}

func TestCredential_WithEvidence(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	evidence := Evidence{
		ID:               "https://example.com/evidence/1",
		Type:             []string{"DocumentVerification"},
		Verifier:         "https://proof.example.com",
		EvidenceDocument: "GitHubAPIResponse",
	}

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithEvidence(evidence)

	if len(cred.Evidence) != 1 {
		t.Fatalf("Evidence count = %v, want 1", len(cred.Evidence))
	}

	if cred.Evidence[0].ID != evidence.ID {
		t.Errorf("Evidence[0].ID = %v, want %v", cred.Evidence[0].ID, evidence.ID)
	}
}

func TestCredential_WithContext(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithContext("https://example.com/custom-context")

	if len(cred.Context) != 3 {
		t.Fatalf("Context count = %v, want 3", len(cred.Context))
	}

	if cred.Context[2] != "https://example.com/custom-context" {
		t.Errorf("Context[2] = %v, want https://example.com/custom-context", cred.Context[2])
	}
}

func TestCredential_Validate(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	// Valid credential
	cred := NewCredential("urn:uuid:test-123", issuerDID, subject)
	if err := cred.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestCredential_Validate_MissingContext(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject)
	cred.Context = nil

	if err := cred.Validate(); err == nil {
		t.Error("Validate() expected error for missing context")
	}
}

func TestCredential_Validate_MissingType(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject)
	cred.Type = nil

	if err := cred.Validate(); err == nil {
		t.Error("Validate() expected error for missing type")
	}
}

func TestCredential_Validate_MissingVerifiableCredentialType(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject)
	cred.Type = []CredentialType{TypeGitHubContributorCredential} // Missing VerifiableCredential

	if err := cred.Validate(); err == nil {
		t.Error("Validate() expected error for missing VerifiableCredential type")
	}
}

func TestCredential_Validate_MissingIssuer(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := &Credential{
		Context:           DefaultContext(),
		Type:              []CredentialType{TypeVerifiableCredential},
		IssuanceDate:      time.Now(),
		CredentialSubject: subject,
	}

	if err := cred.Validate(); err == nil {
		t.Error("Validate() expected error for missing issuer")
	}
}

func TestCredential_Validate_MissingSubject(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	cred := &Credential{
		Context:      DefaultContext(),
		Type:         []CredentialType{TypeVerifiableCredential},
		Issuer:       Issuer{ID: issuerDID},
		IssuanceDate: time.Now(),
	}

	if err := cred.Validate(); err == nil {
		t.Error("Validate() expected error for missing subject")
	}
}

func TestCredential_IsExpired(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	// No expiration
	credNoExp := NewCredential("urn:uuid:test-123", issuerDID, subject)
	if credNoExp.IsExpired() {
		t.Error("IsExpired() should be false when no expiration set")
	}

	// Future expiration
	credFuture := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithExpirationDate(time.Now().Add(24 * time.Hour))
	if credFuture.IsExpired() {
		t.Error("IsExpired() should be false for future expiration")
	}

	// Past expiration
	credPast := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithExpirationDate(time.Now().Add(-24 * time.Hour))
	if !credPast.IsExpired() {
		t.Error("IsExpired() should be true for past expiration")
	}
}

func TestCredential_IsActive(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	// Active credential
	credActive := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithExpirationDate(time.Now().Add(24 * time.Hour))
	if !credActive.IsActive() {
		t.Error("IsActive() should be true for active credential")
	}

	// Expired credential
	credExpired := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithExpirationDate(time.Now().Add(-24 * time.Hour))
	if credExpired.IsActive() {
		t.Error("IsActive() should be false for expired credential")
	}

	// Not yet valid
	credFuture := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithIssuanceDate(time.Now().Add(24 * time.Hour))
	if credFuture.IsActive() {
		t.Error("IsActive() should be false for not-yet-valid credential")
	}
}

func TestCredential_SubjectDID(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject)

	if !cred.SubjectDID().Equals(subjectDID) {
		t.Errorf("SubjectDID() = %v, want %v", cred.SubjectDID(), subjectDID)
	}
}

func TestCredential_IssuerDID(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject)

	if !cred.IssuerDID().Equals(issuerDID) {
		t.Errorf("IssuerDID() = %v, want %v", cred.IssuerDID(), issuerDID)
	}
}

func TestCredential_HasType(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	subject := NewSubject(subjectDID)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithType(TypeGitHubContributorCredential)

	if !cred.HasType(TypeVerifiableCredential) {
		t.Error("HasType(VerifiableCredential) should be true")
	}

	if !cred.HasType(TypeGitHubContributorCredential) {
		t.Error("HasType(GitHubContributorCredential) should be true")
	}

	if cred.HasType(TypeEducationCredential) {
		t.Error("HasType(EducationCredential) should be false")
	}
}

func TestIssuer_JSONMarshal_IDOnly(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	issuer := Issuer{ID: issuerDID}

	data, err := json.Marshal(issuer)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Should be a simple string
	expected := `"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"`
	if string(data) != expected {
		t.Errorf("Marshal() = %v, want %v", string(data), expected)
	}
}

func TestIssuer_JSONMarshal_WithName(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	issuer := Issuer{
		ID:   issuerDID,
		Name: "Proof Platform",
	}

	data, err := json.Marshal(issuer)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Should be an object
	if data[0] != '{' {
		t.Errorf("Marshal() should be object, got: %v", string(data))
	}
}

func TestIssuer_JSONUnmarshal_String(t *testing.T) {
	input := `"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"`

	var issuer Issuer
	if err := json.Unmarshal([]byte(input), &issuer); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	expectedDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	if !issuer.ID.Equals(expectedDID) {
		t.Errorf("ID = %v, want %v", issuer.ID, expectedDID)
	}
}

func TestIssuer_JSONUnmarshal_Object(t *testing.T) {
	input := `{"id":"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK","name":"Proof Platform","url":"https://proof.example.com"}`

	var issuer Issuer
	if err := json.Unmarshal([]byte(input), &issuer); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	expectedDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	if !issuer.ID.Equals(expectedDID) {
		t.Errorf("ID = %v, want %v", issuer.ID, expectedDID)
	}

	if issuer.Name != "Proof Platform" {
		t.Errorf("Name = %v, want Proof Platform", issuer.Name)
	}

	if issuer.URL != "https://proof.example.com" {
		t.Errorf("URL = %v, want https://proof.example.com", issuer.URL)
	}
}

func TestGenerateCredentialID(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	id1 := GenerateCredentialID(issuerDID)
	id2 := GenerateCredentialID(issuerDID)

	if id1 == id2 {
		t.Error("GenerateCredentialID() should produce unique IDs")
	}

	if len(id1) < 10 {
		t.Errorf("GenerateCredentialID() returned short ID: %v", id1)
	}

	// Should start with urn:uuid:
	if id1[:9] != "urn:uuid:" {
		t.Errorf("GenerateCredentialID() should start with 'urn:uuid:', got: %v", id1[:9])
	}
}

func TestGenerateCredentialIDWithPrefix(t *testing.T) {
	id := GenerateCredentialIDWithPrefix("https://proof.example.com/credentials/")

	if len(id) < 40 {
		t.Errorf("GenerateCredentialIDWithPrefix() returned short ID: %v", id)
	}

	// Should start with prefix
	prefix := "https://proof.example.com/credentials/"
	if id[:len(prefix)] != prefix {
		t.Errorf("GenerateCredentialIDWithPrefix() should start with prefix, got: %v", id)
	}
}

func TestCredential_JSONRoundTrip(t *testing.T) {
	issuerDID := did.MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("username", "testuser").
		WithClaim("commits", 500)

	cred := NewCredential("urn:uuid:test-123", issuerDID, subject).
		WithType(TypeGitHubContributorCredential).
		WithIssuerName("Proof Platform").
		WithExpirationDate(time.Now().Add(24 * time.Hour))

	// Marshal
	data, err := json.Marshal(cred)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Unmarshal
	var decoded Credential
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// Verify
	if decoded.ID != cred.ID {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, cred.ID)
	}

	if !decoded.Issuer.ID.Equals(cred.Issuer.ID) {
		t.Errorf("Issuer.ID mismatch: got %v, want %v", decoded.Issuer.ID, cred.Issuer.ID)
	}

	if !decoded.CredentialSubject.ID.Equals(cred.CredentialSubject.ID) {
		t.Errorf("Subject.ID mismatch: got %v, want %v", decoded.CredentialSubject.ID, cred.CredentialSubject.ID)
	}

	if len(decoded.Type) != len(cred.Type) {
		t.Errorf("Type count mismatch: got %v, want %v", len(decoded.Type), len(cred.Type))
	}
}
