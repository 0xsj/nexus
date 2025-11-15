package magic

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/nexus/pkg/security/token/random"
	"github.com/0xsj/result"
	"github.com/google/uuid"
)

// Generator creates magic link tokens.
type Generator interface {
	// Generate creates a new magic link token with default TTL for the purpose.
	Generate(ctx context.Context, userID, email string, purpose Purpose) result.Result[*Token]

	// GenerateWithTTL creates a new magic link token with custom TTL.
	GenerateWithTTL(ctx context.Context, userID, email string, purpose Purpose, ttl time.Duration) result.Result[*Token]

	// GenerateURL creates a complete magic link URL for a token.
	GenerateURL(token *Token) result.Result[string]
}

type generator struct {
	config    Config
	randomGen random.Generator
	store     Store
}

// NewGenerator creates a new magic link generator.
func NewGenerator(config Config, store Store) Generator {
	// Apply defaults if not set
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.LoginTTL == 0 {
		config.LoginTTL = 15 * time.Minute
	}
	if config.VerifyTTL == 0 {
		config.VerifyTTL = 24 * time.Hour
	}
	if config.ResetTTL == 0 {
		config.ResetTTL = 1 * time.Hour
	}
	if config.MaxTokensPerUser == 0 {
		config.MaxTokensPerUser = 3
	}

	return &generator{
		config:    config,
		randomGen: random.NewGenerator(),
		store:     store,
	}
}

// Generate creates a new magic link token with default TTL.
func (g *generator) Generate(ctx context.Context, userID, email string, purpose Purpose) result.Result[*Token] {
	ttl := g.config.GetTTLForPurpose(purpose)
	return g.GenerateWithTTL(ctx, userID, email, purpose, ttl)
}

// GenerateWithTTL creates a new magic link token with custom TTL.
func (g *generator) GenerateWithTTL(ctx context.Context, userID, email string, purpose Purpose, ttl time.Duration) result.Result[*Token] {
	// Validate inputs
	if userID == "" {
		return result.Err[*Token](ErrInvalidUserID)
	}

	if email == "" {
		return result.Err[*Token](ErrInvalidEmail)
	}

	if !purpose.IsValid() {
		return result.Err[*Token](ErrInvalidPurpose{Purpose: string(purpose)})
	}

	if ttl <= 0 {
		return result.Err[*Token](ErrInvalidTTL{TTL: ttl})
	}

	// Check rate limit - max tokens per user per purpose
	rateLimitResult := g.checkRateLimit(ctx, userID, purpose)
	if rateLimitResult.IsErr() {
		return result.Err[*Token](rateLimitResult.UnwrapErr())
	}

	// Generate and store token
	return g.generateToken(ctx, userID, email, purpose, ttl)
}

// checkRateLimit ensures user hasn't exceeded max tokens.
func (g *generator) checkRateLimit(ctx context.Context, userID string, purpose Purpose) result.Result[struct{}] {
	countResult := g.store.CountByUser(ctx, userID, purpose)
	if countResult.IsErr() {
		return result.Err[struct{}](countResult.UnwrapErr())
	}

	count := countResult.Unwrap()
	if count >= g.config.MaxTokensPerUser {
		return result.Err[struct{}](ErrTooManyTokens{
			UserID:  userID,
			Purpose: purpose,
			Count:   count,
			Limit:   g.config.MaxTokensPerUser,
		})
	}

	return result.Ok(struct{}{})
}

// generateToken generates and stores the token.
func (g *generator) generateToken(ctx context.Context, userID, email string, purpose Purpose, ttl time.Duration) result.Result[*Token] {
	// Generate cryptographically secure token value
	tokenValueResult := g.randomGen.GenerateBase64(g.config.TokenLength)
	if tokenValueResult.IsErr() {
		return result.Err[*Token](ErrTokenGeneration{Err: tokenValueResult.UnwrapErr()})
	}

	tokenValue := tokenValueResult.Unwrap()

	// Create token
	token := &Token{
		ID:        uuid.New().String(),
		Value:     tokenValue,
		UserID:    userID,
		Purpose:   purpose,
		Email:     email,
		ExpiresAt: time.Now().Add(ttl),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}

	// Store in Redis (with automatic TTL)
	return g.store.Save(ctx, token)
}

// GenerateURL creates a complete magic link URL.
func (g *generator) GenerateURL(token *Token) result.Result[string] {
	if g.config.BaseURL == "" {
		return result.Err[string](ErrInvalidBaseURL)
	}

	if token == nil {
		return result.Err[string](ErrTokenNotFound)
	}

	// Build URL path based on purpose
	var path string
	switch token.Purpose {
	case PurposeLogin:
		path = "/auth/magic-link/verify"
	case PurposeEmailVerify:
		path = "/auth/verify-email"
	case PurposePasswordReset:
		path = "/auth/reset-password"
	default:
		return result.Err[string](ErrInvalidPurpose{Purpose: string(token.Purpose)})
	}

	// Construct full URL
	url := fmt.Sprintf("%s%s?token=%s", g.config.BaseURL, path, token.Value)

	return result.Ok(url)
}
