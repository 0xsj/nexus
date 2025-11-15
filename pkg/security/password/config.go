package password

// ValidationRules defines password strength requirements.
type ValidationRules struct {
	MinLength      int  // Minimum password length (default: 8)
	RequireUpper   bool // Require at least one uppercase letter (default: true)
	RequireLower   bool // Require at least one lowercase letter (default: true)
	RequireNumber  bool // Require at least one number (default: true)
	RequireSpecial bool // Require at least one special character (default: true)
}

// DefaultValidationRules returns secure default validation rules.
func DefaultValidationRules() ValidationRules {
	return ValidationRules{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
}

// WeakValidationRules returns lenient validation rules (not recommended for production).
func WeakValidationRules() ValidationRules {
	return ValidationRules{
		MinLength:      6,
		RequireUpper:   false,
		RequireLower:   true,
		RequireNumber:  false,
		RequireSpecial: false,
	}
}

// StrongValidationRules returns strict validation rules.
func StrongValidationRules() ValidationRules {
	return ValidationRules{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
}

// WithMinLength sets the minimum length and returns the rules (builder pattern).
func (r ValidationRules) WithMinLength(length int) ValidationRules {
	r.MinLength = length
	return r
}

// WithRequireUpper sets whether uppercase is required and returns the rules (builder pattern).
func (r ValidationRules) WithRequireUpper(require bool) ValidationRules {
	r.RequireUpper = require
	return r
}

// WithRequireLower sets whether lowercase is required and returns the rules (builder pattern).
func (r ValidationRules) WithRequireLower(require bool) ValidationRules {
	r.RequireLower = require
	return r
}

// WithRequireNumber sets whether numbers are required and returns the rules (builder pattern).
func (r ValidationRules) WithRequireNumber(require bool) ValidationRules {
	r.RequireNumber = require
	return r
}

// WithRequireSpecial sets whether special characters are required and returns the rules (builder pattern).
func (r ValidationRules) WithRequireSpecial(require bool) ValidationRules {
	r.RequireSpecial = require
	return r
}

// Config holds password hashing configuration.
type Config struct {
	// BcryptCost is the computational cost for bcrypt (4-31, default: 12)
	// Higher cost = more secure but slower
	BcryptCost int

	// ValidationRules defines password strength requirements
	ValidationRules ValidationRules
}

// DefaultConfig returns configuration with secure defaults.
func DefaultConfig() Config {
	return Config{
		BcryptCost:      12,
		ValidationRules: DefaultValidationRules(),
	}
}

// WithBcryptCost sets the bcrypt cost and returns the config (builder pattern).
func (c Config) WithBcryptCost(cost int) Config {
	c.BcryptCost = cost
	return c
}

// WithValidationRules sets the validation rules and returns the config (builder pattern).
func (c Config) WithValidationRules(rules ValidationRules) Config {
	c.ValidationRules = rules
	return c
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	// Bcrypt cost must be between 4 and 31
	if c.BcryptCost < 4 || c.BcryptCost > 31 {
		return ErrInvalidCost{Cost: c.BcryptCost}
	}

	// Min length must be at least 1
	if c.ValidationRules.MinLength < 1 {
		return ErrPasswordTooShort
	}

	return nil
}
