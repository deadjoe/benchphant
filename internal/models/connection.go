package models

import (
	"database/sql"
	"errors"
	"sync"
	"time"
)

// DBType represents the type of database
type DBType string

// String implements fmt.Stringer interface
func (t DBType) String() string {
	return string(t)
}

const (
	// MySQL database type
	MySQL DBType = "mysql"
	// PostgreSQL database type
	PostgreSQL DBType = "postgresql"
)

// Common errors
var (
	ErrEmptyName     = errors.New("name is required")
	ErrEmptyHost     = errors.New("host is required")
	ErrInvalidPort   = errors.New("port is required")
	ErrEmptyDatabase = errors.New("database is required")
	ErrEmptyUsername = errors.New("username is required")
	ErrEmptyDriver   = errors.New("driver is required")
	ErrEmptyDSN      = errors.New("dsn is required")
)

// DBConnection represents a database connection
type DBConnection struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	Type        DBType            `json:"type"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Database    string            `json:"database"`
	Username    string            `json:"username"`
	Password    string            `json:"-"`
	Description string            `json:"description"`
	Driver      string            `json:"driver"`
	DSN         string            `json:"dsn"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastUsedAt  time.Time         `json:"last_used_at"`
	MaxIdleConn int               `json:"max_idle_conn"`
	MaxOpenConn int               `json:"max_open_conn"`
	Options     map[string]string `json:"options"`
	IsCluster   bool              `json:"is_cluster"`
	RouterHost  string            `json:"router_host"`
	RouterPort  int               `json:"router_port"`

	DB                *sql.DB `json:"-"`
	encryptedPassword string
}

// SetDB sets the database connection
func (c *DBConnection) SetDB(db *sql.DB) {
	c.DB = db
}

// SetEncryptedPassword sets the encrypted password
func (c *DBConnection) SetEncryptedPassword(password string) {
	c.encryptedPassword = password
}

// GetEncryptedPassword gets the encrypted password
func (c *DBConnection) GetEncryptedPassword() string {
	return c.encryptedPassword
}

// Validate validates the connection parameters
func (c *DBConnection) Validate() error {
	if c.Name == "" {
		return ErrEmptyName
	}
	if c.Host == "" {
		return ErrEmptyHost
	}
	if c.Port <= 0 {
		return ErrInvalidPort
	}
	if c.Database == "" {
		return ErrEmptyDatabase
	}
	if c.Username == "" {
		return ErrEmptyUsername
	}
	if c.Driver == "" {
		return ErrEmptyDriver
	}
	if c.DSN == "" {
		return ErrEmptyDSN
	}
	return nil
}

// ConnectionManager manages database connections
type ConnectionManager struct {
	db          *sql.DB
	connections map[int64]*DBConnection
	mu          sync.RWMutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(db *sql.DB) (*ConnectionManager, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &ConnectionManager{
		db:          db,
		connections: make(map[int64]*DBConnection),
	}, nil
}

// AddConnection adds a new connection
func (m *ConnectionManager) AddConnection(conn *DBConnection) error {
	if conn == nil {
		return errors.New("connection is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate new ID if not set
	if conn.ID == 0 {
		conn.ID = time.Now().UnixNano()
	}

	// Set timestamps
	now := time.Now()
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = now
	}
	conn.UpdatedAt = now
	conn.LastUsedAt = now

	// Store connection
	m.connections[conn.ID] = conn
	return nil
}

// GetConnection gets a connection by ID
func (m *ConnectionManager) GetConnection(id int64) (*DBConnection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, ok := m.connections[id]
	if !ok {
		return nil, errors.New("connection not found")
	}
	return conn, nil
}

// ListConnections lists all connections
func (m *ConnectionManager) ListConnections() []*DBConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns := make([]*DBConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		conns = append(conns, conn)
	}
	return conns
}

// UpdateConnection updates a connection
func (m *ConnectionManager) UpdateConnection(conn *DBConnection) error {
	if conn == nil {
		return errors.New("connection is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.connections[conn.ID]
	if !ok {
		return errors.New("connection not found")
	}

	// Update timestamps
	conn.CreatedAt = existing.CreatedAt
	conn.UpdatedAt = time.Now()

	// Store connection
	m.connections[conn.ID] = conn
	return nil
}

// DeleteConnection deletes a connection
func (m *ConnectionManager) DeleteConnection(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.connections[id]; !ok {
		return errors.New("connection not found")
	}

	delete(m.connections, id)
	return nil
}
