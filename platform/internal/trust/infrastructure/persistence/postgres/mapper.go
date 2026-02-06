// Package postgres provides PostgreSQL implementations of Trust domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/trust/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Vouch Mappers
// ============================================================================

// VouchToUpsertParams maps a domain Vouch to SQLC UpsertVouchParams.
func VouchToUpsertParams(vouch *domain.Vouch) (generated.UpsertVouchParams, error) {
	var credentialID *string
	if vouch.CredentialID() != "" {
		c := vouch.CredentialID()
		credentialID = &c
	}

	var claimKey *string
	if vouch.ClaimKey() != "" {
		c := vouch.ClaimKey()
		claimKey = &c
	}

	var statement *string
	if vouch.Statement() != "" {
		s := vouch.Statement()
		statement = &s
	}

	var vouchContext *string
	if vouch.VouchContext() != "" {
		c := vouch.VouchContext()
		vouchContext = &c
	}

	var expiresAt pgtype.Timestamptz
	if !vouch.ExpiresAt().IsZero() {
		expiresAt = pgtype.Timestamptz{Time: vouch.ExpiresAt(), Valid: true}
	}

	var acceptedAt pgtype.Timestamptz
	if !vouch.AcceptedAt().IsZero() {
		acceptedAt = pgtype.Timestamptz{Time: vouch.AcceptedAt(), Valid: true}
	}

	var revokedAt pgtype.Timestamptz
	if !vouch.RevokedAt().IsZero() {
		revokedAt = pgtype.Timestamptz{Time: vouch.RevokedAt(), Valid: true}
	}

	var revocationReason *string
	if vouch.RevocationReason() != "" {
		r := vouch.RevocationReason()
		revocationReason = &r
	}

	now := time.Now().UTC()

	return generated.UpsertVouchParams{
		ID:               uuidFromVouchID(vouch.ID()),
		VoucherID:        vouch.VoucherID(),
		VoucheeID:        vouch.VoucheeID(),
		CredentialID:     credentialID,
		ClaimKey:         claimKey,
		Relationship:     vouch.Relationship().String(),
		Strength:         int32(vouch.Strength().Value()),
		Statement:        statement,
		Context:          vouchContext,
		Status:           vouch.Status().String(),
		ExpiresAt:        expiresAt,
		AcceptedAt:       acceptedAt,
		RevokedAt:        revokedAt,
		RevocationReason: revocationReason,
		Version:          int32(vouch.Version()),
		CreatedAt:        vouch.CreatedAt(),
		UpdatedAt:        now,
	}, nil
}

// ============================================================================
// Event Mappers
// ============================================================================

// VouchEventToInsertParams maps an eventsourcing.Event to SQLC InsertVouchEventParams.
func VouchEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertVouchEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertVouchEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertVouchEventParams{}, err
	}

	return generated.InsertVouchEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeVouch,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// VouchEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func VouchEventRowToEvent(row generated.VouchEvent) (eventsourcing.Event, error) {
	event, err := unmarshalVouchEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// VouchEventRowsToEvents maps multiple SQLC rows to events.
func VouchEventRowsToEvents(rows []generated.VouchEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := VouchEventRowToEvent(row)
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

// unmarshalVouchEvent deserializes a vouch event from JSON based on event type.
func unmarshalVouchEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeVouchGiven:
		var e domain.VouchGivenEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeVouchAccepted:
		var e domain.VouchAcceptedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeVouchRevoked:
		var e domain.VouchRevokedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeVouchExpired:
		var e domain.VouchExpiredEvent
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

// VouchProjection holds read-model data from the vouches table.
type VouchProjection struct {
	ID               string
	VoucherID        string
	VoucheeID        string
	CredentialID     string
	ClaimKey         string
	Relationship     string
	Strength         int
	Statement        string
	Context          string
	Status           string
	ExpiresAt        *time.Time
	AcceptedAt       *time.Time
	RevokedAt        *time.Time
	RevocationReason string
	Version          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// VouchRowToProjection maps a generated.Vouch row to a VouchProjection.
func VouchRowToProjection(row generated.Vouch) VouchProjection {
	proj := VouchProjection{
		ID:           row.ID.String(),
		VoucherID:    row.VoucherID,
		VoucheeID:    row.VoucheeID,
		Relationship: row.Relationship,
		Strength:     int(row.Strength),
		Status:       row.Status,
		Version:      int(row.Version),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}

	// Nullable fields
	if row.CredentialID != nil {
		proj.CredentialID = *row.CredentialID
	}
	if row.ClaimKey != nil {
		proj.ClaimKey = *row.ClaimKey
	}
	if row.Statement != nil {
		proj.Statement = *row.Statement
	}
	if row.Context != nil {
		proj.Context = *row.Context
	}
	if row.ExpiresAt.Valid {
		proj.ExpiresAt = &row.ExpiresAt.Time
	}
	if row.AcceptedAt.Valid {
		proj.AcceptedAt = &row.AcceptedAt.Time
	}
	if row.RevokedAt.Valid {
		proj.RevokedAt = &row.RevokedAt.Time
	}
	if row.RevocationReason != nil {
		proj.RevocationReason = *row.RevocationReason
	}

	return proj
}

// ReputationProjection holds read-model data from the reputations table.
type ReputationProjection struct {
	UserID           string
	OverallScore     int
	CredentialScore  int
	VouchScore       int
	NetworkScore     int
	VouchCount       int
	LastCalculatedAt time.Time
}

// ReputationRowToProjection maps a generated.Reputation row to a ReputationProjection.
func ReputationRowToProjection(row generated.Reputation) ReputationProjection {
	return ReputationProjection{
		UserID:           row.UserID,
		OverallScore:     int(row.OverallScore),
		CredentialScore:  int(row.CredentialScore),
		VouchScore:       int(row.VouchScore),
		NetworkScore:     int(row.NetworkScore),
		VouchCount:       int(row.VouchCount),
		LastCalculatedAt: row.LastCalculatedAt,
	}
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromVouchID converts a domain.VouchID to uuid.UUID.
func uuidFromVouchID(id domain.VouchID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// vouchIDFromUUID converts a uuid.UUID to domain.VouchID.
func vouchIDFromUUID(u uuid.UUID) domain.VouchID {
	id, _ := domain.ParseVouchID(u.String())
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
