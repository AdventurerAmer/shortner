CREATE TABLE IF NOT EXISTS default.analytics (
    event_id   String,
    alias      String,                      
    created_at DateTime DEFAULT now(),
) 
ENGINE = ReplacingMergeTree
PRIMARY KEY (event_id, alias)
PARTITION BY toYYYYMM(created_at)
ORDER BY (event_id, alias);