package vc

import (
	"encoding/json"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
)

// Subject represents the credential subject (the entity the claims are about).
type Subject struct {
	// ID is the DID of the subject.
	ID did.DID `json:"id"`

	// Claims contains the claims about the subject.
	// Stored as a map for flexibility.
	Claims map[string]interface{} `json:"-"`
}

// MarshalJSON implements json.Marshaler.
// Flattens ID and claims into a single object.
func (s Subject) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})

	// Add ID
	if !s.ID.IsZero() {
		m["id"] = s.ID.String()
	}

	// Add claims
	for k, v := range s.Claims {
		if k != "id" { // Don't overwrite ID
			m[k] = v
		}
	}

	return json.Marshal(m)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Subject) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// Extract ID
	if idVal, ok := m["id"]; ok {
		if idStr, ok := idVal.(string); ok {
			parsedDID, err := did.Parse(idStr)
			if err != nil {
				return err
			}
			s.ID = parsedDID
		}
		delete(m, "id")
	}

	// Remaining fields are claims
	s.Claims = m

	return nil
}

// NewSubject creates a new subject with a DID.
func NewSubject(subjectDID did.DID) Subject {
	return Subject{
		ID:     subjectDID,
		Claims: make(map[string]interface{}),
	}
}

// WithClaim adds a claim to the subject.
func (s Subject) WithClaim(key string, value interface{}) Subject {
	if s.Claims == nil {
		s.Claims = make(map[string]interface{})
	}
	s.Claims[key] = value
	return s
}

// WithClaims adds multiple claims to the subject.
func (s Subject) WithClaims(claims map[string]interface{}) Subject {
	if s.Claims == nil {
		s.Claims = make(map[string]interface{})
	}
	for k, v := range claims {
		s.Claims[k] = v
	}
	return s
}

// GetClaim retrieves a claim by key.
func (s Subject) GetClaim(key string) (interface{}, bool) {
	if s.Claims == nil {
		return nil, false
	}
	v, ok := s.Claims[key]
	return v, ok
}

// GetStringClaim retrieves a claim as a string.
func (s Subject) GetStringClaim(key string) (string, bool) {
	v, ok := s.GetClaim(key)
	if !ok {
		return "", false
	}
	str, ok := v.(string)
	return str, ok
}

// GetIntClaim retrieves a claim as an int.
func (s Subject) GetIntClaim(key string) (int, bool) {
	v, ok := s.GetClaim(key)
	if !ok {
		return 0, false
	}

	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, false
	}
}

// GetBoolClaim retrieves a claim as a bool.
func (s Subject) GetBoolClaim(key string) (bool, bool) {
	v, ok := s.GetClaim(key)
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// GetTimeClaim retrieves a claim as a time.Time.
func (s Subject) GetTimeClaim(key string) (time.Time, bool) {
	v, ok := s.GetClaim(key)
	if !ok {
		return time.Time{}, false
	}

	switch val := v.(type) {
	case time.Time:
		return val, true
	case string:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	default:
		return time.Time{}, false
	}
}

// HasClaim returns true if the claim exists.
func (s Subject) HasClaim(key string) bool {
	if s.Claims == nil {
		return false
	}
	_, ok := s.Claims[key]
	return ok
}

// ClaimKeys returns all claim keys.
func (s Subject) ClaimKeys() []string {
	if s.Claims == nil {
		return nil
	}
	keys := make([]string, 0, len(s.Claims))
	for k := range s.Claims {
		keys = append(keys, k)
	}
	return keys
}

// ============================================================================
// Pre-defined Claim Builders
// ============================================================================

// GitHubClaims contains claims for GitHub contribution credentials.
type GitHubClaims struct {
	Username      string    `json:"username"`
	Commits       int       `json:"commits,omitempty"`
	Repositories  int       `json:"repositories,omitempty"`
	Stars         int       `json:"stars,omitempty"`
	Followers     int       `json:"followers,omitempty"`
	Contributions int       `json:"contributions,omitempty"`
	VerifiedAt    time.Time `json:"verifiedAt"`
}

// ToSubject converts GitHubClaims to a Subject.
func (g GitHubClaims) ToSubject(subjectDID did.DID) Subject {
	return NewSubject(subjectDID).
		WithClaim("username", g.Username).
		WithClaim("commits", g.Commits).
		WithClaim("repositories", g.Repositories).
		WithClaim("stars", g.Stars).
		WithClaim("followers", g.Followers).
		WithClaim("contributions", g.Contributions).
		WithClaim("verifiedAt", g.VerifiedAt.Format(time.RFC3339))
}

// LinkedInClaims contains claims for LinkedIn employment credentials.
type LinkedInClaims struct {
	Name        string    `json:"name"`
	Company     string    `json:"company,omitempty"`
	Title       string    `json:"title,omitempty"`
	StartDate   string    `json:"startDate,omitempty"`
	EndDate     string    `json:"endDate,omitempty"`
	Connections int       `json:"connections,omitempty"`
	VerifiedAt  time.Time `json:"verifiedAt"`
}

// ToSubject converts LinkedInClaims to a Subject.
func (l LinkedInClaims) ToSubject(subjectDID did.DID) Subject {
	s := NewSubject(subjectDID).
		WithClaim("name", l.Name).
		WithClaim("verifiedAt", l.VerifiedAt.Format(time.RFC3339))

	if l.Company != "" {
		s = s.WithClaim("company", l.Company)
	}
	if l.Title != "" {
		s = s.WithClaim("title", l.Title)
	}
	if l.StartDate != "" {
		s = s.WithClaim("startDate", l.StartDate)
	}
	if l.EndDate != "" {
		s = s.WithClaim("endDate", l.EndDate)
	}
	if l.Connections > 0 {
		s = s.WithClaim("connections", l.Connections)
	}

	return s
}

// EducationClaims contains claims for education credentials.
type EducationClaims struct {
	Institution    string    `json:"institution"`
	Degree         string    `json:"degree,omitempty"`
	FieldOfStudy   string    `json:"fieldOfStudy,omitempty"`
	StartDate      string    `json:"startDate,omitempty"`
	EndDate        string    `json:"endDate,omitempty"`
	GraduationDate string    `json:"graduationDate,omitempty"`
	Grade          string    `json:"grade,omitempty"`
	VerifiedAt     time.Time `json:"verifiedAt"`
}

// ToSubject converts EducationClaims to a Subject.
func (e EducationClaims) ToSubject(subjectDID did.DID) Subject {
	s := NewSubject(subjectDID).
		WithClaim("institution", e.Institution).
		WithClaim("verifiedAt", e.VerifiedAt.Format(time.RFC3339))

	if e.Degree != "" {
		s = s.WithClaim("degree", e.Degree)
	}
	if e.FieldOfStudy != "" {
		s = s.WithClaim("fieldOfStudy", e.FieldOfStudy)
	}
	if e.StartDate != "" {
		s = s.WithClaim("startDate", e.StartDate)
	}
	if e.EndDate != "" {
		s = s.WithClaim("endDate", e.EndDate)
	}
	if e.GraduationDate != "" {
		s = s.WithClaim("graduationDate", e.GraduationDate)
	}
	if e.Grade != "" {
		s = s.WithClaim("grade", e.Grade)
	}

	return s
}

// CertificationClaims contains claims for professional certifications.
type CertificationClaims struct {
	Name           string    `json:"name"`
	Issuer         string    `json:"issuer"`
	IssueDate      string    `json:"issueDate,omitempty"`
	ExpirationDate string    `json:"expirationDate,omitempty"`
	CredentialID   string    `json:"credentialId,omitempty"`
	CredentialURL  string    `json:"credentialUrl,omitempty"`
	VerifiedAt     time.Time `json:"verifiedAt"`
}

// ToSubject converts CertificationClaims to a Subject.
func (c CertificationClaims) ToSubject(subjectDID did.DID) Subject {
	s := NewSubject(subjectDID).
		WithClaim("name", c.Name).
		WithClaim("issuer", c.Issuer).
		WithClaim("verifiedAt", c.VerifiedAt.Format(time.RFC3339))

	if c.IssueDate != "" {
		s = s.WithClaim("issueDate", c.IssueDate)
	}
	if c.ExpirationDate != "" {
		s = s.WithClaim("expirationDate", c.ExpirationDate)
	}
	if c.CredentialID != "" {
		s = s.WithClaim("credentialId", c.CredentialID)
	}
	if c.CredentialURL != "" {
		s = s.WithClaim("credentialUrl", c.CredentialURL)
	}

	return s
}

// SkillClaims contains claims for skill credentials.
type SkillClaims struct {
	Name         string    `json:"name"`
	Level        string    `json:"level,omitempty"`
	Endorsements int       `json:"endorsements,omitempty"`
	VerifiedAt   time.Time `json:"verifiedAt"`
}

// ToSubject converts SkillClaims to a Subject.
func (s SkillClaims) ToSubject(subjectDID did.DID) Subject {
	subj := NewSubject(subjectDID).
		WithClaim("name", s.Name).
		WithClaim("verifiedAt", s.VerifiedAt.Format(time.RFC3339))

	if s.Level != "" {
		subj = subj.WithClaim("level", s.Level)
	}
	if s.Endorsements > 0 {
		subj = subj.WithClaim("endorsements", s.Endorsements)
	}

	return subj
}
