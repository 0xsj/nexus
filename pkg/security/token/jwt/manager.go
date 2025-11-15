package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/0xsj/result"
	"github.com/golang-jwt/jwt/v5"
)

// TokenPair represents a pair of access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"` // seconds
	ExpiresAt    time.Time `json:"expires_at"`
}

// Manager handles JWT token operations.
type Manager interface {
	// GenerateAccessToken creates a new access token.
	GenerateAccessToken(userID, email, username string, roles []string) result.Result[string]

	// GenerateRefreshToken creates a new refresh token.
	GenerateRefreshToken(userID string) result.Result[string]

	// GenerateTokenPair creates both access and refresh tokens.
	GenerateTokenPair(userID, email, username string, roles []string) result.Result[TokenPair]

	// ValidateToken validates a token and returns its claims.
	ValidateToken(tokenString string) result.Result[*Claims]

	// ValidateAccessToken validates specifically an access token.
	ValidateAccessToken(tokenString string) result.Result[*Claims]

	// ValidateRefreshToken validates specifically a refresh token.
	ValidateRefreshToken(tokenString string) result.Result[*Claims]

	// RefreshAccessToken generates a new access token from a refresh token.
	RefreshAccessToken(refreshToken string, email, username string, roles []string) result.Result[TokenPair]
}

type manager struct {
	config     Config
	signingKey interface{} // Can be []byte for HMAC or *rsa.PrivateKey for RSA
	verifyKey  interface{} // Can be []byte for HMAC or *rsa.PublicKey for RSA
}

// NewManager creates a new JWT manager.
func NewManager(config Config) result.Result[Manager] {
	// Validate config
	if err := config.Validate(); err != nil {
		return result.Err[Manager](err)
	}

	m := &manager{
		config: config,
	}

	// Load keys based on algorithm
	if config.Algorithm.IsSymmetric() {
		// HMAC - use secret key for both signing and verification
		m.signingKey = []byte(config.SecretKey)
		m.verifyKey = []byte(config.SecretKey)
	} else {
		// Asymmetric - load private and public keys
		privateKeyResult := loadPrivateKey(config.PrivateKeyPath)
		if privateKeyResult.IsErr() {
			return result.Err[Manager](privateKeyResult.UnwrapErr())
		}
		m.signingKey = privateKeyResult.Unwrap()

		publicKeyResult := loadPublicKey(config.PublicKeyPath)
		if publicKeyResult.IsErr() {
			return result.Err[Manager](publicKeyResult.UnwrapErr())
		}
		m.verifyKey = publicKeyResult.Unwrap()
	}

	return result.Ok[Manager](m)
}

// GenerateAccessToken creates a new access token.
func (m *manager) GenerateAccessToken(userID, email, username string, roles []string) result.Result[string] {
	claims := NewAccessTokenClaims(
		userID,
		email,
		username,
		roles,
		m.config.Issuer,
		m.config.Audience,
		m.config.AccessTokenTTL,
	)

	return m.generateToken(claims)
}

// GenerateRefreshToken creates a new refresh token.
func (m *manager) GenerateRefreshToken(userID string) result.Result[string] {
	claims := NewRefreshTokenClaims(
		userID,
		m.config.Issuer,
		m.config.Audience,
		m.config.RefreshTokenTTL,
	)

	return m.generateToken(claims)
}

// GenerateTokenPair creates both access and refresh tokens.
func (m *manager) GenerateTokenPair(userID, email, username string, roles []string) result.Result[TokenPair] {
	// Generate access token
	accessTokenResult := m.GenerateAccessToken(userID, email, username, roles)
	if accessTokenResult.IsErr() {
		return result.Err[TokenPair](accessTokenResult.UnwrapErr())
	}

	// Generate refresh token
	refreshTokenResult := m.GenerateRefreshToken(userID)
	if refreshTokenResult.IsErr() {
		return result.Err[TokenPair](refreshTokenResult.UnwrapErr())
	}

	accessToken := accessTokenResult.Unwrap()
	refreshToken := refreshTokenResult.Unwrap()

	pair := TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.config.AccessTokenTTL.Seconds()),
		ExpiresAt:    time.Now().Add(m.config.AccessTokenTTL),
	}

	return result.Ok(pair)
}

// ValidateToken validates a token and returns its claims.
func (m *manager) ValidateToken(tokenString string) result.Result[*Claims] {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if token.Method.Alg() != m.config.Algorithm.String() {
			return nil, ErrInvalidAlgorithm
		}
		return m.verifyKey, nil
	})

	if err != nil {
		return result.Err[*Claims](m.parseError(err))
	}

	if !token.Valid {
		return result.Err[*Claims](ErrInvalidToken)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return result.Err[*Claims](ErrMissingClaims)
	}

	// Validate claims
	return m.validateClaims(claims)
}

// ValidateAccessToken validates specifically an access token.
func (m *manager) ValidateAccessToken(tokenString string) result.Result[*Claims] {
	claimsResult := m.ValidateToken(tokenString)
	if claimsResult.IsErr() {
		return claimsResult
	}

	claims := claimsResult.Unwrap()

	if !claims.IsAccessToken() {
		return result.Err[*Claims](ErrInvalidTokenType{
			Expected: string(TokenTypeAccess),
			Actual:   string(claims.TokenType),
		})
	}

	return result.Ok(claims)
}

// ValidateRefreshToken validates specifically a refresh token.
func (m *manager) ValidateRefreshToken(tokenString string) result.Result[*Claims] {
	claimsResult := m.ValidateToken(tokenString)
	if claimsResult.IsErr() {
		return claimsResult
	}

	claims := claimsResult.Unwrap()

	if !claims.IsRefreshToken() {
		return result.Err[*Claims](ErrInvalidTokenType{
			Expected: string(TokenTypeRefresh),
			Actual:   string(claims.TokenType),
		})
	}

	return result.Ok(claims)
}

// RefreshAccessToken generates a new access token from a refresh token.
func (m *manager) RefreshAccessToken(refreshToken string, email, username string, roles []string) result.Result[TokenPair] {
	// Validate refresh token
	claimsResult := m.ValidateRefreshToken(refreshToken)
	if claimsResult.IsErr() {
		return result.Err[TokenPair](claimsResult.UnwrapErr())
	}

	claims := claimsResult.Unwrap()

	// Generate new token pair with user info
	return m.GenerateTokenPair(claims.UserID, email, username, roles)
}

// generateToken generates a JWT token from claims.
func (m *manager) generateToken(claims *Claims) result.Result[string] {
	var token *jwt.Token

	switch m.config.Algorithm {
	case AlgorithmHS256:
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	case AlgorithmHS384:
		token = jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	case AlgorithmHS512:
		token = jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	case AlgorithmRS256:
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	case AlgorithmRS384:
		token = jwt.NewWithClaims(jwt.SigningMethodRS384, claims)
	case AlgorithmRS512:
		token = jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	case AlgorithmES256:
		token = jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	case AlgorithmES384:
		token = jwt.NewWithClaims(jwt.SigningMethodES384, claims)
	case AlgorithmES512:
		token = jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	default:
		return result.Err[string](ErrInvalidAlgorithm)
	}

	tokenString, err := token.SignedString(m.signingKey)
	if err != nil {
		return result.Err[string](err)
	}

	return result.Ok(tokenString)
}

// validateClaims validates standard and custom claims.
func (m *manager) validateClaims(claims *Claims) result.Result[*Claims] {
	// Check expiration
	if claims.IsExpired() {
		return result.Err[*Claims](ErrTokenExpired)
	}

	// Check not before
	if claims.IsNotYetValid() {
		return result.Err[*Claims](ErrTokenNotYetValid{
			NotBefore: claims.NotBefore.Time,
		})
	}

	// Validate issuer
	if claims.Issuer != m.config.Issuer {
		return result.Err[*Claims](ErrInvalidIssuer{
			Expected: m.config.Issuer,
			Actual:   claims.Issuer,
		})
	}

	// Validate audience
	expectedAudience := []string{m.config.Audience}
	if !containsAudience(claims.Audience, expectedAudience) {
		return result.Err[*Claims](ErrInvalidAudience{
			Expected: expectedAudience,
			Actual:   claims.Audience,
		})
	}

	return result.Ok(claims)
}

// parseError converts jwt parsing errors to our error types.
func (m *manager) parseError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific JWT errors
	switch {
	case err == jwt.ErrTokenExpired:
		return ErrTokenExpired
	case err == jwt.ErrTokenNotValidYet:
		return ErrTokenNotYetValid{NotBefore: time.Now()}
	default:
		return ErrInvalidToken
	}
}

// Helper functions for key loading

// loadPrivateKey loads an RSA private key from a PEM file.
func loadPrivateKey(path string) result.Result[*rsa.PrivateKey] {
	if path == "" {
		return result.Err[*rsa.PrivateKey](ErrInvalidSigningKey)
	}

	keyData, err := os.ReadFile(path)
	if err != nil {
		return result.Err[*rsa.PrivateKey](err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return result.Err[*rsa.PrivateKey](ErrInvalidSigningKey)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return result.Err[*rsa.PrivateKey](err)
	}

	return result.Ok(key)
}

// loadPublicKey loads an RSA public key from a PEM file.
func loadPublicKey(path string) result.Result[*rsa.PublicKey] {
	if path == "" {
		return result.Err[*rsa.PublicKey](ErrInvalidSigningKey)
	}

	keyData, err := os.ReadFile(path)
	if err != nil {
		return result.Err[*rsa.PublicKey](err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return result.Err[*rsa.PublicKey](ErrInvalidSigningKey)
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		// Try ParsePKIXPublicKey as fallback
		pubInterface, err2 := x509.ParsePKIXPublicKey(block.Bytes)
		if err2 != nil {
			return result.Err[*rsa.PublicKey](err)
		}

		rsaKey, ok := pubInterface.(*rsa.PublicKey)
		if !ok {
			return result.Err[*rsa.PublicKey](ErrInvalidSigningKey)
		}
		return result.Ok(rsaKey)
	}

	return result.Ok(key)
}

// containsAudience checks if expected audience is in the token's audience list.
func containsAudience(actual jwt.ClaimStrings, expected []string) bool {
	if len(expected) == 0 {
		return true
	}

	actualSlice := []string(actual)
	for _, exp := range expected {
		for _, act := range actualSlice {
			if exp == act {
				return true
			}
		}
	}
	return false
}
