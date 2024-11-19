package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBenchmark(t *testing.T) {
	t.Run("ValidateBenchmark", func(t *testing.T) {
		tests := []struct {
			name    string
			bench   *Benchmark
			wantErr bool
		}{
			{
				name: "ValidBenchmark",
				bench: &Benchmark{
					Name:           "test_benchmark",
					Description:    "Test benchmark description",
					ConnectionID:   1,
					QueryTemplate: "SELECT * FROM test",
					NumThreads:    10,
					Duration:      time.Minute * 5,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Status:        BenchmarkStatusPending,
				},
				wantErr: false,
			},
			{
				name: "EmptyName",
				bench: &Benchmark{
					Name:           "",
					Description:    "Test benchmark description",
					ConnectionID:   1,
					QueryTemplate: "SELECT * FROM test",
					NumThreads:    10,
					Duration:      time.Minute * 5,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Status:        BenchmarkStatusPending,
				},
				wantErr: true,
			},
			{
				name: "InvalidConnectionID",
				bench: &Benchmark{
					Name:           "test_benchmark",
					Description:    "Test benchmark description",
					ConnectionID:   0,
					QueryTemplate: "SELECT * FROM test",
					NumThreads:    10,
					Duration:      time.Minute * 5,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Status:        BenchmarkStatusPending,
				},
				wantErr: true,
			},
			{
				name: "EmptyQueryTemplate",
				bench: &Benchmark{
					Name:           "test_benchmark",
					Description:    "Test benchmark description",
					ConnectionID:   1,
					QueryTemplate: "",
					NumThreads:    10,
					Duration:      time.Minute * 5,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Status:        BenchmarkStatusPending,
				},
				wantErr: true,
			},
			{
				name: "InvalidNumThreads",
				bench: &Benchmark{
					Name:           "test_benchmark",
					Description:    "Test benchmark description",
					ConnectionID:   1,
					QueryTemplate: "SELECT * FROM test",
					NumThreads:    0,
					Duration:      time.Minute * 5,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Status:        BenchmarkStatusPending,
				},
				wantErr: true,
			},
			{
				name: "InvalidDuration",
				bench: &Benchmark{
					Name:           "test_benchmark",
					Description:    "Test benchmark description",
					ConnectionID:   1,
					QueryTemplate: "SELECT * FROM test",
					NumThreads:    10,
					Duration:      0,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Status:        BenchmarkStatusPending,
				},
				wantErr: true,
			},
			{
				name: "InvalidStatus",
				bench: &Benchmark{
					Name:           "test_benchmark",
					Description:    "Test benchmark description",
					ConnectionID:   1,
					QueryTemplate: "SELECT * FROM test",
					NumThreads:    10,
					Duration:      time.Minute * 5,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Status:        "invalid_status",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.bench.Validate()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("BenchmarkStatus", func(t *testing.T) {
		assert.Equal(t, "pending", string(BenchmarkStatusPending))
		assert.Equal(t, "running", string(BenchmarkStatusRunning))
		assert.Equal(t, "completed", string(BenchmarkStatusCompleted))
		assert.Equal(t, "failed", string(BenchmarkStatusFailed))
		assert.Equal(t, "cancelled", string(BenchmarkStatusCancelled))
	})

	t.Run("BenchmarkResult", func(t *testing.T) {
		result := &BenchmarkResult{
			BenchmarkID:    1,
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(time.Minute),
			TotalQueries:  1000,
			SuccessCount:  950,
			FailureCount:  50,
			AverageLatency: time.Millisecond * 100,
			MinLatency:    time.Millisecond * 50,
			MaxLatency:    time.Millisecond * 200,
			QPS:           100.0,
			Error:         "test error",
		}

		assert.NotNil(t, result)
		assert.Equal(t, int64(1000), result.TotalQueries)
		assert.Equal(t, int64(950), result.SuccessCount)
		assert.Equal(t, int64(50), result.FailureCount)
		assert.Equal(t, time.Millisecond*100, result.AverageLatency)
		assert.Equal(t, time.Millisecond*50, result.MinLatency)
		assert.Equal(t, time.Millisecond*200, result.MaxLatency)
		assert.Equal(t, 100.0, result.QPS)
		assert.Equal(t, "test error", result.Error)
	})
}
