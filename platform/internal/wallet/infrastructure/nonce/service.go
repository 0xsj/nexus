package nonce

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
)

// ============================================================================
// In-Memory Nonce Service
// ============================================================================

// InMemoryService implements domain.NonceService using in-memory storage.
// Suitable for development and single-instance deployments.
// For production with multiple instances, use a Redis-backed implementation.
type InMemoryService struct {
	mu              sync.RWMutex
	nonces          map[string]*nonceEntry
	ttl             time.Duration
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// nonceEntry tracks a nonce and its metadata.
type nonceEntry struct {
	nonce     string
	createdAt time.Time
	expiresAt time.Time
	used      bool
	usedAt    *time.Time
}

// InMemoryConfig contains configuration for the in-memory nonce service.
type InMemoryConfig struct {
	// TTL is how long a nonce is valid.
	TTL time.Duration

	// CleanupInterval is how often to clean up expired nonces.
	CleanupInterval time.Duration

	// NonceLength is the length of generated nonces in bytes (before encoding).
	NonceLength int
}

// DefaultInMemoryConfig returns default configuration.
func DefaultInMemoryConfig() InMemoryConfig {
	return InMemoryConfig{
		TTL:             10 * time.Minute,
		CleanupInterval: 5 * time.Minute,
		NonceLength:     32,
	}
}

// NewInMemoryService creates a new in-memory nonce service.
func NewInMemoryService(config InMemoryConfig) *InMemoryService {
	if config.TTL == 0 {
		config.TTL = 10 * time.Minute
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	s := &InMemoryService{
		nonces:          make(map[string]*nonceEntry),
		ttl:             config.TTL,
		cleanupInterval: config.CleanupInterval,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup
	go s.cleanup()

	return s
}

// Generate generates a new unique nonce.
func (s *InMemoryService) Generate(ctx context.Context) (string, error) {
	const op = "InMemoryService.Generate"

	// Check context
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", domain.ErrInvalidNonce(op, "failed to generate random nonce")
	}

	// Encode as URL-safe base64
	nonce := base64.RawURLEncoding.EncodeToString(bytes)

	// Store nonce
	now := time.Now()
	entry := &nonceEntry{
		nonce:     nonce,
		createdAt: now,
		expiresAt: now.Add(s.ttl),
		used:      false,
	}

	s.mu.Lock()
	s.nonces[nonce] = entry
	s.mu.Unlock()

	return nonce, nil
}

// Validate validates that a nonce is valid and unused.
func (s *InMemoryService) Validate(ctx context.Context, nonce string) error {
	const op = "InMemoryService.Validate"

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.RLock()
	entry, exists := s.nonces[nonce]
	s.mu.RUnlock()

	if !exists {
		return domain.ErrInvalidNonce(op, nonce)
	}

	if entry.used {
		return domain.ErrSignatureUsed(op, nonce)
	}

	if time.Now().After(entry.expiresAt) {
		return domain.ErrSignatureExpired(op)
	}

	return nil
}

// Consume marks a nonce as used.
func (s *InMemoryService) Consume(ctx context.Context, nonce string) error {
	const op = "InMemoryService.Consume"

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.nonces[nonce]
	if !exists {
		return domain.ErrInvalidNonce(op, nonce)
	}

	if entry.used {
		return domain.ErrSignatureUsed(op, nonce)
	}

	if time.Now().After(entry.expiresAt) {
		return domain.ErrSignatureExpired(op)
	}

	now := time.Now()
	entry.used = true
	entry.usedAt = &now

	return nil
}

// IsUsed returns true if the nonce has been used.
func (s *InMemoryService) IsUsed(ctx context.Context, nonce string) (bool, error) {
	// Check context
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	s.mu.RLock()
	entry, exists := s.nonces[nonce]
	s.mu.RUnlock()

	if !exists {
		return false, nil
	}

	return entry.used, nil
}

// cleanup periodically removes expired nonces.
func (s *InMemoryService) cleanup() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeExpired()
		case <-s.stopCleanup:
			return
		}
	}
}

// removeExpired removes all expired nonces.
func (s *InMemoryService) removeExpired() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for nonce, entry := range s.nonces {
		// Remove if expired and used, or if expired for more than 2x TTL
		if entry.used && now.After(entry.expiresAt) {
			delete(s.nonces, nonce)
		} else if now.After(entry.expiresAt.Add(s.ttl)) {
			// Remove very old unused nonces (2x TTL)
			delete(s.nonces, nonce)
		}
	}
}

// Stop stops the background cleanup goroutine.
func (s *InMemoryService) Stop() {
	close(s.stopCleanup)
}

// Stats returns statistics about the nonce store.
func (s *InMemoryService) Stats() NonceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total, used, expired int
	now := time.Now()

	for _, entry := range s.nonces {
		total++
		if entry.used {
			used++
		}
		if now.After(entry.expiresAt) {
			expired++
		}
	}

	return NonceStats{
		Total:   total,
		Used:    used,
		Expired: expired,
		Active:  total - used - expired,
	}
}

// NonceStats contains statistics about the nonce store.
type NonceStats struct {
	Total   int `json:"total"`
	Used    int `json:"used"`
	Expired int `json:"expired"`
	Active  int `json:"active"`
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.NonceService = (*InMemoryService)(nil)
