package challenge

import (
	"context"
	"sync"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/siwe"
)

// ============================================================================
// Service
// ============================================================================

// Service implements domain.ChallengeService.
// Uses in-memory storage by default. For production, use Redis.
type Service struct {
	store  Store
	config Config
}

// Config contains challenge service configuration.
type Config struct {
	// Domain is the SIWE domain (e.g., "proof.io")
	Domain string

	// URI is the SIWE URI (e.g., "https://proof.io")
	URI string

	// Statement is the human-readable message shown to users
	Statement string

	// DefaultTTL is the default challenge lifetime
	DefaultTTL time.Duration

	// NonceLength is the length of generated nonces in bytes
	NonceLength int
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		Domain:      "proof.io",
		URI:         "https://proof.io",
		Statement:   "Sign in to Proof - Verified Identity Platform",
		DefaultTTL:  5 * time.Minute,
		NonceLength: 16,
	}
}

// Ensure Service implements domain.ChallengeService.
var _ domain.ChallengeService = (*Service)(nil)

// NewService creates a new challenge service with in-memory storage.
func NewService(config Config) *Service {
	return &Service{
		store:  NewMemoryStore(),
		config: config,
	}
}

// NewServiceWithStore creates a new challenge service with custom storage.
func NewServiceWithStore(store Store, config Config) *Service {
	return &Service{
		store:  store,
		config: config,
	}
}

// ============================================================================
// Challenge Operations
// ============================================================================

// CreateChallenge creates a new authentication challenge.
func (s *Service) CreateChallenge(ctx context.Context, params domain.CreateChallengeParams) (*domain.Challenge, error) {
	const op = "challenge.Service.CreateChallenge"

	// Generate nonce
	nonce, err := siwe.GenerateNonceWithLength(s.config.NonceLength)
	if err != nil {
		return nil, domain.ErrChallengeInvalid(op, "failed to generate nonce")
	}

	// Determine TTL
	ttl := params.TTL
	if ttl == 0 {
		ttl = s.config.DefaultTTL
	}

	// Determine domain and URI
	dmn := params.Domain
	if dmn == "" {
		dmn = s.config.Domain
	}
	uri := params.URI
	if uri == "" {
		uri = s.config.URI
	}

	// Get chain ID
	chainID := getChainID(params.Chain)

	// Create SIWE message
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)

	siweMsg := siwe.NewMessage(dmn, params.Address, uri, chainID, nonce).
		WithStatement(s.config.Statement).
		WithExpirationTime(expiresAt)

	// Create domain challenge
	challenge := &domain.Challenge{
		Nonce:     nonce,
		Address:   params.Address,
		Chain:     params.Chain,
		Domain:    dmn,
		URI:       uri,
		Statement: s.config.Statement,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
		Message:   siweMsg.String(),
	}

	// Store challenge
	if err := s.store.Save(ctx, challenge); err != nil {
		return nil, err
	}

	return challenge, nil
}

// ValidateChallenge validates and consumes a challenge.
// Returns error if challenge is invalid, expired, or already used.
func (s *Service) ValidateChallenge(ctx context.Context, nonce string) (*domain.Challenge, error) {
	const op = "challenge.Service.ValidateChallenge"

	// Find challenge
	challenge, err := s.store.FindByNonce(ctx, nonce)
	if err != nil {
		if domain.IsChallengeNotFound(err) {
			return nil, domain.ErrChallengeNotFound(op, nonce)
		}
		return nil, err
	}

	// Check if expired
	if time.Now().After(challenge.ExpiresAt) {
		// Clean up expired challenge
		_ = s.store.Delete(ctx, nonce)
		return nil, domain.ErrChallengeExpired(op, nonce)
	}

	// Check if already used
	if challenge.Used {
		return nil, domain.ErrChallengeUsed(op, nonce)
	}

	// Mark as used
	challenge.Used = true
	challenge.UsedAt = timePtr(time.Now())
	if err := s.store.Save(ctx, challenge); err != nil {
		return nil, err
	}

	return challenge, nil
}

// InvalidateChallenge invalidates a challenge without validation.
func (s *Service) InvalidateChallenge(ctx context.Context, nonce string) error {
	return s.store.Delete(ctx, nonce)
}

// ============================================================================
// Store Interface
// ============================================================================

// Store defines the interface for challenge storage.
type Store interface {
	// Save stores a challenge.
	Save(ctx context.Context, challenge *domain.Challenge) error

	// FindByNonce finds a challenge by nonce.
	FindByNonce(ctx context.Context, nonce string) (*domain.Challenge, error)

	// Delete removes a challenge.
	Delete(ctx context.Context, nonce string) error

	// DeleteExpired removes all expired challenges.
	DeleteExpired(ctx context.Context) error
}

// ============================================================================
// Memory Store
// ============================================================================

// MemoryStore is an in-memory challenge store.
// Use for development/testing. Use Redis for production.
type MemoryStore struct {
	challenges map[string]*domain.Challenge
	mu         sync.RWMutex
}

// NewMemoryStore creates a new in-memory challenge store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		challenges: make(map[string]*domain.Challenge),
	}
}

// Save stores a challenge.
func (s *MemoryStore) Save(ctx context.Context, challenge *domain.Challenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.challenges[challenge.Nonce] = challenge
	return nil
}

// FindByNonce finds a challenge by nonce.
func (s *MemoryStore) FindByNonce(ctx context.Context, nonce string) (*domain.Challenge, error) {
	const op = "challenge.MemoryStore.FindByNonce"

	s.mu.RLock()
	defer s.mu.RUnlock()

	challenge, exists := s.challenges[nonce]
	if !exists {
		return nil, domain.ErrChallengeNotFound(op, nonce)
	}

	return challenge, nil
}

// Delete removes a challenge.
func (s *MemoryStore) Delete(ctx context.Context, nonce string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.challenges, nonce)
	return nil
}

// DeleteExpired removes all expired challenges.
func (s *MemoryStore) DeleteExpired(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for nonce, challenge := range s.challenges {
		if now.After(challenge.ExpiresAt) {
			delete(s.challenges, nonce)
		}
	}

	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// getChainID returns the SIWE chain ID for a domain chain.
func getChainID(chain domain.Chain) int {
	switch chain {
	case domain.ChainEthereum:
		return siwe.ChainIDEthereum
	case domain.ChainPolygon:
		return siwe.ChainIDPolygon
	case domain.ChainArbitrum:
		return siwe.ChainIDArbitrum
	case domain.ChainOptimism:
		return siwe.ChainIDOptimism
	case domain.ChainBase:
		return siwe.ChainIDBase
	default:
		return siwe.ChainIDEthereum
	}
}

// timePtr returns a pointer to a time.Time.
func timePtr(t time.Time) *time.Time {
	return &t
}
