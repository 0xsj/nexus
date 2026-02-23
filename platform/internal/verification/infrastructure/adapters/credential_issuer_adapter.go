package adapters

import (
	"context"
	"fmt"

	credentialcmd "github.com/0xsj/nexus/platform/internal/credential/app/command"
	verifdomain "github.com/0xsj/nexus/platform/internal/verification/domain"
)

// CredentialIssuerAdapter implements verification's domain.CredentialIssuer
// by delegating to the Credential context's command handlers.
type CredentialIssuerAdapter struct {
	credentialHandlers *credentialcmd.Handlers
}

// Compile-time check.
var _ verifdomain.CredentialIssuer = (*CredentialIssuerAdapter)(nil)

// NewCredentialIssuerAdapter creates a new CredentialIssuerAdapter.
func NewCredentialIssuerAdapter(handlers *credentialcmd.Handlers) *CredentialIssuerAdapter {
	return &CredentialIssuerAdapter{credentialHandlers: handlers}
}

// IssueCredential creates a new credential from the verified provider data.
func (a *CredentialIssuerAdapter) IssueCredential(
	ctx context.Context,
	userID string,
	provider verifdomain.ProviderType,
	data map[string]any,
) (string, error) {
	result, err := a.credentialHandlers.HandleIssueCredential(ctx, credentialcmd.IssueCredential{
		CredentialType: string(provider),
		IssuerDID:      "did:nexus:platform",
		SubjectDID:     userID,
		Claims:         data,
	})
	if err != nil {
		return "", fmt.Errorf("issuing credential via adapter: %w", err)
	}

	issueResult, ok := result.Data.(credentialcmd.IssueCredentialResult)
	if !ok {
		return result.ID, nil
	}
	return issueResult.CredentialID, nil
}
