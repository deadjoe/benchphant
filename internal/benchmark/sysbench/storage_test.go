package sysbench

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileStorage(t *testing.T) {
	// Create temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "sysbench_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create storage instance
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create test data
	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now()
	config := TestConfig{
		Threads:   4,
		Tables:    10,
		TableSize: 1000,
		Database:  "test_db",
	}
	stats := Stats{
		TPS:               1000.0,
		LatencyAvg:        time.Millisecond * 10,
		LatencyP95:        time.Millisecond * 20,
		LatencyP99:        time.Millisecond * 30,
		TotalTransactions: 10000,
		Errors:            5,
	}
	report := NewReport("test_oltp", config, stats, startTime, endTime)

	// Test SaveReport
	t.Run("SaveReport", func(t *testing.T) {
		if err := storage.SaveReport(report); err != nil {
			t.Errorf("SaveReport failed: %v", err)
		}

		// Verify file exists
		files, err := os.ReadDir(filepath.Join(tmpDir, "reports"))
		if err != nil {
			t.Fatalf("Failed to read reports dir: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Expected 1 report file, got %d", len(files))
		}
	})

	// Test ListReports
	t.Run("ListReports", func(t *testing.T) {
		reports, err := storage.ListReports()
		if err != nil {
			t.Errorf("ListReports failed: %v", err)
		}
		if len(reports) != 1 {
			t.Errorf("Expected 1 report, got %d", len(reports))
		}
		if reports[0].TestName != report.TestName {
			t.Errorf("Expected test name %s, got %s", report.TestName, reports[0].TestName)
		}
	})

	// Test SaveScenarioReports
	t.Run("SaveScenarioReports", func(t *testing.T) {
		scenario := &Scenario{
			Type:        ScenarioTypeBasicOLTP,
			Name:        "Basic OLTP Test",
			Description: "Basic OLTP test scenario",
			Duration:    time.Hour,
		}
		reports := []*Report{report}

		if err := storage.SaveScenarioReports(scenario, reports); err != nil {
			t.Errorf("SaveScenarioReports failed: %v", err)
		}

		// Verify scenario directory exists
		files, err := os.ReadDir(filepath.Join(tmpDir, "scenarios"))
		if err != nil {
			t.Fatalf("Failed to read scenarios dir: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Expected 1 scenario directory, got %d", len(files))
		}
	})

	// Test ListScenarios
	t.Run("ListScenarios", func(t *testing.T) {
		scenarios, err := storage.ListScenarios()
		if err != nil {
			t.Errorf("ListScenarios failed: %v", err)
		}
		if len(scenarios) != 1 {
			t.Errorf("Expected 1 scenario, got %d", len(scenarios))
		}
		if scenarios[0].Type != ScenarioTypeBasicOLTP {
			t.Errorf("Expected scenario type %s, got %s", ScenarioTypeBasicOLTP, scenarios[0].Type)
		}
	})

	// Test LoadScenarioReports
	t.Run("LoadScenarioReports", func(t *testing.T) {
		scenarios, err := storage.ListScenarios()
		if err != nil {
			t.Fatalf("Failed to list scenarios: %v", err)
		}
		if len(scenarios) == 0 {
			t.Fatal("No scenarios found")
		}

		reports, err := storage.LoadScenarioReports(filepath.Base(filepath.Join(tmpDir, "scenarios")))
		if err != nil {
			t.Errorf("LoadScenarioReports failed: %v", err)
		}
		if len(reports) != 1 {
			t.Errorf("Expected 1 report, got %d", len(reports))
		}
		if reports[0].TestName != report.TestName {
			t.Errorf("Expected test name %s, got %s", report.TestName, reports[0].TestName)
		}
	})
}
