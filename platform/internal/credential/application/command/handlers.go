package command

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Request Credential Handler
// ============================================================================

// RequestCredentialHandler handles RequestCredential commands.
type RequestCredentialHandler struct {
	repo domain.CredentialRepository
}

// NewRequestCredentialHandler creates a new RequestCredentialHandler.
func NewRequestCredentialHandler(repo domain.CredentialRepository) *RequestCredentialHandler {
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
	credential := domain.NewCredential(cmd.CredentialID)

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
	repo           domain.CredentialRepository
	signingService domain.SigningService
}

// NewIssueCredentialHandler creates a new IssueCredentialHandler.
func NewIssueCredentialHandler(repo domain.CredentialRepository, signingService domain.SigningService) *IssueCredentialHandler {
	return &IssueCredentialHandler{
		repo:           repo,
		signingService: signingService,
	}
}

// Handle handles the IssueCredential command.
func (h *IssueCredentialHandler) Handle(ctx context.Context, cmd *IssueCredential) (*cqrs.CommandResult, error) {
	// Try to load existing credential (might be from a request)
	credential, err := h.repo.Load(ctx, cmd.CredentialID)
	if err != nil {
		// If not found, create new aggregate for direct issuance
		if eventsourcing.IsAggregateNotFound(err) {
			credential = domain.NewCredential(cmd.CredentialID)
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

	// Sign the credential if signing service is available
	var signedVC string
	if h.signingService != nil {
		issuedAt := time.Now().UTC()
		params := domain.SigningParams{
			CredentialID:   cmd.CredentialID,
			CredentialType: cmd.CredentialType,
			IssuerDID:      cmd.IssuerDID,
			HolderDID:      cmd.HolderDID,
			Claims:         cmd.Claims,
			IssuedAt:       issuedAt,
			ExpiresAt:      cmd.ExpiresAt,
			SchemaID:       cmd.SchemaID,
		}

		signedVC, err = h.signingService.SignCredential(ctx, params)
		if err != nil {
			return nil, err
		}
	}

	// Execute domain logic
	if err := credential.Issue(cmd.IssuerDID, claims, cmd.ExpiresAt, signedVC); err != nil {
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
	repo domain.CredentialRepository
}

// NewRevokeCredentialHandler creates a new RevokeCredentialHandler.
func NewRevokeCredentialHandler(repo domain.CredentialRepository) *RevokeCredentialHandler {
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
	repo domain.CredentialRepository
}

// NewSuspendCredentialHandler creates a new SuspendCredentialHandler.
func NewSuspendCredentialHandler(repo domain.CredentialRepository) *SuspendCredentialHandler {
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
	repo domain.CredentialRepository
}

// NewReinstateCredentialHandler creates a new ReinstateCredentialHandler.
func NewReinstateCredentialHandler(repo domain.CredentialRepository) *ReinstateCredentialHandler {
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
func RegisterHandlers(bus *cqrs.InMemoryCommandBus, repo domain.CredentialRepository, signingService domain.SigningService) error {
	handlers := map[string]any{
		TypeRequestCredential:   NewRequestCredentialHandler(repo),
		TypeIssueCredential:     NewIssueCredentialHandler(repo, signingService),
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
