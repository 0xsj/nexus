-- ============================================================================
-- Notification Event Store Queries
-- ============================================================================

-- name: InsertNotificationEvent :exec
INSERT INTO notification_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetNotificationEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM notification_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestNotificationEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM notification_events
WHERE aggregate_id = $1;

-- name: NotificationAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM notification_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Notification Projection Queries
-- ============================================================================

-- name: UpsertNotification :exec
INSERT INTO notifications (
    id, recipient_id, category, channel, template_id, subject, body,
    action_url, status, error_message, version, created_at, sent_at, read_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    error_message = EXCLUDED.error_message,
    sent_at = EXCLUDED.sent_at,
    read_at = EXCLUDED.read_at,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetNotificationByID :one
SELECT id, recipient_id, category, channel, template_id, subject, body,
       action_url, status, error_message, version, created_at, sent_at, read_at, updated_at
FROM notifications
WHERE id = $1;

-- name: NotificationExistsByID :one
SELECT EXISTS (
    SELECT 1 FROM notifications WHERE id = $1
) AS exists;

-- name: ListNotificationsByRecipient :many
SELECT id, recipient_id, category, channel, template_id, subject, body,
       action_url, status, error_message, version, created_at, sent_at, read_at, updated_at
FROM notifications
WHERE recipient_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountNotificationsByRecipient :one
SELECT COUNT(*)::integer AS count
FROM notifications
WHERE recipient_id = $1;

-- name: CountUnreadByRecipient :one
SELECT COUNT(*)::integer AS count
FROM notifications
WHERE recipient_id = $1
  AND status NOT IN ('read', 'suppressed', 'failed');

-- ============================================================================
-- Notification Preferences Queries
-- ============================================================================

-- name: UpsertPreferences :exec
INSERT INTO notification_preferences (
    user_id, global_enabled, channel_config, category_config,
    digest_enabled, digest_frequency, timezone, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (user_id) DO UPDATE SET
    global_enabled = EXCLUDED.global_enabled,
    channel_config = EXCLUDED.channel_config,
    category_config = EXCLUDED.category_config,
    digest_enabled = EXCLUDED.digest_enabled,
    digest_frequency = EXCLUDED.digest_frequency,
    timezone = EXCLUDED.timezone,
    updated_at = EXCLUDED.updated_at;

-- name: GetPreferencesByUserID :one
SELECT user_id, global_enabled, channel_config, category_config,
       digest_enabled, digest_frequency, timezone, updated_at
FROM notification_preferences
WHERE user_id = $1;

-- name: PreferencesExistByUserID :one
SELECT EXISTS (
    SELECT 1 FROM notification_preferences WHERE user_id = $1
) AS exists;

-- ============================================================================
-- Identity User Projection Queries
-- ============================================================================

-- name: UpsertUserProjection :exec
INSERT INTO notification_user_projections (user_id, email, active, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (user_id) DO UPDATE SET
    email = EXCLUDED.email,
    active = EXCLUDED.active,
    updated_at = EXCLUDED.updated_at;

-- name: GetUserProjection :one
SELECT user_id, email, active, updated_at
FROM notification_user_projections
WHERE user_id = $1;

-- name: UserProjectionExists :one
SELECT EXISTS (
    SELECT 1 FROM notification_user_projections WHERE user_id = $1 AND active = true
) AS exists;

-- name: UpdateUserProjectionEmail :exec
UPDATE notification_user_projections SET email = $2, updated_at = NOW() WHERE user_id = $1;

-- name: UpdateUserProjectionActive :exec
UPDATE notification_user_projections SET active = $2, updated_at = NOW() WHERE user_id = $1;
