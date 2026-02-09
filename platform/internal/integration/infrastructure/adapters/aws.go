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
// AWS Adapter
// ============================================================================

// AWSAdapter implements the ProviderAdapter interface for AWS.
// Note: This uses AWS Cognito OAuth for user authentication.
// AWS certification verification would require additional API integration.
type AWSAdapter struct {
	config     *AWSConfig
	httpClient *http.Client
}

// NewAWSAdapter creates a new AWS adapter.
func NewAWSAdapter(config *AWSConfig) *AWSAdapter {
	return &AWSAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderType returns the provider type.
func (a *AWSAdapter) GetProviderType() domain.ProviderType {
	return domain.ProviderTypeAWS
}

// GetAuthorizationURL generates an OAuth authorization URL.
// Uses AWS Cognito hosted UI
func (a *AWSAdapter) GetAuthorizationURL(ctx context.Context, state string, redirectURI string, scopes []string) (string, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	if len(scopes) == 0 {
		scopes = a.GetDefaultScopes()
	}

	// AWS Cognito domain format: https://<domain-prefix>.auth.<region>.amazoncognito.com
	cognitoDomain := fmt.Sprintf("https://auth.%s.amazoncognito.com", a.config.Region)

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", a.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))

	authURL := fmt.Sprintf("%s/oauth2/authorize?%s", cognitoDomain, params.Encode())
	return authURL, nil
}

// ExchangeCode exchanges an authorization code for an access token.
func (a *AWSAdapter) ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	cognitoDomain := fmt.Sprintf("https://auth.%s.amazoncognito.com", a.config.Region)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/oauth2/token", cognitoDomain),
		strings.NewReader(data.Encode()))
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
		IDToken      string `json:"id_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// RefreshAccessToken refreshes an access token.
func (a *AWSAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error) {
	cognitoDomain := fmt.Sprintf("https://auth.%s.amazoncognito.com", a.config.Region)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/oauth2/token", cognitoDomain),
		strings.NewReader(data.Encode()))
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
		IDToken     string `json:"id_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// FetchUserData fetches user data from AWS Cognito.
func (a *AWSAdapter) FetchUserData(ctx context.Context, accessToken string) (*domain.ProviderData, error) {
	// Fetch user info from Cognito
	userInfo, err := a.fetchUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	// Note: AWS certification data would require separate AWS Training API integration
	// This implementation focuses on Cognito user profile data
	awsData := &domain.AWSData{
		UserID:    userInfo.Sub,
		UserName:  userInfo.Username,
		AccountID: userInfo.Sub, // Using sub as account ID for now
		Region:    a.config.Region,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		// Certifications would be populated by calling AWS Training API
		Certifications: []domain.AWSCertification{},
	}

	providerData := &domain.ProviderData{
		ProviderType:     domain.ProviderTypeAWS,
		ProviderUserID:   userInfo.Sub,
		ProviderUsername: userInfo.Username,
		Email:            userInfo.Email,
		Profile: domain.ProfileData{
			Name: userInfo.Name,
		},
		AWS:       awsData,
		FetchedAt: time.Now().UTC(),
		RawData: map[string]any{
			"userInfo": userInfo,
		},
	}

	return providerData, nil
}

// ValidateScopes validates the requested scopes.
func (a *AWSAdapter) ValidateScopes(scopes []string) error {
	validScopes := map[string]bool{
		"openid":                       true,
		"email":                        true,
		"profile":                      true,
		"phone":                        true,
		"aws.cognito.signin.user.admin": true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

// GetDefaultScopes returns the default scopes.
func (a *AWSAdapter) GetDefaultScopes() []string {
	return []string{"openid", "profile", "email"}
}

// RevokeAccess revokes the access token.
func (a *AWSAdapter) RevokeAccess(ctx context.Context, accessToken string) error {
	cognitoDomain := fmt.Sprintf("https://auth.%s.amazoncognito.com", a.config.Region)

	data := url.Values{}
	data.Set("token", accessToken)
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/oauth2/revoke", cognitoDomain),
		strings.NewReader(data.Encode()))
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
// AWS Cognito API Helpers
// ============================================================================

type cognitoUserInfo struct {
	Sub      string `json:"sub"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

func (a *AWSAdapter) fetchUserInfo(ctx context.Context, accessToken string) (*cognitoUserInfo, error) {
	cognitoDomain := fmt.Sprintf("https://auth.%s.amazoncognito.com", a.config.Region)

	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/oauth2/userInfo", cognitoDomain),
		nil)
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

	var userInfo cognitoUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
