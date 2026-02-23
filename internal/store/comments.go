package store

import (
	"MusicReview/internal/domain"
	"context"
)

func (s *Store) CreateComment(ctx context.Context, trackID int64, author string,
	timestampMS int, text string) (domain.Comment, error) {
	var c domain.Comment
	err := s.DB.QueryRow(ctx,
		`INSERT INTO comments(track_id, author, timestamp_ms, text)
			VALUES($1,$2,$3,$4)
		 	RETURNING id, track_id, author, timestamp_ms, text, created_at`,
		trackID, author, timestampMS, text,
	).Scan(&c.ID, &c.TrackID, &c.Author, &c.TimestampMS, &c.Text, &c.CreatedAt)

	return c, err
}

func (s *Store) ListComments(ctx context.Context, trackID int64, limit int) ([]domain.Comment, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.DB.Query(ctx,
		`SELECT id, track_id, author, timestamp_ms, text, created_at
		 FROM comments
		 WHERE track_id=$1
		 ORDER BY created_at ASC
		 LIMIT $2`,
		trackID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Comment, 0, limit)
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.TrackID, &c.Author, &c.TimestampMS, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
