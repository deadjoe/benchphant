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
		benchmark := NewTPCCBenchmark(config, db, zaptest.NewLogger(t))
		assert.NoError(t, benchmark.Validate())

		// Test invalid config
		invalidConfig := *config
		invalidConfig.Warehouses = 0
		invalidBenchmark := NewTPCCBenchmark(&invalidConfig, db, zaptest.NewLogger(t))
		assert.Error(t, invalidBenchmark.Validate())
	})

	t.Run("Setup", func(t *testing.T) {
		benchmark := NewTPCCBenchmark(config, db, zaptest.NewLogger(t))
		ctx := context.Background()
		assert.NoError(t, benchmark.Setup(ctx))

		// Verify schema
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM warehouse").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, config.Warehouses, count)
	})

	t.Run("Run", func(t *testing.T) {
		benchmark := NewTPCCBenchmark(config, db, zaptest.NewLogger(t))
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
		benchmark := NewTPCCBenchmark(config, db, zaptest.NewLogger(t))
		ctx := context.Background()
		require.NoError(t, benchmark.Setup(ctx))
		assert.NoError(t, benchmark.Cleanup(ctx))

		// Verify schema is dropped
		_, err := db.Query("SELECT * FROM warehouse")
		assert.Error(t, err)
	})
}

func TestFactory(t *testing.T) {
	factory := NewFactory(zaptest.NewLogger(t))

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
		runner := NewRunner(db, config, zaptest.NewLogger(t))
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

type testTransaction struct {
	name     string
	executor func(context.Context) error
}

func runConcurrentTransactions(t *testing.T, ctx context.Context, transactions []testTransaction) {
	var wg sync.WaitGroup
	errChan := make(chan error, len(transactions))

	for _, tx := range transactions {
		wg.Add(1)
		go func(tx testTransaction) {
			defer wg.Done()
			if err := tx.executor(ctx); err != nil {
				t.Logf("Error executing %s: %v", tx.name, err)
				errChan <- err
			}
		}(tx)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		assert.NoError(t, err)
	}
}

func createNewOrderTx(executor *TransactionExecutor) testTransaction {
	return testTransaction{
		name: "NewOrder",
		executor: func(ctx context.Context) error {
			tx := &NewOrder{
				wID:      1,
				dID:      1,
				cID:      1,
				itemIDs:  []int{1, 2},
				supplyWs: []int{1, 1},
				qtys:     []int{1, 1},
				allLocal: true,
			}
			return executor.ExecuteNewOrder(ctx, tx)
		},
	}
}

func createPaymentTx(executor *TransactionExecutor) testTransaction {
	return testTransaction{
		name: "Payment",
		executor: func(ctx context.Context) error {
			tx := &Payment{
				wID:    1,
				dID:    1,
				cID:    1,
				amount: 100.0,
			}
			return executor.ExecutePayment(ctx, tx)
		},
	}
}

func createOrderStatusTx(executor *TransactionExecutor) testTransaction {
	return testTransaction{
		name: "OrderStatus",
		executor: func(ctx context.Context) error {
			tx := &OrderStatus{
				wID: 1,
				dID: 1,
				cID: 1,
			}
			return executor.ExecuteOrderStatus(ctx, tx)
		},
	}
}

func createDeliveryTx(executor *TransactionExecutor) testTransaction {
	return testTransaction{
		name: "Delivery",
		executor: func(ctx context.Context) error {
			tx := &Delivery{
				wID:       1,
				carrierID: 1,
			}
			return executor.ExecuteDelivery(ctx, tx)
		},
	}
}

func createStockLevelTx(executor *TransactionExecutor) testTransaction {
	return testTransaction{
		name: "StockLevel",
		executor: func(ctx context.Context) error {
			tx := &StockLevel{
				wID:       1,
				dID:       1,
				threshold: 10,
			}
			return executor.ExecuteStockLevel(ctx, tx)
		},
	}
}

func TestConcurrentTransactions(t *testing.T) {
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
		transactions := make([]testTransaction, 10)
		for i := range transactions {
			transactions[i] = createNewOrderTx(executor)
		}
		runConcurrentTransactions(t, ctx, transactions)
	})

	// Test concurrent mixed transactions
	t.Run("ConcurrentMixed", func(t *testing.T) {
		transactions := make([]testTransaction, 10)
		for i := range transactions {
			switch i % 5 {
			case 0:
				transactions[i] = createNewOrderTx(executor)
			case 1:
				transactions[i] = createPaymentTx(executor)
			case 2:
				transactions[i] = createOrderStatusTx(executor)
			case 3:
				transactions[i] = createDeliveryTx(executor)
			case 4:
				transactions[i] = createStockLevelTx(executor)
			}
		}
		runConcurrentTransactions(t, ctx, transactions)
	})
}

func TestTransactionExecutor(t *testing.T) {
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
			wID:    1,
			dID:    1,
			cID:    1,
			amount: 100.0,
		}
		assert.NoError(t, executor.ExecutePayment(ctx, tx))
	})

	t.Run("OrderStatus", func(t *testing.T) {
		tx := &OrderStatus{
			wID: 1,
			dID: 1,
			cID: 1,
		}
		assert.NoError(t, executor.ExecuteOrderStatus(ctx, tx))
	})

	t.Run("Delivery", func(t *testing.T) {
		tx := &Delivery{
			wID:       1,
			carrierID: 1,
		}
		assert.NoError(t, executor.ExecuteDelivery(ctx, tx))
	})

	t.Run("StockLevel", func(t *testing.T) {
		tx := &StockLevel{
			wID:       1,
			dID:       1,
			threshold: 10,
		}
		assert.NoError(t, executor.ExecuteStockLevel(ctx, tx))
	})
}

func TestBoundaryConditions(t *testing.T) {
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
			wID:    1,
			dID:    1,
			cID:    99999, // Non-existent customer
			amount: 100.0,
		}
		assert.Error(t, executor.ExecutePayment(ctx, tx))
	})

	t.Run("NegativeAmount", func(t *testing.T) {
		tx := &Payment{
			wID:    1,
			dID:    1,
			cID:    1,
			amount: -100.0, // Negative amount
		}
		assert.Error(t, executor.ExecutePayment(ctx, tx))
	})
}

func TestPerformanceMetrics(t *testing.T) {
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
	benchmark := NewTPCCBenchmark(config, db, zaptest.NewLogger(t))
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
