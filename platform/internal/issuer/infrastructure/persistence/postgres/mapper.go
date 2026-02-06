// Package postgres provides PostgreSQL implementations of Issuer domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/0xsj/nexus/platform/internal/issuer/domain"
	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Issuer Mappers
// ============================================================================

// IssuerToUpsertParams maps a domain Issuer to SQLC UpsertIssuerParams.
func IssuerToUpsertParams(iss *domain.Issuer) (generated.UpsertIssuerParams, error) {
	var brandingJSON json.RawMessage
	if !iss.Branding().IsZero() {
		b, err := json.Marshal(iss.Branding().ToMap())
		if err != nil {
			return generated.UpsertIssuerParams{}, err
		}
		brandingJSON = b
	}

	var description *string
	if iss.Description() != "" {
		d := iss.Description()
		description = &d
	}

	var did *string
	if iss.DID() != "" {
		d := iss.DID()
		did = &d
	}

	var webhookURL *string
	if iss.WebhookURL() != "" {
		w := iss.WebhookURL()
		webhookURL = &w
	}

	var apiKeyHash *string
	if iss.APIKeyHash() != "" {
		a := iss.APIKeyHash()
		apiKeyHash = &a
	}

	now := time.Now().UTC()

	return generated.UpsertIssuerParams{
		ID:             uuidFromIssuerID(iss.ID()),
		OrganizationID: iss.OrganizationID(),
		Name:           iss.Name(),
		Description:    description,
		Did:            did,
		WebhookUrl:     webhookURL,
		ApiKeyHash:     apiKeyHash,
		Status:         iss.Status().String(),
		Branding:       brandingJSON,
		Version:        int32(iss.Version()),
		CreatedAt:      iss.CreatedAt(),
		UpdatedAt:      now,
	}, nil
}

// ============================================================================
// Issuer Event Mappers
// ============================================================================

// IssuerEventToInsertParams maps an eventsourcing.Event to SQLC InsertIssuerEventParams.
func IssuerEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertIssuerEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertIssuerEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertIssuerEventParams{}, err
	}

	return generated.InsertIssuerEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeIssuer,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// IssuerEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func IssuerEventRowToEvent(row generated.IssuerEvent) (eventsourcing.Event, error) {
	event, err := unmarshalIssuerEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// IssuerEventRowsToEvents maps multiple SQLC rows to events.
func IssuerEventRowsToEvents(rows []generated.IssuerEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := IssuerEventRowToEvent(row)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// ============================================================================
// Template Mappers
// ============================================================================

// TemplateToUpsertParams maps a domain Template to SQLC UpsertTemplateParams.
func TemplateToUpsertParams(t *domain.Template) (generated.UpsertTemplateParams, error) {
	var claimMappingsJSON json.RawMessage
	if t.ClaimMappings() != nil {
		b, err := json.Marshal(t.ClaimMappings())
		if err != nil {
			return generated.UpsertTemplateParams{}, err
		}
		claimMappingsJSON = b
	}

	var defaultValuesJSON json.RawMessage
	if t.DefaultValues() != nil {
		b, err := json.Marshal(t.DefaultValues())
		if err != nil {
			return generated.UpsertTemplateParams{}, err
		}
		defaultValuesJSON = b
	}

	var description *string
	if t.Description() != "" {
		d := t.Description()
		description = &d
	}

	now := time.Now().UTC()

	return generated.UpsertTemplateParams{
		ID:             uuidFromTemplateID(t.ID()),
		IssuerID:       uuidFromIssuerID(t.IssuerID()),
		Name:           t.Name(),
		Description:    description,
		SchemaType:     t.SchemaType(),
		ClaimMappings:  claimMappingsJSON,
		DefaultValues:  defaultValuesJSON,
		ExpirationDays: int32(t.ExpirationDays()),
		AutoApprove:    t.AutoApprove(),
		Status:         t.Status().String(),
		Version:        int32(t.TemplateVersion()),
		CreatedAt:      t.CreatedAt(),
		UpdatedAt:      now,
	}, nil
}

// ============================================================================
// Template Event Mappers
// ============================================================================

// TemplateEventToInsertParams maps an eventsourcing.Event to SQLC InsertTemplateEventParams.
func TemplateEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertTemplateEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertTemplateEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertTemplateEventParams{}, err
	}

	return generated.InsertTemplateEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeTemplate,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// TemplateEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func TemplateEventRowToEvent(row generated.TemplateEvent) (eventsourcing.Event, error) {
	event, err := unmarshalTemplateEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// TemplateEventRowsToEvents maps multiple SQLC rows to events.
func TemplateEventRowsToEvents(rows []generated.TemplateEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := TemplateEventRowToEvent(row)
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

// unmarshalIssuerEvent deserializes an issuer event from JSON based on event type.
func unmarshalIssuerEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeIssuerRegistered:
		var e domain.IssuerRegisteredEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeIssuerActivated:
		var e domain.IssuerActivatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeIssuerSuspended:
		var e domain.IssuerSuspendedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeBrandingUpdated:
		var e domain.BrandingUpdatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	default:
		return nil, &UnknownEventTypeError{EventType: eventType}
	}

	return event, nil
}

// unmarshalTemplateEvent deserializes a template event from JSON based on event type.
func unmarshalTemplateEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeTemplateCreated:
		var e domain.TemplateCreatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeTemplateUpdated:
		var e domain.TemplateUpdatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeTemplateArchived:
		var e domain.TemplateArchivedEvent
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

// IssuerProjection holds read-model data from the issuers table.
type IssuerProjection struct {
	ID             string
	OrganizationID string
	Name           string
	Description    string
	DID            string
	WebhookURL     string
	APIKeyHash     string
	Status         string
	Branding       map[string]any
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IssuerRowToProjection maps a generated.Issuer row to an IssuerProjection.
func IssuerRowToProjection(row generated.Issuer) IssuerProjection {
	proj := IssuerProjection{
		ID:             row.ID.String(),
		OrganizationID: row.OrganizationID,
		Name:           row.Name,
		Status:         row.Status,
		Version:        int(row.Version),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	if row.Description != nil {
		proj.Description = *row.Description
	}
	if row.Did != nil {
		proj.DID = *row.Did
	}
	if row.WebhookUrl != nil {
		proj.WebhookURL = *row.WebhookUrl
	}
	if row.ApiKeyHash != nil {
		proj.APIKeyHash = *row.ApiKeyHash
	}
	if row.Branding != nil {
		_ = json.Unmarshal(row.Branding, &proj.Branding)
	}

	return proj
}

// TemplateProjection holds read-model data from the templates table.
type TemplateProjection struct {
	ID             string
	IssuerID       string
	Name           string
	Description    string
	SchemaType     string
	ClaimMappings  map[string]string
	DefaultValues  map[string]any
	ExpirationDays int
	AutoApprove    bool
	Status         string
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TemplateRowToProjection maps a generated.Template row to a TemplateProjection.
func TemplateRowToProjection(row generated.Template) TemplateProjection {
	proj := TemplateProjection{
		ID:             row.ID.String(),
		IssuerID:       row.IssuerID.String(),
		Name:           row.Name,
		SchemaType:     row.SchemaType,
		ExpirationDays: int(row.ExpirationDays),
		AutoApprove:    row.AutoApprove,
		Status:         row.Status,
		Version:        int(row.Version),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	if row.Description != nil {
		proj.Description = *row.Description
	}
	if row.ClaimMappings != nil {
		_ = json.Unmarshal(row.ClaimMappings, &proj.ClaimMappings)
	}
	if row.DefaultValues != nil {
		_ = json.Unmarshal(row.DefaultValues, &proj.DefaultValues)
	}

	return proj
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromIssuerID converts a domain.IssuerID to uuid.UUID.
func uuidFromIssuerID(id domain.IssuerID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// uuidFromTemplateID converts a domain.TemplateID to uuid.UUID.
func uuidFromTemplateID(id domain.TemplateID) uuid.UUID {
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
