-- init/cassandra/002_schema.cql

CREATE KEYSPACE IF NOT EXISTS shortner
  WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
  };

USE shortner;

CREATE TABLE IF NOT EXISTS analytic_stats (
  alias TEXT PRIMARY KEY,
  clicks COUNTER
);

CREATE TABLE analytic_stats_batches (
    batch_id TEXT PRIMARY KEY,
    applied_at TIMESTAMP
);