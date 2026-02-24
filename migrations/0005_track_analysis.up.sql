ALTER TABLE tracks
    ADD COLUMN duration_ms INTEGER,
    ADD COLUMN waveform_peaks JSONB;