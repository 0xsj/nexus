package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements Store using Redis.
type RedisStore struct {
	client *redis.Client
	prefix string
}

// NewRedisStore creates a new Redis-backed idempotency store.
func NewRedisStore(client *redis.Client, prefix string) *RedisStore {
	if prefix == "" {
		prefix = "idempotency:"
	}
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

// Exists checks if an idempotency key has been processed.
func (s *RedisStore) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	redisKey := s.makeKey(key)
	exists, err := s.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}

	return exists > 0, nil
}

// Save stores an idempotency key with optional result.
func (s *RedisStore) Save(ctx context.Context, key string, result interface{}, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	redisKey := s.makeKey(key)

	// Serialize result to JSON
	var value string
	if result != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		value = string(data)
	} else {
		value = "processed" // Just mark as processed with no result
	}

	// Store with TTL
	if err := s.client.Set(ctx, redisKey, value, ttl).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}

	return nil
}

// Get retrieves the stored result for an idempotency key.
func (s *RedisStore) Get(ctx context.Context, key string) (interface{}, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	redisKey := s.makeKey(key)
	value, err := s.client.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}

	// If it's just a marker, return nil
	if value == "processed" {
		return nil, nil
	}

	// Try to unmarshal as JSON
	var result interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		// If unmarshal fails, return raw string
		return value, nil
	}

	return result, nil
}

// Delete removes an idempotency key.
func (s *RedisStore) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	redisKey := s.makeKey(key)
	if err := s.client.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}

	return nil
}

// makeKey creates the full Redis key with prefix.
func (s *RedisStore) makeKey(key string) string {
	return fmt.Sprintf("%s%s", s.prefix, key)
}
