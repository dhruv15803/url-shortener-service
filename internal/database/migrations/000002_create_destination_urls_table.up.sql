CREATE TABLE IF NOT EXISTS destination_urls (
    id SERIAL PRIMARY KEY,
    destination_url TEXT NOT NULL,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, destination_url)
);

CREATE INDEX IF NOT EXISTS destination_urls_user_id_idx ON destination_urls (user_id);
