package domain

import (
	"context"
)

// ============================================================================
// Provider Adapter Port
// ============================================================================

// ProviderAdapter is the port interface that all provider adapters must implement.
// Each adapter is responsible for:
// 1. OAuth authentication with the provider
// 2. Fetching user data from provider APIs
// 3. Normalizing data into ProviderData format
type ProviderAdapter interface {
	// GetProviderType returns the type of provider this adapter handles.
	GetProviderType() ProviderType

	// GetAuthorizationURL generates an OAuth authorization URL for the user to visit.
	// The state parameter is used for CSRF protection.
	// The redirectURI is where the provider will send the user after authorization.
	GetAuthorizationURL(ctx context.Context, state string, redirectURI string, scopes []string) (string, error)

	// ExchangeCode exchanges an OAuth authorization code for an access token.
	// Returns the access token, refresh token (if available), and expiration time.
	ExchangeCode(ctx context.Context, code string, redirectURI string) (*OAuthTokens, error)

	// RefreshAccessToken refreshes an expired access token using a refresh token.
	RefreshAccessToken(ctx context.Context, refreshToken string) (*OAuthTokens, error)

	// FetchUserData fetches and normalizes user data from the provider.
	// The accessToken is used to authenticate API requests.
	FetchUserData(ctx context.Context, accessToken string) (*ProviderData, error)

	// ValidateScopes checks if the requested scopes are valid for this provider.
	ValidateScopes(scopes []string) error

	// GetDefaultScopes returns the default scopes needed for basic user data.
	GetDefaultScopes() []string

	// RevokeAccess revokes the access token with the provider (if supported).
	RevokeAccess(ctx context.Context, accessToken string) error
}

// ============================================================================
// OAuth Tokens
// ============================================================================

// OAuthTokens represents OAuth tokens returned by a provider.
type OAuthTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64 // Seconds until expiration
	Scopes       []string
}

// ============================================================================
// Provider Adapter Registry
// ============================================================================

// ProviderAdapterRegistry manages the collection of registered provider adapters.
type ProviderAdapterRegistry interface {
	// Register registers a provider adapter.
	Register(adapter ProviderAdapter) error

	// Get retrieves a provider adapter by type.
	Get(providerType ProviderType) (ProviderAdapter, error)

	// Has checks if an adapter is registered for the given provider type.
	Has(providerType ProviderType) bool

	// GetAll returns all registered adapters.
	GetAll() []ProviderAdapter

	// GetSupportedProviders returns a list of all supported provider types.
	GetSupportedProviders() []ProviderType
}
