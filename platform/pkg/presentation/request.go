package presentation

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/types"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

// Request represents a presentation request from a verifier.
// The verifier sends this to specify what credentials they want to see.
type Request struct {
	// ID is the unique identifier for this request.
	ID string `json:"id"`

	// Verifier is the DID of the verifier making the request.
	Verifier did.DID `json:"verifier,omitempty"`

	// Challenge is a random value to prevent replay attacks.
	// The holder must include this in their presentation proof.
	Challenge string `json:"challenge"`

	// Domain is the intended audience domain for the presentation.
	// The holder should verify they're presenting to this domain.
	Domain string `json:"domain,omitempty"`

	// RequestedCredentials specifies what credentials are requested.
	RequestedCredentials []CredentialRequest `json:"requestedCredentials,omitempty"`

	// Purpose describes why the credentials are being requested.
	Purpose string `json:"purpose,omitempty"`

	// CreatedAt is when this request was created.
	CreatedAt time.Time `json:"createdAt"`

	// ExpiresAt is when this request expires.
	ExpiresAt time.Time `json:"expiresAt,omitempty"`

	// CallbackURL is where the presentation should be sent.
	CallbackURL string `json:"callbackUrl,omitempty"`

	// State is an opaque value for the verifier to track requests.
	State string `json:"state,omitempty"`
}

// CredentialRequest specifies requirements for a credential.
type CredentialRequest struct {
	// Type is the credential type being requested.
	Type vc.CredentialType `json:"type"`

	// Required indicates if this credential is mandatory.
	Required bool `json:"required"`

	// TrustedIssuers limits which issuers are accepted.
	// Empty means any issuer is accepted.
	TrustedIssuers []did.DID `json:"trustedIssuers,omitempty"`

	// RequestedFields specifies which fields should be disclosed.
	// Used for selective disclosure (BBS+ / Phase 2).
	// Empty means all fields.
	RequestedFields []string `json:"requestedFields,omitempty"`

	// Constraints defines additional constraints on the credential.
	Constraints *CredentialConstraints `json:"constraints,omitempty"`
}

// CredentialConstraints defines constraints on credential values.
type CredentialConstraints struct {
	// MinIssuanceDate requires the credential to be issued after this date.
	MinIssuanceDate *time.Time `json:"minIssuanceDate,omitempty"`

	// MaxAge limits how old the credential can be.
	MaxAge *time.Duration `json:"maxAge,omitempty"`

	// RequireNotExpired requires the credential to not be expired.
	RequireNotExpired bool `json:"requireNotExpired,omitempty"`

	// RequireNotRevoked requires the credential to not be revoked.
	RequireNotRevoked bool `json:"requireNotRevoked,omitempty"`

	// FieldConstraints defines constraints on specific fields.
	FieldConstraints []FieldConstraint `json:"fieldConstraints,omitempty"`
}

// FieldConstraint defines a constraint on a credential field.
type FieldConstraint struct {
	// Field is the JSON path to the field.
	Field string `json:"field"`

	// Operator is the comparison operator.
	Operator ConstraintOperator `json:"operator"`

	// Value is the value to compare against.
	Value interface{} `json:"value"`
}

// ConstraintOperator defines comparison operators for constraints.
type ConstraintOperator string

const (
	OpEquals       ConstraintOperator = "eq"
	OpNotEquals    ConstraintOperator = "neq"
	OpGreaterThan  ConstraintOperator = "gt"
	OpGreaterEqual ConstraintOperator = "gte"
	OpLessThan     ConstraintOperator = "lt"
	OpLessEqual    ConstraintOperator = "lte"
	OpContains     ConstraintOperator = "contains"
	OpExists       ConstraintOperator = "exists"
)

// ============================================================================
// Request Builder
// ============================================================================

// NewRequest creates a new presentation request.
func NewRequest(challenge string) *Request {
	return &Request{
		ID:        GenerateRequestID(),
		Challenge: challenge,
		CreatedAt: time.Now().UTC(),
	}
}

// NewRequestWithChallenge creates a request with an auto-generated challenge.
func NewRequestWithChallenge() *Request {
	return NewRequest(GenerateChallenge())
}

// WithID sets the request ID.
func (r *Request) WithID(id string) *Request {
	r.ID = id
	return r
}

// WithVerifier sets the verifier DID.
func (r *Request) WithVerifier(verifier did.DID) *Request {
	r.Verifier = verifier
	return r
}

// WithDomain sets the domain.
func (r *Request) WithDomain(domain string) *Request {
	r.Domain = domain
	return r
}

// WithPurpose sets the purpose.
func (r *Request) WithPurpose(purpose string) *Request {
	r.Purpose = purpose
	return r
}

// WithExpiration sets the expiration time.
func (r *Request) WithExpiration(expiresAt time.Time) *Request {
	r.ExpiresAt = expiresAt
	return r
}

// WithExpiresIn sets expiration relative to now.
func (r *Request) WithExpiresIn(duration time.Duration) *Request {
	r.ExpiresAt = time.Now().UTC().Add(duration)
	return r
}

// WithCallbackURL sets the callback URL.
func (r *Request) WithCallbackURL(url string) *Request {
	r.CallbackURL = url
	return r
}

// WithState sets the state parameter.
func (r *Request) WithState(state string) *Request {
	r.State = state
	return r
}

// WithCredentialRequest adds a credential request.
func (r *Request) WithCredentialRequest(cr CredentialRequest) *Request {
	r.RequestedCredentials = append(r.RequestedCredentials, cr)
	return r
}

// RequestCredential adds a required credential type.
func (r *Request) RequestCredential(credType vc.CredentialType) *Request {
	return r.WithCredentialRequest(CredentialRequest{
		Type:     credType,
		Required: true,
	})
}

// RequestOptionalCredential adds an optional credential type.
func (r *Request) RequestOptionalCredential(credType vc.CredentialType) *Request {
	return r.WithCredentialRequest(CredentialRequest{
		Type:     credType,
		Required: false,
	})
}

// RequestCredentialFromIssuers adds a required credential with trusted issuers.
func (r *Request) RequestCredentialFromIssuers(credType vc.CredentialType, issuers ...did.DID) *Request {
	return r.WithCredentialRequest(CredentialRequest{
		Type:           credType,
		Required:       true,
		TrustedIssuers: issuers,
	})
}

// ============================================================================
// Request Validation
// ============================================================================

// Validate checks if the request is valid.
func (r *Request) Validate() error {
	const op = "Request.Validate"

	if r.ID == "" {
		return ErrInvalidPresentation(op, "request ID is required")
	}

	if r.Challenge == "" {
		return ErrInvalidChallenge(op)
	}

	if !r.ExpiresAt.IsZero() && time.Now().After(r.ExpiresAt) {
		return ErrPresentationExpired(op, r.ID)
	}

	return nil
}

// IsExpired returns true if the request has expired.
func (r *Request) IsExpired() bool {
	if r.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(r.ExpiresAt)
}

// ============================================================================
// Request Queries
// ============================================================================

// RequiredCredentialTypes returns all required credential types.
func (r *Request) RequiredCredentialTypes() []vc.CredentialType {
	var result []vc.CredentialType
	for _, cr := range r.RequestedCredentials {
		if cr.Required {
			result = append(result, cr.Type)
		}
	}
	return result
}

// OptionalCredentialTypes returns all optional credential types.
func (r *Request) OptionalCredentialTypes() []vc.CredentialType {
	var result []vc.CredentialType
	for _, cr := range r.RequestedCredentials {
		if !cr.Required {
			result = append(result, cr.Type)
		}
	}
	return result
}

// AllCredentialTypes returns all requested credential types.
func (r *Request) AllCredentialTypes() []vc.CredentialType {
	result := make([]vc.CredentialType, len(r.RequestedCredentials))
	for i, cr := range r.RequestedCredentials {
		result[i] = cr.Type
	}
	return result
}

// GetCredentialRequest returns the request for a specific type.
func (r *Request) GetCredentialRequest(credType vc.CredentialType) *CredentialRequest {
	for i := range r.RequestedCredentials {
		if r.RequestedCredentials[i].Type == credType {
			return &r.RequestedCredentials[i]
		}
	}
	return nil
}

// IsTrustedIssuer checks if an issuer is trusted for a credential type.
func (r *Request) IsTrustedIssuer(credType vc.CredentialType, issuer did.DID) bool {
	cr := r.GetCredentialRequest(credType)
	if cr == nil {
		return false
	}

	// Empty list means any issuer is trusted
	if len(cr.TrustedIssuers) == 0 {
		return true
	}

	for _, trusted := range cr.TrustedIssuers {
		if trusted.Equals(issuer) {
			return true
		}
	}
	return false
}

// ============================================================================
// ID and Challenge Generation
// ============================================================================

// GenerateRequestID generates a unique request ID.
func GenerateRequestID() string {
	id := types.NewID()
	return "urn:uuid:" + id.String()
}

// GenerateChallenge generates a random challenge string.
func GenerateChallenge() string {
	id := types.NewID()
	return id.String()
}
