package password

import (
	"github.com/0xsj/result"
	"golang.org/x/crypto/bcrypt"
)

// Hasher handles password hashing and verification.
type Hasher interface {
	// Hash hashes a plaintext password.
	Hash(password string) result.Result[string]

	// Verify checks if a plaintext password matches a hash.
	Verify(password, hash string) result.Result[bool]

	// NeedsRehash checks if a hash was created with a different cost.
	// Useful for upgrading hashes when you increase bcrypt cost.
	NeedsRehash(hash string) bool
}

type hasher struct {
	config Config
}

// NewHasher creates a new password hasher with the given configuration.
func NewHasher(config Config) result.Result[Hasher] {
	// Validate config
	if err := config.Validate(); err != nil {
		return result.Err[Hasher](err)
	}

	return result.Ok[Hasher](&hasher{config: config})
}

// Hash hashes a plaintext password using bcrypt.
func (h *hasher) Hash(password string) result.Result[string] {
	if password == "" {
		return result.Err[string](ErrEmptyPassword)
	}

	// Validate password strength first
	validator := NewValidator(h.config.ValidationRules)
	validationResult := validator.Validate(password)
	if validationResult.IsErr() {
		return result.Err[string](validationResult.UnwrapErr())
	}

	// Hash the password
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		h.config.BcryptCost,
	)
	if err != nil {
		return result.Err[string](ErrHashingFailed{Err: err})
	}

	return result.Ok(string(hashedBytes))
}

// Verify checks if a plaintext password matches a bcrypt hash.
// Uses bcrypt's constant-time comparison to prevent timing attacks.
func (h *hasher) Verify(password, hash string) result.Result[bool] {
	if password == "" {
		return result.Err[bool](ErrEmptyPassword)
	}

	if hash == "" {
		return result.Ok(false)
	}

	// Compare using bcrypt (already constant-time)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			// Password doesn't match - not an error, just return false
			return result.Ok(false)
		}
		// Some other error (malformed hash, etc.)
		return result.Err[bool](err)
	}

	// Password matches
	return result.Ok(true)
}

// NeedsRehash checks if a hash was created with a different cost.
// Returns true if the hash should be regenerated with the current cost.
func (h *hasher) NeedsRehash(hash string) bool {
	if hash == "" {
		return false
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		// If we can't determine cost, assume it needs rehashing
		return true
	}

	return cost != h.config.BcryptCost
}
