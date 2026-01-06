package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
)

// ============================================================================
// List Options
// ============================================================================

// ListOptions contains common pagination and sorting options.
type ListOptions struct {
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

// DefaultListOptions returns sensible defaults.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// ============================================================================
// Read Repository
// ============================================================================

// ReadRepository defines the interface for read-optimized queries.
// This is separate from domain repositories to allow for optimized read models.
type ReadRepository interface {
	WalletReader
	ChallengeReader
}

// ============================================================================
// Wallet Reader
// ============================================================================

// WalletReader provides read operations for wallets.
type WalletReader interface {
	// GetWallet retrieves a wallet by ID.
	GetWallet(ctx context.Context, walletID string) (*WalletView, error)

	// GetWalletByAddress retrieves a wallet by address and chain.
	GetWalletByAddress(ctx context.Context, address string, chainID domain.ChainID) (*WalletView, error)

	// GetWalletByDID retrieves a wallet by its derived DID.
	GetWalletByDID(ctx context.Context, did string) (*WalletView, error)

	// GetPrimaryWallet retrieves the primary wallet for a user.
	GetPrimaryWallet(ctx context.Context, userID string) (*WalletView, error)

	// ListWalletsByUser retrieves wallets for a user.
	ListWalletsByUser(ctx context.Context, userID string, opts ListWalletsOptions) (*WalletListView, error)

	// ListWalletsByChain retrieves wallets by chain.
	ListWalletsByChain(ctx context.Context, chainID domain.ChainID, opts ListOptions) (*WalletListView, error)

	// CountWalletsByUser counts wallets for a user.
	CountWalletsByUser(ctx context.Context, userID string, activeOnly bool) (int, error)

	// CountWalletsByChain counts wallets for a chain.
	CountWalletsByChain(ctx context.Context, chainID domain.ChainID) (int, error)

	// WalletExists checks if a wallet exists.
	WalletExists(ctx context.Context, walletID string) (bool, error)

	// WalletExistsByAddress checks if a wallet exists by address.
	WalletExistsByAddress(ctx context.Context, address string, chainID domain.ChainID) (bool, error)

	// WalletExistsByDID checks if a wallet exists by DID.
	WalletExistsByDID(ctx context.Context, did string) (bool, error)

	// GetUserWalletStats retrieves wallet statistics for a user.
	GetUserWalletStats(ctx context.Context, userID string) (*UserWalletStatsView, error)

	// GetChainStats retrieves wallet statistics for a chain.
	GetChainStats(ctx context.Context, chainID domain.ChainID) (*ChainStatsView, error)
}

// ListWalletsOptions contains options for listing wallets.
type ListWalletsOptions struct {
	ListOptions
	Status *domain.WalletStatus
}

// ============================================================================
// Challenge Reader
// ============================================================================

// ChallengeReader provides read operations for challenges.
type ChallengeReader interface {
	// GetChallenge retrieves a challenge by nonce.
	GetChallenge(ctx context.Context, nonce string) (*ChallengeView, error)

	// ChallengeExists checks if a challenge exists and is valid.
	ChallengeExists(ctx context.Context, nonce string) (bool, error)

	// ChallengeIsValid checks if a challenge is valid (not expired, not used).
	ChallengeIsValid(ctx context.Context, nonce string) (bool, error)
}

// ============================================================================
// Activity Reader
// ============================================================================

// ActivityReader provides read operations for wallet activity.
type ActivityReader interface {
	// ListWalletActivity retrieves activity for a wallet.
	ListWalletActivity(ctx context.Context, walletID string, opts ListOptions) (*WalletActivityListView, error)

	// ListUserActivity retrieves wallet activity for a user.
	ListUserActivity(ctx context.Context, userID string, opts ListOptions) (*WalletActivityListView, error)
}

// ============================================================================
// Composite Read Repository
// ============================================================================

// CompositeReadRepository combines all readers into a single repository.
type CompositeReadRepository struct {
	wallets    WalletReader
	challenges ChallengeReader
	activity   ActivityReader
}

// NewCompositeReadRepository creates a new composite read repository.
func NewCompositeReadRepository(
	wallets WalletReader,
	challenges ChallengeReader,
	activity ActivityReader,
) *CompositeReadRepository {
	return &CompositeReadRepository{
		wallets:    wallets,
		challenges: challenges,
		activity:   activity,
	}
}

// Wallets returns the wallet reader.
func (r *CompositeReadRepository) Wallets() WalletReader {
	return r.wallets
}

// Challenges returns the challenge reader.
func (r *CompositeReadRepository) Challenges() ChallengeReader {
	return r.challenges
}

// Activity returns the activity reader.
func (r *CompositeReadRepository) Activity() ActivityReader {
	return r.activity
}
