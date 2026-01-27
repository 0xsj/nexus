package oauth

import (
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
)

// ============================================================================
// OAuth Configuration
// ============================================================================

// Config holds configuration for all OAuth providers.
type Config struct {
	// Provider-specific configurations
	GitHub   ProviderConfig
	LinkedIn ProviderConfig
	Google   ProviderConfig
	Twitter  ProviderConfig
	Discord  ProviderConfig

	// Global settings
	DefaultTimeout  time.Duration
	StateExpiration time.Duration
	CallbackBaseURL string
}

// ProviderConfig holds configuration for a single OAuth provider.
type ProviderConfig struct {
	// Enabled indicates if this provider is enabled.
	Enabled bool

	// ClientID is the OAuth client ID.
	ClientID string

	// ClientSecret is the OAuth client secret.
	ClientSecret string

	// RedirectURL is the OAuth callback URL.
	RedirectURL string

	// Scopes are the default OAuth scopes to request.
	Scopes []string

	// AuthURL is the authorization endpoint.
	AuthURL string

	// TokenURL is the token endpoint.
	TokenURL string

	// UserInfoURL is the user info endpoint.
	UserInfoURL string

	// Timeout for HTTP requests to this provider.
	Timeout time.Duration
}

// ============================================================================
// Default Configurations
// ============================================================================

// DefaultConfig returns a configuration with default values.
func DefaultConfig() Config {
	return Config{
		DefaultTimeout:  30 * time.Second,
		StateExpiration: 10 * time.Minute,
		GitHub: ProviderConfig{
			AuthURL:     "https://github.com/login/oauth/authorize",
			TokenURL:    "https://github.com/login/oauth/access_token",
			UserInfoURL: "https://api.github.com/user",
			Scopes:      []string{"read:user", "user:email"},
			Timeout:     30 * time.Second,
		},
		LinkedIn: ProviderConfig{
			AuthURL:     "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL:    "https://www.linkedin.com/oauth/v2/accessToken",
			UserInfoURL: "https://api.linkedin.com/v2/userinfo",
			Scopes:      []string{"openid", "profile", "email"},
			Timeout:     30 * time.Second,
		},
		Google: ProviderConfig{
			AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:    "https://oauth2.googleapis.com/token",
			UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
			Scopes:      []string{"openid", "profile", "email"},
			Timeout:     30 * time.Second,
		},
		Twitter: ProviderConfig{
			AuthURL:     "https://twitter.com/i/oauth2/authorize",
			TokenURL:    "https://api.twitter.com/2/oauth2/token",
			UserInfoURL: "https://api.twitter.com/2/users/me",
			Scopes:      []string{"tweet.read", "users.read"},
			Timeout:     30 * time.Second,
		},
		Discord: ProviderConfig{
			AuthURL:     "https://discord.com/api/oauth2/authorize",
			TokenURL:    "https://discord.com/api/oauth2/token",
			UserInfoURL: "https://discord.com/api/users/@me",
			Scopes:      []string{"identify", "email"},
			Timeout:     30 * time.Second,
		},
	}
}

// ============================================================================
// Configuration Methods
// ============================================================================

// GetProviderConfig returns the configuration for a specific provider.
func (c *Config) GetProviderConfig(provider domain.Provider) (*ProviderConfig, error) {
	switch provider {
	case domain.ProviderGitHub:
		return &c.GitHub, nil
	case domain.ProviderLinkedIn:
		return &c.LinkedIn, nil
	case domain.ProviderGoogle:
		return &c.Google, nil
	case domain.ProviderTwitter:
		return &c.Twitter, nil
	case domain.ProviderDiscord:
		return &c.Discord, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

// IsProviderEnabled checks if a provider is enabled.
func (c *Config) IsProviderEnabled(provider domain.Provider) bool {
	cfg, err := c.GetProviderConfig(provider)
	if err != nil {
		return false
	}
	return cfg.Enabled && cfg.ClientID != "" && cfg.ClientSecret != ""
}

// EnabledProviders returns a list of enabled providers.
func (c *Config) EnabledProviders() []domain.Provider {
	var providers []domain.Provider
	for _, p := range domain.AllProviders() {
		if c.IsProviderEnabled(p) {
			providers = append(providers, p)
		}
	}
	return providers
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.CallbackBaseURL == "" {
		return fmt.Errorf("callback base URL is required")
	}

	// At least one provider must be enabled
	if len(c.EnabledProviders()) == 0 {
		return fmt.Errorf("at least one OAuth provider must be enabled")
	}

	// Validate enabled providers
	for _, p := range c.EnabledProviders() {
		cfg, _ := c.GetProviderConfig(p)
		if err := cfg.Validate(p); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates a provider configuration.
func (pc *ProviderConfig) Validate(provider domain.Provider) error {
	if !pc.Enabled {
		return nil
	}

	if pc.ClientID == "" {
		return fmt.Errorf("%s: client ID is required", provider)
	}

	if pc.ClientSecret == "" {
		return fmt.Errorf("%s: client secret is required", provider)
	}

	if pc.AuthURL == "" {
		return fmt.Errorf("%s: auth URL is required", provider)
	}

	if pc.TokenURL == "" {
		return fmt.Errorf("%s: token URL is required", provider)
	}

	return nil
}

// GetRedirectURL returns the redirect URL for a provider.
func (c *Config) GetRedirectURL(provider domain.Provider) string {
	cfg, err := c.GetProviderConfig(provider)
	if err != nil {
		return ""
	}

	if cfg.RedirectURL != "" {
		return cfg.RedirectURL
	}

	// Default: construct from base URL
	return fmt.Sprintf("%s/api/v1/verifications/callback", c.CallbackBaseURL)
}

// GetTimeout returns the timeout for a provider.
func (c *Config) GetTimeout(provider domain.Provider) time.Duration {
	cfg, err := c.GetProviderConfig(provider)
	if err != nil || cfg.Timeout == 0 {
		return c.DefaultTimeout
	}
	return cfg.Timeout
}

// ============================================================================
// Scopes
// ============================================================================

// GetScopes returns the scopes for a provider and credential type.
func (c *Config) GetScopes(provider domain.Provider, credentialType domain.CredentialType) []string {
	cfg, err := c.GetProviderConfig(provider)
	if err != nil {
		return nil
	}

	// Get base scopes
	scopes := make([]string, len(cfg.Scopes))
	copy(scopes, cfg.Scopes)

	// Add credential-type specific scopes
	additionalScopes := getAdditionalScopes(provider, credentialType)
	scopes = append(scopes, additionalScopes...)

	return uniqueScopes(scopes)
}

// getAdditionalScopes returns additional scopes needed for specific credential types.
func getAdditionalScopes(provider domain.Provider, credentialType domain.CredentialType) []string {
	switch provider {
	case domain.ProviderGitHub:
		switch credentialType {
		case domain.CredentialTypeGitHubContributor:
			return []string{"repo"} // Need repo access for contribution stats
		}
	case domain.ProviderLinkedIn:
		switch credentialType {
		case domain.CredentialTypeLinkedInEmployment:
			return []string{} // Basic profile scope is sufficient
		}
	}
	return nil
}

// uniqueScopes removes duplicate scopes.
func uniqueScopes(scopes []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
