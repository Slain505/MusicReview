package store

import (
	"context"
	"encoding/json"

	"MusicReview/internal/domain"
)

func (s *Store) ListTracks(ctx context.Context, limit int) ([]domain.Track, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.DB.Query(ctx,
		`SELECT id, title, audio_path, audio_mime, audio_name, duration_ms, waveform_peaks, created_at
		FROM tracks
		ORDER BY created_at DESC
		LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Track, 0, limit)
	for rows.Next() {
		var t domain.Track
		var peaksJSON []byte
		if err := rows.Scan(&t.ID, &t.Title, &t.AudioPath, &t.AudioMIME, &t.AudioName, &t.DurationMS, &peaksJSON, &t.CreatedAt); err != nil {
			return nil, err
		}
		if len(peaksJSON) > 0 {
			_ = json.Unmarshal(peaksJSON, &t.WaveformPeaks)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
