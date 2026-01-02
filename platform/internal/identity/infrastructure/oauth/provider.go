package oauth

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
)

// Provider defines the interface for OAuth providers.
type Provider interface {
	// Name returns the provider name.
	Name() domain.OAuthProvider

	// AuthURL returns the OAuth authorization URL.
	AuthURL(state string, scopes []string) string

	// Exchange exchanges an authorization code for tokens.
	Exchange(ctx context.Context, code string) (*Tokens, error)

	// Refresh refreshes the access token.
	Refresh(ctx context.Context, refreshToken string) (*Tokens, error)

	// FetchProfile fetches the user profile.
	FetchProfile(ctx context.Context, accessToken string) (*domain.ConnectionProfile, error)

	// RevokeToken revokes the access token.
	RevokeToken(ctx context.Context, accessToken string) error
}

// Tokens represents OAuth tokens.
type Tokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	Scopes       []string
}

// IsExpired returns true if the access token has expired.
func (t *Tokens) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// NeedsRefresh returns true if the token should be refreshed.
func (t *Tokens) NeedsRefresh() bool {
	return time.Until(t.ExpiresAt) < 5*time.Minute
}
