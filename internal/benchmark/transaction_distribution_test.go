package benchmark

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// 设置固定的随机种子以获得可重复的测试结果
	rand.Seed(42)
}

func TestNewTPCCDistribution(t *testing.T) {
	dist := NewTPCCDistribution()
	assert.NotNil(t, dist)
	assert.NotNil(t, dist.counts)
	assert.Equal(t, 45.0, dist.weights.NewOrderWeight)
	assert.Equal(t, 43.0, dist.weights.PaymentWeight)
	assert.Equal(t, 4.0, dist.weights.OrderStatusWeight)
	assert.Equal(t, 4.0, dist.weights.DeliveryWeight)
	assert.Equal(t, 4.0, dist.weights.StockLevelWeight)
}

func TestTPCCDistributionValidation(t *testing.T) {
	tests := []struct {
		name        string
		weights     TransactionWeights
		expectError bool
	}{
		{
			name: "Valid weights",
			weights: TransactionWeights{
				NewOrderWeight:    45.0,
				PaymentWeight:     43.0,
				OrderStatusWeight: 4.0,
				DeliveryWeight:    4.0,
				StockLevelWeight:  4.0,
			},
			expectError: false,
		},
		{
			name: "Invalid weights - sum > 100",
			weights: TransactionWeights{
				NewOrderWeight:    50.0,
				PaymentWeight:     43.0,
				OrderStatusWeight: 4.0,
				DeliveryWeight:    4.0,
				StockLevelWeight:  4.0,
			},
			expectError: true,
		},
		{
			name: "Invalid weights - sum < 100",
			weights: TransactionWeights{
				NewOrderWeight:    40.0,
				PaymentWeight:     43.0,
				OrderStatusWeight: 4.0,
				DeliveryWeight:    4.0,
				StockLevelWeight:  4.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := &TPCCDistribution{weights: tt.weights}
			err := dist.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTPCCDistributionStats(t *testing.T) {
	dist := NewTPCCDistribution()

	// 测试初始状态
	stats := dist.GetDistributionStats()
	assert.Empty(t, stats)

	// 进行大量的事务选择以验证分布
	iterations := 100000
	for i := 0; i < iterations; i++ {
		dist.SelectTransactionType()
	}

	stats = dist.GetDistributionStats()

	// 允许 0.5% 的误差范围
	tolerance := 0.5
	assert.InDelta(t, 45.0, stats[NewOrder], tolerance)
	assert.InDelta(t, 43.0, stats[Payment], tolerance)
	assert.InDelta(t, 4.0, stats[OrderStatus], tolerance)
	assert.InDelta(t, 4.0, stats[Delivery], tolerance)
	assert.InDelta(t, 4.0, stats[StockLevel], tolerance)
}

func TestThinkTimes(t *testing.T) {
	dist := NewTPCCDistribution()

	// 测试每种事务类型的思考时间
	for txType, minTime := range ThinkTimes {
		// 对每种类型进行多次采样以验证分布
		for i := 0; i < 1000; i++ {
			thinkTime := dist.GetThinkTime(txType)
			seconds := float64(thinkTime) / float64(time.Second)

			// 思考时间应该在 minTime 和 minTime*2 之间
			assert.GreaterOrEqual(t, seconds, minTime)
			assert.LessOrEqual(t, seconds, minTime*2)
		}
	}
}

func TestKeyingTimes(t *testing.T) {
	dist := NewTPCCDistribution()

	// 测试每种事务类型的键入时间
	for txType, expectedTime := range KeyingTimes {
		keyingTime := dist.GetKeyingTime(txType)
		assert.Equal(t, time.Duration(expectedTime*float64(time.Second)), keyingTime)
	}
}

func TestConcurrency(t *testing.T) {
	dist := NewTPCCDistribution()
	iterations := 10000
	goroutines := 10
	done := make(chan bool)

	// 并发执行事务选择
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				dist.SelectTransactionType()
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < goroutines; i++ {
		<-done
	}

	stats := dist.GetDistributionStats()
	tolerance := 0.5
	assert.InDelta(t, 45.0, stats[NewOrder], tolerance)
	assert.InDelta(t, 43.0, stats[Payment], tolerance)
	assert.InDelta(t, 4.0, stats[OrderStatus], tolerance)
	assert.InDelta(t, 4.0, stats[Delivery], tolerance)
	assert.InDelta(t, 4.0, stats[StockLevel], tolerance)
}
