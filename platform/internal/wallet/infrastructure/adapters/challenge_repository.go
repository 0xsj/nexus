package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
)

// Compile-time interface check.
var _ domain.ChallengeRepository = (*InMemoryChallengeRepository)(nil)

// challengeEntry stores a challenge with its expiration time.
type challengeEntry struct {
	challenge string
	expiresAt time.Time
}

// InMemoryChallengeRepository stores SIWE challenges in memory with TTL.
// Suitable for single-instance dev/staging. Production would use Redis.
type InMemoryChallengeRepository struct {
	mu         sync.RWMutex
	challenges map[string]challengeEntry
}

// NewInMemoryChallengeRepository creates a new InMemoryChallengeRepository.
func NewInMemoryChallengeRepository() *InMemoryChallengeRepository {
	return &InMemoryChallengeRepository{
		challenges: make(map[string]challengeEntry),
	}
}

// SaveChallenge stores a challenge for the given address with an expiration time.
func (r *InMemoryChallengeRepository) SaveChallenge(_ context.Context, address domain.WalletAddress, challenge string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.challenges[address.String()] = challengeEntry{
		challenge: challenge,
		expiresAt: expiresAt,
	}
	return nil
}

// GetChallenge retrieves the current challenge for the given address.
// Returns ChallengeNotFound if the challenge doesn't exist or has expired.
func (r *InMemoryChallengeRepository) GetChallenge(_ context.Context, address domain.WalletAddress) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := address.String()
	entry, ok := r.challenges[key]
	if !ok {
		return "", domain.ChallengeNotFound("InMemoryChallengeRepository.GetChallenge", key)
	}

	if time.Now().After(entry.expiresAt) {
		delete(r.challenges, key)
		return "", domain.ChallengeNotFound("InMemoryChallengeRepository.GetChallenge", key)
	}

	return entry.challenge, nil
}

// DeleteChallenge removes a challenge after use or expiration.
func (r *InMemoryChallengeRepository) DeleteChallenge(_ context.Context, address domain.WalletAddress) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.challenges, address.String())
	return nil
}
