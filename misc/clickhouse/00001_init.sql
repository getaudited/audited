-- +goose Up
CREATE TABLE events (
    id String,
    source_id String,
    version UInt16,
    action String,
    actor_id String,
    actor_type String,
    actor_name String,
    actor_metadata String DEFAULT '{}',
    targets Nested (
        internal_id String,
        id String,
        name String,
        type LowCardinality(String),
        metadata String
    ),
    context_location String,
    context_user_agent String,
    metadata String DEFAULT '{}',
    occurred_at DateTime DEFAULT now()
)
--- MergeTree-family table engines are designed for high data ingest rates and huge data volumes.
--  Insert operations create table parts which are merged by a background process with other table parts.
--  https://clickhouse.com/docs/engines/table-engines/mergetree-family/mergetree
ENGINE = MergeTree()
-- Audit logs are time-series; partitioning makes range deletes and TTL efficient
PARTITION BY toYYYYMM(occurred_at)
ORDER BY (occurred_at, id);

-- +goose Down
DROP TABLE events;
