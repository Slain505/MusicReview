package httpapi

import (
	"MusicReview/internal/sse"
	"MusicReview/internal/store"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type API struct {
	Store *store.Store
	Hub   *sse.Hub
}

func New(store *store.Store, hub *sse.Hub) *API {
	return &API{Store: store, Hub: hub}
}

func (a *API) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", a.healthz)

	r.Post("/tracks", a.createTrack)
	r.Get("/tracks/{id}", a.getTrack)

	r.Post("/tracks/{id}/comments", a.createComment)
	r.Get("/tracks/{id}/comments", a.listComments)

	r.Get("/tracks/{id}/events", a.trackEvents)

	r.Get("/debug/sse", a.debugSSE)

	return r
}

func (a *API) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	if err := a.Store.Ping(ctx); err != nil {
		http.Error(w, "db not ok", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func idParam(r *http.Request) (int64, error) {
	raw := chi.URLParam(r, "id")
	return strconv.ParseInt(raw, 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
