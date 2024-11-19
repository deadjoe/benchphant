package benchmark

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// OLTPTransactionType represents different types of OLTP transactions
type OLTPTransactionType string

const (
	// NewOrder represents the New-Order transaction
	NewOrder OLTPTransactionType = "new-order"
	// Payment represents the Payment transaction
	Payment OLTPTransactionType = "payment"
	// OrderStatus represents the Order-Status transaction
	OrderStatus OLTPTransactionType = "order-status"
	// Delivery represents the Delivery transaction
	Delivery OLTPTransactionType = "delivery"
	// StockLevel represents the Stock-Level transaction
	StockLevel OLTPTransactionType = "stock-level"
)

// TransactionWeights represents the TPC-C specified weights for each transaction type
type TransactionWeights struct {
	NewOrderWeight     float64
	PaymentWeight     float64
	OrderStatusWeight float64
	DeliveryWeight    float64
	StockLevelWeight  float64
}

// TPCCDistribution represents a TPC-C compliant transaction distribution
type TPCCDistribution struct {
	weights TransactionWeights
	counts  map[OLTPTransactionType]int
	total   int
	mu      sync.Mutex
}

// ThinkTimes represents the minimum think time in seconds for each transaction type
var ThinkTimes = map[OLTPTransactionType]float64{
	NewOrder:    12.0,
	Payment:     12.0,
	OrderStatus: 10.0,
	Delivery:    5.0,
	StockLevel:  5.0,
}

// KeyingTimes represents the fixed keying time in seconds for each transaction type
var KeyingTimes = map[OLTPTransactionType]float64{
	NewOrder:    18.0,
	Payment:     3.0,
	OrderStatus: 2.0,
	Delivery:    2.0,
	StockLevel:  2.0,
}

// NewTPCCDistribution creates a new TPC-C compliant transaction distribution
func NewTPCCDistribution() *TPCCDistribution {
	return &TPCCDistribution{
		weights: TransactionWeights{
			NewOrderWeight:     45.0,
			PaymentWeight:     43.0,
			OrderStatusWeight: 4.0,
			DeliveryWeight:    4.0,
			StockLevelWeight:  4.0,
		},
		counts: make(map[OLTPTransactionType]int),
	}
}

// Validate ensures the transaction weights sum to 100
func (td *TPCCDistribution) Validate() error {
	sum := td.weights.NewOrderWeight +
		td.weights.PaymentWeight +
		td.weights.OrderStatusWeight +
		td.weights.DeliveryWeight +
		td.weights.StockLevelWeight

	if sum != 100.0 {
		return fmt.Errorf("transaction weights must sum to 100, got %.2f", sum)
	}
	return nil
}

// SelectTransactionType selects a transaction type based on TPC-C distribution
func (td *TPCCDistribution) SelectTransactionType() OLTPTransactionType {
	td.mu.Lock()
	defer td.mu.Unlock()

	r := rand.Float64() * 100.0
	sum := 0.0

	// Use cumulative probabilities for more accurate distribution
	sum += td.weights.NewOrderWeight
	if r < sum {
		td.counts[NewOrder]++
		td.total++
		return NewOrder
	}

	sum += td.weights.PaymentWeight
	if r < sum {
		td.counts[Payment]++
		td.total++
		return Payment
	}

	sum += td.weights.OrderStatusWeight
	if r < sum {
		td.counts[OrderStatus]++
		td.total++
		return OrderStatus
	}

	sum += td.weights.DeliveryWeight
	if r < sum {
		td.counts[Delivery]++
		td.total++
		return Delivery
	}

	// Must be StockLevel
	td.counts[StockLevel]++
	td.total++
	return StockLevel
}

// GetDistributionStats returns the current distribution percentages
func (td *TPCCDistribution) GetDistributionStats() map[OLTPTransactionType]float64 {
	td.mu.Lock()
	defer td.mu.Unlock()

	stats := make(map[OLTPTransactionType]float64)
	if td.total == 0 {
		return stats
	}

	for txType, count := range td.counts {
		stats[txType] = float64(count) / float64(td.total) * 100.0
	}
	return stats
}

// GetThinkTime returns a random think time for the given transaction type
func (td *TPCCDistribution) GetThinkTime(txType OLTPTransactionType) time.Duration {
	minTime := ThinkTimes[txType]
	// Generate random think time between minTime and minTime*2 as per TPC-C spec
	thinkTime := minTime + (minTime * rand.Float64())
	return time.Duration(thinkTime * float64(time.Second))
}

// GetKeyingTime returns the fixed keying time for the given transaction type
func (td *TPCCDistribution) GetKeyingTime(txType OLTPTransactionType) time.Duration {
	return time.Duration(KeyingTimes[txType] * float64(time.Second))
}
