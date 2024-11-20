package sysbench

import (
	"context"
	"database/sql"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
)

// ScenarioType represents different predefined test scenarios
type ScenarioType string

const (
	// Common scenarios
	ScenarioTypeBasicOLTP       = ScenarioType(types.ScenarioTypeBasicOLTP)
	ScenarioTypeReadIntensive   = ScenarioType(types.ScenarioTypeReadIntensive)
	ScenarioTypeWriteHeavy      = ScenarioType(types.ScenarioTypeWriteHeavy)
	ScenarioTypeMixedLoad       = ScenarioType(types.ScenarioTypeMixedLoad)
	ScenarioTypeHighConcurrency = ScenarioType(types.ScenarioTypeHighConcurrency)
	ScenarioTypeLongRunning     = ScenarioType(types.ScenarioTypeLongRunning)
	ScenarioTypeStressTest      = ScenarioType(types.ScenarioTypeStressTest)
)

// Scenario represents a test scenario configuration
type Scenario struct {
	Type        ScenarioType
	Name        string
	Description string
	Duration    time.Duration
	Tests       []*ScenarioTest
}

// ScenarioTest represents a single test within a scenario
type ScenarioTest struct {
	Name     string
	Type     types.TestType
	Config   *types.OLTPTestConfig
	Weight   float64 // Weight for mixed workloads (0-1)
	Duration time.Duration
}

// NewScenario creates a new test scenario
func NewScenario(scenarioType ScenarioType) *Scenario {
	scenario := &Scenario{
		Type:  scenarioType,
		Tests: make([]*ScenarioTest, 0),
	}

	switch scenarioType {
	case ScenarioTypeBasicOLTP:
		scenario.Name = "Basic OLTP"
		scenario.Description = "Basic OLTP workload with balanced read/write operations"
		scenario.Duration = 30 * time.Minute
		scenario.addBasicOLTPTests()

	case ScenarioTypeReadIntensive:
		scenario.Name = "Read Intensive"
		scenario.Description = "Read-intensive workload with 95% reads and 5% writes"
		scenario.Duration = 30 * time.Minute
		scenario.addReadIntensiveTests()

	case ScenarioTypeWriteHeavy:
		scenario.Name = "Write Heavy"
		scenario.Description = "Write-heavy workload with 70% writes and 30% reads"
		scenario.Duration = 30 * time.Minute
		scenario.addWriteHeavyTests()

	case ScenarioTypeMixedLoad:
		scenario.Name = "Mixed Load"
		scenario.Description = "Mixed workload with various query types"
		scenario.Duration = 1 * time.Hour
		scenario.addMixedLoadTests()

	case ScenarioTypeHighConcurrency:
		scenario.Name = "High Concurrency"
		scenario.Description = "High concurrency test with many threads"
		scenario.Duration = 15 * time.Minute
		scenario.addHighConcurrencyTests()

	case ScenarioTypeLongRunning:
		scenario.Name = "Long Running"
		scenario.Description = "Long-running stability test"
		scenario.Duration = 24 * time.Hour
		scenario.addLongRunningTests()

	case ScenarioTypeStressTest:
		scenario.Name = "Stress Test"
		scenario.Description = "System stress test with increasing load"
		scenario.Duration = 2 * time.Hour
		scenario.addStressTests()
	}

	return scenario
}

// addBasicOLTPTests adds tests for basic OLTP scenario
func (s *Scenario) addBasicOLTPTests() {
	config := types.NewOLTPTestConfig()
	config.Threads = 8
	config.TableSize = 1000000
	config.TablesCount = 4

	s.Tests = append(s.Tests,
		&ScenarioTest{
			Name:     "Read-Write Mix",
			Type:     types.TestTypeOLTPReadWrite,
			Config:   config,
			Weight:   1.0,
			Duration: s.Duration,
		},
	)
}

// addReadIntensiveTests adds tests for read-intensive scenario
func (s *Scenario) addReadIntensiveTests() {
	config := types.NewOLTPTestConfig()
	config.Threads = 16
	config.TableSize = 1000000
	config.TablesCount = 4

	s.Tests = append(s.Tests,
		&ScenarioTest{
			Name:     "Read Only",
			Type:     types.TestTypeOLTPRead,
			Config:   config,
			Weight:   0.7,
			Duration: s.Duration * 7 / 10,
		},
		&ScenarioTest{
			Name:     "Point Selects",
			Type:     types.TestTypeOLTPPointSelect,
			Config:   config,
			Weight:   0.25,
			Duration: s.Duration * 25 / 100,
		},
		&ScenarioTest{
			Name:     "Write Mix",
			Type:     types.TestTypeOLTPReadWrite,
			Config:   config,
			Weight:   0.05,
			Duration: s.Duration * 5 / 100,
		},
	)
}

// addWriteHeavyTests adds tests for write-heavy scenario
func (s *Scenario) addWriteHeavyTests() {
	config := types.NewOLTPTestConfig()
	config.Threads = 8
	config.TableSize = 500000
	config.TablesCount = 4

	s.Tests = append(s.Tests,
		&ScenarioTest{
			Name:     "Write Only",
			Type:     types.TestTypeOLTPWrite,
			Config:   config,
			Weight:   0.7,
			Duration: s.Duration * 7 / 10,
		},
		&ScenarioTest{
			Name:     "Read-Write Mix",
			Type:     types.TestTypeOLTPReadWrite,
			Config:   config,
			Weight:   0.3,
			Duration: s.Duration * 3 / 10,
		},
	)
}

// addMixedLoadTests adds tests for mixed load scenario
func (s *Scenario) addMixedLoadTests() {
	config := types.NewOLTPTestConfig()
	config.Threads = 16
	config.TableSize = 1000000
	config.TablesCount = 8

	s.Tests = append(s.Tests,
		&ScenarioTest{
			Name:     "Point Selects",
			Type:     types.TestTypeOLTPPointSelect,
			Config:   config,
			Weight:   0.3,
			Duration: s.Duration * 3 / 10,
		},
		&ScenarioTest{
			Name:     "Simple Ranges",
			Type:     types.TestTypeOLTPSimpleSelect,
			Config:   config,
			Weight:   0.2,
			Duration: s.Duration * 2 / 10,
		},
		&ScenarioTest{
			Name:     "Sum Ranges",
			Type:     types.TestTypeOLTPSumRange,
			Config:   config,
			Weight:   0.2,
			Duration: s.Duration * 2 / 10,
		},
		&ScenarioTest{
			Name:     "Write Mix",
			Type:     types.TestTypeOLTPReadWrite,
			Config:   config,
			Weight:   0.3,
			Duration: s.Duration * 3 / 10,
		},
	)
}

// addHighConcurrencyTests adds tests for high concurrency scenario
func (s *Scenario) addHighConcurrencyTests() {
	config := types.NewOLTPTestConfig()
	config.Threads = 64
	config.TableSize = 1000000
	config.TablesCount = 16

	s.Tests = append(s.Tests,
		&ScenarioTest{
			Name:     "Read-Write High Concurrency",
			Type:     types.TestTypeOLTPReadWrite,
			Config:   config,
			Weight:   1.0,
			Duration: s.Duration,
		},
	)
}

// addLongRunningTests adds tests for long-running scenario
func (s *Scenario) addLongRunningTests() {
	config := types.NewOLTPTestConfig()
	config.Threads = 32
	config.TableSize = 5000000
	config.TablesCount = 8

	s.Tests = append(s.Tests,
		&ScenarioTest{
			Name:     "Long Running Mix",
			Type:     types.TestTypeOLTPReadWrite,
			Config:   config,
			Weight:   1.0,
			Duration: s.Duration,
		},
	)
}

// addStressTests adds tests for stress testing scenario
func (s *Scenario) addStressTests() {
	baseConfig := types.NewOLTPTestConfig()
	baseConfig.TableSize = 1000000
	baseConfig.TablesCount = 8

	// Add tests with increasing thread counts
	threadCounts := []int{8, 16, 32, 64, 128}
	duration := s.Duration / time.Duration(len(threadCounts))

	for _, threads := range threadCounts {
		config := *baseConfig
		config.Threads = threads
		s.Tests = append(s.Tests,
			&ScenarioTest{
				Name:     fmt.Sprintf("Stress Test (%d threads)", threads),
				Type:     types.TestTypeOLTPReadWrite,
				Config:   &config,
				Weight:   1.0 / float64(len(threadCounts)),
				Duration: duration,
			},
		)
	}
}

// Run executes all tests in the scenario
func (s *Scenario) Run(ctx context.Context) ([]*types.Report, error) {
	reports := make([]*types.Report, 0, len(s.Tests))

	for _, test := range s.Tests {
		// Create test context with duration limit
		testCtx, cancel := context.WithTimeout(ctx, test.Duration)
		defer cancel()

		// Create and prepare test
		oltpTest := types.NewOLTPTest(test.Config)
		if err := oltpTest.Prepare(testCtx); err != nil {
			return reports, fmt.Errorf("failed to prepare test %s: %w", test.Name, err)
		}

		// Run test
		report, err := oltpTest.Run(testCtx)
		if err != nil {
			return reports, fmt.Errorf("failed to run test %s: %w", test.Name, err)
		}

		// Clean up test
		if err := oltpTest.Cleanup(testCtx); err != nil {
			return reports, fmt.Errorf("failed to clean up test %s: %w", test.Name, err)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// GetPredefinedScenarios returns a list of predefined test scenarios
func GetPredefinedScenarios() []Scenario {
	return []Scenario{
		{
			Type:        ScenarioTypeBasicOLTP,
			Name:        "basic_oltp",
			Description: "Basic OLTP test with balanced read/write operations",
			Duration:    10 * time.Minute,
		},
		{
			Type:        ScenarioTypeReadIntensive,
			Name:        "read_intensive",
			Description: "OLTP test focused on read operations",
			Duration:    15 * time.Minute,
		},
		{
			Type:        ScenarioTypeWriteHeavy,
			Name:        "write_heavy",
			Description: "OLTP test with high write load",
			Duration:    15 * time.Minute,
		},
		{
			Type:        ScenarioTypeMixedLoad,
			Name:        "mixed_load",
			Description: "OLTP test with mixed read/write operations",
			Duration:    20 * time.Minute,
		},
		{
			Type:        ScenarioTypeHighConcurrency,
			Name:        "high_concurrency",
			Description: "OLTP test with high thread count",
			Duration:    30 * time.Minute,
		},
		{
			Type:        ScenarioTypeLongRunning,
			Name:        "long_running",
			Description: "Extended duration OLTP test",
			Duration:    2 * time.Hour,
		},
		{
			Type:        ScenarioTypeStressTest,
			Name:        "stress_test",
			Description: "High load stress test",
			Duration:    1 * time.Hour,
		},
	}
}

// GetScenarioByType returns a predefined scenario by its type
func GetScenarioByType(scenarioType ScenarioType) (*Scenario, error) {
	scenarios := GetPredefinedScenarios()
	for _, s := range scenarios {
		if s.Type == scenarioType {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("scenario type %s not found", scenarioType)
}

// GetScenarioByName returns a predefined scenario by its name
func GetScenarioByName(name string) (*Scenario, error) {
	scenarios := GetPredefinedScenarios()
	for _, s := range scenarios {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("scenario with name %s not found", name)
}
