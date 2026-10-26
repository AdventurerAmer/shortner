CREATE TYPE token_type AS ENUM ('verification', 'passwordReset');

CREATE TABLE IF NOT EXISTS tokens (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type 		  token_type NOT NULL,
	user_id 	  TEXT NOT NULL,
    hash   		  TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
	expires_at    TIMESTAMPTZ NOT NULL,
    used_at    	  TIMESTAMPTZ NULL
);