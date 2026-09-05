-- status now stores user intent only. 'scheduled' and 'expired' are derived
-- from starts_at/expires_at at read time, so they must never be stored.
UPDATE short_urls SET status = 'active' WHERE status IN ('scheduled', 'expired');

ALTER TABLE short_urls DROP CONSTRAINT short_urls_status_check;
ALTER TABLE short_urls ADD CONSTRAINT short_urls_status_check CHECK (status IN ('active', 'disabled'));
