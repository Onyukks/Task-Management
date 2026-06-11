package task

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tech4mation/tasks-api/internal/middleware"
)

// Stream is a Server-Sent Events endpoint that pushes live task changes
// (created/updated/deleted) for the authenticated user. The frontend connects
// here to reflect changes in real time without polling.
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := h.broker.Subscribe(p.UserID)
	defer unsubscribe()

	// Initial comment so the client's onopen fires immediately.
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	// Heartbeat keeps proxies from closing an idle connection.
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case ev, open := <-ch:
			if !open {
				return
			}
			payload, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, payload)
			flusher.Flush()
		}
	}
}
