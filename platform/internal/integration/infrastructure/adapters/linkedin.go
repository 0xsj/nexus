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
// LinkedIn Adapter
// ============================================================================

// LinkedInAdapter implements the ProviderAdapter interface for LinkedIn.
type LinkedInAdapter struct {
	config     *LinkedInConfig
	httpClient *http.Client
}

// NewLinkedInAdapter creates a new LinkedIn adapter.
func NewLinkedInAdapter(config *LinkedInConfig) *LinkedInAdapter {
	return &LinkedInAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderType returns the provider type.
func (a *LinkedInAdapter) GetProviderType() domain.ProviderType {
	return domain.ProviderTypeLinkedIn
}

// GetAuthorizationURL generates an OAuth authorization URL.
func (a *LinkedInAdapter) GetAuthorizationURL(ctx context.Context, state string, redirectURI string, scopes []string) (string, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	if len(scopes) == 0 {
		scopes = a.GetDefaultScopes()
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", a.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))

	authURL := fmt.Sprintf("https://www.linkedin.com/oauth/v2/authorization?%s", params.Encode())
	return authURL, nil
}

// ExchangeCode exchanges an authorization code for an access token.
func (a *LinkedInAdapter) ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.linkedin.com/oauth/v2/accessToken", strings.NewReader(data.Encode()))
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
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
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
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.ExpiresIn,
		Scopes:       scopes,
	}, nil
}

// RefreshAccessToken refreshes an access token.
func (a *LinkedInAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.linkedin.com/oauth/v2/accessToken", strings.NewReader(data.Encode()))
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
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &domain.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// FetchUserData fetches user data from LinkedIn.
func (a *LinkedInAdapter) FetchUserData(ctx context.Context, accessToken string) (*domain.ProviderData, error) {
	// Fetch user profile
	profile, err := a.fetchProfile(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}

	// Fetch email
	email, err := a.fetchEmail(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch email: %w", err)
	}

	// Build LinkedIn data
	linkedInData := &domain.LinkedInData{
		ID:          profile.ID,
		FirstName:   profile.FirstName.Localized.EnUS,
		LastName:    profile.LastName.Localized.EnUS,
		Headline:    profile.Headline.Localized.EnUS,
		ProfileURL:  fmt.Sprintf("https://www.linkedin.com/in/%s", profile.VanityName),
		PictureURL:  profile.ProfilePicture.DisplayImage,
		// Note: LinkedIn API v2 requires additional permissions for positions, education, etc.
		// These would need separate API calls with appropriate scopes
	}

	providerData := &domain.ProviderData{
		ProviderType:     domain.ProviderTypeLinkedIn,
		ProviderUserID:   profile.ID,
		ProviderUsername: profile.VanityName,
		Email:            email,
		Profile: domain.ProfileData{
			Name:     fmt.Sprintf("%s %s", linkedInData.FirstName, linkedInData.LastName),
			Bio:      linkedInData.Headline,
			Website:  linkedInData.ProfileURL,
		},
		LinkedIn:  linkedInData,
		FetchedAt: time.Now().UTC(),
		RawData: map[string]any{
			"profile": profile,
		},
	}

	return providerData, nil
}

// ValidateScopes validates the requested scopes.
func (a *LinkedInAdapter) ValidateScopes(scopes []string) error {
	validScopes := map[string]bool{
		"r_liteprofile":   true,
		"r_emailaddress":  true,
		"r_basicprofile":  true,
		"r_fullprofile":   true,
		"w_member_social": true,
		"r_ads":           true,
		"r_ads_reporting": true,
		"rw_ads":          true,
		"r_organization_social": true,
		"w_organization_social": true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

// GetDefaultScopes returns the default scopes.
func (a *LinkedInAdapter) GetDefaultScopes() []string {
	return []string{"r_liteprofile", "r_emailaddress"}
}

// RevokeAccess revokes the access token.
func (a *LinkedInAdapter) RevokeAccess(ctx context.Context, accessToken string) error {
	// LinkedIn doesn't provide a token revocation endpoint
	// Tokens expire after 60 days automatically
	return nil
}

// ============================================================================
// LinkedIn API Helpers
// ============================================================================

type linkedInProfile struct {
	ID        string `json:"id"`
	FirstName struct {
		Localized struct {
			EnUS string `json:"en_US"`
		} `json:"localized"`
	} `json:"firstName"`
	LastName struct {
		Localized struct {
			EnUS string `json:"en_US"`
		} `json:"localized"`
	} `json:"lastName"`
	Headline struct {
		Localized struct {
			EnUS string `json:"en_US"`
		} `json:"localized"`
	} `json:"headline"`
	VanityName     string `json:"vanityName"`
	ProfilePicture struct {
		DisplayImage string `json:"displayImage~"`
	} `json:"profilePicture"`
}

type linkedInEmail struct {
	Elements []struct {
		Handle      string `json:"handle~"`
		HandleValue struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"handle"`
	} `json:"elements"`
}

func (a *LinkedInAdapter) fetchProfile(ctx context.Context, accessToken string) (*linkedInProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.linkedin.com/v2/me?projection=(id,firstName,lastName,headline,vanityName,profilePicture(displayImage~:playableStreams))",
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
		return nil, fmt.Errorf("failed to fetch profile: status=%d body=%s", resp.StatusCode, string(body))
	}

	var profile linkedInProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (a *LinkedInAdapter) fetchEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))",
		nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch email: status=%d body=%s", resp.StatusCode, string(body))
	}

	var emailResp linkedInEmail
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		return "", err
	}

	if len(emailResp.Elements) == 0 {
		return "", fmt.Errorf("no email found")
	}

	return emailResp.Elements[0].HandleValue.EmailAddress, nil
}
