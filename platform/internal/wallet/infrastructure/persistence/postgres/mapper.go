// Package postgres provides PostgreSQL implementations of Wallet domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Wallet Mappers
// ============================================================================

// WalletToUpsertParams maps a domain Wallet to SQLC UpsertWalletParams.
func WalletToUpsertParams(w *domain.Wallet) (generated.UpsertWalletParams, error) {
	var label *string
	if !w.Label().IsZero() {
		l := w.Label().String()
		label = &l
	}

	var did *string
	if w.DID() != "" {
		d := w.DID()
		did = &d
	}

	var verifiedAt pgtype.Timestamptz
	if !w.VerifiedAt().IsZero() {
		verifiedAt = pgtype.Timestamptz{Time: w.VerifiedAt(), Valid: true}
	}

	now := time.Now().UTC()

	return generated.UpsertWalletParams{
		ID:         uuidFromWalletID(w.ID()),
		UserID:     w.UserID(),
		Address:    w.Address().String(),
		Chain:      w.Chain().String(),
		Label:      label,
		Status:     w.Status().String(),
		IsPrimary:  w.IsPrimary(),
		Did:        did,
		LinkedAt:   w.CreatedAt(),
		VerifiedAt: verifiedAt,
		Version:    int32(w.Version()),
		CreatedAt:  w.CreatedAt(),
		UpdatedAt:  now,
	}, nil
}

// ============================================================================
// Event Mappers
// ============================================================================

// WalletEventToInsertParams maps an eventsourcing.Event to SQLC InsertWalletEventParams.
func WalletEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertWalletEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertWalletEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertWalletEventParams{}, err
	}

	return generated.InsertWalletEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeWallet,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// WalletEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func WalletEventRowToEvent(row generated.WalletEvent) (eventsourcing.Event, error) {
	event, err := unmarshalWalletEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// WalletEventRowsToEvents maps multiple SQLC rows to events.
func WalletEventRowsToEvents(rows []generated.WalletEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := WalletEventRowToEvent(row)
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

// unmarshalWalletEvent deserializes a wallet event from JSON based on event type.
func unmarshalWalletEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeWalletLinked:
		var e domain.WalletLinkedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeWalletVerified:
		var e domain.WalletVerifiedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeWalletUnlinked:
		var e domain.WalletUnlinkedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeWalletSetPrimary:
		var e domain.WalletSetPrimaryEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeWalletLabelUpdated:
		var e domain.WalletLabelUpdatedEvent
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

// WalletProjection holds read-model data from the wallets table.
type WalletProjection struct {
	ID         string
	UserID     string
	Address    string
	Chain      string
	Label      string
	Status     string
	IsPrimary  bool
	DID        string
	LinkedAt   time.Time
	VerifiedAt *time.Time
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// WalletRowToProjection maps a generated.Wallet row to a WalletProjection.
func WalletRowToProjection(row generated.Wallet) WalletProjection {
	proj := WalletProjection{
		ID:        row.ID.String(),
		UserID:    row.UserID,
		Address:   row.Address,
		Chain:     row.Chain,
		Status:    row.Status,
		IsPrimary: row.IsPrimary,
		LinkedAt:  row.LinkedAt,
		Version:   int(row.Version),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	// Nullable fields
	if row.Label != nil {
		proj.Label = *row.Label
	}
	if row.Did != nil {
		proj.DID = *row.Did
	}
	if row.VerifiedAt.Valid {
		proj.VerifiedAt = &row.VerifiedAt.Time
	}

	return proj
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromWalletID converts a domain.WalletID to uuid.UUID.
func uuidFromWalletID(id domain.WalletID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// walletIDFromUUID converts a uuid.UUID to domain.WalletID.
func walletIDFromUUID(u uuid.UUID) domain.WalletID {
	id, _ := domain.ParseWalletID(u.String())
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
