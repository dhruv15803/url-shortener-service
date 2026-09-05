CREATE TABLE IF NOT EXISTS clicks (
    id SERIAL PRIMARY KEY,
    short_url_id INT NOT NULL REFERENCES short_urls(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip INET,
    user_agent TEXT,
    referrer TEXT,
    country VARCHAR(2),
    device VARCHAR(64),
    browser VARCHAR(64)
);

CREATE INDEX IF NOT EXISTS clicks_short_url_id_idx ON clicks (short_url_id);
