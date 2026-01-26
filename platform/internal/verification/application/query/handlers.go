package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Handler Dependencies
// ============================================================================

// HandlerDependencies contains all dependencies for query handlers.
type HandlerDependencies struct {
	ReadRepo domain.VerificationReadRepository
}

// ============================================================================
// Get Verification Handler
// ============================================================================

// GetVerificationHandler handles GetVerification queries.
type GetVerificationHandler struct {
	readRepo domain.VerificationReadRepository
}

// NewGetVerificationHandler creates a new GetVerificationHandler.
func NewGetVerificationHandler(readRepo domain.VerificationReadRepository) *GetVerificationHandler {
	return &GetVerificationHandler{
		readRepo: readRepo,
	}
}

// Handle handles the GetVerification query.
func (h *GetVerificationHandler) Handle(ctx context.Context, q *GetVerification) (*domain.VerificationView, error) {
	return h.readRepo.GetByID(ctx, q.VerificationID)
}

// ============================================================================
// Get Verification By State Handler
// ============================================================================

// GetVerificationByStateHandler handles GetVerificationByState queries.
type GetVerificationByStateHandler struct {
	readRepo domain.VerificationReadRepository
}

// NewGetVerificationByStateHandler creates a new GetVerificationByStateHandler.
func NewGetVerificationByStateHandler(readRepo domain.VerificationReadRepository) *GetVerificationByStateHandler {
	return &GetVerificationByStateHandler{
		readRepo: readRepo,
	}
}

// Handle handles the GetVerificationByState query.
func (h *GetVerificationByStateHandler) Handle(ctx context.Context, q *GetVerificationByState) (*domain.VerificationView, error) {
	// Note: This requires adding a method to the read repository
	// For now, we'll return not found - implementation depends on read repo
	return nil, domain.ErrVerificationNotFound("GetVerificationByStateHandler.Handle", q.State)
}

// ============================================================================
// List Verifications By User Handler
// ============================================================================

// ListVerificationsByUserHandler handles ListVerificationsByUser queries.
type ListVerificationsByUserHandler struct {
	readRepo domain.VerificationReadRepository
}

// NewListVerificationsByUserHandler creates a new ListVerificationsByUserHandler.
func NewListVerificationsByUserHandler(readRepo domain.VerificationReadRepository) *ListVerificationsByUserHandler {
	return &ListVerificationsByUserHandler{
		readRepo: readRepo,
	}
}

// Handle handles the ListVerificationsByUser query.
func (h *ListVerificationsByUserHandler) Handle(ctx context.Context, q *ListVerificationsByUser) (*VerificationListResult, error) {
	opts := domain.VerificationListOptions{
		Limit:    q.Limit,
		Offset:   q.Offset,
		SortBy:   q.SortBy,
		SortDesc: q.SortDesc,
	}

	items, total, err := h.readRepo.GetByUser(ctx, q.UserID, opts)
	if err != nil {
		return nil, err
	}

	return &VerificationListResult{
		Items:   items,
		Total:   total,
		Limit:   q.Limit,
		Offset:  q.Offset,
		HasMore: q.Offset+len(items) < total,
	}, nil
}

// ============================================================================
// Get Provider Connections Handler
// ============================================================================

// GetProviderConnectionsHandler handles GetProviderConnections queries.
type GetProviderConnectionsHandler struct {
	readRepo domain.VerificationReadRepository
}

// NewGetProviderConnectionsHandler creates a new GetProviderConnectionsHandler.
func NewGetProviderConnectionsHandler(readRepo domain.VerificationReadRepository) *GetProviderConnectionsHandler {
	return &GetProviderConnectionsHandler{
		readRepo: readRepo,
	}
}

// Handle handles the GetProviderConnections query.
func (h *GetProviderConnectionsHandler) Handle(ctx context.Context, q *GetProviderConnections) ([]*domain.ProviderConnectionView, error) {
	return h.readRepo.GetProviderConnections(ctx, q.UserID)
}

// ============================================================================
// Get Latest By User And Provider Handler
// ============================================================================

// GetLatestByUserAndProviderHandler handles GetLatestByUserAndProvider queries.
type GetLatestByUserAndProviderHandler struct {
	readRepo domain.VerificationReadRepository
}

// NewGetLatestByUserAndProviderHandler creates a new GetLatestByUserAndProviderHandler.
func NewGetLatestByUserAndProviderHandler(readRepo domain.VerificationReadRepository) *GetLatestByUserAndProviderHandler {
	return &GetLatestByUserAndProviderHandler{
		readRepo: readRepo,
	}
}

// Handle handles the GetLatestByUserAndProvider query.
func (h *GetLatestByUserAndProviderHandler) Handle(ctx context.Context, q *GetLatestByUserAndProvider) (*domain.VerificationView, error) {
	return h.readRepo.GetLatestByUserAndProvider(ctx, q.UserID, q.Provider)
}

// ============================================================================
// Check Provider Connected Handler
// ============================================================================

// CheckProviderConnectedHandler handles CheckProviderConnected queries.
type CheckProviderConnectedHandler struct {
	readRepo domain.VerificationReadRepository
}

// NewCheckProviderConnectedHandler creates a new CheckProviderConnectedHandler.
func NewCheckProviderConnectedHandler(readRepo domain.VerificationReadRepository) *CheckProviderConnectedHandler {
	return &CheckProviderConnectedHandler{
		readRepo: readRepo,
	}
}

// Handle handles the CheckProviderConnected query.
func (h *CheckProviderConnectedHandler) Handle(ctx context.Context, q *CheckProviderConnected) (*CheckProviderConnectedResult, error) {
	verification, err := h.readRepo.GetLatestByUserAndProvider(ctx, q.UserID, q.Provider)
	if err != nil {
		if domain.IsVerificationNotFound(err) {
			return &CheckProviderConnectedResult{
				Connected: false,
			}, nil
		}
		return nil, err
	}

	// Check if the latest verification was completed
	if verification.Status != string(domain.StatusCompleted) {
		return &CheckProviderConnectedResult{
			Connected: false,
		}, nil
	}

	return &CheckProviderConnectedResult{
		Connected:    true,
		CredentialID: &verification.CredentialID,
		Username:     &verification.Username,
	}, nil
}

// ============================================================================
// Handler Registration
// ============================================================================

// RegisterHandlers registers all verification query handlers with the query bus.
func RegisterHandlers(bus *cqrs.InMemoryQueryBus, deps HandlerDependencies) error {
	handlers := map[string]any{
		TypeGetVerification:            NewGetVerificationHandler(deps.ReadRepo),
		TypeGetVerificationByState:     NewGetVerificationByStateHandler(deps.ReadRepo),
		TypeListVerificationsByUser:    NewListVerificationsByUserHandler(deps.ReadRepo),
		TypeGetProviderConnections:     NewGetProviderConnectionsHandler(deps.ReadRepo),
		TypeGetLatestByUserAndProvider: NewGetLatestByUserAndProviderHandler(deps.ReadRepo),
		TypeCheckProviderConnected:     NewCheckProviderConnectedHandler(deps.ReadRepo),
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
	_ cqrs.QueryHandler[*GetVerification, *domain.VerificationView]                = (*GetVerificationHandler)(nil)
	_ cqrs.QueryHandler[*GetVerificationByState, *domain.VerificationView]         = (*GetVerificationByStateHandler)(nil)
	_ cqrs.QueryHandler[*ListVerificationsByUser, *VerificationListResult]         = (*ListVerificationsByUserHandler)(nil)
	_ cqrs.QueryHandler[*GetProviderConnections, []*domain.ProviderConnectionView] = (*GetProviderConnectionsHandler)(nil)
	_ cqrs.QueryHandler[*GetLatestByUserAndProvider, *domain.VerificationView]     = (*GetLatestByUserAndProviderHandler)(nil)
	_ cqrs.QueryHandler[*CheckProviderConnected, *CheckProviderConnectedResult]    = (*CheckProviderConnectedHandler)(nil)
)
