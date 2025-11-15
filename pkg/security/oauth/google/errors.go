package google

import "errors"

var (
	// Configuration errors
	ErrMissingClientID     = errors.New("google oauth: client ID is required")
	ErrMissingClientSecret = errors.New("google oauth: client secret is required")
	ErrMissingRedirectURL  = errors.New("google oauth: redirect URL is required")

	// Request errors
	ErrInvalidState        = errors.New("google oauth: state parameter is required")
	ErrInvalidCode         = errors.New("google oauth: authorization code is required")
	ErrInvalidAccessToken  = errors.New("google oauth: access token is required")
	ErrInvalidRefreshToken = errors.New("google oauth: refresh token is required")

	// API errors
	ErrTokenExchangeFailed = errors.New("google oauth: token exchange failed")
	ErrUserInfoFailed      = errors.New("google oauth: failed to fetch user info")
	ErrTokenRefreshFailed  = errors.New("google oauth: token refresh failed")
)
