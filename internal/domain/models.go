package domain

import "time"

type Track struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID          int64     `json:"id"`
	TrackID     int64     `json:"track_id"`
	Author      string    `json:"author"`
	TimestampMS int       `json:"timestamp_ms"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
}
