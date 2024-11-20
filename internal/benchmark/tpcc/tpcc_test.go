package tpcc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestTPCCBenchmark(t *testing.T) {
	// Create test logger
	logger := zaptest.NewLogger(t)

	// Create test database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test config
	config := &Config{
		Warehouses:            1,
		Terminals:             2,
		Duration:              5 * time.Second,
		ReportInterval:        1 * time.Second,
		NewOrderItemsMin:      5,
		NewOrderItemsMax:      15,
		NewOrderPercentage:    45,
		PaymentPercentage:     43,
		OrderStatusPercentage: 4,
		DeliveryPercentage:    4,
		StockLevelPercentage:  4,
	}

	t.Run("Validate", func(t *testing.T) {
		benchmark := NewTPCCBenchmark(config, db, logger)
		assert.NoError(t, benchmark.Validate())

		// Test invalid config
		invalidConfig := *config
		invalidConfig.Warehouses = 0
		invalidBenchmark := NewTPCCBenchmark(&invalidConfig, db, logger)
		assert.Error(t, invalidBenchmark.Validate())
	})

	t.Run("Setup", func(t *testing.T) {
		benchmark := NewTPCCBenchmark(config, db, logger)
		ctx := context.Background()
		assert.NoError(t, benchmark.Setup(ctx))

		// Verify schema
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM warehouse").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, config.Warehouses, count)
	})

	t.Run("Run", func(t *testing.T) {
		benchmark := NewTPCCBenchmark(config, db, logger)
		ctx := context.Background()
		require.NoError(t, benchmark.Setup(ctx))

		result, err := benchmark.Run(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "tpcc", result.Name)
		assert.True(t, result.Duration >= config.Duration)
		assert.NotEmpty(t, result.Metrics)
	})

	t.Run("Cleanup", func(t *testing.T) {
		benchmark := NewTPCCBenchmark(config, db, logger)
		ctx := context.Background()
		require.NoError(t, benchmark.Setup(ctx))
		assert.NoError(t, benchmark.Cleanup(ctx))

		// Verify schema is dropped
		_, err := db.Query("SELECT * FROM warehouse")
		assert.Error(t, err)
	})
}

func TestFactory(t *testing.T) {
	logger := zaptest.NewLogger(t)
	factory := NewFactory(logger)

	t.Run("Create", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		config := Config{
			Warehouses:     1,
			Terminals:      2,
			Duration:       5 * time.Second,
			ReportInterval: 1 * time.Second,
		}

		configJSON, err := json.Marshal(config)
		require.NoError(t, err)

		benchmark, err := factory.Create(db, configJSON)
		assert.NoError(t, err)
		assert.NotNil(t, benchmark)
		assert.Equal(t, "tpcc", benchmark.Name())
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		invalidConfig := struct {
			Invalid string `json:"invalid"`
		}{
			Invalid: "invalid",
		}

		configJSON, err := json.Marshal(invalidConfig)
		require.NoError(t, err)

		benchmark, err := factory.Create(db, configJSON)
		assert.Error(t, err)
		assert.Nil(t, benchmark)
	})
}

func TestRunner(t *testing.T) {
	logger := zaptest.NewLogger(t)
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	config := &Config{
		Warehouses:            1,
		Terminals:             2,
		Duration:              2 * time.Second,
		ReportInterval:        1 * time.Second,
		NewOrderItemsMin:      5,
		NewOrderItemsMax:      15,
		NewOrderPercentage:    45,
		PaymentPercentage:     43,
		OrderStatusPercentage: 4,
		DeliveryPercentage:    4,
		StockLevelPercentage:  4,
	}

	t.Run("Run", func(t *testing.T) {
		runner := NewRunner(db, config, logger)
		ctx := context.Background()

		// Setup schema and load data
		require.NoError(t, CreateSchema(ctx, db))
		loader := NewLoader(db, config)
		require.NoError(t, loader.Load(ctx))

		// Run benchmark
		assert.NoError(t, runner.Run(ctx))

		// Check stats
		stats := runner.GetStats()
		assert.True(t, stats.TotalTransactions > 0)
		assert.True(t, stats.TPMc > 0)
		assert.True(t, stats.Efficiency > 0)
	})
}

func TestTransactionExecutor(t *testing.T) {
	logger := zaptest.NewLogger(t)
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	config := &Config{
		Warehouses: 1,
		Terminals:  1,
	}

	ctx := context.Background()
	require.NoError(t, CreateSchema(ctx, db))
	loader := NewLoader(db, config)
	require.NoError(t, loader.Load(ctx))

	executor := NewTransactionExecutor(db, config)

	t.Run("NewOrder", func(t *testing.T) {
		tx := &NewOrder{
			db:       db,
			wID:      1,
			dID:      1,
			cID:      1,
			itemIDs:  []int{1, 2},
			supplyWs: []int{1, 1},
			qtys:     []int{1, 1},
			allLocal: true,
		}
		assert.NoError(t, executor.ExecuteNewOrder(ctx, tx))
	})

	t.Run("Payment", func(t *testing.T) {
		tx := &Payment{
			db:     db,
			wID:    1,
			dID:    1,
			cID:    1,
			amount: 100.0,
		}
		assert.NoError(t, executor.ExecutePayment(ctx, tx))
	})

	t.Run("OrderStatus", func(t *testing.T) {
		tx := &OrderStatus{
			db:  db,
			wID: 1,
			dID: 1,
			cID: 1,
		}
		assert.NoError(t, executor.ExecuteOrderStatus(ctx, tx))
	})

	t.Run("Delivery", func(t *testing.T) {
		tx := &Delivery{
			db:        db,
			wID:       1,
			carrierID: 1,
		}
		assert.NoError(t, executor.ExecuteDelivery(ctx, tx))
	})

	t.Run("StockLevel", func(t *testing.T) {
		tx := &StockLevel{
			db:        db,
			wID:       1,
			dID:       1,
			threshold: 10,
		}
		assert.NoError(t, executor.ExecuteStockLevel(ctx, tx))
	})
}

func TestConcurrentTransactions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	config := &Config{
		Warehouses: 2,
		Terminals:  10,
	}

	ctx := context.Background()
	require.NoError(t, CreateSchema(ctx, db))
	loader := NewLoader(db, config)
	require.NoError(t, loader.Load(ctx))

	executor := NewTransactionExecutor(db, config)

	// Test concurrent new order transactions
	t.Run("ConcurrentNewOrders", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tx := &NewOrder{
					db:       db,
					wID:      1,
					dID:      1,
					cID:      1,
					itemIDs:  []int{1, 2},
					supplyWs: []int{1, 1},
					qtys:     []int{1, 1},
					allLocal: true,
				}
				if err := executor.ExecuteNewOrder(ctx, tx); err != nil {
					errChan <- err
				}
			}()
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			assert.NoError(t, err)
		}
	})

	// Test concurrent mixed transactions
	t.Run("ConcurrentMixed", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				var err error
				switch i % 5 {
				case 0:
					tx := &NewOrder{
						db:       db,
						wID:      1,
						dID:      1,
						cID:      1,
						itemIDs:  []int{1, 2},
						supplyWs: []int{1, 1},
						qtys:     []int{1, 1},
						allLocal: true,
					}
					err = executor.ExecuteNewOrder(ctx, tx)
				case 1:
					tx := &Payment{
						db:     db,
						wID:    1,
						dID:    1,
						cID:    1,
						amount: 100.0,
					}
					err = executor.ExecutePayment(ctx, tx)
				case 2:
					tx := &OrderStatus{
						db:  db,
						wID: 1,
						dID: 1,
						cID: 1,
					}
					err = executor.ExecuteOrderStatus(ctx, tx)
				case 3:
					tx := &Delivery{
						db:        db,
						wID:       1,
						carrierID: 1,
					}
					err = executor.ExecuteDelivery(ctx, tx)
				case 4:
					tx := &StockLevel{
						db:        db,
						wID:       1,
						dID:       1,
						threshold: 10,
					}
					err = executor.ExecuteStockLevel(ctx, tx)
				}
				if err != nil {
					errChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			assert.NoError(t, err)
		}
	})
}

func TestBoundaryConditions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	config := &Config{
		Warehouses: 1,
		Terminals:  1,
	}

	ctx := context.Background()
	require.NoError(t, CreateSchema(ctx, db))
	loader := NewLoader(db, config)
	require.NoError(t, loader.Load(ctx))

	executor := NewTransactionExecutor(db, config)

	t.Run("MaxValues", func(t *testing.T) {
		tx := &NewOrder{
			db:       db,
			wID:      1,
			dID:      10,              // Max district ID
			cID:      3000,            // Max customer ID
			itemIDs:  make([]int, 15), // Max items
			supplyWs: make([]int, 15),
			qtys:     make([]int, 15),
			allLocal: true,
		}
		for i := range tx.itemIDs {
			tx.itemIDs[i] = 100000 // Max item ID
			tx.supplyWs[i] = 1
			tx.qtys[i] = 10 // Max quantity
		}
		assert.NoError(t, executor.ExecuteNewOrder(ctx, tx))
	})

	t.Run("MinValues", func(t *testing.T) {
		tx := &NewOrder{
			db:       db,
			wID:      1,
			dID:      1,
			cID:      1,
			itemIDs:  []int{1},
			supplyWs: []int{1},
			qtys:     []int{1},
			allLocal: true,
		}
		assert.NoError(t, executor.ExecuteNewOrder(ctx, tx))
	})
}

func TestErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	config := &Config{
		Warehouses: 1,
		Terminals:  1,
	}

	ctx := context.Background()
	require.NoError(t, CreateSchema(ctx, db))
	loader := NewLoader(db, config)
	require.NoError(t, loader.Load(ctx))

	executor := NewTransactionExecutor(db, config)

	t.Run("InvalidWarehouse", func(t *testing.T) {
		tx := &NewOrder{
			db:       db,
			wID:      999, // Non-existent warehouse
			dID:      1,
			cID:      1,
			itemIDs:  []int{1},
			supplyWs: []int{1},
			qtys:     []int{1},
			allLocal: true,
		}
		assert.Error(t, executor.ExecuteNewOrder(ctx, tx))
	})

	t.Run("InvalidItem", func(t *testing.T) {
		tx := &NewOrder{
			db:       db,
			wID:      1,
			dID:      1,
			cID:      1,
			itemIDs:  []int{999999}, // Non-existent item
			supplyWs: []int{1},
			qtys:     []int{1},
			allLocal: true,
		}
		assert.Error(t, executor.ExecuteNewOrder(ctx, tx))
	})

	t.Run("InvalidCustomer", func(t *testing.T) {
		tx := &Payment{
			db:     db,
			wID:    1,
			dID:    1,
			cID:    99999, // Non-existent customer
			amount: 100.0,
		}
		assert.Error(t, executor.ExecutePayment(ctx, tx))
	})

	t.Run("NegativeAmount", func(t *testing.T) {
		tx := &Payment{
			db:     db,
			wID:    1,
			dID:    1,
			cID:    1,
			amount: -100.0, // Negative amount
		}
		assert.Error(t, executor.ExecutePayment(ctx, tx))
	})
}

func TestPerformanceMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	config := &Config{
		Warehouses:            2,
		Terminals:             4,
		Duration:              5 * time.Second,
		ReportInterval:        1 * time.Second,
		NewOrderItemsMin:      5,
		NewOrderItemsMax:      15,
		NewOrderPercentage:    45,
		PaymentPercentage:     43,
		OrderStatusPercentage: 4,
		DeliveryPercentage:    4,
		StockLevelPercentage:  4,
	}

	ctx := context.Background()
	benchmark := NewTPCCBenchmark(config, db, logger)
	require.NoError(t, benchmark.Setup(ctx))

	// Run benchmark
	result, err := benchmark.Run(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify performance metrics
	assert.True(t, result.Metrics["tpmC"] > 0, "tpmC should be positive")
	assert.True(t, result.Metrics["efficiency"] > 0, "Efficiency should be positive")
	assert.True(t, result.Metrics["efficiency"] <= 100, "Efficiency should be <= 100")
	assert.True(t, result.Metrics["total_transactions"] > 0, "Should have some transactions")
	assert.True(t, result.Metrics["overall_latency_avg"] > 0, "Average latency should be positive")
}
