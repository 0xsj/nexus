package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	integrationdomain "github.com/0xsj/nexus/platform/internal/integration/domain"
	verifdomain "github.com/0xsj/nexus/platform/internal/verification/domain"
)

// IntegrationBridge implements Verification's OAuthURLGenerator, OAuthCodeExchanger,
// and DataFetcher ports by delegating to Integration's ProviderAdapterRegistry.
type IntegrationBridge struct {
	registry integrationdomain.ProviderAdapterRegistry
}

// Compile-time checks.
var (
	_ verifdomain.OAuthURLGenerator  = (*IntegrationBridge)(nil)
	_ verifdomain.OAuthCodeExchanger = (*IntegrationBridge)(nil)
	_ verifdomain.DataFetcher        = (*IntegrationBridge)(nil)
)

// NewIntegrationBridge creates a new IntegrationBridge.
func NewIntegrationBridge(registry integrationdomain.ProviderAdapterRegistry) *IntegrationBridge {
	return &IntegrationBridge{registry: registry}
}

// GenerateAuthURL generates an OAuth authorization URL by delegating to the
// Integration context's provider adapter.
func (b *IntegrationBridge) GenerateAuthURL(ctx context.Context, provider verifdomain.ProviderType, state string, redirectURI string) (string, error) {
	adapter, err := b.registry.Get(integrationdomain.ProviderType(provider.String()))
	if err != nil {
		return "", fmt.Errorf("getting provider adapter for %s: %w", provider, err)
	}

	url, err := adapter.GetAuthorizationURL(ctx, state, redirectURI, nil)
	if err != nil {
		return "", fmt.Errorf("generating auth URL for %s: %w", provider, err)
	}

	return url, nil
}

// ExchangeCode exchanges an OAuth authorization code for an access token by
// delegating to the Integration context's provider adapter.
func (b *IntegrationBridge) ExchangeCode(ctx context.Context, provider verifdomain.ProviderType, code string, redirectURI string) (string, error) {
	adapter, err := b.registry.Get(integrationdomain.ProviderType(provider.String()))
	if err != nil {
		return "", fmt.Errorf("getting provider adapter for %s: %w", provider, err)
	}

	tokens, err := adapter.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return "", fmt.Errorf("exchanging code for %s: %w", provider, err)
	}

	return tokens.AccessToken, nil
}

// FetchData fetches user data from an external provider by delegating to the
// Integration context's provider adapter. Converts ProviderData to map[string]any
// via JSON round-trip for the Verification domain's generic interface.
func (b *IntegrationBridge) FetchData(ctx context.Context, provider verifdomain.ProviderType, accessToken string) (map[string]any, error) {
	adapter, err := b.registry.Get(integrationdomain.ProviderType(provider.String()))
	if err != nil {
		return nil, fmt.Errorf("getting provider adapter for %s: %w", provider, err)
	}

	providerData, err := adapter.FetchUserData(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("fetching data from %s: %w", provider, err)
	}

	// Convert ProviderData → map[string]any via JSON round-trip
	raw, err := json.Marshal(providerData)
	if err != nil {
		return nil, fmt.Errorf("marshaling provider data: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling provider data: %w", err)
	}

	return result, nil
}
