// Package postgres provides PostgreSQL implementations of Credential domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Credential Mappers
// ============================================================================

// CredentialToUpsertParams maps a domain Credential to SQLC UpsertCredentialParams.
func CredentialToUpsertParams(cred *domain.Credential) (generated.UpsertCredentialParams, error) {
	claimsJSON, err := json.Marshal(cred.Claims().ToMap())
	if err != nil {
		return generated.UpsertCredentialParams{}, err
	}

	var expiresAt pgtype.Timestamptz
	if !cred.ExpiresAt().IsZero() {
		expiresAt = pgtype.Timestamptz{Time: cred.ExpiresAt(), Valid: true}
	}

	var revokedAt pgtype.Timestamptz
	if !cred.RevokedAt().IsZero() {
		revokedAt = pgtype.Timestamptz{Time: cred.RevokedAt(), Valid: true}
	}

	var revocationReason *string
	if cred.RevocationReason() != "" {
		r := cred.RevocationReason()
		revocationReason = &r
	}

	var jwt *string
	if cred.JWT() != "" {
		j := cred.JWT()
		jwt = &j
	}

	var verificationID *string
	if cred.VerificationID() != "" {
		v := cred.VerificationID()
		verificationID = &v
	}

	now := time.Now().UTC()

	return generated.UpsertCredentialParams{
		ID:               uuidFromCredentialID(cred.ID()),
		CredentialType:   cred.Type().String(),
		IssuerDid:        cred.IssuerDID(),
		SubjectDid:       cred.SubjectDID(),
		Claims:           claimsJSON,
		IssuedAt:         cred.IssuedAt(),
		ExpiresAt:        expiresAt,
		Status:           cred.Status().String(),
		RevokedAt:        revokedAt,
		RevocationReason: revocationReason,
		Jwt:              jwt,
		VerificationID:   verificationID,
		Version:          int32(cred.Version()),
		CreatedAt:        cred.IssuedAt(),
		UpdatedAt:        now,
	}, nil
}

// ============================================================================
// Event Mappers
// ============================================================================

// CredentialEventToInsertParams maps an eventsourcing.Event to SQLC InsertCredentialEventParams.
func CredentialEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertCredentialEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertCredentialEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertCredentialEventParams{}, err
	}

	return generated.InsertCredentialEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeCredential,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// CredentialEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func CredentialEventRowToEvent(row generated.CredentialEvent) (eventsourcing.Event, error) {
	event, err := unmarshalCredentialEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// CredentialEventRowsToEvents maps multiple SQLC rows to events.
func CredentialEventRowsToEvents(rows []generated.CredentialEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := CredentialEventRowToEvent(row)
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

// unmarshalCredentialEvent deserializes a credential event from JSON based on event type.
func unmarshalCredentialEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeCredentialIssued:
		var e domain.CredentialIssuedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeCredentialRevoked:
		var e domain.CredentialRevokedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeCredentialExpired:
		var e domain.CredentialExpiredEvent
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

// CredentialProjection holds read-model data from the credentials table.
type CredentialProjection struct {
	ID               string
	CredentialType   string
	IssuerDID        string
	SubjectDID       string
	Claims           map[string]any
	IssuedAt         time.Time
	ExpiresAt        *time.Time
	Status           string
	RevokedAt        *time.Time
	RevocationReason string
	JWT              string
	VerificationID   string
	Version          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CredentialRowToProjection maps a generated.Credential row to a CredentialProjection.
func CredentialRowToProjection(row generated.Credential) CredentialProjection {
	proj := CredentialProjection{
		ID:             row.ID.String(),
		CredentialType: row.CredentialType,
		IssuerDID:      row.IssuerDid,
		SubjectDID:     row.SubjectDid,
		IssuedAt:       row.IssuedAt,
		Status:         row.Status,
		Version:        int(row.Version),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	// Unmarshal claims
	if row.Claims != nil {
		_ = json.Unmarshal(row.Claims, &proj.Claims)
	}

	// Nullable fields
	if row.ExpiresAt.Valid {
		proj.ExpiresAt = &row.ExpiresAt.Time
	}
	if row.RevokedAt.Valid {
		proj.RevokedAt = &row.RevokedAt.Time
	}
	if row.RevocationReason != nil {
		proj.RevocationReason = *row.RevocationReason
	}
	if row.Jwt != nil {
		proj.JWT = *row.Jwt
	}
	if row.VerificationID != nil {
		proj.VerificationID = *row.VerificationID
	}

	return proj
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromCredentialID converts a domain.CredentialID to uuid.UUID.
func uuidFromCredentialID(id domain.CredentialID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// credentialIDFromUUID converts a uuid.UUID to domain.CredentialID.
func credentialIDFromUUID(u uuid.UUID) domain.CredentialID {
	id, _ := domain.ParseCredentialID(u.String())
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
