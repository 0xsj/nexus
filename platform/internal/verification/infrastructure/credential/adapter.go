package credential

import (
	"context"
	"fmt"
	"time"

	credentialcmd "github.com/0xsj/nexus/platform/internal/credential/application/command"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/id"
)

// ============================================================================
// Credential Module Adapter
// ============================================================================

// ModuleAdapter adapts the credential module's command handlers to the
// CredentialService interface used by the verification module.
type ModuleAdapter struct {
	commandBus cqrs.CommandBus
	idGen      id.Generator
	issuerDID  string
}

// NewModuleAdapter creates a new ModuleAdapter.
func NewModuleAdapter(commandBus cqrs.CommandBus, idGen id.Generator, issuerDID string) *ModuleAdapter {
	return &ModuleAdapter{
		commandBus: commandBus,
		idGen:      idGen,
		issuerDID:  issuerDID,
	}
}

// IssueCredential issues a credential through the credential module.
func (a *ModuleAdapter) IssueCredential(ctx context.Context, params IssueCredentialParams) (*IssuedCredential, error) {
	const op = "ModuleAdapter.IssueCredential"

	// Generate credential ID
	credentialID := a.idGen.Generate().String()

	// Use the issuer DID from params or fall back to default
	issuerDID := params.IssuerDID
	if issuerDID == "" {
		issuerDID = a.issuerDID
	}

	// Calculate expiration time if ExpiresIn is provided
	var expiresAt *time.Time
	if params.ExpiresIn != nil {
		exp := time.Now().Add(time.Duration(*params.ExpiresIn) * time.Second)
		expiresAt = &exp
	}

	// Create issue command using credential module's command type
	cmd := credentialcmd.NewIssueCredential(
		credentialID,
		params.HolderDID,
		issuerDID,
		params.CredentialType,
		params.Claims,
	)

	// Set expiration if provided
	if expiresAt != nil {
		cmd.WithExpiration(*expiresAt)
	}

	// Dispatch command
	result, err := a.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to dispatch issue command: %w", op, err)
	}

	// Build issued credential response
	issuedAt := time.Now().UTC().Format(time.RFC3339)
	var expiresAtStr *string
	if expiresAt != nil {
		s := expiresAt.Format(time.RFC3339)
		expiresAtStr = &s
	}

	return &IssuedCredential{
		ID:        result.ID,
		Type:      params.CredentialType,
		SignedVC:  "", // SignedVC would need to be fetched separately if needed
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAtStr,
	}, nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ CredentialService = (*ModuleAdapter)(nil)
