CREATE TABLE share_links (
    token       TEXT PRIMARY KEY,
    track_id    BIGINT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_share_links_track_id ON share_links(track_id);