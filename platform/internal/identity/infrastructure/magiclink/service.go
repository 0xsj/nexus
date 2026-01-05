package magiclink

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/id"
)

// ============================================================================
// Configuration
// ============================================================================

// Config contains magic link service configuration.
type Config struct {
	// DefaultTTL is the default token lifetime.
	DefaultTTL time.Duration

	// TokenLength is the length of generated tokens in bytes.
	TokenLength int

	// BaseURL is the base URL for magic link verification.
	BaseURL string
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		DefaultTTL:  15 * time.Minute,
		TokenLength: 32,
		BaseURL:     "https://proof.io/auth/verify",
	}
}

// ============================================================================
// Service
// ============================================================================

// Service implements domain.MagicLinkService.
type Service struct {
	store       Store
	idGenerator id.Generator
	config      Config
}

// Ensure Service implements domain.MagicLinkService.
var _ domain.MagicLinkService = (*Service)(nil)

// NewService creates a new magic link service with in-memory storage.
func NewService(config Config, idGenerator id.Generator) *Service {
	return &Service{
		store:       NewMemoryStore(),
		idGenerator: idGenerator,
		config:      config,
	}
}

// NewServiceWithStore creates a new magic link service with custom storage.
func NewServiceWithStore(store Store, config Config, idGenerator id.Generator) *Service {
	return &Service{
		store:       store,
		idGenerator: idGenerator,
		config:      config,
	}
}

// ============================================================================
// Token Operations
// ============================================================================

// CreateToken creates a new magic link token.
func (s *Service) CreateToken(ctx context.Context, params domain.CreateMagicLinkParams) (*domain.MagicLinkToken, error) {
	const op = "magiclink.Service.CreateToken"

	// Generate token ID
	tokenID := s.idGenerator.Generate().String()

	// Generate random token
	token, err := generateSecureToken(s.config.TokenLength)
	if err != nil {
		return nil, domain.ErrMagicLinkInvalid(op, "failed to generate token")
	}

	// Hash the token for storage
	tokenHash := hashToken(token)

	// Determine TTL
	ttl := params.TTL
	if ttl == 0 {
		ttl = s.config.DefaultTTL
	}

	// Create token entity
	magicLink := domain.NewMagicLinkToken(
		tokenID,
		params.Email,
		token,
		tokenHash,
		params.Purpose,
		ttl,
		params.IPAddress,
		params.UserAgent,
	)

	// Store token
	if err := s.store.Save(ctx, magicLink); err != nil {
		return nil, err
	}

	return magicLink, nil
}

// ValidateToken validates and consumes a magic link token.
func (s *Service) ValidateToken(ctx context.Context, token string) (*domain.MagicLinkToken, error) {
	const op = "magiclink.Service.ValidateToken"

	// Hash the provided token
	tokenHash := hashToken(token)

	// Find token by hash
	magicLink, err := s.store.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if domain.IsMagicLinkNotFound(err) {
			return nil, domain.ErrMagicLinkNotFound(op, "token")
		}
		return nil, err
	}

	// Check if expired
	if magicLink.IsExpired() {
		// Update status to expired
		_ = s.store.Delete(ctx, magicLink.ID())
		return nil, domain.ErrMagicLinkExpired(op, magicLink.ID())
	}

	// Check if already used
	if magicLink.IsUsed() {
		return nil, domain.ErrMagicLinkAlreadyUsed(op, magicLink.ID())
	}

	// Mark as used
	if err := magicLink.MarkUsed(); err != nil {
		return nil, err
	}

	// Update in store
	if err := s.store.Save(ctx, magicLink); err != nil {
		return nil, err
	}

	return magicLink, nil
}

// RevokeToken revokes a magic link token.
func (s *Service) RevokeToken(ctx context.Context, tokenID string) error {
	const op = "magiclink.Service.RevokeToken"

	magicLink, err := s.store.FindByID(ctx, tokenID)
	if err != nil {
		if domain.IsMagicLinkNotFound(err) {
			return domain.ErrMagicLinkNotFound(op, tokenID)
		}
		return err
	}

	magicLink.Revoke()

	return s.store.Save(ctx, magicLink)
}

// RevokeAllForEmail revokes all pending tokens for an email.
func (s *Service) RevokeAllForEmail(ctx context.Context, email string) error {
	return s.store.DeleteAllForEmail(ctx, email)
}

// ============================================================================
// URL Generation
// ============================================================================

// GenerateVerificationURL generates the full verification URL for a token.
func (s *Service) GenerateVerificationURL(token string) string {
	return s.config.BaseURL + "?token=" + token
}

// ============================================================================
// Store Interface
// ============================================================================

// Store defines the interface for magic link token storage.
type Store interface {
	// Save stores a magic link token.
	Save(ctx context.Context, token *domain.MagicLinkToken) error

	// FindByID finds a token by ID.
	FindByID(ctx context.Context, id string) (*domain.MagicLinkToken, error)

	// FindByTokenHash finds a token by its hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*domain.MagicLinkToken, error)

	// Delete removes a token.
	Delete(ctx context.Context, id string) error

	// DeleteAllForEmail removes all tokens for an email.
	DeleteAllForEmail(ctx context.Context, email string) error

	// DeleteExpired removes all expired tokens.
	DeleteExpired(ctx context.Context) error
}

// ============================================================================
// Memory Store
// ============================================================================

// MemoryStore is an in-memory magic link token store.
type MemoryStore struct {
	tokens map[string]*domain.MagicLinkToken // keyed by ID
	mu     sync.RWMutex
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tokens: make(map[string]*domain.MagicLinkToken),
	}
}

// Save stores a magic link token.
func (s *MemoryStore) Save(ctx context.Context, token *domain.MagicLinkToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[token.ID()] = token
	return nil
}

// FindByID finds a token by ID.
func (s *MemoryStore) FindByID(ctx context.Context, id string) (*domain.MagicLinkToken, error) {
	const op = "magiclink.MemoryStore.FindByID"

	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[id]
	if !exists {
		return nil, domain.ErrMagicLinkNotFound(op, id)
	}

	return token, nil
}

// FindByTokenHash finds a token by its hash.
func (s *MemoryStore) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.MagicLinkToken, error) {
	const op = "magiclink.MemoryStore.FindByTokenHash"

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, token := range s.tokens {
		if token.TokenHash() == tokenHash {
			return token, nil
		}
	}

	return nil, domain.ErrMagicLinkNotFound(op, "hash")
}

// Delete removes a token.
func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, id)
	return nil
}

// DeleteAllForEmail removes all tokens for an email.
func (s *MemoryStore) DeleteAllForEmail(ctx context.Context, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, token := range s.tokens {
		if token.Email() == email {
			delete(s.tokens, id)
		}
	}

	return nil
}

// DeleteExpired removes all expired tokens.
func (s *MemoryStore) DeleteExpired(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, token := range s.tokens {
		if token.IsExpired() {
			delete(s.tokens, id)
		}
	}

	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA-256 hash of the token.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
