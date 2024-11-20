package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	// DatabaseTypeMySQL represents MySQL database
	DatabaseTypeMySQL DatabaseType = "mysql"
	// DatabaseTypePostgreSQL represents PostgreSQL database
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
)

// Database represents a database instance
type Database struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Type        DatabaseType `json:"type"`
	Host        string       `json:"host"`
	Port        int          `json:"port"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	Database    string       `json:"database"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Validate checks if the database configuration is valid
func (c *Database) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Type == "" {
		return errors.New("type is required")
	}
	if c.Type != DatabaseTypeMySQL && c.Type != DatabaseTypePostgreSQL {
		return fmt.Errorf("invalid database type: %s", c.Type)
	}
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port == 0 {
		return errors.New("port is required")
	}
	if c.Username == "" {
		return errors.New("username is required")
	}
	if c.Database == "" {
		return errors.New("database is required")
	}
	return nil
}

// DSN returns the data source name for the database connection
func (c *Database) DSN() string {
	switch c.Type {
	case DatabaseTypeMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
	case DatabaseTypePostgreSQL:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", c.Host, c.Port, c.Username, c.Password, c.Database)
	default:
		return ""
	}
}

// TestConnection tests if the database connection is working
func (c *Database) TestConnection() error {
	db, err := sql.Open(string(c.Type), c.DSN())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()
	
	return db.Ping()
}

// ConnectionManager represents a manager of database connections
type ConnectionManager struct {
	mu          sync.Mutex
	connections []*sql.DB
	available   chan *sql.DB
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(db *sql.DB) (*ConnectionManager, error) {
	if db == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}
	return &ConnectionManager{
		connections: []*sql.DB{db},
		available:   make(chan *sql.DB, 1),
	}, nil
}

// Get gets a connection from the manager
func (p *ConnectionManager) Get() (*sql.DB, error) {
	select {
	case conn := <-p.available:
		return conn, nil
	default:
		p.mu.Lock()
		defer p.mu.Unlock()

		if len(p.connections) == 0 {
			return nil, nil
		}

		conn := p.connections[0]
		p.connections = p.connections[1:]
		return conn, nil
	}
}

// Put puts a connection back into the manager
func (p *ConnectionManager) Put(conn *sql.DB) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connections = append(p.connections, conn)
	select {
	case p.available <- conn:
	default:
	}
}

// Close closes all connections in the manager
func (p *ConnectionManager) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for _, conn := range p.connections {
		if err := conn.Close(); err != nil {
			lastErr = err
		}
	}
	p.connections = nil
	close(p.available)
	return lastErr
}

// Stats returns the connection manager statistics
func (p *ConnectionManager) Stats() json.RawMessage {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := map[string]interface{}{
		"total_connections": len(p.connections),
		"available":         len(p.available),
	}
	data, _ := json.Marshal(stats)
	return data
}
