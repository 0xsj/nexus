package event

import (
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// NewRegistry creates a new event registry pre-populated with credential events.
func NewRegistry() *eventsourcing.EventRegistry {
	registry := eventsourcing.NewEventRegistry()
	RegisterEvents(registry)
	return registry
}

// RegisterEvents registers all credential events with the given registry.
func RegisterEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(TypeCredentialRequested, func() eventsourcing.Event {
		return &CredentialRequested{}
	})

	registry.Register(TypeCredentialIssued, func() eventsourcing.Event {
		return &CredentialIssued{}
	})

	registry.Register(TypeCredentialRevoked, func() eventsourcing.Event {
		return &CredentialRevoked{}
	})

	registry.Register(TypeCredentialSuspended, func() eventsourcing.Event {
		return &CredentialSuspended{}
	})

	registry.Register(TypeCredentialReinstated, func() eventsourcing.Event {
		return &CredentialReinstated{}
	})

	registry.Register(TypeCredentialExpired, func() eventsourcing.Event {
		return &CredentialExpired{}
	})
}

// RegisterWithDefault registers all credential events with the default registry.
func RegisterWithDefault() {
	RegisterEvents(eventsourcing.DefaultRegistry)
}
