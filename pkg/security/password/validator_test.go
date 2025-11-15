package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	assert.NotNil(t, validator)
}

func TestValidate_Success(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	validPasswords := []string{
		"SecureP@ssw0rd",
		"MyP@ssw0rd123",
		"C0mpl3x!Pass",
		"Str0ng&Password",
	}

	for _, password := range validPasswords {
		t.Run(password, func(t *testing.T) {
			result := validator.Validate(password)

			assert.True(t, result.IsOk(), "should validate strong password")
		})
	}
}

func TestValidate_EmptyPassword(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	result := validator.Validate("")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrEmptyPassword, result.UnwrapErr())
}

func TestValidate_TooShort(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	result := validator.Validate("Short1!")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrPasswordTooShort, result.UnwrapErr())
}

func TestValidate_MissingUppercase(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	result := validator.Validate("nouppercase123!")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrMissingUppercase, result.UnwrapErr())
}

func TestValidate_MissingLowercase(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	result := validator.Validate("NOLOWERCASE123!")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrMissingLowercase, result.UnwrapErr())
}

func TestValidate_MissingNumber(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	result := validator.Validate("NoNumbers!")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrMissingNumber, result.UnwrapErr())
}

func TestValidate_MissingSpecialChar(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	result := validator.Validate("NoSpecial123")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrMissingSpecialChar, result.UnwrapErr())
}

func TestValidate_WeakRules(t *testing.T) {
	rules := WeakValidationRules()
	validator := NewValidator(rules)

	// Password that would fail default rules
	password := "simple"

	result := validator.Validate(password)

	assert.True(t, result.IsOk(), "should validate with weak rules")
}

func TestValidate_StrongRules(t *testing.T) {
	rules := StrongValidationRules()
	validator := NewValidator(rules)

	t.Run("too short for strong rules", func(t *testing.T) {
		password := "Short1!"

		result := validator.Validate(password)

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrPasswordTooShort, result.UnwrapErr())
	})

	t.Run("meets strong rules", func(t *testing.T) {
		password := "VeryStr0ng&SecurePassword"

		result := validator.Validate(password)

		assert.True(t, result.IsOk(), "should validate strong password")
	})
}

func TestValidate_CustomRules(t *testing.T) {
	// Custom rules: 10 chars, no uppercase required
	rules := DefaultValidationRules().
		WithMinLength(10).
		WithRequireUpper(false)

	validator := NewValidator(rules)

	password := "lowercase123!"

	result := validator.Validate(password)

	assert.True(t, result.IsOk(), "should validate with custom rules")
}

func TestValidate_SpecialCharacters(t *testing.T) {
	rules := DefaultValidationRules()
	validator := NewValidator(rules)

	specialChars := []string{
		"Password123!",
		"Password123@",
		"Password123#",
		"Password123$",
		"Password123%",
		"Password123^",
		"Password123&",
		"Password123*",
		"Password123(",
		"Password123)",
		"Password123-",
		"Password123_",
		"Password123=",
		"Password123+",
	}

	for _, password := range specialChars {
		t.Run(password, func(t *testing.T) {
			result := validator.Validate(password)

			assert.True(t, result.IsOk(), "should accept various special characters")
		})
	}
}

func TestValidationRules_BuilderPattern(t *testing.T) {
	rules := DefaultValidationRules().
		WithMinLength(16).
		WithRequireUpper(false).
		WithRequireSpecial(false)

	assert.Equal(t, 16, rules.MinLength)
	assert.False(t, rules.RequireUpper)
	assert.False(t, rules.RequireSpecial)
	assert.True(t, rules.RequireLower)
	assert.True(t, rules.RequireNumber)
}
