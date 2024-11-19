package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/deadjoe/benchphant/internal/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// ConnectionPool manages a pool of database connections
type ConnectionPool struct {
	db     *sql.DB
	config *models.DBConnection
	mu     sync.RWMutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *models.DBConnection) (*ConnectionPool, error) {
	dsn := generateDSN(config)
	db, err := sql.Open(string(config.Type), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure pool settings
	db.SetMaxIdleConns(config.MaxIdleConn)
	db.SetMaxOpenConns(config.MaxOpenConn)
	db.SetConnMaxLifetime(time.Hour)

	return &ConnectionPool{
		db:     db,
		config: config,
	}, nil
}

// Get gets a connection from the pool
func (p *ConnectionPool) Get() (*sql.DB, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.db == nil {
		return nil, fmt.Errorf("connection pool is closed")
	}

	return p.db, nil
}

// GetDB returns the underlying sql.DB instance
func (p *ConnectionPool) GetDB() *sql.DB {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.db
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.db != nil {
		err := p.db.Close()
		p.db = nil
		return err
	}

	return nil
}

// generateDSN generates a Data Source Name for the database connection
func generateDSN(config *models.DBConnection) string {
	switch config.Type {
	case models.MySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
			config.Options)

	case models.PostgreSQL:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
			config.Host,
			config.Port,
			config.Username,
			config.Password,
			config.Database,
			config.Options)

	default:
		return ""
	}
}
