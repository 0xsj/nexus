// Package postgres provides PostgreSQL implementations of Integration domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/integration/domain"
	"github.com/0xsj/nexus/platform/internal/integration/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Integration Mappers
// ============================================================================

// IntegrationToUpsertParams maps a domain Integration to SQLC UpsertIntegrationParams.
func IntegrationToUpsertParams(integ *domain.Integration) (generated.UpsertIntegrationParams, error) {
	scopesJSON, err := json.Marshal(integ.Scopes())
	if err != nil {
		return generated.UpsertIntegrationParams{}, err
	}

	metadataJSON, err := json.Marshal(integ.Metadata())
	if err != nil {
		return generated.UpsertIntegrationParams{}, err
	}

	var providerUserID *string
	if integ.ProviderUserID() != "" {
		v := integ.ProviderUserID()
		providerUserID = &v
	}

	var providerUsername *string
	if integ.ProviderUsername() != "" {
		v := integ.ProviderUsername()
		providerUsername = &v
	}

	var lastFetchAt pgtype.Timestamptz
	if !integ.LastFetchAt().IsZero() {
		lastFetchAt = pgtype.Timestamptz{Time: integ.LastFetchAt(), Valid: true}
	}

	var connectedAt pgtype.Timestamptz
	if !integ.ConnectedAt().IsZero() {
		connectedAt = pgtype.Timestamptz{Time: integ.ConnectedAt(), Valid: true}
	}

	var disconnectedAt pgtype.Timestamptz
	if !integ.DisconnectedAt().IsZero() {
		disconnectedAt = pgtype.Timestamptz{Time: integ.DisconnectedAt(), Valid: true}
	}

	var suspendedAt pgtype.Timestamptz
	if !integ.SuspendedAt().IsZero() {
		suspendedAt = pgtype.Timestamptz{Time: integ.SuspendedAt(), Valid: true}
	}

	var suspensionReason *string
	if integ.SuspensionReason() != "" {
		r := integ.SuspensionReason()
		suspensionReason = &r
	}

	now := time.Now().UTC()

	return generated.UpsertIntegrationParams{
		ID:               uuidFromIntegrationID(integ.ID()),
		UserID:           integ.UserID(),
		ProviderType:     integ.ProviderType().String(),
		Status:           integ.Status().String(),
		ProviderUserID:   providerUserID,
		ProviderUsername: providerUsername,
		Scopes:           scopesJSON,
		LastFetchAt:      lastFetchAt,
		FetchCount:       int32(integ.FetchCount()),
		Metadata:         metadataJSON,
		ConnectedAt:      connectedAt,
		DisconnectedAt:   disconnectedAt,
		SuspendedAt:      suspendedAt,
		SuspensionReason: suspensionReason,
		Version:          int32(integ.Version()),
		CreatedAt:        integ.CreatedAt(),
		UpdatedAt:        now,
	}, nil
}

// ============================================================================
// Event Mappers
// ============================================================================

// IntegrationEventToInsertParams maps an eventsourcing.Event to SQLC InsertIntegrationEventParams.
func IntegrationEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertIntegrationEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertIntegrationEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertIntegrationEventParams{}, err
	}

	return generated.InsertIntegrationEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeIntegration,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// IntegrationEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func IntegrationEventRowToEvent(row generated.IntegrationEvent) (eventsourcing.Event, error) {
	event, err := unmarshalIntegrationEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// IntegrationEventRowsToEvents maps multiple SQLC rows to events.
func IntegrationEventRowsToEvents(rows []generated.IntegrationEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := IntegrationEventRowToEvent(row)
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

// unmarshalIntegrationEvent deserializes an integration event from JSON based on event type.
func unmarshalIntegrationEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeProviderConnected:
		var e domain.ProviderConnectedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeProviderDisconnected:
		var e domain.ProviderDisconnectedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeCredentialsRefreshed:
		var e domain.CredentialsRefreshedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeDataFetched:
		var e domain.DataFetchedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeIntegrationSuspended:
		var e domain.IntegrationSuspendedEvent
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

// IntegrationProjection holds read-model data from the integrations table.
type IntegrationProjection struct {
	ID               string
	UserID           string
	ProviderType     string
	Status           string
	ProviderUserID   string
	ProviderUsername string
	Scopes           []string
	LastFetchAt      *time.Time
	FetchCount       int
	Metadata         map[string]string
	ConnectedAt      *time.Time
	DisconnectedAt   *time.Time
	SuspendedAt      *time.Time
	SuspensionReason string
	Version          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IntegrationRowToProjection maps a generated.Integration row to an IntegrationProjection.
func IntegrationRowToProjection(row generated.Integration) IntegrationProjection {
	proj := IntegrationProjection{
		ID:           row.ID.String(),
		UserID:       row.UserID,
		ProviderType: row.ProviderType,
		Status:       row.Status,
		FetchCount:   int(row.FetchCount),
		Version:      int(row.Version),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}

	// Nullable string fields
	if row.ProviderUserID != nil {
		proj.ProviderUserID = *row.ProviderUserID
	}
	if row.ProviderUsername != nil {
		proj.ProviderUsername = *row.ProviderUsername
	}
	if row.SuspensionReason != nil {
		proj.SuspensionReason = *row.SuspensionReason
	}

	// Unmarshal scopes
	if row.Scopes != nil {
		_ = json.Unmarshal(row.Scopes, &proj.Scopes)
	}

	// Unmarshal metadata
	if row.Metadata != nil {
		_ = json.Unmarshal(row.Metadata, &proj.Metadata)
	}

	// Nullable timestamp fields
	if row.LastFetchAt.Valid {
		proj.LastFetchAt = &row.LastFetchAt.Time
	}
	if row.ConnectedAt.Valid {
		proj.ConnectedAt = &row.ConnectedAt.Time
	}
	if row.DisconnectedAt.Valid {
		proj.DisconnectedAt = &row.DisconnectedAt.Time
	}
	if row.SuspendedAt.Valid {
		proj.SuspendedAt = &row.SuspendedAt.Time
	}

	return proj
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromIntegrationID converts a domain.IntegrationID to uuid.UUID.
func uuidFromIntegrationID(id domain.IntegrationID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// integrationIDFromUUID converts a uuid.UUID to domain.IntegrationID.
func integrationIDFromUUID(u uuid.UUID) domain.IntegrationID {
	id, _ := domain.ParseIntegrationID(u.String())
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
