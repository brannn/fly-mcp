package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/config"
	"github.com/brannn/fly-mcp/pkg/mcp"
)

// Server represents the MCP server
type Server struct {
	config     *config.Config
	logger     *logger.Logger
	mcpHandler *mcp.Handler
	httpServer *http.Server
	router     *mux.Router
}

// New creates a new server instance
func New(cfg *config.Config, log *logger.Logger) (*Server, error) {
	// Create MCP handler
	mcpHandler, err := mcp.NewHandler(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP handler: %w", err)
	}
	
	// Create router
	router := mux.NewRouter()
	
	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
	
	server := &Server{
		config:     cfg,
		logger:     log,
		mcpHandler: mcpHandler,
		httpServer: httpServer,
		router:     router,
	}
	
	// Setup routes
	server.setupRoutes()
	
	return server, nil
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info().
		Str("address", s.httpServer.Addr).
		Msg("Starting HTTP server")
	
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()
	
	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down server")
	
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	
	return nil
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	// Metrics endpoint (if enabled)
	s.router.HandleFunc("/metrics", s.handleMetrics).Methods("GET")
	
	// MCP endpoint - this is where MCP clients will connect
	s.router.HandleFunc("/mcp", s.handleMCP).Methods("POST")
	
	// Add middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.corsMiddleware)
	
	if s.config.Security.RateLimitEnabled {
		s.router.Use(s.rateLimitMiddleware)
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status":      "healthy",
		"timestamp":   time.Now().UTC(),
		"version":     s.config.MCP.ServerInfo.Version,
		"environment": s.config.Environment,
	}
	
	if err := writeJSON(w, response); err != nil {
		s.logger.Error().Err(err).Msg("Failed to write health check response")
	}
}

// handleMetrics handles metrics requests
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	
	// TODO: Implement actual metrics collection
	fmt.Fprintf(w, "# HELP fly_mcp_requests_total Total number of MCP requests\n")
	fmt.Fprintf(w, "# TYPE fly_mcp_requests_total counter\n")
	fmt.Fprintf(w, "fly_mcp_requests_total 0\n")
}

// handleMCP handles MCP protocol requests
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Handle the MCP request
	if err := s.mcpHandler.HandleRequest(w, r); err != nil {
		s.logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("MCP request failed")
		
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	s.logger.Debug().
		Dur("duration", time.Since(start)).
		Msg("MCP request completed")
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	// TODO: Implement JSON encoding
	return nil
}
