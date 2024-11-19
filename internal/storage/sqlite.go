package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements persistent storage using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return storage, nil
}

// initialize creates the necessary tables if they don't exist
func (s *SQLiteStorage) initialize() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS benchmarks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			config JSON,
			results JSON
		)`,
		`CREATE TABLE IF NOT EXISTS metrics (
			id TEXT PRIMARY KEY,
			benchmark_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			value REAL NOT NULL,
			labels JSON,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY(benchmark_id) REFERENCES benchmarks(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_benchmark_id ON metrics(benchmark_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", query, err)
		}
	}

	return nil
}

// StoreBenchmark stores a benchmark result
func (s *SQLiteStorage) StoreBenchmark(id, name, description string, config, results map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	now := time.Now()
	_, err = s.db.Exec(
		`INSERT INTO benchmarks (id, name, description, created_at, updated_at, config, results)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, name, description, now, now, configJSON, resultsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to store benchmark: %w", err)
	}

	return nil
}

// StoreMetric stores a metric
func (s *SQLiteStorage) StoreMetric(benchmarkID, name, metricType string, value float64, labels map[string]string) error {
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO metrics (id, benchmark_id, name, type, value, labels, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		fmt.Sprintf("%s-%s-%d", benchmarkID, name, time.Now().UnixNano()),
		benchmarkID,
		name,
		metricType,
		value,
		labelsJSON,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to store metric: %w", err)
	}

	return nil
}

// GetBenchmark retrieves a benchmark by ID
func (s *SQLiteStorage) GetBenchmark(id string) (map[string]interface{}, error) {
	var (
		name        string
		description string
		createdAt   time.Time
		updatedAt   time.Time
		configJSON  []byte
		resultsJSON []byte
	)

	err := s.db.QueryRow(
		`SELECT name, description, created_at, updated_at, config, results
		FROM benchmarks WHERE id = ?`,
		id,
	).Scan(&name, &description, &createdAt, &updatedAt, &configJSON, &resultsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query benchmark: %w", err)
	}

	var config, results map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if err := json.Unmarshal(resultsJSON, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": description,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
		"config":      config,
		"results":     results,
	}, nil
}

// GetMetrics retrieves metrics for a benchmark
func (s *SQLiteStorage) GetMetrics(benchmarkID string) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(
		`SELECT name, type, value, labels, timestamp
		FROM metrics WHERE benchmark_id = ?
		ORDER BY timestamp`,
		benchmarkID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []map[string]interface{}
	for rows.Next() {
		var (
			name      string
			typ       string
			value     float64
			labelsJSON []byte
			timestamp time.Time
		)

		if err := rows.Scan(&name, &typ, &value, &labelsJSON, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}

		var labels map[string]string
		if err := json.Unmarshal(labelsJSON, &labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}

		metrics = append(metrics, map[string]interface{}{
			"name":      name,
			"type":      typ,
			"value":     value,
			"labels":    labels,
			"timestamp": timestamp,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metric rows: %w", err)
	}

	return metrics, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
