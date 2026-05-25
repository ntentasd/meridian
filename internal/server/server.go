package server

import (
	"context"
	"net/http"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/ntentasd/meridian/internal/store"
)

type Server struct {
	store *store.Store
	addr  string
}

func New(s *store.Store, addr string) *Server {
	return &Server{s, addr}
}

func (s *Server) Start(ctx context.Context) error {
	log := ctrl.Log.WithName("meridian-server")
	log.Info("Starting Meridian API server", "addr", s.addr)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/routes", s.handleRoutes)
	mux.HandleFunc("/api/events", s.handleEvents)

	srv := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
