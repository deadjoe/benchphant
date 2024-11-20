package sysbench

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
	"go.uber.org/zap"
)

// ScenarioType represents different predefined test scenarios
type ScenarioType types.ScenarioType

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
	Tests       []*types.Test
}

// NewScenario creates a new test scenario
func NewScenario(scenarioType ScenarioType) *Scenario {
	scenario := &Scenario{
		Type:  scenarioType,
		Tests: make([]*types.Test, 0),
	}

	switch scenarioType {
	case ScenarioTypeBasicOLTP:
		scenario.Name = "Basic OLTP"
		scenario.Description = "Basic OLTP test with balanced read/write operations"
		scenario.Duration = 10 * time.Minute
		scenario.addBasicOLTPTests()

	case ScenarioTypeReadIntensive:
		scenario.Name = "Read Intensive"
		scenario.Description = "OLTP test focused on read operations"
		scenario.Duration = 15 * time.Minute
		scenario.addReadIntensiveTests()

	case ScenarioTypeWriteHeavy:
		scenario.Name = "Write Heavy"
		scenario.Description = "OLTP test with high write load"
		scenario.Duration = 15 * time.Minute
		scenario.addWriteHeavyTests()

	case ScenarioTypeMixedLoad:
		scenario.Name = "Mixed Load"
		scenario.Description = "OLTP test with mixed read/write operations"
		scenario.Duration = 20 * time.Minute
		scenario.addMixedLoadTests()

	case ScenarioTypeHighConcurrency:
		scenario.Name = "High Concurrency"
		scenario.Description = "OLTP test with high thread count"
		scenario.Duration = 30 * time.Minute
		scenario.addHighConcurrencyTests()

	case ScenarioTypeLongRunning:
		scenario.Name = "Long Running"
		scenario.Description = "Extended duration OLTP test"
		scenario.Duration = 2 * time.Hour
		scenario.addLongRunningTests()

	case ScenarioTypeStressTest:
		scenario.Name = "Stress Test"
		scenario.Description = "High load stress test"
		scenario.Duration = 1 * time.Hour
		scenario.addStressTests()
	}

	return scenario
}

// addBasicOLTPTests adds tests for basic OLTP scenario
func (s *Scenario) addBasicOLTPTests() {
	config := types.NewOLTPTestConfig()
	config.NumThreads = 8
	config.TableSize = 1000000
	config.NumTables = 4
	config.Duration = s.Duration

	s.Tests = append(s.Tests,
		&types.Test{
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
	config.NumThreads = 16
	config.TableSize = 1000000
	config.NumTables = 4
	config.Duration = s.Duration

	s.Tests = append(s.Tests,
		&types.Test{
			Name:     "Read Only",
			Type:     types.TestTypeOLTPRead,
			Config:   config,
			Weight:   0.7,
			Duration: s.Duration * 7 / 10,
		},
		&types.Test{
			Name:     "Point Selects",
			Type:     types.TestTypeOLTPPointSelect,
			Config:   config,
			Weight:   0.25,
			Duration: s.Duration * 25 / 100,
		},
		&types.Test{
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
	config.NumThreads = 8
	config.TableSize = 500000
	config.NumTables = 4
	config.Duration = s.Duration

	s.Tests = append(s.Tests,
		&types.Test{
			Name:     "Write Only",
			Type:     types.TestTypeOLTPWrite,
			Config:   config,
			Weight:   0.7,
			Duration: s.Duration * 7 / 10,
		},
		&types.Test{
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
	config.NumThreads = 16
	config.TableSize = 1000000
	config.NumTables = 8
	config.Duration = s.Duration

	s.Tests = append(s.Tests,
		&types.Test{
			Name:     "Point Selects",
			Type:     types.TestTypeOLTPPointSelect,
			Config:   config,
			Weight:   0.3,
			Duration: s.Duration * 3 / 10,
		},
		&types.Test{
			Name:     "Simple Ranges",
			Type:     types.TestTypeOLTPSimpleSelect,
			Config:   config,
			Weight:   0.2,
			Duration: s.Duration * 2 / 10,
		},
		&types.Test{
			Name:     "Sum Ranges",
			Type:     types.TestTypeOLTPSumRange,
			Config:   config,
			Weight:   0.2,
			Duration: s.Duration * 2 / 10,
		},
		&types.Test{
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
	config.NumThreads = 32
	config.TableSize = 1000000
	config.NumTables = 4
	config.Duration = s.Duration

	s.Tests = append(s.Tests,
		&types.Test{
			Name:     "High Concurrency Mix",
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
	config.NumThreads = 16
	config.TableSize = 1000000
	config.NumTables = 4
	config.Duration = s.Duration

	s.Tests = append(s.Tests,
		&types.Test{
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
	config := types.NewOLTPTestConfig()
	config.NumThreads = 64
	config.TableSize = 2000000
	config.NumTables = 8
	config.Duration = s.Duration

	s.Tests = append(s.Tests,
		&types.Test{
			Name:     "Stress Test Mix",
			Type:     types.TestTypeOLTPReadWrite,
			Config:   config,
			Weight:   1.0,
			Duration: s.Duration,
		},
	)
}

// Run executes the scenario
func (s *Scenario) Run(ctx context.Context, db *sql.DB, logger *zap.Logger) ([]*types.Report, error) {
	var reports []*types.Report

	// Create test configuration
	config := types.NewOLTPTestConfig()
	config.TestType = types.TestType(s.Type)
	config.Duration = s.Duration
	config.NumThreads = 8
	config.NumTables = 4
	config.TableSize = 1000000

	// Create test
	test := NewOLTPTest(config)
	test.SetDB(db)
	test.SetLogger(logger)

	// Prepare test
	if err := test.Prepare(ctx); err != nil {
		return nil, fmt.Errorf("prepare test failed: %w", err)
	}

	// Run test
	if err := test.Run(ctx); err != nil {
		return nil, fmt.Errorf("run test failed: %w", err)
	}

	// Get report
	report := test.GetReport()
	reports = append(reports, report)

	// Clean up
	if err := test.Cleanup(ctx); err != nil {
		return nil, fmt.Errorf("cleanup test failed: %w", err)
	}

	return reports, nil
}

// GetPredefinedScenarios returns a list of predefined test scenarios
func GetPredefinedScenarios() []Scenario {
	scenarios := []Scenario{
		*NewScenario(ScenarioTypeBasicOLTP),
		*NewScenario(ScenarioTypeReadIntensive),
		*NewScenario(ScenarioTypeWriteHeavy),
		*NewScenario(ScenarioTypeMixedLoad),
		*NewScenario(ScenarioTypeHighConcurrency),
		*NewScenario(ScenarioTypeLongRunning),
		*NewScenario(ScenarioTypeStressTest),
	}

	return scenarios
}

// GetScenarioByType returns a predefined scenario by its type
func GetScenarioByType(scenarioType ScenarioType) (*Scenario, error) {
	return NewScenario(scenarioType), nil
}

// GetScenarioByName returns a predefined scenario by its name
func GetScenarioByName(name string) (*Scenario, error) {
	scenarios := GetPredefinedScenarios()
	for _, s := range scenarios {
		if s.Name == name {
			scenario := s
			return &scenario, nil
		}
	}
	return nil, fmt.Errorf("scenario not found: %s", name)
}
