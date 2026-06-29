-- init/cassandra/002_schema.cql

CREATE KEYSPACE IF NOT EXISTS shortner
  WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
  };

USE shortner;

CREATE TABLE IF NOT EXISTS analytics (
  alias text PRIMARY KEY,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  clicks bigint,
  version bigint
);

CREATE INDEX IF NOT EXISTS ON analytics (created_at);