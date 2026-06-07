-- +goose Up
-- +goose StatementBegin

CREATE TABLE sources (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT un_source_name UNIQUE (name)
);

CREATE TABLE event_types (
    id TEXT NOT NULL PRIMARY KEY,
    version INTEGER DEFAULT 1 NOT NULL,
    action TEXT NOT NULL,
    target_types TEXT[] NOT NULL,
    should_validate_metadata_schema BOOLEAN NOT NULL DEFAULT false,
    event_schema JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT un_event_type_name UNIQUE (action)
);

CREATE TABLE tokens (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    source_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_token_belongs_to_source FOREIGN KEY (source_id) REFERENCES sources (id)
);

-- TODO: add event_type_schemas

CREATE TABLE events (
    id TEXT NOT NULL PRIMARY KEY,
    source_id TEXT NOT NULL,
    version INT NOT NULL,
    action TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    actor_type TEXT NOT NULL,
    actor_name TEXT,
    actor_metadata JSONB,
    context_location TEXT NOT NULL,
    context_user_agent TEXT,
    metadata JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_event_belongs_to_source FOREIGN KEY (source_id) REFERENCES sources (id),
    CONSTRAINT fk_event_has_action FOREIGN KEY (action) REFERENCES event_types (action)
);

CREATE TABLE event_targets (
    internal_id TEXT NOT NULL PRIMARY KEY,
    id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    name TEXT,
    type TEXT NOT NULL,
    metadata JSONB,

    CONSTRAINT fk_event_target_belongs_to_event FOREIGN KEY (event_id) REFERENCES events (id)
);

CREATE INDEX idx_events_source_occurred_id ON events(source_id, occurred_at DESC, id DESC);
CREATE INDEX idx_event_targets_id ON event_targets(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS event_targets;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS event_types;
DROP TABLE IF EXISTS sources;
-- +goose StatementEnd
