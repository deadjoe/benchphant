package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DBType represents the type of database
type DBType string

const (
	// MySQL database type
	MySQL DBType = "mysql"
	// PostgreSQL database type
	PostgreSQL DBType = "postgresql"
)

// String implements fmt.Stringer interface
func (t DBType) String() string {
	return string(t)
}

// DBConnection represents a database connection configuration
type DBConnection struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        DBType    `json:"type"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Database    string    `json:"database"`
	Options     string    `json:"options"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	IsCluster   bool      `json:"is_cluster"`
	RouterHost  string    `json:"router_host,omitempty"`
	RouterPort  int       `json:"router_port,omitempty"`
	MaxIdleConn int       `json:"max_idle_conn"`
	MaxOpenConn int       `json:"max_open_conn"`

	// Internal fields for encrypted data
	encryptedPassword string

	// Underlying database connection
	db *sql.DB
}

// Validate checks if the connection configuration is valid
func (c *DBConnection) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Type == "" {
		return errors.New("type is required")
	}
	if c.Type != MySQL && c.Type != PostgreSQL {
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
	if c.MaxIdleConn <= 0 {
		return errors.New("max idle connections must be positive")
	}
	if c.MaxOpenConn <= 0 {
		return errors.New("max open connections must be positive")
	}
	if c.IsCluster {
		if c.RouterHost == "" {
			return errors.New("router host is required for cluster mode")
		}
		if c.RouterPort == 0 {
			return errors.New("router port is required for cluster mode")
		}
	}
	return nil
}

// SetEncryptedPassword sets the encrypted password and clears the plain text password
func (c *DBConnection) SetEncryptedPassword(encrypted string) {
	c.encryptedPassword = encrypted
	c.Password = "" // Clear plain text password for security
}

// GetEncryptedPassword returns the encrypted password
func (c *DBConnection) GetEncryptedPassword() string {
	return c.encryptedPassword
}

// DSN returns the data source name for the database connection
func (c *DBConnection) DSN() string {
	host := c.getHost()
	port := c.getPort()
	options := c.getOptions()

	switch c.Type {
	case MySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.Username, c.Password, host, port, c.Database, options)
	case PostgreSQL:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s", host, port, c.Username, c.Password, c.Database, options)
	default:
		return ""
	}
}

// getHost returns the appropriate host based on whether it's a cluster or not
func (c *DBConnection) getHost() string {
	if c.IsCluster && c.RouterHost != "" {
		return c.RouterHost
	}
	return c.Host
}

// getPort returns the appropriate port based on whether it's a cluster or not
func (c *DBConnection) getPort() int {
	if c.IsCluster && c.RouterPort != 0 {
		return c.RouterPort
	}
	return c.Port
}

// getOptions returns the connection options with defaults if not specified
func (c *DBConnection) getOptions() string {
	if c.Options != "" {
		return c.Options
	}

	switch c.Type {
	case MySQL:
		return "parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci"
	case PostgreSQL:
		return "sslmode=disable"
	default:
		return ""
	}
}

// DB returns the underlying sql.DB
func (c *DBConnection) DB() *sql.DB {
	return c.db
}

// SetDB sets the underlying sql.DB
func (c *DBConnection) SetDB(db *sql.DB) {
	c.db = db
}

// Close closes the database connection
func (c *DBConnection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// NewDBConnection creates a new database connection
func NewDBConnection(db *sql.DB) *DBConnection {
	conn := &DBConnection{}
	conn.SetDB(db)
	return conn
}

// ConnectionManager represents a manager of database connections
type ConnectionManager struct {
	mu          sync.Mutex
	connections []*DBConnection
	available   chan *DBConnection
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(conn *DBConnection) (*ConnectionManager, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}
	return &ConnectionManager{
		connections: []*DBConnection{conn},
		available:   make(chan *DBConnection, 1),
	}, nil
}

// Get gets a connection from the manager
func (p *ConnectionManager) Get() (*DBConnection, error) {
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
func (p *ConnectionManager) Put(conn *DBConnection) {
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

// TestConnection tests if the database connection is working
func (p *ConnectionManager) TestConnection() error {
	conn, err := p.Get()
	if err != nil {
		return err
	}
	defer p.Put(conn)

	if conn.db == nil {
		return fmt.Errorf("no database connection available")
	}
	return conn.db.Ping()
}

// Stats returns the connection manager statistics
func (p *ConnectionManager) Stats() json.RawMessage {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := map[string]interface{}{
		"total_connections": len(p.connections),
		"available":        len(p.available),
	}
	data, _ := json.Marshal(stats)
	return data
}
