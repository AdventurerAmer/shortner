-- init/cassandra/001_schema.cql

CREATE KEYSPACE IF NOT EXISTS shortner
  WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
  };

USE shortner;
