package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetIntegration               = "integration.GetIntegration"
	QueryListIntegrationsByUser       = "integration.ListIntegrationsByUser"
	QueryGetIntegrationByUserProvider = "integration.GetIntegrationByUserAndProvider"
)

// ============================================================================
// GetIntegration
// ============================================================================

// GetIntegration retrieves a single integration by ID.
type GetIntegration struct {
	IntegrationID types.ID `json:"integration_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetIntegration) QueryName() string {
	return QueryGetIntegration
}

// Validate implements cqrs.Validatable.
func (q GetIntegration) Validate() error {
	if q.IntegrationID.IsZero() {
		return cqrs.ErrQueryValidation("GetIntegration.Validate", "integration_id is required")
	}
	return nil
}

// ============================================================================
// ListIntegrationsByUser
// ============================================================================

// ListIntegrationsByUser lists integrations for a user.
type ListIntegrationsByUser struct {
	UserID string `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q ListIntegrationsByUser) QueryName() string {
	return QueryListIntegrationsByUser
}

// Validate implements cqrs.Validatable.
func (q ListIntegrationsByUser) Validate() error {
	if q.UserID == "" {
		return cqrs.ErrQueryValidation("ListIntegrationsByUser.Validate", "user_id is required")
	}
	return nil
}

// ============================================================================
// GetIntegrationByUserAndProvider
// ============================================================================

// GetIntegrationByUserAndProvider retrieves an integration for a user and provider type.
type GetIntegrationByUserAndProvider struct {
	UserID       string `json:"user_id" validate:"required"`
	ProviderType string `json:"provider_type" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetIntegrationByUserAndProvider) QueryName() string {
	return QueryGetIntegrationByUserProvider
}

// Validate implements cqrs.Validatable.
func (q GetIntegrationByUserAndProvider) Validate() error {
	if q.UserID == "" {
		return cqrs.ErrQueryValidation("GetIntegrationByUserAndProvider.Validate", "user_id is required")
	}
	if q.ProviderType == "" {
		return cqrs.ErrQueryValidation("GetIntegrationByUserAndProvider.Validate", "provider_type is required")
	}
	return nil
}
