CREATE TABLE IF NOT EXISTS users (
    guid uuid PRIMARY KEY,
    username text,
    email text
);

CREATE TABLE IF NOT EXISTS refresh_tokens (hash text PRIMARY KEY);