-- ============================================================================
-- Notification Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Notification events store
CREATE TABLE IF NOT EXISTS notification_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Notification',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_notification_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_notification_events_aggregate_id ON notification_events(aggregate_id);
CREATE INDEX idx_notification_events_event_type ON notification_events(event_type);
CREATE INDEX idx_notification_events_occurred_at ON notification_events(occurred_at);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Notifications projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY,
    recipient_id    TEXT NOT NULL,
    category        TEXT NOT NULL,
    channel         TEXT NOT NULL,
    template_id     TEXT,
    subject         TEXT NOT NULL,
    body            TEXT NOT NULL,
    action_url      TEXT,
    status          TEXT NOT NULL DEFAULT 'queued',
    error_message   TEXT,
    version         INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at         TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_category ON notifications(category);
CREATE INDEX idx_notifications_recipient_status ON notifications(recipient_id, status);

-- ============================================================================
-- Notification Preferences Table (CRUD, not event-sourced)
-- ============================================================================

CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id          TEXT PRIMARY KEY,
    global_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    channel_config   JSONB,
    category_config  JSONB,
    digest_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    digest_frequency TEXT NOT NULL DEFAULT 'daily',
    timezone         TEXT NOT NULL DEFAULT 'UTC',
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- Cross-Context Projection Tables
-- ============================================================================

-- Identity user projection (populated via User.* events from Identity context)
CREATE TABLE IF NOT EXISTS notification_user_projections (
    user_id     VARCHAR(255) PRIMARY KEY,
    email       VARCHAR(255) NOT NULL DEFAULT '',
    active      BOOLEAN NOT NULL DEFAULT true,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
