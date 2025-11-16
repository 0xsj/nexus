package domain

import (
	"regexp"
	"strings"

	"github.com/0xsj/result"
)

const (
	// emailMaxLength is the maximum allowed length for an email address.
	// RFC 5321 specifies 254 as the maximum length.
	emailMaxLength = 254
)

// emailRegex validates email format.
// This is a simplified regex for basic email validation.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is a value object representing a validated email address.
// Email is immutable and always guaranteed to be valid.
type Email struct {
	value string
}

// NewEmail creates a new Email value object.
// Returns a Result containing the Email or an error if validation fails.
func NewEmail(email string) result.Result[Email] {
	// Trim whitespace
	email = strings.TrimSpace(email)

	// Check if empty
	if email == "" {
		return result.Err[Email](ErrInvalidEmail{
			Email:  email,
			Reason: "email cannot be empty",
		})
	}

	// Check length
	if len(email) > emailMaxLength {
		return result.Err[Email](ErrInvalidEmail{
			Email:  email,
			Reason: "email cannot exceed 254 characters",
		})
	}

	// Convert to lowercase for consistency
	email = strings.ToLower(email)

	// Validate format
	if !emailRegex.MatchString(email) {
		return result.Err[Email](ErrInvalidEmail{
			Email:  email,
			Reason: "invalid email format",
		})
	}

	return result.Ok(Email{value: email})
}

// Value returns the string value of the email.
func (e Email) Value() string {
	return e.value
}

// String implements the Stringer interface.
func (e Email) String() string {
	return e.value
}

// Equals checks if two emails are equal.
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// IsEmpty checks if the email is empty (zero value).
func (e Email) IsEmpty() bool {
	return e.value == ""
}

// Domain returns the domain part of the email (after @).
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// LocalPart returns the local part of the email (before @).
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// MarshalText implements encoding.TextMarshaler for JSON/XML serialization.
func (e Email) MarshalText() ([]byte, error) {
	return []byte(e.value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON/XML deserialization.
func (e *Email) UnmarshalText(text []byte) error {
	emailResult := NewEmail(string(text))
	if emailResult.IsErr() {
		return emailResult.UnwrapErr()
	}
	*e = emailResult.Unwrap()
	return nil
}
