package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// LinkedIn Adapter
// ============================================================================

// LinkedInAdapter implements ProviderAdapter for LinkedIn OAuth.
type LinkedInAdapter struct {
	config *ProviderConfig
	client *http.Client
	logger log.Logger
}

// NewLinkedInAdapter creates a new LinkedInAdapter.
func NewLinkedInAdapter(config *ProviderConfig, client *http.Client, logger log.Logger) *LinkedInAdapter {
	return &LinkedInAdapter{
		config: config,
		client: client,
		logger: logger,
	}
}

// Provider returns the provider type.
func (a *LinkedInAdapter) Provider() domain.Provider {
	return domain.ProviderLinkedIn
}

// ============================================================================
// OAuth Flow
// ============================================================================

// GetAuthorizationURL returns the LinkedIn OAuth authorization URL.
func (a *LinkedInAdapter) GetAuthorizationURL(state string, redirectURI string, scopes []string) (string, error) {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", a.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", a.config.AuthURL, params.Encode()), nil
}

// ExchangeCode exchanges an authorization code for tokens.
func (a *LinkedInAdapter) ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error) {
	const op = "LinkedInAdapter.ExchangeCode"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", op, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read response: %w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: unexpected status %d: %s", op, resp.StatusCode, string(body))
	}

	var tokenResp linkedInTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%s: failed to parse response: %w", op, err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%s: %s: %s", op, tokenResp.Error, tokenResp.ErrorDescription)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    expiresAt,
		Scopes:       strings.Split(tokenResp.Scope, " "),
	}, nil
}

// RefreshTokens refreshes expired tokens.
func (a *LinkedInAdapter) RefreshTokens(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error) {
	const op = "LinkedInAdapter.RefreshTokens"

	if refreshToken == "" {
		return nil, fmt.Errorf("%s: refresh token not available", op)
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", op, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read response: %w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: unexpected status %d: %s", op, resp.StatusCode, string(body))
	}

	var tokenResp linkedInTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%s: failed to parse response: %w", op, err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    expiresAt,
		Scopes:       strings.Split(tokenResp.Scope, " "),
	}, nil
}

// RevokeTokens revokes tokens with LinkedIn.
func (a *LinkedInAdapter) RevokeTokens(ctx context.Context, accessToken string) error {
	const op = "LinkedInAdapter.RevokeTokens"

	revokeURL := "https://www.linkedin.com/oauth/v2/revoke"

	data := url.Values{}
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("token", accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: request failed: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: unexpected status %d: %s", op, resp.StatusCode, string(body))
	}

	return nil
}

// ============================================================================
// Profile & Data Fetching
// ============================================================================

// FetchProfile fetches the user's LinkedIn profile.
func (a *LinkedInAdapter) FetchProfile(ctx context.Context, accessToken string) (*domain.ProviderProfile, error) {
	const op = "LinkedInAdapter.FetchProfile"

	// LinkedIn uses OpenID Connect userinfo endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", op, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read response: %w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: unexpected status %d: %s", op, resp.StatusCode, string(body))
	}

	var userInfo linkedInUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("%s: failed to parse response: %w", op, err)
	}

	// Build raw data map
	var rawData map[string]any
	_ = json.Unmarshal(body, &rawData)

	// Construct display name
	displayName := strings.TrimSpace(userInfo.GivenName + " " + userInfo.FamilyName)
	if displayName == "" {
		displayName = userInfo.Name
	}

	// Construct username from email or sub
	username := userInfo.Email
	if username == "" {
		username = userInfo.Sub
	}

	profile := domain.NewProviderProfile(domain.ProviderLinkedIn, userInfo.Sub, username).
		WithDisplayName(displayName).
		WithEmail(userInfo.Email).
		WithAvatarURL(userInfo.Picture).
		WithProfileURL(""). // LinkedIn doesn't provide profile URL in userinfo
		WithRawData(rawData)

	return &profile, nil
}

// FetchCredentialData fetches data for credential issuance.
func (a *LinkedInAdapter) FetchCredentialData(ctx context.Context, credentialType domain.CredentialType, accessToken string, profile *domain.ProviderProfile) (*domain.CredentialData, error) {
	const op = "LinkedInAdapter.FetchCredentialData"

	switch credentialType {
	case domain.CredentialTypeLinkedInEmployment:
		return a.fetchEmploymentData(ctx, accessToken, profile)
	case domain.CredentialTypeLinkedInAccount:
		return a.fetchAccountData(ctx, accessToken, profile)
	default:
		return nil, fmt.Errorf("%s: unsupported credential type: %s", op, credentialType)
	}
}

// SupportedCredentialTypes returns the credential types LinkedIn supports.
func (a *LinkedInAdapter) SupportedCredentialTypes() []domain.CredentialType {
	return []domain.CredentialType{
		domain.CredentialTypeLinkedInEmployment,
		domain.CredentialTypeLinkedInAccount,
	}
}

// ============================================================================
// Credential Data Fetching
// ============================================================================

// fetchEmploymentData fetches LinkedIn employment/position data.
func (a *LinkedInAdapter) fetchEmploymentData(ctx context.Context, accessToken string, profile *domain.ProviderProfile) (*domain.CredentialData, error) {
	// Note: LinkedIn's Member Data Portability API requires special approval
	// For now, we'll use the basic profile data available through OpenID Connect
	// In production, you would need to apply for the r_liteprofile or r_fullprofile scopes

	claims := map[string]any{
		"linkedin_user_id":  profile.ProviderUserID,
		"display_name":      profile.DisplayName,
		"email":             profile.Email,
		"email_verified":    profile.RawData["email_verified"],
		"locale":            profile.RawData["locale"],
		"verification_date": time.Now().UTC().Format(time.RFC3339),
	}

	// Add given_name and family_name if available
	if givenName, ok := profile.RawData["given_name"].(string); ok {
		claims["given_name"] = givenName
	}
	if familyName, ok := profile.RawData["family_name"].(string); ok {
		claims["family_name"] = familyName
	}

	evidence := []domain.CredentialEvidence{
		{
			Type:       "APIVerification",
			Source:     "LinkedIn API",
			VerifiedAt: time.Now().UTC().Format(time.RFC3339),
			Data: map[string]any{
				"endpoint": "api.linkedin.com",
				"method":   "OAuth2 OpenID Connect",
			},
		},
	}

	return &domain.CredentialData{
		Provider:       domain.ProviderLinkedIn,
		CredentialType: domain.CredentialTypeLinkedInEmployment,
		Claims:         claims,
		Evidence:       evidence,
		Profile:        profile,
	}, nil
}

// fetchAccountData fetches basic LinkedIn account data.
func (a *LinkedInAdapter) fetchAccountData(ctx context.Context, accessToken string, profile *domain.ProviderProfile) (*domain.CredentialData, error) {
	claims := map[string]any{
		"linkedin_user_id":  profile.ProviderUserID,
		"display_name":      profile.DisplayName,
		"email":             profile.Email,
		"email_verified":    profile.RawData["email_verified"],
		"avatar_url":        profile.AvatarURL,
		"verification_date": time.Now().UTC().Format(time.RFC3339),
	}

	// Add given_name and family_name if available
	if givenName, ok := profile.RawData["given_name"].(string); ok {
		claims["given_name"] = givenName
	}
	if familyName, ok := profile.RawData["family_name"].(string); ok {
		claims["family_name"] = familyName
	}

	evidence := []domain.CredentialEvidence{
		{
			Type:       "APIVerification",
			Source:     "LinkedIn API",
			VerifiedAt: time.Now().UTC().Format(time.RFC3339),
			Data: map[string]any{
				"endpoint": "api.linkedin.com",
				"method":   "OAuth2 OpenID Connect",
			},
		},
	}

	return &domain.CredentialData{
		Provider:       domain.ProviderLinkedIn,
		CredentialType: domain.CredentialTypeLinkedInAccount,
		Claims:         claims,
		Evidence:       evidence,
		Profile:        profile,
	}, nil
}

// ============================================================================
// Response Types
// ============================================================================

type linkedInTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	Scope            string `json:"scope"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type linkedInUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ ProviderAdapter = (*LinkedInAdapter)(nil)
