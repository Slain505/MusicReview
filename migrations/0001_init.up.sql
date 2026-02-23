CREATE TABLE tracks (
                        id            BIGSERIAL PRIMARY KEY,
                        title         TEXT NOT NULL,
                        created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE comments (
                          id            BIGSERIAL PRIMARY KEY,
                          track_id      BIGINT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
                          author        TEXT NOT NULL,
                          timestamp_ms  INTEGER NOT NULL CHECK (timestamp_ms >= 0),
                          text          TEXT NOT NULL,
                          created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_comments_track_created ON comments(track_id, created_at);