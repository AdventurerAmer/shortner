-- init/cassandra/001_schema.cql

CREATE KEYSPACE IF NOT EXISTS shortner
  WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
  };

USE shortner;

CREATE TABLE IF NOT EXISTS url_mappings (
  alias text PRIMARY KEY,
  long_url text,
  created_at TIMESTAMP,
  user_id  text
);

CREATE INDEX IF NOT EXISTS ON url_mappings (created_at);
CREATE INDEX IF NOT EXISTS ON url_mappings (user_id);