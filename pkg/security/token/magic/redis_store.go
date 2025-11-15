package magic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xsj/result"
	"github.com/redis/go-redis/v9"
)

// RedisStore is a Redis implementation of Store.
type RedisStore struct {
	client *redis.Client
	prefix string // Key prefix for namespacing (e.g., "magic:")
}

// NewRedisStore creates a new Redis magic token store.
func NewRedisStore(client *redis.Client, keyPrefix string) *RedisStore {
	if keyPrefix == "" {
		keyPrefix = "magic:"
	}

	return &RedisStore{
		client: client,
		prefix: keyPrefix,
	}
}

// Key generation helpers
func (s *RedisStore) tokenKey(value string, purpose Purpose) string {
	return fmt.Sprintf("%s%s:%s", s.prefix, purpose, value)
}

func (s *RedisStore) tokenIDKey(tokenID string) string {
	return fmt.Sprintf("%sid:%s", s.prefix, tokenID)
}

func (s *RedisStore) userPurposeKey(userID string, purpose Purpose) string {
	return fmt.Sprintf("%suser:%s:%s", s.prefix, userID, purpose)
}

// Save stores a token with automatic TTL based on ExpiresAt.
func (s *RedisStore) Save(ctx context.Context, token *Token) result.Result[*Token] {
	if token == nil {
		return result.Err[*Token](ErrTokenNotFound)
	}

	// Calculate TTL
	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		return result.Err[*Token](ErrTokenExpired{
			TokenID:   token.ID,
			ExpiresAt: token.ExpiresAt,
		})
	}

	// Serialize token
	data, err := json.Marshal(token)
	if err != nil {
		return result.Err[*Token](err)
	}

	// Use pipeline for atomic operation
	pipe := s.client.Pipeline()

	// Store by token value (primary key for verification)
	pipe.Set(ctx, s.tokenKey(token.Value, token.Purpose), data, ttl)

	// Store by token ID (for direct ID lookups)
	pipe.Set(ctx, s.tokenIDKey(token.ID), data, ttl)

	// Add to user's token set for counting/listing
	pipe.SAdd(ctx, s.userPurposeKey(token.UserID, token.Purpose), token.ID)
	pipe.Expire(ctx, s.userPurposeKey(token.UserID, token.Purpose), ttl)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return result.Err[*Token](err)
	}

	return result.Ok(token)
}

// GetByValue retrieves a token by its value and purpose.
func (s *RedisStore) GetByValue(ctx context.Context, value string, purpose Purpose) result.Result[*Token] {
	data, err := s.client.Get(ctx, s.tokenKey(value, purpose)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Err[*Token](ErrTokenNotFound)
		}
		return result.Err[*Token](err)
	}

	var token Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return result.Err[*Token](err)
	}

	return result.Ok(&token)
}

// GetByID retrieves a token by its ID.
func (s *RedisStore) GetByID(ctx context.Context, tokenID string) result.Result[*Token] {
	data, err := s.client.Get(ctx, s.tokenIDKey(tokenID)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Err[*Token](ErrTokenNotFound)
		}
		return result.Err[*Token](err)
	}

	var token Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return result.Err[*Token](err)
	}

	return result.Ok(&token)
}

// MarkAsUsed marks a token as used, preventing reuse.
func (s *RedisStore) MarkAsUsed(ctx context.Context, tokenID string) result.Result[struct{}] {
	// Get current token
	tokenResult := s.GetByID(ctx, tokenID)
	if tokenResult.IsErr() {
		return result.Err[struct{}](tokenResult.UnwrapErr())
	}

	token := tokenResult.Unwrap()

	// Mark as used
	token.MarkAsUsed()

	// Calculate remaining TTL
	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		// Token expired, just delete it
		return s.Delete(ctx, tokenID)
	}

	// Serialize updated token
	data, err := json.Marshal(token)
	if err != nil {
		return result.Err[struct{}](err)
	}

	// Update both keys with the marked token
	pipe := s.client.Pipeline()
	pipe.Set(ctx, s.tokenKey(token.Value, token.Purpose), data, ttl)
	pipe.Set(ctx, s.tokenIDKey(token.ID), data, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return result.Err[struct{}](err)
	}

	return result.Ok(struct{}{})
}

// Delete removes a token from storage.
func (s *RedisStore) Delete(ctx context.Context, tokenID string) result.Result[struct{}] {
	// Get token first to clean up all keys
	tokenResult := s.GetByID(ctx, tokenID)
	if tokenResult.IsErr() {
		// If token doesn't exist, consider it deleted (idempotent)
		return result.Ok(struct{}{})
	}

	token := tokenResult.Unwrap()

	// Use pipeline for atomic operation
	pipe := s.client.Pipeline()

	// Delete by value
	pipe.Del(ctx, s.tokenKey(token.Value, token.Purpose))

	// Delete by ID
	pipe.Del(ctx, s.tokenIDKey(token.ID))

	// Remove from user's set
	pipe.SRem(ctx, s.userPurposeKey(token.UserID, token.Purpose), token.ID)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return result.Err[struct{}](err)
	}

	return result.Ok(struct{}{})
}

// DeleteByValue removes a token by its value.
func (s *RedisStore) DeleteByValue(ctx context.Context, value string, purpose Purpose) result.Result[struct{}] {
	// Get token first
	tokenResult := s.GetByValue(ctx, value, purpose)
	if tokenResult.IsErr() {
		// If token doesn't exist, consider it deleted (idempotent)
		return result.Ok(struct{}{})
	}

	token := tokenResult.Unwrap()

	// Delete by ID (which cleans up all keys)
	return s.Delete(ctx, token.ID)
}

// DeleteExpired removes all expired tokens (cleanup job).
// Note: Redis automatically expires keys with TTL, so this is mostly for cleanup of orphaned data.
func (s *RedisStore) DeleteExpired(ctx context.Context) result.Result[int] {
	// Redis handles expiration automatically via TTL
	// This method is here for interface compatibility
	// In a real implementation, you might scan for orphaned keys

	// For now, just return 0 as Redis handles this automatically
	return result.Ok(0)
}

// DeleteAllForUser removes all tokens for a specific user.
func (s *RedisStore) DeleteAllForUser(ctx context.Context, userID string) result.Result[int] {
	count := 0

	// Delete for each purpose
	for _, purpose := range []Purpose{PurposeLogin, PurposeEmailVerify, PurposePasswordReset} {
		// Get all token IDs for this user and purpose
		tokenIDs, err := s.client.SMembers(ctx, s.userPurposeKey(userID, purpose)).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return result.Err[int](err)
		}

		// Delete each token
		for _, tokenID := range tokenIDs {
			deleteResult := s.Delete(ctx, tokenID)
			if deleteResult.IsOk() {
				count++
			}
		}

		// Clean up user set
		s.client.Del(ctx, s.userPurposeKey(userID, purpose))
	}

	return result.Ok(count)
}

// CountByUser counts active (not expired, not used) tokens for a user and purpose.
func (s *RedisStore) CountByUser(ctx context.Context, userID string, purpose Purpose) result.Result[int] {
	// Get all token IDs
	tokenIDs, err := s.client.SMembers(ctx, s.userPurposeKey(userID, purpose)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Ok(0)
		}
		return result.Err[int](err)
	}

	// Count only active tokens (not expired, not used)
	activeCount := 0
	for _, tokenID := range tokenIDs {
		tokenResult := s.GetByID(ctx, tokenID)
		if tokenResult.IsOk() {
			token := tokenResult.Unwrap()
			if token.IsValid() {
				activeCount++
			}
		}
	}

	return result.Ok(activeCount)
}

// ListByUser retrieves all active tokens for a user and purpose.
func (s *RedisStore) ListByUser(ctx context.Context, userID string, purpose Purpose) result.Result[[]*Token] {
	// Get all token IDs
	tokenIDs, err := s.client.SMembers(ctx, s.userPurposeKey(userID, purpose)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Ok([]*Token{})
		}
		return result.Err[[]*Token](err)
	}

	if len(tokenIDs) == 0 {
		return result.Ok([]*Token{})
	}

	// Get all tokens
	tokens := make([]*Token, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		tokenResult := s.GetByID(ctx, tokenID)
		if tokenResult.IsOk() {
			token := tokenResult.Unwrap()
			// Only include active (valid) tokens
			if token.IsValid() {
				tokens = append(tokens, token)
			}
		}
	}

	return result.Ok(tokens)
}
