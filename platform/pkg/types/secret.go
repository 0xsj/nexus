package types

import (
	"database/sql/driver"
	"encoding/json"
)

// Secret is a sensitive string that is redacted in logs and JSON output.
// Use for API keys, tokens, passwords, and other sensitive data.
//
// The underlying value is never exposed through String(), MarshalJSON(),
// or other standard output methods. Use Expose() to access the raw value.
type Secret struct {
	value string
}

// RedactedValue is the placeholder shown when a secret is printed.
const RedactedValue = "[REDACTED]"

// NewSecret creates a new Secret from a string.
func NewSecret(s string) Secret {
	return Secret{value: s}
}

// Expose returns the underlying secret value.
// Use with caution — only when the actual value is needed.
func (s Secret) Expose() string {
	return s.value
}

// IsEmpty returns true if the secret is empty.
func (s Secret) IsEmpty() bool {
	return s.value == ""
}

// IsValid returns true if the secret is non-empty.
func (s Secret) IsValid() bool {
	return s.value != ""
}

// Length returns the length of the secret without exposing it.
func (s Secret) Length() int {
	return len(s.value)
}

// Equals compares two secrets for equality.
// Uses constant-time comparison would be ideal for crypto,
// but for simplicity we use standard comparison here.
func (s Secret) Equals(other Secret) bool {
	return s.value == other.value
}

// String returns a redacted placeholder.
// Implements fmt.Stringer to prevent accidental logging.
func (s Secret) String() string {
	if s.IsEmpty() {
		return ""
	}
	return RedactedValue
}

// GoString returns a redacted placeholder.
// Implements fmt.GoStringer to prevent accidental logging via %#v.
func (s Secret) GoString() string {
	if s.IsEmpty() {
		return "Secret{empty}"
	}
	return "Secret{" + RedactedValue + "}"
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON returns a redacted placeholder.
// Prevents accidental exposure in JSON responses or logs.
func (s Secret) MarshalJSON() ([]byte, error) {
	if s.IsEmpty() {
		return []byte("null"), nil
	}
	return json.Marshal(RedactedValue)
}

// UnmarshalJSON reads the secret value from JSON.
// Allows secrets to be read from config files or requests.
func (s *Secret) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	// Don't store the redacted placeholder as an actual value
	if v == RedactedValue {
		*s = Secret{}
		return nil
	}

	*s = Secret{value: v}
	return nil
}

// ============================================================================
// SQL Scanning
// ============================================================================

// Scan implements sql.Scanner.
func (s *Secret) Scan(value interface{}) error {
	if value == nil {
		*s = Secret{}
		return nil
	}

	switch v := value.(type) {
	case string:
		*s = Secret{value: v}
		return nil

	case []byte:
		*s = Secret{value: string(v)}
		return nil

	default:
		*s = Secret{}
		return nil
	}
}

// Value implements driver.Valuer.
// Returns the actual value for database storage.
func (s Secret) Value() (driver.Value, error) {
	if s.IsEmpty() {
		return nil, nil
	}
	return s.value, nil
}

// ============================================================================
// Text Marshaling
// ============================================================================

// MarshalText returns a redacted placeholder.
func (s Secret) MarshalText() ([]byte, error) {
	if s.IsEmpty() {
		return []byte{}, nil
	}
	return []byte(RedactedValue), nil
}

// UnmarshalText reads the secret value from text.
func (s *Secret) UnmarshalText(text []byte) error {
	v := string(text)

	// Don't store the redacted placeholder as an actual value
	if v == RedactedValue {
		*s = Secret{}
		return nil
	}

	*s = Secret{value: v}
	return nil
}
