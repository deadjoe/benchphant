package metrics

import (
	"sync"
	"time"
)

// MetricType represents the type of metric being collected
type MetricType string

const (
	// MetricTypeGauge represents a metric that can go up and down
	MetricTypeGauge MetricType = "gauge"
	// MetricTypeCounter represents a metric that can only go up
	MetricTypeCounter MetricType = "counter"
	// MetricTypeHistogram represents a metric that tracks the distribution of values
	MetricTypeHistogram MetricType = "histogram"
)

// Metric represents a single metric with its metadata
type Metric struct {
	Name        string            // Name of the metric
	Type        MetricType        // Type of metric (gauge, counter, histogram)
	Value       float64           // Current value for gauge/counter
	Labels      map[string]string // Labels/tags associated with the metric
	LastUpdated time.Time         // Last time the metric was updated
}

// Collector manages the collection and storage of metrics
type Collector struct {
	metrics map[string]*Metric
	mu      sync.RWMutex
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		metrics: make(map[string]*Metric),
	}
}

// SetGauge sets a gauge metric value
func (c *Collector) SetGauge(name string, value float64, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics[name] = &Metric{
		Name:        name,
		Type:        MetricTypeGauge,
		Value:       value,
		Labels:      labels,
		LastUpdated: time.Now(),
	}
}

// IncrementCounter increments a counter metric by the given value
func (c *Collector) IncrementCounter(name string, value float64, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metric, exists := c.metrics[name]; exists {
		metric.Value += value
		metric.LastUpdated = time.Now()
	} else {
		c.metrics[name] = &Metric{
			Name:        name,
			Type:        MetricTypeCounter,
			Value:       value,
			Labels:      labels,
			LastUpdated: time.Now(),
		}
	}
}

// GetMetric returns a metric by name
func (c *Collector) GetMetric(name string) (*Metric, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metric, exists := c.metrics[name]
	return metric, exists
}

// GetAllMetrics returns all metrics
func (c *Collector) GetAllMetrics() []*Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := make([]*Metric, 0, len(c.metrics))
	for _, metric := range c.metrics {
		metrics = append(metrics, metric)
	}
	return metrics
}

// Reset resets all metrics
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = make(map[string]*Metric)
}
