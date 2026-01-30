package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Actor Error Codes
// ============================================================================

const (
	CodeInvalidActorID   pkgerrors.Code = "LEDGER_INVALID_ACTOR_ID"
	CodeInvalidActorType pkgerrors.Code = "LEDGER_INVALID_ACTOR_TYPE"
)

// ============================================================================
// Actor Sentinel Errors
// ============================================================================

var (
	ErrInvalidActorID   = errors.New("invalid actor ID")
	ErrInvalidActorType = errors.New("invalid actor type")
)

// ============================================================================
// Actor Error Constructors
// ============================================================================

// InvalidActorID creates an invalid actor ID error.
func InvalidActorID(operation string, value string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid actor ID: "+reason).
		WithCode(CodeInvalidActorID).
		WithMeta("value", value)
}

// InvalidActorType creates an invalid actor type error.
func InvalidActorType(operation string, value string) *pkgerrors.Error {
	return pkgerrors.Validationf(operation, "invalid actor type: %s", value).
		WithCode(CodeInvalidActorType).
		WithMeta("value", value)
}

// ============================================================================
// ActorType Value Object
// ============================================================================

// ActorType represents the type of entity that caused an event.
type ActorType int

const (
	// ActorTypeUnknown indicates an unknown actor type.
	ActorTypeUnknown ActorType = iota

	// ActorTypeUser indicates the actor is a platform user.
	ActorTypeUser

	// ActorTypeSystem indicates the actor is the system itself (automated actions).
	ActorTypeSystem

	// ActorTypeExternalVerifier indicates the actor is an external party verifying credentials.
	ActorTypeExternalVerifier

	// ActorTypeIssuer indicates the actor is a credential issuer (organization).
	ActorTypeIssuer

	// ActorTypeAdmin indicates the actor is a platform administrator.
	ActorTypeAdmin
)

// String returns the string representation of the ActorType.
func (t ActorType) String() string {
	switch t {
	case ActorTypeUser:
		return "user"
	case ActorTypeSystem:
		return "system"
	case ActorTypeExternalVerifier:
		return "external_verifier"
	case ActorTypeIssuer:
		return "issuer"
	case ActorTypeAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

// ParseActorType parses a string into an ActorType.
func ParseActorType(s string) (ActorType, error) {
	switch s {
	case "user":
		return ActorTypeUser, nil
	case "system":
		return ActorTypeSystem, nil
	case "external_verifier":
		return ActorTypeExternalVerifier, nil
	case "issuer":
		return ActorTypeIssuer, nil
	case "admin":
		return ActorTypeAdmin, nil
	case "unknown":
		return ActorTypeUnknown, nil
	default:
		return ActorTypeUnknown, InvalidActorType("domain.ParseActorType", s)
	}
}

// IsValid returns true if the ActorType is a known, valid type.
func (t ActorType) IsValid() bool {
	switch t {
	case ActorTypeUser, ActorTypeSystem, ActorTypeExternalVerifier, ActorTypeIssuer, ActorTypeAdmin:
		return true
	default:
		return false
	}
}

// ============================================================================
// ActorID Value Object
// ============================================================================

// ActorID identifies the specific actor that caused an event.
// The format depends on ActorType:
//   - User: user UUID
//   - System: service name (e.g., "scheduler", "verification-worker")
//   - ExternalVerifier: verifier identifier or request ID
//   - Issuer: organization UUID
//   - Admin: admin user UUID
type ActorID struct {
	value string
}

// NewActorID creates a new ActorID.
func NewActorID(value string) (ActorID, error) {
	if value == "" {
		return ActorID{}, InvalidActorID("domain.NewActorID", value, "cannot be empty")
	}
	return ActorID{value: value}, nil
}

// MustNewActorID creates a new ActorID and panics if invalid.
// Only use for constants or tests.
func MustNewActorID(value string) ActorID {
	id, err := NewActorID(value)
	if err != nil {
		panic(err)
	}
	return id
}

// SystemActorID returns an ActorID for system actions.
func SystemActorID(serviceName string) ActorID {
	if serviceName == "" {
		serviceName = "system"
	}
	return ActorID{value: serviceName}
}

// String returns the string representation.
func (id ActorID) String() string {
	return id.value
}

// IsZero returns true if the ActorID is the zero value.
func (id ActorID) IsZero() bool {
	return id.value == ""
}

// IsValid returns true if the ActorID is valid (non-empty).
func (id ActorID) IsValid() bool {
	return id.value != ""
}

// Equals checks if two ActorIDs are equal.
func (id ActorID) Equals(other ActorID) bool {
	return id.value == other.value
}

// ============================================================================
// Actor Value Object (Composite)
// ============================================================================

// Actor represents who caused an event, combining type and identifier.
type Actor struct {
	actorType ActorType
	actorID   ActorID
}

// NewActor creates a new Actor.
func NewActor(actorType ActorType, actorID ActorID) (Actor, error) {
	if !actorType.IsValid() {
		return Actor{}, InvalidActorType("domain.NewActor", actorType.String())
	}
	if actorID.IsZero() {
		return Actor{}, InvalidActorID("domain.NewActor", "", "actor ID cannot be empty")
	}
	return Actor{
		actorType: actorType,
		actorID:   actorID,
	}, nil
}

// MustNewActor creates a new Actor and panics if invalid.
// Only use for constants or tests.
func MustNewActor(actorType ActorType, actorID ActorID) Actor {
	actor, err := NewActor(actorType, actorID)
	if err != nil {
		panic(err)
	}
	return actor
}

// SystemActor creates an Actor for system-initiated actions.
func SystemActor(serviceName string) Actor {
	return Actor{
		actorType: ActorTypeSystem,
		actorID:   SystemActorID(serviceName),
	}
}

// Type returns the actor type.
func (a Actor) Type() ActorType {
	return a.actorType
}

// ID returns the actor ID.
func (a Actor) ID() ActorID {
	return a.actorID
}

// IsZero returns true if the Actor is the zero value.
func (a Actor) IsZero() bool {
	return !a.actorType.IsValid() && a.actorID.IsZero()
}

// Equals checks if two Actors are equal.
func (a Actor) Equals(other Actor) bool {
	return a.actorType == other.actorType && a.actorID.Equals(other.actorID)
}
