package vc

import (
	"encoding/json"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Credential represents a W3C Verifiable Credential.
// Spec: https://www.w3.org/TR/vc-data-model/
type Credential struct {
	// Context defines the JSON-LD context(s).
	Context []string `json:"@context"`

	// ID is the unique identifier for this credential.
	ID string `json:"id,omitempty"`

	// Type specifies the credential type(s).
	// Must include "VerifiableCredential".
	Type []CredentialType `json:"type"`

	// Issuer is the DID of the credential issuer.
	Issuer Issuer `json:"issuer"`

	// IssuanceDate is when the credential was issued.
	IssuanceDate time.Time `json:"issuanceDate"`

	// ExpirationDate is when the credential expires (optional).
	ExpirationDate *time.Time `json:"expirationDate,omitempty"`

	// ValidFrom is when the credential becomes valid (optional, VC 2.0).
	ValidFrom *time.Time `json:"validFrom,omitempty"`

	// ValidUntil is when the credential stops being valid (optional, VC 2.0).
	ValidUntil *time.Time `json:"validUntil,omitempty"`

	// CredentialSubject contains the claims about the subject.
	CredentialSubject Subject `json:"credentialSubject"`

	// CredentialStatus provides revocation/status information (optional).
	CredentialStatus *CredentialStatus `json:"credentialStatus,omitempty"`

	// CredentialSchema specifies the schema for validation (optional).
	CredentialSchema *CredentialSchema `json:"credentialSchema,omitempty"`

	// Evidence provides supporting evidence for the claims (optional).
	Evidence []Evidence `json:"evidence,omitempty"`

	// Proof contains the cryptographic proof (for JSON-LD format).
	// For JWT format, the proof is in the JWT signature.
	Proof *Proof `json:"proof,omitempty"`
}

// Issuer represents the credential issuer.
// Can be a simple DID string or an object with additional properties.
type Issuer struct {
	// ID is the DID of the issuer.
	ID did.DID

	// Name is the human-readable name of the issuer (optional).
	Name string

	// URL is the issuer's website (optional).
	URL string

	// Image is the issuer's logo URL (optional).
	Image string
}

// MarshalJSON implements json.Marshaler.
func (i Issuer) MarshalJSON() ([]byte, error) {
	// If only ID is set, marshal as string
	if i.Name == "" && i.URL == "" && i.Image == "" {
		return json.Marshal(i.ID.String())
	}

	// Otherwise marshal as object
	return json.Marshal(map[string]interface{}{
		"id":    i.ID.String(),
		"name":  i.Name,
		"url":   i.URL,
		"image": i.Image,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *Issuer) UnmarshalJSON(data []byte) error {
	// Try string first
	var idStr string
	if err := json.Unmarshal(data, &idStr); err == nil {
		parsedDID, err := did.Parse(idStr)
		if err != nil {
			return err
		}
		i.ID = parsedDID
		return nil
	}

	// Try object
	var obj struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		URL   string `json:"url"`
		Image string `json:"image"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	parsedDID, err := did.Parse(obj.ID)
	if err != nil {
		return err
	}

	i.ID = parsedDID
	i.Name = obj.Name
	i.URL = obj.URL
	i.Image = obj.Image

	return nil
}

// CredentialStatus provides information about credential revocation status.
type CredentialStatus struct {
	// ID is the unique identifier for the status entry.
	ID string `json:"id"`

	// Type specifies the status method type.
	Type string `json:"type"`

	// StatusListIndex is the index in a status list (for StatusList2021).
	StatusListIndex string `json:"statusListIndex,omitempty"`

	// StatusListCredential is the URL of the status list credential.
	StatusListCredential string `json:"statusListCredential,omitempty"`

	// StatusPurpose indicates the purpose (revocation, suspension).
	StatusPurpose string `json:"statusPurpose,omitempty"`
}

// CredentialSchema specifies the schema for credential validation.
type CredentialSchema struct {
	// ID is the URL of the schema.
	ID string `json:"id"`

	// Type specifies the schema type.
	Type string `json:"type"`
}

// Evidence provides supporting evidence for credential claims.
type Evidence struct {
	// ID is the unique identifier for this evidence.
	ID string `json:"id,omitempty"`

	// Type specifies the evidence type.
	Type []string `json:"type"`

	// Verifier is who verified the evidence.
	Verifier string `json:"verifier,omitempty"`

	// EvidenceDocument is the document reference.
	EvidenceDocument string `json:"evidenceDocument,omitempty"`

	// SubjectPresence indicates subject presence during verification.
	SubjectPresence string `json:"subjectPresence,omitempty"`

	// DocumentPresence indicates document presence during verification.
	DocumentPresence string `json:"documentPresence,omitempty"`
}

// ============================================================================
// Credential Builder
// ============================================================================

// NewCredential creates a new credential with required fields.
func NewCredential(id string, issuer did.DID, subject Subject) *Credential {
	return &Credential{
		Context:           DefaultContext(),
		ID:                id,
		Type:              []CredentialType{TypeVerifiableCredential},
		Issuer:            Issuer{ID: issuer},
		IssuanceDate:      time.Now().UTC(),
		CredentialSubject: subject,
	}
}

// WithType adds a credential type.
func (c *Credential) WithType(credType CredentialType) *Credential {
	c.Type = append(c.Type, credType)
	return c
}

// WithTypes sets multiple credential types (preserves VerifiableCredential).
func (c *Credential) WithTypes(credTypes ...CredentialType) *Credential {
	c.Type = append([]CredentialType{TypeVerifiableCredential}, credTypes...)
	return c
}

// WithIssuanceDate sets the issuance date.
func (c *Credential) WithIssuanceDate(t time.Time) *Credential {
	c.IssuanceDate = t.UTC()
	return c
}

// WithExpirationDate sets the expiration date.
func (c *Credential) WithExpirationDate(t time.Time) *Credential {
	utc := t.UTC()
	c.ExpirationDate = &utc
	return c
}

// WithValidityPeriod sets both issuance and expiration dates.
func (c *Credential) WithValidityPeriod(from, until time.Time) *Credential {
	c.IssuanceDate = from.UTC()
	utcUntil := until.UTC()
	c.ExpirationDate = &utcUntil
	return c
}

// WithIssuerName sets the issuer's name.
func (c *Credential) WithIssuerName(name string) *Credential {
	c.Issuer.Name = name
	return c
}

// WithIssuerURL sets the issuer's URL.
func (c *Credential) WithIssuerURL(url string) *Credential {
	c.Issuer.URL = url
	return c
}

// WithStatus sets the credential status.
func (c *Credential) WithStatus(status *CredentialStatus) *Credential {
	c.CredentialStatus = status
	return c
}

// WithSchema sets the credential schema.
func (c *Credential) WithSchema(schema *CredentialSchema) *Credential {
	c.CredentialSchema = schema
	return c
}

// WithEvidence adds evidence to the credential.
func (c *Credential) WithEvidence(evidence Evidence) *Credential {
	c.Evidence = append(c.Evidence, evidence)
	return c
}

// WithContext adds a context URI.
func (c *Credential) WithContext(ctx string) *Credential {
	c.Context = append(c.Context, ctx)
	return c
}

// WithProof sets the proof.
func (c *Credential) WithProof(proof *Proof) *Credential {
	c.Proof = proof
	return c
}

// ============================================================================
// Credential Validation
// ============================================================================

// Validate checks if the credential structure is valid.
func (c *Credential) Validate() error {
	const op = "Credential.Validate"

	if len(c.Context) == 0 {
		return ErrInvalidCredential(op, "context is required")
	}

	if len(c.Type) == 0 {
		return ErrInvalidCredential(op, "type is required")
	}

	// Must include VerifiableCredential type
	hasVC := false
	for _, t := range c.Type {
		if t == TypeVerifiableCredential {
			hasVC = true
			break
		}
	}
	if !hasVC {
		return ErrInvalidCredential(op, "type must include 'VerifiableCredential'")
	}

	if c.Issuer.ID.IsZero() {
		return ErrInvalidIssuer(op, "issuer ID is required")
	}

	if c.IssuanceDate.IsZero() {
		return ErrInvalidCredential(op, "issuanceDate is required")
	}

	if c.CredentialSubject.ID.IsZero() {
		return ErrInvalidSubject(op, "credentialSubject.id is required")
	}

	return nil
}

// IsExpired returns true if the credential has expired.
func (c *Credential) IsExpired() bool {
	if c.ExpirationDate == nil {
		return false
	}
	return time.Now().After(*c.ExpirationDate)
}

// IsActive returns true if the credential is currently valid (not expired, past issuance).
func (c *Credential) IsActive() bool {
	now := time.Now()

	// Check if issued yet
	if now.Before(c.IssuanceDate) {
		return false
	}

	// Check if expired
	if c.ExpirationDate != nil && now.After(*c.ExpirationDate) {
		return false
	}

	return true
}

// SubjectDID returns the DID of the credential subject.
func (c *Credential) SubjectDID() did.DID {
	return c.CredentialSubject.ID
}

// IssuerDID returns the DID of the credential issuer.
func (c *Credential) IssuerDID() did.DID {
	return c.Issuer.ID
}

// HasType returns true if the credential has the specified type.
func (c *Credential) HasType(credType CredentialType) bool {
	for _, t := range c.Type {
		if t == credType {
			return true
		}
	}
	return false
}

// ============================================================================
// Credential ID Generation
// ============================================================================

// GenerateCredentialID generates a unique credential ID.
func GenerateCredentialID(issuerDID did.DID) string {
	id := types.NewID()
	return "urn:uuid:" + id.String()
}

// GenerateCredentialIDWithPrefix generates a credential ID with a custom prefix.
func GenerateCredentialIDWithPrefix(prefix string) string {
	id := types.NewID()
	return prefix + id.String()
}
