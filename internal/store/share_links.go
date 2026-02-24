package store

import (
	"context"
	"time"
)

type ShareLink struct {
	Token     string     `json:"token"`
	TrackID   int64      `json:"track_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (s *Store) CreateShareLink(ctx context.Context, token string, trackID int64, expiresAt *time.Time) (ShareLink, error) {
	var sl ShareLink
	err := s.DB.QueryRow(ctx,
		`INSERT INTO share_links(token, track_id, expires_at)
		 	VALUES($1,$2,$3)
			RETURNING token, track_id, expires_at, created_at`,
		token, trackID, expiresAt,
	).Scan(&sl.Token, &sl.TrackID, &sl.ExpiresAt, &sl.CreatedAt)
	return sl, err
}

func (s *Store) GetShareLink(ctx context.Context, token string) (ShareLink, error) {
	var sl ShareLink
	err := s.DB.QueryRow(ctx,
		`SELECT token, track_id, expires_at, created_at
		 FROM share_links WHERE token=$1`,
		token,
	).Scan(&sl.Token, &sl.TrackID, &sl.ExpiresAt, &sl.CreatedAt)
	return sl, err
}
