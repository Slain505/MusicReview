package httpapi

import (
	"net/http"

	"MusicReview/internal/domain"
)

type createTrackReq struct {
	Title string `json:"title"`
}

func (a *API) createTrack(w http.ResponseWriter, r *http.Request) {
	var req createTrackReq
	if err := decodeJSON(r, &req); err != nil || req.Title == "" {
		http.Error(w, "invalid json or title", http.StatusBadRequest)
		return
	}

	t, err := a.Store.CreateTrack(r.Context(), req.Title)
	if err != nil {
		http.Error(w, "failed to create track", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (a *API) getTrack(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	t, err := a.Store.GetTrack(r.Context(), id)
	if err != nil {
		if isNotFound(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get track", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// чтобы компилятор видел импорт domain (иначе можно убрать)
var _ domain.Track
