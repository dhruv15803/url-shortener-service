ALTER TABLE short_urls DROP CONSTRAINT short_urls_status_check;
ALTER TABLE short_urls ADD CONSTRAINT short_urls_status_check CHECK (status IN ('scheduled', 'active', 'disabled', 'expired'));
