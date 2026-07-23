CREATE TABLE IF NOT EXISTS shortner.url_mappings (
  alias text PRIMARY KEY,
  long_url text,
  created_at TIMESTAMP,
  user_id  text
);

CREATE INDEX IF NOT EXISTS ON shortner.url_mappings (created_at);
CREATE INDEX IF NOT EXISTS ON shortner.url_mappings (user_id);