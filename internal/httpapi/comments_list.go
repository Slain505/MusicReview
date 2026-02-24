package httpapi

import (
	"net/http"
	"strconv"
)

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
