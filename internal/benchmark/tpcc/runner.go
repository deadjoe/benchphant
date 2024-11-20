package tpcc

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Runner coordinates the execution of TPC-C transactions
type Runner struct {
	db        *sql.DB
	config    *Config
	logger    *zap.Logger
	stats     *Stats
	executor  *TransactionExecutor
	stopChan  chan struct{}
	terminals []*Terminal
	wg        sync.WaitGroup
}

// Terminal represents a client terminal that executes transactions
type Terminal struct {
	id       int
	wID      int // Home warehouse ID
	dID      int // Home district ID
	runner   *Runner
	stopChan chan struct{}
	rng      *rand.Rand // Per-terminal random number generator
}

// NewRunner creates a new TPC-C test runner
func NewRunner(db *sql.DB, config *Config, logger *zap.Logger) *Runner {
	return &Runner{
		db:       db,
		config:   config,
		logger:   logger,
		stats:    NewStats(),
		executor: NewTransactionExecutor(db, config),
		stopChan: make(chan struct{}),
	}
}

// Run executes the TPC-C test
func (r *Runner) Run(ctx context.Context) error {
	// Initialize test
	if err := r.initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize test: %w", err)
	}

	// Create worker channels
	workCh := make(chan struct{}, r.config.NumThreads)
	errCh := make(chan error, r.config.NumThreads)

	// Start workers
	for i := 0; i < r.config.NumThreads; i++ {
		go r.worker(ctx, workCh, errCh)
	}

	// Parse duration
	duration, err := time.ParseDuration(r.config.Duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	// Run test for the specified duration
	timer := time.NewTimer(duration)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case <-timer.C:
			close(workCh)
			r.stats.Finalize()
			return nil
		default:
			workCh <- struct{}{}
		}
	}
}

// GetStats returns the current test statistics
func (r *Runner) GetStats() *Stats {
	return r.stats
}

// initialize prepares the database for testing
func (r *Runner) initialize(ctx context.Context) error {
	r.logger.Info("Initializing TPC-C test",
		zap.Int("warehouses", r.config.Warehouses),
		zap.Int("terminals", r.config.Terminals),
		zap.Duration("duration", r.config.Duration),
		zap.Float64("new_order_percentage", r.config.NewOrderPercentage),
		zap.Float64("payment_percentage", r.config.PaymentPercentage),
		zap.Float64("order_status_percentage", r.config.OrderStatusPercentage),
		zap.Float64("delivery_percentage", r.config.DeliveryPercentage),
		zap.Float64("stock_level_percentage", r.config.StockLevelPercentage),
	)

	// Initialize statistics
	r.stats.StartTime = time.Now()

	return nil
}

// startTerminals starts all client terminals
func (r *Runner) startTerminals() {
	r.terminals = make([]*Terminal, r.config.Terminals)
	r.wg.Add(r.config.Terminals)

	for i := 0; i < r.config.Terminals; i++ {
		terminal := &Terminal{
			id:       i + 1,
			wID:      (i % r.config.Warehouses) + 1,
			dID:      (i % 10) + 1, // Each warehouse has 10 districts
			runner:   r,
			stopChan: make(chan struct{}),
			rng:      rand.New(rand.NewSource(time.Now().UnixNano() + int64(i))),
		}
		r.terminals[i] = terminal
		go terminal.run()
	}
}

// stop stops all terminals and the test
func (r *Runner) stop() {
	close(r.stopChan)
	for _, t := range r.terminals {
		close(t.stopChan)
	}
}

// monitor periodically reports test progress
func (r *Runner) monitor(ctx context.Context) {
	ticker := time.NewTicker(r.config.ReportInterval)
	defer ticker.Stop()

	lastStats := *r.stats
	lastTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentStats := *r.stats
			currentTime := time.Now()
			interval := currentTime.Sub(lastTime)

			// Calculate interval metrics
			newOrders := currentStats.NewOrderCount - lastStats.NewOrderCount
			tpmC := float64(newOrders) / interval.Minutes()

			r.logger.Info("Test progress",
				zap.Duration("elapsed", time.Since(r.stats.StartTime)),
				zap.Int64("total_transactions", currentStats.TotalTransactions),
				zap.Int64("total_errors", currentStats.TotalErrors),
				zap.Float64("current_tpmC", tpmC),
				zap.Float64("overall_tpmC", float64(currentStats.NewOrderCount)/time.Since(r.stats.StartTime).Minutes()),
				zap.Float64("efficiency", float64(currentStats.TotalTransactions-currentStats.TotalErrors)/float64(currentStats.TotalTransactions)*100),
				zap.Int64("new_orders", currentStats.NewOrderCount),
				zap.Int64("payments", currentStats.PaymentCount),
				zap.Int64("order_status", currentStats.OrderStatusCount),
				zap.Int64("deliveries", currentStats.DeliveryCount),
				zap.Int64("stock_level", currentStats.StockLevelCount),
			)

			lastStats = currentStats
			lastTime = currentTime
		}
	}
}

// calculateStats calculates final test statistics
func (r *Runner) calculateStats() {
	r.stats.EndTime = time.Now()
	r.stats.Duration = r.stats.EndTime.Sub(r.stats.StartTime)
	r.stats.TPMc = float64(r.stats.NewOrderCount) / r.stats.Duration.Minutes()
	if r.stats.TotalTransactions > 0 {
		r.stats.Efficiency = float64(r.stats.TotalTransactions-r.stats.TotalErrors) / float64(r.stats.TotalTransactions) * 100
	}

	r.logger.Info("Test completed",
		zap.Duration("duration", r.stats.Duration),
		zap.Int64("total_transactions", r.stats.TotalTransactions),
		zap.Int64("total_errors", r.stats.TotalErrors),
		zap.Float64("tpmC", r.stats.TPMc),
		zap.Float64("efficiency", r.stats.Efficiency),
		zap.Float64("latency_avg_ms", r.stats.OverallLatencyAvg),
	)
}

// run executes transactions for a terminal
func (t *Terminal) run() {
	defer t.runner.wg.Done()

	for {
		select {
		case <-t.stopChan:
			return
		default:
			if err := t.executeTransaction(); err != nil {
				t.runner.logger.Error("Transaction error",
					zap.Int("terminal", t.id),
					zap.Error(err),
				)
			}
		}
	}
}

// executeTransaction executes a random transaction based on the configured mix
func (t *Terminal) executeTransaction() error {
	r := t.rng.Float64() * 100
	start := time.Now()
	var err error

	switch {
	case r < t.runner.config.NewOrderPercentage:
		err = t.executeNewOrderTransaction()
		if err == nil {
			atomic.AddInt64(&t.runner.stats.NewOrderCount, 1)
		} else {
			atomic.AddInt64(&t.runner.stats.NewOrderErrors, 1)
		}

	case r < t.runner.config.NewOrderPercentage+t.runner.config.PaymentPercentage:
		err = t.executePaymentTransaction()
		if err == nil {
			atomic.AddInt64(&t.runner.stats.PaymentCount, 1)
		} else {
			atomic.AddInt64(&t.runner.stats.PaymentErrors, 1)
		}

	case r < t.runner.config.NewOrderPercentage+t.runner.config.PaymentPercentage+t.runner.config.OrderStatusPercentage:
		err = t.executeOrderStatusTransaction()
		if err == nil {
			atomic.AddInt64(&t.runner.stats.OrderStatusCount, 1)
		} else {
			atomic.AddInt64(&t.runner.stats.OrderStatusErrors, 1)
		}

	case r < t.runner.config.NewOrderPercentage+t.runner.config.PaymentPercentage+t.runner.config.OrderStatusPercentage+t.runner.config.DeliveryPercentage:
		err = t.executeDeliveryTransaction()
		if err == nil {
			atomic.AddInt64(&t.runner.stats.DeliveryCount, 1)
		} else {
			atomic.AddInt64(&t.runner.stats.DeliveryErrors, 1)
		}

	default:
		err = t.executeStockLevelTransaction()
		if err == nil {
			atomic.AddInt64(&t.runner.stats.StockLevelCount, 1)
		} else {
			atomic.AddInt64(&t.runner.stats.StockLevelErrors, 1)
		}
	}

	if err == nil {
		atomic.AddInt64(&t.runner.stats.TotalTransactions, 1)
		t.updateLatencyStats(time.Since(start))
	} else {
		atomic.AddInt64(&t.runner.stats.TotalErrors, 1)
	}

	return err
}

// executeNewOrderTransaction executes a New-Order transaction
func (t *Terminal) executeNewOrderTransaction() error {
	numItems := t.rng.Intn(t.runner.config.NewOrderItemsMax-t.runner.config.NewOrderItemsMin+1) + t.runner.config.NewOrderItemsMin
	itemIDs := make([]int, numItems)
	supplyWs := make([]int, numItems)
	qtys := make([]int, numItems)
	allLocal := true

	for i := 0; i < numItems; i++ {
		itemIDs[i] = t.rng.Intn(100000) + 1
		if t.rng.Float64() < 0.01 { // 1% remote
			supplyWs[i] = t.rng.Intn(t.runner.config.Warehouses) + 1
			if supplyWs[i] != t.wID {
				allLocal = false
			}
		} else {
			supplyWs[i] = t.wID
		}
		qtys[i] = t.rng.Intn(10) + 1 // 1-10 quantities
	}

	tx := &NewOrder{
		db:       t.runner.db,
		wID:      t.wID,
		dID:      t.dID,
		cID:      t.rng.Intn(3000) + 1,
		itemIDs:  itemIDs,
		supplyWs: supplyWs,
		qtys:     qtys,
		allLocal: allLocal,
	}

	return t.runner.executor.ExecuteNewOrder(context.Background(), tx)
}

// executePaymentTransaction executes a Payment transaction
func (t *Terminal) executePaymentTransaction() error {
	tx := &Payment{
		db:     t.runner.db,
		wID:    t.wID,
		dID:    t.dID,
		cID:    t.rng.Intn(3000) + 1,
		amount: float64(t.rng.Intn(5000)+1) / 100.0, // $1.00-$50.00
	}

	return t.runner.executor.ExecutePayment(context.Background(), tx)
}

// executeOrderStatusTransaction executes an Order-Status transaction
func (t *Terminal) executeOrderStatusTransaction() error {
	tx := &OrderStatus{
		db:  t.runner.db,
		wID: t.wID,
		dID: t.dID,
		cID: t.rng.Intn(3000) + 1,
	}

	return t.runner.executor.ExecuteOrderStatus(context.Background(), tx)
}

// executeDeliveryTransaction executes a Delivery transaction
func (t *Terminal) executeDeliveryTransaction() error {
	tx := &Delivery{
		db:        t.runner.db,
		wID:       t.wID,
		carrierID: t.rng.Intn(10) + 1,
	}

	return t.runner.executor.ExecuteDelivery(context.Background(), tx)
}

// executeStockLevelTransaction executes a Stock-Level transaction
func (t *Terminal) executeStockLevelTransaction() error {
	tx := &StockLevel{
		db:        t.runner.db,
		wID:       t.wID,
		dID:       t.dID,
		threshold: t.rng.Intn(11) + 10, // 10-20 threshold
	}

	return t.runner.executor.ExecuteStockLevel(context.Background(), tx)
}

// updateLatencyStats updates latency statistics for a transaction
func (t *Terminal) updateLatencyStats(d time.Duration) {
	ms := float64(d.Milliseconds())
	atomic.StoreFloat64(&t.runner.stats.OverallLatencyAvg,
		(atomic.LoadFloat64(&t.runner.stats.OverallLatencyAvg)*float64(atomic.LoadInt64(&t.runner.stats.TotalTransactions)-1)+ms)/
			float64(atomic.LoadInt64(&t.runner.stats.TotalTransactions)))
}
