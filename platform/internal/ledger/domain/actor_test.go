package domain

import (
	"testing"
)

// ============================================================================
// ActorType Tests
// ============================================================================

func TestActorType_String(t *testing.T) {
	tests := []struct {
		actorType ActorType
		want      string
	}{
		{ActorTypeUnknown, "unknown"},
		{ActorTypeUser, "user"},
		{ActorTypeSystem, "system"},
		{ActorTypeExternalVerifier, "external_verifier"},
		{ActorTypeIssuer, "issuer"},
		{ActorTypeAdmin, "admin"},
		{ActorType(999), "unknown"}, // invalid value
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.actorType.String()
			if got != tt.want {
				t.Errorf("ActorType(%d).String() = %v, want %v", tt.actorType, got, tt.want)
			}
		})
	}
}

func TestParseActorType_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  ActorType
	}{
		{"user", ActorTypeUser},
		{"system", ActorTypeSystem},
		{"external_verifier", ActorTypeExternalVerifier},
		{"issuer", ActorTypeIssuer},
		{"admin", ActorTypeAdmin},
		{"unknown", ActorTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseActorType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ParseActorType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseActorType_Invalid(t *testing.T) {
	invalidTypes := []string{
		"invalid",
		"USER",
		"System",
		"",
		"external-verifier",
		"verifier",
	}

	for _, input := range invalidTypes {
		t.Run(input, func(t *testing.T) {
			_, err := ParseActorType(input)
			if err == nil {
				t.Errorf("expected error for ParseActorType(%q)", input)
			}
		})
	}
}

func TestActorType_IsValid(t *testing.T) {
	tests := []struct {
		actorType ActorType
		want      bool
	}{
		{ActorTypeUnknown, false},
		{ActorTypeUser, true},
		{ActorTypeSystem, true},
		{ActorTypeExternalVerifier, true},
		{ActorTypeIssuer, true},
		{ActorTypeAdmin, true},
		{ActorType(999), false},
	}

	for _, tt := range tests {
		t.Run(tt.actorType.String(), func(t *testing.T) {
			got := tt.actorType.IsValid()
			if got != tt.want {
				t.Errorf("ActorType(%d).IsValid() = %v, want %v", tt.actorType, got, tt.want)
			}
		})
	}
}

// ============================================================================
// ActorID Tests
// ============================================================================

func TestNewActorID_Valid(t *testing.T) {
	tests := []string{
		"user-123",
		"system",
		"abc",
		"user@example.com",
		"123e4567-e89b-12d3-a456-426614174000",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			id, err := NewActorID(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id.String() != input {
				t.Errorf("ActorID.String() = %v, want %v", id.String(), input)
			}
			if id.IsZero() {
				t.Error("expected non-zero ActorID")
			}
			if !id.IsValid() {
				t.Error("expected valid ActorID")
			}
		})
	}
}

func TestNewActorID_Empty(t *testing.T) {
	_, err := NewActorID("")

	if err == nil {
		t.Error("expected error for empty ActorID")
	}
}

func TestMustNewActorID_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	id := MustNewActorID("user-123")
	if id.String() != "user-123" {
		t.Errorf("unexpected ActorID: %v", id.String())
	}
}

func TestMustNewActorID_Empty_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty ActorID")
		}
	}()

	MustNewActorID("")
}

func TestSystemActorID(t *testing.T) {
	tests := []struct {
		serviceName string
		want        string
	}{
		{"scheduler", "scheduler"},
		{"verification-worker", "verification-worker"},
		{"", "system"},
	}

	for _, tt := range tests {
		t.Run(tt.serviceName, func(t *testing.T) {
			id := SystemActorID(tt.serviceName)
			if id.String() != tt.want {
				t.Errorf("SystemActorID(%q) = %v, want %v", tt.serviceName, id.String(), tt.want)
			}
		})
	}
}

func TestActorID_IsZero(t *testing.T) {
	var zeroID ActorID

	if !zeroID.IsZero() {
		t.Error("expected zero value to be zero")
	}
	if zeroID.IsValid() {
		t.Error("expected zero value to be invalid")
	}
}

func TestActorID_Equals(t *testing.T) {
	id1, _ := NewActorID("user-123")
	id2, _ := NewActorID("user-123")
	id3, _ := NewActorID("user-456")

	if !id1.Equals(id2) {
		t.Error("expected equal ActorIDs to be equal")
	}
	if id1.Equals(id3) {
		t.Error("expected different ActorIDs to not be equal")
	}
}

// ============================================================================
// Actor Tests
// ============================================================================

func TestNewActor_Valid(t *testing.T) {
	actorID, _ := NewActorID("user-123")
	actor, err := NewActor(ActorTypeUser, actorID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor.Type() != ActorTypeUser {
		t.Errorf("Actor.Type() = %v, want %v", actor.Type(), ActorTypeUser)
	}
	if !actor.ID().Equals(actorID) {
		t.Errorf("Actor.ID() = %v, want %v", actor.ID(), actorID)
	}
	if actor.IsZero() {
		t.Error("expected non-zero Actor")
	}
}

func TestNewActor_InvalidType(t *testing.T) {
	actorID, _ := NewActorID("user-123")
	_, err := NewActor(ActorTypeUnknown, actorID)

	if err == nil {
		t.Error("expected error for invalid ActorType")
	}
}

func TestNewActor_InvalidID(t *testing.T) {
	var zeroID ActorID
	_, err := NewActor(ActorTypeUser, zeroID)

	if err == nil {
		t.Error("expected error for zero ActorID")
	}
}

func TestMustNewActor_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	actorID := MustNewActorID("user-123")
	actor := MustNewActor(ActorTypeUser, actorID)

	if actor.Type() != ActorTypeUser {
		t.Errorf("unexpected ActorType: %v", actor.Type())
	}
}

func TestMustNewActor_InvalidType_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid ActorType")
		}
	}()

	actorID := MustNewActorID("user-123")
	MustNewActor(ActorTypeUnknown, actorID)
}

func TestSystemActor(t *testing.T) {
	tests := []struct {
		serviceName string
		wantID      string
	}{
		{"scheduler", "scheduler"},
		{"worker", "worker"},
		{"", "system"},
	}

	for _, tt := range tests {
		t.Run(tt.serviceName, func(t *testing.T) {
			actor := SystemActor(tt.serviceName)

			if actor.Type() != ActorTypeSystem {
				t.Errorf("SystemActor.Type() = %v, want %v", actor.Type(), ActorTypeSystem)
			}
			if actor.ID().String() != tt.wantID {
				t.Errorf("SystemActor.ID() = %v, want %v", actor.ID().String(), tt.wantID)
			}
		})
	}
}

func TestActor_IsZero(t *testing.T) {
	var zeroActor Actor

	if !zeroActor.IsZero() {
		t.Error("expected zero value to be zero")
	}
}

func TestActor_Equals(t *testing.T) {
	actorID1, _ := NewActorID("user-123")
	actorID2, _ := NewActorID("user-456")

	actor1, _ := NewActor(ActorTypeUser, actorID1)
	actor2, _ := NewActor(ActorTypeUser, actorID1)
	actor3, _ := NewActor(ActorTypeUser, actorID2)
	actor4, _ := NewActor(ActorTypeAdmin, actorID1)

	if !actor1.Equals(actor2) {
		t.Error("expected equal Actors to be equal")
	}
	if actor1.Equals(actor3) {
		t.Error("expected Actors with different IDs to not be equal")
	}
	if actor1.Equals(actor4) {
		t.Error("expected Actors with different types to not be equal")
	}
}
