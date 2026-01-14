package token

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/jwt"
)

// ============================================================================
// Service
// ============================================================================

// Service implements domain.TokenService using JWT.
type Service struct {
	signer   crypto.Signer
	verifier crypto.Verifier
	issuer   string
	audience []string

	// Token revocation (in-memory for now, use Redis in production)
	revoked   map[string]time.Time
	revokedMu sync.RWMutex

	// Config
	config Config
}

// Config contains token service configuration.
type Config struct {
	// Issuer is the token issuer (e.g., "https://proof.io")
	Issuer string

	// Audience is the intended audience
	Audience []string

	// AccessTokenTTL is the access token lifetime
	AccessTokenTTL time.Duration

	// RefreshTokenTTL is the refresh token lifetime
	RefreshTokenTTL time.Duration

	// KeyID is the key identifier for the JWT header
	KeyID string
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		Issuer:          "https://proof.io",
		Audience:        []string{"https://proof.io"},
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		KeyID:           "",
	}
}

// Ensure Service implements domain.TokenService.
var _ domain.TokenService = (*Service)(nil)

// NewService creates a new token service.
func NewService(signer crypto.Signer, verifier crypto.Verifier, config Config) *Service {
	return &Service{
		signer:   signer,
		verifier: verifier,
		issuer:   config.Issuer,
		audience: config.Audience,
		revoked:  make(map[string]time.Time),
		config:   config,
	}
}

// ============================================================================
// Token Generation
// ============================================================================

// GenerateAccessToken generates a short-lived access token.
func (s *Service) GenerateAccessToken(ctx context.Context, params domain.AccessTokenParams) (string, error) {
	const op = "token.Service.GenerateAccessToken"

	// Generate token ID
	tokenID, err := generateTokenID()
	if err != nil {
		return "", errors.Internal(op, err)
	}

	// Determine expiration
	expiresIn := params.ExpiresIn
	if expiresIn == 0 {
		expiresIn = s.config.AccessTokenTTL
	}

	// Build claims
	now := time.Now()

	fmt.Printf("[DEBUG] GenerateAccessToken: now=%v, expiresIn=%v, expiresAt=%v\n",
		now, expiresIn, now.Add(expiresIn))

	baseClaims := jwt.NewClaims(s.issuer, params.UserID).
		WithAudience(s.audience...).
		WithExpiresAt(now.Add(expiresIn)).
		WithIssuedAt(now).
		WithNotBefore(now).
		WithID(tokenID)

	fmt.Printf("[DEBUG] baseClaims: exp=%v, iat=%v, nbf=%v\n",
		baseClaims.ExpiresAt, baseClaims.IssuedAt, baseClaims.NotBefore)

	claims := AccessTokenClaims{
		Claims:    baseClaims,
		DID:       params.DID,
		SessionID: params.SessionID,
		Scopes:    params.Scopes,
		TokenType: domain.TokenTypeAccess,
	}

	fmt.Printf("[DEBUG] AccessTokenClaims.Claims: exp=%v, iat=%v, nbf=%v\n",
		claims.Claims.ExpiresAt, claims.Claims.IssuedAt, claims.Claims.NotBefore)

	// Build header
	header := jwt.NewHeader(s.signer.Algorithm().JWAName())
	if s.config.KeyID != "" {
		header = header.WithKeyID(s.config.KeyID)
	}

	// Sign token
	token, err := jwt.Sign(header, claims, s.signer)
	if err != nil {
		return "", errors.Wrap(err, op)
	}

	return token, nil
}

// GenerateRefreshToken generates a long-lived refresh token.
func (s *Service) GenerateRefreshToken(ctx context.Context, params domain.RefreshTokenParams) (string, error) {
	const op = "token.Service.GenerateRefreshToken"

	// Generate token ID
	tokenID, err := generateTokenID()
	if err != nil {
		return "", errors.Internal(op, err)
	}

	// Determine expiration
	expiresIn := params.ExpiresIn
	if expiresIn == 0 {
		expiresIn = s.config.RefreshTokenTTL
	}

	// Build claims
	now := time.Now()
	claims := RefreshTokenClaims{
		Claims: jwt.NewClaims(s.issuer, params.UserID).
			WithAudience(s.audience...).
			WithExpiresAt(now.Add(expiresIn)).
			WithIssuedAt(now).
			WithNotBefore(now).
			WithID(tokenID),
		SessionID: params.SessionID,
		TokenType: domain.TokenTypeRefresh,
	}

	// Build header
	header := jwt.NewHeader(s.signer.Algorithm().JWAName())
	if s.config.KeyID != "" {
		header = header.WithKeyID(s.config.KeyID)
	}

	// Sign token
	token, err := jwt.Sign(header, claims, s.signer)
	if err != nil {
		return "", errors.Wrap(err, op)
	}

	return token, nil
}

// ============================================================================
// Token Validation
// ============================================================================

// ValidateAccessToken validates an access token and returns claims.
func (s *Service) ValidateAccessToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	const op = "token.Service.ValidateAccessToken"

	// Parse and verify
	token, err := jwt.Verify(tokenString, s.verifier)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Parse claims
	var claims AccessTokenClaims
	if err := token.ParseClaims(&claims); err != nil {
		return nil, jwt.ErrInvalidClaims(op, "failed to parse claims")
	}

	fmt.Printf("[DEBUG] ValidateAccessToken: parsed TokenType=%q, expected=%q\n", claims.TokenType, domain.TokenTypeAccess)
	fmt.Printf("[DEBUG] ValidateAccessToken: claims=%+v\n", claims)

	// Validate standard claims
	if err := claims.Claims.Validate(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Validate issuer
	if err := claims.Claims.ValidateIssuer(s.issuer); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Check token type
	if claims.TokenType != domain.TokenTypeAccess {
		return nil, jwt.ErrInvalidClaims(op, "invalid token type")
	}

	// Check revocation
	revoked, err := s.IsRevoked(ctx, claims.Claims.ID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	if revoked {
		return nil, jwt.ErrTokenExpired(op)
	}

	return &domain.TokenClaims{
		TokenID:   claims.Claims.ID,
		UserID:    claims.Claims.Subject,
		DID:       claims.DID,
		SessionID: claims.SessionID,
		Scopes:    claims.Scopes,
		IssuedAt:  *claims.Claims.IssuedAt,
		ExpiresAt: *claims.Claims.ExpiresAt,
		TokenType: claims.TokenType,
	}, nil
}

// ValidateRefreshToken validates a refresh token and returns claims.
func (s *Service) ValidateRefreshToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	const op = "token.Service.ValidateRefreshToken"

	// Parse and verify
	token, err := jwt.Verify(tokenString, s.verifier)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Parse claims
	var claims RefreshTokenClaims
	if err := token.ParseClaims(&claims); err != nil {
		return nil, jwt.ErrInvalidClaims(op, "failed to parse claims")
	}

	// Validate standard claims
	if err := claims.Claims.Validate(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Validate issuer
	if err := claims.Claims.ValidateIssuer(s.issuer); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Check token type
	if claims.TokenType != domain.TokenTypeRefresh {
		return nil, jwt.ErrInvalidClaims(op, "invalid token type")
	}

	// Check revocation
	revoked, err := s.IsRevoked(ctx, claims.Claims.ID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	if revoked {
		return nil, jwt.ErrTokenExpired(op)
	}

	return &domain.TokenClaims{
		TokenID:   claims.Claims.ID,
		UserID:    claims.Claims.Subject,
		SessionID: claims.SessionID,
		IssuedAt:  *claims.Claims.IssuedAt,
		ExpiresAt: *claims.Claims.ExpiresAt,
		TokenType: claims.TokenType,
	}, nil
}

// ============================================================================
// Token Revocation
// ============================================================================

// RevokeToken revokes a token by its ID.
func (s *Service) RevokeToken(ctx context.Context, tokenString string) error {
	const op = "token.Service.RevokeToken"

	// Extract token ID without full verification
	token, err := jwt.Parse(tokenString)
	if err != nil {
		return errors.Wrap(err, op)
	}

	stdClaims, err := token.StandardClaims()
	if err != nil {
		return errors.Wrap(err, op)
	}

	if stdClaims.ID == "" {
		return jwt.ErrMissingClaim(op, "jti")
	}

	s.revokedMu.Lock()
	defer s.revokedMu.Unlock()

	// Store with expiration time for cleanup
	expiry := time.Now().Add(s.config.RefreshTokenTTL)
	if stdClaims.ExpiresAt != nil {
		expiry = *stdClaims.ExpiresAt
	}
	s.revoked[stdClaims.ID] = expiry

	return nil
}

// IsRevoked checks if a token ID is revoked.
func (s *Service) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	s.revokedMu.RLock()
	defer s.revokedMu.RUnlock()

	_, exists := s.revoked[tokenID]
	return exists, nil
}

// CleanupRevoked removes expired entries from the revocation list.
// Should be called periodically.
func (s *Service) CleanupRevoked() {
	s.revokedMu.Lock()
	defer s.revokedMu.Unlock()

	now := time.Now()
	for id, expiry := range s.revoked {
		if now.After(expiry) {
			delete(s.revoked, id)
		}
	}
}

// ============================================================================
// Claims Types
// ============================================================================

// AccessTokenClaims contains claims for access tokens.
type AccessTokenClaims struct {
	jwt.Claims

	// DID is the user's primary DID.
	DID string `json:"did,omitempty"`

	// SessionID is the session that issued this token.
	SessionID string `json:"sid,omitempty"`

	// Scopes are the permissions granted.
	Scopes []string `json:"scopes,omitempty"`

	// TokenType identifies this as an access token.
	TokenType domain.TokenType `json:"type"`
}

// RefreshTokenClaims contains claims for refresh tokens.
type RefreshTokenClaims struct {
	jwt.Claims

	// SessionID is the session that issued this token.
	SessionID string `json:"sid,omitempty"`

	// TokenType identifies this as a refresh token.
	TokenType domain.TokenType `json:"type"`
}

// MarshalJSON implements json.Marshaler for AccessTokenClaims.
// This is needed because jwt.Claims has its own MarshalJSON that would
// otherwise override the embedded struct serialization.
func (c AccessTokenClaims) MarshalJSON() ([]byte, error) {
	type Alias AccessTokenClaims

	// Create a map to hold all claims
	m := make(map[string]any)

	// Add standard claims
	if c.Claims.Issuer != "" {
		m["iss"] = c.Claims.Issuer
	}
	if c.Claims.Subject != "" {
		m["sub"] = c.Claims.Subject
	}
	if len(c.Claims.Audience) > 0 {
		if len(c.Claims.Audience) == 1 {
			m["aud"] = c.Claims.Audience[0]
		} else {
			m["aud"] = c.Claims.Audience
		}
	}
	if c.Claims.ExpiresAt != nil {
		m["exp"] = c.Claims.ExpiresAt.Unix()
	}
	if c.Claims.NotBefore != nil {
		m["nbf"] = c.Claims.NotBefore.Unix()
	}
	if c.Claims.IssuedAt != nil {
		m["iat"] = c.Claims.IssuedAt.Unix()
	}
	if c.Claims.ID != "" {
		m["jti"] = c.Claims.ID
	}

	// Add custom claims
	if c.DID != "" {
		m["did"] = c.DID
	}
	if c.SessionID != "" {
		m["sid"] = c.SessionID
	}
	if len(c.Scopes) > 0 {
		m["scopes"] = c.Scopes
	}
	m["type"] = c.TokenType

	return json.Marshal(m)
}

// UnmarshalJSON implements json.Unmarshaler for AccessTokenClaims.
func (c *AccessTokenClaims) UnmarshalJSON(data []byte) error {
	// First unmarshal into the embedded Claims
	if err := json.Unmarshal(data, &c.Claims); err != nil {
		return err
	}

	// Then unmarshal the custom fields using an alias to avoid recursion
	type customFields struct {
		DID       string           `json:"did"`
		SessionID string           `json:"sid"`
		Scopes    []string         `json:"scopes"`
		TokenType domain.TokenType `json:"type"`
	}

	var custom customFields
	if err := json.Unmarshal(data, &custom); err != nil {
		return err
	}

	c.DID = custom.DID
	c.SessionID = custom.SessionID
	c.Scopes = custom.Scopes
	c.TokenType = custom.TokenType

	return nil
}

// UnmarshalJSON implements json.Unmarshaler for RefreshTokenClaims.
func (c *RefreshTokenClaims) UnmarshalJSON(data []byte) error {
	// First unmarshal into the embedded Claims
	if err := json.Unmarshal(data, &c.Claims); err != nil {
		return err
	}

	// Then unmarshal the custom fields
	type customFields struct {
		SessionID string           `json:"sid"`
		TokenType domain.TokenType `json:"type"`
	}

	var custom customFields
	if err := json.Unmarshal(data, &custom); err != nil {
		return err
	}

	c.SessionID = custom.SessionID
	c.TokenType = custom.TokenType

	return nil
}

// MarshalJSON implements json.Marshaler for RefreshTokenClaims.
func (c RefreshTokenClaims) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)

	// Add standard claims
	if c.Claims.Issuer != "" {
		m["iss"] = c.Claims.Issuer
	}
	if c.Claims.Subject != "" {
		m["sub"] = c.Claims.Subject
	}
	if len(c.Claims.Audience) > 0 {
		if len(c.Claims.Audience) == 1 {
			m["aud"] = c.Claims.Audience[0]
		} else {
			m["aud"] = c.Claims.Audience
		}
	}
	if c.Claims.ExpiresAt != nil {
		m["exp"] = c.Claims.ExpiresAt.Unix()
	}
	if c.Claims.NotBefore != nil {
		m["nbf"] = c.Claims.NotBefore.Unix()
	}
	if c.Claims.IssuedAt != nil {
		m["iat"] = c.Claims.IssuedAt.Unix()
	}
	if c.Claims.ID != "" {
		m["jti"] = c.Claims.ID
	}

	// Add custom claims
	if c.SessionID != "" {
		m["sid"] = c.SessionID
	}
	m["type"] = c.TokenType

	return json.Marshal(m)
}

// ============================================================================
// Helpers
// ============================================================================

// generateTokenID generates a unique token ID.
func generateTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
