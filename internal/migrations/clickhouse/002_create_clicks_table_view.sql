CREATE MATERIALIZED VIEW IF NOT EXISTS default.analytic_clicks_view
ENGINE = SummingMergeTree((total_clicks, record_count))
ORDER BY (alias)
AS SELECT alias, sum(clicks) AS total_clicks, count() AS record_count
FROM default.analytic_clicks
GROUP BY alias;