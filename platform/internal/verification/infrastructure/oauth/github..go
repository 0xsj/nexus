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
// GitHub Adapter
// ============================================================================

// GitHubAdapter implements ProviderAdapter for GitHub OAuth.
type GitHubAdapter struct {
	config *ProviderConfig
	client *http.Client
	logger log.Logger
}

// NewGitHubAdapter creates a new GitHubAdapter.
func NewGitHubAdapter(config *ProviderConfig, client *http.Client, logger log.Logger) *GitHubAdapter {
	return &GitHubAdapter{
		config: config,
		client: client,
		logger: logger,
	}
}

// Provider returns the provider type.
func (a *GitHubAdapter) Provider() domain.Provider {
	return domain.ProviderGitHub
}

// ============================================================================
// OAuth Flow
// ============================================================================

// GetAuthorizationURL returns the GitHub OAuth authorization URL.
func (a *GitHubAdapter) GetAuthorizationURL(state string, redirectURI string, scopes []string) (string, error) {
	params := url.Values{}
	params.Set("client_id", a.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)
	params.Set("allow_signup", "false")

	return fmt.Sprintf("%s?%s", a.config.AuthURL, params.Encode()), nil
}

// ExchangeCode exchanges an authorization code for tokens.
func (a *GitHubAdapter) ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error) {
	const op = "GitHubAdapter.ExchangeCode"

	data := url.Values{}
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

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

	var tokenResp gitHubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%s: failed to parse response: %w", op, err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%s: %s: %s", op, tokenResp.Error, tokenResp.ErrorDescription)
	}

	// GitHub tokens don't expire by default, but we set a reasonable time
	expiresAt := time.Now().Add(8760 * time.Hour) // 1 year
	if tokenResp.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    expiresAt,
		Scopes:       strings.Split(tokenResp.Scope, ","),
	}, nil
}

// RefreshTokens refreshes expired tokens.
func (a *GitHubAdapter) RefreshTokens(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error) {
	const op = "GitHubAdapter.RefreshTokens"

	// GitHub doesn't support token refresh for most OAuth apps
	// Only GitHub Apps with expiring tokens support this
	if refreshToken == "" {
		return nil, fmt.Errorf("%s: refresh token not available", op)
	}

	data := url.Values{}
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

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

	var tokenResp gitHubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%s: failed to parse response: %w", op, err)
	}

	expiresAt := time.Now().Add(8760 * time.Hour)
	if tokenResp.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    expiresAt,
		Scopes:       strings.Split(tokenResp.Scope, ","),
	}, nil
}

// RevokeTokens revokes tokens with GitHub.
func (a *GitHubAdapter) RevokeTokens(ctx context.Context, accessToken string) error {
	const op = "GitHubAdapter.RevokeTokens"

	// GitHub uses DELETE to revoke OAuth tokens
	revokeURL := fmt.Sprintf("https://api.github.com/applications/%s/token", a.config.ClientID)

	reqBody := fmt.Sprintf(`{"access_token":"%s"}`, accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, revokeURL, strings.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: request failed: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: unexpected status %d: %s", op, resp.StatusCode, string(body))
	}

	return nil
}

// ============================================================================
// Profile & Data Fetching
// ============================================================================

// FetchProfile fetches the user's GitHub profile.
func (a *GitHubAdapter) FetchProfile(ctx context.Context, accessToken string) (*domain.ProviderProfile, error) {
	const op = "GitHubAdapter.FetchProfile"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

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

	var user gitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("%s: failed to parse response: %w", op, err)
	}

	// Fetch email if not public
	email := user.Email
	if email == "" {
		email, _ = a.fetchPrimaryEmail(ctx, accessToken)
	}

	// Build raw data map
	var rawData map[string]any
	_ = json.Unmarshal(body, &rawData)

	profile := domain.NewProviderProfile(domain.ProviderGitHub, fmt.Sprintf("%d", user.ID), user.Login).
		WithDisplayName(user.Name).
		WithEmail(email).
		WithAvatarURL(user.AvatarURL).
		WithProfileURL(user.HTMLURL).
		WithRawData(rawData)

	return &profile, nil
}

// FetchCredentialData fetches data for credential issuance.
func (a *GitHubAdapter) FetchCredentialData(ctx context.Context, credentialType domain.CredentialType, accessToken string, profile *domain.ProviderProfile) (*domain.CredentialData, error) {
	const op = "GitHubAdapter.FetchCredentialData"

	switch credentialType {
	case domain.CredentialTypeGitHubContributor:
		return a.fetchContributorData(ctx, accessToken, profile)
	case domain.CredentialTypeGitHubAccount:
		return a.fetchAccountData(ctx, accessToken, profile)
	default:
		return nil, fmt.Errorf("%s: unsupported credential type: %s", op, credentialType)
	}
}

// SupportedCredentialTypes returns the credential types GitHub supports.
func (a *GitHubAdapter) SupportedCredentialTypes() []domain.CredentialType {
	return []domain.CredentialType{
		domain.CredentialTypeGitHubContributor,
		domain.CredentialTypeGitHubAccount,
	}
}

// ============================================================================
// Credential Data Fetching
// ============================================================================

// fetchContributorData fetches GitHub contribution statistics.
func (a *GitHubAdapter) fetchContributorData(ctx context.Context, accessToken string, profile *domain.ProviderProfile) (*domain.CredentialData, error) {
	const op = "GitHubAdapter.fetchContributorData"

	// Fetch user repos
	repos, err := a.fetchUserRepos(ctx, accessToken)
	if err != nil {
		a.logger.Warn("failed to fetch repos", log.Err(err))
		repos = []gitHubRepo{}
	}

	// Calculate statistics
	totalRepos := len(repos)
	totalStars := 0
	totalForks := 0
	languages := make(map[string]int)

	for _, repo := range repos {
		totalStars += repo.StargazersCount
		totalForks += repo.ForksCount
		if repo.Language != "" {
			languages[repo.Language]++
		}
	}

	// Find top languages
	topLanguages := getTopLanguages(languages, 5)

	// Build claims
	claims := map[string]any{
		"github_username":   profile.Username,
		"github_user_id":    profile.ProviderUserID,
		"public_repos":      totalRepos,
		"total_stars":       totalStars,
		"total_forks":       totalForks,
		"top_languages":     topLanguages,
		"account_created":   profile.RawData["created_at"],
		"profile_url":       profile.ProfileURL,
		"verification_date": time.Now().UTC().Format(time.RFC3339),
	}

	// Add follower count if available
	if followers, ok := profile.RawData["followers"].(float64); ok {
		claims["followers"] = int(followers)
	}

	// Build evidence
	evidence := []domain.CredentialEvidence{
		{
			Type:       "APIVerification",
			Source:     "GitHub API",
			VerifiedAt: time.Now().UTC().Format(time.RFC3339),
			Data: map[string]any{
				"endpoint": "api.github.com",
				"method":   "OAuth2",
			},
		},
	}

	return &domain.CredentialData{
		Provider:       domain.ProviderGitHub,
		CredentialType: domain.CredentialTypeGitHubContributor,
		Claims:         claims,
		Evidence:       evidence,
		Profile:        profile,
	}, nil
}

// fetchAccountData fetches basic GitHub account data.
func (a *GitHubAdapter) fetchAccountData(ctx context.Context, accessToken string, profile *domain.ProviderProfile) (*domain.CredentialData, error) {
	claims := map[string]any{
		"github_username":   profile.Username,
		"github_user_id":    profile.ProviderUserID,
		"display_name":      profile.DisplayName,
		"email":             profile.Email,
		"profile_url":       profile.ProfileURL,
		"avatar_url":        profile.AvatarURL,
		"account_created":   profile.RawData["created_at"],
		"verification_date": time.Now().UTC().Format(time.RFC3339),
	}

	evidence := []domain.CredentialEvidence{
		{
			Type:       "APIVerification",
			Source:     "GitHub API",
			VerifiedAt: time.Now().UTC().Format(time.RFC3339),
			Data: map[string]any{
				"endpoint": "api.github.com",
				"method":   "OAuth2",
			},
		},
	}

	return &domain.CredentialData{
		Provider:       domain.ProviderGitHub,
		CredentialType: domain.CredentialTypeGitHubAccount,
		Claims:         claims,
		Evidence:       evidence,
		Profile:        profile,
	}, nil
}

// ============================================================================
// API Helpers
// ============================================================================

// fetchPrimaryEmail fetches the user's primary email.
func (a *GitHubAdapter) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var emails []gitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	return "", nil
}

// fetchUserRepos fetches the user's repositories.
func (a *GitHubAdapter) fetchUserRepos(ctx context.Context, accessToken string) ([]gitHubRepo, error) {
	var allRepos []gitHubRepo
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/user/repos?per_page=%d&page=%d&sort=updated", perPage, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := a.client.Do(req)
		if err != nil {
			return nil, err
		}

		var repos []gitHubRepo
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if len(repos) == 0 {
			break
		}

		allRepos = append(allRepos, repos...)

		if len(repos) < perPage {
			break
		}

		page++

		// Limit to first 500 repos
		if len(allRepos) >= 500 {
			break
		}
	}

	return allRepos, nil
}

// ============================================================================
// Response Types
// ============================================================================

type gitHubTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type gitHubUser struct {
	ID          int64  `json:"id"`
	Login       string `json:"login"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	HTMLURL     string `json:"html_url"`
	CreatedAt   string `json:"created_at"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	PublicRepos int    `json:"public_repos"`
}

type gitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type gitHubRepo struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	FullName        string `json:"full_name"`
	Private         bool   `json:"private"`
	Fork            bool   `json:"fork"`
	Language        string `json:"language"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// ============================================================================
// Helpers
// ============================================================================

// getTopLanguages returns the top N languages by repo count.
func getTopLanguages(languages map[string]int, n int) []string {
	if len(languages) == 0 {
		return []string{}
	}

	type langCount struct {
		lang  string
		count int
	}

	var sorted []langCount
	for lang, count := range languages {
		sorted = append(sorted, langCount{lang, count})
	}

	// Simple bubble sort (small dataset)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].count < sorted[j+1].count {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	result := make([]string, 0, n)
	for i := 0; i < len(sorted) && i < n; i++ {
		result = append(result, sorted[i].lang)
	}

	return result
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ ProviderAdapter = (*GitHubAdapter)(nil)
