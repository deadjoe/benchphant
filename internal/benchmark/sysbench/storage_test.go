package sysbench

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testHelper contains common test utilities and assertions
type testHelper struct {
	t       *testing.T
	tmpDir  string
	storage *FileStorage
}

// newTestHelper creates a new test helper with temporary directory
func newTestHelper(t *testing.T) *testHelper {
	tmpDir, err := os.MkdirTemp("", "sysbench_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	return &testHelper{
		t:       t,
		tmpDir:  tmpDir,
		storage: storage,
	}
}

// cleanup removes the temporary directory
func (h *testHelper) cleanup() {
	os.RemoveAll(h.tmpDir)
}

// createTestReport creates a test report with sample data
func (h *testHelper) createTestReport() *Report {
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
	return NewReport("test_oltp", config, stats, startTime, endTime)
}

// assertFileCount verifies the number of files in a directory
func (h *testHelper) assertFileCount(dir string, expected int) {
	files, err := os.ReadDir(dir)
	if err != nil {
		h.t.Fatalf("Failed to read directory %s: %v", dir, err)
	}
	if len(files) != expected {
		h.t.Errorf("Expected %d files in %s, got %d", expected, dir, len(files))
	}
}

func TestFileStorage(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	report := h.createTestReport()

	// Test SaveReport
	t.Run("SaveReport", func(t *testing.T) {
		if err := h.storage.SaveReport(report); err != nil {
			t.Errorf("SaveReport failed: %v", err)
		}
		h.assertFileCount(filepath.Join(h.tmpDir, "reports"), 1)
	})

	// Test ListReports
	t.Run("ListReports", func(t *testing.T) {
		reports, err := h.storage.ListReports()
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

		if err := h.storage.SaveScenarioReports(scenario, reports); err != nil {
			t.Errorf("SaveScenarioReports failed: %v", err)
		}
		h.assertFileCount(filepath.Join(h.tmpDir, "scenarios"), 1)
	})

	// Test ListScenarios
	t.Run("ListScenarios", func(t *testing.T) {
		scenarios, err := h.storage.ListScenarios()
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
		scenarios, err := h.storage.ListScenarios()
		if err != nil {
			t.Fatalf("Failed to list scenarios: %v", err)
		}
		if len(scenarios) == 0 {
			t.Fatal("No scenarios found")
		}

		reports, err := h.storage.LoadScenarioReports(filepath.Base(filepath.Join(h.tmpDir, "scenarios")))
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
