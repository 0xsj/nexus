// Package google provides Google OAuth 2.0 integration.
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/0xsj/result"
)

// Client handles Google OAuth 2.0 flow.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new Google OAuth client.
func NewClient(cfg *Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetAuthorizationURL generates the OAuth authorization URL.
func (c *Client) GetAuthorizationURL(state string, scopes []string) result.Result[string] {
	if state == "" {
		return result.Err[string](ErrInvalidState)
	}

	if len(scopes) == 0 {
		scopes = DefaultScopes
	}

	params := url.Values{
		"client_id":     {c.config.ClientID},
		"redirect_uri":  {c.config.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(scopes, " ")},
		"state":         {state},
		"access_type":   {"offline"}, // Request refresh token
		"prompt":        {"consent"}, // Force consent screen to get refresh token
	}

	authURL := fmt.Sprintf("%s?%s", GoogleAuthURL, params.Encode())
	return result.Ok(authURL)
}

// ExchangeCode exchanges authorization code for access token.
func (c *Client) ExchangeCode(ctx context.Context, code string) result.Result[*TokenResponse] {
	if code == "" {
		return result.Err[*TokenResponse](ErrInvalidCode)
	}

	data := url.Values{
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {c.config.RedirectURL},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		GoogleTokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return result.Err[*TokenResponse](err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result.Err[*TokenResponse](err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return result.Err[*TokenResponse](
			fmt.Errorf("google oauth token exchange failed: status %d, body: %s", resp.StatusCode, string(body)),
		)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return result.Err[*TokenResponse](err)
	}

	// Calculate expiration time
	if tokenResp.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		tokenResp.ExpiresAt = &expiresAt
	}

	return result.Ok(&tokenResp)
}

// GetUserInfo fetches user information from Google.
func (c *Client) GetUserInfo(ctx context.Context, accessToken string) result.Result[*UserInfo] {
	if accessToken == "" {
		return result.Err[*UserInfo](ErrInvalidAccessToken)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GoogleUserInfoURL, nil)
	if err != nil {
		return result.Err[*UserInfo](err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result.Err[*UserInfo](err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return result.Err[*UserInfo](
			fmt.Errorf("google userinfo request failed: status %d, body: %s", resp.StatusCode, string(body)),
		)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return result.Err[*UserInfo](err)
	}

	return result.Ok(&userInfo)
}

// RefreshToken refreshes an access token using a refresh token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) result.Result[*TokenResponse] {
	if refreshToken == "" {
		return result.Err[*TokenResponse](ErrInvalidRefreshToken)
	}

	data := url.Values{
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		GoogleTokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return result.Err[*TokenResponse](err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result.Err[*TokenResponse](err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return result.Err[*TokenResponse](
			fmt.Errorf("google token refresh failed: status %d, body: %s", resp.StatusCode, string(body)),
		)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return result.Err[*TokenResponse](err)
	}

	// Calculate expiration time
	if tokenResp.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		tokenResp.ExpiresAt = &expiresAt
	}

	return result.Ok(&tokenResp)
}
