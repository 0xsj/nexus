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
// GitHub Adapter
// ============================================================================

// GitHubAdapter implements the ProviderAdapter interface for GitHub.
type GitHubAdapter struct {
	config     *GitHubConfig
	httpClient *http.Client
}

// NewGitHubAdapter creates a new GitHub adapter.
func NewGitHubAdapter(config *GitHubConfig) *GitHubAdapter {
	return &GitHubAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderType returns the provider type.
func (a *GitHubAdapter) GetProviderType() domain.ProviderType {
	return domain.ProviderTypeGitHub
}

// GetAuthorizationURL generates an OAuth authorization URL.
func (a *GitHubAdapter) GetAuthorizationURL(ctx context.Context, state string, redirectURI string, scopes []string) (string, error) {
	// Use redirect URI from config if not provided
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	// Use default scopes if none provided
	if len(scopes) == 0 {
		scopes = a.GetDefaultScopes()
	}

	params := url.Values{}
	params.Set("client_id", a.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("allow_signup", "true")

	authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?%s", params.Encode())
	return authURL, nil
}

// ExchangeCode exchanges an authorization code for an access token.
func (a *GitHubAdapter) ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error) {
	// Use redirect URI from config if not provided
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	data := url.Values{}
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

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
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	scopes := []string{}
	if tokenResp.Scope != "" {
		scopes = strings.Split(tokenResp.Scope, ",")
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: "", // GitHub doesn't provide refresh tokens
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    0, // GitHub tokens don't expire
		Scopes:       scopes,
	}, nil
}

// RefreshAccessToken refreshes an access token.
// Note: GitHub OAuth tokens don't expire, so this is a no-op.
func (a *GitHubAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error) {
	return nil, fmt.Errorf("GitHub OAuth tokens do not expire and cannot be refreshed")
}

// FetchUserData fetches user data from GitHub.
func (a *GitHubAdapter) FetchUserData(ctx context.Context, accessToken string) (*domain.ProviderData, error) {
	// Fetch user profile
	user, err := a.fetchUser(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Fetch repositories
	repos, err := a.fetchRepositories(ctx, accessToken, user.Login)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Fetch organizations
	orgs, err := a.fetchOrganizations(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organizations: %w", err)
	}

	// Normalize to ProviderData
	githubData := &domain.GitHubData{
		Login:         user.Login,
		Name:          user.Name,
		Bio:           user.Bio,
		AvatarURL:     user.AvatarURL,
		Location:      user.Location,
		Company:       user.Company,
		Blog:          user.Blog,
		Email:         user.Email,
		PublicRepos:   user.PublicRepos,
		PublicGists:   user.PublicGists,
		Followers:     user.Followers,
		Following:     user.Following,
		Repositories:  repos,
		Organizations: orgs,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	providerData := &domain.ProviderData{
		ProviderType:     domain.ProviderTypeGitHub,
		ProviderUserID:   fmt.Sprintf("%d", user.ID),
		ProviderUsername: user.Login,
		Email:            user.Email,
		Profile: domain.ProfileData{
			Name:        user.Name,
			Bio:         user.Bio,
			AvatarURL:   user.AvatarURL,
			Location:    user.Location,
			Website:     user.Blog,
			CompanyName: user.Company,
		},
		GitHub:    githubData,
		FetchedAt: time.Now().UTC(),
		RawData: map[string]any{
			"user": user,
		},
	}

	return providerData, nil
}

// ValidateScopes validates the requested scopes.
func (a *GitHubAdapter) ValidateScopes(scopes []string) error {
	validScopes := map[string]bool{
		"user":              true,
		"user:email":        true,
		"read:user":         true,
		"repo":              true,
		"public_repo":       true,
		"read:org":          true,
		"admin:org":         true,
		"gist":              true,
		"notifications":     true,
		"delete_repo":       true,
		"workflow":          true,
		"write:discussion":  true,
		"read:discussion":   true,
		"read:packages":     true,
		"write:packages":    true,
		"delete:packages":   true,
		"admin:gpg_key":     true,
		"admin:public_key":  true,
		"admin:repo_hook":   true,
		"admin:org_hook":    true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

// GetDefaultScopes returns the default scopes.
func (a *GitHubAdapter) GetDefaultScopes() []string {
	return []string{"user", "user:email", "read:org", "public_repo"}
}

// RevokeAccess revokes the access token.
func (a *GitHubAdapter) RevokeAccess(ctx context.Context, accessToken string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE",
		fmt.Sprintf("https://api.github.com/applications/%s/token", a.config.ClientID),
		nil)
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke access: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("revoke failed: status=%d", resp.StatusCode)
	}

	return nil
}

// ============================================================================
// GitHub API Helpers
// ============================================================================

type githubUser struct {
	ID          int64     `json:"id"`
	Login       string    `json:"login"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Bio         string    `json:"bio"`
	AvatarURL   string    `json:"avatar_url"`
	Location    string    `json:"location"`
	Company     string    `json:"company"`
	Blog        string    `json:"blog"`
	PublicRepos int       `json:"public_repos"`
	PublicGists int       `json:"public_gists"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type githubRepo struct {
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	HTMLURL     string    `json:"html_url"`
	Homepage    string    `json:"homepage"`
	Language    string    `json:"language"`
	Stars       int       `json:"stargazers_count"`
	Forks       int       `json:"forks_count"`
	Private     bool      `json:"private"`
	Fork        bool      `json:"fork"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PushedAt    time.Time `json:"pushed_at"`
}

type githubOrg struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

func (a *GitHubAdapter) fetchUser(ctx context.Context, accessToken string) (*githubUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user: status=%d body=%s", resp.StatusCode, string(body))
	}

	var user githubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *GitHubAdapter) fetchRepositories(ctx context.Context, accessToken string, username string) ([]domain.GitHubRepository, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&sort=updated", username),
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch repos: status=%d", resp.StatusCode)
	}

	var repos []githubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	result := make([]domain.GitHubRepository, len(repos))
	for i, repo := range repos {
		result[i] = domain.GitHubRepository{
			Name:        repo.Name,
			FullName:    repo.FullName,
			Description: repo.Description,
			URL:         repo.HTMLURL,
			Homepage:    repo.Homepage,
			Language:    repo.Language,
			Stars:       repo.Stars,
			Forks:       repo.Forks,
			IsPrivate:   repo.Private,
			IsFork:      repo.Fork,
			CreatedAt:   repo.CreatedAt,
			UpdatedAt:   repo.UpdatedAt,
			PushedAt:    repo.PushedAt,
		}
	}

	return result, nil
}

func (a *GitHubAdapter) fetchOrganizations(ctx context.Context, accessToken string) ([]domain.GitHubOrganization, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/orgs", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch orgs: status=%d", resp.StatusCode)
	}

	var orgs []githubOrg
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, err
	}

	result := make([]domain.GitHubOrganization, len(orgs))
	for i, org := range orgs {
		result[i] = domain.GitHubOrganization{
			Login:       org.Login,
			Name:        org.Name,
			Description: org.Description,
			AvatarURL:   org.AvatarURL,
		}
	}

	return result, nil
}
