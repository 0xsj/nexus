package session

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
	prefix string // Key prefix for namespacing (e.g., "session:")
}

// NewRedisStore creates a new Redis session store.
func NewRedisStore(client *redis.Client, keyPrefix string) *RedisStore {
	if keyPrefix == "" {
		keyPrefix = "session:"
	}

	return &RedisStore{
		client: client,
		prefix: keyPrefix,
	}
}

// Key generation helpers
func (s *RedisStore) sessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s", s.prefix, sessionID)
}

func (s *RedisStore) tokenKey(refreshToken string) string {
	return fmt.Sprintf("%stoken:%s", s.prefix, refreshToken)
}

func (s *RedisStore) userKey(userID string) string {
	return fmt.Sprintf("%suser:%s", s.prefix, userID)
}

// Create creates a new session.
func (s *RedisStore) Create(ctx context.Context, session *Session) result.Result[*Session] {
	if session == nil {
		return result.Err[*Session](ErrSessionNotFound)
	}

	// Serialize session
	data, err := json.Marshal(session)
	if err != nil {
		return result.Err[*Session](err)
	}

	// Calculate TTL
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return result.Err[*Session](ErrSessionExpired)
	}

	// Use pipeline for atomic operation
	pipe := s.client.Pipeline()

	// Store session data with TTL
	pipe.Set(ctx, s.sessionKey(session.ID), data, ttl)

	// Map refresh token -> session ID
	pipe.Set(ctx, s.tokenKey(session.RefreshToken), session.ID, ttl)

	// Add to user's session set
	pipe.SAdd(ctx, s.userKey(session.UserID), session.ID)
	pipe.Expire(ctx, s.userKey(session.UserID), ttl)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return result.Err[*Session](err)
	}

	return result.Ok(session)
}

// GetByID retrieves a session by its ID.
func (s *RedisStore) GetByID(ctx context.Context, sessionID string) result.Result[*Session] {
	data, err := s.client.Get(ctx, s.sessionKey(sessionID)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Err[*Session](ErrSessionNotFound)
		}
		return result.Err[*Session](err)
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return result.Err[*Session](err)
	}

	return result.Ok(&session)
}

// GetByRefreshToken retrieves a session by its refresh token.
func (s *RedisStore) GetByRefreshToken(ctx context.Context, refreshToken string) result.Result[*Session] {
	// Get session ID from token mapping
	sessionID, err := s.client.Get(ctx, s.tokenKey(refreshToken)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Err[*Session](ErrSessionNotFound)
		}
		return result.Err[*Session](err)
	}

	// Get session by ID
	return s.GetByID(ctx, sessionID)
}

// Update updates an existing session.
func (s *RedisStore) Update(ctx context.Context, session *Session) result.Result[*Session] {
	if session == nil {
		return result.Err[*Session](ErrSessionNotFound)
	}

	// Check if session exists
	exists, err := s.client.Exists(ctx, s.sessionKey(session.ID)).Result()
	if err != nil {
		return result.Err[*Session](err)
	}
	if exists == 0 {
		return result.Err[*Session](ErrSessionNotFound)
	}

	// Get old session to clean up old token mapping
	oldSessionResult := s.GetByID(ctx, session.ID)
	if oldSessionResult.IsErr() {
		return result.Err[*Session](oldSessionResult.UnwrapErr())
	}
	oldSession := oldSessionResult.Unwrap()

	// Serialize new session
	data, err := json.Marshal(session)
	if err != nil {
		return result.Err[*Session](err)
	}

	// Calculate TTL
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return result.Err[*Session](ErrSessionExpired)
	}

	// Use pipeline for atomic operation
	pipe := s.client.Pipeline()

	// Update session data
	pipe.Set(ctx, s.sessionKey(session.ID), data, ttl)

	// If refresh token changed, update mappings
	if oldSession.RefreshToken != session.RefreshToken {
		pipe.Del(ctx, s.tokenKey(oldSession.RefreshToken))
		pipe.Set(ctx, s.tokenKey(session.RefreshToken), session.ID, ttl)
	}

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return result.Err[*Session](err)
	}

	return result.Ok(session)
}

// Delete deletes a session by ID.
func (s *RedisStore) Delete(ctx context.Context, sessionID string) result.Result[struct{}] {
	// Get session first to clean up all mappings
	sessionResult := s.GetByID(ctx, sessionID)
	if sessionResult.IsErr() {
		// If session doesn't exist, consider it deleted (idempotent)
		return result.Ok(struct{}{})
	}

	session := sessionResult.Unwrap()

	// Use pipeline for atomic operation
	pipe := s.client.Pipeline()

	// Delete session data
	pipe.Del(ctx, s.sessionKey(sessionID))

	// Delete token mapping
	pipe.Del(ctx, s.tokenKey(session.RefreshToken))

	// Remove from user's session set
	pipe.SRem(ctx, s.userKey(session.UserID), sessionID)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return result.Err[struct{}](err)
	}

	return result.Ok(struct{}{})
}

// DeleteByRefreshToken deletes a session by its refresh token.
func (s *RedisStore) DeleteByRefreshToken(ctx context.Context, refreshToken string) result.Result[struct{}] {
	// Get session ID from token mapping
	sessionID, err := s.client.Get(ctx, s.tokenKey(refreshToken)).Result()
	if err != nil {
		if err == redis.Nil {
			// Token doesn't exist, consider it deleted (idempotent)
			return result.Ok(struct{}{})
		}
		return result.Err[struct{}](err)
	}

	// Delete the session
	return s.Delete(ctx, sessionID)
}

// DeleteAllForUser deletes all sessions for a specific user.
func (s *RedisStore) DeleteAllForUser(ctx context.Context, userID string) result.Result[int] {
	// Get all session IDs for the user
	sessionIDs, err := s.client.SMembers(ctx, s.userKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Ok(0)
		}
		return result.Err[int](err)
	}

	if len(sessionIDs) == 0 {
		return result.Ok(0)
	}

	// Delete each session
	count := 0
	for _, sessionID := range sessionIDs {
		deleteResult := s.Delete(ctx, sessionID)
		if deleteResult.IsOk() {
			count++
		}
	}

	// Clean up user set
	s.client.Del(ctx, s.userKey(userID))

	return result.Ok(count)
}

// ListByUser retrieves all active sessions for a user.
func (s *RedisStore) ListByUser(ctx context.Context, userID string) result.Result[[]*Session] {
	// Get all session IDs for the user
	sessionIDs, err := s.client.SMembers(ctx, s.userKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Ok([]*Session{})
		}
		return result.Err[[]*Session](err)
	}

	if len(sessionIDs) == 0 {
		return result.Ok([]*Session{})
	}

	// Get all sessions
	sessions := make([]*Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		sessionResult := s.GetByID(ctx, sessionID)
		if sessionResult.IsOk() {
			sessions = append(sessions, sessionResult.Unwrap())
		}
	}

	return result.Ok(sessions)
}

// CountByUser counts active sessions for a user.
func (s *RedisStore) CountByUser(ctx context.Context, userID string) result.Result[int] {
	count, err := s.client.SCard(ctx, s.userKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return result.Ok(0)
		}
		return result.Err[int](err)
	}

	return result.Ok(int(count))
}

// DeleteExpired removes all expired sessions.
// Note: Redis automatically expires keys with TTL, so this is mostly for cleanup of orphaned data.
func (s *RedisStore) DeleteExpired(ctx context.Context) result.Result[int] {
	// Redis handles expiration automatically via TTL
	// This method is here for interface compatibility
	// In a real implementation, you might scan for orphaned keys

	// For now, just return 0 as Redis handles this automatically
	return result.Ok(0)
}

// ExtendExpiry extends a session's expiry time.
func (s *RedisStore) ExtendExpiry(ctx context.Context, sessionID string, duration time.Duration) result.Result[*Session] {
	// Get session
	sessionResult := s.GetByID(ctx, sessionID)
	if sessionResult.IsErr() {
		return sessionResult
	}

	session := sessionResult.Unwrap()

	// Extend expiry
	session.Extend(duration)

	// Update in Redis
	return s.Update(ctx, session)
}
