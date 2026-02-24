package store

import (
	"MusicReview/internal/domain"
	"context"
	"encoding/json"
)

func (s *Store) CreateTrack(ctx context.Context, title string, audioPath, audioMIME, audioName *string) (domain.Track, error) {
	var t domain.Track
	err := s.DB.QueryRow(ctx,
		`INSERT INTO tracks(title, audio_path, audio_mime, audio_name)
		 VALUES($1, $2, $3, $4)
		 RETURNING id, title, audio_path, audio_mime, audio_name, created_at`,
		title, audioPath, audioMIME, audioName,
	).Scan(&t.ID, &t.Title, &t.AudioPath, &t.AudioMIME, &t.AudioName, &t.CreatedAt)
	return t, err
}

func (s *Store) GetTrack(ctx context.Context, id int64) (domain.Track, error) {
	var t domain.Track
	var peaksJSON []byte // NULL -> nil
	err := s.DB.QueryRow(ctx,
		`SELECT id, title, audio_path, audio_mime, audio_name, duration_ms, waveform_peaks, created_at
     FROM tracks WHERE id=$1`,
		id,
	).Scan(&t.ID, &t.Title, &t.AudioPath, &t.AudioMIME, &t.AudioName, &t.DurationMS, &peaksJSON, &t.CreatedAt)

	if len(peaksJSON) > 0 {
		_ = json.Unmarshal(peaksJSON, &t.WaveformPeaks)
	}

	return t, err
}

func (s *Store) SetTrackAudioMeta(ctx context.Context, id int64, path, mime, name string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE tracks SET audio_path=$2, audio_mime=$3, audio_name=$4 WHERE id=$1`,
		id, path, mime, name,
	)
	return err
}

func (s *Store) SetTrackAnalysis(ctx context.Context, id int64, durationMS int, peaks []int) error {
	b, err := json.Marshal(peaks)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx,
		`UPDATE tracks SET duration_ms=$2, waveform_peaks=$3 WHERE id=$1`,
		id, durationMS, b,
	)
	return err
}
