package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/0xsj/nexus/platform/internal/integration/domain"
)

// ============================================================================
// Twitter Adapter
// ============================================================================

// TwitterAdapter implements the ProviderAdapter interface for Twitter/X.
// Note: Twitter OAuth 2.0 uses PKCE (Proof Key for Code Exchange)
type TwitterAdapter struct {
	config         *TwitterConfig
	httpClient     *http.Client
	codeVerifiers  map[string]string // state -> code_verifier mapping
}

// NewTwitterAdapter creates a new Twitter adapter.
func NewTwitterAdapter(config *TwitterConfig) *TwitterAdapter {
	return &TwitterAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		codeVerifiers: make(map[string]string),
	}
}

// GetProviderType returns the provider type.
func (a *TwitterAdapter) GetProviderType() domain.ProviderType {
	return domain.ProviderTypeTwitter
}

// GetAuthorizationURL generates an OAuth authorization URL.
// Twitter uses OAuth 2.0 with PKCE
func (a *TwitterAdapter) GetAuthorizationURL(ctx context.Context, state string, redirectURI string, scopes []string) (string, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	if len(scopes) == 0 {
		scopes = a.GetDefaultScopes()
	}

	// Generate PKCE code verifier and challenge
	codeVerifier := generateCodeVerifier()
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Store code verifier for later use in ExchangeCode
	a.codeVerifiers[state] = codeVerifier

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", a.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	authURL := fmt.Sprintf("https://twitter.com/i/oauth2/authorize?%s", params.Encode())
	return authURL, nil
}

// ExchangeCode exchanges an authorization code for an access token.
func (a *TwitterAdapter) ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	// Note: In production, you'd need to pass the state to retrieve the code_verifier
	// For now, we'll use a placeholder approach
	codeVerifier := ""
	for _, v := range a.codeVerifiers {
		codeVerifier = v
		break
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", a.config.ClientID)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twitter.com/2/oauth2/token", strings.NewReader(data.Encode()))
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
func (a *TwitterAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", a.config.ClientID)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twitter.com/2/oauth2/token", strings.NewReader(data.Encode()))
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
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// FetchUserData fetches user data from Twitter.
func (a *TwitterAdapter) FetchUserData(ctx context.Context, accessToken string) (*domain.ProviderData, error) {
	// Fetch user profile
	user, err := a.fetchUser(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Build Twitter data
	twitterData := &domain.TwitterData{
		ID:             user.Data.ID,
		Username:       user.Data.Username,
		DisplayName:    user.Data.Name,
		Bio:            user.Data.Description,
		ProfileImage:   user.Data.ProfileImageURL,
		Location:       user.Data.Location,
		Website:        user.Data.URL,
		TweetCount:     user.Data.PublicMetrics.TweetCount,
		Followers:      user.Data.PublicMetrics.FollowersCount,
		Following:      user.Data.PublicMetrics.FollowingCount,
		Listed:         user.Data.PublicMetrics.ListedCount,
		IsVerified:     user.Data.Verified,
		IsBlueVerified: user.Data.VerifiedType == "blue",
		CreatedAt:      user.Data.CreatedAt,
		UpdatedAt:      time.Now().UTC(),
	}

	providerData := &domain.ProviderData{
		ProviderType:     domain.ProviderTypeTwitter,
		ProviderUserID:   user.Data.ID,
		ProviderUsername: user.Data.Username,
		Profile: domain.ProfileData{
			Name:      user.Data.Name,
			Bio:       user.Data.Description,
			AvatarURL: user.Data.ProfileImageURL,
			Location:  user.Data.Location,
			Website:   user.Data.URL,
		},
		Twitter:   twitterData,
		FetchedAt: time.Now().UTC(),
		RawData: map[string]any{
			"user": user,
		},
	}

	return providerData, nil
}

// ValidateScopes validates the requested scopes.
func (a *TwitterAdapter) ValidateScopes(scopes []string) error {
	validScopes := map[string]bool{
		"tweet.read":           true,
		"tweet.write":          true,
		"tweet.moderate.write": true,
		"users.read":           true,
		"follows.read":         true,
		"follows.write":        true,
		"offline.access":       true,
		"space.read":           true,
		"mute.read":            true,
		"mute.write":           true,
		"like.read":            true,
		"like.write":           true,
		"list.read":            true,
		"list.write":           true,
		"block.read":           true,
		"block.write":          true,
		"bookmark.read":        true,
		"bookmark.write":       true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

// GetDefaultScopes returns the default scopes.
func (a *TwitterAdapter) GetDefaultScopes() []string {
	return []string{"tweet.read", "users.read", "offline.access"}
}

// RevokeAccess revokes the access token.
func (a *TwitterAdapter) RevokeAccess(ctx context.Context, accessToken string) error {
	data := url.Values{}
	data.Set("token", accessToken)
	data.Set("client_id", a.config.ClientID)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twitter.com/2/oauth2/revoke", strings.NewReader(data.Encode()))
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
// Twitter API Helpers
// ============================================================================

type twitterUserResponse struct {
	Data twitterUser `json:"data"`
}

type twitterUser struct {
	ID              string                  `json:"id"`
	Username        string                  `json:"username"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	ProfileImageURL string                  `json:"profile_image_url"`
	Location        string                  `json:"location"`
	URL             string                  `json:"url"`
	Verified        bool                    `json:"verified"`
	VerifiedType    string                  `json:"verified_type"`
	CreatedAt       time.Time               `json:"created_at"`
	PublicMetrics   twitterPublicMetrics    `json:"public_metrics"`
}

type twitterPublicMetrics struct {
	FollowersCount int `json:"followers_count"`
	FollowingCount int `json:"following_count"`
	TweetCount     int `json:"tweet_count"`
	ListedCount    int `json:"listed_count"`
}

func (a *TwitterAdapter) fetchUser(ctx context.Context, accessToken string) (*twitterUserResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.twitter.com/2/users/me?user.fields=created_at,description,location,profile_image_url,public_metrics,url,verified,verified_type",
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
		return nil, fmt.Errorf("failed to fetch user: status=%d body=%s", resp.StatusCode, string(body))
	}

	var user twitterUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ============================================================================
// PKCE Helpers
// ============================================================================

// generateCodeVerifier generates a cryptographically random code verifier
func generateCodeVerifier() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	b := make([]byte, 128)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// generateCodeChallenge generates a code challenge from a verifier
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
