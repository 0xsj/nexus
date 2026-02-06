package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeIntegration = "Integration"
)

// Event type constants
const (
	EventTypeProviderConnected    = "Integration.ProviderConnected"
	EventTypeProviderDisconnected = "Integration.ProviderDisconnected"
	EventTypeCredentialsRefreshed = "Integration.CredentialsRefreshed"
	EventTypeDataFetched          = "Integration.DataFetched"
	EventTypeIntegrationSuspended = "Integration.Suspended"
)

// ============================================================================
// Integration Events
// ============================================================================

// ProviderConnectedEvent is emitted when a provider is connected.
type ProviderConnectedEvent struct {
	eventsourcing.BaseEvent

	IntegrationID    string    `json:"integration_id"`
	UserID           string    `json:"user_id"`
	ProviderType     string    `json:"provider_type"`
	ProviderUserID   string    `json:"provider_user_id"`
	ProviderUsername string    `json:"provider_username"`
	Scopes           []string  `json:"scopes"`
	ConnectedAt      time.Time `json:"connected_at"`
}

// EventType returns the event type.
func (e ProviderConnectedEvent) EventType() string {
	return EventTypeProviderConnected
}

// ProviderDisconnectedEvent is emitted when a provider is disconnected.
type ProviderDisconnectedEvent struct {
	eventsourcing.BaseEvent

	IntegrationID  string    `json:"integration_id"`
	DisconnectedAt time.Time `json:"disconnected_at"`
}

// EventType returns the event type.
func (e ProviderDisconnectedEvent) EventType() string {
	return EventTypeProviderDisconnected
}

// CredentialsRefreshedEvent is emitted when integration credentials are refreshed.
type CredentialsRefreshedEvent struct {
	eventsourcing.BaseEvent

	IntegrationID string    `json:"integration_id"`
	Scopes        []string  `json:"scopes"`
	RefreshedAt   time.Time `json:"refreshed_at"`
}

// EventType returns the event type.
func (e CredentialsRefreshedEvent) EventType() string {
	return EventTypeCredentialsRefreshed
}

// DataFetchedEvent is emitted when data is fetched from the provider.
type DataFetchedEvent struct {
	eventsourcing.BaseEvent

	IntegrationID string    `json:"integration_id"`
	FetchedAt     time.Time `json:"fetched_at"`
}

// EventType returns the event type.
func (e DataFetchedEvent) EventType() string {
	return EventTypeDataFetched
}

// IntegrationSuspendedEvent is emitted when an integration is suspended.
type IntegrationSuspendedEvent struct {
	eventsourcing.BaseEvent

	IntegrationID string    `json:"integration_id"`
	Reason        string    `json:"reason"`
	SuspendedAt   time.Time `json:"suspended_at"`
}

// EventType returns the event type.
func (e IntegrationSuspendedEvent) EventType() string {
	return EventTypeIntegrationSuspended
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterIntegrationEvents registers all Integration domain events with the event registry.
func RegisterIntegrationEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeProviderConnected, func() eventsourcing.Event { return &ProviderConnectedEvent{} })
	registry.Register(EventTypeProviderDisconnected, func() eventsourcing.Event { return &ProviderDisconnectedEvent{} })
	registry.Register(EventTypeCredentialsRefreshed, func() eventsourcing.Event { return &CredentialsRefreshedEvent{} })
	registry.Register(EventTypeDataFetched, func() eventsourcing.Event { return &DataFetchedEvent{} })
	registry.Register(EventTypeIntegrationSuspended, func() eventsourcing.Event { return &IntegrationSuspendedEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterIntegrationEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newIntegrationBaseEvent creates a base event for integration aggregate.
func newIntegrationBaseEvent(id IntegrationID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeIntegration, id.String())
}
