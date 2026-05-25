package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ntentasd/meridian/internal/store"
)

func (s *Server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.store.List())
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := s.store.Subscribe()
	defer s.store.Unsubscribe(ch)

	// push full state immediately on connect
	for _, entry := range s.store.List() {
		writeEvent(w, store.Event{Kind: store.EventUpsert, Entry: entry})
	}
	flusher.Flush()

	for {
		select {
		case ev := <-ch:
			writeEvent(w, ev)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func writeEvent(w http.ResponseWriter, ev store.Event) {
	data, _ := json.Marshal(ev)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Kind, data)
}
