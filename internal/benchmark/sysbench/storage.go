package sysbench

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
)

// FileStorage handles file operations for test reports and scenarios
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(baseDir string) (*FileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %v", err)
	}
	return &FileStorage{baseDir: baseDir}, nil
}

// SaveReport saves a test report to a file
func (s *FileStorage) SaveReport(report types.Report) error {
	reportDir := filepath.Join(s.baseDir, "reports")
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %v", err)
	}

	filename := fmt.Sprintf("report_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(reportDir, filename)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %v", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %v", err)
	}

	return nil
}

// LoadReport loads a test report from a file
func (s *FileStorage) LoadReport(filename string) (*types.Report, error) {
	filepath := filepath.Join(s.baseDir, "reports", filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read report file: %v", err)
	}

	var report types.Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %v", err)
	}

	return &report, nil
}

// SaveScenario saves a test scenario to a file
func (s *FileStorage) SaveScenario(scenario types.Scenario) error {
	scenarioDir := filepath.Join(s.baseDir, "scenarios")
	if err := os.MkdirAll(scenarioDir, 0755); err != nil {
		return fmt.Errorf("failed to create scenarios directory: %v", err)
	}

	filename := fmt.Sprintf("%s.json", scenario.Name)
	filepath := filepath.Join(scenarioDir, filename)

	data, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scenario: %v", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write scenario file: %v", err)
	}

	return nil
}

// LoadScenario loads a test scenario from a file
func (s *FileStorage) LoadScenario(name string) (*types.Scenario, error) {
	filepath := filepath.Join(s.baseDir, "scenarios", fmt.Sprintf("%s.json", name))

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %v", err)
	}

	var scenario types.Scenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scenario: %v", err)
	}

	return &scenario, nil
}

// ListScenarios returns a list of available test scenarios
func (s *FileStorage) ListScenarios() ([]string, error) {
	scenarioDir := filepath.Join(s.baseDir, "scenarios")
	if err := os.MkdirAll(scenarioDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create scenarios directory: %v", err)
	}

	files, err := os.ReadDir(scenarioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios directory: %v", err)
	}

	var scenarios []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			scenarios = append(scenarios, file.Name()[:len(file.Name())-5])
		}
	}

	return scenarios, nil
}
