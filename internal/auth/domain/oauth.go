package domain

import (
	"time"

	"github.com/google/uuid"
)

// OAuthProvider represents supported OAuth providers.
type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderApple  OAuthProvider = "apple"
	OAuthProviderGitHub OAuthProvider = "github"
)

// IsValid checks if the OAuth provider is supported.
func (p OAuthProvider) IsValid() bool {
	switch p {
	case OAuthProviderGoogle, OAuthProviderApple, OAuthProviderGitHub:
		return true
	default:
		return false
	}
}

// String returns the string representation of the provider.
func (p OAuthProvider) String() string {
	return string(p)
}

// OAuthConnection represents a user's OAuth provider connection.
// Note: This will need a migration to be created in the database.
type OAuthConnection struct {
	ID                string
	UserID            string
	Provider          OAuthProvider
	ProviderUserID    string
	ProviderEmail     string
	ProviderName      string
	ProviderAvatarURL string
	AccessToken       string
	RefreshToken      string
	TokenExpiresAt    *time.Time
	Scopes            []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LastUsedAt        *time.Time
}

// NewOAuthConnection creates a new OAuth connection.
func NewOAuthConnection(
	userID string,
	provider OAuthProvider,
	providerUserID string,
	providerEmail string,
	providerName string,
	accessToken string,
) *OAuthConnection {
	now := time.Now().UTC()

	return &OAuthConnection{
		ID:             uuid.New().String(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		ProviderEmail:  providerEmail,
		ProviderName:   providerName,
		AccessToken:    accessToken,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UpdateTokens updates the OAuth tokens.
func (o *OAuthConnection) UpdateTokens(accessToken, refreshToken string, expiresAt *time.Time) {
	o.AccessToken = accessToken
	o.RefreshToken = refreshToken
	o.TokenExpiresAt = expiresAt
	o.UpdatedAt = time.Now().UTC()
}

// MarkAsUsed updates the last used timestamp.
func (o *OAuthConnection) MarkAsUsed() {
	now := time.Now().UTC()
	o.LastUsedAt = &now
	o.UpdatedAt = now
}

// IsTokenExpired checks if the access token has expired.
func (o *OAuthConnection) IsTokenExpired() bool {
	if o.TokenExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*o.TokenExpiresAt)
}

// NeedsRefresh checks if the token needs to be refreshed.
func (o *OAuthConnection) NeedsRefresh() bool {
	return o.IsTokenExpired() && o.RefreshToken != ""
}

// OAuthState represents the OAuth state parameter for CSRF protection.
// This is stored in Redis/cache temporarily during OAuth flow.
type OAuthState struct {
	State     string
	Provider  OAuthProvider
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewOAuthState creates a new OAuth state.
func NewOAuthState(provider OAuthProvider, ttl time.Duration) *OAuthState {
	now := time.Now().UTC()

	return &OAuthState{
		State:     uuid.New().String(),
		Provider:  provider,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
}

// IsExpired checks if the OAuth state has expired.
func (s *OAuthState) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// IsValid checks if the OAuth state is valid (not expired).
func (s *OAuthState) IsValid() bool {
	return !s.IsExpired()
}
