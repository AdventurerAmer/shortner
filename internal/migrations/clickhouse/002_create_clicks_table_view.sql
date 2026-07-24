CREATE TABLE IF NOT EXISTS default.analytic_clicks_view_target (
    unique_id String,
	alias String,
	total_clicks UInt64,
	record_count UInt64 
) ENGINE = AggregatingMergeTree()
ORDER BY (unique_id, alias);

CREATE MATERIALIZED VIEW IF NOT EXISTS default.analytic_clicks_view 
TO analytic_clicks_view_target AS 
SELECT 
    uniqState(id) as unique_id, 
	alias,
	sum(clicks) AS total_clicks,
	count() AS record_count
FROM default.analytic_clicks
GROUP BY (id, alias);