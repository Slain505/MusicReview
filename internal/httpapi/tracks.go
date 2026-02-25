package httpapi

import (
	"MusicReview/internal/audioanalyzer"
	"MusicReview/internal/sse"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"MusicReview/internal/domain"
)

type createTrackReq struct {
	Title string `json:"title"`
}

func (a *API) createTrack(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")

	// 1) multipart upload
	if strings.HasPrefix(ct, "multipart/form-data") {
		a.createTrackMultipart(w, r)
		return
	}

	// 2) JSON (old path)
	var req createTrackReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		http.Error(w, "invalid json or title", http.StatusBadRequest)
		return
	}

	t, err := a.Store.CreateTrack(r.Context(), req.Title, nil, nil, nil)
	if err != nil {
		http.Error(w, "failed to create track", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (a *API) createTrackMultipart(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB
		http.Error(w, "expected multipart/form-data", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "audio file is required (field name: audio)", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// create track without audio_path
	t, err := a.Store.CreateTrack(r.Context(), title, nil, nil, nil)
	if err != nil {
		http.Error(w, "failed to create track", http.StatusInternalServerError)
		return
	}

	// save
	dir := fmt.Sprintf("storage/tracks/%d", t.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		http.Error(w, "failed to create storage dir", http.StatusInternalServerError)
		return
	}

	name := sanitizeFilename(header.Filename)
	dstPath := fmt.Sprintf("%s/original_%s", dir, name)

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	// write path to db
	mime := detectMime(dstPath, header.Header.Get("Content-Type"))

	ext := strings.ToLower(filepath.Ext(name))
	if mime == "application/octet-stream" {
		switch ext {
		case ".mp3":
			mime = "audio/mpeg"
		case ".wav":
			mime = "audio/wav"
		case ".flac":
			mime = "audio/flac"
		case ".ogg":
			mime = "audio/ogg"
		case ".m4a":
			mime = "audio/mp4"
		}
	}

	if err := a.Store.SetTrackAudioMeta(r.Context(), t.ID, dstPath, mime, name); err != nil {
		http.Error(w, "failed to set audio meta", http.StatusInternalServerError)
		return
	}

	go func(trackID int64, path string) {
		ctx := context.Background()

		res, err := audioanalyzer.Analyze(ctx, path, 1000)
		if err != nil {
			return
		}

		if err := a.Store.SetTrackAnalysis(ctx, trackID, res.DurationMS, res.Peaks); err != nil {
			return
		}

		// Notify clients that analysis is ready (clients can refetch track JSON).
		a.Hub.Publish(trackID, sse.Event{
			Type: "track.analyzed",
			Data: map[string]any{
				"track_id": trackID,
			},
		})
	}(t.ID, dstPath)

	// return reload track
	t, err = a.Store.GetTrack(r.Context(), t.ID)
	if err != nil {
		http.Error(w, "failed to reload track", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return name
}

func detectMime(path string, fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}

	f, err := os.Open(path)
	if err != nil {
		return "application/octet-stream"
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	return http.DetectContentType(buf[:n])
}

var _ domain.Track
