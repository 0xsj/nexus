package domain

import (
	"time"
)

// ============================================================================
// Connection Entity
// ============================================================================

// Connection represents a linked external identity provider connection.
// This is a domain entity — it doesn't know about OAuth tokens or API calls.
// Infrastructure layer handles the actual OAuth implementation.
type Connection struct {
	id             string
	userID         string
	provider       OAuthProvider
	providerUserID string
	status         ConnectionStatus

	// Profile data from provider (cached, provider-agnostic)
	displayName string
	email       string
	username    string
	avatarURL   string
	profileURL  string
	metadata    map[string]any

	// Timestamps
	connectedAt  time.Time
	updatedAt    time.Time
	lastSyncedAt time.Time
}

// ============================================================================
// Connection Status
// ============================================================================

// ConnectionStatus represents the status of a connection.
type ConnectionStatus string

const (
	ConnectionStatusActive   ConnectionStatus = "active"
	ConnectionStatusInactive ConnectionStatus = "inactive"
	ConnectionStatusRevoked  ConnectionStatus = "revoked"
	ConnectionStatusError    ConnectionStatus = "error"
)

// String returns the string representation.
func (s ConnectionStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid.
func (s ConnectionStatus) IsValid() bool {
	switch s {
	case ConnectionStatusActive, ConnectionStatusInactive, ConnectionStatusRevoked, ConnectionStatusError:
		return true
	default:
		return false
	}
}

// IsUsable returns true if the connection can be used.
func (s ConnectionStatus) IsUsable() bool {
	return s == ConnectionStatusActive
}

// ============================================================================
// Constructor
// ============================================================================

// NewConnection creates a new connection.
func NewConnection(
	id string,
	userID string,
	provider OAuthProvider,
	providerUserID string,
) *Connection {
	now := time.Now()
	return &Connection{
		id:             id,
		userID:         userID,
		provider:       provider,
		providerUserID: providerUserID,
		status:         ConnectionStatusActive,
		metadata:       make(map[string]any),
		connectedAt:    now,
		updatedAt:      now,
		lastSyncedAt:   now,
	}
}

// ReconstituteConnection creates a Connection from persisted data.
func ReconstituteConnection(
	id string,
	userID string,
	provider OAuthProvider,
	providerUserID string,
	status ConnectionStatus,
	displayName string,
	email string,
	username string,
	avatarURL string,
	profileURL string,
	metadata map[string]any,
	connectedAt time.Time,
	updatedAt time.Time,
	lastSyncedAt time.Time,
) *Connection {
	return &Connection{
		id:             id,
		userID:         userID,
		provider:       provider,
		providerUserID: providerUserID,
		status:         status,
		displayName:    displayName,
		email:          email,
		username:       username,
		avatarURL:      avatarURL,
		profileURL:     profileURL,
		metadata:       metadata,
		connectedAt:    connectedAt,
		updatedAt:      updatedAt,
		lastSyncedAt:   lastSyncedAt,
	}
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the connection ID.
func (c *Connection) ID() string {
	return c.id
}

// UserID returns the user ID.
func (c *Connection) UserID() string {
	return c.userID
}

// Provider returns the OAuth provider.
func (c *Connection) Provider() OAuthProvider {
	return c.provider
}

// ProviderUserID returns the user ID from the provider.
func (c *Connection) ProviderUserID() string {
	return c.providerUserID
}

// Status returns the connection status.
func (c *Connection) Status() ConnectionStatus {
	return c.status
}

// DisplayName returns the display name.
func (c *Connection) DisplayName() string {
	return c.displayName
}

// Email returns the email from the provider.
func (c *Connection) Email() string {
	return c.email
}

// Username returns the username from the provider.
func (c *Connection) Username() string {
	return c.username
}

// AvatarURL returns the avatar URL.
func (c *Connection) AvatarURL() string {
	return c.avatarURL
}

// ProfileURL returns the profile URL.
func (c *Connection) ProfileURL() string {
	return c.profileURL
}

// Metadata returns the provider-specific metadata.
func (c *Connection) Metadata() map[string]any {
	result := make(map[string]any, len(c.metadata))
	for k, v := range c.metadata {
		result[k] = v
	}
	return result
}

// ConnectedAt returns when the connection was created.
func (c *Connection) ConnectedAt() time.Time {
	return c.connectedAt
}

// UpdatedAt returns when the connection was last updated.
func (c *Connection) UpdatedAt() time.Time {
	return c.updatedAt
}

// LastSyncedAt returns when the profile was last synced.
func (c *Connection) LastSyncedAt() time.Time {
	return c.lastSyncedAt
}

// ============================================================================
// Status Checks
// ============================================================================

// IsActive returns true if the connection is active.
func (c *Connection) IsActive() bool {
	return c.status == ConnectionStatusActive
}

// IsUsable returns true if the connection can be used.
func (c *Connection) IsUsable() bool {
	return c.status.IsUsable()
}

// NeedsSync returns true if the profile data is stale.
func (c *Connection) NeedsSync(staleDuration time.Duration) bool {
	return time.Since(c.lastSyncedAt) > staleDuration
}

// ============================================================================
// Commands
// ============================================================================

// UpdateProfile updates the cached profile data.
func (c *Connection) UpdateProfile(profile ConnectionProfile) {
	c.displayName = profile.DisplayName
	c.email = profile.Email
	c.username = profile.Username
	c.avatarURL = profile.AvatarURL
	c.profileURL = profile.ProfileURL
	c.metadata = profile.Metadata
	c.updatedAt = time.Now()
	c.lastSyncedAt = time.Now()
}

// Activate activates the connection.
func (c *Connection) Activate() {
	c.status = ConnectionStatusActive
	c.updatedAt = time.Now()
}

// Deactivate deactivates the connection (tokens expired, needs reauth).
func (c *Connection) Deactivate() {
	c.status = ConnectionStatusInactive
	c.updatedAt = time.Now()
}

// MarkError marks the connection as having an error.
func (c *Connection) MarkError() {
	c.status = ConnectionStatusError
	c.updatedAt = time.Now()
}

// Revoke revokes the connection.
func (c *Connection) Revoke() {
	c.status = ConnectionStatusRevoked
	c.updatedAt = time.Now()
}

// ============================================================================
// Connection Profile (Value Object)
// ============================================================================

// ConnectionProfile represents profile data from a provider.
// Used to update a connection's cached profile.
type ConnectionProfile struct {
	DisplayName string
	Email       string
	Username    string
	AvatarURL   string
	ProfileURL  string
	Metadata    map[string]any
}

// ============================================================================
// Credential Claims
// ============================================================================

// ToCredentialClaims returns claims for credential issuance.
// Provider-specific claim extraction is done in infrastructure layer.
func (c *Connection) ToCredentialClaims() map[string]any {
	claims := map[string]any{
		"provider":         c.provider.String(),
		"provider_user_id": c.providerUserID,
		"connected_at":     c.connectedAt.UTC().Format(time.RFC3339),
	}

	if c.email != "" {
		claims["email"] = c.email
	}
	if c.username != "" {
		claims["username"] = c.username
	}
	if c.displayName != "" {
		claims["display_name"] = c.displayName
	}

	return claims
}
