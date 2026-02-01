// Package postgres provides PostgreSQL implementations of Identity domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// User Mappers
// ============================================================================

// UserToUpsertParams maps a domain User to SQLC UpsertUserParams.
func UserToUpsertParams(user *domain.User) generated.UpsertUserParams {
	var email *string
	if !user.Email().IsEmpty() {
		e := user.Email().String()
		email = &e
	}

	return generated.UpsertUserParams{
		ID:          uuidFromUserID(user.ID()),
		Email:       email,
		DisplayName: user.DisplayName().String(),
		Status:      user.Status().String(),
		PrimaryDid:  user.PrimaryDID(),
		CreatedAt:   user.CreatedAt(),
		UpdatedAt:   user.UpdatedAt(),
		Version:     int32(user.Version()),
	}
}

// UserDIDToInsertParams maps a DID to SQLC InsertUserDIDParams.
func UserDIDToInsertParams(userID domain.UserID, did string, isPrimary bool, addedAt time.Time) generated.InsertUserDIDParams {
	return generated.InsertUserDIDParams{
		ID:        uuid.New(),
		UserID:    uuidFromUserID(userID),
		Did:       did,
		IsPrimary: isPrimary,
		AddedAt:   addedAt,
	}
}

// UserOAuthLinkToInsertParams maps an OAuth link to SQLC InsertUserOAuthLinkParams.
func UserOAuthLinkToInsertParams(userID domain.UserID, subject domain.OAuthSubject, email string, linkedAt time.Time) generated.InsertUserOAuthLinkParams {
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	return generated.InsertUserOAuthLinkParams{
		ID:         uuid.New(),
		UserID:     uuidFromUserID(userID),
		Provider:   subject.Provider().String(),
		ExternalID: subject.ExternalID(),
		Email:      emailPtr,
		LinkedAt:   linkedAt,
	}
}

// ============================================================================
// Session Mappers
// ============================================================================

// SessionToUpsertParams maps a domain Session to SQLC UpsertSessionParams.
func SessionToUpsertParams(session *domain.Session) generated.UpsertSessionParams {
	var ipAddress, userAgent *string
	if session.IPAddress() != "" {
		ip := session.IPAddress()
		ipAddress = &ip
	}
	if session.UserAgent() != "" {
		ua := session.UserAgent()
		userAgent = &ua
	}

	return generated.UpsertSessionParams{
		ID:         uuidFromSessionID(session.ID()),
		UserID:     uuidFromUserID(session.UserID()),
		TokenHash:  session.TokenHash(),
		AuthMethod: session.AuthMethod().String(),
		Status:     session.Status().String(),
		IpAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  session.ExpiresAt(),
		CreatedAt:  session.CreatedAt(),
		UpdatedAt:  session.UpdatedAt(),
		Version:    int32(session.Version()),
	}
}

// SessionData holds session data for read models (not aggregate reconstruction).
type SessionData struct {
	ID         domain.SessionID
	UserID     domain.UserID
	TokenHash  string
	AuthMethod domain.AuthMethod
	Status     domain.SessionStatus
	IPAddress  string
	UserAgent  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Version    int
}

// SessionRowToData maps a SQLC row to SessionData.
func SessionRowToData(row generated.IdentitySession) SessionData {
	sessionID, _ := domain.ParseSessionID(row.ID.String())
	userID, _ := domain.ParseUserID(row.UserID.String())
	authMethod, _ := domain.ParseAuthMethod(row.AuthMethod)
	status, _ := domain.ParseSessionStatus(row.Status)

	ipAddress := ""
	if row.IpAddress != nil {
		ipAddress = *row.IpAddress
	}
	userAgent := ""
	if row.UserAgent != nil {
		userAgent = *row.UserAgent
	}

	return SessionData{
		ID:         sessionID,
		UserID:     userID,
		TokenHash:  row.TokenHash,
		AuthMethod: authMethod,
		Status:     status,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		Version:    int(row.Version),
	}
}

// ============================================================================
// Magic Link Mappers
// ============================================================================

// MagicLinkRecordToInsertParams maps a domain MagicLinkRecord to SQLC params.
func MagicLinkRecordToInsertParams(record *domain.MagicLinkRecord) generated.InsertMagicLinkParams {
	return generated.InsertMagicLinkParams{
		ID:        uuid.New(),
		TokenHash: record.TokenHash,
		Email:     record.Email,
		ExpiresAt: time.Unix(record.ExpiresAt, 0),
		Used:      record.Used,
		CreatedAt: time.Unix(record.CreatedAt, 0),
	}
}

// MagicLinkRowToRecord maps a SQLC row to domain MagicLinkRecord.
func MagicLinkRowToRecord(row generated.IdentityMagicLink) *domain.MagicLinkRecord {
	return &domain.MagicLinkRecord{
		TokenHash: row.TokenHash,
		Email:     row.Email,
		ExpiresAt: row.ExpiresAt.Unix(),
		Used:      row.Used,
		CreatedAt: row.CreatedAt.Unix(),
	}
}

// ============================================================================
// OAuth State Mappers
// ============================================================================

// OAuthStateRecordToInsertParams maps a domain OAuthStateRecord to SQLC params.
func OAuthStateRecordToInsertParams(record *domain.OAuthStateRecord) generated.InsertOAuthStateParams {
	return generated.InsertOAuthStateParams{
		ID:          uuid.New(),
		State:       record.State,
		Provider:    record.Provider,
		RedirectUrl: record.RedirectURL,
		ExpiresAt:   time.Unix(record.ExpiresAt, 0),
		CreatedAt:   time.Unix(record.CreatedAt, 0),
	}
}

// OAuthStateRowToRecord maps a SQLC row to domain OAuthStateRecord.
func OAuthStateRowToRecord(row generated.IdentityOauthState) *domain.OAuthStateRecord {
	return &domain.OAuthStateRecord{
		State:       row.State,
		Provider:    row.Provider,
		RedirectURL: row.RedirectUrl,
		ExpiresAt:   row.ExpiresAt.Unix(),
		CreatedAt:   row.CreatedAt.Unix(),
	}
}

// ============================================================================
// Event Mappers
// ============================================================================

// UserEventToInsertParams maps an eventsourcing.Event to SQLC InsertUserEventParams.
// The version parameter is the event's version in the aggregate stream.
func UserEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertUserEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertUserEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertUserEventParams{}, err
	}

	return generated.InsertUserEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeUser,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// UserEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func UserEventRowToEvent(row generated.IdentityUserEvent) (eventsourcing.Event, error) {
	event, err := unmarshalUserEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// UserEventRowsToEvents maps multiple SQLC rows to events.
func UserEventRowsToEvents(rows []generated.IdentityUserEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := UserEventRowToEvent(row)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// SessionEventToInsertParams maps an eventsourcing.Event to SQLC InsertSessionEventParams.
// The version parameter is the event's version in the aggregate stream.
func SessionEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertSessionEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertSessionEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertSessionEventParams{}, err
	}

	return generated.InsertSessionEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeSession,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// SessionEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func SessionEventRowToEvent(row generated.IdentitySessionEvent) (eventsourcing.Event, error) {
	event, err := unmarshalSessionEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// SessionEventRowsToEvents maps multiple SQLC rows to events.
func SessionEventRowsToEvents(rows []generated.IdentitySessionEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := SessionEventRowToEvent(row)
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

// unmarshalUserEvent deserializes a user event from JSON based on event type.
func unmarshalUserEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeUserRegistered:
		var e domain.UserRegisteredEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserActivated:
		var e domain.UserActivatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserSuspended:
		var e domain.UserSuspendedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserReactivated:
		var e domain.UserReactivatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserDeleted:
		var e domain.UserDeletedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserDisplayNameChanged:
		var e domain.UserDisplayNameChangedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserEmailChanged:
		var e domain.UserEmailChangedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserDIDAdded:
		var e domain.UserDIDAddedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserDIDRemoved:
		var e domain.UserDIDRemovedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserOAuthLinked:
		var e domain.UserOAuthLinkedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeUserOAuthUnlinked:
		var e domain.UserOAuthUnlinkedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	default:
		return nil, &UnknownEventTypeError{EventType: eventType}
	}

	return event, nil
}

// unmarshalSessionEvent deserializes a session event from JSON based on event type.
func unmarshalSessionEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeSessionStarted:
		var e domain.SessionStartedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeSessionRefreshed:
		var e domain.SessionRefreshedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeSessionRevoked:
		var e domain.SessionRevokedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeSessionExpired:
		var e domain.SessionExpiredEvent
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
// ID Conversion Helpers
// ============================================================================

// uuidFromUserID converts a domain.UserID to uuid.UUID.
func uuidFromUserID(id domain.UserID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// uuidFromSessionID converts a domain.SessionID to uuid.UUID.
func uuidFromSessionID(id domain.SessionID) uuid.UUID {
	u, _ := uuid.Parse(id.String())
	return u
}

// userIDFromUUID converts a uuid.UUID to domain.UserID.
func userIDFromUUID(u uuid.UUID) domain.UserID {
	id, _ := domain.ParseUserID(u.String())
	return id
}

// sessionIDFromUUID converts a uuid.UUID to domain.SessionID.
func sessionIDFromUUID(u uuid.UUID) domain.SessionID {
	id, _ := domain.ParseSessionID(u.String())
	return id
}

// typesIDFromUUID converts a uuid.UUID to types.ID.
func typesIDFromUUID(u uuid.UUID) types.ID {
	id, _ := types.ParseID(u.String())
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
