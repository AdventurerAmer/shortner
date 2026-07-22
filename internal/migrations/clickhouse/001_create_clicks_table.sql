CREATE TABLE IF NOT EXISTS default.analytic_clicks (
    id         String,
    alias      String,           
    clicks     UInt64,           
    created_at DateTime DEFAULT now(),
) 
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(created_at)
ORDER BY (alias, id);