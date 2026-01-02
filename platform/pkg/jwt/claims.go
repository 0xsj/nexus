package jwt

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// ============================================================================
// Standard Claims
// ============================================================================

// Claims represents the standard JWT claims (RFC 7519).
type Claims struct {
	// Issuer identifies the principal that issued the JWT.
	Issuer string `json:"iss,omitempty"`

	// Subject identifies the principal that is the subject of the JWT.
	Subject string `json:"sub,omitempty"`

	// Audience identifies the recipients that the JWT is intended for.
	Audience Audience `json:"aud,omitempty"`

	// ExpiresAt identifies the expiration time on or after which the JWT MUST NOT be accepted.
	ExpiresAt *time.Time `json:"exp,omitempty"`

	// NotBefore identifies the time before which the JWT MUST NOT be accepted.
	NotBefore *time.Time `json:"nbf,omitempty"`

	// IssuedAt identifies the time at which the JWT was issued.
	IssuedAt *time.Time `json:"iat,omitempty"`

	// ID provides a unique identifier for the JWT.
	ID string `json:"jti,omitempty"`
}

// NewClaims creates a new Claims with issuer and subject.
func NewClaims(issuer, subject string) Claims {
	now := time.Now()
	return Claims{
		Issuer:   issuer,
		Subject:  subject,
		IssuedAt: &now,
	}
}

// WithAudience sets the audience.
func (c Claims) WithAudience(aud ...string) Claims {
	c.Audience = aud
	return c
}

// WithExpiresAt sets the expiration time.
func (c Claims) WithExpiresAt(exp time.Time) Claims {
	c.ExpiresAt = &exp
	return c
}

// WithExpiresIn sets expiration relative to now.
func (c Claims) WithExpiresIn(d time.Duration) Claims {
	exp := time.Now().Add(d)
	c.ExpiresAt = &exp
	return c
}

// WithNotBefore sets the not-before time.
func (c Claims) WithNotBefore(nbf time.Time) Claims {
	c.NotBefore = &nbf
	return c
}

// WithIssuedAt sets the issued-at time.
func (c Claims) WithIssuedAt(iat time.Time) Claims {
	c.IssuedAt = &iat
	return c
}

// WithID sets the JWT ID.
func (c Claims) WithID(id string) Claims {
	c.ID = id
	return c
}

// ============================================================================
// Validation
// ============================================================================

// Validate validates the claims.
func (c Claims) Validate() error {
	return c.ValidateWithTime(time.Now())
}

// ValidateWithTime validates the claims against a specific time.
func (c Claims) ValidateWithTime(now time.Time) error {
	const op = "jwt.Claims.Validate"

	// Check expiration
	if c.ExpiresAt != nil && now.After(*c.ExpiresAt) {
		return ErrTokenExpired(op)
	}

	// Check not-before
	if c.NotBefore != nil && now.Before(*c.NotBefore) {
		return ErrTokenNotYetValid(op)
	}

	return nil
}

// ValidateAudience validates that the expected audience is present.
func (c Claims) ValidateAudience(expected string) error {
	const op = "jwt.Claims.ValidateAudience"

	if !c.Audience.Contains(expected) {
		return ErrInvalidClaims(op, "audience mismatch")
	}

	return nil
}

// ValidateIssuer validates that the issuer matches.
func (c Claims) ValidateIssuer(expected string) error {
	const op = "jwt.Claims.ValidateIssuer"

	if c.Issuer != expected {
		return ErrInvalidClaims(op, "issuer mismatch")
	}

	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// IsExpired returns true if the token has expired.
func (c Claims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// TimeUntilExpiry returns the duration until expiration.
func (c Claims) TimeUntilExpiry() time.Duration {
	if c.ExpiresAt == nil {
		return time.Duration(1<<63 - 1) // Max duration
	}
	return time.Until(*c.ExpiresAt)
}

// ============================================================================
// Audience
// ============================================================================

// Audience represents the JWT audience claim.
// Can be a single string or an array of strings.
type Audience []string

// Contains checks if the audience contains a specific value.
func (a Audience) Contains(value string) bool {
	for _, v := range a {
		if v == value {
			return true
		}
	}
	return false
}

// MarshalJSON implements json.Marshaler.
// Encodes as a string if single element, array otherwise.
func (a Audience) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("null"), nil
	}
	if len(a) == 1 {
		return json.Marshal(a[0])
	}
	return json.Marshal([]string(a))
}

// UnmarshalJSON implements json.Unmarshaler.
// Accepts both string and array of strings.
func (a *Audience) UnmarshalJSON(data []byte) error {
	// Try string first
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = Audience{single}
		return nil
	}

	// Try array
	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		*a = multiple
		return nil
	}

	return ErrInvalidClaims("jwt.Audience.UnmarshalJSON", "invalid audience format")
}

// ============================================================================
// Claims JSON Marshaling (Unix timestamps)
// ============================================================================

// claimsJSON is the JSON representation with Unix timestamps.
type claimsJSON struct {
	Issuer    string   `json:"iss,omitempty"`
	Subject   string   `json:"sub,omitempty"`
	Audience  Audience `json:"aud,omitempty"`
	ExpiresAt *int64   `json:"exp,omitempty"`
	NotBefore *int64   `json:"nbf,omitempty"`
	IssuedAt  *int64   `json:"iat,omitempty"`
	ID        string   `json:"jti,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (c Claims) MarshalJSON() ([]byte, error) {
	j := claimsJSON{
		Issuer:   c.Issuer,
		Subject:  c.Subject,
		Audience: c.Audience,
		ID:       c.ID,
	}

	if c.ExpiresAt != nil {
		exp := c.ExpiresAt.Unix()
		j.ExpiresAt = &exp
	}
	if c.NotBefore != nil {
		nbf := c.NotBefore.Unix()
		j.NotBefore = &nbf
	}
	if c.IssuedAt != nil {
		iat := c.IssuedAt.Unix()
		j.IssuedAt = &iat
	}

	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Claims) UnmarshalJSON(data []byte) error {
	var j claimsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	c.Issuer = j.Issuer
	c.Subject = j.Subject
	c.Audience = j.Audience
	c.ID = j.ID

	if j.ExpiresAt != nil {
		t := time.Unix(*j.ExpiresAt, 0).UTC()
		c.ExpiresAt = &t
	}
	if j.NotBefore != nil {
		t := time.Unix(*j.NotBefore, 0).UTC()
		c.NotBefore = &t
	}
	if j.IssuedAt != nil {
		t := time.Unix(*j.IssuedAt, 0).UTC()
		c.IssuedAt = &t
	}

	return nil
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// Encode encodes the claims to a base64url string.
func (c Claims) Encode() (string, error) {
	const op = "jwt.Claims.Encode"

	data, err := json.Marshal(c)
	if err != nil {
		return "", ErrEncodingFailed(op, err)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

// DecodeClaims decodes base64url encoded claims into standard Claims.
func DecodeClaims(encoded string) (Claims, error) {
	const op = "jwt.DecodeClaims"

	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Claims{}, ErrInvalidClaims(op, "invalid base64 encoding")
	}

	var claims Claims
	if err := json.Unmarshal(data, &claims); err != nil {
		return Claims{}, ErrInvalidClaims(op, "invalid JSON")
	}

	return claims, nil
}

// DecodeClaimsRaw decodes base64url encoded claims into raw JSON.
// Use this when you need to decode custom claims.
func DecodeClaimsRaw(encoded string) (json.RawMessage, error) {
	const op = "jwt.DecodeClaimsRaw"

	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrInvalidClaims(op, "invalid base64 encoding")
	}

	// Validate it's valid JSON
	if !json.Valid(data) {
		return nil, ErrInvalidClaims(op, "invalid JSON")
	}

	return json.RawMessage(data), nil
}

// DecodeClaimsInto decodes base64url encoded claims into a custom struct.
func DecodeClaimsInto(encoded string, v any) error {
	const op = "jwt.DecodeClaimsInto"

	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return ErrInvalidClaims(op, "invalid base64 encoding")
	}

	if err := json.Unmarshal(data, v); err != nil {
		return ErrInvalidClaims(op, "invalid JSON")
	}

	return nil
}
