package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
)

// Compile-time interface check.
var _ domain.OAuthService = (*OAuthService)(nil)

// OAuthProviderConfig holds OAuth settings for a single provider.
type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

// OAuthServiceConfig holds configuration for the OAuthService.
type OAuthServiceConfig struct {
	Providers map[string]OAuthProviderConfig // Keyed by provider name ("github", "google")
}

// OAuthService handles OAuth authorization flows using raw net/http.
type OAuthService struct {
	providers  map[string]OAuthProviderConfig
	httpClient *http.Client
}

// NewOAuthService creates a new OAuthService.
func NewOAuthService(cfg OAuthServiceConfig) *OAuthService {
	return &OAuthService{
		providers:  cfg.Providers,
		httpClient: &http.Client{},
	}
}

// GetAuthorizationURL returns the OAuth authorization URL for a provider.
func (s *OAuthService) GetAuthorizationURL(_ context.Context, provider string, state string, redirectURL string) (string, error) {
	cfg, ok := s.providers[provider]
	if !ok {
		return "", fmt.Errorf("oauth not configured for provider: %s", provider)
	}

	params := url.Values{
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {redirectURL},
		"response_type": {"code"},
		"state":         {state},
		"scope":         {strings.Join(cfg.Scopes, " ")},
	}

	return cfg.AuthURL + "?" + params.Encode(), nil
}

// ExchangeCode exchanges an authorization code for user profile data.
func (s *OAuthService) ExchangeCode(ctx context.Context, provider string, code string, redirectURL string) (*domain.OAuthProfile, error) {
	cfg, ok := s.providers[provider]
	if !ok {
		return nil, fmt.Errorf("oauth not configured for provider: %s", provider)
	}

	// Exchange code for access token.
	accessToken, err := s.exchangeToken(ctx, cfg, code, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("exchanging oauth code: %w", err)
	}

	// Fetch user profile.
	profile, err := s.fetchUserInfo(ctx, provider, cfg, accessToken)
	if err != nil {
		return nil, fmt.Errorf("fetching oauth user info: %w", err)
	}

	return profile, nil
}

// exchangeToken POSTs to the token endpoint to exchange a code for an access token.
func (s *OAuthService) exchangeToken(ctx context.Context, cfg OAuthProviderConfig, code string, redirectURL string) (string, error) {
	data := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURL},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	return tokenResp.AccessToken, nil
}

// fetchUserInfo GETs the user info endpoint and maps the response to an OAuthProfile.
func (s *OAuthService) fetchUserInfo(ctx context.Context, provider string, cfg OAuthProviderConfig, accessToken string) (*domain.OAuthProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading user info response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	switch provider {
	case "github":
		return s.parseGitHubProfile(body)
	case "google":
		return s.parseGoogleProfile(body)
	default:
		return nil, fmt.Errorf("unsupported oauth provider for profile parsing: %s", provider)
	}
}

// parseGitHubProfile maps GitHub's /user response to an OAuthProfile.
func (s *OAuthService) parseGitHubProfile(body []byte) (*domain.OAuthProfile, error) {
	var gh struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &gh); err != nil {
		return nil, fmt.Errorf("parsing github profile: %w", err)
	}

	return &domain.OAuthProfile{
		Provider:   "github",
		ExternalID: fmt.Sprintf("%d", gh.ID),
		Email:      gh.Email,
		Name:       gh.Name,
		AvatarURL:  gh.AvatarURL,
	}, nil
}

// parseGoogleProfile maps Google's /oauth2/v2/userinfo response to an OAuthProfile.
func (s *OAuthService) parseGoogleProfile(body []byte) (*domain.OAuthProfile, error) {
	var g struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &g); err != nil {
		return nil, fmt.Errorf("parsing google profile: %w", err)
	}

	return &domain.OAuthProfile{
		Provider:   "google",
		ExternalID: g.ID,
		Email:      g.Email,
		Name:       g.Name,
		AvatarURL:  g.Picture,
	}, nil
}
