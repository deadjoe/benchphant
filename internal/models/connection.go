package models

import (
	"database/sql"
	"errors"
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
