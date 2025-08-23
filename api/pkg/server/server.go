package server

import (
	"api/pkg/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Server represents the API server with its configuration.
type Server struct {
	hostname string
	key      string
}

// New creates a new Server instance with the provided configuration.
func New(cfg *config.Config) *Server {
	return &Server{
		hostname: fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port),
		key:      cfg.Key,
	}
}

// Run starts the HTTP server and listens for incoming requests.
func (s *Server) Run(ctx context.Context) error {
	srv := http.Server{
		Addr:    s.hostname,
		Handler: s,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)

}

// ServeHTTP is the main handler that processes incoming HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, respUnauthorized, http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, respUnauthorized, http.StatusUnauthorized)
		return
	}

	if !s.validateAuthorization(authHeader) {
		http.Error(w, respUnauthorized, http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	// write response body
	fmt.Fprint(w, respSuccess)

}

func (s *Server) validateAuthorization(authHeader string) bool {
	bearerVal := strings.TrimPrefix(authHeader, "Bearer ")
	if bearerVal == "" {
		return false
	}

	parts := strings.Split(bearerVal, ":")

	if len(parts) < 2 {
		return false
	}

	return parts[1] == s.key
}
