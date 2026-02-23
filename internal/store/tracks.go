package store

import (
	"MusicReview/internal/domain"
	"context"
)

func (s *Store) CreateTrack(ctx context.Context, title string) (domain.Track, error) {
	var t domain.Track
	err := s.DB.QueryRow(ctx,
		`INSERT INTO tracks(title) VALUES($1) RETURNING id, title, created_at`,
		title,
	).Scan(&t.ID, &t.Title, &t.CreatedAt)

	return t, err
}

func (s *Store) GetTrack(ctx context.Context, id int64) (domain.Track, error) {
	var t domain.Track
	err := s.DB.QueryRow(ctx,
		`SELECT id, title, created_at FROM tracks WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Title, &t.CreatedAt)

	return t, err
}
