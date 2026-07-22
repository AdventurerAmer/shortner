CREATE TABLE IF NOT EXISTS default.analytic_clicks (
    id         String,
    alias      String,           
    clicks     UInt32,           
    created_at DateTime DEFAULT now(),
    _version UInt64 DEFAULT 1
) 
ENGINE = ReplacingMergeTree(_version)
ORDER BY (alias, id)
PARTITION BY toYYYYMM(created_at);