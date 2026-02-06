// Package postgres provides PostgreSQL implementations of Presentation domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/presentation/domain"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Presentation Mappers
// ============================================================================

// PresentationToUpsertParams maps a domain Presentation to SQLC UpsertPresentationParams.
func PresentationToUpsertParams(pres *domain.Presentation) (generated.UpsertPresentationParams, error) {
	credIDsJSON, err := json.Marshal(pres.CredentialIDs())
	if err != nil {
		return generated.UpsertPresentationParams{}, err
	}

	policyJSON, err := json.Marshal(pres.DisclosurePolicy().ToMap())
	if err != nil {
		return generated.UpsertPresentationParams{}, err
	}

	var vpJWT *string
	if pres.VPJWT() != "" {
		j := pres.VPJWT()
		vpJWT = &j
	}

	var purpose *string
	if pres.Purpose() != "" {
		p := pres.Purpose()
		purpose = &p
	}

	var revokedAt pgtype.Timestamptz
	if !pres.RevokedAt().IsZero() {
		revokedAt = pgtype.Timestamptz{Time: pres.RevokedAt(), Valid: true}
	}

	var revocationReason *string
	if pres.RevocationReason() != "" {
		r := pres.RevocationReason()
		revocationReason = &r
	}

	now := time.Now().UTC()

	return generated.UpsertPresentationParams{
		ID:               uuidFromPresentationID(pres.ID()),
		HolderDid:        pres.HolderDID(),
		CredentialIds:    credIDsJSON,
		DisclosurePolicy: policyJSON,
		VpJwt:            vpJWT,
		Purpose:          purpose,
		Status:           pres.Status().String(),
		RevokedAt:        revokedAt,
		RevocationReason: revocationReason,
		Version:          int32(pres.Version()),
		CreatedAt:        pres.CreatedAt(),
		UpdatedAt:        now,
	}, nil
}

// ShareLinkToUpsertParams maps a domain ShareLink to SQLC UpsertShareLinkParams.
func ShareLinkToUpsertParams(sl *domain.ShareLink) (generated.UpsertShareLinkParams, error) {
	var expiresAt pgtype.Timestamptz
	if !sl.ExpiresAt().IsZero() {
		expiresAt = pgtype.Timestamptz{Time: sl.ExpiresAt(), Valid: true}
	}

	var pinHash *string
	if sl.PinHash() != "" {
		p := sl.PinHash()
		pinHash = &p
	}

	var audience *string
	if sl.Audience() != "" {
		a := sl.Audience()
		audience = &a
	}

	now := time.Now().UTC()

	return generated.UpsertShareLinkParams{
		ID:             uuidFromShareLinkID(sl.ID()),
		PresentationID: uuidFromPresentationID(sl.PresentationID()),
		Token:          sl.Token(),
		ExpiresAt:      expiresAt,
		MaxViews:       int32(sl.MaxViews()),
		CurrentViews:   int32(sl.CurrentViews()),
		PinHash:        pinHash,
		Audience:       audience,
		Status:         sl.Status().String(),
		Version:        int32(sl.Version()),
		CreatedAt:      sl.CreatedAt(),
		UpdatedAt:      now,
	}, nil
}

// ============================================================================
// Event Mappers
// ============================================================================

// PresentationEventToInsertParams maps an eventsourcing.Event to SQLC InsertPresentationEventParams.
func PresentationEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertPresentationEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertPresentationEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertPresentationEventParams{}, err
	}

	return generated.InsertPresentationEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypePresentation,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// ShareLinkEventToInsertParams maps an eventsourcing.Event to SQLC InsertShareLinkEventParams.
func ShareLinkEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertShareLinkEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertShareLinkEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertShareLinkEventParams{}, err
	}

	return generated.InsertShareLinkEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeShareLink,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// PresentationEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func PresentationEventRowToEvent(row generated.PresentationEvent) (eventsourcing.Event, error) {
	event, err := unmarshalPresentationEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// PresentationEventRowsToEvents maps multiple SQLC rows to events.
func PresentationEventRowsToEvents(rows []generated.PresentationEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := PresentationEventRowToEvent(row)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// ShareLinkEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func ShareLinkEventRowToEvent(row generated.ShareLinkEvent) (eventsourcing.Event, error) {
	event, err := unmarshalShareLinkEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// ShareLinkEventRowsToEvents maps multiple SQLC rows to events.
func ShareLinkEventRowsToEvents(rows []generated.ShareLinkEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := ShareLinkEventRowToEvent(row)
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

// unmarshalPresentationEvent deserializes a presentation event from JSON based on event type.
func unmarshalPresentationEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypePresentationCreated:
		var e domain.PresentationCreatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypePresentationRevoked:
		var e domain.PresentationRevokedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	default:
		return nil, &UnknownEventTypeError{EventType: eventType}
	}

	return event, nil
}

// unmarshalShareLinkEvent deserializes a share link event from JSON based on event type.
func unmarshalShareLinkEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeShareLinkCreated:
		var e domain.ShareLinkCreatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeShareLinkAccessed:
		var e domain.ShareLinkAccessedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeShareLinkRevoked:
		var e domain.ShareLinkRevokedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeShareLinkExpired:
		var e domain.ShareLinkExpiredEvent
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

// PresentationProjection holds read-model data from the presentations table.
type PresentationProjection struct {
	ID               string
	HolderDID        string
	CredentialIDs    []string
	DisclosurePolicy map[string]any
	VPJWT            string
	Purpose          string
	Status           string
	RevokedAt        *time.Time
	RevocationReason string
	Version          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PresentationRowToProjection maps a generated.Presentation row to a PresentationProjection.
func PresentationRowToProjection(row generated.Presentation) PresentationProjection {
	proj := PresentationProjection{
		ID:        row.ID.String(),
		HolderDID: row.HolderDid,
		Status:    row.Status,
		Version:   int(row.Version),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	// Unmarshal credential IDs
	if row.CredentialIds != nil {
		_ = json.Unmarshal(row.CredentialIds, &proj.CredentialIDs)
	}

	// Unmarshal disclosure policy
	if row.DisclosurePolicy != nil {
		_ = json.Unmarshal(row.DisclosurePolicy, &proj.DisclosurePolicy)
	}

	// Nullable fields
	if row.VpJwt != nil {
		proj.VPJWT = *row.VpJwt
	}
	if row.Purpose != nil {
		proj.Purpose = *row.Purpose
	}
	if row.RevokedAt.Valid {
		proj.RevokedAt = &row.RevokedAt.Time
	}
	if row.RevocationReason != nil {
		proj.RevocationReason = *row.RevocationReason
	}

	return proj
}

// ShareLinkProjection holds read-model data from the share_links table.
type ShareLinkProjection struct {
	ID             string
	PresentationID string
	Token          string
	ExpiresAt      *time.Time
	MaxViews       int
	CurrentViews   int
	PinHash        string
	Audience       string
	Status         string
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ShareLinkRowToProjection maps a generated.ShareLink row to a ShareLinkProjection.
func ShareLinkRowToProjection(row generated.ShareLink) ShareLinkProjection {
	proj := ShareLinkProjection{
		ID:             row.ID.String(),
		PresentationID: row.PresentationID.String(),
		Token:          row.Token,
		MaxViews:       int(row.MaxViews),
		CurrentViews:   int(row.CurrentViews),
		Status:         row.Status,
		Version:        int(row.Version),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	// Nullable fields
	if row.ExpiresAt.Valid {
		proj.ExpiresAt = &row.ExpiresAt.Time
	}
	if row.PinHash != nil {
		proj.PinHash = *row.PinHash
	}
	if row.Audience != nil {
		proj.Audience = *row.Audience
	}

	return proj
}

// AccessGrantProjection holds read-model data from the access_grants table.
type AccessGrantProjection struct {
	ID              string
	ShareLinkID     string
	VerifierDID     string
	AccessedAt      time.Time
	IPAddress       string
	DisclosedClaims map[string]any
}

// AccessGrantRowToProjection maps a generated.AccessGrant row to an AccessGrantProjection.
func AccessGrantRowToProjection(row generated.AccessGrant) AccessGrantProjection {
	proj := AccessGrantProjection{
		ID:          row.ID.String(),
		ShareLinkID: row.ShareLinkID.String(),
		AccessedAt:  row.AccessedAt,
	}

	if row.VerifierDid != nil {
		proj.VerifierDID = *row.VerifierDid
	}
	if row.IpAddress != nil {
		proj.IPAddress = *row.IpAddress
	}
	if row.DisclosedClaims != nil {
		_ = json.Unmarshal(row.DisclosedClaims, &proj.DisclosedClaims)
	}

	return proj
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromPresentationID converts a domain.PresentationID to uuid.UUID.
func uuidFromPresentationID(id domain.PresentationID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// presentationIDFromUUID converts a uuid.UUID to domain.PresentationID.
func presentationIDFromUUID(u uuid.UUID) domain.PresentationID {
	id, _ := domain.ParsePresentationID(u.String())
	return id
}

// uuidFromShareLinkID converts a domain.ShareLinkID to uuid.UUID.
func uuidFromShareLinkID(id domain.ShareLinkID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// shareLinkIDFromUUID converts a uuid.UUID to domain.ShareLinkID.
func shareLinkIDFromUUID(u uuid.UUID) domain.ShareLinkID {
	id, _ := domain.ParseShareLinkID(u.String())
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
