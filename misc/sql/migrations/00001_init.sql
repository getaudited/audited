-- +goose Up
-- +goose StatementBegin
-- CREATE TABLE tenants (
--     id TEXT NOT NULL PRIMARY KEY,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
--     disabled_at TIMESTAMPTZ
-- );
--
-- INSERT INTO tenants (id) VALUES ('tnt_01JZ3DXWVZKFMVJ500BCDK7BHP');

CREATE TABLE sources (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE event_types (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL, -- TODO: REMOVE THIS
    version INTEGER DEFAULT 1 NOT NULL,
    action TEXT NOT NULL,
    target_types TEXT[] NOT NULL,
    should_validate_metadata_schema BOOLEAN NOT NULL DEFAULT false,
    event_schema JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT un_event_type_name UNIQUE (action, tenant_id)
    -- TODO: Enforce this later
    -- CONSTRAINT fk_event_types_belongs_to_tenants
    --     FOREIGN KEY (tenant_id)
    --     REFERENCES tenants (id)
    --     ON DELETE CASCADE
);

-- TODO: add event_type_schemas

CREATE TABLE events (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    source_id TEXT NOT NULL, -- TODO: enforce constraint later
    version INT NOT NULL,
    actor_id TEXT NOT NULL,
    actor_type TEXT NOT NULL,
    actor_name TEXT,
    actor_metadata JSONB,
    context_location TEXT NOT NULL,
    context_user_agent TEXT,
    metadata JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()

-- TODO: Enforce this later
-- CONSTRAINT fk_event_types_belongs_to_tenants
--     FOREIGN KEY (tenant_id)
--     REFERENCES tenants (id)
--     ON DELETE CASCADE
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sources;
DROP TABLE event_targets;
DROP TABLE events;
DROP TABLE event_types;
-- +goose StatementEnd
