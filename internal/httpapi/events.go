package httpapi

import (
	"net/http"
	"time"

	"MusicReview/internal/sse"
)

func (a *API) trackEvents(w http.ResponseWriter, r *http.Request) {
	trackID, err := idParam(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch, unsubscribe := a.Hub.Subscribe(trackID)
	defer unsubscribe()

	// “комментарий”: периодический ping, чтобы соединение не тухло у прокси
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// сразу отправим hello
	_ = sse.WriteSSE(w, []byte(`{"type":"hello"}`))
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case b := <-ch:
			_ = sse.WriteSSE(w, b)
			flusher.Flush()
		case <-ticker.C:
			_, _ = w.Write([]byte(": ping\n\n")) // SSE comment
			flusher.Flush()
		}
	}
}
