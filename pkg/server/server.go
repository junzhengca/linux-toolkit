package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	port       int
	bind       string
	uiEnabled  bool
	interval   int
}

// NewServer creates a new server instance
func NewServer(port int, bind string, uiEnabled bool, interval int) *Server {
	return &Server{
		port:       port,
		bind:       bind,
		uiEnabled:  uiEnabled,
		interval:   interval,
		httpServer: nil,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.bind, s.port)
	mux := http.NewServeMux()

	// Register routes
	s.registerRoutes(mux, s.uiEnabled)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	return s.startWithGracefulShutdown()
}

// startWithGracefulShutdown starts the server with graceful shutdown handling
func (s *Server) startWithGracefulShutdown() error {
	// Channel to listen for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Channel to capture server errors
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s (UI enabled: %v, refresh interval: %ds)", s.httpServer.Addr, s.uiEnabled, s.interval)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for either shutdown signal or server error
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		return s.Stop()
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Printf("Server stopped gracefully")
	return nil
}
