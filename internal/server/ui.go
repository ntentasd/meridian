package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ntentasd/meridian/internal/store"
	"github.com/ntentasd/meridian/web/templates"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.Index(s.store.List()).Render(r.Context(), w)
}

func (s *Server) handleUIEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := s.store.Subscribe()
	defer s.store.Unsubscribe(ch)

	if err := sendRowsEvent(r.Context(), w, s.store.List()); err != nil {
		return
	}
	flusher.Flush()

	for {
		select {
		case <-ch:
			if err := sendRowsEvent(r.Context(), w, s.store.List()); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func sendRowsEvent(ctx context.Context, w http.ResponseWriter, entries []store.RouteEntry) error {
	var buf bytes.Buffer
	if err := templates.Cards(entries).Render(ctx, &buf); err != nil {
		return err
	}
	html := strings.ReplaceAll(buf.String(), "\n", "")
	_, err := fmt.Fprintf(w, "event: update\ndata: %s\n\n", html)
	return err
}
