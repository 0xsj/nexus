package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider Service
// ============================================================================

// ProviderService implements domain.ProviderService for OAuth providers.
type ProviderService struct {
	config   *Config
	adapters map[domain.Provider]ProviderAdapter
	client   *http.Client
	logger   log.Logger
}

// ProviderAdapter is the interface for provider-specific implementations.
type ProviderAdapter interface {
	// Provider returns the provider type.
	Provider() domain.Provider

	// GetAuthorizationURL returns the OAuth authorization URL.
	GetAuthorizationURL(state string, redirectURI string, scopes []string) (string, error)

	// ExchangeCode exchanges an authorization code for tokens.
	ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error)

	// RefreshTokens refreshes expired tokens.
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error)

	// FetchProfile fetches the user's profile.
	FetchProfile(ctx context.Context, accessToken string) (*domain.ProviderProfile, error)

	// FetchCredentialData fetches data for credential issuance.
	FetchCredentialData(ctx context.Context, credentialType domain.CredentialType, accessToken string, profile *domain.ProviderProfile) (*domain.CredentialData, error)

	// RevokeTokens revokes tokens with the provider.
	RevokeTokens(ctx context.Context, accessToken string) error

	// SupportedCredentialTypes returns the credential types this provider supports.
	SupportedCredentialTypes() []domain.CredentialType
}

// NewProviderService creates a new ProviderService.
func NewProviderService(config *Config, logger log.Logger) *ProviderService {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.DefaultTimeout,
	}

	ps := &ProviderService{
		config:   config,
		adapters: make(map[domain.Provider]ProviderAdapter),
		client:   client,
		logger:   logger,
	}

	// Register enabled providers
	ps.registerProviders()

	return ps
}

// registerProviders registers all enabled provider adapters.
func (ps *ProviderService) registerProviders() {
	// Register GitHub if enabled
	if ps.config.IsProviderEnabled(domain.ProviderGitHub) {
		ps.adapters[domain.ProviderGitHub] = NewGitHubAdapter(&ps.config.GitHub, ps.client, ps.logger)
	}

	// Register LinkedIn if enabled
	if ps.config.IsProviderEnabled(domain.ProviderLinkedIn) {
		ps.adapters[domain.ProviderLinkedIn] = NewLinkedInAdapter(&ps.config.LinkedIn, ps.client, ps.logger)
	}

	// Additional providers can be registered here
}

// ============================================================================
// Provider Service Interface Implementation
// ============================================================================

// GetAuthorizationURL returns the OAuth authorization URL for a provider.
func (ps *ProviderService) GetAuthorizationURL(ctx context.Context, params domain.AuthorizationURLParams) (string, error) {
	const op = "ProviderService.GetAuthorizationURL"

	adapter, err := ps.getAdapter(params.Provider)
	if err != nil {
		return "", err
	}

	redirectURI := params.RedirectURI
	if redirectURI == "" {
		redirectURI = ps.config.GetRedirectURL(params.Provider)
	}

	scopes := params.Scopes
	if len(scopes) == 0 {
		scopes = ps.config.GetScopes(params.Provider, "")
	}

	url, err := adapter.GetAuthorizationURL(params.State, redirectURI, scopes)
	if err != nil {
		return "", domain.ErrProviderAuthFailed(op, string(params.Provider), err.Error())
	}

	return url, nil
}

// ExchangeCode exchanges an authorization code for tokens.
func (ps *ProviderService) ExchangeCode(ctx context.Context, params domain.ExchangeCodeParams) (*domain.OAuthTokens, error) {
	const op = "ProviderService.ExchangeCode"

	adapter, err := ps.getAdapter(params.Provider)
	if err != nil {
		return nil, err
	}

	redirectURI := params.RedirectURI
	if redirectURI == "" {
		redirectURI = ps.config.GetRedirectURL(params.Provider)
	}

	tokens, err := adapter.ExchangeCode(ctx, params.Code, redirectURI)
	if err != nil {
		ps.logger.Error("failed to exchange code",
			log.String("provider", string(params.Provider)),
			log.Err(err),
		)
		return nil, domain.ErrProviderAuthFailed(op, string(params.Provider), "failed to exchange authorization code")
	}

	return tokens, nil
}

// RefreshTokens refreshes expired access tokens.
func (ps *ProviderService) RefreshTokens(ctx context.Context, params domain.RefreshTokensParams) (*domain.OAuthTokens, error) {
	const op = "ProviderService.RefreshTokens"

	adapter, err := ps.getAdapter(params.Provider)
	if err != nil {
		return nil, err
	}

	tokens, err := adapter.RefreshTokens(ctx, params.RefreshToken)
	if err != nil {
		ps.logger.Error("failed to refresh tokens",
			log.String("provider", string(params.Provider)),
			log.Err(err),
		)
		return nil, domain.ErrProviderTokenExpired(op, string(params.Provider))
	}

	return tokens, nil
}

// FetchProfile fetches the user's profile from the provider.
func (ps *ProviderService) FetchProfile(ctx context.Context, params domain.FetchProfileParams) (*domain.ProviderProfile, error) {
	const op = "ProviderService.FetchProfile"

	adapter, err := ps.getAdapter(params.Provider)
	if err != nil {
		return nil, err
	}

	profile, err := adapter.FetchProfile(ctx, params.AccessToken)
	if err != nil {
		ps.logger.Error("failed to fetch profile",
			log.String("provider", string(params.Provider)),
			log.Err(err),
		)
		return nil, domain.ErrProviderFetchFailed(op, string(params.Provider), err)
	}

	return profile, nil
}

// FetchCredentialData fetches data needed for credential issuance.
func (ps *ProviderService) FetchCredentialData(ctx context.Context, params domain.FetchCredentialDataParams) (*domain.CredentialData, error) {
	const op = "ProviderService.FetchCredentialData"

	adapter, err := ps.getAdapter(params.Provider)
	if err != nil {
		return nil, err
	}

	data, err := adapter.FetchCredentialData(ctx, params.CredentialType, params.AccessToken, params.Profile)
	if err != nil {
		ps.logger.Error("failed to fetch credential data",
			log.String("provider", string(params.Provider)),
			log.String("credential_type", string(params.CredentialType)),
			log.Err(err),
		)
		return nil, domain.ErrProviderFetchFailed(op, string(params.Provider), err)
	}

	return data, nil
}

// RevokeTokens revokes the user's tokens with the provider.
func (ps *ProviderService) RevokeTokens(ctx context.Context, provider domain.Provider, accessToken string) error {
	const op = "ProviderService.RevokeTokens"

	adapter, err := ps.getAdapter(provider)
	if err != nil {
		return err
	}

	if err := adapter.RevokeTokens(ctx, accessToken); err != nil {
		ps.logger.Warn("failed to revoke tokens (non-fatal)",
			log.String("provider", string(provider)),
			log.Err(err),
		)
		// Don't return error - token revocation failure is non-fatal
	}

	return nil
}

// SupportsProvider returns true if the provider is supported.
func (ps *ProviderService) SupportsProvider(provider domain.Provider) bool {
	_, exists := ps.adapters[provider]
	return exists
}

// ============================================================================
// Helper Methods
// ============================================================================

// getAdapter returns the adapter for a provider.
func (ps *ProviderService) getAdapter(provider domain.Provider) (ProviderAdapter, error) {
	adapter, exists := ps.adapters[provider]
	if !exists {
		return nil, domain.ErrProviderNotSupported("ProviderService.getAdapter", string(provider))
	}
	return adapter, nil
}

// ListProviders returns all available providers.
func (ps *ProviderService) ListProviders() []domain.Provider {
	providers := make([]domain.Provider, 0, len(ps.adapters))
	for p := range ps.adapters {
		providers = append(providers, p)
	}
	return providers
}

// ============================================================================
// OAuth State Service Implementation
// ============================================================================

// OAuthStateService implements domain.OAuthStateService.
type OAuthStateService struct {
	stateRepo domain.OAuthStateRepository
	config    *Config
	logger    log.Logger
}

// NewOAuthStateService creates a new OAuthStateService.
func NewOAuthStateService(stateRepo domain.OAuthStateRepository, config *Config, logger log.Logger) *OAuthStateService {
	return &OAuthStateService{
		stateRepo: stateRepo,
		config:    config,
		logger:    logger,
	}
}

// GenerateState generates a new OAuth state.
func (s *OAuthStateService) GenerateState(ctx context.Context, params domain.GenerateStateParams) (*domain.OAuthState, error) {
	const op = "OAuthStateService.GenerateState"

	// Generate random state value
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("%s: failed to generate state: %w", op, err)
	}
	stateValue := base64.URLEncoding.EncodeToString(stateBytes)

	// Calculate TTL
	ttl := s.config.StateExpiration
	if params.TTL > 0 {
		ttl = time.Duration(params.TTL) * time.Second
	}

	// Create state
	state := domain.NewOAuthState(
		stateValue,
		params.UserID,
		params.Provider,
		params.CredentialType,
		params.RedirectURL,
		ttl,
	)

	// Persist state
	if err := s.stateRepo.Save(ctx, &state); err != nil {
		return nil, fmt.Errorf("%s: failed to save state: %w", op, err)
	}

	return &state, nil
}

// ValidateState validates an OAuth state and returns the associated data.
func (s *OAuthStateService) ValidateState(ctx context.Context, value string) (*domain.OAuthState, error) {
	const op = "OAuthStateService.ValidateState"

	state, err := s.stateRepo.FindByValue(ctx, value)
	if err != nil {
		return nil, err
	}

	if state.IsExpired() {
		return nil, domain.ErrOAuthCodeExpired(op, string(state.Provider))
	}

	return state, nil
}

// ConsumeState validates and consumes an OAuth state (one-time use).
func (s *OAuthStateService) ConsumeState(ctx context.Context, value string) (*domain.OAuthState, error) {
	const op = "OAuthStateService.ConsumeState"

	state, err := s.ValidateState(ctx, value)
	if err != nil {
		return nil, err
	}

	// Delete the state (one-time use)
	if err := s.stateRepo.Delete(ctx, value); err != nil {
		s.logger.Warn("failed to delete OAuth state after consumption",
			log.String("state", value[:8]+"..."),
			log.Err(err),
		)
	}

	return state, nil
}

// InvalidateState invalidates an OAuth state without consuming it.
func (s *OAuthStateService) InvalidateState(ctx context.Context, value string) error {
	return s.stateRepo.Delete(ctx, value)
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ domain.ProviderService   = (*ProviderService)(nil)
	_ domain.OAuthStateService = (*OAuthStateService)(nil)
)
