package domain

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/0xsj/result"
)

// usernameRegex validates username format.
// Allows alphanumeric and underscores only (matching DB constraint).
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Username is a value object representing a validated username.
type Username struct {
	value string
}

// UsernameConstraints defines the validation rules for usernames.
type UsernameConstraints struct {
	MinLength     int
	MaxLength     int
	ReservedNames []string
}

// DefaultUsernameConstraints provides sensible defaults matching DB schema.
var DefaultUsernameConstraints = UsernameConstraints{
	MinLength:     3,
	MaxLength:     50,
	ReservedNames: []string{"admin", "root", "system", "api", "www", "support", "nexus"},
}

// NewUsername creates a new Username value object with default constraints.
func NewUsername(username string) result.Result[Username] {
	return NewUsernameWithConstraints(username, DefaultUsernameConstraints)
}

// NewUsernameWithConstraints creates a new Username with custom constraints.
func NewUsernameWithConstraints(username string, constraints UsernameConstraints) result.Result[Username] {
	// Trim whitespace
	username = strings.TrimSpace(username)

	// Check if empty
	if username == "" {
		return result.Err[Username](ErrEmptyUsername())
	}

	// Check length
	if len(username) < constraints.MinLength {
		return result.Err[Username](ErrUsernameTooShort(len(username), constraints.MinLength))
	}

	if len(username) > constraints.MaxLength {
		return result.Err[Username](ErrUsernameTooLong(len(username), constraints.MaxLength))
	}

	// Validate format (alphanumeric and underscore only)
	if !usernameRegex.MatchString(username) {
		return result.Err[Username](ErrInvalidUsernameChar(username))
	}

	// Check if starts with number
	if unicode.IsDigit(rune(username[0])) {
		return result.Err[Username](ErrUsernameStartsWithNumber())
	}

	// Convert to lowercase for consistency (matching DB constraint ~*)
	username = strings.ToLower(username)

	// Check reserved names
	for _, reserved := range constraints.ReservedNames {
		if username == strings.ToLower(reserved) {
			return result.Err[Username](ErrReservedUsername(username))
		}
	}

	return result.Ok(Username{value: username})
}

// Value returns the string value of the username.
func (u Username) Value() string {
	return u.value
}

// String implements the Stringer interface.
func (u Username) String() string {
	return u.value
}

// Equals checks if two usernames are equal.
func (u Username) Equals(other Username) bool {
	return u.value == other.value
}
