package vc

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
)

func TestNewSubject(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID)

	if !subject.ID.Equals(subjectDID) {
		t.Errorf("ID = %v, want %v", subject.ID, subjectDID)
	}

	if subject.Claims == nil {
		t.Error("Claims should be initialized")
	}

	if len(subject.Claims) != 0 {
		t.Errorf("Claims should be empty, got %v", len(subject.Claims))
	}
}

func TestSubject_WithClaim(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("username", "testuser").
		WithClaim("commits", 500)

	if len(subject.Claims) != 2 {
		t.Errorf("Claims count = %v, want 2", len(subject.Claims))
	}

	if subject.Claims["username"] != "testuser" {
		t.Errorf("Claims[username] = %v, want testuser", subject.Claims["username"])
	}

	if subject.Claims["commits"] != 500 {
		t.Errorf("Claims[commits] = %v, want 500", subject.Claims["commits"])
	}
}

func TestSubject_WithClaims(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	claims := map[string]interface{}{
		"username": "testuser",
		"commits":  500,
		"verified": true,
	}

	subject := NewSubject(subjectDID).WithClaims(claims)

	if len(subject.Claims) != 3 {
		t.Errorf("Claims count = %v, want 3", len(subject.Claims))
	}

	if subject.Claims["username"] != "testuser" {
		t.Errorf("Claims[username] = %v, want testuser", subject.Claims["username"])
	}
}

func TestSubject_GetClaim(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("username", "testuser")

	// Existing claim
	val, ok := subject.GetClaim("username")
	if !ok {
		t.Error("GetClaim() should return true for existing claim")
	}
	if val != "testuser" {
		t.Errorf("GetClaim() = %v, want testuser", val)
	}

	// Non-existing claim
	_, ok = subject.GetClaim("nonexistent")
	if ok {
		t.Error("GetClaim() should return false for non-existing claim")
	}
}

func TestSubject_GetStringClaim(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("username", "testuser").
		WithClaim("commits", 500)

	// String claim
	val, ok := subject.GetStringClaim("username")
	if !ok {
		t.Error("GetStringClaim() should return true for string claim")
	}
	if val != "testuser" {
		t.Errorf("GetStringClaim() = %v, want testuser", val)
	}

	// Non-string claim
	_, ok = subject.GetStringClaim("commits")
	if ok {
		t.Error("GetStringClaim() should return false for non-string claim")
	}

	// Non-existing claim
	_, ok = subject.GetStringClaim("nonexistent")
	if ok {
		t.Error("GetStringClaim() should return false for non-existing claim")
	}
}

func TestSubject_GetIntClaim(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("commits", 500).
		WithClaim("commits_int64", int64(1000)).
		WithClaim("commits_float", 1500.0).
		WithClaim("username", "testuser")

	// Int claim
	val, ok := subject.GetIntClaim("commits")
	if !ok {
		t.Error("GetIntClaim() should return true for int claim")
	}
	if val != 500 {
		t.Errorf("GetIntClaim() = %v, want 500", val)
	}

	// Int64 claim
	val, ok = subject.GetIntClaim("commits_int64")
	if !ok {
		t.Error("GetIntClaim() should return true for int64 claim")
	}
	if val != 1000 {
		t.Errorf("GetIntClaim() = %v, want 1000", val)
	}

	// Float64 claim (JSON numbers unmarshal as float64)
	val, ok = subject.GetIntClaim("commits_float")
	if !ok {
		t.Error("GetIntClaim() should return true for float64 claim")
	}
	if val != 1500 {
		t.Errorf("GetIntClaim() = %v, want 1500", val)
	}

	// Non-int claim
	_, ok = subject.GetIntClaim("username")
	if ok {
		t.Error("GetIntClaim() should return false for non-int claim")
	}
}

func TestSubject_GetBoolClaim(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("verified", true).
		WithClaim("premium", false).
		WithClaim("username", "testuser")

	// True claim
	val, ok := subject.GetBoolClaim("verified")
	if !ok {
		t.Error("GetBoolClaim() should return true for bool claim")
	}
	if val != true {
		t.Errorf("GetBoolClaim() = %v, want true", val)
	}

	// False claim
	val, ok = subject.GetBoolClaim("premium")
	if !ok {
		t.Error("GetBoolClaim() should return true for bool claim")
	}
	if val != false {
		t.Errorf("GetBoolClaim() = %v, want false", val)
	}

	// Non-bool claim
	_, ok = subject.GetBoolClaim("username")
	if ok {
		t.Error("GetBoolClaim() should return false for non-bool claim")
	}
}

func TestSubject_GetTimeClaim(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	now := time.Now().UTC().Truncate(time.Second)
	nowStr := now.Format(time.RFC3339)

	subject := NewSubject(subjectDID).
		WithClaim("verifiedAt", now).
		WithClaim("verifiedAtStr", nowStr).
		WithClaim("username", "testuser")

	// Time claim
	val, ok := subject.GetTimeClaim("verifiedAt")
	if !ok {
		t.Error("GetTimeClaim() should return true for time claim")
	}
	if !val.Equal(now) {
		t.Errorf("GetTimeClaim() = %v, want %v", val, now)
	}

	// String time claim
	val, ok = subject.GetTimeClaim("verifiedAtStr")
	if !ok {
		t.Error("GetTimeClaim() should return true for string time claim")
	}
	if !val.Equal(now) {
		t.Errorf("GetTimeClaim() = %v, want %v", val, now)
	}

	// Non-time claim
	_, ok = subject.GetTimeClaim("username")
	if ok {
		t.Error("GetTimeClaim() should return false for non-time claim")
	}
}

func TestSubject_HasClaim(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("username", "testuser")

	if !subject.HasClaim("username") {
		t.Error("HasClaim() should return true for existing claim")
	}

	if subject.HasClaim("nonexistent") {
		t.Error("HasClaim() should return false for non-existing claim")
	}
}

func TestSubject_ClaimKeys(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("username", "testuser").
		WithClaim("commits", 500).
		WithClaim("verified", true)

	keys := subject.ClaimKeys()

	if len(keys) != 3 {
		t.Errorf("ClaimKeys() count = %v, want 3", len(keys))
	}

	// Check all keys are present (order not guaranteed)
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	if !keyMap["username"] || !keyMap["commits"] || !keyMap["verified"] {
		t.Errorf("ClaimKeys() missing expected keys: %v", keys)
	}
}

func TestSubject_ClaimKeys_Empty(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID)
	keys := subject.ClaimKeys()

	if len(keys) != 0 {
		t.Errorf("ClaimKeys() should be empty, got %v", keys)
	}
}

func TestSubject_JSONMarshal(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	subject := NewSubject(subjectDID).
		WithClaim("username", "testuser").
		WithClaim("commits", 500)

	data, err := json.Marshal(subject)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Should be an object with id and claims flattened
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if m["id"] != subjectDID.String() {
		t.Errorf("id = %v, want %v", m["id"], subjectDID.String())
	}

	if m["username"] != "testuser" {
		t.Errorf("username = %v, want testuser", m["username"])
	}

	// JSON numbers are float64
	if m["commits"] != float64(500) {
		t.Errorf("commits = %v, want 500", m["commits"])
	}
}

func TestSubject_JSONUnmarshal(t *testing.T) {
	input := `{"id":"did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy","username":"testuser","commits":500}`

	var subject Subject
	if err := json.Unmarshal([]byte(input), &subject); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	expectedDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")
	if !subject.ID.Equals(expectedDID) {
		t.Errorf("ID = %v, want %v", subject.ID, expectedDID)
	}

	if subject.Claims["username"] != "testuser" {
		t.Errorf("Claims[username] = %v, want testuser", subject.Claims["username"])
	}

	// JSON numbers are float64
	if subject.Claims["commits"] != float64(500) {
		t.Errorf("Claims[commits] = %v, want 500", subject.Claims["commits"])
	}

	// ID should not be in claims
	if _, ok := subject.Claims["id"]; ok {
		t.Error("Claims should not contain 'id'")
	}
}

func TestSubject_JSONRoundTrip(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	original := NewSubject(subjectDID).
		WithClaim("username", "testuser").
		WithClaim("commits", 500).
		WithClaim("verified", true)

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Unmarshal
	var decoded Subject
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// Verify
	if !decoded.ID.Equals(original.ID) {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, original.ID)
	}

	if len(decoded.Claims) != len(original.Claims) {
		t.Errorf("Claims count mismatch: got %v, want %v", len(decoded.Claims), len(original.Claims))
	}
}

// ============================================================================
// Pre-defined Claims Tests
// ============================================================================

func TestGitHubClaims_ToSubject(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	claims := GitHubClaims{
		Username:      "testuser",
		Commits:       500,
		Repositories:  10,
		Stars:         100,
		Followers:     50,
		Contributions: 1000,
		VerifiedAt:    time.Now().UTC(),
	}

	subject := claims.ToSubject(subjectDID)

	if !subject.ID.Equals(subjectDID) {
		t.Errorf("ID = %v, want %v", subject.ID, subjectDID)
	}

	username, ok := subject.GetStringClaim("username")
	if !ok || username != "testuser" {
		t.Errorf("username = %v, want testuser", username)
	}

	commits, ok := subject.GetIntClaim("commits")
	if !ok || commits != 500 {
		t.Errorf("commits = %v, want 500", commits)
	}
}

func TestLinkedInClaims_ToSubject(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	claims := LinkedInClaims{
		Name:        "Test User",
		Company:     "Acme Inc",
		Title:       "Software Engineer",
		StartDate:   "2020-01-01",
		Connections: 500,
		VerifiedAt:  time.Now().UTC(),
	}

	subject := claims.ToSubject(subjectDID)

	if !subject.ID.Equals(subjectDID) {
		t.Errorf("ID = %v, want %v", subject.ID, subjectDID)
	}

	name, ok := subject.GetStringClaim("name")
	if !ok || name != "Test User" {
		t.Errorf("name = %v, want Test User", name)
	}

	company, ok := subject.GetStringClaim("company")
	if !ok || company != "Acme Inc" {
		t.Errorf("company = %v, want Acme Inc", company)
	}
}

func TestLinkedInClaims_ToSubject_OptionalFields(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	// Only required fields
	claims := LinkedInClaims{
		Name:       "Test User",
		VerifiedAt: time.Now().UTC(),
	}

	subject := claims.ToSubject(subjectDID)

	// Optional fields should not be present
	if subject.HasClaim("company") {
		t.Error("company should not be present when empty")
	}

	if subject.HasClaim("title") {
		t.Error("title should not be present when empty")
	}

	if subject.HasClaim("connections") {
		t.Error("connections should not be present when zero")
	}
}

func TestEducationClaims_ToSubject(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	claims := EducationClaims{
		Institution:    "MIT",
		Degree:         "Bachelor of Science",
		FieldOfStudy:   "Computer Science",
		GraduationDate: "2020-05-15",
		Grade:          "3.8",
		VerifiedAt:     time.Now().UTC(),
	}

	subject := claims.ToSubject(subjectDID)

	institution, ok := subject.GetStringClaim("institution")
	if !ok || institution != "MIT" {
		t.Errorf("institution = %v, want MIT", institution)
	}

	degree, ok := subject.GetStringClaim("degree")
	if !ok || degree != "Bachelor of Science" {
		t.Errorf("degree = %v, want Bachelor of Science", degree)
	}
}

func TestCertificationClaims_ToSubject(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	claims := CertificationClaims{
		Name:           "AWS Solutions Architect",
		Issuer:         "Amazon Web Services",
		IssueDate:      "2023-01-15",
		ExpirationDate: "2026-01-15",
		CredentialID:   "AWS-123456",
		CredentialURL:  "https://aws.amazon.com/verify/AWS-123456",
		VerifiedAt:     time.Now().UTC(),
	}

	subject := claims.ToSubject(subjectDID)

	name, ok := subject.GetStringClaim("name")
	if !ok || name != "AWS Solutions Architect" {
		t.Errorf("name = %v, want AWS Solutions Architect", name)
	}

	issuer, ok := subject.GetStringClaim("issuer")
	if !ok || issuer != "Amazon Web Services" {
		t.Errorf("issuer = %v, want Amazon Web Services", issuer)
	}

	credentialID, ok := subject.GetStringClaim("credentialId")
	if !ok || credentialID != "AWS-123456" {
		t.Errorf("credentialId = %v, want AWS-123456", credentialID)
	}
}

func TestSkillClaims_ToSubject(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	claims := SkillClaims{
		Name:         "Go Programming",
		Level:        "Expert",
		Endorsements: 25,
		VerifiedAt:   time.Now().UTC(),
	}

	subject := claims.ToSubject(subjectDID)

	name, ok := subject.GetStringClaim("name")
	if !ok || name != "Go Programming" {
		t.Errorf("name = %v, want Go Programming", name)
	}

	level, ok := subject.GetStringClaim("level")
	if !ok || level != "Expert" {
		t.Errorf("level = %v, want Expert", level)
	}

	endorsements, ok := subject.GetIntClaim("endorsements")
	if !ok || endorsements != 25 {
		t.Errorf("endorsements = %v, want 25", endorsements)
	}
}

func TestSkillClaims_ToSubject_OptionalFields(t *testing.T) {
	subjectDID := did.MustParse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	// Only required fields
	claims := SkillClaims{
		Name:       "Go Programming",
		VerifiedAt: time.Now().UTC(),
	}

	subject := claims.ToSubject(subjectDID)

	// Optional fields should not be present
	if subject.HasClaim("level") {
		t.Error("level should not be present when empty")
	}

	if subject.HasClaim("endorsements") {
		t.Error("endorsements should not be present when zero")
	}
}
