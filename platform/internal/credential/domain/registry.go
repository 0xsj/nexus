package domain

import (
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// NewRegistry creates a new event registry with all credential events registered.
func NewRegistry() *eventsourcing.EventRegistry {
	registry := eventsourcing.NewEventRegistry()

	registry.Register(EventTypeCredentialRequested, func() eventsourcing.Event {
		return &CredentialRequested{}
	})

	registry.Register(EventTypeCredentialIssued, func() eventsourcing.Event {
		return &CredentialIssued{}
	})

	registry.Register(EventTypeCredentialRevoked, func() eventsourcing.Event {
		return &CredentialRevoked{}
	})

	registry.Register(EventTypeCredentialSuspended, func() eventsourcing.Event {
		return &CredentialSuspended{}
	})

	registry.Register(EventTypeCredentialReinstated, func() eventsourcing.Event {
		return &CredentialReinstated{}
	})

	registry.Register(EventTypeCredentialExpired, func() eventsourcing.Event {
		return &CredentialExpired{}
	})

	return registry
}

// RegisterEvents registers all credential events with the given registry.
func RegisterEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeCredentialRequested, func() eventsourcing.Event {
		return &CredentialRequested{}
	})

	registry.Register(EventTypeCredentialIssued, func() eventsourcing.Event {
		return &CredentialIssued{}
	})

	registry.Register(EventTypeCredentialRevoked, func() eventsourcing.Event {
		return &CredentialRevoked{}
	})

	registry.Register(EventTypeCredentialSuspended, func() eventsourcing.Event {
		return &CredentialSuspended{}
	})

	registry.Register(EventTypeCredentialReinstated, func() eventsourcing.Event {
		return &CredentialReinstated{}
	})

	registry.Register(EventTypeCredentialExpired, func() eventsourcing.Event {
		return &CredentialExpired{}
	})
}

// RegisterWithDefault registers all credential events with the default registry.
func RegisterWithDefault() {
	RegisterEvents(eventsourcing.DefaultRegistry)
}
