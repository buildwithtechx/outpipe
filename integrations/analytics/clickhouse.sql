CREATE TABLE IF NOT EXISTS default.outpipe_telemetry
(
    id String,
    organization_id String,
    signal LowCardinality(String),
    payload String,
    timestamp DateTime64(9, 'UTC')
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(timestamp)
ORDER BY (organization_id, signal, timestamp, id)
TTL timestamp + INTERVAL 30 DAY;
