package request

import (
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
)

// ============================================================================
// Validator
// ============================================================================

// Validator accumulates validation errors.
type Validator struct {
	errors *pkghttp.ValidationErrors
}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{
		errors: pkghttp.NewValidationErrors(),
	}
}

// AddError adds a validation error.
func (v *Validator) AddError(field, message string) {
	v.errors.Add(field, message)
}

// HasErrors returns true if there are validation errors.
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// Errors returns the validation errors.
func (v *Validator) Errors() *pkghttp.ValidationErrors {
	return v.errors
}

// Error returns the validation errors as an error (nil if no errors).
func (v *Validator) Error() error {
	if !v.HasErrors() {
		return nil
	}
	return v.errors
}

// ============================================================================
// String Validations
// ============================================================================

// Required checks that a string is not empty.
func (v *Validator) Required(field, value string) bool {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
		return false
	}
	return true
}

// MinLength checks that a string has a minimum length.
func (v *Validator) MinLength(field, value string, min int) bool {
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, "must be at least "+strconv.Itoa(min)+" characters")
		return false
	}
	return true
}

// MaxLength checks that a string has a maximum length.
func (v *Validator) MaxLength(field, value string, max int) bool {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, "must be at most "+strconv.Itoa(max)+" characters")
		return false
	}
	return true
}

// LengthBetween checks that a string length is within a range.
func (v *Validator) LengthBetween(field, value string, min, max int) bool {
	length := utf8.RuneCountInString(value)
	if length < min || length > max {
		v.AddError(field, "must be between "+strconv.Itoa(min)+" and "+strconv.Itoa(max)+" characters")
		return false
	}
	return true
}

// Matches checks that a string matches a regex pattern.
func (v *Validator) Matches(field, value string, pattern *regexp.Regexp, message string) bool {
	if !pattern.MatchString(value) {
		if message == "" {
			message = "is invalid"
		}
		v.AddError(field, message)
		return false
	}
	return true
}

// OneOf checks that a value is one of the allowed values.
func (v *Validator) OneOf(field, value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	v.AddError(field, "must be one of: "+strings.Join(allowed, ", "))
	return false
}

// NotOneOf checks that a value is not one of the disallowed values.
func (v *Validator) NotOneOf(field, value string, disallowed []string) bool {
	for _, d := range disallowed {
		if value == d {
			v.AddError(field, "cannot be: "+d)
			return false
		}
	}
	return true
}

// ============================================================================
// Format Validations
// ============================================================================

// Email checks that a string is a valid email address.
func (v *Validator) Email(field, value string) bool {
	if value == "" {
		return true // Use Required() for empty check
	}

	_, err := mail.ParseAddress(value)
	if err != nil {
		v.AddError(field, "must be a valid email address")
		return false
	}
	return true
}

// URL checks that a string is a valid URL.
func (v *Validator) URL(field, value string) bool {
	if value == "" {
		return true
	}

	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		v.AddError(field, "must be a valid URL starting with http:// or https://")
		return false
	}
	return true
}

// UUID checks that a string is a valid UUID format.
func (v *Validator) UUID(field, value string) bool {
	if value == "" {
		return true
	}

	if !uuidPattern.MatchString(value) {
		v.AddError(field, "must be a valid UUID")
		return false
	}
	return true
}

// DID checks that a string is a valid DID format.
func (v *Validator) DID(field, value string) bool {
	if value == "" {
		return true
	}

	if !strings.HasPrefix(value, "did:") {
		v.AddError(field, "must be a valid DID (did:method:identifier)")
		return false
	}

	parts := strings.SplitN(value, ":", 3)
	if len(parts) < 3 || parts[1] == "" || parts[2] == "" {
		v.AddError(field, "must be a valid DID (did:method:identifier)")
		return false
	}

	return true
}

// Slug checks that a string is a valid slug (lowercase alphanumeric with hyphens).
func (v *Validator) Slug(field, value string) bool {
	if value == "" {
		return true
	}

	if !slugPattern.MatchString(value) {
		v.AddError(field, "must contain only lowercase letters, numbers, and hyphens")
		return false
	}
	return true
}

// AlphaNumeric checks that a string contains only alphanumeric characters.
func (v *Validator) AlphaNumeric(field, value string) bool {
	if value == "" {
		return true
	}

	if !alphaNumericPattern.MatchString(value) {
		v.AddError(field, "must contain only letters and numbers")
		return false
	}
	return true
}

// ============================================================================
// Number Validations
// ============================================================================

// Min checks that an integer is at least a minimum value.
func (v *Validator) Min(field string, value, min int) bool {
	if value < min {
		v.AddError(field, "must be at least "+strconv.Itoa(min))
		return false
	}
	return true
}

// Max checks that an integer is at most a maximum value.
func (v *Validator) Max(field string, value, max int) bool {
	if value > max {
		v.AddError(field, "must be at most "+strconv.Itoa(max))
		return false
	}
	return true
}

// Between checks that an integer is within a range.
func (v *Validator) Between(field string, value, min, max int) bool {
	if value < min || value > max {
		v.AddError(field, "must be between "+strconv.Itoa(min)+" and "+strconv.Itoa(max))
		return false
	}
	return true
}

// Positive checks that an integer is positive.
func (v *Validator) Positive(field string, value int) bool {
	if value <= 0 {
		v.AddError(field, "must be positive")
		return false
	}
	return true
}

// NonNegative checks that an integer is non-negative.
func (v *Validator) NonNegative(field string, value int) bool {
	if value < 0 {
		v.AddError(field, "must not be negative")
		return false
	}
	return true
}

// ============================================================================
// Time Validations
// ============================================================================

// NotInPast checks that a time is not in the past.
func (v *Validator) NotInPast(field string, value time.Time) bool {
	if value.Before(time.Now()) {
		v.AddError(field, "must not be in the past")
		return false
	}
	return true
}

// NotInFuture checks that a time is not in the future.
func (v *Validator) NotInFuture(field string, value time.Time) bool {
	if value.After(time.Now()) {
		v.AddError(field, "must not be in the future")
		return false
	}
	return true
}

// Before checks that a time is before another time.
func (v *Validator) Before(field string, value, before time.Time) bool {
	if !value.Before(before) {
		v.AddError(field, "must be before "+before.Format(time.RFC3339))
		return false
	}
	return true
}

// After checks that a time is after another time.
func (v *Validator) After(field string, value, after time.Time) bool {
	if !value.After(after) {
		v.AddError(field, "must be after "+after.Format(time.RFC3339))
		return false
	}
	return true
}

// ============================================================================
// Slice Validations
// ============================================================================

// NotEmpty checks that a slice is not empty.
func (v *Validator) NotEmpty(field string, value []string) bool {
	if len(value) == 0 {
		v.AddError(field, "must not be empty")
		return false
	}
	return true
}

// MaxItems checks that a slice has at most a maximum number of items.
func (v *Validator) MaxItems(field string, value []string, max int) bool {
	if len(value) > max {
		v.AddError(field, "must have at most "+strconv.Itoa(max)+" items")
		return false
	}
	return true
}

// MinItems checks that a slice has at least a minimum number of items.
func (v *Validator) MinItems(field string, value []string, min int) bool {
	if len(value) < min {
		v.AddError(field, "must have at least "+strconv.Itoa(min)+" items")
		return false
	}
	return true
}

// UniqueItems checks that all items in a slice are unique.
func (v *Validator) UniqueItems(field string, value []string) bool {
	seen := make(map[string]bool, len(value))
	for _, item := range value {
		if seen[item] {
			v.AddError(field, "must have unique items")
			return false
		}
		seen[item] = true
	}
	return true
}

// ============================================================================
// Map Validations
// ============================================================================

// MapNotEmpty checks that a map is not empty.
func (v *Validator) MapNotEmpty(field string, value map[string]any) bool {
	if len(value) == 0 {
		v.AddError(field, "must not be empty")
		return false
	}
	return true
}

// MapMaxKeys checks that a map has at most a maximum number of keys.
func (v *Validator) MapMaxKeys(field string, value map[string]any, max int) bool {
	if len(value) > max {
		v.AddError(field, "must have at most "+strconv.Itoa(max)+" keys")
		return false
	}
	return true
}

// ============================================================================
// Conditional Validations
// ============================================================================

// RequiredIf checks that a value is required if a condition is true.
func (v *Validator) RequiredIf(field, value string, condition bool) bool {
	if condition && strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
		return false
	}
	return true
}

// RequiredUnless checks that a value is required unless a condition is true.
func (v *Validator) RequiredUnless(field, value string, condition bool) bool {
	if !condition && strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
		return false
	}
	return true
}

// ============================================================================
// Custom Validation
// ============================================================================

// Custom runs a custom validation function.
func (v *Validator) Custom(field string, fn func() bool, message string) bool {
	if !fn() {
		v.AddError(field, message)
		return false
	}
	return true
}

// ============================================================================
// Patterns
// ============================================================================

var (
	uuidPattern         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	slugPattern         = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	alphaNumericPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)
