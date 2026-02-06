// Package postgres provides PostgreSQL implementations of Profile domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/0xsj/nexus/platform/internal/profile/domain"
	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Profile Mappers
// ============================================================================

// ProfileToUpsertParams maps a domain Profile to SQLC UpsertProfileParams.
func ProfileToUpsertParams(p *domain.Profile) generated.UpsertProfileParams {
	now := time.Now().UTC()

	return generated.UpsertProfileParams{
		ID:          uuidFromProfileID(p.ID()),
		UserID:      p.UserID(),
		DisplayName: p.DisplayName(),
		Headline:    p.Headline(),
		Bio:         p.Bio(),
		VanitySlug:  p.VanitySlug(),
		BadgeCount:  int32(len(p.Badges())),
		Version:     int32(p.Version()),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   now,
	}
}

// ============================================================================
// Event Mappers
// ============================================================================

// ProfileEventToInsertParams maps an eventsourcing.Event to SQLC InsertProfileEventParams.
func ProfileEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertProfileEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertProfileEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertProfileEventParams{}, err
	}

	return generated.InsertProfileEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeProfile,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// ProfileEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func ProfileEventRowToEvent(row generated.ProfileEvent) (eventsourcing.Event, error) {
	event, err := unmarshalProfileEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// ProfileEventRowsToEvents maps multiple SQLC rows to events.
func ProfileEventRowsToEvents(rows []generated.ProfileEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := ProfileEventRowToEvent(row)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// ============================================================================
// Event Unmarshaling
// ============================================================================

// unmarshalProfileEvent deserializes a profile event from JSON based on event type.
func unmarshalProfileEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeProfileCreated:
		var e domain.ProfileCreatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeProfileUpdated:
		var e domain.ProfileUpdatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeBadgeAdded:
		var e domain.BadgeAddedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeBadgeRemoved:
		var e domain.BadgeRemovedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeBadgeVisibilityChanged:
		var e domain.BadgeVisibilityChangedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeVanityURLClaimed:
		var e domain.VanityURLClaimedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	default:
		return nil, &UnknownEventTypeError{EventType: eventType}
	}

	return event, nil
}

// ============================================================================
// Projection Mapping
// ============================================================================

// ProfileProjection holds read-model data from the profiles table.
type ProfileProjection struct {
	ID          string
	UserID      string
	DisplayName string
	Headline    string
	Bio         string
	VanitySlug  string
	BadgeCount  int
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProfileRowToProjection maps a generated.Profile row to a ProfileProjection.
func ProfileRowToProjection(row generated.Profile) ProfileProjection {
	return ProfileProjection{
		ID:          row.ID.String(),
		UserID:      row.UserID,
		DisplayName: row.DisplayName,
		Headline:    row.Headline,
		Bio:         row.Bio,
		VanitySlug:  row.VanitySlug,
		BadgeCount:  int(row.BadgeCount),
		Version:     int(row.Version),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromProfileID converts a domain.ProfileID to uuid.UUID.
func uuidFromProfileID(id domain.ProfileID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// profileIDFromUUID converts a uuid.UUID to domain.ProfileID.
func profileIDFromUUID(u uuid.UUID) domain.ProfileID {
	id, _ := domain.ParseProfileID(u.String())
	return id
}

// ============================================================================
// Errors
// ============================================================================

// UnknownEventTypeError is returned when an unknown event type is encountered.
type UnknownEventTypeError struct {
	EventType string
}

func (e *UnknownEventTypeError) Error() string {
	return "unknown event type: " + e.EventType
}
