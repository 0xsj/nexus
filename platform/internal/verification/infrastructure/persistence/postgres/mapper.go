// Package postgres provides PostgreSQL implementations of Verification domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/internal/verification/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Verification Mappers
// ============================================================================

// VerificationToUpsertParams maps a domain Verification to SQLC UpsertVerificationParams.
func VerificationToUpsertParams(v *domain.Verification) (generated.UpsertVerificationParams, error) {
	var errorMessage *string
	if v.VerificationError().Message() != "" {
		m := v.VerificationError().Message()
		errorMessage = &m
	}

	var errorCode *string
	if v.VerificationError().Code() != "" {
		c := v.VerificationError().Code()
		errorCode = &c
	}

	var credentialID *string
	if v.CredentialID() != "" {
		c := v.CredentialID()
		credentialID = &c
	}

	var completedAt pgtype.Timestamptz
	if !v.CompletedAt().IsZero() {
		completedAt = pgtype.Timestamptz{Time: v.CompletedAt(), Valid: true}
	}

	now := time.Now().UTC()

	return generated.UpsertVerificationParams{
		ID:           uuidFromVerificationID(v.ID()),
		UserID:       v.UserID(),
		ProviderType: v.Provider().String(),
		Status:       v.Status().String(),
		OauthState:   v.OAuthState(),
		ErrorMessage: errorMessage,
		ErrorCode:    errorCode,
		CredentialID: credentialID,
		Version:      int32(v.Version()),
		StartedAt:    v.StartedAt(),
		CompletedAt:  completedAt,
		CreatedAt:    v.StartedAt(),
		UpdatedAt:    now,
	}, nil
}

// ============================================================================
// Event Mappers
// ============================================================================

// VerificationEventToInsertParams maps an eventsourcing.Event to SQLC InsertVerificationEventParams.
func VerificationEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertVerificationEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertVerificationEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertVerificationEventParams{}, err
	}

	return generated.InsertVerificationEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeVerification,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// VerificationEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func VerificationEventRowToEvent(row generated.VerificationEvent) (eventsourcing.Event, error) {
	event, err := unmarshalVerificationEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// VerificationEventRowsToEvents maps multiple SQLC rows to events.
func VerificationEventRowsToEvents(rows []generated.VerificationEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := VerificationEventRowToEvent(row)
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

// unmarshalVerificationEvent deserializes a verification event from JSON based on event type.
func unmarshalVerificationEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeVerificationStarted:
		var e domain.VerificationStartedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeOAuthCallbackReceived:
		var e domain.OAuthCallbackReceivedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeDataFetchCompleted:
		var e domain.DataFetchCompletedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeDataFetchFailed:
		var e domain.DataFetchFailedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeCredentialIssueRequested:
		var e domain.CredentialIssueRequestedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeVerificationCompleted:
		var e domain.VerificationCompletedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeVerificationFailed:
		var e domain.VerificationFailedEvent
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

// VerificationProjection holds read-model data from the verifications table.
type VerificationProjection struct {
	ID           string
	UserID       string
	ProviderType string
	Status       string
	OAuthState   string
	ErrorMessage string
	ErrorCode    string
	CredentialID string
	Version      int
	StartedAt    time.Time
	CompletedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// VerificationRowToProjection maps a generated.Verification row to a VerificationProjection.
func VerificationRowToProjection(row generated.Verification) VerificationProjection {
	proj := VerificationProjection{
		ID:           row.ID.String(),
		UserID:       row.UserID,
		ProviderType: row.ProviderType,
		Status:       row.Status,
		OAuthState:   row.OauthState,
		Version:      int(row.Version),
		StartedAt:    row.StartedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}

	// Nullable fields
	if row.CompletedAt.Valid {
		proj.CompletedAt = &row.CompletedAt.Time
	}
	if row.ErrorMessage != nil {
		proj.ErrorMessage = *row.ErrorMessage
	}
	if row.ErrorCode != nil {
		proj.ErrorCode = *row.ErrorCode
	}
	if row.CredentialID != nil {
		proj.CredentialID = *row.CredentialID
	}

	return proj
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromVerificationID converts a domain.VerificationID to uuid.UUID.
func uuidFromVerificationID(id domain.VerificationID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// verificationIDFromUUID converts a uuid.UUID to domain.VerificationID.
func verificationIDFromUUID(u uuid.UUID) domain.VerificationID {
	id, _ := domain.ParseVerificationID(u.String())
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
