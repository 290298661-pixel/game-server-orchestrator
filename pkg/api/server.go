package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	httpServer  *http.Server
	handler     *Handler
	tlsCertFile string
	tlsKeyFile  string
}

func NewServer(addr string, handler *Handler, tlsCertFile, tlsKeyFile string) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/allocate", handler.Allocate)
	mux.HandleFunc("/api/v1/release", handler.Release)
	mux.HandleFunc("/api/v1/fleets/", handler.GetFleetStatus)
	mux.HandleFunc("/healthz", handler.Healthz)

	// Wrap with middleware.
	var h http.Handler = mux
	if handler.middleware != nil {
		h = handler.middleware.Wrap(mux)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      h,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		handler:     handler,
		tlsCertFile: tlsCertFile,
		tlsKeyFile:  tlsKeyFile,
	}
}

func (s *Server) Start() error {
	if s.tlsCertFile != "" && s.tlsKeyFile != "" {
		log.Printf("[OK] [api] listening on %s (TLS)", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServeTLS(s.tlsCertFile, s.tlsKeyFile); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("api server: %w", err)
		}
	} else {
		log.Printf("[OK] [api] listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("api server: %w", err)
		}
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("[OK] [api] shutting down")
	return s.httpServer.Shutdown(ctx)
}
