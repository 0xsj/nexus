package domain

import (
	"strings"
	"unicode"

	"github.com/0xsj/result"
	"golang.org/x/crypto/bcrypt"
)

// Password is a value object representing a hashed password.
type Password struct {
	hash string
}

// PasswordConstraints defines password validation rules.
type PasswordConstraints struct {
	MinLength          int
	MaxLength          int
	RequireUppercase   bool
	RequireLowercase   bool
	RequireNumber      bool
	RequireSpecialChar bool
	BCryptCost         int
}

// DefaultPasswordConstraints provides sensible defaults.
var DefaultPasswordConstraints = PasswordConstraints{
	MinLength:          8,
	MaxLength:          128,
	RequireUppercase:   true,
	RequireLowercase:   true,
	RequireNumber:      true,
	RequireSpecialChar: false,
	BCryptCost:         12, // bcrypt.DefaultCost is 10, 12 is more secure
}

// NewPassword creates a new Password by hashing the plain text password.
func NewPassword(plainPassword string) result.Result[Password] {
	return NewPasswordWithConstraints(plainPassword, DefaultPasswordConstraints)
}

// NewPasswordWithConstraints creates a new Password with custom constraints.
func NewPasswordWithConstraints(plainPassword string, constraints PasswordConstraints) result.Result[Password] {
	// Validate password strength
	if err := validatePasswordStrength(plainPassword, constraints); err != nil {
		return result.Err[Password](err)
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), constraints.BCryptCost)
	if err != nil {
		return result.Err[Password](ErrInvalidPassword("failed to hash password"))
	}

	return result.Ok(Password{hash: string(hash)})
}

// NewPasswordFromHash creates a Password from an existing hash.
// Used when loading from database.
func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

// Hash returns the bcrypt hash of the password.
func (p Password) Hash() string {
	return p.hash
}

// Compare checks if a plain text password matches this hashed password.
func (p Password) Compare(plainPassword string) result.Result[bool] {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plainPassword))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return result.Ok(false)
		}
		return result.Err[bool](ErrPasswordMismatch())
	}
	return result.Ok(true)
}

// validatePasswordStrength validates password against constraints.
func validatePasswordStrength(password string, constraints PasswordConstraints) error {
	// Check if empty
	if strings.TrimSpace(password) == "" {
		return ErrInvalidPassword("password cannot be empty")
	}

	// Check length
	if len(password) < constraints.MinLength {
		return ErrWeakPassword("password is too short")
	}

	if len(password) > constraints.MaxLength {
		return ErrInvalidPassword("password is too long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Check requirements
	if constraints.RequireUppercase && !hasUpper {
		return ErrWeakPassword("password must contain at least one uppercase letter")
	}

	if constraints.RequireLowercase && !hasLower {
		return ErrWeakPassword("password must contain at least one lowercase letter")
	}

	if constraints.RequireNumber && !hasNumber {
		return ErrWeakPassword("password must contain at least one number")
	}

	if constraints.RequireSpecialChar && !hasSpecial {
		return ErrWeakPassword("password must contain at least one special character")
	}

	return nil
}
