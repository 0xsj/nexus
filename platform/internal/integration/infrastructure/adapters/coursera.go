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
// Coursera Adapter
// ============================================================================

// CourseraAdapter implements the ProviderAdapter interface for Coursera.
type CourseraAdapter struct {
	config     *CourseraConfig
	httpClient *http.Client
}

// NewCourseraAdapter creates a new Coursera adapter.
func NewCourseraAdapter(config *CourseraConfig) *CourseraAdapter {
	return &CourseraAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderType returns the provider type.
func (a *CourseraAdapter) GetProviderType() domain.ProviderType {
	return domain.ProviderTypeCoursera
}

// GetAuthorizationURL generates an OAuth authorization URL.
func (a *CourseraAdapter) GetAuthorizationURL(ctx context.Context, state string, redirectURI string, scopes []string) (string, error) {
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

	authURL := fmt.Sprintf("https://accounts.coursera.org/oauth2/v1/auth?%s", params.Encode())
	return authURL, nil
}

// ExchangeCode exchanges an authorization code for an access token.
func (a *CourseraAdapter) ExchangeCode(ctx context.Context, code string, redirectURI string) (*domain.OAuthTokens, error) {
	if redirectURI == "" {
		redirectURI = a.config.RedirectURI
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://accounts.coursera.org/oauth2/v1/token", strings.NewReader(data.Encode()))
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
func (a *CourseraAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (*domain.OAuthTokens, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://accounts.coursera.org/oauth2/v1/token", strings.NewReader(data.Encode()))
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

// FetchUserData fetches user data from Coursera.
func (a *CourseraAdapter) FetchUserData(ctx context.Context, accessToken string) (*domain.ProviderData, error) {
	// Fetch user profile
	profile, err := a.fetchProfile(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}

	// Fetch enrollments
	enrollments, err := a.fetchEnrollments(ctx, accessToken, profile.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch enrollments: %w", err)
	}

	// Fetch certificates
	certificates, err := a.fetchCertificates(ctx, accessToken, profile.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch certificates: %w", err)
	}

	// Build Coursera data
	courseraData := &domain.CourseraData{
		UserID:            profile.ID,
		FullName:          profile.FullName,
		Email:             profile.Email,
		ProfileURL:        fmt.Sprintf("https://www.coursera.org/user/%s", profile.ID),
		Enrollments:       enrollments,
		Certifications:    certificates,
		TotalCourses:      len(enrollments),
		TotalCertificates: len(certificates),
		JoinedAt:          time.Now().UTC(), // Coursera doesn't expose this
		UpdatedAt:         time.Now().UTC(),
	}

	// Count completed courses
	for _, enrollment := range enrollments {
		if enrollment.IsCompleted {
			courseraData.CompletedCourses++
		}
	}

	providerData := &domain.ProviderData{
		ProviderType:     domain.ProviderTypeCoursera,
		ProviderUserID:   profile.ID,
		ProviderUsername: profile.Email,
		Email:            profile.Email,
		Profile: domain.ProfileData{
			Name: profile.FullName,
		},
		Coursera:  courseraData,
		FetchedAt: time.Now().UTC(),
		RawData: map[string]any{
			"profile": profile,
		},
	}

	return providerData, nil
}

// ValidateScopes validates the requested scopes.
func (a *CourseraAdapter) ValidateScopes(scopes []string) error {
	validScopes := map[string]bool{
		"view_profile":        true,
		"view_enrollments":    true,
		"view_certificates":   true,
		"view_specializations": true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

// GetDefaultScopes returns the default scopes.
func (a *CourseraAdapter) GetDefaultScopes() []string {
	return []string{"view_profile", "view_enrollments", "view_certificates"}
}

// RevokeAccess revokes the access token.
func (a *CourseraAdapter) RevokeAccess(ctx context.Context, accessToken string) error {
	// Coursera doesn't provide a public revocation endpoint
	// Tokens expire automatically
	return nil
}

// ============================================================================
// Coursera API Helpers
// ============================================================================

type courseraProfile struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

type courseraEnrollment struct {
	ID         string    `json:"id"`
	CourseID   string    `json:"courseId"`
	CourseName string    `json:"courseName"`
	EnrolledAt time.Time `json:"enrolledTimestamp"`
	Grade      *float64  `json:"grade"`
	Completed  bool      `json:"completed"`
}

type courseraCertificate struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	CourseID        string    `json:"courseId"`
	CompletedAt     time.Time `json:"completedTimestamp"`
	VerificationURL string    `json:"verifyUrl"`
}

func (a *CourseraAdapter) fetchProfile(ctx context.Context, accessToken string) (*courseraProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.coursera.org/api/externalBasicProfiles.v1/me", nil)
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

	var profileResp struct {
		Elements []courseraProfile `json:"elements"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&profileResp); err != nil {
		return nil, err
	}

	if len(profileResp.Elements) == 0 {
		return nil, fmt.Errorf("no profile found")
	}

	return &profileResp.Elements[0], nil
}

func (a *CourseraAdapter) fetchEnrollments(ctx context.Context, accessToken string, userID string) ([]domain.CourseraEnrollment, error) {
	url := fmt.Sprintf("https://api.coursera.org/api/onDemandCourseEnrollments.v1?userId=%s&includes=courses&limit=100", userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return nil, fmt.Errorf("failed to fetch enrollments: status=%d", resp.StatusCode)
	}

	var enrollmentResp struct {
		Elements []courseraEnrollment `json:"elements"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&enrollmentResp); err != nil {
		return nil, err
	}

	result := make([]domain.CourseraEnrollment, len(enrollmentResp.Elements))
	for i, enrollment := range enrollmentResp.Elements {
		var completedAt *time.Time
		if enrollment.Completed {
			completedAt = &enrollment.EnrolledAt // Approximation
		}

		result[i] = domain.CourseraEnrollment{
			CourseID:    enrollment.CourseID,
			CourseName:  enrollment.CourseName,
			EnrolledAt:  enrollment.EnrolledAt,
			CompletedAt: completedAt,
			Grade:       enrollment.Grade,
			IsCompleted: enrollment.Completed,
		}
	}

	return result, nil
}

func (a *CourseraAdapter) fetchCertificates(ctx context.Context, accessToken string, userID string) ([]domain.CourseraCertification, error) {
	url := fmt.Sprintf("https://api.coursera.org/api/onDemandCourseCertificates.v1?userId=%s&limit=100", userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return nil, fmt.Errorf("failed to fetch certificates: status=%d", resp.StatusCode)
	}

	var certResp struct {
		Elements []courseraCertificate `json:"elements"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return nil, err
	}

	result := make([]domain.CourseraCertification, len(certResp.Elements))
	for i, cert := range certResp.Elements {
		result[i] = domain.CourseraCertification{
			CertificateID:   cert.ID,
			CertificateURL:  cert.VerificationURL,
			CourseID:        cert.CourseID,
			CourseName:      cert.Name,
			CompletedAt:     cert.CompletedAt,
			VerificationURL: cert.VerificationURL,
			IsVerified:      true,
		}
	}

	return result, nil
}
