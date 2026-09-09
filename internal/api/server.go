package api

import (
	"context"
	"net/http"
	"time"

	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

// Server holds the dependencies handler needs. Handlers are methods on Server so we don't have to use globals.
type Server struct {
	store *persistence.Store
}

// Returns a *Server with the provided store
func New(store *persistence.Store) *Server {
	return &Server{store: store}
}

// httpServer wraps http.Server with timeouts built from a *Server's routes.
type httpServer struct {
	inner *http.Server
}

// Returns a *httpServer with the provided address and *Server's routes + timeouts.
func NewHTTPServer(addr string, s *Server) *httpServer {
	return &httpServer{
		inner: &http.Server{
			Addr:              addr,
			Handler:           s.routes(),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

// Start blocks until the server stops(returns http.ErrServerClosed on shutdown)
func (h *httpServer) Start() error {
	return h.inner.ListenAndServe()
}

// Shutdown stops accepting new connections and waits for in-flight requests to finish
func (h *httpServer) Shutdown(ctx context.Context) error {
	return h.inner.Shutdown(ctx)
}
