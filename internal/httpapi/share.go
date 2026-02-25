package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (a *API) createShare(w http.ResponseWriter, r *http.Request) {
	trackID, err := idParam(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if _, err := a.Store.GetTrack(r.Context(), trackID); err != nil {
		if isNotFound(err) {
			http.Error(w, "track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load track", http.StatusInternalServerError)
		return
	}

	token, err := newToken(24)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	var expiresAt *time.Time = nil

	sl, err := a.Store.CreateShareLink(r.Context(), token, trackID, expiresAt)
	if err != nil {
		http.Error(w, "failed to create share link", http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/share/%s", scheme, r.Host, sl.Token)

	writeJSON(w, http.StatusCreated, map[string]any{
		"token": sl.Token,
		"url":   url,
	})
}

func newToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (a *API) getShareJSON(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	sl, err := a.Store.GetShareLink(r.Context(), token)
	if err != nil {
		if isNotFound(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get share link", http.StatusInternalServerError)
		return
	}

	if sl.ExpiresAt != nil && time.Now().After(*sl.ExpiresAt) {
		http.Error(w, "link expired", http.StatusGone)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token":      sl.Token,
		"track_id":   sl.TrackID,
		"expires_at": sl.ExpiresAt,
	})
}

func (a *API) sharePage(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	sl, err := a.Store.GetShareLink(r.Context(), token)
	if err != nil {
		if isNotFound(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get share link", http.StatusInternalServerError)
		return
	}

	// exp check (for future, even if nil)
	if sl.ExpiresAt != nil && time.Now().After(*sl.ExpiresAt) {
		http.Error(w, "link expired", http.StatusGone)
		return
	}

	track, err := a.Store.GetTrack(r.Context(), sl.TrackID)
	if err != nil {
		http.Error(w, "failed to get track", http.StatusInternalServerError)
		return
	}

	// HTML (MVP)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintf(w, `<!doctype html>
<html>
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>%s</title>
  <style>
    body { font-family: system-ui, -apple-system, sans-serif; max-width: 860px; margin: 24px auto; padding: 0 14px; }
    .row { display:flex; gap:10px; align-items:center; flex-wrap: wrap; }
    input, button { padding: 10px 12px; border-radius: 10px; border:1px solid #ddd; }
    button { cursor:pointer; }
    #list { margin-top: 12px; }
    .c { padding:10px 0; border-bottom:1px solid #eee; cursor:pointer; }
    .meta { color:#666; font-size: 12px; margin-bottom: 3px; }
    #log { white-space: pre-wrap; background:#f7f7f7; padding:12px; border-radius: 12px; }
  </style>
</head>
<body>
  <h2>%s</h2>

  <audio id="player" controls style="width: 100%%" src="/tracks/%d/audio"></audio>

  <h3>Add comment</h3>
  <div class="row">
    <input id="author" placeholder="name" value="guest" />
    <input id="text" placeholder="comment" style="min-width: 320px;" />
    <button id="send">Send at current time</button>
  </div>

  <h3>Comments (click to jump)</h3>
  <div id="list"></div>

  <h3>Live events</h3>
  <pre id="log"></pre>

<script>
const trackId = %d;
const player = document.getElementById("player");
const list = document.getElementById("list");
const logEl = document.getElementById("log");

function log(m){ logEl.textContent += m + "\n"; }

function msToTime(ms) {
  const s = Math.floor(ms/1000);
  const m = Math.floor(s/60);
  const ss = String(s%%60).padStart(2,'0');
  return m + ":" + ss;
}

function addCommentToUI(c) {
  const el = document.createElement("div");
  el.className = "c";
  el.innerHTML = '<div class="meta">' + msToTime(c.timestamp_ms) + ' — ' + (c.author||'') + '</div>' +
                 '<div>' + (c.text||'') + '</div>';
  el.onclick = () => { player.currentTime = (c.timestamp_ms||0)/1000.0; player.play(); };
  list.appendChild(el);
}

async function loadComments() {
  const res = await fetch("/tracks/" + trackId + "/comments");
  const items = await res.json();
  list.innerHTML = "";
  for (const c of items) addCommentToUI(c);
}

document.getElementById("send").onclick = async () => {
  const author = document.getElementById("author").value || "guest";
  const text = document.getElementById("text").value || "";
  const timestamp_ms = Math.floor(player.currentTime * 1000);

  const res = await fetch("/tracks/" + trackId + "/comments", {
    method: "POST",
    headers: {"Content-Type":"application/json"},
    body: JSON.stringify({author, text, timestamp_ms})
  });

  if (!res.ok) {
    log("POST failed: " + res.status);
    return;
  }

  document.getElementById("text").value = "";
};

const es = new EventSource("/tracks/" + trackId + "/events");
es.onmessage = (e) => {
  try {
    const msg = JSON.parse(e.data);
    if (msg.type === "comment.created") addCommentToUI(msg.data);
    log("event: " + e.data);
  } catch {
    log("event(raw): " + e.data);
  }
};

loadComments();
</script>
</body>
</html>`, track.Title, track.Title, track.ID, track.ID)
}
