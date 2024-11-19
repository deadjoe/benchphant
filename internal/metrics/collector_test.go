package metrics

import (
	"testing"
	"time"
)

func TestCollector_SetGauge(t *testing.T) {
	c := NewCollector()
	labels := map[string]string{"db": "test"}
	
	c.SetGauge("test_gauge", 42.0, labels)
	
	metric, exists := c.GetMetric("test_gauge")
	if !exists {
		t.Fatal("Expected metric to exist")
	}
	
	if metric.Type != MetricTypeGauge {
		t.Errorf("Expected metric type %s, got %s", MetricTypeGauge, metric.Type)
	}
	
	if metric.Value != 42.0 {
		t.Errorf("Expected metric value 42.0, got %f", metric.Value)
	}
}

func TestCollector_IncrementCounter(t *testing.T) {
	c := NewCollector()
	labels := map[string]string{"db": "test"}
	
	// First increment
	c.IncrementCounter("test_counter", 1.0, labels)
	metric, exists := c.GetMetric("test_counter")
	if !exists {
		t.Fatal("Expected metric to exist")
	}
	if metric.Value != 1.0 {
		t.Errorf("Expected metric value 1.0, got %f", metric.Value)
	}
	
	// Second increment
	c.IncrementCounter("test_counter", 2.0, labels)
	metric, exists = c.GetMetric("test_counter")
	if !exists {
		t.Fatal("Expected metric to exist")
	}
	if metric.Value != 3.0 {
		t.Errorf("Expected metric value 3.0, got %f", metric.Value)
	}
}

func TestCollector_GetAllMetrics(t *testing.T) {
	c := NewCollector()
	labels := map[string]string{"db": "test"}
	
	c.SetGauge("gauge1", 1.0, labels)
	c.SetGauge("gauge2", 2.0, labels)
	
	metrics := c.GetAllMetrics()
	if len(metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(metrics))
	}
}

func TestCollector_Reset(t *testing.T) {
	c := NewCollector()
	labels := map[string]string{"db": "test"}
	
	c.SetGauge("gauge1", 1.0, labels)
	c.SetGauge("gauge2", 2.0, labels)
	
	c.Reset()
	
	metrics := c.GetAllMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics after reset, got %d", len(metrics))
	}
}

func TestMetric_LastUpdated(t *testing.T) {
	c := NewCollector()
	labels := map[string]string{"db": "test"}
	
	beforeSet := time.Now()
	time.Sleep(time.Millisecond) // Ensure some time passes
	
	c.SetGauge("test_gauge", 42.0, labels)
	
	metric, exists := c.GetMetric("test_gauge")
	if !exists {
		t.Fatal("Expected metric to exist")
	}
	
	if metric.LastUpdated.Before(beforeSet) {
		t.Error("Expected LastUpdated to be after the set time")
	}
}
