package magic

import "time"

// Config holds configuration for magic link token generation.
type Config struct {
	// TokenLength is the number of random bytes to generate (default: 32)
	TokenLength int

	// LoginTTL is the time-to-live for login tokens (default: 15 minutes)
	LoginTTL time.Duration

	// VerifyTTL is the time-to-live for email verification tokens (default: 24 hours)
	VerifyTTL time.Duration

	// ResetTTL is the time-to-live for password reset tokens (default: 1 hour)
	ResetTTL time.Duration

	// BaseURL is the base URL for generating magic links (e.g., "https://app.example.com")
	BaseURL string

	// MaxTokensPerUser is the maximum number of active tokens per user per purpose (default: 3)
	// Used for rate limiting token generation
	MaxTokensPerUser int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		TokenLength:      32,
		LoginTTL:         15 * time.Minute,
		VerifyTTL:        24 * time.Hour,
		ResetTTL:         1 * time.Hour,
		BaseURL:          "",
		MaxTokensPerUser: 3,
	}
}

// WithTokenLength sets the token length and returns the config (builder pattern).
func (c Config) WithTokenLength(length int) Config {
	c.TokenLength = length
	return c
}

// WithLoginTTL sets the login token TTL and returns the config (builder pattern).
func (c Config) WithLoginTTL(ttl time.Duration) Config {
	c.LoginTTL = ttl
	return c
}

// WithVerifyTTL sets the email verification token TTL and returns the config (builder pattern).
func (c Config) WithVerifyTTL(ttl time.Duration) Config {
	c.VerifyTTL = ttl
	return c
}

// WithResetTTL sets the password reset token TTL and returns the config (builder pattern).
func (c Config) WithResetTTL(ttl time.Duration) Config {
	c.ResetTTL = ttl
	return c
}

// WithBaseURL sets the base URL and returns the config (builder pattern).
func (c Config) WithBaseURL(url string) Config {
	c.BaseURL = url
	return c
}

// WithMaxTokensPerUser sets the max tokens per user and returns the config (builder pattern).
func (c Config) WithMaxTokensPerUser(max int) Config {
	c.MaxTokensPerUser = max
	return c
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.TokenLength <= 0 {
		return ErrInvalidLength{Length: c.TokenLength}
	}

	if c.LoginTTL <= 0 {
		return ErrInvalidTTL{TTL: c.LoginTTL}
	}

	if c.VerifyTTL <= 0 {
		return ErrInvalidTTL{TTL: c.VerifyTTL}
	}

	if c.ResetTTL <= 0 {
		return ErrInvalidTTL{TTL: c.ResetTTL}
	}

	if c.MaxTokensPerUser <= 0 {
		return ErrInvalidMaxTokens{Max: c.MaxTokensPerUser}
	}

	// BaseURL is optional - can be empty if generating tokens without URLs

	return nil
}

// GetTTLForPurpose returns the TTL for a given purpose.
func (c Config) GetTTLForPurpose(purpose Purpose) time.Duration {
	switch purpose {
	case PurposeLogin:
		return c.LoginTTL
	case PurposeEmailVerify:
		return c.VerifyTTL
	case PurposePasswordReset:
		return c.ResetTTL
	default:
		return c.LoginTTL
	}
}
