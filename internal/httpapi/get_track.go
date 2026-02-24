package httpapi

import "net/http"

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
