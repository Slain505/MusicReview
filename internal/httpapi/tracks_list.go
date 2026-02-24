package httpapi

import (
	"net/http"
	"strconv"
)

func (a *API) listTracks(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}

	items, err := a.Store.ListTracks(r.Context(), limit)
	if err != nil {
		http.Error(w, "failed to list tracks", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
