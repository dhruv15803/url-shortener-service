CREATE TABLE IF NOT EXISTS short_urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(32) NOT NULL UNIQUE,
    destination_id INT NOT NULL REFERENCES destination_urls(id) ON DELETE CASCADE,
    name VARCHAR(255),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('scheduled', 'active', 'disabled', 'expired')),
    starts_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS short_urls_destination_id_idx ON short_urls (destination_id);
