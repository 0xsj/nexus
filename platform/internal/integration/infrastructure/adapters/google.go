package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/0xsj/nexus/platform/internal/integration/domain"
)

// ============================================================================
// Google Adapter
// ============================================================================

// GoogleAdapter implements the ProviderAdapter interface for Google.
type GoogleAdapter struct {
	config     *GoogleConfig
	httpClient *http.Client
}

// NewGoogleAdapter creates a new Google adapter.
func NewGoogleAdapter(config *GoogleConfig) *GoogleAdapter {
	return &GoogleAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderType returns the provider type.
func (a *GoogleAdapter) GetProviderType() domain.ProviderType {
	return domain.ProviderTypeGoogle
}

// GetAuthorizationURL generates an OAuth authorization URL.
func (a *GoogleAdapter) GetAuthorizationURL(ctx context.Context, state string, redirectURI string, scopes []string) (string, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	if len(scopes) == 0 {
		scopes = a.GetDefaultScopes()
	}

	params := url.Values{}
	params.Set("client_id", a.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)
	params.Set("access_type", "offline") // Request refresh token
	params.Set("prompt", "consent")      // Force consent to get refresh token

	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?%s", params.Encode())
	return authURL, nil
}

// ExchangeCode exchanges an authorization code for an access token.
func (a *GoogleAdapter) ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	scopes := []string{}
	if tokenResp.Scope != "" {
		scopes = strings.Split(tokenResp.Scope, " ")
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		Scopes:       scopes,
	}, nil
}

// RefreshAccessToken refreshes an access token.
func (a *GoogleAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error) {
	data := url.Values{}
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	scopes := []string{}
	if tokenResp.Scope != "" {
		scopes = strings.Split(tokenResp.Scope, " ")
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: refreshToken, // Google doesn't return a new refresh token
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		Scopes:       scopes,
	}, nil
}

// FetchUserData fetches user data from Google.
func (a *GoogleAdapter) FetchUserData(ctx context.Context, accessToken string) (*domain.ProviderData, error) {
	// Fetch user info
	userInfo, err := a.fetchUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	// Build Google data
	googleData := &domain.GoogleData{
		ID:            userInfo.Sub,
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		GivenName:     userInfo.GivenName,
		FamilyName:    userInfo.FamilyName,
		Picture:       userInfo.Picture,
		Locale:        userInfo.Locale,
		EmailVerified: userInfo.EmailVerified,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	providerData := &domain.ProviderData{
		ProviderType:     domain.ProviderTypeGoogle,
		ProviderUserID:   userInfo.Sub,
		ProviderUsername: userInfo.Email,
		Email:            userInfo.Email,
		Profile: domain.ProfileData{
			Name:      userInfo.Name,
			AvatarURL: userInfo.Picture,
		},
		Google:    googleData,
		FetchedAt: time.Now().UTC(),
		RawData: map[string]any{
			"userInfo": userInfo,
		},
	}

	return providerData, nil
}

// ValidateScopes validates the requested scopes.
func (a *GoogleAdapter) ValidateScopes(scopes []string) error {
	// Google has many scopes, validating against common ones
	// In production, you might want a more comprehensive list
	validPrefixes := []string{
		"https://www.googleapis.com/auth/",
		"openid",
		"profile",
		"email",
	}

	for _, scope := range scopes {
		valid := false
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(scope, prefix) || scope == prefix {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

// GetDefaultScopes returns the default scopes.
func (a *GoogleAdapter) GetDefaultScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"openid",
	}
}

// RevokeAccess revokes the access token.
func (a *GoogleAdapter) RevokeAccess(ctx context.Context, accessToken string) error {
	params := url.Values{}
	params.Set("token", accessToken)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://oauth2.googleapis.com/revoke",
		strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke access: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke failed: status=%d", resp.StatusCode)
	}

	return nil
}

// ============================================================================
// Google API Helpers
// ============================================================================

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
}

func (a *GoogleAdapter) fetchUserInfo(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user info: status=%d body=%s", resp.StatusCode, string(body))
	}

	var userInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
