// Package postgres provides PostgreSQL implementations of Notification domain repositories.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/notification/domain"
	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Notification Mappers
// ============================================================================

// NotificationToUpsertParams maps a domain Notification to SQLC UpsertNotificationParams.
func NotificationToUpsertParams(n *domain.Notification) generated.UpsertNotificationParams {
	var templateID *string
	if n.TemplateID() != "" {
		t := n.TemplateID()
		templateID = &t
	}

	var actionURL *string
	if n.ActionURL() != "" {
		a := n.ActionURL()
		actionURL = &a
	}

	var errorMessage *string
	// We derive error message from delivery attempts if status is failed.
	if n.Status() == domain.DeliveryStatusFailed {
		attempts := n.DeliveryAttempts()
		if len(attempts) > 0 {
			lastAttempt := attempts[len(attempts)-1]
			if lastAttempt.ErrorMessage() != "" {
				msg := lastAttempt.ErrorMessage()
				errorMessage = &msg
			}
		}
	}

	var sentAt pgtype.Timestamptz
	if !n.SentAt().IsZero() {
		sentAt = pgtype.Timestamptz{Time: n.SentAt(), Valid: true}
	}

	var readAt pgtype.Timestamptz
	if !n.ReadAt().IsZero() {
		readAt = pgtype.Timestamptz{Time: n.ReadAt(), Valid: true}
	}

	now := time.Now().UTC()

	return generated.UpsertNotificationParams{
		ID:           uuidFromNotificationID(n.ID()),
		RecipientID:  n.RecipientID(),
		Category:     n.Category().String(),
		Channel:      n.Channel().String(),
		TemplateID:   templateID,
		Subject:      n.Subject(),
		Body:         n.Body(),
		ActionUrl:    actionURL,
		Status:       n.Status().String(),
		ErrorMessage: errorMessage,
		Version:      int32(n.Version()),
		CreatedAt:    n.CreatedAt(),
		SentAt:       sentAt,
		ReadAt:       readAt,
		UpdatedAt:    now,
	}
}

// ============================================================================
// Event Mappers
// ============================================================================

// NotificationEventToInsertParams maps an eventsourcing.Event to SQLC InsertNotificationEventParams.
func NotificationEventToInsertParams(aggregateID string, event eventsourcing.Event, version int) (generated.InsertNotificationEventParams, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return generated.InsertNotificationEventParams{}, err
	}

	aggUUID, err := uuid.Parse(aggregateID)
	if err != nil {
		return generated.InsertNotificationEventParams{}, err
	}

	return generated.InsertNotificationEventParams{
		ID:            uuid.New(),
		AggregateID:   aggUUID,
		AggregateType: domain.AggregateTypeNotification,
		EventType:     event.EventType(),
		EventData:     eventData,
		Version:       int32(version),
		OccurredAt:    event.OccurredAt(),
	}, nil
}

// NotificationEventRowToEvent maps a SQLC row to an eventsourcing.Event.
func NotificationEventRowToEvent(row generated.NotificationEvent) (eventsourcing.Event, error) {
	event, err := unmarshalNotificationEvent(row.EventType, row.EventData)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// NotificationEventRowsToEvents maps multiple SQLC rows to events.
func NotificationEventRowsToEvents(rows []generated.NotificationEvent) ([]eventsourcing.Event, error) {
	events := make([]eventsourcing.Event, 0, len(rows))
	for _, row := range rows {
		event, err := NotificationEventRowToEvent(row)
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

// unmarshalNotificationEvent deserializes a notification event from JSON based on event type.
func unmarshalNotificationEvent(eventType string, data json.RawMessage) (eventsourcing.Event, error) {
	var event eventsourcing.Event

	switch eventType {
	case domain.EventTypeNotificationCreated:
		var e domain.NotificationCreatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeNotificationSent:
		var e domain.NotificationSentEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeNotificationFailed:
		var e domain.NotificationFailedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeNotificationRead:
		var e domain.NotificationReadEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = &e

	case domain.EventTypeNotificationSuppressed:
		var e domain.NotificationSuppressedEvent
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

// NotificationProjection holds read-model data from the notifications table.
type NotificationProjection struct {
	ID           string
	RecipientID  string
	Category     string
	Channel      string
	TemplateID   string
	Subject      string
	Body         string
	ActionURL    string
	Status       string
	ErrorMessage string
	Version      int
	CreatedAt    time.Time
	SentAt       *time.Time
	ReadAt       *time.Time
	UpdatedAt    time.Time
}

// NotificationRowToProjection maps a generated.Notification row to a NotificationProjection.
func NotificationRowToProjection(row generated.Notification) NotificationProjection {
	proj := NotificationProjection{
		ID:          row.ID.String(),
		RecipientID: row.RecipientID,
		Category:    row.Category,
		Channel:     row.Channel,
		Subject:     row.Subject,
		Body:        row.Body,
		Status:      row.Status,
		Version:     int(row.Version),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}

	// Nullable fields
	if row.TemplateID != nil {
		proj.TemplateID = *row.TemplateID
	}
	if row.ActionUrl != nil {
		proj.ActionURL = *row.ActionUrl
	}
	if row.ErrorMessage != nil {
		proj.ErrorMessage = *row.ErrorMessage
	}
	if row.SentAt.Valid {
		proj.SentAt = &row.SentAt.Time
	}
	if row.ReadAt.Valid {
		proj.ReadAt = &row.ReadAt.Time
	}

	return proj
}

// ============================================================================
// Preferences Mapping
// ============================================================================

// PreferencesToUpsertParams maps a domain NotificationPreferences to SQLC UpsertPreferencesParams.
func PreferencesToUpsertParams(prefs domain.NotificationPreferences) (generated.UpsertPreferencesParams, error) {
	channelConfig, err := marshalChannelConfig(prefs.ChannelEnabled())
	if err != nil {
		return generated.UpsertPreferencesParams{}, err
	}

	categoryConfig, err := marshalCategoryConfig(prefs.CategoryChannels())
	if err != nil {
		return generated.UpsertPreferencesParams{}, err
	}

	now := time.Now().UTC()

	return generated.UpsertPreferencesParams{
		UserID:          prefs.UserID(),
		GlobalEnabled:   prefs.GlobalEnabled(),
		ChannelConfig:   channelConfig,
		CategoryConfig:  categoryConfig,
		DigestEnabled:   prefs.DigestEnabled(),
		DigestFrequency: prefs.DigestFrequency().String(),
		Timezone:        prefs.Timezone(),
		UpdatedAt:       now,
	}, nil
}

// PreferencesFromRow maps a generated.NotificationPreference row to a domain NotificationPreferences.
func PreferencesFromRow(row generated.NotificationPreference) (domain.NotificationPreferences, error) {
	prefs := domain.NewDefaultPreferences(row.UserID)
	prefs = prefs.WithGlobalEnabled(row.GlobalEnabled)
	prefs = prefs.WithDigestEnabled(row.DigestEnabled)

	// Parse digest frequency
	if row.DigestFrequency != "" {
		freq, err := domain.ParseDigestFrequency(row.DigestFrequency)
		if err == nil {
			prefs = prefs.WithDigestFrequency(freq)
		}
	}

	// Parse timezone
	if row.Timezone != "" {
		prefs = prefs.WithTimezone(row.Timezone)
	}

	// Unmarshal channel config
	if len(row.ChannelConfig) > 0 {
		channelMap := make(map[string]bool)
		if err := json.Unmarshal(row.ChannelConfig, &channelMap); err == nil {
			for chStr, enabled := range channelMap {
				ch, err := domain.ParseChannel(chStr)
				if err == nil {
					prefs = prefs.WithChannelEnabled(ch, enabled)
				}
			}
		}
	}

	// Unmarshal category config
	if len(row.CategoryConfig) > 0 {
		catMap := make(map[string]map[string]bool)
		if err := json.Unmarshal(row.CategoryConfig, &catMap); err == nil {
			for catStr, channels := range catMap {
				cat, err := domain.ParseCategory(catStr)
				if err != nil {
					continue
				}
				for chStr, enabled := range channels {
					ch, err := domain.ParseChannel(chStr)
					if err == nil {
						prefs = prefs.WithCategoryChannel(cat, ch, enabled)
					}
				}
			}
		}
	}

	return prefs, nil
}

// ============================================================================
// JSON Marshaling Helpers
// ============================================================================

// marshalChannelConfig serializes channel enabled map to JSON.
func marshalChannelConfig(channels map[domain.Channel]bool) ([]byte, error) {
	m := make(map[string]bool, len(channels))
	for ch, enabled := range channels {
		m[ch.String()] = enabled
	}
	return json.Marshal(m)
}

// marshalCategoryConfig serializes category-channel map to JSON.
func marshalCategoryConfig(categories map[domain.Category]map[domain.Channel]bool) ([]byte, error) {
	m := make(map[string]map[string]bool, len(categories))
	for cat, channels := range categories {
		chMap := make(map[string]bool, len(channels))
		for ch, enabled := range channels {
			chMap[ch.String()] = enabled
		}
		m[cat.String()] = chMap
	}
	return json.Marshal(m)
}

// ============================================================================
// ID Conversion Helpers
// ============================================================================

// uuidFromNotificationID converts a domain.NotificationID to uuid.UUID.
func uuidFromNotificationID(id domain.NotificationID) uuid.UUID {
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
