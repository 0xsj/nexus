package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Get Wallet By ID Handler
// ============================================================================

// GetWalletByIDHandler handles GetWalletByID queries.
type GetWalletByIDHandler struct {
	reader WalletReader
}

// NewGetWalletByIDHandler creates a new GetWalletByIDHandler.
func NewGetWalletByIDHandler(reader WalletReader) *GetWalletByIDHandler {
	return &GetWalletByIDHandler{reader: reader}
}

// Handle handles the GetWalletByID query.
func (h *GetWalletByIDHandler) Handle(ctx context.Context, q *GetWalletByID) (*WalletView, error) {
	wallet, err := h.reader.GetWallet(ctx, q.WalletID)
	if err != nil {
		return nil, err
	}

	// Verify ownership if UserID provided
	if q.UserID != "" && wallet.UserID != q.UserID {
		return nil, domain.ErrWalletNotFound("GetWalletByIDHandler.Handle", q.WalletID)
	}

	return wallet, nil
}

// ============================================================================
// Get Wallet By Address Handler
// ============================================================================

// GetWalletByAddressHandler handles GetWalletByAddress queries.
type GetWalletByAddressHandler struct {
	reader WalletReader
}

// NewGetWalletByAddressHandler creates a new GetWalletByAddressHandler.
func NewGetWalletByAddressHandler(reader WalletReader) *GetWalletByAddressHandler {
	return &GetWalletByAddressHandler{reader: reader}
}

// Handle handles the GetWalletByAddress query.
func (h *GetWalletByAddressHandler) Handle(ctx context.Context, q *GetWalletByAddress) (*WalletView, error) {
	return h.reader.GetWalletByAddress(ctx, q.Address, q.ChainID)
}

// ============================================================================
// Get Wallet By DID Handler
// ============================================================================

// GetWalletByDIDHandler handles GetWalletByDID queries.
type GetWalletByDIDHandler struct {
	reader WalletReader
}

// NewGetWalletByDIDHandler creates a new GetWalletByDIDHandler.
func NewGetWalletByDIDHandler(reader WalletReader) *GetWalletByDIDHandler {
	return &GetWalletByDIDHandler{reader: reader}
}

// Handle handles the GetWalletByDID query.
func (h *GetWalletByDIDHandler) Handle(ctx context.Context, q *GetWalletByDID) (*WalletView, error) {
	return h.reader.GetWalletByDID(ctx, q.DID)
}

// ============================================================================
// List Wallets By User Handler
// ============================================================================

// ListWalletsByUserHandler handles ListWalletsByUser queries.
type ListWalletsByUserHandler struct {
	reader WalletReader
}

// NewListWalletsByUserHandler creates a new ListWalletsByUserHandler.
func NewListWalletsByUserHandler(reader WalletReader) *ListWalletsByUserHandler {
	return &ListWalletsByUserHandler{reader: reader}
}

// Handle handles the ListWalletsByUser query.
func (h *ListWalletsByUserHandler) Handle(ctx context.Context, q *ListWalletsByUser) (*WalletListView, error) {
	opts := ListWalletsOptions{
		ListOptions: ListOptions{
			Limit:     q.Limit,
			Offset:    q.Offset,
			SortBy:    q.SortBy,
			SortOrder: q.SortOrder,
		},
		Status: q.Status,
	}
	return h.reader.ListWalletsByUser(ctx, q.UserID, opts)
}

// ============================================================================
// List Wallets By Chain Handler
// ============================================================================

// ListWalletsByChainHandler handles ListWalletsByChain queries.
type ListWalletsByChainHandler struct {
	reader WalletReader
}

// NewListWalletsByChainHandler creates a new ListWalletsByChainHandler.
func NewListWalletsByChainHandler(reader WalletReader) *ListWalletsByChainHandler {
	return &ListWalletsByChainHandler{reader: reader}
}

// Handle handles the ListWalletsByChain query.
func (h *ListWalletsByChainHandler) Handle(ctx context.Context, q *ListWalletsByChain) (*WalletListView, error) {
	opts := ListOptions{
		Limit:     q.Limit,
		Offset:    q.Offset,
		SortBy:    q.SortBy,
		SortOrder: q.SortOrder,
	}
	return h.reader.ListWalletsByChain(ctx, q.ChainID, opts)
}

// ============================================================================
// Get Primary Wallet Handler
// ============================================================================

// GetPrimaryWalletHandler handles GetPrimaryWallet queries.
type GetPrimaryWalletHandler struct {
	reader WalletReader
}

// NewGetPrimaryWalletHandler creates a new GetPrimaryWalletHandler.
func NewGetPrimaryWalletHandler(reader WalletReader) *GetPrimaryWalletHandler {
	return &GetPrimaryWalletHandler{reader: reader}
}

// Handle handles the GetPrimaryWallet query.
func (h *GetPrimaryWalletHandler) Handle(ctx context.Context, q *GetPrimaryWallet) (*WalletView, error) {
	return h.reader.GetPrimaryWallet(ctx, q.UserID)
}

// ============================================================================
// Count Wallets By User Handler
// ============================================================================

// CountWalletsByUserHandler handles CountWalletsByUser queries.
type CountWalletsByUserHandler struct {
	reader WalletReader
}

// NewCountWalletsByUserHandler creates a new CountWalletsByUserHandler.
func NewCountWalletsByUserHandler(reader WalletReader) *CountWalletsByUserHandler {
	return &CountWalletsByUserHandler{reader: reader}
}

// Handle handles the CountWalletsByUser query.
func (h *CountWalletsByUserHandler) Handle(ctx context.Context, q *CountWalletsByUser) (*WalletCountView, error) {
	count, err := h.reader.CountWalletsByUser(ctx, q.UserID, q.ActiveOnly)
	if err != nil {
		return nil, err
	}

	return &WalletCountView{
		UserID: q.UserID,
		Count:  count,
	}, nil
}

// ============================================================================
// Check Wallet Exists Handler
// ============================================================================

// CheckWalletExistsHandler handles CheckWalletExists queries.
type CheckWalletExistsHandler struct {
	reader WalletReader
}

// NewCheckWalletExistsHandler creates a new CheckWalletExistsHandler.
func NewCheckWalletExistsHandler(reader WalletReader) *CheckWalletExistsHandler {
	return &CheckWalletExistsHandler{reader: reader}
}

// Handle handles the CheckWalletExists query.
func (h *CheckWalletExistsHandler) Handle(ctx context.Context, q *CheckWalletExists) (bool, error) {
	if q.Address != nil {
		return h.reader.WalletExistsByAddress(ctx, q.Address.Address, q.Address.ChainID)
	}
	if q.DID != nil {
		return h.reader.WalletExistsByDID(ctx, *q.DID)
	}
	return false, nil
}

// ============================================================================
// Handler Dependencies
// ============================================================================

// HandlerDependencies contains all dependencies needed for query handlers.
type HandlerDependencies struct {
	WalletReader    WalletReader
	ChallengeReader ChallengeReader
}

// ============================================================================
// Handler Registration
// ============================================================================

// RegisterHandlers registers all wallet query handlers with the query bus.
func RegisterHandlers(bus *cqrs.InMemoryQueryBus, deps HandlerDependencies) error {
	handlers := map[string]any{
		TypeGetWalletByID:      NewGetWalletByIDHandler(deps.WalletReader),
		TypeGetWalletByAddress: NewGetWalletByAddressHandler(deps.WalletReader),
		TypeGetWalletByDID:     NewGetWalletByDIDHandler(deps.WalletReader),
		TypeListWalletsByUser:  NewListWalletsByUserHandler(deps.WalletReader),
		TypeListWalletsByChain: NewListWalletsByChainHandler(deps.WalletReader),
		TypeGetPrimaryWallet:   NewGetPrimaryWalletHandler(deps.WalletReader),
		TypeCountWalletsByUser: NewCountWalletsByUserHandler(deps.WalletReader),
		TypeCheckWalletExists:  NewCheckWalletExistsHandler(deps.WalletReader),
	}

	for queryType, handler := range handlers {
		if err := bus.Register(queryType, handler); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.QueryHandler[*GetWalletByID, *WalletView]           = (*GetWalletByIDHandler)(nil)
	_ cqrs.QueryHandler[*GetWalletByAddress, *WalletView]      = (*GetWalletByAddressHandler)(nil)
	_ cqrs.QueryHandler[*GetWalletByDID, *WalletView]          = (*GetWalletByDIDHandler)(nil)
	_ cqrs.QueryHandler[*ListWalletsByUser, *WalletListView]   = (*ListWalletsByUserHandler)(nil)
	_ cqrs.QueryHandler[*ListWalletsByChain, *WalletListView]  = (*ListWalletsByChainHandler)(nil)
	_ cqrs.QueryHandler[*GetPrimaryWallet, *WalletView]        = (*GetPrimaryWalletHandler)(nil)
	_ cqrs.QueryHandler[*CountWalletsByUser, *WalletCountView] = (*CountWalletsByUserHandler)(nil)
	_ cqrs.QueryHandler[*CheckWalletExists, bool]              = (*CheckWalletExistsHandler)(nil)
)
