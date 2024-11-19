package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/deadjoe/benchphant/internal/benchmark"
	"github.com/deadjoe/benchphant/internal/config"
	"github.com/deadjoe/benchphant/internal/database"
	"github.com/deadjoe/benchphant/internal/models"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	cfg             *config.Config
	manager         *database.Manager
	logger          *zap.Logger
	mu              sync.RWMutex
	activeBenchmark *benchmark.Benchmark
	server          *http.Server
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, manager *database.Manager, logger *zap.Logger) *Server {
	s := &Server{
		cfg:     cfg,
		manager: manager,
		logger:  logger,
	}

	s.server = &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
	}

	s.registerRoutes()

	return s
}

// Start starts the API server
func (s *Server) Start() error {
	s.logger.Info("Starting API server", zap.Int("port", s.cfg.Port))
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.server.Close()
}

func (s *Server) registerRoutes() {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/connections", s.handleConnections)
	mux.HandleFunc("/api/v1/connections/test", s.handleTestConnection)

	// Static files
	mux.Handle("/", http.FileServer(http.Dir("web/dist")))

	s.server.Handler = mux
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List all connections
		connections, err := s.manager.ListConnections()
		if err != nil {
			s.logger.Error("Failed to list connections", zap.Error(err))
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to list connections"})
			return
		}
		writeJSON(w, http.StatusOK, connections)

	case http.MethodPost:
		// Add new connection
		var conn models.DBConnection
		if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
			s.logger.Error("Failed to decode connection", zap.Error(err))
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid connection data"})
			return
		}

		if err := s.manager.AddConnection(&conn); err != nil {
			s.logger.Error("Failed to add connection", zap.Error(err))
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to add connection"})
			return
		}
		writeJSON(w, http.StatusCreated, conn)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var conn models.DBConnection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		s.logger.Error("Failed to decode connection", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid connection data"})
		return
	}

	if err := s.manager.TestConnection(&conn); err != nil {
		s.logger.Error("Connection test failed", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Connection test failed: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
