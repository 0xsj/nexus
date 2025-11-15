package password

import (
	"unicode"

	"github.com/0xsj/result"
)

// Validator validates password strength.
type Validator interface {
	// Validate checks if a password meets the strength requirements.
	Validate(password string) result.Result[struct{}]
}

type validator struct {
	rules ValidationRules
}

// NewValidator creates a new password validator with the given rules.
func NewValidator(rules ValidationRules) Validator {
	return &validator{rules: rules}
}

// Validate checks if a password meets the strength requirements.
func (v *validator) Validate(password string) result.Result[struct{}] {
	// Check if password is empty
	if password == "" {
		return result.Err[struct{}](ErrEmptyPassword)
	}

	// Check minimum length
	if len(password) < v.rules.MinLength {
		return result.Err[struct{}](ErrPasswordTooShort)
	}

	// Check uppercase requirement
	if v.rules.RequireUpper && !containsUpper(password) {
		return result.Err[struct{}](ErrMissingUppercase)
	}

	// Check lowercase requirement
	if v.rules.RequireLower && !containsLower(password) {
		return result.Err[struct{}](ErrMissingLowercase)
	}

	// Check number requirement
	if v.rules.RequireNumber && !containsNumber(password) {
		return result.Err[struct{}](ErrMissingNumber)
	}

	// Check special character requirement
	if v.rules.RequireSpecial && !containsSpecial(password) {
		return result.Err[struct{}](ErrMissingSpecialChar)
	}

	return result.Ok(struct{}{})
}

// containsUpper checks if the string contains at least one uppercase letter.
func containsUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// containsLower checks if the string contains at least one lowercase letter.
func containsLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

// containsNumber checks if the string contains at least one digit.
func containsNumber(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// containsSpecial checks if the string contains at least one special character.
func containsSpecial(s string) bool {
	for _, r := range s {
		// Special characters are non-alphanumeric printable characters
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return true
		}
	}
	return false
}
