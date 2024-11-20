package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark"
	"go.uber.org/zap"
)

// BenchmarkRequest represents a request to start a benchmark
type BenchmarkRequest struct {
	ConnectionID int64     `json:"connection_id"`
	Duration     string    `json:"duration"`
	Concurrency  int       `json:"concurrency"`
	QueryRate    int       `json:"query_rate"`
	Queries      []string  `json:"queries"`
	Distribution string    `json:"distribution"`
	QueryWeights []float64 `json:"query_weights,omitempty"`
}

// handleBenchmarkStart handles starting a benchmark
func (s *Server) handleBenchmarkStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BenchmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode request", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request data"})
		return
	}

	// Validate request
	if req.ConnectionID == 0 || len(req.Queries) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Missing required fields"})
		return
	}

	// Get connection
	conn, err := s.manager.GetConnection(req.ConnectionID)
	if err != nil {
		s.logger.Error("Failed to get connection", zap.Error(err))
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "Connection not found"})
		return
	}

	// Parse duration
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		s.logger.Error("Invalid duration", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid duration"})
		return
	}

	// Create benchmark config
	config := &benchmark.Config{
		Duration:          duration,
		Concurrency:       req.Concurrency,
		QueryRate:         req.QueryRate,
		Queries:           req.Queries,
		QueryDistribution: benchmark.QueryDistributionType(req.Distribution),
		QueryWeights:      req.QueryWeights,
	}

	// Create and start benchmark
	b, err := benchmark.New(config, conn, s.logger)
	if err != nil {
		s.logger.Error("Failed to create benchmark", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to create benchmark"})
		return
	}

	if err := b.Start(r.Context()); err != nil {
		s.logger.Error("Failed to start benchmark", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to start benchmark"})
		return
	}

	// Store benchmark for status checks
	s.mu.Lock()
	s.activeBenchmark = b
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// handleBenchmarkStop handles stopping a benchmark
func (s *Server) handleBenchmarkStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	b := s.activeBenchmark
	s.mu.Unlock()

	if b == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "No active benchmark"})
		return
	}

	b.Stop() // Stop() doesn't return an error
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// handleBenchmarkStatus handles getting benchmark status
func (s *Server) handleBenchmarkStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	b := s.activeBenchmark
	s.mu.RUnlock()

	if b == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "No active benchmark"})
		return
	}

	status := struct {
		Status benchmark.BenchmarkStatus `json:"status"`
		Result *benchmark.Result        `json:"result,omitempty"`
	}{
		Status: b.Status(),
		Result: b.Result(),
	}

	writeJSON(w, http.StatusOK, status)
}
