CREATE TABLE IF NOT EXISTS default.analytics_view_target (
	alias String,
	clicks UInt64
) ENGINE = AggregatingMergeTree()
ORDER BY (alias);

CREATE MATERIALIZED VIEW IF NOT EXISTS default.analytics_view 
TO analytics_view_target AS 
SELECT  
	alias,
	count() AS clicks
FROM default.analytics
GROUP BY (alias);