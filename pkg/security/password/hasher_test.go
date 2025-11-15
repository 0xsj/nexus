package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewHasher_Success(t *testing.T) {
	config := DefaultConfig()

	result := NewHasher(config)

	require.True(t, result.IsOk(), "should create hasher successfully")
	hasher := result.Unwrap()
	assert.NotNil(t, hasher)
}

func TestNewHasher_InvalidConfig(t *testing.T) {
	t.Run("invalid bcrypt cost - too low", func(t *testing.T) {
		config := DefaultConfig().WithBcryptCost(3)

		result := NewHasher(config)

		assert.True(t, result.IsErr())
		_, ok := result.UnwrapErr().(ErrInvalidCost)
		assert.True(t, ok)
	})

	t.Run("invalid bcrypt cost - too high", func(t *testing.T) {
		config := DefaultConfig().WithBcryptCost(32)

		result := NewHasher(config)

		assert.True(t, result.IsErr())
		_, ok := result.UnwrapErr().(ErrInvalidCost)
		assert.True(t, ok)
	})
}

func TestHash_Success(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	password := "SecureP@ssw0rd"

	result := hasher.Hash(password)

	require.True(t, result.IsOk(), "should hash password successfully")
	hash := result.Unwrap()

	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "hash should not equal plaintext")
	assert.True(t, len(hash) > 50, "bcrypt hash should be long")
}

func TestHash_DifferentHashesForSamePassword(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	password := "SecureP@ssw0rd"

	// Hash the same password twice
	result1 := hasher.Hash(password)
	result2 := hasher.Hash(password)

	require.True(t, result1.IsOk())
	require.True(t, result2.IsOk())

	hash1 := result1.Unwrap()
	hash2 := result2.Unwrap()

	// Bcrypt includes a salt, so hashes should be different
	assert.NotEqual(t, hash1, hash2, "bcrypt should produce different hashes with different salts")
}

func TestHash_EmptyPassword(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	result := hasher.Hash("")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrEmptyPassword, result.UnwrapErr())
}

func TestHash_WeakPassword(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	weakPasswords := []struct {
		password    string
		expectedErr error
	}{
		{"short", ErrPasswordTooShort},
		{"alllowercase", ErrMissingUppercase},
		{"ALLUPPERCASE", ErrMissingLowercase},
		{"NoNumbers!", ErrMissingNumber},
		{"NoSpecial123", ErrMissingSpecialChar},
	}

	for _, tc := range weakPasswords {
		t.Run(tc.password, func(t *testing.T) {
			result := hasher.Hash(tc.password)

			assert.True(t, result.IsErr(), "should reject weak password")
			assert.Equal(t, tc.expectedErr, result.UnwrapErr())
		})
	}
}

func TestVerify_Success(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	password := "SecureP@ssw0rd"

	// Hash the password
	hashResult := hasher.Hash(password)
	require.True(t, hashResult.IsOk())
	hash := hashResult.Unwrap()

	// Verify correct password
	result := hasher.Verify(password, hash)

	require.True(t, result.IsOk())
	assert.True(t, result.Unwrap(), "should verify correct password")
}

func TestVerify_WrongPassword(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	password := "SecureP@ssw0rd"
	wrongPassword := "WrongP@ssw0rd"

	// Hash the correct password
	hashResult := hasher.Hash(password)
	require.True(t, hashResult.IsOk())
	hash := hashResult.Unwrap()

	// Verify with wrong password
	result := hasher.Verify(wrongPassword, hash)

	require.True(t, result.IsOk())
	assert.False(t, result.Unwrap(), "should reject wrong password")
}

func TestVerify_EmptyPassword(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	result := hasher.Verify("", "some-hash")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrEmptyPassword, result.UnwrapErr())
}

func TestVerify_EmptyHash(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	result := hasher.Verify("SomePassword123!", "")

	require.True(t, result.IsOk())
	assert.False(t, result.Unwrap(), "should return false for empty hash")
}

func TestVerify_InvalidHash(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	invalidHashes := []string{
		"not-a-bcrypt-hash",
		"$2a$12$invalid",
		"completely-wrong",
	}

	for _, hash := range invalidHashes {
		t.Run(hash, func(t *testing.T) {
			result := hasher.Verify("Password123!", hash)

			// Should error or return false
			if result.IsOk() {
				assert.False(t, result.Unwrap())
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

func TestNeedsRehash_SameCost(t *testing.T) {
	config := DefaultConfig().WithBcryptCost(10)
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	password := "SecureP@ssw0rd"

	// Hash with cost 10
	hashResult := hasher.Hash(password)
	require.True(t, hashResult.IsOk())
	hash := hashResult.Unwrap()

	// Check if needs rehash (should be false - same cost)
	needsRehash := hasher.NeedsRehash(hash)

	assert.False(t, needsRehash, "should not need rehash with same cost")
}

func TestNeedsRehash_DifferentCost(t *testing.T) {
	// Create hash with cost 10
	config10 := DefaultConfig().WithBcryptCost(10)
	hasher10Result := NewHasher(config10)
	require.True(t, hasher10Result.IsOk())
	hasher10 := hasher10Result.Unwrap()

	password := "SecureP@ssw0rd"
	hashResult := hasher10.Hash(password)
	require.True(t, hashResult.IsOk())
	hash := hashResult.Unwrap()

	// Check with hasher configured for cost 12
	config12 := DefaultConfig().WithBcryptCost(12)
	hasher12Result := NewHasher(config12)
	require.True(t, hasher12Result.IsOk())
	hasher12 := hasher12Result.Unwrap()

	needsRehash := hasher12.NeedsRehash(hash)

	assert.True(t, needsRehash, "should need rehash when cost is different")
}

func TestNeedsRehash_EmptyHash(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	needsRehash := hasher.NeedsRehash("")

	assert.False(t, needsRehash, "should return false for empty hash")
}

func TestNeedsRehash_InvalidHash(t *testing.T) {
	config := DefaultConfig()
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	needsRehash := hasher.NeedsRehash("not-a-valid-hash")

	assert.True(t, needsRehash, "should return true for invalid hash (to trigger regeneration)")
}

func TestDifferentBcryptCosts(t *testing.T) {
	costs := []int{4, 8, 10, 12}

	for _, cost := range costs {
		t.Run(string(rune(cost)), func(t *testing.T) {
			config := DefaultConfig().WithBcryptCost(cost)
			hasherResult := NewHasher(config)
			require.True(t, hasherResult.IsOk())
			hasher := hasherResult.Unwrap()

			password := "SecureP@ssw0rd"

			// Hash
			hashResult := hasher.Hash(password)
			require.True(t, hashResult.IsOk())
			hash := hashResult.Unwrap()

			// Verify
			verifyResult := hasher.Verify(password, hash)
			require.True(t, verifyResult.IsOk())
			assert.True(t, verifyResult.Unwrap())

			// Check cost
			actualCost, err := bcrypt.Cost([]byte(hash))
			require.NoError(t, err)
			assert.Equal(t, cost, actualCost)
		})
	}
}

func TestHash_WithCustomValidationRules(t *testing.T) {
	// Weak rules - only require 6 chars and lowercase
	weakRules := WeakValidationRules()
	config := DefaultConfig().WithValidationRules(weakRules)
	hasherResult := NewHasher(config)
	require.True(t, hasherResult.IsOk())
	hasher := hasherResult.Unwrap()

	// This would fail default rules but should pass weak rules
	password := "simple"

	result := hasher.Hash(password)

	require.True(t, result.IsOk(), "should hash password with weak rules")
	hash := result.Unwrap()
	assert.NotEmpty(t, hash)
}

func TestRehashWorkflow(t *testing.T) {
	// Simulate upgrading from cost 10 to cost 12

	// Step 1: User registered with old config (cost 10)
	oldConfig := DefaultConfig().WithBcryptCost(10)
	oldHasherResult := NewHasher(oldConfig)
	require.True(t, oldHasherResult.IsOk())
	oldHasher := oldHasherResult.Unwrap()

	password := "SecureP@ssw0rd"
	oldHashResult := oldHasher.Hash(password)
	require.True(t, oldHashResult.IsOk())
	oldHash := oldHashResult.Unwrap()

	// Step 2: System upgraded to cost 12
	newConfig := DefaultConfig().WithBcryptCost(12)
	newHasherResult := NewHasher(newConfig)
	require.True(t, newHasherResult.IsOk())
	newHasher := newHasherResult.Unwrap()

	// Step 3: User logs in - verify with old hash
	verifyResult := newHasher.Verify(password, oldHash)
	require.True(t, verifyResult.IsOk())
	assert.True(t, verifyResult.Unwrap(), "should still verify old hash")

	// Step 4: Check if needs rehash
	assert.True(t, newHasher.NeedsRehash(oldHash), "old hash should need rehash")

	// Step 5: Rehash password with new cost
	newHashResult := newHasher.Hash(password)
	require.True(t, newHashResult.IsOk())
	newHash := newHashResult.Unwrap()

	// Step 6: Verify new hash doesn't need rehash
	assert.False(t, newHasher.NeedsRehash(newHash), "new hash should not need rehash")

	// Step 7: Verify password works with new hash
	verifyNewResult := newHasher.Verify(password, newHash)
	require.True(t, verifyNewResult.IsOk())
	assert.True(t, verifyNewResult.Unwrap())
}

func TestConfig_BuilderPattern(t *testing.T) {
	config := DefaultConfig().
		WithBcryptCost(14).
		WithValidationRules(
			StrongValidationRules().
				WithMinLength(16).
				WithRequireSpecial(true),
		)

	assert.Equal(t, 14, config.BcryptCost)
	assert.Equal(t, 16, config.ValidationRules.MinLength)
	assert.True(t, config.ValidationRules.RequireSpecial)
}
