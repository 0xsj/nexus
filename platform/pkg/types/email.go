package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
)

// Email is a validated email address value object.
// Ensures email addresses are syntactically valid and normalized.
//
// Zero value is invalid — use NewEmail() to create valid emails.
type Email struct {
	value string
}

// NewEmail parses and validates an email address.
// Returns an error if the email is invalid.
func NewEmail(s string) (Email, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}

	// Normalize to lowercase
	s = strings.ToLower(s)

	// Validate using net/mail
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return Email{}, fmt.Errorf("invalid email format: %w", err)
	}

	return Email{value: addr.Address}, nil
}

// MustNewEmail parses an email and panics if invalid.
// Only use for constants or tests.
func MustNewEmail(s string) Email {
	email, err := NewEmail(s)
	if err != nil {
		panic(fmt.Sprintf("invalid email: %v", err))
	}
	return email
}

// String returns the email address string.
func (e Email) String() string {
	return e.value
}

// IsEmpty returns true if the email is empty.
func (e Email) IsEmpty() bool {
	return e.value == ""
}

// IsValid returns true if the email is valid (non-empty).
func (e Email) IsValid() bool {
	return e.value != ""
}

// Local returns the local part (before @).
func (e Email) Local() string {
	if e.IsEmpty() {
		return ""
	}
	parts := strings.SplitN(e.value, "@", 2)
	if len(parts) < 1 {
		return ""
	}
	return parts[0]
}

// Domain returns the domain part (after @).
func (e Email) Domain() string {
	if e.IsEmpty() {
		return ""
	}
	parts := strings.SplitN(e.value, "@", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// Equals checks if two emails are equal.
// Comparison is case-insensitive (emails are normalized to lowercase).
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Masked returns a masked version for display.
// Example: "john.doe@example.com" -> "j*****e@example.com"
func (e Email) Masked() string {
	if e.IsEmpty() {
		return ""
	}

	local := e.Local()
	domain := e.Domain()

	if len(local) <= 2 {
		return local + "@" + domain
	}

	masked := string(local[0]) + strings.Repeat("*", len(local)-2) + string(local[len(local)-1])
	return masked + "@" + domain
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
func (e Email) MarshalJSON() ([]byte, error) {
	if e.IsEmpty() {
		return []byte("null"), nil
	}
	return json.Marshal(e.value)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Email) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" || s == "null" {
		*e = Email{}
		return nil
	}

	parsed, err := NewEmail(s)
	if err != nil {
		return err
	}

	*e = parsed
	return nil
}

// ============================================================================
// SQL Scanning
// ============================================================================

// Scan implements sql.Scanner.
func (e *Email) Scan(value interface{}) error {
	if value == nil {
		*e = Email{}
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			*e = Email{}
			return nil
		}
		parsed, err := NewEmail(v)
		if err != nil {
			return err
		}
		*e = parsed
		return nil

	case []byte:
		if len(v) == 0 {
			*e = Email{}
			return nil
		}
		parsed, err := NewEmail(string(v))
		if err != nil {
			return err
		}
		*e = parsed
		return nil

	default:
		return fmt.Errorf("cannot scan %T into Email", value)
	}
}

// Value implements driver.Valuer.
func (e Email) Value() (driver.Value, error) {
	if e.IsEmpty() {
		return nil, nil
	}
	return e.value, nil
}

// ============================================================================
// Text Marshaling
// ============================================================================

// MarshalText implements encoding.TextMarshaler.
func (e Email) MarshalText() ([]byte, error) {
	if e.IsEmpty() {
		return []byte{}, nil
	}
	return []byte(e.value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (e *Email) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*e = Email{}
		return nil
	}

	parsed, err := NewEmail(string(text))
	if err != nil {
		return err
	}

	*e = parsed
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// GoString implements fmt.GoStringer for debugging.
func (e Email) GoString() string {
	if e.IsEmpty() {
		return "Email{empty}"
	}
	return fmt.Sprintf("Email{%s}", e.value)
}
