package google

const (
	// Google OAuth 2.0 endpoints
	GoogleAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	GoogleTokenURL    = "https://oauth2.googleapis.com/token"
	GoogleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	GoogleRevokeURL   = "https://oauth2.googleapis.com/revoke"
)

// Default scopes for Google OAuth
var DefaultScopes = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

// Config holds Google OAuth configuration.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewConfig creates a new Google OAuth config.
func NewConfig(clientID, clientSecret, redirectURL string) *Config {
	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return ErrMissingClientID
	}
	if c.ClientSecret == "" {
		return ErrMissingClientSecret
	}
	if c.RedirectURL == "" {
		return ErrMissingRedirectURL
	}
	return nil
}
