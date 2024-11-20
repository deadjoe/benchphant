package storage

import (
	"os"
	"testing"
	"time"
)

func TestSQLiteStorage(t *testing.T) {
	// Create a temporary database file
	dbPath := "test.db"
	defer os.Remove(dbPath)

	// Create storage instance
	storage, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Test storing and retrieving a benchmark
	benchmarkID := "test-benchmark-1"
	config := map[string]interface{}{
		"database": "test_db",
		"threads":  10,
	}
	results := map[string]interface{}{
		"tps": 1000,
		"latency": map[string]interface{}{
			"avg": 10.5,
			"p95": 20.0,
		},
	}

	err = storage.StoreBenchmark(benchmarkID, "Test Benchmark", "A test benchmark", config, results)
	if err != nil {
		t.Fatalf("Failed to store benchmark: %v", err)
	}

	// Test retrieving the benchmark
	benchmark, err := storage.GetBenchmark(benchmarkID)
	if err != nil {
		t.Fatalf("Failed to get benchmark: %v", err)
	}
	if benchmark == nil {
		t.Fatal("Expected benchmark to exist")
	}
	if benchmark["name"] != "Test Benchmark" {
		t.Errorf("Expected benchmark name 'Test Benchmark', got %v", benchmark["name"])
	}

	// Test storing and retrieving metrics
	labels := map[string]string{
		"database":  "test_db",
		"operation": "insert",
	}

	err = storage.StoreMetric(benchmarkID, "transactions", "counter", 100.0, labels)
	if err != nil {
		t.Fatalf("Failed to store metric: %v", err)
	}

	// Wait a bit to ensure different timestamps
	time.Sleep(time.Millisecond)

	err = storage.StoreMetric(benchmarkID, "latency", "gauge", 15.5, labels)
	if err != nil {
		t.Fatalf("Failed to store metric: %v", err)
	}

	// Test retrieving metrics
	metrics, err := storage.GetMetrics(benchmarkID)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(metrics))
	}

	// Verify first metric
	if metrics[0]["name"] != "transactions" {
		t.Errorf("Expected first metric name 'transactions', got %v", metrics[0]["name"])
	}
	if metrics[0]["value"].(float64) != 100.0 {
		t.Errorf("Expected first metric value 100.0, got %v", metrics[0]["value"])
	}

	// Verify second metric
	if metrics[1]["name"] != "latency" {
		t.Errorf("Expected second metric name 'latency', got %v", metrics[1]["name"])
	}
	if metrics[1]["value"].(float64) != 15.5 {
		t.Errorf("Expected second metric value 15.5, got %v", metrics[1]["value"])
	}
}

func TestSQLiteStorage_NonExistentBenchmark(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	storage, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	benchmark, err := storage.GetBenchmark("non-existent")
	if err != nil {
		t.Fatalf("Expected no error for non-existent benchmark, got: %v", err)
	}
	if benchmark != nil {
		t.Error("Expected nil result for non-existent benchmark")
	}
}

func TestSQLiteStorage_EmptyMetrics(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	storage, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	metrics, err := storage.GetMetrics("non-existent")
	if err != nil {
		t.Fatalf("Expected no error for non-existent benchmark metrics, got: %v", err)
	}
	if len(metrics) != 0 {
		t.Error("Expected empty metrics array for non-existent benchmark")
	}
}
