package command

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/credential/aggregate"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Repository Interface
// ============================================================================

// CredentialRepository defines the interface for credential persistence.
type CredentialRepository interface {
	// Load loads a credential aggregate by ID.
	Load(ctx context.Context, id string) (*aggregate.Credential, error)

	// Save saves a credential aggregate.
	Save(ctx context.Context, credential *aggregate.Credential) error

	// Exists checks if a credential exists.
	Exists(ctx context.Context, id string) (bool, error)
}

// ============================================================================
// Request Credential Handler
// ============================================================================

// RequestCredentialHandler handles RequestCredential commands.
type RequestCredentialHandler struct {
	repo CredentialRepository
}

// NewRequestCredentialHandler creates a new RequestCredentialHandler.
func NewRequestCredentialHandler(repo CredentialRepository) *RequestCredentialHandler {
	return &RequestCredentialHandler{repo: repo}
}

// Handle handles the RequestCredential command.
func (h *RequestCredentialHandler) Handle(ctx context.Context, cmd *RequestCredential) (*cqrs.CommandResult, error) {
	// Check if credential already exists
	exists, err := h.repo.Exists(ctx, cmd.CredentialID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, eventsourcing.ErrAggregateValidation(
			"RequestCredentialHandler",
			"credential already exists: "+cmd.CredentialID,
		)
	}

	// Create new aggregate
	credential := aggregate.NewCredential(cmd.CredentialID)

	// Execute domain logic
	if err := credential.Request(cmd.HolderDID, cmd.IssuerDID, cmd.CredentialType, cmd.Claims); err != nil {
		return nil, err
	}

	// Persist
	if err := h.repo.Save(ctx, credential); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.CredentialID,
		Version: credential.Version(),
	}, nil
}

// ============================================================================
// Issue Credential Handler
// ============================================================================

// IssueCredentialHandler handles IssueCredential commands.
type IssueCredentialHandler struct {
	repo CredentialRepository
}

// NewIssueCredentialHandler creates a new IssueCredentialHandler.
func NewIssueCredentialHandler(repo CredentialRepository) *IssueCredentialHandler {
	return &IssueCredentialHandler{repo: repo}
}

// Handle handles the IssueCredential command.
func (h *IssueCredentialHandler) Handle(ctx context.Context, cmd *IssueCredential) (*cqrs.CommandResult, error) {
	// Try to load existing credential (might be from a request)
	credential, err := h.repo.Load(ctx, cmd.CredentialID)
	if err != nil {
		// If not found, create new aggregate for direct issuance
		if eventsourcing.IsAggregateNotFound(err) {
			credential = aggregate.NewCredential(cmd.CredentialID)
		} else {
			return nil, err
		}
	}

	// Prepare claims with holder and type info for direct issuance
	claims := make(map[string]any, len(cmd.Claims)+2)
	for k, v := range cmd.Claims {
		claims[k] = v
	}

	// Add holder and type if this is a direct issuance (no prior request)
	if credential.Status() == "" {
		claims["holder_did"] = cmd.HolderDID
		claims["credential_type"] = cmd.CredentialType
	}

	// Execute domain logic
	if err := credential.Issue(cmd.IssuerDID, claims, cmd.ExpiresAt); err != nil {
		return nil, err
	}

	// Persist
	if err := h.repo.Save(ctx, credential); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.CredentialID,
		Version: credential.Version(),
	}, nil
}

// ============================================================================
// Revoke Credential Handler
// ============================================================================

// RevokeCredentialHandler handles RevokeCredential commands.
type RevokeCredentialHandler struct {
	repo CredentialRepository
}

// NewRevokeCredentialHandler creates a new RevokeCredentialHandler.
func NewRevokeCredentialHandler(repo CredentialRepository) *RevokeCredentialHandler {
	return &RevokeCredentialHandler{repo: repo}
}

// Handle handles the RevokeCredential command.
func (h *RevokeCredentialHandler) Handle(ctx context.Context, cmd *RevokeCredential) (*cqrs.CommandResult, error) {
	// Load credential
	credential, err := h.repo.Load(ctx, cmd.CredentialID)
	if err != nil {
		return nil, err
	}

	// Execute domain logic
	if err := credential.Revoke(cmd.RevokedBy, cmd.Reason); err != nil {
		return nil, err
	}

	// Persist
	if err := h.repo.Save(ctx, credential); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.CredentialID,
		Version: credential.Version(),
	}, nil
}

// ============================================================================
// Suspend Credential Handler
// ============================================================================

// SuspendCredentialHandler handles SuspendCredential commands.
type SuspendCredentialHandler struct {
	repo CredentialRepository
}

// NewSuspendCredentialHandler creates a new SuspendCredentialHandler.
func NewSuspendCredentialHandler(repo CredentialRepository) *SuspendCredentialHandler {
	return &SuspendCredentialHandler{repo: repo}
}

// Handle handles the SuspendCredential command.
func (h *SuspendCredentialHandler) Handle(ctx context.Context, cmd *SuspendCredential) (*cqrs.CommandResult, error) {
	// Load credential
	credential, err := h.repo.Load(ctx, cmd.CredentialID)
	if err != nil {
		return nil, err
	}

	// Execute domain logic
	if err := credential.Suspend(cmd.SuspendedBy, cmd.Reason, cmd.Until); err != nil {
		return nil, err
	}

	// Persist
	if err := h.repo.Save(ctx, credential); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.CredentialID,
		Version: credential.Version(),
	}, nil
}

// ============================================================================
// Reinstate Credential Handler
// ============================================================================

// ReinstateCredentialHandler handles ReinstateCredential commands.
type ReinstateCredentialHandler struct {
	repo CredentialRepository
}

// NewReinstateCredentialHandler creates a new ReinstateCredentialHandler.
func NewReinstateCredentialHandler(repo CredentialRepository) *ReinstateCredentialHandler {
	return &ReinstateCredentialHandler{repo: repo}
}

// Handle handles the ReinstateCredential command.
func (h *ReinstateCredentialHandler) Handle(ctx context.Context, cmd *ReinstateCredential) (*cqrs.CommandResult, error) {
	// Load credential
	credential, err := h.repo.Load(ctx, cmd.CredentialID)
	if err != nil {
		return nil, err
	}

	// Execute domain logic
	if err := credential.Reinstate(cmd.ReinstatedBy, cmd.Reason); err != nil {
		return nil, err
	}

	// Persist
	if err := h.repo.Save(ctx, credential); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.CredentialID,
		Version: credential.Version(),
	}, nil
}

// ============================================================================
// Handler Registration
// ============================================================================

// RegisterHandlers registers all credential command handlers with the command bus.
func RegisterHandlers(bus *cqrs.InMemoryCommandBus, repo CredentialRepository) error {
	handlers := map[string]any{
		TypeRequestCredential:   NewRequestCredentialHandler(repo),
		TypeIssueCredential:     NewIssueCredentialHandler(repo),
		TypeRevokeCredential:    NewRevokeCredentialHandler(repo),
		TypeSuspendCredential:   NewSuspendCredentialHandler(repo),
		TypeReinstateCredential: NewReinstateCredentialHandler(repo),
	}

	for cmdType, handler := range handlers {
		if err := bus.Register(cmdType, handler); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.CommandHandler[*RequestCredential]   = (*RequestCredentialHandler)(nil)
	_ cqrs.CommandHandler[*IssueCredential]     = (*IssueCredentialHandler)(nil)
	_ cqrs.CommandHandler[*RevokeCredential]    = (*RevokeCredentialHandler)(nil)
	_ cqrs.CommandHandler[*SuspendCredential]   = (*SuspendCredentialHandler)(nil)
	_ cqrs.CommandHandler[*ReinstateCredential] = (*ReinstateCredentialHandler)(nil)
)
