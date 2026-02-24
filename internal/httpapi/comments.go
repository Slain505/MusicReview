package httpapi

import (
	"MusicReview/internal/sse"
	"net/http"
)

type createCommentReq struct {
	Author      string `json:"author"`
	TimestampMS int    `json:"timestamp_ms"`
	Text        string `json:"text"`
}

func (a *API) createComment(w http.ResponseWriter, r *http.Request) {
	trackID, err := idParam(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req createCommentReq
	if err := decodeJSON(r, &req); err != nil || req.Author == "" || req.Text == "" || req.TimestampMS < 0 {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	track, err := a.Store.GetTrack(r.Context(), trackID)
	if err != nil {
		if isNotFound(err) {
			http.Error(w, "track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load track", http.StatusInternalServerError)
		return
	}

	// If duration is known, validate that timestamp_ms is within the track length.
	if track.DurationMS != nil && req.TimestampMS > *track.DurationMS {
		http.Error(w, "timestamp_ms exceeds track duration", http.StatusBadRequest)
		return
	}

	c, err := a.Store.CreateComment(r.Context(), trackID, req.Author, req.TimestampMS, req.Text)
	if err != nil {
		http.Error(w, "failed to create comment", http.StatusInternalServerError)
		return
	}

	a.Hub.Publish(trackID, sse.Event{
		Type: "comment.created",
		Data: c,
	})

	writeJSON(w, http.StatusCreated, c)
}
