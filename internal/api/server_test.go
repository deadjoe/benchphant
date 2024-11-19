package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deadjoe/benchphant/internal/config"
	"github.com/deadjoe/benchphant/internal/database"
	"github.com/deadjoe/benchphant/internal/models"
	"go.uber.org/zap"
)

func TestServer_HandleConnections(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{Port: 8080}
	storage := database.NewMemoryStorage()
	key := []byte("12345678901234567890123456789012") // 32 bytes encryption key
	manager, err := database.NewManager(storage, key, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	server := NewServer(cfg, manager, logger)

	// Test GET connections
	t.Run("GET connections", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/connections", nil)
		w := httptest.NewRecorder()
		server.handleConnections(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test POST connection
	t.Run("POST connection", func(t *testing.T) {
		conn := models.DBConnection{
			Name:     "test-db",
			Host:     "localhost",
			Port:     3306,
			Username: "test-user",
			Password: "test-pass",
			Database: "test-db",
			Type:     models.MySQL,
		}

		body, _ := json.Marshal(conn)
		req := httptest.NewRequest(http.MethodPost, "/api/connections", bytes.NewReader(body))
		w := httptest.NewRecorder()
		server.handleConnections(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}
	})
}

func TestServer_HandleTestConnection(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{Port: 8080}
	storage := database.NewMemoryStorage()
	key := []byte("12345678901234567890123456789012") // 32 bytes encryption key
	manager, err := database.NewManager(storage, key, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	server := NewServer(cfg, manager, logger)

	// Test connection test endpoint
	t.Run("Test connection", func(t *testing.T) {
		conn := models.DBConnection{
			Name:     "test-db",
			Host:     "localhost",
			Port:     3306,
			Username: "test-user",
			Password: "test-pass",
			Database: "test-db",
			Type:     models.MySQL,
		}

		body, _ := json.Marshal(conn)
		req := httptest.NewRequest(http.MethodPost, "/api/connections/test", bytes.NewReader(body))
		w := httptest.NewRecorder()
		server.handleTestConnection(w, req)

		// Note: This will likely fail since we're not actually connecting to a database
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d (since no real DB), got %d", http.StatusBadRequest, w.Code)
		}
	})
}
