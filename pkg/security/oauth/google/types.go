package google

import "time"

// TokenResponse represents Google's OAuth token response.
type TokenResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	ExpiresIn    int        `json:"expires_in"`
	TokenType    string     `json:"token_type"`
	Scope        string     `json:"scope,omitempty"`
	IDToken      string     `json:"id_token,omitempty"`
	ExpiresAt    *time.Time `json:"-"` // Calculated from ExpiresIn
}

// UserInfo represents Google user information.
type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}
