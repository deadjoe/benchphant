package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/deadjoe/benchphant/internal/config"
	"github.com/deadjoe/benchphant/internal/database"
	"github.com/deadjoe/benchphant/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTestServer(t *testing.T) *Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	dbManager, err := database.NewManager(&database.Config{
		StorageType: "memory",
	})
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	return NewServer(cfg, dbManager, logger)
}

func TestHandleBenchmarkStart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.POST("/api/v1/benchmark/start", gin.WrapF(server.handleBenchmarkStart))

	// Test case 1: Valid request
	validConfig := &models.BenchmarkConfig{
		ConnectionID: "test-conn",
		Duration:     60,
		Threads:      10,
		Query:        "SELECT 1",
		Name:         "Test Benchmark",
		Description:  "Test benchmark description",
	}

	validBody, err := json.Marshal(validConfig)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(validBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	// Test case 2: Invalid request - missing required fields
	invalidConfig := &models.BenchmarkConfig{
		Duration: 60,
		Threads:  10,
	}

	invalidBody, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test case 3: Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleBenchmarkStop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.POST("/api/v1/benchmark/stop", gin.WrapF(server.handleBenchmarkStop))

	// Test case 1: No active benchmark
	req := httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/stop", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test case 2: Active benchmark
	validConfig := &models.BenchmarkConfig{
		ConnectionID: "test-conn",
		Duration:     60,
		Threads:      10,
		Query:        "SELECT 1",
		Name:         "Test Benchmark",
		Description:  "Test benchmark description",
	}

	// Start a benchmark first
	body, _ := json.Marshal(validConfig)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Now stop it
	req = httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/stop", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleBenchmarkStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.GET("/api/v1/benchmark/status", gin.WrapF(server.handleBenchmarkStatus))

	// Test case 1: No active benchmark
	req := httptest.NewRequest(http.MethodGet, "/api/v1/benchmark/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "no active benchmark", response["status"])

	// Test case 2: Active benchmark
	validConfig := &models.BenchmarkConfig{
		ConnectionID: "test-conn",
		Duration:     60,
		Threads:      10,
		Query:        "SELECT 1",
		Name:         "Test Benchmark",
		Description:  "Test benchmark description",
	}

	// Start a benchmark first
	body, _ := json.Marshal(validConfig)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Now check status
	req = httptest.NewRequest(http.MethodGet, "/api/v1/benchmark/status", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "running", response["status"])
}

func TestHandleConnections(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.GET("/api/v1/connections", gin.WrapF(server.handleConnections))
	router.POST("/api/v1/connections", gin.WrapF(server.handleConnections))

	// Test case 1: List connections (GET)
	req := httptest.NewRequest("GET", "/api/v1/connections", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test case 2: Add connection (POST)
	connection := &models.DBConnection{
		ID:       "test-conn",
		Name:     "Test Connection",
		Type:     models.DBTypePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Username: "test",
		Password: "test",
	}

	body, _ := json.Marshal(connection)
	req = httptest.NewRequest("POST", "/api/v1/connections", bytes.NewBuffer(body))
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test case 3: Invalid connection (POST)
	invalidConnection := &models.DBConnection{
		ID:   "", // Missing required field
		Type: models.DBTypePostgreSQL,
	}

	body, _ = json.Marshal(invalidConnection)
	req = httptest.NewRequest("POST", "/api/v1/connections", bytes.NewBuffer(body))
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test case 4: Invalid JSON (POST)
	req = httptest.NewRequest("POST", "/api/v1/connections", bytes.NewBufferString("invalid json"))
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleTestConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.POST("/api/v1/connections/test", gin.WrapF(server.handleTestConnection))

	// Test case 1: Test valid connection
	connection := &models.DBConnection{
		ID:       "test-conn",
		Name:     "Test Connection",
		Type:     models.DBTypePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Username: "test",
		Password: "test",
	}

	body, _ := json.Marshal(connection)
	req := httptest.NewRequest("POST", "/api/v1/connections/test", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test case 2: Invalid connection
	invalidConnection := &models.DBConnection{
		ID:   "", // Missing required field
		Type: models.DBTypePostgreSQL,
	}

	body, _ = json.Marshal(invalidConnection)
	req = httptest.NewRequest("POST", "/api/v1/connections/test", bytes.NewBuffer(body))
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test case 3: Invalid JSON
	req = httptest.NewRequest("POST", "/api/v1/connections/test", bytes.NewBufferString("invalid json"))
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleBenchmarkStartWithDatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.POST("/api/v1/benchmark/start", gin.WrapF(server.handleBenchmarkStart))

	// Test case: Invalid connection ID
	invalidConfig := &models.BenchmarkConfig{
		ConnectionID: "non-existent-conn",
		Duration:     60,
		Threads:      10,
		Query:        "SELECT 1",
		Name:         "Test Benchmark",
		Description:  "Test benchmark description",
	}

	body, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "connection not found")
}

func TestHandleBenchmarkStartWithInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.POST("/api/v1/benchmark/start", gin.WrapF(server.handleBenchmarkStart))

	// First add a valid connection
	conn := &models.DBConnection{
		ID:       1,
		Name:     "test-conn",
		Type:     models.MySQL,
		Host:     "localhost",
		Port:     3306,
		Username: "test",
		Password: "test",
		Database: "test",
	}
	err := server.manager.AddConnection(conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Test case: Invalid SQL query
	invalidConfig := &models.BenchmarkConfig{
		ConnectionID: "test-conn",
		Duration:     60,
		Threads:      10,
		Query:        "INVALID SQL QUERY",
		Name:         "Test Benchmark",
		Description:  "Test benchmark description",
	}

	body, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check status after a short delay
	time.Sleep(100 * time.Millisecond)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/benchmark/status", nil)
	w = httptest.NewRecorder()
	router.GET("/api/v1/benchmark/status", gin.WrapF(server.handleBenchmarkStatus))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	metrics := response["metrics"].(map[string]interface{})
	assert.True(t, metrics["errors"].(float64) > 0)
}

func TestHandleBenchmarkStartWithConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.POST("/api/v1/benchmark/start", gin.WrapF(server.handleBenchmarkStart))

	// First add a valid connection
	conn := &models.DBConnection{
		ID:       1,
		Name:     "test-conn",
		Type:     models.MySQL,
		Host:     "localhost",
		Port:     3306,
		Username: "test",
		Password: "test",
		Database: "test",
	}
	err := server.manager.AddConnection(conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Valid config
	validConfig := &models.BenchmarkConfig{
		ConnectionID: "test-conn",
		Duration:     60,
		Threads:      10,
		Query:        "SELECT 1",
		Name:         "Test Benchmark",
		Description:  "Test benchmark description",
	}

	body, err := json.Marshal(validConfig)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Start first benchmark
	req := httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Try to start another benchmark while the first one is running
	req = httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "benchmark already running")
}

func TestHandleBenchmarkStopWithoutActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.POST("/api/v1/benchmark/stop", gin.WrapF(server.handleBenchmarkStop))

	// Test case: Stop without active benchmark
	req := httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/stop", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "no active benchmark")
}

func TestHandleBenchmarkStatusProgress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.POST("/api/v1/benchmark/start", gin.WrapF(server.handleBenchmarkStart))
	router.GET("/api/v1/benchmark/status", gin.WrapF(server.handleBenchmarkStatus))

	// First add a valid connection
	conn := &models.DBConnection{
		ID:       1,
		Name:     "test-conn",
		Type:     models.MySQL,
		Host:     "localhost",
		Port:     3306,
		Username: "test",
		Password: "test",
		Database: "test",
	}
	err := server.manager.AddConnection(conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Start a short benchmark
	validConfig := &models.BenchmarkConfig{
		ConnectionID: "test-conn",
		Duration:     1,
		Threads:      1,
		Query:        "SELECT 1",
		Name:         "Test Benchmark",
		Description:  "Test benchmark description",
	}

	body, err := json.Marshal(validConfig)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/benchmark/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check progress multiple times
	var lastProgress float64
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)

		req = httptest.NewRequest(http.MethodGet, "/api/v1/benchmark/status", nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		progress := response["progress"].(float64)
		assert.True(t, progress >= lastProgress)
		lastProgress = progress
	}

	// Wait for benchmark to complete
	time.Sleep(1 * time.Second)

	// Check final status
	req = httptest.NewRequest(http.MethodGet, "/api/v1/benchmark/status", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(100), response["progress"].(float64))
}

func TestHandleBenchmarkWithInvalidMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := setupTestServer(t)
	router.Any("/api/v1/benchmark/start", gin.WrapF(server.handleBenchmarkStart))
	router.Any("/api/v1/benchmark/stop", gin.WrapF(server.handleBenchmarkStop))
	router.Any("/api/v1/benchmark/status", gin.WrapF(server.handleBenchmarkStatus))

	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
	}{
		{"Start with GET", http.MethodGet, "/api/v1/benchmark/start", http.StatusMethodNotAllowed},
		{"Start with PUT", http.MethodPut, "/api/v1/benchmark/start", http.StatusMethodNotAllowed},
		{"Start with DELETE", http.MethodDelete, "/api/v1/benchmark/start", http.StatusMethodNotAllowed},
		{"Stop with GET", http.MethodGet, "/api/v1/benchmark/stop", http.StatusMethodNotAllowed},
		{"Stop with PUT", http.MethodPut, "/api/v1/benchmark/stop", http.StatusMethodNotAllowed},
		{"Stop with DELETE", http.MethodDelete, "/api/v1/benchmark/stop", http.StatusMethodNotAllowed},
		{"Status with POST", http.MethodPost, "/api/v1/benchmark/status", http.StatusMethodNotAllowed},
		{"Status with PUT", http.MethodPut, "/api/v1/benchmark/status", http.StatusMethodNotAllowed},
		{"Status with DELETE", http.MethodDelete, "/api/v1/benchmark/status", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"].(string), "method not allowed")
		})
	}
}
