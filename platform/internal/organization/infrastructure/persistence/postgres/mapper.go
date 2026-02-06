// Package postgres provides PostgreSQL implementations of Organization domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/0xsj/nexus/platform/internal/organization/domain"
	"github.com/0xsj/nexus/platform/internal/organization/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Organization Mappers
// ============================================================================

// OrganizationToUpsertParams maps a domain Organization to SQLC UpsertOrganizationParams.
func OrganizationToUpsertParams(org *domain.Organization) (generated.UpsertOrganizationParams, error) {
	var description *string
	if org.Description() != "" {
		d := org.Description()
		description = &d
	}

	var did *string
	if org.DID() != "" {
		d := org.DID()
		did = &d
	}

	now := time.Now().UTC()

	return generated.UpsertOrganizationParams{
		ID:                 uuidFromOrganizationID(org.ID()),
		Name:               org.Name(),
		Slug:               org.Slug(),
		OrgType:            org.OrgType().String(),
		Description:        description,
		VerificationStatus: org.VerificationStatus().String(),
		Did:                did,
		OwnerMemberID:      org.OwnerMemberID().String(),
		MemberCount:        int32(org.MemberCount()),
		Deleted:            org.IsDeleted(),
		Version:            int32(org.Version()),
		CreatedAt:          org.CreatedAt(),
		UpdatedAt:          now,
	}, nil
}

// ============================================================================
// Event Mappers
// ============================================================================

// OrganizationEventToInsertParams maps an eventsourcing.Event to SQLC InsertOrganizationEventParams.
func OrganizationEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertOrganizationEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertOrganizationEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertOrganizationEventParams{}, err
	}

	return generated.InsertOrganizationEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeOrganization,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// OrganizationEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func OrganizationEventRowToEvent(row generated.OrganizationEvent) (eventsourcing.Event, error) {
	event, err := unmarshalOrganizationEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// OrganizationEventRowsToEvents maps multiple SQLC rows to events.
func OrganizationEventRowsToEvents(rows []generated.OrganizationEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := OrganizationEventRowToEvent(row)
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

// unmarshalOrganizationEvent deserializes an organization event from JSON based on event type.
func unmarshalOrganizationEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeOrganizationCreated:
		var e domain.OrganizationCreatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeOrganizationUpdated:
		var e domain.OrganizationUpdatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeOrganizationVerified:
		var e domain.OrganizationVerifiedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeOrganizationDeleted:
		var e domain.OrganizationDeletedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeMemberAdded:
		var e domain.MemberAddedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeMemberRemoved:
		var e domain.MemberRemovedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeMemberRoleChanged:
		var e domain.MemberRoleChangedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeOwnershipTransferred:
		var e domain.OwnershipTransferredEvent
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

// OrganizationProjection holds read-model data from the organizations table.
type OrganizationProjection struct {
	ID                 string
	Name               string
	Slug               string
	OrgType            string
	Description        string
	VerificationStatus string
	DID                string
	OwnerMemberID      string
	MemberCount        int
	Deleted            bool
	Version            int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// OrganizationRowToProjection maps a generated.Organization row to an OrganizationProjection.
func OrganizationRowToProjection(row generated.Organization) OrganizationProjection {
	proj := OrganizationProjection{
		ID:                 row.ID.String(),
		Name:               row.Name,
		Slug:               row.Slug,
		OrgType:            row.OrgType,
		VerificationStatus: row.VerificationStatus,
		OwnerMemberID:      row.OwnerMemberID,
		MemberCount:        int(row.MemberCount),
		Deleted:            row.Deleted,
		Version:            int(row.Version),
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}

	// Nullable fields
	if row.Description != nil {
		proj.Description = *row.Description
	}
	if row.Did != nil {
		proj.DID = *row.Did
	}

	return proj
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromOrganizationID converts a domain.OrganizationID to uuid.UUID.
func uuidFromOrganizationID(id domain.OrganizationID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
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
