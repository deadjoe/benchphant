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
	"github.com/gin-gonic/gin"
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

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	router := gin.Default()

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/connections", gin.WrapF(s.handleConnections))
		v1.POST("/connections", gin.WrapF(s.handleConnections))
		v1.POST("/connections/test", gin.WrapF(s.handleTestConnection))
	}

	// Static files
	router.Static("/", "web/dist")

	s.server.Handler = router
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
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	var config models.DBConnection
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := config.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
