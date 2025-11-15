package magic

import (
	"context"

	"github.com/0xsj/result"
)

// Verifier validates magic link tokens.
type Verifier interface {
	// Verify validates a token without consuming it (read-only check).
	Verify(ctx context.Context, tokenValue string, purpose Purpose) result.Result[*Token]

	// VerifyAndConsume validates and marks token as used (one-time use).
	// This should be used when actually logging in or verifying email.
	VerifyAndConsume(ctx context.Context, tokenValue string, purpose Purpose) result.Result[*Token]
}

type verifier struct {
	store Store
}

// NewVerifier creates a new token verifier.
func NewVerifier(store Store) Verifier {
	return &verifier{store: store}
}

// Verify validates a token without consuming it.
func (v *verifier) Verify(ctx context.Context, tokenValue string, purpose Purpose) result.Result[*Token] {
	// Validate inputs
	if tokenValue == "" {
		return result.Err[*Token](ErrTokenNotFound)
	}

	if !purpose.IsValid() {
		return result.Err[*Token](ErrInvalidPurpose{Purpose: string(purpose)})
	}

	// Retrieve token from store
	tokenResult := v.store.GetByValue(ctx, tokenValue, purpose)
	if tokenResult.IsErr() {
		return result.Err[*Token](ErrTokenNotFound)
	}

	token := tokenResult.Unwrap()

	// Validate token
	return v.validateToken(token)
}

// VerifyAndConsume validates and marks the token as used (one-time use).
func (v *verifier) VerifyAndConsume(ctx context.Context, tokenValue string, purpose Purpose) result.Result[*Token] {
	// First verify the token
	tokenResult := v.Verify(ctx, tokenValue, purpose)
	if tokenResult.IsErr() {
		return tokenResult
	}

	token := tokenResult.Unwrap()

	// Mark as used in store
	markResult := v.store.MarkAsUsed(ctx, token.ID)
	if markResult.IsErr() {
		return result.Err[*Token](markResult.UnwrapErr())
	}

	// Update local copy
	token.MarkAsUsed()

	return result.Ok(token)
}

// validateToken checks if a token is valid.
func (v *verifier) validateToken(token *Token) result.Result[*Token] {
	// Check if expired
	if token.IsExpired() {
		return result.Err[*Token](ErrTokenExpired{
			TokenID:   token.ID,
			ExpiresAt: token.ExpiresAt,
		})
	}

	// Check if already used
	if token.IsUsed() {
		return result.Err[*Token](ErrTokenAlreadyUsed{
			TokenID: token.ID,
			UsedAt:  *token.UsedAt,
		})
	}

	return result.Ok(token)
}
