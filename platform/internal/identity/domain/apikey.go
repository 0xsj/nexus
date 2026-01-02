package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// ============================================================================
// API Key Entity
// ============================================================================

// APIKey represents an API key for programmatic access.
// The raw key is only available at creation time — we store only the hash.
type APIKey struct {
	id        string
	userID    string
	name      string
	keyHash   string
	keyPrefix string // First 8 chars for identification

	// Permissions
	scopes APIKeyScopes

	// Status
	status       APIKeyStatus
	revokedAt    time.Time
	revokedBy    string
	revokeReason string

	// Expiration
	expiresAt time.Time

	// Usage tracking
	lastUsedAt time.Time
	usageCount int64

	// Metadata
	description string
	createdAt   time.Time
	updatedAt   time.Time
}

// ============================================================================
// API Key Status
// ============================================================================

// APIKeyStatus represents the status of an API key.
type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusExpired APIKeyStatus = "expired"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
)

// String returns the string representation.
func (s APIKeyStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid.
func (s APIKeyStatus) IsValid() bool {
	switch s {
	case APIKeyStatusActive, APIKeyStatusExpired, APIKeyStatusRevoked:
		return true
	default:
		return false
	}
}

// IsUsable returns true if the key can be used.
func (s APIKeyStatus) IsUsable() bool {
	return s == APIKeyStatusActive
}

// ============================================================================
// Constructor
// ============================================================================

// APIKeyCreateResult contains the result of creating an API key.
// The raw key is only available here — it cannot be retrieved later.
type APIKeyCreateResult struct {
	APIKey *APIKey
	RawKey string // Only available at creation time!
}

// NewAPIKey creates a new API key.
// Returns the API key entity and the raw key (only available at creation).
func NewAPIKey(
	id string,
	userID string,
	name string,
	scopes APIKeyScopes,
	expiresAt time.Time,
	description string,
) (*APIKeyCreateResult, error) {
	// Generate raw key
	rawKey, err := generateRawAPIKey()
	if err != nil {
		return nil, err
	}

	// Hash the key
	keyHash := hashAPIKey(rawKey)
	keyPrefix := rawKey[:8]

	now := time.Now()

	apiKey := &APIKey{
		id:          id,
		userID:      userID,
		name:        name,
		keyHash:     keyHash,
		keyPrefix:   keyPrefix,
		scopes:      scopes,
		status:      APIKeyStatusActive,
		expiresAt:   expiresAt,
		description: description,
		createdAt:   now,
		updatedAt:   now,
	}

	return &APIKeyCreateResult{
		APIKey: apiKey,
		RawKey: rawKey,
	}, nil
}

// ReconstituteAPIKey creates an APIKey from persisted data.
func ReconstituteAPIKey(
	id string,
	userID string,
	name string,
	keyHash string,
	keyPrefix string,
	scopes APIKeyScopes,
	status APIKeyStatus,
	expiresAt time.Time,
	revokedAt time.Time,
	revokedBy string,
	revokeReason string,
	lastUsedAt time.Time,
	usageCount int64,
	description string,
	createdAt time.Time,
	updatedAt time.Time,
) *APIKey {
	return &APIKey{
		id:           id,
		userID:       userID,
		name:         name,
		keyHash:      keyHash,
		keyPrefix:    keyPrefix,
		scopes:       scopes,
		status:       status,
		expiresAt:    expiresAt,
		revokedAt:    revokedAt,
		revokedBy:    revokedBy,
		revokeReason: revokeReason,
		lastUsedAt:   lastUsedAt,
		usageCount:   usageCount,
		description:  description,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the API key ID.
func (k *APIKey) ID() string {
	return k.id
}

// UserID returns the user ID.
func (k *APIKey) UserID() string {
	return k.userID
}

// Name returns the key name.
func (k *APIKey) Name() string {
	return k.name
}

// KeyHash returns the key hash.
func (k *APIKey) KeyHash() string {
	return k.keyHash
}

// KeyPrefix returns the key prefix (for display/identification).
func (k *APIKey) KeyPrefix() string {
	return k.keyPrefix
}

// Scopes returns the key scopes.
func (k *APIKey) Scopes() APIKeyScopes {
	result := make(APIKeyScopes, len(k.scopes))
	copy(result, k.scopes)
	return result
}

// Status returns the key status.
func (k *APIKey) Status() APIKeyStatus {
	return k.status
}

// ExpiresAt returns when the key expires.
func (k *APIKey) ExpiresAt() time.Time {
	return k.expiresAt
}

// RevokedAt returns when the key was revoked.
func (k *APIKey) RevokedAt() time.Time {
	return k.revokedAt
}

// RevokedBy returns who revoked the key.
func (k *APIKey) RevokedBy() string {
	return k.revokedBy
}

// RevokeReason returns why the key was revoked.
func (k *APIKey) RevokeReason() string {
	return k.revokeReason
}

// LastUsedAt returns when the key was last used.
func (k *APIKey) LastUsedAt() time.Time {
	return k.lastUsedAt
}

// UsageCount returns how many times the key has been used.
func (k *APIKey) UsageCount() int64 {
	return k.usageCount
}

// Description returns the key description.
func (k *APIKey) Description() string {
	return k.description
}

// CreatedAt returns when the key was created.
func (k *APIKey) CreatedAt() time.Time {
	return k.createdAt
}

// UpdatedAt returns when the key was last updated.
func (k *APIKey) UpdatedAt() time.Time {
	return k.updatedAt
}

// ============================================================================
// Status Checks
// ============================================================================

// IsActive returns true if the key is active.
func (k *APIKey) IsActive() bool {
	return k.status == APIKeyStatusActive
}

// IsExpired returns true if the key has expired.
func (k *APIKey) IsExpired() bool {
	if k.expiresAt.IsZero() {
		return false // No expiration
	}
	return time.Now().After(k.expiresAt)
}

// IsRevoked returns true if the key was revoked.
func (k *APIKey) IsRevoked() bool {
	return k.status == APIKeyStatusRevoked
}

// IsUsable returns true if the key can be used.
func (k *APIKey) IsUsable() bool {
	return k.IsActive() && !k.IsExpired()
}

// TimeUntilExpiry returns the duration until expiration.
func (k *APIKey) TimeUntilExpiry() time.Duration {
	if k.expiresAt.IsZero() {
		return time.Duration(1<<63 - 1) // Max duration (no expiry)
	}
	return time.Until(k.expiresAt)
}

// HasScope returns true if the key has the specified scope.
func (k *APIKey) HasScope(scope APIKeyScope) bool {
	return k.scopes.Contains(scope)
}

// HasAnyScope returns true if the key has any of the specified scopes.
func (k *APIKey) HasAnyScope(scopes ...APIKeyScope) bool {
	return k.scopes.HasAny(scopes...)
}

// HasAllScopes returns true if the key has all of the specified scopes.
func (k *APIKey) HasAllScopes(scopes ...APIKeyScope) bool {
	return k.scopes.HasAll(scopes...)
}

// ============================================================================
// Commands
// ============================================================================

// RecordUsage records that the key was used.
func (k *APIKey) RecordUsage() {
	k.lastUsedAt = time.Now()
	k.usageCount++
}

// Revoke revokes the API key.
func (k *APIKey) Revoke(revokedBy, reason string) error {
	if k.status == APIKeyStatusRevoked {
		return nil // Already revoked
	}

	k.status = APIKeyStatusRevoked
	k.revokedAt = time.Now()
	k.revokedBy = revokedBy
	k.revokeReason = reason
	k.updatedAt = time.Now()
	return nil
}

// Expire marks the key as expired.
func (k *APIKey) Expire() {
	k.status = APIKeyStatusExpired
	k.updatedAt = time.Now()
}

// UpdateName updates the key name.
func (k *APIKey) UpdateName(name string) {
	k.name = name
	k.updatedAt = time.Now()
}

// UpdateDescription updates the key description.
func (k *APIKey) UpdateDescription(description string) {
	k.description = description
	k.updatedAt = time.Now()
}

// UpdateScopes updates the key scopes.
func (k *APIKey) UpdateScopes(scopes APIKeyScopes) {
	k.scopes = scopes
	k.updatedAt = time.Now()
}

// ExtendExpiration extends the expiration time.
func (k *APIKey) ExtendExpiration(newExpiresAt time.Time) error {
	if !k.IsActive() {
		return ErrAPIKeyRevoked("APIKey.ExtendExpiration", k.id)
	}

	k.expiresAt = newExpiresAt
	k.updatedAt = time.Now()
	return nil
}

// ============================================================================
// Validation
// ============================================================================

// Validate validates the API key for use.
func (k *APIKey) Validate() error {
	if k.status == APIKeyStatusRevoked {
		return ErrAPIKeyRevoked("APIKey.Validate", k.id)
	}

	if k.IsExpired() {
		return ErrAPIKeyExpired("APIKey.Validate", k.id)
	}

	return nil
}

// ValidateKey validates that the provided raw key matches.
func (k *APIKey) ValidateKey(rawKey string) error {
	if err := k.Validate(); err != nil {
		return err
	}

	if hashAPIKey(rawKey) != k.keyHash {
		return ErrAPIKeyNotFound("APIKey.ValidateKey", k.id)
	}

	return nil
}

// ValidateScope validates that the key has the required scope.
func (k *APIKey) ValidateScope(required APIKeyScope) error {
	if err := k.Validate(); err != nil {
		return err
	}

	if !k.HasScope(required) && !k.HasScope(APIScopeAdmin) {
		return ErrAPIKeyRevoked("APIKey.ValidateScope", k.id)
	}

	return nil
}

// ============================================================================
// Display
// ============================================================================

// MaskedKey returns a masked version of the key for display.
// Example: "nxs_abc1****"
func (k *APIKey) MaskedKey() string {
	return k.keyPrefix + "****"
}

// ============================================================================
// Key Generation
// ============================================================================

const (
	apiKeyPrefix = "nxs_" // Nexus API key prefix
	apiKeyLength = 32     // 32 bytes = 64 hex chars
)

// generateRawAPIKey generates a new raw API key.
func generateRawAPIKey() (string, error) {
	bytes := make([]byte, apiKeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return apiKeyPrefix + hex.EncodeToString(bytes), nil
}

// hashAPIKey hashes an API key for storage.
func hashAPIKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}

// ParseAPIKey extracts the prefix from a raw API key.
func ParseAPIKey(rawKey string) (prefix string, isValid bool) {
	if !strings.HasPrefix(rawKey, apiKeyPrefix) {
		return "", false
	}
	if len(rawKey) < 12 { // prefix (4) + at least 8 chars
		return "", false
	}
	return rawKey[len(apiKeyPrefix) : len(apiKeyPrefix)+8], true
}

// HashAPIKey hashes a raw API key (exported for use in queries).
func HashAPIKey(rawKey string) string {
	return hashAPIKey(rawKey)
}
