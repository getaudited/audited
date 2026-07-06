-- +goose Up
CREATE TABLE users (
    id String,
    email String,
    password String,
    role LowCardinality(String),
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY id;

CREATE TABLE sources (
    id String,
    name String,
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY id;

CREATE TABLE event_types (
    action String,
    should_validate_metadata_schema Bool DEFAULT false,
    version UInt16,
    schema String,
    target_types Array(LowCardinality(String)),
    created_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(created_at)
ORDER BY (action, version);

CREATE TABLE events (
    id String CODEC(ZSTD(1)),
    source_id String CODEC(ZSTD(1)),
    version UInt16,
    action LowCardinality(String),
    actor_id String CODEC(ZSTD(1)),
    actor_type LowCardinality(String),
    actor_name String,
    actor_metadata String DEFAULT '{}' CODEC(ZSTD(3)),
    targets Nested (
        internal_id String,
        id String,
        name String,
        type LowCardinality(String),
        metadata String
    ),
    context_location LowCardinality(String),
    context_user_agent LowCardinality(String),
    metadata String DEFAULT '{}' CODEC(ZSTD(3)),
    occurred_at DateTime  CODEC(Delta, ZSTD(1)),

    -- exact-match filter on actor_id → skip index prunes granules
    INDEX idx_actor_id actor_id TYPE bloom_filter GRANULARITY 4,
    -- WHERE id = ? without source_id → minmax works because id is time-ordered
    INDEX idx_id id TYPE minmax GRANULARITY 4
)
--- MergeTree-family table engines are designed for high data ingest rates and huge data volumes.
--  Insert operations create table parts which are merged by a background process with other table parts.
--  https://clickhouse.com/docs/engines/table-engines/mergetree-family/mergetree
ENGINE = MergeTree()
-- Audit logs are time-series; partitioning makes range deletes and TTL efficient
PARTITION BY toYYYYMM(occurred_at)
ORDER BY (source_id, id);

CREATE TABLE tokens (
    id String,
    name String,
    value String,
    source_id String,
    created_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(created_at)
ORDER BY value;

-- +goose Down
DROP TABLE tokens;
DROP TABLE users;
DROP TABLE event_types;
DROP TABLE sources;
DROP TABLE events;
