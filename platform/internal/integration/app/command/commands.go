package command

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandConnectProvider    = "integration.ConnectProvider"
	CommandDisconnectProvider = "integration.DisconnectProvider"
	CommandRefreshCredentials = "integration.RefreshCredentials"
	CommandSuspendIntegration = "integration.SuspendIntegration"
)

// ============================================================================
// ConnectProvider
// ============================================================================

// ConnectProvider connects a new external provider for a user.
type ConnectProvider struct {
	UserID           string   `json:"user_id" validate:"required"`
	ProviderType     string   `json:"provider_type" validate:"required"`
	ProviderUserID   string   `json:"provider_user_id" validate:"required"`
	ProviderUsername string   `json:"provider_username" validate:"omitempty"`
	Scopes           []string `json:"scopes" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c ConnectProvider) CommandName() string {
	return CommandConnectProvider
}

// Validate implements cqrs.Validatable.
func (c ConnectProvider) Validate() error {
	if c.UserID == "" {
		return cqrs.ErrCommandValidation("ConnectProvider.Validate", "user_id is required")
	}
	if c.ProviderType == "" {
		return cqrs.ErrCommandValidation("ConnectProvider.Validate", "provider_type is required")
	}
	if c.ProviderUserID == "" {
		return cqrs.ErrCommandValidation("ConnectProvider.Validate", "provider_user_id is required")
	}
	return nil
}

// ConnectProviderResult is the result data for ConnectProvider.
type ConnectProviderResult struct {
	IntegrationID string `json:"integration_id"`
	ProviderType  string `json:"provider_type"`
	Status        string `json:"status"`
}

// ============================================================================
// DisconnectProvider
// ============================================================================

// DisconnectProvider disconnects an existing provider integration.
type DisconnectProvider struct {
	IntegrationID types.ID `json:"integration_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c DisconnectProvider) CommandName() string {
	return CommandDisconnectProvider
}

// Validate implements cqrs.Validatable.
func (c DisconnectProvider) Validate() error {
	if c.IntegrationID.IsZero() {
		return cqrs.ErrCommandValidation("DisconnectProvider.Validate", "integration_id is required")
	}
	return nil
}

// DisconnectProviderResult is the result data for DisconnectProvider.
type DisconnectProviderResult struct {
	IntegrationID string `json:"integration_id"`
	Status        string `json:"status"`
}

// ============================================================================
// RefreshCredentials
// ============================================================================

// RefreshCredentials refreshes the credentials for an integration.
type RefreshCredentials struct {
	IntegrationID types.ID `json:"integration_id" validate:"required"`
	Scopes        []string `json:"scopes" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c RefreshCredentials) CommandName() string {
	return CommandRefreshCredentials
}

// Validate implements cqrs.Validatable.
func (c RefreshCredentials) Validate() error {
	if c.IntegrationID.IsZero() {
		return cqrs.ErrCommandValidation("RefreshCredentials.Validate", "integration_id is required")
	}
	return nil
}

// RefreshCredentialsResult is the result data for RefreshCredentials.
type RefreshCredentialsResult struct {
	IntegrationID string `json:"integration_id"`
	Status        string `json:"status"`
}

// ============================================================================
// SuspendIntegration
// ============================================================================

// SuspendIntegration suspends an existing integration.
type SuspendIntegration struct {
	IntegrationID types.ID `json:"integration_id" validate:"required"`
	Reason        string   `json:"reason" validate:"required,max=500"`
}

// CommandName implements cqrs.Command.
func (c SuspendIntegration) CommandName() string {
	return CommandSuspendIntegration
}

// Validate implements cqrs.Validatable.
func (c SuspendIntegration) Validate() error {
	if c.IntegrationID.IsZero() {
		return cqrs.ErrCommandValidation("SuspendIntegration.Validate", "integration_id is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation("SuspendIntegration.Validate", "reason is required")
	}
	if len(c.Reason) > 500 {
		return cqrs.ErrCommandValidation("SuspendIntegration.Validate", "reason must be 500 characters or less")
	}
	return nil
}

// SuspendIntegrationResult is the result data for SuspendIntegration.
type SuspendIntegrationResult struct {
	IntegrationID string `json:"integration_id"`
	Status        string `json:"status"`
}
