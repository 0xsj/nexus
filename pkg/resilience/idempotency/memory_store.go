package idempotency

import (
	"context"
	"sync"
	"time"
)

// entry represents a stored idempotency entry.
type entry struct {
	result    interface{}
	expiresAt time.Time
}

// MemoryStore implements Store using in-memory storage.
// Useful for testing and development.
type MemoryStore struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

// NewMemoryStore creates a new in-memory idempotency store.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		entries: make(map[string]*entry),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Exists checks if an idempotency key has been processed.
func (s *MemoryStore) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.entries[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if time.Now().After(e.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Save stores an idempotency key with optional result.
func (s *MemoryStore) Save(ctx context.Context, key string, result interface{}, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = &entry{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Get retrieves the stored result for an idempotency key.
func (s *MemoryStore) Get(ctx context.Context, key string) (interface{}, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.entries[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	// Check if expired
	if time.Now().After(e.expiresAt) {
		return nil, ErrKeyNotFound
	}

	return e.result, nil
}

// Delete removes an idempotency key.
func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
	return nil
}

// cleanup periodically removes expired entries.
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, e := range s.entries {
			if now.After(e.expiresAt) {
				delete(s.entries, key)
			}
		}
		s.mu.Unlock()
	}
}
