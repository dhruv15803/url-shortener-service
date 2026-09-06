-- Analytics filters on (short_url_id, clicked_at). With only short_url_id
-- indexed, every click ever recorded for a campaign is read and then filtered
-- by time; this makes the range an index scan instead.
CREATE INDEX IF NOT EXISTS clicks_short_url_id_clicked_at_idx
    ON clicks (short_url_id, clicked_at DESC);
