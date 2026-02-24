package httpapi

import (
	"net/http"
	"os"
)

func (a *API) getTrackAudio(w http.ResponseWriter, r *http.Request) {
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

	if t.AudioPath == nil || *t.AudioPath == "" {
		http.Error(w, "audio not uploaded", http.StatusNotFound)
		return
	}

	f, err := os.Open(*t.AudioPath)
	if err != nil {
		http.Error(w, "audio missing on disk", http.StatusNotFound)
		return
	}
	defer f.Close()

	if t.AudioMIME != nil && *t.AudioMIME != "" {
		w.Header().Set("Content-Type", *t.AudioMIME)
	}

	filename := "audio"
	if t.AudioName != nil && *t.AudioName != "" {
		filename = *t.AudioName
	}

	// inline - browser will try to play if possible
	w.Header().Set("Content-Disposition", `inline; filename="`+filename+`"`)

	http.ServeContent(w, r, "audio", t.CreatedAt, f)
}
