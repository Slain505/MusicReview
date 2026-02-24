package domain

import "time"

type Track struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	AudioPath *string `json:"audio_path,omitempty"`
	AudioMIME *string `json:"audio_mime,omitempty"`
	AudioName *string `json:"audio_name,omitempty"`

	DurationMS    *int  `json:"duration_ms,omitempty"`
	WaveformPeaks []int `json:"waveform_peaks,omitempty"`

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
