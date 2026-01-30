package domain

import (
	"strings"
	"unicode/utf8"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// DisplayName constraints
const (
	DisplayNameMinLength = 1
	DisplayNameMaxLength = 100
)

// DisplayName represents a user's display name.
// Validated to ensure it meets length and content requirements.
type DisplayName struct {
	value string
}

// NewDisplayName creates a new DisplayName from a string.
// Validates length and trims whitespace.
func NewDisplayName(s string) (DisplayName, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return DisplayName{}, pkgerrors.Validation("DisplayName.New", "display name cannot be empty")
	}

	length := utf8.RuneCountInString(s)
	if length < DisplayNameMinLength {
		return DisplayName{}, pkgerrors.Validation("DisplayName.New", "display name too short").
			WithMeta("min_length", DisplayNameMinLength).
			WithMeta("actual_length", length)
	}

	if length > DisplayNameMaxLength {
		return DisplayName{}, pkgerrors.Validation("DisplayName.New", "display name too long").
			WithMeta("max_length", DisplayNameMaxLength).
			WithMeta("actual_length", length)
	}

	return DisplayName{value: s}, nil
}

// MustNewDisplayName creates a DisplayName and panics if invalid.
// Only use for constants or tests.
func MustNewDisplayName(s string) DisplayName {
	name, err := NewDisplayName(s)
	if err != nil {
		panic(err)
	}
	return name
}

// String returns the display name string.
func (n DisplayName) String() string {
	return n.value
}

// IsZero returns true if the display name is empty.
func (n DisplayName) IsZero() bool {
	return n.value == ""
}

// IsValid returns true if the display name is valid (non-empty).
func (n DisplayName) IsValid() bool {
	return n.value != ""
}

// Equals checks if two display names are equal.
func (n DisplayName) Equals(other DisplayName) bool {
	return n.value == other.value
}

// Length returns the length in runes (Unicode characters).
func (n DisplayName) Length() int {
	return utf8.RuneCountInString(n.value)
}
