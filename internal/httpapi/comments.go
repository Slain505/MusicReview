package httpapi

import (
	"net/http"
	"strconv"

	"MusicReview/internal/sse"
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

	// (необязательно) проверить, что трек существует
	if _, err := a.Store.GetTrack(r.Context(), trackID); err != nil {
		if isNotFound(err) {
			http.Error(w, "track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to validate track", http.StatusInternalServerError)
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

func (a *API) listComments(w http.ResponseWriter, r *http.Request) {
	trackID, err := idParam(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}

	items, err := a.Store.ListComments(r.Context(), trackID, limit)
	if err != nil {
		http.Error(w, "failed to list comments", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
