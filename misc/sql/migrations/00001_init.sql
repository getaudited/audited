-- +goose Up
-- +goose StatementBegin
-- CREATE TABLE tenants (
--     id TEXT NOT NULL PRIMARY KEY,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
--     disabled_at TIMESTAMPTZ
-- );
--
-- INSERT INTO tenants (id) VALUES ('tnt_01JZ3DXWVZKFMVJ500BCDK7BHP');

CREATE TABLE event_types (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    action TEXT NOT NULL,
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE event_types;
-- +goose StatementEnd
