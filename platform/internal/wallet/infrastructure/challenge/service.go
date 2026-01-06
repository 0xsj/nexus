package challenge

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
)

// ============================================================================
// Challenge Service
// ============================================================================

// Service implements domain.ChallengeService.
// It creates and manages SIWE authentication challenges.
type Service struct {
	mu           sync.RWMutex
	challenges   map[string]*domain.Challenge
	nonceService domain.NonceService
	config       Config
	stopCleanup  chan struct{}
}

// Config contains configuration for the challenge service.
type Config struct {
	// TTL is how long a challenge is valid.
	TTL time.Duration

	// CleanupInterval is how often to clean up expired challenges.
	CleanupInterval time.Duration

	// DefaultStatement is the default statement for SIWE messages.
	DefaultStatement string

	// Version is the SIWE message version.
	Version string
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		TTL:              10 * time.Minute,
		CleanupInterval:  5 * time.Minute,
		DefaultStatement: "Sign in with your wallet",
		Version:          "1",
	}
}

// NewService creates a new challenge service.
func NewService(nonceService domain.NonceService, config Config) *Service {
	if config.TTL == 0 {
		config.TTL = 10 * time.Minute
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}
	if config.Version == "" {
		config.Version = "1"
	}

	s := &Service{
		challenges:   make(map[string]*domain.Challenge),
		nonceService: nonceService,
		config:       config,
		stopCleanup:  make(chan struct{}),
	}

	// Start background cleanup
	go s.cleanup()

	return s
}

// CreateChallenge creates a new authentication challenge.
func (s *Service) CreateChallenge(ctx context.Context, params domain.CreateChallengeParams) (*domain.Challenge, error) {
	const op = "ChallengeService.CreateChallenge"

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate address
	if params.Address.IsZero() {
		return nil, domain.ErrInvalidAddress(op, "", "address is required")
	}

	// Validate domain
	if params.Domain == "" {
		return nil, domain.ErrVerificationFailed(op, "domain is required")
	}

	// Validate URI
	if params.URI == "" {
		return nil, domain.ErrVerificationFailed(op, "URI is required")
	}

	// Generate nonce
	nonce, err := s.nonceService.Generate(ctx)
	if err != nil {
		return nil, err
	}

	// Build timestamps
	now := time.Now().UTC()
	expiresAt := now.Add(s.config.TTL)

	// Build SIWE message
	message := s.buildSIWEMessage(params, nonce, now, expiresAt)

	// Create challenge
	challenge := &domain.Challenge{
		Nonce:     nonce,
		Message:   message,
		Address:   params.Address,
		Domain:    params.Domain,
		URI:       params.URI,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	// Store challenge
	s.mu.Lock()
	s.challenges[nonce] = challenge
	s.mu.Unlock()

	return challenge, nil
}

// ValidateChallenge validates and consumes a challenge.
func (s *Service) ValidateChallenge(ctx context.Context, nonce string) (*domain.Challenge, error) {
	const op = "ChallengeService.ValidateChallenge"

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate nonce with nonce service
	if err := s.nonceService.Validate(ctx, nonce); err != nil {
		return nil, err
	}

	// Get challenge
	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, exists := s.challenges[nonce]
	if !exists {
		return nil, domain.ErrInvalidNonce(op, nonce)
	}

	// Check if already used
	if challenge.Used {
		return nil, domain.ErrSignatureUsed(op, nonce)
	}

	// Check expiration
	if challenge.IsExpired() {
		return nil, domain.ErrSignatureExpired(op)
	}

	// Consume nonce
	if err := s.nonceService.Consume(ctx, nonce); err != nil {
		return nil, err
	}

	// Mark challenge as used
	now := time.Now()
	challenge.Used = true
	challenge.UsedAt = &now

	return challenge, nil
}

// InvalidateChallenge invalidates a challenge without consuming it.
func (s *Service) InvalidateChallenge(ctx context.Context, nonce string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, exists := s.challenges[nonce]
	if !exists {
		// Already gone, that's fine
		return nil
	}

	// Mark as used without consuming the nonce
	now := time.Now()
	challenge.Used = true
	challenge.UsedAt = &now

	return nil
}

// ============================================================================
// SIWE Message Building
// ============================================================================

// buildSIWEMessage builds an EIP-4361 compliant SIWE message.
func (s *Service) buildSIWEMessage(
	params domain.CreateChallengeParams,
	nonce string,
	issuedAt time.Time,
	expiresAt time.Time,
) string {
	// Get statement
	statement := params.Statement
	if statement == "" {
		statement = s.config.DefaultStatement
	}

	// Get chain ID for EVM, default to "1" for Ethereum mainnet
	chainID := "1"
	if !params.Address.ChainID().IsEmpty() {
		chainID = s.chainIDToEIP155(params.Address.ChainID())
	}

	// Build message according to EIP-4361 format
	// https://eips.ethereum.org/EIPS/eip-4361
	message := fmt.Sprintf(
		"%s wants you to sign in with your Ethereum account:\n%s\n\n%s\n\nURI: %s\nVersion: %s\nChain ID: %s\nNonce: %s\nIssued At: %s\nExpiration Time: %s",
		params.Domain,
		params.Address.Normalized(),
		statement,
		params.URI,
		s.config.Version,
		chainID,
		nonce,
		issuedAt.Format(time.RFC3339),
		expiresAt.Format(time.RFC3339),
	)

	// Add resources if provided
	if len(params.Resources) > 0 {
		message += "\nResources:"
		for _, resource := range params.Resources {
			message += fmt.Sprintf("\n- %s", resource)
		}
	}

	return message
}

// chainIDToEIP155 converts a domain.ChainID to EIP-155 chain ID string.
func (s *Service) chainIDToEIP155(chainID domain.ChainID) string {
	switch chainID {
	case domain.ChainIDEthereumMainnet:
		return "1"
	case domain.ChainIDEthereumSepolia:
		return "11155111"
	case domain.ChainIDPolygon:
		return "137"
	case domain.ChainIDArbitrum:
		return "42161"
	case domain.ChainIDOptimism:
		return "10"
	case domain.ChainIDBase:
		return "8453"
	default:
		// For unknown chains, try to use the chain ID string directly
		return chainID.String()
	}
}

// ============================================================================
// Background Cleanup
// ============================================================================

// cleanup periodically removes expired challenges.
func (s *Service) cleanup() {
	ticker := time.NewTicker(s.config.CleanupInterval)
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

// removeExpired removes all expired challenges.
func (s *Service) removeExpired() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for nonce, challenge := range s.challenges {
		// Remove if used and expired, or if expired for more than 2x TTL
		if challenge.Used && now.After(challenge.ExpiresAt) {
			delete(s.challenges, nonce)
		} else if now.After(challenge.ExpiresAt.Add(s.config.TTL)) {
			// Remove very old unused challenges (2x TTL)
			delete(s.challenges, nonce)
		}
	}
}

// Stop stops the background cleanup goroutine.
func (s *Service) Stop() {
	close(s.stopCleanup)
}

// ============================================================================
// Statistics
// ============================================================================

// Stats returns statistics about the challenge store.
func (s *Service) Stats() ChallengeStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total, used, expired, active int
	now := time.Now()

	for _, challenge := range s.challenges {
		total++
		if challenge.Used {
			used++
		}
		if now.After(challenge.ExpiresAt) {
			expired++
		}
		if !challenge.Used && !now.After(challenge.ExpiresAt) {
			active++
		}
	}

	return ChallengeStats{
		Total:   total,
		Used:    used,
		Expired: expired,
		Active:  active,
	}
}

// ChallengeStats contains statistics about the challenge store.
type ChallengeStats struct {
	Total   int `json:"total"`
	Used    int `json:"used"`
	Expired int `json:"expired"`
	Active  int `json:"active"`
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.ChallengeService = (*Service)(nil)
