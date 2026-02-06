package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Integration is the aggregate root for provider integrations.
// It manages the lifecycle of an external provider connection including
// connecting, disconnecting, credential refresh, data fetching, and suspension.
type Integration struct {
	eventsourcing.AggregateRoot

	id               IntegrationID
	userID           string
	providerType     ProviderType
	status           IntegrationStatus
	providerUserID   string
	providerUsername string
	scopes           []string
	lastFetchAt      time.Time
	fetchCount       int
	metadata         map[string]string
	connectedAt      time.Time
	disconnectedAt   time.Time
	suspendedAt      time.Time
	suspensionReason string
	createdAt        time.Time
	updatedAt        time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// ConnectProvider creates a new Integration aggregate for a provider connection.
func ConnectProvider(
	id IntegrationID,
	userID string,
	providerType ProviderType,
	providerUserID string,
	providerUsername string,
	scopes []string,
) (*Integration, error) {
	if id.IsZero() {
		return nil, IntegrationInvalid("Integration.ConnectProvider", "integration ID is required")
	}

	if userID == "" {
		return nil, IntegrationInvalid("Integration.ConnectProvider", "user ID is required")
	}

	if !providerType.IsValid() {
		return nil, ProviderNotSupported("Integration.ConnectProvider", string(providerType))
	}

	if providerUserID == "" {
		return nil, IntegrationInvalid("Integration.ConnectProvider", "provider user ID is required")
	}

	i := &Integration{}
	i.InitAggregate(AggregateTypeIntegration, id.String())

	now := time.Now().UTC()

	i.Raise(i, &ProviderConnectedEvent{
		BaseEvent:        newIntegrationBaseEvent(id),
		IntegrationID:    id.String(),
		UserID:           userID,
		ProviderType:     providerType.String(),
		ProviderUserID:   providerUserID,
		ProviderUsername: providerUsername,
		Scopes:           scopes,
		ConnectedAt:      now,
	})

	return i, nil
}

// NewIntegrationFromEvents reconstructs an Integration from events (for hydration).
func NewIntegrationFromEvents(id string) *Integration {
	i := &Integration{}
	i.InitAggregate(AggregateTypeIntegration, id)
	return i
}

// IntegrationFactory creates a factory for Integration aggregates.
func IntegrationFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewIntegrationFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the integration's ID.
func (i *Integration) ID() IntegrationID {
	return i.id
}

// UserID returns the user ID that owns this integration.
func (i *Integration) UserID() string {
	return i.userID
}

// ProviderType returns the provider type.
func (i *Integration) ProviderType() ProviderType {
	return i.providerType
}

// Status returns the integration's status.
func (i *Integration) Status() IntegrationStatus {
	return i.status
}

// ProviderUserID returns the user's ID on the provider platform.
func (i *Integration) ProviderUserID() string {
	return i.providerUserID
}

// ProviderUsername returns the user's username on the provider platform.
func (i *Integration) ProviderUsername() string {
	return i.providerUsername
}

// Scopes returns the granted scopes.
func (i *Integration) Scopes() []string {
	return i.scopes
}

// LastFetchAt returns when data was last fetched.
func (i *Integration) LastFetchAt() time.Time {
	return i.lastFetchAt
}

// FetchCount returns the total number of data fetches.
func (i *Integration) FetchCount() int {
	return i.fetchCount
}

// Metadata returns the integration metadata.
func (i *Integration) Metadata() map[string]string {
	return i.metadata
}

// ConnectedAt returns when the integration was connected.
func (i *Integration) ConnectedAt() time.Time {
	return i.connectedAt
}

// DisconnectedAt returns when the integration was disconnected.
func (i *Integration) DisconnectedAt() time.Time {
	return i.disconnectedAt
}

// SuspendedAt returns when the integration was suspended.
func (i *Integration) SuspendedAt() time.Time {
	return i.suspendedAt
}

// SuspensionReason returns the reason for suspension.
func (i *Integration) SuspensionReason() string {
	return i.suspensionReason
}

// CreatedAt returns when the integration was created.
func (i *Integration) CreatedAt() time.Time {
	return i.createdAt
}

// UpdatedAt returns when the integration was last updated.
func (i *Integration) UpdatedAt() time.Time {
	return i.updatedAt
}

// ============================================================================
// Command Methods
// ============================================================================

// Disconnect disconnects the integration from the provider.
func (i *Integration) Disconnect() error {
	if !i.status.CanTransitionTo(IntegrationStatusDisconnected) {
		return IntegrationDisconnected("Integration.Disconnect", i.id.String()).
			WithMessage("cannot disconnect integration in current status: " + i.status.String())
	}

	i.Raise(i, &ProviderDisconnectedEvent{
		BaseEvent:      newIntegrationBaseEvent(i.id),
		IntegrationID:  i.id.String(),
		DisconnectedAt: time.Now().UTC(),
	})

	return nil
}

// RefreshCredentials refreshes the integration's credentials with updated scopes.
func (i *Integration) RefreshCredentials(scopes []string) error {
	if !i.status.IsConnected() {
		return IntegrationDisconnected("Integration.RefreshCredentials", i.id.String()).
			WithMessage("cannot refresh credentials for integration in status: " + i.status.String())
	}

	i.Raise(i, &CredentialsRefreshedEvent{
		BaseEvent:     newIntegrationBaseEvent(i.id),
		IntegrationID: i.id.String(),
		Scopes:        scopes,
		RefreshedAt:   time.Now().UTC(),
	})

	return nil
}

// RecordFetch records a data fetch from the provider.
func (i *Integration) RecordFetch() error {
	if !i.status.IsConnected() {
		return IntegrationDisconnected("Integration.RecordFetch", i.id.String()).
			WithMessage("cannot fetch data for integration in status: " + i.status.String())
	}

	i.Raise(i, &DataFetchedEvent{
		BaseEvent:     newIntegrationBaseEvent(i.id),
		IntegrationID: i.id.String(),
		FetchedAt:     time.Now().UTC(),
	})

	return nil
}

// Suspend suspends the integration.
func (i *Integration) Suspend(reason string) error {
	if !i.status.CanTransitionTo(IntegrationStatusSuspended) {
		return IntegrationSuspendedErr("Integration.Suspend", i.id.String()).
			WithMessage("cannot suspend integration in current status: " + i.status.String())
	}

	i.Raise(i, &IntegrationSuspendedEvent{
		BaseEvent:     newIntegrationBaseEvent(i.id),
		IntegrationID: i.id.String(),
		Reason:        reason,
		SuspendedAt:   time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (i *Integration) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *ProviderConnectedEvent:
		i.onProviderConnected(e)
	case *ProviderDisconnectedEvent:
		i.onProviderDisconnected(e)
	case *CredentialsRefreshedEvent:
		i.onCredentialsRefreshed(e)
	case *DataFetchedEvent:
		i.onDataFetched(e)
	case *IntegrationSuspendedEvent:
		i.onIntegrationSuspended(e)
	}
}

func (i *Integration) onProviderConnected(e *ProviderConnectedEvent) {
	var err error
	i.id, err = ParseIntegrationID(e.IntegrationID)
	if err != nil {
		panic("corrupt event store: ProviderConnected has invalid IntegrationID: " + e.IntegrationID)
	}
	i.providerType, err = ParseProviderType(e.ProviderType)
	if err != nil {
		panic("corrupt event store: ProviderConnected has invalid ProviderType: " + e.ProviderType)
	}
	i.userID = e.UserID
	i.providerUserID = e.ProviderUserID
	i.providerUsername = e.ProviderUsername
	i.scopes = e.Scopes
	i.status = IntegrationStatusConnected
	i.connectedAt = e.ConnectedAt
	i.createdAt = e.ConnectedAt
	i.updatedAt = e.ConnectedAt
	i.metadata = make(map[string]string)
}

func (i *Integration) onProviderDisconnected(e *ProviderDisconnectedEvent) {
	i.status = IntegrationStatusDisconnected
	i.disconnectedAt = e.DisconnectedAt
	i.updatedAt = e.DisconnectedAt
}

func (i *Integration) onCredentialsRefreshed(e *CredentialsRefreshedEvent) {
	i.scopes = e.Scopes
	i.updatedAt = e.RefreshedAt
}

func (i *Integration) onDataFetched(e *DataFetchedEvent) {
	i.lastFetchAt = e.FetchedAt
	i.fetchCount++
	i.updatedAt = e.FetchedAt
}

func (i *Integration) onIntegrationSuspended(e *IntegrationSuspendedEvent) {
	i.status = IntegrationStatusSuspended
	i.suspendedAt = e.SuspendedAt
	i.suspensionReason = e.Reason
	i.updatedAt = e.SuspendedAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (i *Integration) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &i.AggregateRoot
}
