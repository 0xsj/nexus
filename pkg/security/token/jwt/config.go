package jwt

import "time"

// Algorithm defines the signing algorithm for JWT tokens.
type Algorithm string

const (
	AlgorithmHS256 Algorithm = "HS256" // HMAC with SHA-256 (symmetric)
	AlgorithmHS384 Algorithm = "HS384" // HMAC with SHA-384 (symmetric)
	AlgorithmHS512 Algorithm = "HS512" // HMAC with SHA-512 (symmetric)
	AlgorithmRS256 Algorithm = "RS256" // RSA with SHA-256 (asymmetric)
	AlgorithmRS384 Algorithm = "RS384" // RSA with SHA-384 (asymmetric)
	AlgorithmRS512 Algorithm = "RS512" // RSA with SHA-512 (asymmetric)
	AlgorithmES256 Algorithm = "ES256" // ECDSA with SHA-256 (asymmetric)
	AlgorithmES384 Algorithm = "ES384" // ECDSA with SHA-384 (asymmetric)
	AlgorithmES512 Algorithm = "ES512" // ECDSA with SHA-512 (asymmetric)
)

// String returns the string representation of the algorithm.
func (a Algorithm) String() string {
	return string(a)
}

// IsSymmetric checks if the algorithm uses symmetric keys (HMAC).
func (a Algorithm) IsSymmetric() bool {
	switch a {
	case AlgorithmHS256, AlgorithmHS384, AlgorithmHS512:
		return true
	default:
		return false
	}
}

// IsAsymmetric checks if the algorithm uses asymmetric keys (RSA, ECDSA).
func (a Algorithm) IsAsymmetric() bool {
	return !a.IsSymmetric()
}

// Config holds JWT token configuration.
type Config struct {
	// Algorithm is the signing algorithm (default: HS256)
	Algorithm Algorithm

	// SecretKey is used for HMAC algorithms (HS256, HS384, HS512)
	SecretKey string

	// PrivateKeyPath is the path to the private key file for asymmetric algorithms
	PrivateKeyPath string

	// PublicKeyPath is the path to the public key file for asymmetric algorithms
	PublicKeyPath string

	// Issuer is the token issuer (typically your service name)
	Issuer string

	// Audience is the intended audience (typically your API/service)
	Audience string

	// AccessTokenTTL is the time-to-live for access tokens (default: 15 minutes)
	AccessTokenTTL time.Duration

	// RefreshTokenTTL is the time-to-live for refresh tokens (default: 7 days)
	RefreshTokenTTL time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Algorithm:       AlgorithmHS256,
		SecretKey:       "", // Must be set by user
		Issuer:          "nexus",
		Audience:        "nexus-api",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 7 days
	}
}

// WithAlgorithm sets the algorithm and returns the config (builder pattern).
func (c Config) WithAlgorithm(algo Algorithm) Config {
	c.Algorithm = algo
	return c
}

// WithSecretKey sets the secret key and returns the config (builder pattern).
func (c Config) WithSecretKey(key string) Config {
	c.SecretKey = key
	return c
}

// WithPrivateKeyPath sets the private key path and returns the config (builder pattern).
func (c Config) WithPrivateKeyPath(path string) Config {
	c.PrivateKeyPath = path
	return c
}

// WithPublicKeyPath sets the public key path and returns the config (builder pattern).
func (c Config) WithPublicKeyPath(path string) Config {
	c.PublicKeyPath = path
	return c
}

// WithIssuer sets the issuer and returns the config (builder pattern).
func (c Config) WithIssuer(issuer string) Config {
	c.Issuer = issuer
	return c
}

// WithAudience sets the audience and returns the config (builder pattern).
func (c Config) WithAudience(audience string) Config {
	c.Audience = audience
	return c
}

// WithAccessTokenTTL sets the access token TTL and returns the config (builder pattern).
func (c Config) WithAccessTokenTTL(ttl time.Duration) Config {
	c.AccessTokenTTL = ttl
	return c
}

// WithRefreshTokenTTL sets the refresh token TTL and returns the config (builder pattern).
func (c Config) WithRefreshTokenTTL(ttl time.Duration) Config {
	c.RefreshTokenTTL = ttl
	return c
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Algorithm == "" {
		return ErrInvalidAlgorithm
	}

	// Validate symmetric key setup
	if c.Algorithm.IsSymmetric() {
		if c.SecretKey == "" {
			return ErrInvalidSigningKey
		}
		if len(c.SecretKey) < 32 {
			return ErrWeakSecretKey{Length: len(c.SecretKey)}
		}
	}

	// Validate asymmetric key setup
	if c.Algorithm.IsAsymmetric() {
		if c.PrivateKeyPath == "" && c.PublicKeyPath == "" {
			return ErrMissingKeyPaths
		}
	}

	if c.Issuer == "" {
		return ErrMissingIssuer
	}

	if c.Audience == "" {
		return ErrMissingAudience
	}

	if c.AccessTokenTTL <= 0 {
		return ErrInvalidTTL{Name: "AccessTokenTTL", TTL: c.AccessTokenTTL}
	}

	if c.RefreshTokenTTL <= 0 {
		return ErrInvalidTTL{Name: "RefreshTokenTTL", TTL: c.RefreshTokenTTL}
	}

	if c.RefreshTokenTTL <= c.AccessTokenTTL {
		return ErrRefreshTTLTooShort{
			AccessTTL:  c.AccessTokenTTL,
			RefreshTTL: c.RefreshTokenTTL,
		}
	}

	return nil
}
