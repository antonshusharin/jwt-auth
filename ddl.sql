CREATE TABLE IF NOT EXISTS users (
    guid uuid PRIMARY KEY,
    username text,
    email text
);

CREATE TABLE IF NOT EXISTS refresh_tokens (refresh_id uuid PRIMARY KEY, hash text);