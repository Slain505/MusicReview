package httpapi

import (
	"fmt"
	"net/http"
)

func (a *API) debugSSE(w http.ResponseWriter, r *http.Request) {
	track := r.URL.Query().Get("track")
	if track == "" {
		track = "1"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html>
  <body>
    <h3>SSE debug (track=%s)</h3>
    <pre id="log"></pre>
    <script>
      const log = (m) => document.getElementById('log').textContent += m + "\n";
      const es = new EventSource("/tracks/%s/events");
      es.onmessage = (e) => log("message: " + e.data);
      es.onerror = () => log("error (see DevTools)");
    </script>
  </body>
</html>`, track, track)
}
