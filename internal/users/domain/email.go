package domain

import (
	"regexp"
	"strings"

	"github.com/0xsj/result"
)

// emailRegex is a simple regex for email validation.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is a value object representing a validated email address.
type Email struct {
	value string
}

// NewEmail creates a new Email value object.
// Returns an error if the email is invalid.
func NewEmail(email string) result.Result[Email] {
	// Trim whitespace
	email = strings.TrimSpace(email)

	// Check if empty
	if email == "" {
		return result.Err[Email](ErrEmptyEmail())
	}

	// Check length (RFC 5321 specifies 254 as max)
	if len(email) > 254 {
		return result.Err[Email](ErrEmailTooLong(len(email), 254))
	}

	// Convert to lowercase for consistency
	email = strings.ToLower(email)

	// Validate format
	if !emailRegex.MatchString(email) {
		return result.Err[Email](ErrInvalidEmail())
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

// Domain returns the domain part of the email.
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
