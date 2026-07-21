CREATE TABLE IF NOT EXISTS default.analytic_clicks (
    id         String,
    alias      String,           
    clicks     UInt32,           
    created_at DateTime DEFAULT now()
) 
ENGINE = MergeTree()
ORDER BY (alias)
PARTITION BY toYYYYMM(created_at)
SETTINGS index_granularity = 8192;